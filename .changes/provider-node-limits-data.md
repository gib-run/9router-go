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

### feat(chat): enforce a custom provider's own RPM, TPM, concurrency and queue timeout

The columns above are now enforced. A custom provider is one upstream endpoint
that every connection under it shares, so what it tolerates belongs to the node,
not to the caller — a different axis from the per-key limits, which bound what one
caller may spend. The key is still checked first, so a caller that has exhausted
its own allowance never consumes a provider's capacity on the way to being refused.

- The existing limiter serves both scopes through a namespaced key
  (`prov:<node id>`), so there is one sliding window and one FIFO queue rather than
  two implementations free to drift. `AcquireProvider` charges TPM first (not
  queueable: a bucket refills continuously, so a parked request could be admitted
  and immediately refused again) and then RPM/concurrency through the same queue,
  up to the provider's own timeout.
- Enforcement runs inside `tryForwardWithConnection`, once per upstream attempt,
  because the provider is only known after model resolution and because a refusal
  must fail over rather than end the request. The slot is released when the
  attempt ends — including on failover — so one slow pass cannot lock a node out
  of the concurrency it was granted.
- A refusal is a failed candidate: the combo loops try the next one and only the
  last candidate produces a `429`. It is deliberately NOT a cooldown —
  `RecordConnectionScopedError` records `lastError` for the dashboard while the
  account stays in rotation, because a busy endpoint is not a dead credential.
- Debug trail for the provider that was skipped: one structured log line naming
  provider/axis/limit/window count/concurrency/retry-after, the existing
  `router_rate_limit_rejects_total{scope="provider"}` counter, the `lastError` row,
  and an `attempts[]` array in the final 429 body so a client can see which
  providers were tried and why. The wait is reported in the body's `retryAfter`
  field; the response writers pass upstream status and body through and do not copy
  upstream headers, so no transport-level `Retry-After` is set on this path.
- A token charge the queue refuses is refunded in full (that attempt never left),
  and the reservation is reconciled with the turn's real prompt+completion total
  through the same settler the per-key TPM charge uses — an attempt that died
  before metering gets its charge back instead of spending a provider's budget on a
  turn nobody was served.
- One request remembers which providers refused it, so a pass with several
  accounts under the same saturated node waits out the queue timeout once instead
  of once per account; the memory dies with the request, so the next request is
  free to try the same provider again.
- An unreadable policy reads as unlimited: a storage blip must not take every
  custom provider offline.

Verified: 9 limiter tests (window per provider, slot released on failover, queue
timeout names its axis, TPM refusal, refund on queue refusal, settle only corrects
an under-estimate, release refunds an unsettled reservation, snapshot ignores stale
stamps) and 6 request-side tests (no slot for an unlimited provider, refusal is a
429 carrying `attempts`, a refused provider is not waited on twice, the policy is
memoised per request, fail-open on a read error, `lastError` written without a
cooldown), plus `golangci-lint --new-from-rev` clean and `go test ./...` green.

### feat(chat): refuse a custom provider whose contract expired or whose lifetime quota is spent

The two provider-level limits that no queue can absorb — an expired contract and a
spent lifetime token quota — are now enforced on the node itself, and enforced
once: `providerNodeBlocked`/`providerNodeVerdict` in
`internal/handlers/chat/provider_limits.go` is the only place a node's usability
is decided, and both callers ask it. Candidate selection (`connections.go`,
before strategy rotation) skips an unusable node, and dispatch
(`acquireProviderAttempt`) asks the same helper **before** it touches the limiter,
so a node skipped while picking a connection cannot be admitted when the request
is forwarded, or the reverse. The verdict is memoised per request alongside the
policy (`providerLimitState.verdicts`): one limit read and one ledger sum per
node per request, not one per caller.

The refusal's status now follows its axis instead of being 429 for everything:
`contract` and `token_quota` answer **403** with
`provider_contract_expired` / `provider_token_quota_exhausted`, while congestion
stays **429** `rate_limit_exceeded`. No `Retry-After` is sent for a node state —
waiting does not end a contract or give back quota. A spent quota means
`used >= tokenQuota` (99 tokens under a 100 quota stays usable, 100 is refused);
`ResetProviderQuota` makes the node usable again without a restart. When the
ledger cannot be read the node is **failed open**: a storage blip must not drop a
healthy provider out of rotation.

Reason survives the combo path too: both combo selection-error sites capture the
typed refusal and keep it as `lastErr`, instead of replacing the cause with
"all combo models failed".

