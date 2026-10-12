### Added
- **SumoPod provider**: `sumopod` registered as an API-key provider (bearer transport, OpenAI-compatible) pointing at `ai.sumopod.com/v1/chat/completions` with the `sp` short alias and a curated 50-model catalogue (Claude, Gemini, GPT, DeepSeek, Qwen, GLM, Kimi, MiniMax, Seed, Mimo). The live `/v1/models` requires a token, so the dashboard import uses the seeded list.
