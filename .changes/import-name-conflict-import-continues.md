### Fixed

- **"Import from /models" ended at the first duplicate name.** `POST
  /api/models/custom` refuses an id that already addresses a model alias or a
  combo with a typed `409` (`MODEL_ALIAS_CONFLICT` / `COMBO_NAME_CONFLICT`), and
  the import loop awaited every save inside one `try`/`catch`. The first refused
  id therefore threw past the ids listed after it: the run stopped, the alert
  carried the raw 409 body, and every model below the collision was silently
  never imported. A `/models` list only has to carry one already-aliased name
  (a bare-name alias usually points at the same model on another node) for the
  import to lose most of its rows.
- Each row is now attempted on its own and the refusals come back as data.
  `importCustomModelsBestEffort` (new, in `web/src/lib/customModels.ts`) returns
  `{ imported, refused }`, both import paths (`/models` and the live-catalog
  import) report the refused ids with the server's own reason, and the ids that
  follow a collision are still imported.

### Verified

- `bun test src scripts` — 405 pass, including a new
  `web/src/lib/customModels.test.ts` pinning that a refusal costs one row and
  not the batch.
- `web/e2e/modelImportNameConflict.test.ts` drives the reported flow against the
  real gateway: an OpenAI-compatible node whose `/models` lists the aliased id
  first, then a fresh one. Before the fix the same run stored **zero** rows
  (`Received: []`); after it the alert reads `Imported 1 model. 1 model could
  not be registered:` and the fresh id is in `/api/models/custom`.
- `bun run build`, `bun run ratchet:svelte` (81 errors, at baseline) and
  `bunx oxlint --deny-warnings` are clean.
