import { describe, expect, it } from 'vitest'
import { checkStub, stubChecker } from './contract'

const json = 'application/json'
const problemJson = 'application/problem+json'
const note = { id: '00000000-0000-4000-8000-000000000001', text: 'Buy milk', createdAt: '2026-10-02T09:30:00Z' }
const problem = { type: 'about:blank', title: 'Unprocessable Entity', status: 422, detail: 'Too long.' }

describe("checking a stub against the service's contract", () => {
  it('accepts an answer the contract describes', () => {
    expect(checkStub('GET', '/api/v1/notes', 200, json, { items: [note] })).toBeNull()
  })

  it('refuses a field the contract does not allow', () => {
    expect(checkStub('GET', '/api/v1/notes', 200, json, { items: [], next: 'x' })).toMatch(/must NOT have additional properties/)
  })

  it('refuses an answer missing a required field', () => {
    expect(checkStub('POST', '/api/v1/notes', 201, json, { id: note.id })).toMatch(/must have required property 'text'/)
  })

  it('accepts a problem the operation names by status', () => {
    expect(checkStub('POST', '/api/v1/notes', 422, problemJson, problem)).toBeNull()
  })

  it('checks a status the operation does not name against its default answer', () => {
    expect(checkStub('GET', '/api/v1/notes', 500, problemJson, { title: 'Internal Server Error', reason: 'no' })).not.toBeNull()
  })

  it('refuses a path the contract does not have', () => {
    expect(checkStub('GET', '/api/v1/nothing', 200, json, {})).toBe('the contract has no GET /api/v1/nothing')
  })

  it('refuses a method the path does not take', () => {
    expect(checkStub('DELETE', '/api/v1/notes', 204, null, null)).toBe('the contract has no DELETE /api/v1/notes')
  })

  it('refuses an empty answer where the contract has a body', () => {
    expect(checkStub('GET', '/api/v1/notes', 200, null, null)).toBe('GET /api/v1/notes answers 200 as application/json, not no body')
  })

  it('refuses a media type the answer does not have', () => {
    expect(checkStub('GET', '/api/v1/notes', 200, 'text/plain', 'x')).toBe('GET /api/v1/notes answers 200 as application/json, not text/plain')
  })
})

// A contract with what this service's does not have yet: a path parameter, a status range and an
// answer without a body.
const other = stubChecker(`
openapi: 3.1.0
info: { title: other, version: '1' }
paths:
  /things/{id}:
    delete:
      responses:
        '204': { description: Deleted }
        4XX:
          description: Refused
          content: { application/json: { schema: { type: object, required: [why] } } }
        default: { description: Failed }
components: { responses: {} }
`)

describe('checking a stub against any contract', () => {
  it('matches a path with parameters to its template', () => {
    expect(other('DELETE', '/things/42', 204, null, null)).toBeNull()
  })

  it('refuses a body on an answer the contract says has none', () => {
    expect(other('DELETE', '/things/42', 204, json, { ok: true })).toBe('DELETE /things/42 answers 204 with no body')
  })

  it('checks a status against the range that covers it', () => {
    expect(other('DELETE', '/things/42', 409, json, {})).toMatch(/must have required property 'why'/)
  })
})
