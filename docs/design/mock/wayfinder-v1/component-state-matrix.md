# Wayfinder v1 component and state matrix

## Metadata

- Created: 2026-10-02
- Last updated: 2026-10-02
- Author: OpenAI Codex, Android Product Design Specialist
- Mode: Design; implementation handoff only
- Status: Canonical companion to `wayfinder-v1-design-spec.md`

## State ownership rule

Compose renders explicit immutable UI state and emits user intents. It must not infer retrieval, support, timeout, model, import, or citation state from visible strings. ViewModels coordinate use cases; repositories, importers, retrieval, and AI runtime own dispatcher choice and cancellation. Window reflow is presentation state and must never restart domain work.

## Application shell

| State | Visible contract | Primary intents | Semantics and persistence | Current mapping |
| --- | --- | --- | --- | --- |
| Compact | Bottom navigation with Ask, Library, Settings | Select destination | Labels always visible; selected state announced | `app/GameMasterNotesApp.kt`; current routes need redesign |
| Medium/expanded | Labelled rail; drawer only at large widths or when real secondary items exist | Select destination; keyboard shortcut | Stable order and independent destination back stacks | No adaptive shell exists yet |
| Offline ready | `Offline · ready` with indexed count | None | Text plus icon, not color-only | Repository count exists; shell status does not |
| Blocking setup | `Model needed` or `Add sourcebooks` | Open relevant setup | Reason included in accessible description | Currently distributed across Assistant/import UI |

## Ask UI state

| Canonical state | Required data | Enabled actions | Disabled/frozen | Transition target |
| --- | --- | --- | --- | --- |
| `NoBooks(draft)` | Draft, indexed count `0` | Add sourcebooks, explain imports, edit draft | Submit | Successful import → `Ready` |
| `Ready(draft, scope, mode, model)` | Indexed count, editable question, stable scope/mode/model | Edit, choose scope/mode, submit | Cancel | Submit → `Retrieving` |
| `Retrieving(request, stage)` | Submitted question, request id, real stage, optional elapsed time | Cancel, navigate away | Duplicate submit and in-flight configuration | Evidence found → `Generating`; none → `NoEvidence` |
| `Generating(request, evidence, stage)` | Frozen question/scope/mode, cited candidate identity, real stage | Cancel, navigate away | Duplicate submit and request mutation | Supported → `SupportedAnswer`; rejected output → `CitedFallback`; deadline → `Timeout` |
| `Stopping(request)` | Same request identity | Navigate away | Cancel and submit | Runtime acknowledgement → `Ready` with draft restored |
| `SupportedAnswer(answer, citations)` | Answer text, support status, citation IDs, evidence records | Open citation, follow-up, copy if implemented | Nothing implicit | Citation → `EvidenceDetail`; follow-up → `Retrieving` |
| `CitedFallback(excerpts, citations, reason)` | Deterministic excerpt result and explicit fallback reason | Open evidence, narrower retry | Do not label as model answer | Retry → `Retrieving` |
| `NoEvidence(draft, scope)` | Original question and scope | Edit, change scope | Speculative generation | Retry → `Retrieving` |
| `Timeout(draft, fallback?)` | Deadline reason and any valid cited fallback | Narrow retry, use excerpts, model settings | Silent automatic retry | Chosen action determines target |
| `ModelMissing` / `Unsupported` | Compatibility reason and fallback availability | Model setup or excerpts-only | Unsupported model selection | Setup → Settings; fallback → Ready |
| `Failure(draft, safeMessage, diagnosticId?)` | User-safe message and preserved draft | Retry; diagnostics if available | Raw exception/source text | Retry → last safe preflight state |

### Ask component behavior

| Component | Focus and input | Compact | Expanded |
| --- | --- | --- | --- |
| `QuestionComposer` | `/` focuses when no editor owns focus; `Ctrl+Enter` submits; IME action submits only when valid | Sticky above IME/navigation when practical | Anchored at main-pane bottom; max readable width |
| `GenerationProgress` | Polite live-region announcement only when stage changes | In main flow with text Cancel | In main pane; navigation remains operable |
| `CitationLink` | Enter/Space opens; announces ordinal, source, and page | Opens full-screen evidence detail | Selects and focuses matching evidence card |
| `EvidencePane` | Independent scroll; heading is a navigation landmark | Related full-screen destination; Back restores answer | 300-400 dp supporting pane when minimums fit |
| `RecoveryPanel` | Focus moves to heading after state change, then primary recovery | Full-width single card | Constrained card in main pane; no empty support pane |

### Current tablet correction seams

