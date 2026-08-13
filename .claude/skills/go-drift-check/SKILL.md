---
name: go-drift-check
description: This skill should be used when the user asks to "audit the Go code", "check for idiomatic Go", "review Mongo indexes", "check for missing indexes", "check dependency freshness", "check for deprecated library usage", "check for performance or memory regressions", "run a drift check", "check for drift", or wants a health check of this repo's Go code against this repo's coding-standards rules (CLAUDE.md plus its linked files under `.claude/rules/`): idiomatic Go, usage-driven Mongo indexing, performance, memory footprint, up-to-date libraries. Also trigger for periodic or pre-release code-quality audits of this repository, not for reviewing a single PR's diff (use /code-review for that).
---

# Go Drift Check

Audit this repo's Go code for drift across four dimensions and report findings — do not silently
fix anything unless the user asks. This is a reporting skill: read code, compare it against a
baseline, and list what has drifted with file:line references and a concrete fix suggestion.

## Scope

Default to auditing the whole repo (`api/`, `db/`, `suggest/`, `downstream/`, `model/`,
`validation/`, `main.go`), excluding `testing/` mocks. If the user asks to check only recent work,
scope to `git diff release...HEAD` instead. For a repo this size, direct `Bash`/`Grep`/`Read` use
is fine — no need to spawn a subagent unless the user asks for one.

Run the four checks below, then produce the report in the format at the bottom.

## 1. Idiomatic Go

The "Idiomatic Go" section of `.claude/rules/coding-priorities.md` (linked from CLAUDE.md's Coding
priorities section) is the authoritative checklist for this repo — re-read it before judging
anything, since it is the source of truth and may be updated independently of this skill.

Mechanical baseline:
- `gofmt -l .` — anything listed is drift.
- `go vet ./...` — any warning is drift.

Targeted checks beyond the vet/fmt baseline:
- `grep -rn "slog\." api/ db/ suggest/ downstream/` for bare `slog` calls in request paths —
  `.claude/rules/coding-priorities.md` requires `cUtil.RetrieveLogger(ctx)` there instead.
- Any new/changed method on `SKCSuggestionEngineDAO` (`db/skc_suggestion_engine_db.go`) — confirm it
  exists on `SKCSuggestionEngineDAOImplementation` *and* has a matching entry in
  `testing/skc_suggestion_engine_dao_mocks.go`. A method present on one but not the others is drift.
- Downstream/DB-layer functions should return `*cModel.APIError`, not a bare `error`, at the
  boundary that feeds a handler; handlers should surface it via `err.HandleServerResponse(res)`.
- New concurrent DB/downstream fan-out should follow `cUtil.AtomicWaitGroup[T]` or the
  `sync.WaitGroup` + goroutine pattern already used in `suggest.FetchMetadata` — flag fan-out that's
  written sequentially where the calls are actually independent.
- Per-card downstream loops: `grep -rn "for .*range" -A5 api/ suggest/ db/` and check whether the
  loop body calls a single-item downstream method (e.g. a per-ID `ygo-service` call) where a batch
  call (`GetCardsByID` with `BatchCardIDs`) exists and would work instead.

## 2. Mongo index coverage (usage-driven, not a static list)

Don't rely on a hardcoded index inventory — the schema changes over time. Derive it fresh each run:

1. Read `db/connect.go`'s `createIndexes()` and `createSearchIndexes()` to build the *current*
   index inventory: for each `*mongo.Collection` package var, list its indexed key(s) in order.
   Remember the compound-index prefix rule — an index on `{a:1, b:1}` only serves queries that filter
   or sort on `a` alone, or `a` then `b`; it does not serve a query that filters on `b` alone.
2. For every collection var declared in `db/connect.go` (grep `db/*.go` for `Collection\.`), find
   every `.Find(`, `.FindOne(`, `.CountDocuments(`, `.Aggregate(`, `.UpdateOne(`, `.DeleteOne(` call
   against it and read the filter (`bson.M{...}` / `bson.D{...}`) or the `$match`/`$sort`/`$group _id`
   stage fields actually used.
3. For each query, check whether an existing index covers the filtered/sorted field(s) as a usable
   prefix per the rule above. `_id` lookups are always covered and need no action.
