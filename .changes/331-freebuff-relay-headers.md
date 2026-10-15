### fix(freebuff): forward `x-relay-*` static headers on control-plane requests

- Thread `StaticHeaders` through the freebuff session/run helpers (session request, run start/finish, session release, model switch) and copy only the `x-relay-*` prefix onto those requests via `applyRelayHeaders`, so relay-routed connections no longer bypass the relay on control-plane calls. Freebuff still owns its CLI identity headers; tests pass `nil` to keep single-process behavior.
- Dashboard compatible-model testing: per-model test status/error/latency replaces the single in-flight `compatibleTestId`, new **Test all** button probes with a concurrency ceiling of 2 (each probe consumes account quota), and the import dialog can filter to free-tier models only, mirroring `providers.IsFreeTierModel`.
