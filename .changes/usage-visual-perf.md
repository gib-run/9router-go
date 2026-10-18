### Dashboard: the Usage visuals stop repainting every frame, and the Usage panels become arrangeable

- **Combo flow card.** The route dot was an SVG `<animate>` on `cx` plus a
  `drop-shadow` filter, running at 60fps for as long as the card was mounted —
  including after the combo run had already finished. It is now a single
  composited CSS hop that renders only while a member is being tried, and the
  blurred glows on the trying/connected rows and order discs are gone (the tint,
  border, icon and label still carry the status).
- **Provider topology map**, the measured hot spot on that page (~95% of its
  main-thread task time while requests are in flight): removed the
  `topo-electric` filter (`feTurbulence` with an animated `baseFrequency` over a
  10000x10000 user-unit region plus `feDisplacementMap`) and every inline
  `drop-shadow` on animated elements; one token per direction instead of three
  forward + four reverse dots, and the four electric sparks are gone — ~5 SMIL
  animations per active node instead of ~25; the router pulse and icon shake
  animate `transform` only, with their blurred shadow animation replaced by a
  static glow.
- **Details panels.** Activity Heatmap and Node Health now use the same header
  strip as Recent Request Details (`bg-surface-2`, `font-headline` title,
  caption, chevron) instead of the 10router `Card title icon padding="md"`
  header, which also left a dangling gap under the title while collapsed.
- **Usage dashboard preferences (Settings).** Every Usage panel — Overview,
  Details, Cache Analytics and Compression Analytics, 22 in all — can be hidden
  individually, and the leading groups reordered by drag or by the arrow
  buttons. Three presets (Advanced / Simple / Super Lite) fill the whole toggle
  set, and "Reset order" restores one group. The card has no Save of its own: an
  edit raises the same bottom-right "Unsaved changes" reminder the rest of the
  settings form uses, and that Save writes the arrangement (a `usagePanelLayout`
  localStorage preference) alongside the server settings. The panels whose
  position is fixed — the paged request list and its payload dialog — say so and
  cannot be dragged.
- **Details period filter.** The Request Details table now takes the same
  `?period` window as the Overview instead of ignoring the picker:
  `GetRequestDetailsPaged` gained a `since` predicate, the handler derives it
  from `period` (`detailsSince`), and the table, its total and its pagination
  follow the window. The heatmap, lifetime totals and node health keep their own
  windows and still label them.

Verified with `bun run build`, `bunx oxlint --deny-warnings`, the svelte-check
ratchet (0 unresolved identifiers, 81 errors = baseline) and
`go test ./internal/db/ ./internal/handlers/`, then rendered in a real browser
against a seeded local gateway: the combo flow card holds no `<animate>`, no
`filter` and no blurred shadow (1→0, 1→0, 6→0), the topology map still draws
every node, edge and token, a dragged panel reorders the list and the page,
storage stays untouched until the reminder's Save is pressed (and then matches
the arranged order exactly), and `period=7d` narrows the request table and its
total.

Upstream parity: `decolua/9router#381` is the combo flow card; the header change
follows the house strip in `RequestDetailsTab.svelte`. No API, schema or `/v1`
behaviour is touched; the only backend change is the additive `since` filter on
the request-detail query.
