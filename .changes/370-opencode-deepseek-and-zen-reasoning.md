### Fixes

- **opencode**: drop `response_format` for DeepSeek models in `opencode-go` to prevent HTTP 400 relay rejection, and inject `reasoning_content` placeholder for `opencode-zen` multi-turn assistant turns (closes #370).
