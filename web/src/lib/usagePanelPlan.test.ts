import assert from 'node:assert/strict'
import { test } from 'node:test'
import {
  USAGE_PANEL_DEFAULT_ORDER,
  blocksFor,
  moveBefore,
  moveInOrder,
  normalizeOrder,
} from './usagePanelPlan'
import { USAGE_PANEL_IDS, USAGE_PANEL_PRESETS } from './usagePanel'
import type { UsagePanelId } from './usagePanel'

// The arrangement decides what the Usage page draws and in which order, and the
// two ways it can go wrong are both silent: an unknown id in storage hides a
// panel from its section for good, and a mis-paired row puts the topology map and
// the Recent Requests list in the wrong half of the grid.

test('normalizeOrder appends panels this build knows that storage predates', () => {
  const stored = ['overview.breakdown', 'overview.kpi']
  assert.deepEqual(normalizeOrder('overview', stored), [
    'overview.breakdown',
    'overview.kpi',
    'overview.topology',
    'overview.recent',
    'overview.trend',
    'overview.comboFlow',
  ])
})

test('the Default preset hides the combo flow and the three Details-only cards', () => {
  const preset = USAGE_PANEL_PRESETS.find((p) => p.id === 'default')
  assert.ok(preset, 'the default preset must exist')
  const hidden = USAGE_PANEL_IDS.filter((id) => !preset.visible.includes(id))
  assert.deepEqual(hidden, [
    'overview.comboFlow',
    'details.lifetime',
    'details.heatmap',
    'details.nodeHealth',
  ])
})

test('normalizeOrder drops unknown and duplicated ids', () => {
  const stored = ['overview.kpi', 'overview.gone', 'overview.kpi']
  const order = normalizeOrder('overview', stored)
  assert.equal(order[0], 'overview.kpi')
  assert.equal(order.length, USAGE_PANEL_DEFAULT_ORDER.overview.length)
})

test('normalizeOrder keeps a group to its own panels', () => {
  const order = normalizeOrder('details', ['cache.entries', 'details.nodeHealth'])
  assert.deepEqual(order.slice(0, 4), [
    'details.nodeHealth',
    'details.lifetime',
    'details.heatmap',
    'details.requests',
  ])
})

test('moveInOrder swaps one place and refuses to leave the list', () => {
  const order: UsagePanelId[] = ['overview.kpi', 'overview.trend', 'overview.breakdown']
  assert.deepEqual(moveInOrder(order, 'overview.trend', -1), [
    'overview.trend',
    'overview.kpi',
    'overview.breakdown',
  ])
  assert.deepEqual(moveInOrder(order, 'overview.trend', 1), [
    'overview.kpi',
    'overview.breakdown',
    'overview.trend',
  ])
  // Past an end: unchanged, which is what a disabled arrow button relies on.
  assert.deepEqual(moveInOrder(order, 'overview.kpi', -1), order)
  assert.deepEqual(moveInOrder(order, 'overview.breakdown', 1), order)
  assert.deepEqual(moveInOrder(order, 'overview.missing' as UsagePanelId, 1), order)
})

test('moveBefore drops a panel into the slot above or below its target', () => {
  const order: UsagePanelId[] = [
    'overview.kpi',
    'overview.comboFlow',
    'overview.trend',
    'overview.breakdown',
  ]
  assert.deepEqual(moveBefore(order, 'overview.breakdown', 'overview.kpi', true), [
    'overview.breakdown',
    'overview.kpi',
    'overview.comboFlow',
    'overview.trend',
  ])
  assert.deepEqual(moveBefore(order, 'overview.kpi', 'overview.trend', false), [
    'overview.comboFlow',
    'overview.trend',
    'overview.kpi',
    'overview.breakdown',
  ])
})

test('moveBefore leaves the order alone when the drop changes nothing', () => {
  const order: UsagePanelId[] = ['overview.kpi', 'overview.trend', 'overview.breakdown']
  // Onto itself, and into the gap the panel already sits in.
  assert.deepEqual(moveBefore(order, 'overview.trend', 'overview.trend', true), order)
  assert.deepEqual(moveBefore(order, 'overview.trend', 'overview.kpi', false), order)
  assert.deepEqual(moveBefore(order, 'overview.trend', 'overview.breakdown', true), order)
  // An id this build does not know is not a drop target either.
  assert.deepEqual(moveBefore(order, 'overview.gone' as UsagePanelId, 'overview.kpi', true), order)
})

test('blocksFor pairs the two panels that share a row when they stay adjacent', () => {
  const blocks = blocksFor(['overview.kpi', 'overview.topology', 'overview.recent', 'overview.trend'])
  assert.deepEqual(
    blocks.map((b) => b.ids),
    [['overview.kpi'], ['overview.topology', 'overview.recent'], ['overview.trend']]
  )
})

test('blocksFor splits the row once the arrangement separates the pair', () => {
  const blocks = blocksFor(['overview.topology', 'overview.kpi', 'overview.recent'])
  assert.deepEqual(
    blocks.map((b) => b.ids),
    [['overview.topology'], ['overview.kpi'], ['overview.recent']]
  )
})

test('blocksFor leaves a lone row panel as its own full-width block', () => {
  const blocks = blocksFor(['overview.recent', 'overview.kpi'])
  assert.deepEqual(
    blocks.map((b) => b.ids),
    [['overview.recent'], ['overview.kpi']]
  )
})

test('blocksFor keeps single-panel groups one block per panel', () => {
  const blocks = blocksFor(['details.heatmap', 'details.lifetime'])
  assert.deepEqual(
    blocks.map((b) => b.ids),
    [['details.heatmap'], ['details.lifetime']]
  )
})