| Observed seam | Immediate contract | Broader Wayfinder contract |
| --- | --- | --- |
| Dark semantic text is rendered over the default white window | Paint the root destination canvas with `background` and controls/cards with their semantic surface before drawing content | Keep the same semantic role mapping in both themes; verify every theme/window combination |
| Shell navigation uses literal colors and letter marks | Map rail colors to `navContainer`, `onNavContainer`, `navSelected`, and `onNavSelected`; use labelled accessible icons | Preserve stable Ask, Library, Settings order and selected marker across navigation forms |
| Assistant is one full-width column with 16 dp margins | Constrain the expanded main work region to 720 dp and center it within remaining space | Add the supporting evidence pane only in answer states where both panes meet minimum width |
| Full-width selectors and fixed-height content crowd the 706 dp height | Use content-sized controls, wrapping rows, and a body that owns remaining scroll space without fixed clipping | Keep the composer anchored/reachable; compact-height reduces nonessential spacing/examples first |
| Model import and diagnostics compete with Ask | Keep only compact model readiness on Ask; place import/setup in Settings; demote diagnostics to overflow/troubleshooting | Preserve explicit setup/recovery routes without making configuration the ready-state hierarchy |

## Library and source management state

| Canonical state | Presentation | Action contract |
| --- | --- | --- |
| No folder | Empty Library with `Choose source folder` | Opens SAF folder picker; cancellation is harmless |
| Empty folder | Folder exists but contains no valid packs | Scan again, change folder, import help |
| Populated / no selection | Sourcebook list and index health | Compact opens detail; expanded selects detail pane |
| Populated / selected | User-facing book detail, safe metadata, indexed status | Manage sources; return to list on compact |
| Permission lost | Folder name retained with blocked status | Re-authorize or choose folder |
| Scan preflight | Authoritative-folder consequence shown | Confirm scan/change; cancel preflight |
| Scanning | Real stage and counts only when reliable | Cancel only after importer exposes safe cancellation |
| Scan complete | Imported, unchanged, removed, failed counts | Return to Library; expand failures/removals |
| Partial failure | Successful imports remain visible; failures named safely | Retry scan; inspect safe diagnostics |

The current `core:data` schema allows a pack to contain multiple documents. Until the Library unit is approved, use fields already present in `SourcebookPackSummary` and label the item accurately; do not fabricate book-level metadata.

## Settings and model state

| State | Presentation | Action |
| --- | --- | --- |
| Missing | Model requirement, expected compatible file, local-only handling | Import model file; keep grounded fallback available |
| Picking file | Android system picker owns focus | Picker cancellation restores prior state |
| Installing/loading | Real stage, indeterminate unless measurable | Do not allow replacement races |
| Ready/selected | Model name, quantization, approximate size, ready label | Replace with confirmation if active work would be interrupted |
| Unsupported | Compatibility reason in text | Disable selection; retain fallback |
| Failed | Safe error and unchanged prior valid model if any | Retry or choose another file |

## Current Compose module mapping

| Design responsibility | Current source | Generation guidance |
| --- | --- | --- |
| Root navigation and default route | `android/app/.../GameMasterNotesApp.kt`, `AppRoute.kt` | Make Ask the start destination; retain Navigation Compose initially unless migration is separately approved |
| Ask state/content | `android/feature/assistant/.../AssistantScreen.kt` | Extract stateless content sections and a screen state; avoid moving runtime ownership into composables |
| Library list | `android/feature/library/.../LibraryScreen.kt` | Add empty/list/detail presentations against real repository data |
| Source scan | `android/feature/import/.../ImportPacksScreen.kt` | Route from Library and add explicit authority warning around existing behavior |
| Settings | `android/feature/settings/.../SettingsScreen.kt` | Replace placeholder; surface shared model state rather than duplicate Assistant logic |
| Session | `android/feature/session/.../SessionScreen.kt` | Hide from production navigation until contracts exist |
| Theme/components | `android/core/design/.../GameMasterNotesTheme.kt` | Add semantic light/dark tokens and reusable primitives before screen styling |

## Code-generation gates

Before generating implementation code:

1. Resolve Library's user-visible unit and authoritative-folder semantics.
2. Confirm whether Navigation Compose remains for the first slice; the design does not require Navigation 3.
3. Define an explicit Ask screen-state contract including stop reason, support/fallback result, runtime stage, and cited IDs.
4. Confirm safe import cancellation before exposing Cancel during scans.
5. Create deterministic fixtures for every state in this matrix.
6. Implement in the slice order in the main specification; do not combine schema, runtime, navigation, and full visual redesign in one change.

## Review acceptance

- Every matrix row has a deterministic preview or screenshot fixture at compact and expanded width where its layout differs.
- TalkBack names headings, status, disabled reasons, citations, selected items, and progress stages.
- Keyboard-only use reaches all actions with a visible 3 dp focus indicator; hover never substitutes for focus.
- At 200% font scale, actions reflow and remain available; labels and warnings are not ellipsized into ambiguity.
- Window resize preserves request, draft, destination, selection, and scroll and never triggers generation or scanning.
