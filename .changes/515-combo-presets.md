### feat(combos): serve the preset catalog from the same source the dashboard shows

A combo is a named failover chain, and building one by hand means picking the
models and the order. The dashboard already had a catalog of the chains that
make sense for a given source, but the API had no way to read it — the preset
list and the combo list were two different things, and only one of them could
be created from the other.

`GET /api/combos/presets?source=<source>` now returns the catalog, and
`POST /api/combos/presets` creates the combo the selected preset describes.
The preset list comes from the same catalog the dashboard renders, so a name
the operator clicked is the name that is created — the two cannot drift.

An unknown or empty `source` is refused rather than answered with an empty
list, so a caller that asked the wrong question is told so instead of being
handed a success it has to interpret.
