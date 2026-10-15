### Fixes

- **combo**: advance to the next model in a fallback combo when a member returns a request-scoped context overflow error (e.g. `context_length_exceeded`, `maximum context length`, `prompt is too long`) without triggering connection cooldown, allowing subsequent models with larger context windows to serve the request (fixes #338, porting decolua/9router#4820).
