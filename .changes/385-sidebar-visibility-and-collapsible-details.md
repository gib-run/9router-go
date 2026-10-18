### Sidebar visibility (Plans) and collapsible usage details

- **One switch, one entry.** Settings gains a **Sidebar** card beside **Display** with a single
  switch: **Plans** (the selling system). It is the only entry the panel offers, so a
  deployment that sells nothing does not carry the page around. The switch is a saved
  setting: it lives in the same form as the rest, so the floating "Unsaved changes — Save"
  reminder covers it.
- **Defaults.** A deployment that never saved a list starts with **Plans** hidden; Security is
  visible like every other entry. Saving an explicit empty list restores everything — only an
  absent or unreadable stored value falls back to the default.
- **Usage detail panels collapse.** In Usage > Details, **Activity Heatmap** and **Node Health**
  start collapsed and open from the chevron in their card header. The starting state is the new
  **Display > Expand usage details** switch (localStorage `detailsExpandedByDefault`, off by
  default), rendered next to Unit abbreviation and applied to both panels.
- **Storage.** `settings.navVisibility.hidden` (a list of entry ids) through the existing
  `GET/PUT /api/settings`, so it survives a reload, travels with the database export/import, and
  applies to any browser. `internal/handlers/dashboard/navvisibility.go` normalises the payload —
  lower-cased, de-duplicated, capped at 32 ids, `settings` stripped — while `web/src/lib/nav.ts`
  owns the canonical entry list and the toggleable set.
- **Deep links.** A hidden tab is not reachable by its own URL: `App.svelte` redirects to
  Endpoint & Key, alias paths included (`/dashboard/keys`, `/dashboard/terminal`,
  `/dashboard/usage/cache`), and a hidden Media Providers accordion hides every child.
- **One entry list.** The sidebar and the panel render entries from `web/src/lib/nav.ts` instead
  of three hand-written arrays plus a hardcoded Plans link, so a switch cannot address an id the
  sidebar does not consume (the phantom-config-path failure mode).
- **Verified.** `go test -count=1 ./internal/handlers/dashboard/`, `bun test src/lib/nav.test.ts
  src/lib/displayPrefs.test.ts` (22 pass), `bun run build`, the svelte-check ratchet (82 errors =
  baseline, 0 unresolved identifiers), and a browser run against an isolated `DATA_DIR`: a fresh
  database hides Plans while Security stays, the Sidebar card sits directly under Display with a
  single switch, and on Usage > Details both cards report `aria-expanded="false"` with no body —
  the Display switch then opens both (expanded, "Last 7 days (≥50 real requests)" rendered).
  `/dashboard/plans` redirects to `/dashboard/endpoint` while Plans is hidden.
