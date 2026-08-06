# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A Go HTTP API that extends [SKC API](https://github.com/ygo-skc/skc-api) with Yu-Gi-Oh! card
suggestions: material/reference parsing, support cards, product & archetype suggestions,
semantic/similarity search, card-of-the-day, and traffic/trending analysis. See `README.md`
for the feature list and `SYSTEM_DESIGN.md` for per-endpoint sequence diagrams (kept current —
consult it before changing handler flows).

## Coding priorities (read before writing or changing any Go)

When adding or modifying code in this repo, optimize in this order: **correctness → idiomatic Go →
performance → low memory footprint**. The last two are first-class goals here, not afterthoughts —
this is a hot-path suggestion API doing fan-out downstream calls per request.

**Idiomatic Go**
- Follow Effective Go / Go Code Review Comments: short names in small scopes, `err != nil` handled
  immediately, no needless getters, accept interfaces & return concrete types, keep interfaces small
  and defined at the consumer (as `SKCSuggestionEngineDAO` is).
- Match the surrounding code: package-level dependency vars (not framework DI), request-scoped logger
  pulled from `ctx` via `cUtil.RetrieveLogger` (never bare `slog` in request paths), errors as
  `*cModel.APIError` surfaced through `HandleServerResponse`.
- Prefer the standard library and the shared `common/v3` helpers over new dependencies. Use `go vet ./...`
  and `gofmt` semantics; leave no vet warnings.

**Performance**
- Fan out independent downstream/DB work concurrently with `cUtil.AtomicWaitGroup[T]` or
  `sync.WaitGroup` (see `suggest.FetchMetadata`); keep genuinely dependent/fail-fast calls sequential.
- Batch downstream requests (e.g. `GetCardsByID` with `BatchCardIDs`) instead of calling per-card in a loop.
- Reuse expensive objects rather than reallocating per request — follow the existing `sync.Pool` pattern
  used for gzip writers.
- Do work once: hoist invariants out of loops, avoid redundant downstream/DB round-trips, and don't
  re-parse or re-fetch data already in hand.

**Low memory footprint**
- Preallocate slices and maps with a known capacity (`make([]T, 0, n)` / `make(map[K]V, n)`) when the
  size is predictable from the input.
- Avoid copying large structs/maps (`CardDataMap`, card lists) — pass pointers or indexes, and range
  with an index when the element is large.
- Don't hold whole result sets longer than needed; filter/project early (in the Mongo pipeline where
  possible) so less data crosses the wire and lives in memory.
- Prefer streaming/in-place transforms over building throwaway intermediate slices.

Don't add speculative micro-optimizations that hurt readability, and don't add defensive guards for
failure modes that can't occur in this environment — keep changes idiomatic and measured.

## Commands

| Command | Purpose |
| --- | --- |
| `go run .` | Run locally — serves **HTTPS on port 9000** (needs certs, `.env`, and IP DB, see Setup) |
| `make test` / `go clean -testcache && go test ./...` | Run all tests (cache cleared) |
| `go test ./api -run TestName` | Run a single test |
| `make coverage` | Tests with coverage, opens HTML report |
| `make build` | `go mod tidy` + `go vet ./...` + cross-compile Linux ARM64 static binary |
| `go vet ./...` | Vet (also part of `make build`) |

There is no separate lint step beyond `go vet`. CI (`.github/workflows`) runs unit tests + CodeQL.

## Architecture

Layered, with no framework-based DI — dependencies are package-level variables (some are
interfaces) that tests reassign to mocks.

```
main.go → downstream.ConnectToYGOService() + db.EstablishSKCSuggestionEngineDBConn() + api.RunHttpServer()

api/         chi router, middleware, one file per handler group. Registers v1 (/api/v1/suggestions)
             and v2 (/api/v2/suggestions) routes. HTTPS-only: TLS 1.3 + HTTP/2, gzip via sync.Pool,
             API-Key middleware guards admin routes (traffic-analysis).
suggest/     Business logic: parse card effect text into name tokens, dedupe archetypes/references,
             concurrent metadata fetch (FetchMetadata), sorting.
db/          MongoDB DAO. SKCSuggestionEngineDAO interface + ...Implementation struct.
downstream/  External calls: YGO (gRPC client to ygo-service) and Voyage AI (HTTP: embeddings + rerank).
model/       Shared request/response + persistence structs.
validation/  go-playground/validator wrappers for request payloads.
testing/     Mocks (DAO, ygo-service) + setup.go.
```

### Key patterns

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

### Semantic / similar-card search flow

Both `GET /card/{id}/similar` and `GET /card/search?q=` follow: **Voyage embed (query) → Mongo
`$rankFusion`** (Reciprocal Rank Fusion, k=60, weighting vector 0.65 / BM25 text 0.35; `/similar` adds
shared type/attribute/monster-type metadata boosts, `search` does not) **→ Voyage rerank → hydrate via
`ygo-service.GetCardsByID`**. The RRF constants and boost math are documented inline in `db/skc_suggestion_engine_db.go`
and `db/mongo_pipeline.go`. Embeddings are `voyage-4` @ 512 dims; reranking is `rerank-2.5`.

### Testing notes

- `testing/setup.go` `chdir`s to the repo root in its `init()` so relative paths (`./certs`, `./data`)
  resolve in tests. Tests set `IS_CICD` / run as `*.test` so the IP DB and env-file loading are skipped.
- Mocks live in `testing/` (`skc_suggestion_engine_dao_mocks.go`, `ygo_service_mock.go`, plus card/archetype fixtures).