4. For `cardEmbeddingCollection`, cross-check the Atlas Search/vectorSearch index `path`s defined in
   `createSearchIndexes()` against the fields actually referenced in `db/mongo_pipeline.go`'s
   `$vectorSearch`/`$search` stages (currently `textEmbedding` and `text`), and confirm
   `numDimensions` still matches the embedding model in use (see `.claude/rules/semantic-search.md`
   for the current model/dims — don't hardcode a number here since it can change).
5. Weigh severity by how hot the path is: DAO methods called on every request (card-of-the-day,
   similar-card search, archetype lookups) matter far more than a rarely-hit admin path or a
   collection that will only ever hold a handful of documents. Don't demand an index for every field
   ever filtered — use judgment about query frequency and collection growth, consistent with
   `.claude/rules/coding-priorities.md`'s instruction not to add defensive work for scenarios that
   don't matter here.

Report each finding as: `<collection> is queried on {field(s)} by <DAO method> (db/....go:LINE) but
no current index covers that access pattern.`

## 3. Performance & memory

The "Performance" and "Low memory footprint" sections of `.claude/rules/coding-priorities.md` are
the checklist; apply them, don't re-derive new rules. Concretely look for:
- Sequential per-item downstream/DB calls in a loop that could batch (see check in section 1).
- Slices built with `append` in a loop whose final size is already known (e.g. `len(input)`) but
  declared without a capacity hint (`make([]T, 0, n)`).
- `for _, v := range someSlice` where `v` is a large struct (e.g. `cModel.YGOCard`, `CardDataMap`
  entries) — prefer ranging by index and using `someSlice[i]` to avoid copying.
- New allocation-heavy hot-path code (per-request buffers, encoders, etc.) that doesn't reuse a
  `sync.Pool` the way the gzip writer in `api/` does.
- Result sets read further than needed — Mongo queries that don't project (`SetProjection`) down to
  only the fields actually used downstream, letting unused fields cross the wire.
- New goroutine fan-out that's missing a `Wait()`/could deadlock, or that fans out work which is
  actually dependent (should be sequential/fail-fast instead of concurrent).

## 4. Library freshness / deprecated usage

- `go vet ./...` already flags some deprecated stdlib usage (covered in section 1, worth re-noting
  here if relevant).
- `go list -m -u all` against `go.mod` to see what has newer versions available.
- `github.com/ygo-skc/skc-go/common/v3` is treated as a vendored dependency per
  `.claude/rules/architecture.md` — don't inspect its internals. Only compare the pinned version in `go.mod` against latest via
  `go list -m -versions github.com/ygo-skc/skc-go/common/v3`, and note if newer.
- For the other direct dependencies in `go.mod` (`chi`, `go-playground/validator`,
  `mongo-driver/v2`, etc.), use `WebFetch`/`WebSearch` on their release notes/changelogs for: (a) any
  newer major/minor version, (b) deprecation notices affecting symbols actually imported here (grep
  `go.mod`'s module path across the repo to see what's imported), (c) new features genuinely
  relevant to this repo's usage (e.g. a mongo-driver bulk-write helper, a validator tag that could
  replace hand-rolled validation in `validation/`). Don't suggest an upgrade or new feature just
  because it exists — tie it back to something this repo's code actually does.

## What NOT to flag

This skill finds drift from the standard this repo's CLAUDE.md and `.claude/rules/` already state —
it is not license to invent stricter rules. Don't flag speculative micro-optimizations, defensive
guards against failure modes that can't occur in this environment, or style preferences those files
don't hold. When in doubt, quote the CLAUDE.md/`.claude/rules/` line the finding is drifting from.

## Report format

Respond directly in the conversation (no file needed unless the user asks for one). Structure:

```
## Go Drift Check

### 1. Idiomatic Go
- <file>:<line> — <what drifted> — <fix>
(or: "No drift found.")

### 2. Mongo index coverage
- <collection> queried on {field(s)} by <method> (<file>:<line>) — no covering index — <suggested index>
(or: "No drift found.")

### 3. Performance & memory
- <file>:<line> — <what drifted> — <fix>
(or: "No drift found.")

### 4. Library freshness / deprecated usage
- <module> <current> → <latest> — <why it matters here>
(or: "No drift found.")
```

If the user then asks to fix any finding, treat it as a normal edit: make the change, and rely on
the repo's `PostToolUse` hook (gofmt + vet + test on the changed package) rather than re-running the
whole audit.
