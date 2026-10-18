import type { ActiveTab } from './router'

/**
 * Single source of truth for the dashboard sidebar.
 *
 * The sidebar renders NAV_GROUPS and the Feature Visibility panel in Settings
 * writes the hide list for exactly the same entries. A hand-written nav entry in
 * one place and a toggle list in the other is the phantom-config-path failure
 * mode: the panel saves an id that no sidebar entry ever consumes, the save
 * reports success, and nothing changes.
 */

/** Where a dashboard feature came from, so the panel can name its provenance. */
export interface NavProvenance {
  /** What the feature is, in one line. */
  summary: string
  /** Branch that carried or carries the feature; absent when there is none yet. */
  branch?: string
  /** Port-plan section, for a feature that is specified but not implemented. */
  plan?: string
  /** Pull requests (PR) that introduced or extended it. */
  prs: number[]
  /** Issues it closes or belongs to. */
  issues: number[]
  /** Repository slug the branch, PRs and issues live in. */
  repo: string
  status: 'in-main' | 'awaiting-merge' | 'not-ported'
}

export interface NavLink {
  kind: 'link'
  id: ActiveTab
  label: string
  icon: string
  provenance?: NavProvenance
}

export interface NavAccordion {
  kind: 'accordion'
  /** The id the hide list stores for the whole group. */
  id: string
  label: string
  icon: string
  children: NavLink[]
  provenance?: NavProvenance
}

export type NavEntry = NavLink | NavAccordion

export interface NavGroup {
  id: string
  /** Absent for the top group, which renders without a heading. */
  label?: string
  entries: NavEntry[]
}

/**
 * Resale metadata on an API key (F-14 in docs/keirouter-port-plan.md) already
 * ships in the Endpoint & Key page: `expiry`/`usedCount` and the per-key
 * `rateLimitRPM/TPM/concurrency` columns.
 */
const RESALE_KEY_METADATA: NavProvenance = {
  summary:
    'Per-key resale metadata: expiry, usedCount and rateLimitRPM/TPM/concurrency on Endpoint & Key.',
  branch: 'main',
  prs: [185, 316, 342],
  issues: [176],
  repo: 'luqman-v1/9router-go',
  status: 'in-main',
}

/**
 * Plans (F-3 in docs/keirouter-port-plan.md): reusable policy templates every
 * API key inherits from. Merged on main as PR #354 from the branch below.
 */
const PLANS: NavProvenance = {
  summary: 'Reusable policy templates (rate limits, budgets, model access) an API key inherits.',
  branch: 'feat/plans-policy-resolver',
  prs: [354],
  issues: [176],
  repo: 'luqman-v1/9router-go',
  status: 'in-main',
}

/** Sidebar entries the panel must never hide: the panel lives behind Settings. */
export const ALWAYS_VISIBLE_NAV_IDS: readonly string[] = ['settings']

/** Mirrors the server's cap in internal/handlers/dashboard/navvisibility.go. */
const MAX_HIDDEN_NAV_ENTRIES = 32
const NAV_ID_PATTERN = /^[a-z0-9-]{1,32}$/

