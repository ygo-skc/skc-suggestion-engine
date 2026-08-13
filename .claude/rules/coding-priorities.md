---
paths:
  - "**/*.go"
---

# Coding priorities (read before writing or changing any Go)

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
