### fix(translator): keep a thought signature across a restart

A thought signature is the provider's proof that the model produced a given
thought. The gateway replays it on the next tool turn, because a provider
that receives a function call without the signature it issued refuses the
call.

The signature was kept in the process only. A restart — a crash, a deploy, a
`service restart` — threw them away, and the next tool turn was refused with
a signature error that no amount of retrying could fix: the proof had never
been written down anywhere.

The store now writes each signature to SQLite and reads it back on the next
turn. The table is created with the rest of the core schema, so an existing
database picks it up on the next start without a migration step. Hydration
happens when the repository is built, and it fails open: a database that
cannot be read leaves the store exactly as process-local as it was before, so
a broken database costs the persistence, not the gateway.

The signature itself is still keyed by the same identity it always was, and
an expired row is treated as absent rather than replayed — a signature the
provider has already forgotten is worse than no signature.
