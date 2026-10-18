### feat(db): per-custom-provider limit storage, token quota readers, and a ledger carryover for them

A custom provider (`providerNodes`) is one upstream endpoint shared by every
connection under it, and until now the gateway had no way to say what that
endpoint tolerates. This lands the storage and repository layer only — nothing
enforces yet — for the six limits planned in `docs/provider-limits-plan.md`:
RPM, TPM, concurrency, a lifetime token quota, a queue timeout, and a contract
expiry date.

- Seven Go-only columns on `providerNodes` (`rpmLimit`, `tpmLimit`,
  `concurrencyLimit`, `tokenQuota`, `queueTimeoutMs`, `quotaResetId`,
  `contractExpiresAt`), applied by the existing `EnsureAdditiveColumns` bootstrap,
  so an existing install picks them up on the next start. The dashboard's node
  `data` JSON blob is deliberately left alone: that blob is rewritten by node
  updates and mirrored into every connection's `providerSpecificData`.
- `internal/db/provider_limits.go` reads and writes the policy.
  `GetProviderLimits` answers the zero value for an unknown node, which reads as
  unlimited; `SetProviderLimits` rejects negative counts and an unparseable
  contract timestamp (the column is parsed on every dispatch, and a value that
  never parses would silently mean "no expiry"), normalises the timestamp to
  RFC3339 UTC, and refuses an unknown node. `quotaResetId` is not writable there —
  it belongs to `ResetProviderQuota`, and a policy save that moved it would either
  forgive spend or charge it twice.
- `internal/db/provider_quota.go` mirrors the per-key budget reset (#365): a
  high-water mark on the node row is the `usageHistory` row id the quota is
  measured from, `ProviderTokenUsage` sums the node's ledger rows from that mark
  forward plus any carried spend, and the mark plus the node's carryover are
  written in one transaction so a request landing between them cannot slip onto
  the wrong side of the boundary.
- The ledger prune now carries provider spend too. `usageHistory.provider` holds
  the node id for a custom provider, so a node's quota is derived from rows a
  retention prune deletes: without `providerLedgerCarryover` the prune would have
  silently refunded an exhausted node's quota. The chunk is folded once against
  both the key marks and the node marks, in the same transaction that deletes the
  rows. A fixture without a `providerNodes` table still prunes: the dimension is
  probed and left nil.

Verified: `go test ./... -count=1` passes, including
`TestPruneUsageHistory_ProviderSpendInvariant` (a node's usage is identical before
and after a prune) and `TestPruneUsageHistory_ProviderResetIsNotReBanked` (a reset
that precedes the prune is not charged twice). Formatting checked against LF
content, because a `core.autocrlf=true` worktree makes `gofmt -l` flag every file.

Rename this file to `<pr-number>-provider-node-limits-data.md` when the PR is
opened, so the fragment orders correctly in the dashboard changelog.
