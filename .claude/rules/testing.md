---
paths:
  - "**/*_test.go"
  - "testing/**"
---

# Testing notes

- `testing/setup.go` `chdir`s to the repo root in its `init()` so relative paths (`./certs`, `./data`)
  resolve in tests. Tests set `IS_CICD` / run as `*.test` so the IP DB and env-file loading are skipped.
- Mocks live in `testing/` (`skc_suggestion_engine_dao_mocks.go`, `ygo_service_mock.go`, plus card/archetype fixtures).
