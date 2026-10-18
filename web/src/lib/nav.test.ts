import { describe, expect, it } from 'bun:test'
import { TAB_ROUTES } from './router'
import {
  ALWAYS_VISIBLE_NAV_IDS,
  NAV_GROUPS,
  NAV_ORDER,
  NOT_ON_MAIN_NAV_FEATURES,
  isNavHidden,
  isNavLinkActive,
  normalizeHiddenNav,
  provenanceNote,
  setNavVisible,
  visibleNavGroups,
  type NavEntry,
} from './nav'

const links: Extract<NavEntry, { kind: 'link' }>[] = NAV_GROUPS.flatMap((group) =>
  group.entries.flatMap((entry) =>
    entry.kind === 'accordion' ? entry.children : [entry]
  )
)

const renderedIds = (hidden: unknown): string[] =>
  visibleNavGroups(hidden).flatMap((group) =>
    group.entries.flatMap((entry) =>
      entry.kind === 'accordion' ? [entry.id, ...entry.children.map((child) => child.id)] : [entry.id]
    )
  )

const MEDIA_CHILD_IDS = NAV_GROUPS.flatMap((group) =>
  group.entries.flatMap((entry) => (entry.kind === 'accordion' ? entry.children.map((c) => c.id) : []))
)

describe('nav model', () => {
  it('addresses a route the router knows for every entry', () => {
    // The consumer check: an id with no route is an entry the sidebar would
    // render but the app could not open.
    for (const link of links) {
      expect(TAB_ROUTES[link.id as keyof typeof TAB_ROUTES]).toBeTruthy()
    }
  })

  it('uses every id exactly once', () => {
    expect(new Set(NAV_ORDER).size).toBe(NAV_ORDER.length)
    expect(NAV_ORDER.length).toBeGreaterThan(0)
  })

  it('keeps the panel behind an entry it can never hide', () => {
    expect(ALWAYS_VISIBLE_NAV_IDS).toContain('settings')
    expect(renderedIds({ hidden: ['settings'] })).toContain('settings')
    expect(normalizeHiddenNav({ hidden: ['settings'] })).toEqual([])
    expect(setNavVisible([], 'settings', false)).toEqual([])
  })
})

describe('visibility filtering', () => {
  it('removes only the hidden link', () => {
    const ids = renderedIds({ hidden: ['quota'] })
    expect(ids).not.toContain('quota')
    expect(ids).toContain('endpoint')
    expect(ids).toContain('media-video')
  })

  it('removes an accordion with its children', () => {
    const ids = renderedIds({ hidden: ['media-providers'] })
    expect(ids).not.toContain('media-providers')
    for (const child of MEDIA_CHILD_IDS) {
      expect(ids).not.toContain(child)
    }
  })

  it('drops an accordion whose children are all hidden', () => {
    const ids = renderedIds({ hidden: MEDIA_CHILD_IDS })
    expect(ids).not.toContain('media-providers')
    expect(ids).toContain('proxy-pools')
  })

  it('keeps an accordion that still has one visible child', () => {
    const groups = visibleNavGroups({ hidden: MEDIA_CHILD_IDS.filter((id) => id !== 'media-video') })
    const accordion = groups
      .flatMap((group) => group.entries)
      .find((entry) => entry.kind === 'accordion')
    expect(accordion?.kind).toBe('accordion')
    expect(accordion?.kind === 'accordion' && accordion.children.map((c) => c.id)).toEqual(['media-video'])
  })

  it('ignores ids it cannot address and lower-cases the rest', () => {
    expect(normalizeHiddenNav({ hidden: ['QUOTA', 'not-an-entry', 7, 'quota'] })).toEqual(['quota'])
    expect(normalizeHiddenNav({ hidden: 'quota' })).toEqual([])
    expect(normalizeHiddenNav(null)).toEqual([])
    expect(renderedIds({ hidden: ['not-an-entry'] })).toEqual(NAV_ORDER)
  })
})

describe('stored hide list', () => {
  it('stores ids in render order', () => {
    const first = setNavVisible([], 'security', false)
    expect(first).toEqual(['security'])
    expect(setNavVisible(first, 'quota', false)).toEqual(['quota', 'security'])
    expect(setNavVisible(['quota', 'security'], 'quota', true)).toEqual(['security'])
  })

  it('survives a junk payload without losing a valid entry', () => {
    expect(setNavVisible({ hidden: ['quota', 'bogus'] }, 'media-tts', false)).toEqual([
      'quota',
      'media-tts',
    ])
  })
})