export const NAV_GROUPS: NavGroup[] = [
  {
    id: 'main',
    entries: [
      { kind: 'link', id: 'endpoint', label: 'Endpoint & Key', icon: 'api', provenance: RESALE_KEY_METADATA },
      { kind: 'link', id: 'connections', label: 'Providers', icon: 'dns' },
      { kind: 'link', id: 'combos', label: 'Combo & Vision Adapter', icon: 'layers' },
      { kind: 'link', id: 'analytics', label: 'Usage', icon: 'bar_chart' },
      { kind: 'link', id: 'quota', label: 'Quota Tracker', icon: 'data_usage' },
      { kind: 'link', id: 'token-saver', label: 'Token Saver', icon: 'savings' },
      { kind: 'link', id: 'cli-tools', label: 'CLI Tools', icon: 'terminal' },
    ],
  },
  {
    id: 'system',
    label: 'System',
    entries: [
      {
        kind: 'accordion',
        id: 'media-providers',
        label: 'Media Providers',
        icon: 'perm_media',
        children: [
          { kind: 'link', id: 'media-embedding', label: 'Embedding', icon: 'data_array' },
          { kind: 'link', id: 'media-image', label: 'Text to Image', icon: 'brush' },
          { kind: 'link', id: 'media-tts', label: 'Text To Speech', icon: 'record_voice_over' },
          { kind: 'link', id: 'media-stt', label: 'Speech To Text', icon: 'mic' },
          { kind: 'link', id: 'media-video', label: 'Video', icon: 'movie' },
          { kind: 'link', id: 'media-systemone', label: 'System One', icon: 'psychology' },
          { kind: 'link', id: 'media-web', label: 'Web Fetch & Search', icon: 'travel_explore' },
        ],
      },
      { kind: 'link', id: 'proxy-pools', label: 'Proxy Pools', icon: 'lan' },
      { kind: 'link', id: 'skills', label: 'Skills', icon: 'extension' },
      { kind: 'link', id: 'console-log', label: 'Console Log', icon: 'terminal' },
      { kind: 'link', id: 'settings', label: 'Settings', icon: 'settings' },
      { kind: 'link', id: 'security', label: 'Security', icon: 'shield_lock' },
      { kind: 'link', id: 'plans', label: 'Plans', icon: 'workspace_premium', provenance: PLANS },
    ],
  },
]

/**
 * Features from the KeiRouter port that have no sidebar entry to toggle, so the
 * panel can only name where they live. An operator hiding "the selling system"
 * has to be able to see which branch carries that code and which PR or issue
 * tracks it.
 */
export interface UpcomingNavFeature {
  id: string
  label: string
  provenance: NavProvenance
}

export const NOT_ON_MAIN_NAV_FEATURES: UpcomingNavFeature[] = [
  {
    id: 'usage-portal',
    label: 'Public usage portal (selling system)',
    provenance: {
      summary:
        'The page a key buyer logs into to read their own usage (F-8). Not implemented, so it has no branch.',
      plan: 'docs/keirouter-port-plan.md §F-8',
      prs: [],
      issues: [176],
      repo: 'luqman-v1/9router-go',
      status: 'not-ported',
    },
  },
]

/**
 * The provenance note the panel renders under an entry: the branch and the PRs
 * when the feature shipped, the port-plan section when it is only specified.
 * Every note names a branch or a plan plus the PRs or issues that track it,
 * which is what the panel has to state.
 */
