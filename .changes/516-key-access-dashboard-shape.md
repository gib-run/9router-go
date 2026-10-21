### feat(dashboard): read and write a key's model access in the Node shape

The chat gate already enforced a key's model allow-list, but the dashboard had
no way to see or change it: the listing and the single-key update spoke only
the flat `models` list, while the Node dashboard speaks
`access: {restricted, allow}`. A client that understood only the Node shape
saw an unrestricted key and could not restrict one — the policy the request
path enforced was invisible to the operator who set it.

Both endpoints now carry the `access` object as well, in both directions. The
mapping is not a second store: every write still goes through
`SetAllowedModels` and every read through `GetAllowedModels`, so the chat gate
keeps seeing one source of truth and the dashboard cannot disagree with it. A
separate `restricted` column would be a second place for the policy to
contradict itself.

A null allow-list is reported as unrestricted, matching what the gate
answers when the table is absent, so the listing never claims a restriction
the request path is not enforcing. The listing reads the whole table in one
query rather than one per key, and a read failure falls back to the same
answer — an empty list is a fact about the request, not a failure to read.

An omitted `access` leaves the stored allowlist alone; a present one is
validated before the first write, so a refused body leaves the allowlist
exactly as it was.
