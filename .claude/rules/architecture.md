---
paths:
  - "**/*.go"
---

# Architecture: key patterns

- **DAO injection:** handlers call the package-level `skcSuggestionEngineDBInterface db.SKCSuggestionEngineDAO`
  var (`api/server.go`). Tests swap it for a mock from `testing/`. Add new DB operations to the
  `SKCSuggestionEngineDAO` interface, the `...Implementation`, and the mock together.
- **Downstream clients are package-level vars:** `downstream.YGO` (gRPC) is set at startup. `suggest`
  and handlers call it directly. Voyage is a plain HTTP client with a generic `doVoyageRequest[T]` helper.
- **Two Mongo connections, one cluster** (`db/connect.go`): a general connection with `ReadConcern:
  Available` (eventually-consistent reads) and a separate `vectorSearch` connection with `ReadConcern:
  Local` — required because `$vectorSearch` won't run under `Available`. Indexes and Atlas Search indexes
  are created idempotently at startup.
- **Shared common lib:** `github.com/ygo-skc/skc-go/common/v3` (imported as `cModel`/`cUtil`) provides
  models (`APIError`, `YGOCard`, `CardDataMap`, `BatchCardIDs`), env loading, request-scoped logging
  (`cUtil.InitRequest` / `RetrieveLogger` — pull the logger from `ctx`, don't use `slog` directly in
  request paths), and concurrency helpers (`AtomicWaitGroup`). Treat it as a vendored dependency; don't
  inspect or modify its internals — use the exported API and rely on `go build`/`go vet`.
- **Errors:** downstream/DB layers return `*cModel.APIError`; handlers surface it via
  `err.HandleServerResponse(res)`.
- **Concurrency:** fan-out work uses `cUtil.AtomicWaitGroup[T]` (see `suggest.FetchMetadata`) or
  `sync.WaitGroup` + goroutines, loading results after kickoff.
