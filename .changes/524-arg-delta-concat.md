### fix(chat): concatenate repeated tool-call argument deltas instead of deduplicating them

- The forced-SSE JSON path folds an OpenAI stream back into a single Claude
  message, and it merged the `function.arguments` deltas with a heuristic
  rather than by appending them: a delta was dropped when it equalled the
  whole accumulated value, when it appeared anywhere as a substring of that
  value, or when the accumulated value already ended with it
  (`internal/handlers/chat/forward.go`, `sseToClaudeJSON`). The first case
  loses a genuinely repeated chunk — two identical deltas `"a"`, `"a"` folded
  to `"a"` instead of `"aa"` — and the second loses real content, so an
  argument whose payload repeated an earlier fragment arrived with that
  fragment missing and the tool call's arguments stopped parsing.
- The deltas are now appended as they arrive, which is what the JavaScript
  gateway this path was ported from does
  (`open-sse/handlers/chatCore/sseToJsonHandler.js`: `arguments += delta`).
  A stream is an ordered sequence of partials, not a set of replacements, so
  concatenation is the only merge that can be correct for every stream. The
  sibling `function.name` merge is left alone: a name is not streamed in
  partials the way arguments are.
- `internal/handlers/chat/parity_forced_sse_test.go` pins the behaviour three
  ways: a repeated delta, an escaped-JSON delta whose payload repeats a
  phrase, and two interleaved tool indices each carrying a repeated delta.
- Verified by `go build ./...`, `go vet ./...` and `go test ./... -count=1`,
  which pass except for the pre-existing `internal/browserlogin` failure on
  `TestBrowserNamesCoverTheBrowsersTheFlowCanDrive` (upstream's
  `browserNames()` still answers `brave`/`msedge` where the test asks for
  `brave-browser`/`microsoft-edge`). Replaying the upstream merge over the
  same three fixtures folds `"abc"`/`"abc"` to `"abc"` and `"dup"`/`"dup"` to
  `"dup"`, so the new tests fail against the old logic.
