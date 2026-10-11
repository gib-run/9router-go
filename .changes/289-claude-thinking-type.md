### 🐛 `reasoning_effort` to a Claude upstream no longer fails with `thinking.type: Field required` (#289)

- An OpenAI-shaped client sending `reasoning_effort` (or Responses `reasoning.effort`) to a Claude
  upstream got `thinking: {"display": "summarized"}` with no `type`, and Anthropic answered
  `400 thinking.type: Field required`.
- The request translation now sets the missing `type`, mirroring the `claude-adaptive` /
  `claude-budget` branches of upstream `thinkingUnified.js`: `adaptive` + `output_config.effort`
  for adaptive models, `enabled` + `budget_tokens` (clamped below `max_tokens`) otherwise.
  An explicit `thinking.type` from the client is never overwritten.
