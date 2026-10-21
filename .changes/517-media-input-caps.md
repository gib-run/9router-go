### feat(media): cap the media a single request may carry, and report what it carried

The transport ceiling (`middleware.MediaMaxBodySize`, 128MB) protects the
process. It does not protect the bill: a request that passes it is still
dispatched upstream, and the upstream call is what gets billed. Every media
lane accepted that body without measuring it, so a caller could send an
arbitrary number of media items of an arbitrary size and the cost was the
provider's problem.

- **Caps, refused before any upstream call.** At most 8 media items per
  request, at most 25MB per item (OpenAI's own ceiling for vision and audio
  input), and at most 64MB of media in total. The total sits below the 128MB
  transport ceiling on purpose: the transport protects the process, the caps
  protect the bill. The refusal is a 413 whose message names the cap and the
  actual figure — a caller told only "too large" cannot tell which limit it
  hit or by how much.
- **The measurement counts what the lanes accept, not what a filename
  suggests.** A multipart part is media when it carries a filename *or* when
  its form field is one the media lanes read binary input from (`file`,
  `image`, `audio`, `video`, `input`, `source`). `antigravity_stt` selects its
  audio by `FormName() == "file"` and reads `FileName()` only for a label, so
  counting by filename alone let that audio through every cap. In a JSON body,
  a media item is a `data:` URI anywhere in the tree, or raw base64 in the
  `image`, `images`, `mask` and `reference` fields, which the image lanes wrap
  into a data URL themselves. The base64 sniff requires a plausible payload —
  long enough to matter and free of whitespace and punctuation base64 never
  contains — so a prompt that merely mentions an image is not counted as one.
  A large text body is never refused by a media cap.
- **The measurement is reported back to the caller.** `X-Media-Items` and
  `X-Media-Bytes`, the same shape `X-Token-Quota-*` already uses.
- **Those two headers are gateway-owned.** Upstream response headers now go
  through `copyUpstreamHeaders`, which drops them if the provider emits them:
  a provider that emits `X-Media-Items` could otherwise replace the number
  the gateway measured with its own — the usage the gateway claims to own was
  writable by the party it measures. Unrelated upstream headers are still
  relayed.
- **The guard sits at the one funnel every media lane passes through**
  (`forwardMediaRequest`), before model resolution, access control and
  connection selection — an oversized request must never reach an upstream
  call. The two lanes that leave that funnel on their own path,
  `/v1/systemone` and the Xiaomi MiMo speech lane, are guarded at their own
  entry points.

Verified with `internal/handlers/media/media_input_cap_test.go` and
`internal/handlers/media/media_cap_verify_gaps_test.go`: the item-count cap,
the per-item cap, the total-size cap, the filename-less media part, the bare
base64 image field, the prose control, the upstream header overwrite, and the
two lanes that bypass the funnel. The refusal tests count the upstream
requests that actually arrived and require zero — the cap has to fire before
the call, not after it.
