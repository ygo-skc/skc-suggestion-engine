---
name: releasing-the-api
description: Use when cutting a release for skc-suggestion-engine — choosing the version, creating and pushing a version tag, or publishing GitHub release notes.
---

# Releasing the API

## Overview

skc-suggestion-engine is a single Go module at the repo root and a deployed **application** — no
other repo imports it. So tags are bare `vX.Y.Z` with no module prefix, and the version that
actually matters is the one the running binary reports at `/status`.

That version is hardcoded in `api/status_handler.go`:

```go
status := cModel.APIHealth{Version: "3.2.0", Downstream: downstreamHealth}
```

The string has no `v` prefix; the tag does.

## Pre-flight

The default branch is `release`. Run from an up-to-date checkout of it.

```bash
grep -n 'APIHealth{Version' api/status_handler.go
PREV=$(git tag --list 'v*' --sort=-v:refname | head -1)
git log --oneline "$PREV"..HEAD
git diff --stat "$PREV"..HEAD
make test
```

**The version string is the source of truth — read it first.** If it already names an unreleased
version (says `3.2.0` while the newest tag is `v3.1.4`), that is the version to cut. Match it and
skip the bump table.

**If the string is stale** — still naming the version `$PREV` already released — **stop. Do not
tag.** Report which version it should become and let the user commit and push that bump. Tagging a
commit whose binary self-reports the previous version ships a `/status` endpoint that lies.

`make test` must pass before tagging. The Unit Test workflow runs on `tags: v**`, so a failing test
becomes a permanent red check on an already-published tag.

## Choosing the version

Only needed when the version string hasn't already been decided. Semver here describes the **HTTP
API surface** — the `/api/v1/suggestions` and `/api/v2/suggestions` namespaces — not Go identifiers.

| Bump | When |
| --- | --- |
| Patch | Dependency bumps, docs, config, internal refactors, bug fixes — no change to any request or response shape |
| Minor | New endpoint, new optional response field, opt-in behavior — backward compatible for existing clients |
| Major | Breaking change to an existing endpoint's contract. Prefer adding a new `/api/vN` namespace alongside the old one (how v2 was introduced) over breaking v1 in place |

## Release notes

**Title:** `vX.Y.Z`, or `vX.Y.Z: Short Theme` when the release has a headline —
`v3.1.0: Improving Archetype Suggestions`. Bare tag name is right for routine releases.

**Body**, in this order:

1. `## Changes`
2. One `*` bullet per user-visible change; two-space-indented sub-bullets for detail
3. On dependency-heavy releases, the renovate lines: `* <title> by @renovate[bot] in <PR url>`
4. Blank line, then
   `**Full Changelog**: https://github.com/ygo-skc/skc-suggestion-engine/compare/<PREV>...<NEW>`

Use `*` for bullets, not `-`. The heading is `## Changes`; `## What's Changed` is GitHub's
auto-generated default and appears only on releases nobody hand-wrote.

## Sequence

Show the version, the diff, the `make test` result, and the drafted notes. Get approval **once**.
Then run all three steps without stopping again:

```bash
git tag vX.Y.Z <commit>          # lightweight: no -a, no -m
git push origin vX.Y.Z
gh release create vX.Y.Z --repo ygo-skc/skc-suggestion-engine \
  --title "vX.Y.Z" --notes-file notes.md
```

Write `notes.md` to a scratch directory, not into the repo.

Push the tag first. `gh release create` attaches to an existing tag but invents one from the
default branch when the tag is missing.

## Why approval comes before the push

The pushed tag is what production gets built from and what `/status` gets checked against, and the
GitHub Release announces the change to API consumers. Approval is the last cheap moment — after the
push, a wrong version is corrected with another release, not an edit.

## Common mistakes

- **`git tag -a`.** Every tag in this repo is lightweight (`git cat-file -t v3.1.4` → `commit`). An
  annotated tag carries a message nobody reads; the notes belong in the GitHub Release.
- **Borrowing skc-go's tag prefix.** There is no `common/` or `ygo-service/` prefix here. Bare
  `vX.Y.Z`.
- **Tagging without reading `api/status_handler.go` first.** The tag and the binary's self-reported
  version silently disagree, and nothing catches it until someone hits `/status` in prod.
- **Pasting GitHub's generated notes wholesale.** That is how `v3.1.4` acquired a stray bullet
  referencing `v3.1.1`'s PR, and `v3.1.1` a duplicated `## Changes` heading. Write the bullets by
  hand; borrow only the renovate lines and the Full Changelog footer.
- **Dropping the Full Changelog footer.** `v3.1.0` is missing it. It is the last line of every
  release.
