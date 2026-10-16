import { afterEach, describe, expect, it } from 'vitest'
import { client } from './api/client.gen'
import { configureApi } from './http'

afterEach(() => {
  configureApi()
})

describe('pointing the API client', () => {
  it("calls the page's own origin when given nothing", () => {
    configureApi({ baseUrl: 'http://elsewhere.test' })
    configureApi()
    expect(client.getConfig().baseUrl).toBe('')
  })

  it('calls through the fetch it is given', () => {
    const fetch = () => Promise.resolve(new Response())
    configureApi({ fetch })
    expect(client.getConfig().fetch).toBe(fetch)
  })
})