describe('route guard', () => {
  it('follows the aliases the sidebar keeps lit', () => {
    expect(isNavHidden('keys', { hidden: ['endpoint'] })).toBe(true)
    expect(isNavHidden('terminal', { hidden: ['console-log'] })).toBe(true)
    expect(isNavHidden('usage-cache', { hidden: ['analytics'] })).toBe(true)
    expect(isNavHidden('usage-compression', { hidden: ['analytics'] })).toBe(true)
  })

  it('treats a hidden accordion as hiding its children', () => {
    expect(isNavHidden('media-video', { hidden: ['media-providers'] })).toBe(true)
    expect(isNavHidden('media-video', { hidden: ['media-tts'] })).toBe(false)
  })

  it('never blocks the login page or an entry nobody hid', () => {
    expect(isNavHidden('login', { hidden: ['login'] })).toBe(false)
    expect(isNavHidden('quota', { hidden: [] })).toBe(false)
    expect(isNavHidden('quota', null)).toBe(false)
  })
})

describe('active highlighting', () => {
  it('keeps the aliased entries lit', () => {
    expect(isNavLinkActive('endpoint', 'keys')).toBe(true)
    expect(isNavLinkActive('console-log', 'terminal')).toBe(true)
    expect(isNavLinkActive('analytics', 'usage-cache')).toBe(true)
    expect(isNavLinkActive('analytics', 'usage-compression')).toBe(true)
    expect(isNavLinkActive('quota', 'analytics')).toBe(false)
  })
})

describe('feature provenance notes', () => {
  const noted = [
    ...NAV_GROUPS.flatMap((group) =>
      group.entries.filter((entry) => entry.provenance !== undefined)
    ).map((entry) => ({ label: entry.label, provenance: entry.provenance! })),
    ...NOT_ON_MAIN_NAV_FEATURES.map((feature) => ({
      label: feature.label,
      provenance: feature.provenance,
    })),
  ]

  it('names a branch or a plan, plus a PR or issue, for every noted feature', () => {
    expect(noted.length).toBeGreaterThan(0)
    for (const { label, provenance } of noted) {
      expect(provenance.branch ?? provenance.plan, label).toBeTruthy()
      expect(provenance.repo, label).toContain('/')
      expect(provenance.prs.length + provenance.issues.length, label).toBeGreaterThan(0)
      expect(provenance.summary.length, label).toBeGreaterThan(0)
    }
  })

  it('points the selling system at the branch, PR and issue that carry it', () => {
    const plans = links.find((link) => link.id === 'plans')
    expect(plans?.provenance?.status).toBe('in-main')
    expect(plans?.provenance?.branch).toBe('feat/plans-policy-resolver')
    expect(plans?.provenance?.prs).toContain(354)
    expect(plans?.provenance?.issues).toContain(176)

    // The portal half has no branch yet, so its note names the port plan.
    const portal = NOT_ON_MAIN_NAV_FEATURES.find((entry) => entry.id === 'usage-portal')
    expect(portal?.provenance.status).toBe('not-ported')
    expect(portal?.provenance.plan).toContain('F-8')
    expect(portal?.provenance.issues).toContain(176)
  })

  it('points the resale metadata at the merged PRs on main', () => {
    const endpoint = links.find((link) => link.id === 'endpoint')
    expect(endpoint?.provenance?.status).toBe('in-main')
    expect(endpoint?.provenance?.branch).toBe('main')
    expect(endpoint?.provenance?.prs).toEqual([185, 316, 342])
  })

  it('renders a note that names the branch or plan, the PRs and the issue', () => {
    const plans = links.find((link) => link.id === 'plans')
    expect(provenanceNote(plans!.provenance!)).toBe(
      'Carried by branch feat/plans-policy-resolver — PR #354, issue #176 in luqman-v1/9router-go.'
    )
    const portal = NOT_ON_MAIN_NAV_FEATURES[0]
    expect(provenanceNote(portal.provenance)).toContain('docs/keirouter-port-plan.md §F-8')
    expect(provenanceNote(portal.provenance)).toContain('issue #176')
  })
})
