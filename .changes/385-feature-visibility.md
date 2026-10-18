### Feature Visibility — hide the sidebar entries this deployment does not use

- **What changed.** Settings gains a **Feature Visibility** section: one switch per
  sidebar entry (the seven main links, the Media Providers accordion with each of its
  seven children, Proxy Pools, Skills, Console Log, Security, Plans) that hides the entry
  from the sidebar. Settings itself is never hideable, so the panel can always be
  reopened.
- **Storage.** `settings.navVisibility.hidden` (a list of entry ids) through the existing
  `GET/PUT /api/settings`, so it survives a reload, travels with the database
  export/import, and applies to any browser. `internal/handlers/dashboard/navvisibility.go`
  normalises the payload — lower-cased, de-duplicated, capped at 32 ids, `settings`
  stripped — while `web/src/lib/nav.ts` owns the canonical entry list.
- **Deep links.** A hidden tab is not reachable by its own URL: `App.svelte` redirects to
  Endpoint & Key, alias paths included (`/dashboard/keys`, `/dashboard/terminal`,
  `/dashboard/usage/cache`), and a hidden Media Providers accordion hides every child.
- **One entry list.** The sidebar and the panel render `NAV_GROUPS` from
  `web/src/lib/nav.ts` instead of three hand-written arrays plus a hardcoded Plans link,
  so a toggle cannot address an id the sidebar does not consume (the
  phantom-config-path failure mode).
- **Provenance.** Entries that came from the KeiRouter port carry a note naming the
  branch plus the PRs and the issue: Endpoint & Key (per-key resale metadata, PR #185
  closing issue #176, extended by #316, hardened by #342) and Plans (PR #354 on branch
  `feat/plans-policy-resolver`). Under "Not on main yet", the public usage portal (F-8)
  is listed with the port-plan section it is specified in — it has no branch.
- **Verified.** `go test ./internal/handlers/dashboard/`, `bun test src/lib/nav.test.ts`,
  `bun run build`, the svelte-check ratchet, and a browser run against an isolated
  `DATA_DIR`: toggle → sidebar updates after Save → reload keeps it hidden →
  `/dashboard/quota` redirects to `/dashboard/endpoint`, then hide/show a Media Providers
  child and the whole accordion and confirm restoration.
