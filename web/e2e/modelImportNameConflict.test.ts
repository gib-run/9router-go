/**
 * "Import from /models" stopped at the first duplicate name.
 *
 * POST /api/models/custom refuses an id that already addresses a model alias or
 * a combo with a typed 409 (MODEL_ALIAS_CONFLICT). The import loop awaited every
 * save inside one try/catch, so the first refused id threw past the ids listed
 * after it: the run ended, the alert carried the raw 409, and every model below
 * the collision was silently never imported — the report was against an
 * OpenAI-compatible node whose /models list contained `claude-opus-4-6`, which a
 * model alias already owned.
 *
 * Driven as reported: a `/models` endpoint listing the colliding id FIRST, then
 * a fresh one, with the alias in place. The observable contract is the alert
 * naming the refused id and the fresh id still landing in /api/models/custom —
 * an assertion that the loop merely "did not throw" would pass with the import
 * still half-done.
 */

import { afterAll, beforeAll, describe, expect, test } from 'bun:test'
import type { Page } from '@playwright/test'

import { parseCustomModelsResponse } from '../src/lib/customModels'
import { startDashboard, type Dashboard } from './harness'

/** Collides with the model alias created below. */
const ALIASED_ID = 'claude-opus-4-6'
/** Listed after the collision: must still be imported. */
const FRESH_ID = 'zzz-fresh-model'

const PORT = 20441
const UPSTREAM_PORT = 20442

const NODE_SUFFIX = 'e2eimportconflict'
const NODE_ID = `openai-compatible-chat-${NODE_SUFFIX}`

let app: Dashboard
let page: Page
/** Stopper for the fake /models endpoint; undefined until it listens. */
let stopUpstream: (() => void) | undefined

/** The /models list, aliased id first so the old loop imported nothing at all. */
function modelsBody(): string {
  return JSON.stringify({
    object: 'list',
    data: [{ id: ALIASED_ID }, { id: FRESH_ID }],
  })
}

/** The rows the gateway stored, as "<providerAlias>|<id>|llm" keys. */
async function customModelKeys(): Promise<string[]> {
  const res = await page.request.get(`${app.baseURL}/api/models/custom`)
  return parseCustomModelsResponse(await res.json())
    .filter((m) => typeof m.providerAlias === 'string')
    .map((m) => `${m.providerAlias}|${m.id}|llm`)
}

beforeAll(async () => {
  const server = Bun.serve({
    port: UPSTREAM_PORT,
    fetch(req) {
      if (new URL(req.url).pathname === '/v1/models') {
        return new Response(modelsBody(), { headers: { 'Content-Type': 'application/json' } })
      }
      return new Response('not found', { status: 404 })
    },
  })
  stopUpstream = () => {
    void server.stop(true)
  }

  app = await startDashboard(PORT)
  page = app.page
}, 90_000)

afterAll(async () => {
  await app.stop()
  stopUpstream?.()
})

describe('Import from /models with a duplicate name', () => {
  test('imports the ids after a refused one and names the refusal', async () => {
    const node = await page.request.post(`${app.baseURL}/api/provider-nodes`, {
      data: {
        name: 'E2E Import Conflict',
        prefix: 'e2econflict',
        type: 'openai-compatible',
        apiType: 'chat',
        baseUrl: `http://127.0.0.1:${UPSTREAM_PORT}/v1`,
        urlSuffix: NODE_SUFFIX,
      },
    })
    expect(node.ok()).toBe(true)

    const conn = await page.request.post(`${app.baseURL}/api/connections`, {
      data: {
        provider: NODE_ID,
        name: 'e2e-import-conflict',
        apiKey: 'e2e-dummy-key',
        authType: 'apiKey',
        isActive: 1,
      },
    })
    expect(conn.ok()).toBe(true)

    // The occupant: a bare-name alias for the id the upstream list opens with.
    const aliased = await page.request.put(`${app.baseURL}/api/models/alias`, {
      data: { model: 'openai/gpt-4o', alias: ALIASED_ID },
    })
    expect(aliased.ok()).toBe(true)

    await page.goto(`${app.baseURL}/dashboard/providers/${NODE_ID}`)

    const dialogOpened = page.waitForEvent('dialog', { timeout: 30_000 })
    // The button's accessible name carries its Material Symbols ligature
    // ("download Import from /models"), so an exact name match never hits.
    await page.getByRole('button', { name: /Import from \/models/ }).click()
    const dialog = await dialogOpened
    const message = dialog.message()
    await dialog.accept()

    // The summary names what was refused (imported ids are not listed), so the
    // import itself is asserted against the stored rows below.
    expect(message).toContain(ALIASED_ID)

    const keys = await customModelKeys()
    expect(keys).toContain(`${NODE_ID}|${FRESH_ID}|llm`)
    expect(keys.filter((k) => k.startsWith(`${NODE_ID}|${ALIASED_ID}|`))).toEqual([])
  }, 90_000)
})