export function provenanceNote(provenance: NavProvenance): string {
  const refs: string[] = []
  if (provenance.prs.length > 0) refs.push(`PR ${provenance.prs.map((n) => `#${n}`).join(', ')}`)
  if (provenance.issues.length > 0) {
    refs.push(`issue ${provenance.issues.map((n) => `#${n}`).join(', ')}`)
  }
  const where = refs.length > 0 ? ` — ${refs.join(', ')} in ${provenance.repo}` : ` in ${provenance.repo}`
  const origin = provenance.plan
    ? `Specified in ${provenance.plan}`
    : provenance.branch
      ? `Carried by branch ${provenance.branch}`
      : 'No branch or plan recorded'
  return `${origin}${where}.`
}

/** Every entry id in render order, accordions before their children. */
export const NAV_ORDER: string[] = NAV_GROUPS.flatMap((group) =>
  group.entries.flatMap((entry) =>
    entry.kind === 'accordion' ? [entry.id, ...entry.children.map((child) => child.id)] : [entry.id]
  )
)

/** Which sidebar ids govern a tab: its own entry, plus the entry it aliases. */
const TAB_OWNERS: Record<string, string[]> = buildTabOwners()

function buildTabOwners(): Record<string, string[]> {
  const owners: Record<string, string[]> = {}
  const own = (tab: string, id: string) => {
    ;(owners[tab] ??= []).push(id)
  }
  for (const group of NAV_GROUPS) {
    for (const entry of group.entries) {
      if (entry.kind === 'link') {
        own(entry.id, entry.id)
        continue
      }
      own(entry.id, entry.id)
      for (const child of entry.children) {
        own(child.id, child.id)
        // Hiding the accordion hides every child, so the child's tab is gone too.
        own(child.id, entry.id)
      }
    }
  }
  // Aliases the sidebar keeps lit for another path: /dashboard/keys is the
  // legacy Endpoint & Key route, Terminal lives inside Console Log, and the
  // Cache/Compression sections live inside Usage.
  own('keys', 'endpoint')
  own('terminal', 'console-log')
  own('usage-cache', 'analytics')
  own('usage-compression', 'analytics')
  return owners
}

/**
 * Coerces whatever the server stored into a usable hide list, mirroring
 * internal/handlers/dashboard/navvisibility.go. The server only checks the
 * shape; the sidebar entry list lives here, so an id that addresses no entry is
 * dropped rather than stored as dead configuration.
 *
 * Accepts both the stored `{ hidden: [...] }` object and a bare array, because
 * every helper below hands its own working list back in.
 */
export function normalizeHiddenNav(value: unknown): string[] {
  const raw = Array.isArray(value) ? value : (value as { hidden?: unknown } | null | undefined)?.hidden
  if (!Array.isArray(raw)) return []
  const hidden: string[] = []
  for (const item of raw) {
    if (typeof item !== 'string') continue
    const id = item.trim().toLowerCase()
    if (!NAV_ID_PATTERN.test(id)) continue
    if (!NAV_ORDER.includes(id)) continue
    if (ALWAYS_VISIBLE_NAV_IDS.includes(id)) continue
    if (hidden.includes(id)) continue
    if (hidden.length >= MAX_HIDDEN_NAV_ENTRIES) break
    hidden.push(id)
  }
  return hidden
}

/** Store the list in render order, so the saved JSON stays stable and readable. */
export function orderHiddenNav(value: unknown): string[] {
  const hidden = normalizeHiddenNav(value)
  return NAV_ORDER.filter((id) => hidden.includes(id))
}

/** The next hide list after the operator flipped one entry's switch. */
export function setNavVisible(hidden: unknown, id: string, visible: boolean): string[] {
  const current = normalizeHiddenNav(hidden)
  if (visible || ALWAYS_VISIBLE_NAV_IDS.includes(id)) {
    return orderHiddenNav(current.filter((entry) => entry !== id))
  }
  // normalizeHiddenNav drops an id that addresses no entry, so this cannot
  // store the panel's own id or a typo.
  return orderHiddenNav([...current, id])
}

/** The sidebar with the hidden entries removed; empty groups disappear too. */
export function visibleNavGroups(hidden: unknown): NavGroup[] {
  const ids = normalizeHiddenNav(hidden)
  const groups: NavGroup[] = []
  for (const group of NAV_GROUPS) {
    const entries: NavEntry[] = []
    for (const entry of group.entries) {
      if (ids.includes(entry.id)) continue
      if (entry.kind === 'accordion') {
        const children = entry.children.filter((child) => !ids.includes(child.id))
        // An accordion whose children are all hidden would render as a header
        // that opens onto nothing, so it is dropped with them.
        if (children.length === 0) continue
        entries.push({ ...entry, children })
        continue
      }
      entries.push(entry)
    }
    if (entries.length > 0) groups.push({ ...group, entries })
  }
  return groups
}

/** Is the tab the operator is looking at hidden? */
export function isNavHidden(tab: ActiveTab, hidden: unknown): boolean {
  const ids = normalizeHiddenNav(hidden)
  return (TAB_OWNERS[tab] ?? []).some((owner) => ids.includes(owner))
}

/**
 * The sidebar entry that lights up for a tab. Endpoint & Key stays lit for the
 * legacy /dashboard/keys path, Console Log for Terminal, and Usage for the
 * Cache and Compression sections.
 */
export function isNavLinkActive(id: ActiveTab, activeTab: ActiveTab): boolean {
  if (id === 'endpoint') return activeTab === 'endpoint' || activeTab === 'keys'
  if (id === 'console-log') return activeTab === 'console-log' || activeTab === 'terminal'
  if (id === 'analytics') {
    return (
      activeTab === 'analytics' || activeTab === 'usage-cache' || activeTab === 'usage-compression'
    )
  }
  return activeTab === id
}
