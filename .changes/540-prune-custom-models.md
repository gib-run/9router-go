### fix(db): prune custom models when provider node is deleted

- **Deleting a provider node orphaned its custom models.** `DeleteProviderNode`
  removed the node row and its connections, but left behind custom models in
  the `customModels` KV scope. A deleted node therefore continued to surface its
  models in the catalog.
- **Port upstream decolua/9router#3115 (commit a8f63f4d).** `DeleteProviderNode`
  now cleans up associated `customModels` rows atomically inside the same
  transaction as the node and connection rows.
- **Support both key separators.** Uses `splitCustomModelKey` to match both
  pipe-separated (`<id>|<model>|...`) and slash-separated (`<id>/<model>`) keys,
  leaving rows of other aliases or unparseable keys intact.
- **Missing node is a no-op.** Deleting an unknown node id returns without
  modifying any tables.
