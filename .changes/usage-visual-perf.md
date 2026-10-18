### Dashboard: the Usage visuals stop repainting every frame

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

Verified with `bun run build`, `bunx oxlint --deny-warnings` and the
svelte-check ratchet (0 unresolved identifiers, 81 errors = baseline), and
rendered in a real browser against a seeded local gateway: the combo flow card
holds no `<animate>`, no `filter` and no blurred shadow (1→0, 1→0, 6→0), and the
topology map still draws every node, edge and token.

Upstream parity: `decolua/9router#381` is the combo flow card; the header change
follows the house strip in `RequestDetailsTab.svelte`. No API, schema or
`/v1` behaviour is touched.
