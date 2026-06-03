---
name: Feature Spec
about: Spec-first feature definition — contract before code
title: "spec: "
labels: [spec]
assignees: []
---

## Summary

<!-- One-sentence description of the capability. -->

## Motivation

<!-- Why does this matter? What problem does it solve? -->

## Contract

### Public Types

<!-- New or modified types in pkg/tooldef/ or exported interfaces in internal/cli/deps.go. -->

```go
// TODO: define types / interface changes
```

### YAML Schema

<!-- New or modified fields in configs/default.yaml. Leave blank if no config changes. -->

```yaml
# TODO: schema snippet
```

### CLI Interface

<!-- New subcommand, flags, or output format changes. Leave blank if no CLI changes. -->

```
devops-starter <subcommand> [flags]
```

## Acceptance Criteria

- [ ] <!-- criterion 1 -->
- [ ] <!-- criterion 2 -->
- [ ] Contract tests pass: preconditions, postconditions, and invariants verified
- [ ] `make check` passes (fmt + vet + lint + test)

## Implementation Phases

> Break into waves for agent swarming. Each wave can be parallelised; review gates between waves.

### Wave 1 — Foundation (sequential, review gate before Wave 2)

- [ ] Define public types / constants
- [ ] Write contract tests (red → green)
- [ ] Update YAML schema and bump `ConfigVersion` if needed

### Wave 2 — Implementation (can be parallelised)

- [ ] <!-- task A -->
- [ ] <!-- task B -->

### Wave 3 — Integration

- [ ] Wire CLI command(s)
- [ ] Update config / state if schema changed

### Wave 4 — Validation

- [ ] All acceptance criteria checked
- [ ] README / docs updated

## Notes

<!-- Constraints, open questions, links to related issues or ADRs. -->