Verified: 6 new cases in `internal/handlers/chat/provider_limits_test.go` — an
expired contract is refused with 403 `provider_contract_expired` even with no
congestion limits configured, a spent quota is refused at the exact boundary
(3 subtests), a quota reset restores the node, an unreadable ledger fails open,
selection skips an expired node, and an expired node as the only candidate
answers 403 rather than `no active connections`. Test fixtures were moved to the
production schema rather than the production code being loosened:
`dbtest.CreateTables` now also creates the ledger tables (`EnsureLedgerTables`),
and the fixture's `usageHistory` carries the `id INTEGER PRIMARY KEY AUTOINCREMENT`
the quota reset mark reads.

It also hardens `TestQueueIsFIFO`: it established enqueue order with a 5 ms sleep
stagger, and a loaded machine broke that: goroutine 4 reached `Acquire` before
2 and 3, so the recorder saw `0,1,4,2,3,…` and reported a FIFO violation the
queue had not made. Each waiter now starts only once the previous one is parked
(`waitParked`, white-box on the queue depth), so the test asserts the queue's
fairness instead of the scheduler's timing: 50 consecutive runs pass, including
20 with six CPU burners saturating the machine.

### feat(dashboard): a custom provider node's limits, its spend, and the live window

The policy the data plane enforces is now editable and observable from the
dashboard. `PUT /api/provider-nodes/{id}/limits` merges rather than replaces —
an omitted field keeps its stored value, an explicit `0` clears it, a number
sets it — with the same tri-state `mergeLimit` the per-key editor uses, so an
operator does not have to learn a second one. A negative count and an
unparseable contract timestamp answer 400, an id that names no node answers 404
(a silent success would report a policy nobody can find again), and
`POST /api/provider-nodes/{id}/quota/reset` moves the measurement mark without
touching the history the analytics reads. `GET /api/provider-nodes` now carries a
`limits` object per node — policy, `usedTokens`, `rpmUsedWindow`,
`activeConcurrency`, `expired` — filled from two queries for the whole list
(`ListProviderLimits` plus a new `ProviderTokenUsageByNode`, which unions the
carryover so a node whose rows a prune already deleted still reports its spend),
and from the same limiter the gateway meters with, so a dashboard figure and a
refusal cannot disagree.

Three gaps surfaced while writing the integration test, all of which the unit
suite was structurally unable to see:

1. **`SetRateLimiter` had no production caller.** The provider gate returns early
   when the handler holds no limiter, so the whole per-node feature was inert on
   a real boot while every unit test — which sets that field by hand — stayed
   green. `SetupServerRouter` now hands the limiter over right after building it.
2. **`resolvePrefixProvider` swallowed the node's refusal.** With an expired
   contract the lookup fails, resolution fell through to the bare prefix, and the
   client read `502 no active connections for provider: <prefix>` instead of the
   typed 403 that names the cause. A node that is merely unusable now stays
   pinned; a prefix that is also a built-in alias keeps its alias path, so an
   unusable node cannot silently take requests meant for another provider.
3. **The two write routes were mounted only in `dashboard.RegisterRoutes`**, which
   nothing but the package's own tests calls — the binary mounts the dashboard API
   from `SetupDashboardRoutes`. Both tests passed and the live server answered a
   plain `404 page not found`. The routes are registered in both tables now, and
   `TestProviderNodeLimitsAreEditableThroughTheRouter` drives them through the real
   router so the next one added in a single place fails a test instead of a user.

DB export/import carries the seven node limit columns and `quotaResetId`: a
backup that quietly lifted a node's ceiling would turn disaster recovery into an
accidental opening of the throttle, and restoring spend without its mark would
charge the node for tokens it had been forgiven.

Verified: 4 new handler cases (`provider_node_limits_test.go` — merge leaves other
fields alone, negative/malformed/unparseable answers 400 without writing, an
unknown node answers 404, a reset clears spend while keeping the budget and the
history, and the list reports policy, spend and live window/concurrency from a
limiter with one request in flight), 4 integration cases through the real router
and a real SQLite file (an RPM ceiling answers 429 and the refused attempt never
reaches the fake upstream; a spent quota answers 403
`provider_token_quota_exhausted`; an expired contract answers 403
`provider_contract_expired`; and the limits editor saves, lists and resets a
node's policy through the router the binary actually mounts, which is where the
route tables differ), a resolution case pinning that an unusable node still
resolves to the node, and the `usageHistory` fixture in
`provider_nodes_update_test.go` moved to the production column set (it was
missing `promptTokens`/`completionTokens`, so the node list answered 500 rather
than the test failing on the behaviour it was about).
Rename this file to `<pr-number>-provider-node-limits-data.md` when the PR is
opened, so the fragment orders correctly in the dashboard changelog.
