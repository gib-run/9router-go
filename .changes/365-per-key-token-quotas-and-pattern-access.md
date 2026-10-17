### feat(keys): per-key token quotas and pattern-based model access

Ported per-key token limits and wildcard/negation model access patterns from upstream decolua/9router (issue #365).

- **Database**: Added `token_limit` and `used_tokens` columns to `apiKeys` schema. Added `IncrementKeyUsedTokens` and `ResetKeyUsedTokens` repository methods for atomic quota tracking and operator reset.
- **Model Pattern Access**: Implemented `IsModelAllowedForKey` supporting exact matching, wildcard prefixes/infixes/suffixes (`oc/*`, `*claude*`, `google/*`), negations (`!opus`, `-haiku`), and open-except-denied semantics.
- **Quota Enforcement**: Added `CheckApiKeyQuota` in auth middleware returning HTTP 429 (`key_quota_exceeded`) when `usedTokens >= tokenLimit`. Wired token usage counter increments in chat usage logging.
- **Dashboard API**: Added `POST /api/keys/{id}/reset-quota` endpoint and updated key creation/update handlers to accept and return `tokenLimit`.
- **Frontend UI**: Updated `ApiKey` client interface and Svelte UI in `ApiKeysTable.svelte` / `ApiKeyPolicyModal.svelte` / `ApiKeysView.svelte` with quota progress bar, token limit configuration, and reset quota action.

Verified with `go test ./...`, `bun run ratchet:svelte`, and `bun run build`.
