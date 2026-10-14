### Fixes

- **opencode**: migrate free-tier models and Zen payg free-tier routes to `https://opencode.ai/inference/v1/chat/completions` and `https://opencode.ai/inference/v1/responses` following upstream OpenCode Zen's decommissioning of `/zen/*` (fixes #317, upstream decolua/9router#4781, #4837).
- **opencode**: align client fingerprinting with v2 CLI identity contract (`opencode/latest/2.0.29/cli`), send 4-mirror session headers (`x-opencode-session`, `x-opencode-session-id`, `x-session-affinity`, `x-session-id`), and drop conflicting `x-api-key` and `x-opencode-request` on the new inference gate.
- **translator**: provide realistic non-empty parameter schemas on injected canonical tools in `ConcealFingerprintTools` (`command`, `filePath`, `pattern`), preventing upstream 403 `FreeTierError` on tool-less free model requests.
