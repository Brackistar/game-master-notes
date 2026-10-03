# Code Quality Audit Reports

## Metadata

- Created: 2026-09-29
- Last updated: 2026-09-29
- Author: Codex
- File: `docs/audit/README.md`

## Purpose

This directory contains evidence-backed, read-only code-quality and maintainability audits for first-party project code. Audits emphasize cognitive-complexity risks, duplicated logic, oversized responsibilities, and refactoring safety. They do not include vendored or generated code unless the audit request explicitly changes the scope.

Use a dated descriptive filename:

```text
docs/audit/YYYY-MM-DD-code-smell-audit.md
```

## Required Audit Report

Every report must contain the following sections.

### Metadata And Scope

Record:

- Creation and last-updated dates.
- Author or agent role.
- Repository branch or revision reviewed.
- Included first-party paths.
- Explicit exclusions.
- Tools, Semgrep rules/configuration, and manual validation methods used.
- Failed or incomplete checks and their exact effect on coverage.

### Overview Of Findings

Summarize:

- Number of confirmed findings by priority and smell category.
- The highest-leverage maintainability risks.
- Areas scanned with no validated findings.
- Important limitations, hypotheses, or inconclusive coverage.

### Table Of Code Smells

Use at least these columns:

| ID | Priority | Smell | Location | Evidence | Impact | Confidence | Recommended direction |
| --- | --- | --- | --- | --- | --- | --- | --- |

Locations must identify a first-party file and the narrowest useful line or declaration. Evidence must describe observed structure rather than only naming a rule. Separate confirmed findings from tool-only candidates.

### Deep Explanation

For every confirmed finding, explain:

1. The affected responsibility and execution path.
2. The structural evidence and why it is a smell in this project.
3. Cognitive-complexity or duplication contributors.
4. Maintenance, correctness, performance, or testing consequences.
5. False-positive checks and relevant counter-evidence.
6. A minimal remediation direction that preserves behavior and architecture.
7. Characterization tests or validation needed before and after refactoring.

Do not present a heuristic Semgrep match as an exact cognitive-complexity score or complete duplication measurement. State the metric source and method whenever a numeric score is reported.

### Prioritized Action Plan

Order work by dependency and value:

- `P0`: correctness or severe change-risk blocker.
- `P1`: high-leverage maintainability improvement for active POC paths.
- `P2`: bounded cleanup that reduces future cost.
- `P3`: optional or deferred improvement requiring measured justification.

For each action include scope, likely files/modules, dependencies, expected benefit, estimated effort, regression risk, tests, and completion criteria. Prefer focused change sets and avoid combining unrelated refactors.

## Audit Method

The specialist should:

1. Check the worktree and establish first-party scope.
2. Run broad Semgrep analysis where available.
3. Run focused custom rules for repeated project-specific structures and complexity proxies.
4. Inspect relevant ASTs and source paths directly.
5. Search for sibling implementations to confirm duplication.
6. Validate every reported finding and record false positives or scanner failures.
7. Write only the report unless remediation was separately authorized.

Semgrep is one source of evidence, not the sole authority. Cognitive complexity and clone detection require careful manual validation or a dedicated metric tool when exact scores are needed.
