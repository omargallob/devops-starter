## Summary

<!-- What does this PR do and why? -->

Closes #<!-- spec / bug issue number -->

## Changes

- 

## Contract-Driven Development

- [ ] **Spec first** — a spec or bug issue is linked above and was opened before implementation
- [ ] **Contract defined** — public types in `pkg/tooldef/` or interface docs in `internal/cli/deps.go` updated
- [ ] **Contract tests** — pre/postconditions and invariants tested, not just happy path
- [ ] **Error handling** — contract violations produce a clear error or `panic`, not silent failure
- [ ] **Config version** — bumped `ConfigVersion` if config schema changed, migration provided
- [ ] **State version** — bumped `StateVersion` if state schema changed, migration provided

## Agentic Agile

- [ ] **Phase complete** — this PR closes one or more waves defined in the linked spec issue
- [ ] **No spec drift** — implementation matches the contract in the spec; spec updated if requirements changed
- [ ] **Review gate respected** — if this closes Wave 1 or Wave 2, the next wave has not been started

## Test Plan

- [ ] `make test` passes
- [ ] `make lint` passes
- [ ] Tested on: <!-- platform(s) -->
