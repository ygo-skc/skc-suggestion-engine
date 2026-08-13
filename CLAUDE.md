# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A Go HTTP API that extends [SKC API](https://github.com/ygo-skc/skc-api) with Yu-Gi-Oh! card
suggestions: material/reference parsing, support cards, product & archetype suggestions,
semantic/similarity search, card-of-the-day, and traffic/trending analysis. See `README.md`
for the feature list and `SYSTEM_DESIGN.md` for per-endpoint sequence diagrams (kept current —
consult it before changing handler flows).

Detailed rules — coding priorities, architecture patterns, the semantic-search flow, and testing
notes — live in `.claude/rules/` and load automatically; this file stays a short entry point.

## Coding priorities

Optimize in this order: **correctness → idiomatic Go → performance → low memory footprint**. The
last two are first-class goals here, not afterthoughts — this is a hot-path suggestion API doing
fan-out downstream calls per request. Full checklist (idiomatic Go / performance / memory):
`.claude/rules/coding-priorities.md`.

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

Key patterns (DAO injection, downstream client vars, the two Mongo connections, shared common lib,
error handling, concurrency): `.claude/rules/architecture.md`.

Semantic/similar-card search flow (Voyage embed → `$rankFusion` → rerank → hydrate):
`.claude/rules/semantic-search.md`.

Testing setup notes (`testing/setup.go` chdir behavior, mock locations): `.claude/rules/testing.md`.
