### Fixes

- **providers**: add `big-pickle` to `openCodeFreeTierModels` and allow models with `-free` suffix in `IsOpencodeFreeTierModel` on OpenCode providers, enabling uncataloged/preview free models to route to `/inference/v1` on `opencode-zen` connections without 403 (fixes #327).
