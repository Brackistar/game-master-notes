# Wayfinder v1 design index

## Metadata

- Created: 2026-09-30
- Last updated: 2026-10-02
- Author: OpenAI Codex, Android Product Design Specialist
- Mode: Design; Android source and resources are read-only
- Status: Approved design direction; implementation-ready specification
- Selected direction: Wayfinder, navigation-first

## Start here

- [`wayfinder-v1-design-spec.md`](wayfinder-v1-design-spec.md): product behavior, visual system, responsive rules, component contracts, implementation mapping, acceptance criteria, and unresolved decisions.
- [`tokens.md`](tokens.md): implementation-oriented semantic tokens for light and dark themes.
- [`component-state-matrix.md`](component-state-matrix.md): canonical UI-state ownership, actions, semantics, and current Compose mapping.
- [`navigation-flow.mmd`](navigation-flow.mmd): top-level navigation and related-detail transitions.
- [`ask-state-flow.mmd`](ask-state-flow.mmd): complete Ask state and recovery model.

## Review boards

Every board contains a compact phone composition and an expanded tablet composition, except the recovery board, which compares the reusable recovery variants in one responsive system. They specify hierarchy and behavior rather than pixel-perfect rendering.

1. [`01-ask-ready.svg`](01-ask-ready.svg) - default launch, ready to ask.
2. [`02-ask-generating.svg`](02-ask-generating.svg) - staged progress and immediate Cancel.
3. [`03-ask-cited-answer.svg`](03-ask-cited-answer.svg) - answer, inline citations, and evidence.
4. [`04-ask-recovery-states.svg`](04-ask-recovery-states.svg) - no books, no evidence, and timeout.
5. [`05-library-list-detail.svg`](05-library-list-detail.svg) - browse indexed sourcebooks and inspect one.
6. [`06-source-management.svg`](06-source-management.svg) - authoritative folder, scan/import, warning, and results.
7. [`07-session-shell.svg`](07-session-shell.svg) - explicitly future, disabled in v1 unless implemented.
8. [`08-settings-models.svg`](08-settings-models.svg) - local model setup and device-aware status.
9. [`09-ask-ready-dark-tablet.svg`](09-ask-ready-dark-tablet.svg) - implementation-ready dark-theme Ask state for the measured 1129 × 706 dp tablet viewport.

## Decision boundary

The approved visual system does not by itself authorize a Navigation 3 migration, new dependencies, persistence changes, or unimplemented Session features. Those remain engineering/product decisions called out in the handoff.
