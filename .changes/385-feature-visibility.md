### Feature Visibility — choose which sidebar entries this deployment shows

- **What changed.** Settings gains a **Feature Visibility** section: one switch per
  sidebar entry (the seven main links, the Media Providers accordion with each of its
  seven children, Proxy Pools, Skills, Console Log, Security, Plans). Settings is not
  listed at all — it is the page the panel lives on, so it can never be hidden.
- **Defaults.** A deployment that never saved a list starts with **Security** and
  **Plans** hidden, because most operators open neither. Saving an explicit empty list
  restores every entry: only an absent or unreadable stored value falls back to the
  defaults.
- **Save reminder.** A switch is not saved until the operator presses Save, so flipping
  one raises a floating "Unsaved changes — Save" button for four seconds. It clears on a
  successful save and on unmount.
- **Storage.** `settings.navVisibility.hidden` (a list of entry ids) through the existing
  `GET/PUT /api/settings`, so it survives a reload, travels with the database
  export/import, and applies to any browser.
  `internal/handlers/dashboard/navvisibility.go` normalises the payload — lower-cased,
  de-duplicated, capped at 32 ids, `settings` stripped — while `web/src/lib/nav.ts` owns
  the canonical entry list.
- **Deep links.** A hidden tab is not reachable by its own URL: `App.svelte` redirects to
  Endpoint & Key, alias paths included (`/dashboard/keys`, `/dashboard/terminal`,
  `/dashboard/usage/cache`), and a hidden Media Providers accordion hides every child.
- **One entry list.** The sidebar and the panel render `NAV_GROUPS` from
  `web/src/lib/nav.ts` instead of three hand-written arrays plus a hardcoded Plans link,
  so a toggle cannot address an id the sidebar does not consume (the
  phantom-config-path failure mode).
- **Verified.** `go test ./internal/handlers/dashboard/`, `bun test src/lib/nav.test.ts`,
  `bun run build`, the svelte-check ratchet, and a browser run against an isolated
  `DATA_DIR`: Security and Plans hidden on a fresh database → flip a switch → the save
  reminder appears and disappears → Save → reload keeps the list → `/dashboard/plans`
  redirects to `/dashboard/endpoint` → restore and confirm the sidebar returns.
