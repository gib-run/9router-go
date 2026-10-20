### fix(executor): arm the completion watchdog only when a completion is pending

`completionWatchdog.arm` started its flush timer whenever nothing had been
emitted yet, without asking whether the bridge actually owed a terminal event.
On a stream that was still producing content the watchdog could therefore fire
`PendingCompletionFlushTimeout` after the stream began, hand a fabricated
completion to the client mid-stream, and then sit on `emitted` so the real
`finish_reason` was dropped.

`arm` now also checks `w.bridge.CompletionPending()` and refuses to schedule the
timer while the stream is still live. The timer keeps its original purpose: once
the terminal event is owed but the usage trailer has not arrived, it flushes
rather than waiting forever.

Verified with `TestParityWatchdogWaitsForPending`, which feeds a content delta,
waits past the flush timeout, and asserts nothing was completed, then feeds the
`finish_reason` and asserts exactly one completion arrives. It fails on the
pre-fix code.
