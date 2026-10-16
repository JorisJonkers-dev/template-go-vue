import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'
import { flushPromises, mount } from '@vue/test-utils'
import { expect, vi } from 'vitest'
import { createMemoryHistory } from 'vue-router'
import { configureApi } from '../infrastructure/http'
import App from './App.vue'
import { checkStub } from './contract'
import { createAppRouter } from './router'

/** How a test answers the app's requests. */
export type Handler = (request: Request) => Response | Promise<Response>

export function json(body: unknown, status = 200, contentType = 'application/json'): Response {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': contentType } })
}

/**
 * Mounts the whole app at path, with the network answered by handle. Every answer must be one the
 * contract allows, unless the test is about an answer the server should never send: offContract.
 */
export async function mountApp(handle: Handler, path = '/', { offContract = false } = {}) {
  const fetch = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const request = new Request(input, init)
    const response = await handle(request)
    if (!offContract) await keepsTheContract(request.method, new URL(request.url).pathname, response)
    return response
  })
  configureApi({ baseUrl: 'http://app.test', fetch })
  const router = createAppRouter(createMemoryHistory())
  await router.push(path)
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const wrapper = mount(App, { global: { plugins: [router, [VueQueryPlugin, { queryClient }]] } })
  await router.isReady()
  await settle()
  return { wrapper, fetch, router }
}

/**
 * Fails the running test, without stopping it, when a stub answers what the server never would.
 * The page may swallow the bad answer, so the failure cannot rest on what the page shows.
 */
async function keepsTheContract(method: string, path: string, response: Response) {
  const text = await response.clone().text()
  const mediaType = response.headers.get('Content-Type')?.split(';')[0] ?? null
  const violation = checkStub(method, path, response.status, text ? mediaType : null, text ? (JSON.parse(text) as unknown) : null)
  expect.soft(violation, `the stub for ${method} ${path} disagrees with the contract`).toBeNull()
}

export async function settle() {
  await flushPromises()
  await vi.dynamicImportSettled()
  await flushPromises()
}
