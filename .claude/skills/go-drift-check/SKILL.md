---
name: go-drift-check
description: This skill should be used when the user asks to "audit the Go code", "check for idiomatic Go", "review Mongo indexes", "check for missing indexes", "check for new library features we should adopt", "are we behind on any dependencies", "check for deprecated library usage", "check for performance or memory regressions", "run a drift check", "check for drift", or wants a health check of this repo's Go code against this repo's coding-standards rules (CLAUDE.md plus its linked files under `.claude/rules/`): idiomatic Go, usage-driven Mongo indexing, performance, memory footprint, and adoptable new library/stdlib capabilities. Also trigger for periodic or pre-release code-quality audits of this repository, not for reviewing a single PR's diff (use /code-review for that).
---

# Go Drift Check

Audit this repo's Go code for drift across four dimensions and report findings — do not silently
fix anything unless the user asks. This is a reporting skill: read code, compare it against a
baseline, and list what has drifted with file:line references and a concrete fix suggestion.

## Scope

Default to auditing the whole repo (`api/`, `db/`, `suggest/`, `downstream/`, `model/`,
`validation/`, `main.go`), excluding `testing/` mocks. If the user asks to check only recent work,
scope to `git diff release...HEAD` instead.

For a repo this size (~3.2k non-test LOC), run checks 1–3 directly with `Bash`/`Grep`/`Read` — they
read overlapping files and produce findings that must be deduped against each other (the batching
check appears in both section 1 and section 3), so splitting them across subagents costs more than
it saves, and second-hand `file:line` references can't be trusted without re-reading the code anyway.

**Exception:** dispatch check 4 (library capability drift) as a background subagent *before* starting
check 1, then collect its result when assembling the report. It's network-bound, reads only `go.mod`
plus release notes, and shares no findings with the other checks, so it runs while the main context
reads Go code. Revisit this split if the repo grows past ~10k LOC.

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

## 4. Library capability drift (new features worth adopting)

Per Scope, this check should already be running in a background subagent dispatched before check 1 —
what follows is that agent's brief. Run it inline only if no agent was dispatched.

**This is not a version-lag check.** `.github/renovate.json` automerges `minor`, `patch`, `pin`, and
`digest` updates, so this repo is on the latest minor/patch by construction — reporting "x.y.z →
x.y.z+1" is noise. The actual drift is the opposite: because those bumps merge without review, **new
capabilities arrive in `go.mod` and are never adopted**. Find those.

Scope the search to **releases from roughly the last 12 months**. Don't read a dependency's full
history.

Method:
1. `go list -m -u all` — one command, used only to spot a **major** version available (see below).
   Don't build a report section out of its minor/patch output.
2. For each direct dependency whose API this repo actually exercises — `mongo-driver/v2`, `chi`,
   `go-playground/validator`, `grpc`, `ip2location`, `cors` — `WebFetch`/`WebSearch` its release
   notes for that ~12-month window and look for additions that map onto something this repo already
   does by hand. Grep the module path across the repo first so you know which symbols are in use.
   Skip deps this repo barely touches directly (`locales`, `universal-translator`, `uuid`, `x/net`) —
   they're plumbing, and their release notes won't map to adoptable code here.
3. Same treatment for the **Go stdlib**, per `.claude/rules/coding-priorities.md`'s "prefer the
   standard library" line: check Go releases from the last ~12 months (`go.mod` currently pins
   `go 1.26`) for stdlib or language additions that would replace hand-rolled code here — e.g. a new
   `slices`/`maps` helper standing in for a manual loop in `suggest/`.
4. `github.com/ygo-skc/skc-go/common/v3` — check its GitHub **release notes only** for new shared
   helpers worth adopting. Never inspect its source; it's treated as a vendored dependency per
   `.claude/rules/architecture.md`.

**Major version bumps are the exception that deserves real time.** If a direct dependency has a new
major available (v2 → v3, i.e. a changed module path), don't just note the number — report what
changed, which files here import the module, and roughly what migration would involve. Renovate does
*not* automerge majors, so one may also be sitting in an open PR; `gh pr list` is worth a look.

Deprecations still count when they affect a symbol this repo imports, but they're secondary — lead
with adoptable additions.

Every finding must tie back to concrete code. "mongo-driver added X" is not a finding; "mongo-driver
added X, which would replace the hand-rolled Y at `db/foo.go:120`" is. If a release window turns up
nothing that maps to this repo, say so and move on — a short, empty section 4 is the expected result
most runs.

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

### 4. Library capability drift
- <module/stdlib> added <feature> in <version> — could replace <what> at <file>:<line>
- MAJOR AVAILABLE: <module> <current> → <new major> — <what changed> — <files importing it>
(or: "No adoptable additions found in the last ~12 months.")
```

If the user then asks to fix any finding, treat it as a normal edit: make the change, and rely on
the repo's `PostToolUse` hook (gofmt + vet + test on the changed package) rather than re-running the
whole audit.
