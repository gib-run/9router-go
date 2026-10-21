### Hide the MMF provider from the catalog and refuse cloud metadata for local callers

- **A provider that exists to serve traffic was being offered as one to pick.**
  `mimo-free` is a no-auth registry entry, and the no-auth rule publishes every
  such entry into `/v1/models` so a caller can use the free tier without
  connecting anything. That rule is right for providers a person chooses between
  and wrong for one the platform runs on its own: `mimo-free` was showing up in
  the catalog a person browses, and so was its `mmf` alias, as two separate
  entries for the same provider. `ProviderConfig` now carries a `Hidden` flag,
  `mimo-free` sets it, and the catalog skips hidden providers — so the entry
  stays fully reachable while it stops being advertised.
- **Hidden is a listing concern, not a registration one.** A stored connection
  pointing at `mimo-free` still resolves and still routes; the provider keeps
  its published alias. The catalog filter runs on the canonical id, so every
  spelling of the same provider leaves the catalog together — an alias-only
  check would let one spelling through while blocking the other.
- **A local caller could point a provider node at the cloud metadata service.**
  `HandleValidateProviderNode` skips the SSRF gate for loopback peers, so a
  browser on the same machine could hand the gateway `169.254.169.254` and have
  it dial the instance-metadata service, which answers with instance
  credentials to anything that asks from inside the VM. The metadata guard now
  runs *before* the loopback bypass rather than inside it, so the function reads
  as one rule with one exception instead of an exception inside an exception.
- **Only the metadata addresses are refused.** `169.254.0.0/16`, `fe80::/10`,
  `100.100.100.200` (Alibaba), `192.0.0.192` (Oracle) and `fd00:ec2::254` are
  refused for every caller, local or not. Loopback and RFC1918 stay the
  caller's decision, because a self-hosted provider is exactly a request to
  reach those on purpose: a local caller keeps its `127.0.0.1:11434` node and
  its LAN nodes, and a remote caller is still refused both.

Evidence: `internal/handlers/dashboard/provider_nodes_g23b_test.go` and
`internal/handlers/chat/models_hidden_test.go`. The catalog test carries a
control — it requires a visible no-auth provider to be present — because an
earlier version asserted an absence under a listing mode that never reaches the
publishing pass, and so passed for the wrong reason.
