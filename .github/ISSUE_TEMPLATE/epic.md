---
name: Epic
about: Large capability decomposed into child spec issues and agent work streams
title: "epic: "
labels: [epic]
assignees: []
---

## Goal

<!-- What this epic delivers end-to-end. One paragraph. -->

## Child Spec Issues

<!-- Create a spec issue for each work stream before implementation begins. -->

- [ ] #<!-- spec -->
- [ ] #<!-- spec -->

## Agent Work Streams

> Streams within the same wave can run in parallel. No Wave N+1 work begins until Wave N is reviewed and merged.

### Wave 1 — Contracts & Schema (sequential)

| Stream | Owner | Spec issue |
|--------|-------|------------|
| Public types | | |
| YAML / config schema | | |
| CLI skeleton | | |

### Wave 2 — Core Implementation (parallel)

| Stream | Owner | Spec issue |
|--------|-------|------------|
| | | |
| | | |

### Wave 3 — Integration & Docs (sequential)

| Stream | Owner | Spec issue |
|--------|-------|------------|
| | | |

## Review Gates

- [ ] Wave 1 reviewed and merged before Wave 2 begins
- [ ] Wave 2 integration test passes before Wave 3 begins

## Definition of Done

- [ ] All child spec issues closed
- [ ] End-to-end acceptance test passes
- [ ] `ARCHITECTURE.md` updated if design decisions changed
- [ ] No regressions in `make check`

## Notes

<!-- Architecture decisions, constraints, external dependencies. -->
