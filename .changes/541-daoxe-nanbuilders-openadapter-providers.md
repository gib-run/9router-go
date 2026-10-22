### Added

- **New providers**: DaoXE (`daoxe` / `dx`), NaN.Builders (`nanbuilders` / `nan`), OpenAdapter (`openadapter` / `oad`)
  - All three are OpenAI-compatible free-tier providers
  - Support standard OpenAI models: `gpt-4o`, `gpt-4o-mini`, `gpt-3.5-turbo`
  - Include vision, reasoning, search, and tool capabilities for supported models
  - Registered in provider catalog with correct BaseURLs and auth schemes
  - Frontend metadata added for dashboard display (icons, colors, notices)

- **Provider aliases**: `dx` → `daoxe`, `nan` → `nanbuilders`, `oad` → `openadapter`
- **Registry aliases**: `daoxe` → `dx`, `nanbuilders` → `nan`, `openadapter` → `oad`
- **Model capabilities**: Added per-model capability entries for all three providers in `capabilities.go`
- **Unit tests**: Added `TestNewProvidersRegistry_v060` covering provider registration, alias resolution, RegistryAliases, model catalog, and capabilities

### Verified

- `go build ./...` — compiles cleanly
- `go vet ./...` — no issues
- `go test -count=1 ./internal/providers/...` — all provider tests pass
- `go test -tags=integration -count=1 ./internal/integration/...` — integration tests pass
- `cd web && bun run ratchet:svelte` — 0 unresolved identifiers, 78 errors (baseline maintained)