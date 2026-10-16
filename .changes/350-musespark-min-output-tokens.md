### Fixes

- **executor**: clamp output token budget (`max_output_tokens` / `max_tokens`) to a minimum floor of 512 for Muse Spark Responses API models and parse reasoning effort suffixes (e.g. `(high)`), preventing empty completions and 502 stream stalls caused by hidden reasoning token starvation (fixes #349).
