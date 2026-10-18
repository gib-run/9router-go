/**
 * The arrangement half of the Usage panel preferences: which order the panels of
 * a section are drawn in, and how that order folds into renderable blocks.
 *
 * Pure on purpose — the reactive state (and its localStorage draft) lives in
 * lib/usagePanel.svelte.ts, so this module can be unit-tested without the Svelte
 * compiler, the way lib/period.ts and lib/comboMatch.ts are.
 */
import { USAGE_PANEL_GROUPS, type UsagePanelGroupId, type UsagePanelId } from './usagePanel'

/** The order a section draws its panels in when nothing has been arranged. */
export const USAGE_PANEL_DEFAULT_ORDER: Record<UsagePanelGroupId, UsagePanelId[]> =
  Object.fromEntries(USAGE_PANEL_GROUPS.map((group) => [group.id, group.items.map((item) => item.id)])) as Record<
    UsagePanelGroupId,
    UsagePanelId[]
  >

/**
 * Two panels that share one layout row. In the Overview the provider map and the
 * Recent Requests list sit side by side; they still hide independently, so the
 * arrangement treats them as a row only while they are both visible and adjacent.
 */
export const USAGE_PANEL_ROW_PARTNERS: Partial<Record<UsagePanelId, UsagePanelId>> = {
  'overview.topology': 'overview.recent',
  'overview.recent': 'overview.topology',
}

export interface UsagePanelBlock {
  /** Stable list key: the row's two ids, or the single panel's id. */
  key: string
  /** The panels this block draws, in the order it draws them. */
  ids: UsagePanelId[]
}

/**
 * normalizeOrder repairs a stored order against the ids this build knows: unknown
 * ids (an older or newer build) are dropped, and anything missing is appended in
 * declaration order. Without it, a stored order that predates a panel would hide
 * that panel from the section for good.
 */
export function normalizeOrder(group: UsagePanelGroupId, stored: readonly string[]): UsagePanelId[] {
  const known = new Set<UsagePanelId>(USAGE_PANEL_DEFAULT_ORDER[group])
  const kept: UsagePanelId[] = []
  for (const id of stored) {
    if (known.has(id as UsagePanelId) && !kept.includes(id as UsagePanelId)) kept.push(id as UsagePanelId)
  }
  for (const id of USAGE_PANEL_DEFAULT_ORDER[group]) {
    if (!kept.includes(id)) kept.push(id)
  }
  return kept
}

/**
 * moveInOrder returns the order with `id` swapped one place up (-1) or down (+1).
 * A move past either end is a no-op, which is what the disabled arrow buttons rely
 * on rather than each caller re-checking the bounds.
 */
export function moveInOrder(order: readonly UsagePanelId[], id: UsagePanelId, delta: -1 | 1): UsagePanelId[] {
  const next = [...order]
  const from = next.indexOf(id)
  if (from < 0) return next
  const to = from + delta
  if (to < 0 || to >= next.length) return next
  next[from] = next[to]
  next[to] = id
  return next
}

/**
 * moveBefore returns the order with `id` dropped into `targetId`'s slot — above it
 * when `before`, below it otherwise. A drop that cannot move anything (the panel
 * onto itself, or an id this build does not know) comes back unchanged, which is
 * what lets the drag handler call this without pre-checking the pair.
 */
export function moveBefore(
  order: readonly UsagePanelId[],
  id: UsagePanelId,
  targetId: UsagePanelId,
  before: boolean
): UsagePanelId[] {
  const next = [...order]
  if (id === targetId) return next
  const from = next.indexOf(id)
  if (from < 0 || next.indexOf(targetId) < 0) return next
  next.splice(from, 1)
  const slot = next.indexOf(targetId) + (before ? 0 : 1)
  if (slot === from) return [...order]
  next.splice(slot, 0, id)
  return next
}

/**
 * blocksFor turns one section's ordered, visible panels into the blocks the
 * template draws: a row block for the two panel ids that share a row and ended up
 * next to each other, a panel block for everything else. Ordering is preserved
 * exactly as given, so what the settings list shows is what the page draws.
 */
export function blocksFor(orderedVisible: readonly UsagePanelId[]): UsagePanelBlock[] {
  const blocks: UsagePanelBlock[] = []
  const consumed = new Set<UsagePanelId>()
  for (let i = 0; i < orderedVisible.length; i++) {
    const id = orderedVisible[i]
    if (consumed.has(id)) continue
    const partner = USAGE_PANEL_ROW_PARTNERS[id]
    if (partner && i + 1 < orderedVisible.length && orderedVisible[i + 1] === partner) {
      blocks.push({ key: `${id}+${partner}`, ids: [id, partner] })
      consumed.add(id)
      consumed.add(partner)
      continue
    }
    blocks.push({ key: id, ids: [id] })
    consumed.add(id)
  }
  return blocks
}
