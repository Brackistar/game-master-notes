# Wayfinder v1 - implementation-ready design specification

## Metadata

- Created: 2026-09-30
- Last updated: 2026-10-02
- Author: OpenAI Codex, Android Product Design Specialist
- File: `wayfinder-v1-design-spec.md`
- Mode: Design; no Android source, resources, manifests, Gradle files, or tests changed
- Status: Approved visual direction and implementation handoff
- Target: Android 10+, low-requirement ARM tablet first; phone, foldable, resizable window, keyboard, and pointer supported
- Primary destination: Ask the Books

## Product intent

Wayfinder is a practical GM command surface. It favors immediate orientation, predictable navigation, visible source scope, honest AI state, and rapid recovery over visual novelty. The interface should feel like dependable field equipment: dense enough for play, calm enough for reading, and never mysterious about what the local system is doing.

The product promise is: **ask the books, verify the evidence, and move on without losing your place.**

### Design principles

1. **Ask is home.** Launch into a ready composer; never auto-start retrieval or inference.
2. **Location is visible.** Top-level destination, current scope, answer mode, model state, and indexed-book count remain discoverable.
3. **Evidence is adjacent.** Supporting passages are beside answers when space permits and one action away otherwise.
4. **Long work stays controllable.** Show the current stage, elapsed feedback, and a Cancel action that remains independently responsive.
5. **Library is the collection; packs are plumbing.** Users browse sourcebooks in Library and manage the `.gmnpack` source folder from Library actions.
6. **Failure is actionable.** No books, no evidence, timeout, invalid output, and unavailable model each have distinct copy and recovery.
7. **Adapt to the window, not a device label.** Preserve state and reflow across rotation, split screen, folding, and resize.
8. **Do not pretend future features exist.** Session is a clearly labelled preview until its contracts are implemented.

## Information architecture

Top-level destinations, in stable order:

1. **Ask** - default and primary workflow.
2. **Library** - usable indexed sourcebooks; includes source management.
3. **Session** - future shell only; hide from production v1 if not useful.
4. **Settings** - local model and application configuration.

`Home` is removed as a destination because Ask performs the home role. `Packs` is removed from primary navigation because a pack is an import artifact, not a user task. Recent questions and pinned evidence are secondary workspace shortcuts, not top-level routes. They appear in an expanded drawer only after the underlying data is real; until then, omit them rather than showing dead controls.

See [`navigation-flow.mmd`](navigation-flow.mmd) and [`ask-state-flow.mmd`](ask-state-flow.mmd).

## Responsive layout contract

Use current window size and input capabilities. Do not infer layout from orientation or a hard-coded tablet boolean.

| Available width | Navigation | Content behavior |
| --- | --- | --- |
| Compact, under 600 dp | Labelled bottom bar; max 4 destinations | One pane. Related detail opens as a full screen. Evidence returns with Back. Composer stays above system/IME insets. |
| Medium, 600-839 dp | Labelled or expanded rail depending on height | One primary pane with 24 dp margins; list/detail or answer/evidence may use two panes only if each retains its minimum useful width. |
| Expanded, 840-1199 dp | Labelled rail by default | Ask uses main + supporting pane. Library uses list + detail. Drawer is optional only when real secondary workspace items justify it. |
| Large, 1200-1599 dp | Persistent labelled drawer | Drawer 240-280 dp; content maximum width prevents excessive line length. Supporting pane 320-400 dp. |
| Extra large, 1600 dp and above | Persistent labelled drawer | Center the working region; do not endlessly stretch answer prose or lists. |

Compact height below 480 dp overrides width assumptions: use the least obstructive navigation form, collapse nonessential headers, keep the composer and Cancel accessible, and prefer a single pane.

### Minimum pane sizes

- Ask main: 440 dp preferred, 360 dp absolute minimum.
- Evidence supporting pane: 300 dp preferred, 280 dp minimum.
- Library list: 320-400 dp.
- Library detail: 440 dp minimum.
- Gap/divider: 16-24 dp or 1 dp divider, depending on window width.

### Resize and posture behavior

- Preserve draft, messages, active request, selected source, evidence expansion, list scroll, and current destination.
- Resizing does not submit, cancel, re-run retrieval, or change scope.
- When two panes collapse, keep the currently focused pane visible. Back returns to its logical parent.
- Avoid content under a fold/hinge. Place panes on separate sides only when both meet minimum widths.
- Phone landscape and tabletop postures prioritize composer/action reachability over decorative headers.

### Adaptive implementation note

Android's current adaptive guidance supports `NavigationSuiteScaffold` for switching navigation forms and Navigation 3 scene strategies for canonical list-detail/supporting-pane relationships. The project currently uses Navigation Compose. A first implementation may preserve it and adapt panes inside each destination. Treat Navigation 3 and adaptive-navigation dependencies as a separately reviewed migration, not a hidden requirement of the visuals.

## Visual system

The canonical values live in [`tokens.md`](tokens.md).

The canonical state and ownership handoff lives in [`component-state-matrix.md`](component-state-matrix.md). Code generation should use that matrix to avoid inferring domain state from display strings.

- Deep navy navigation forms a stable spatial anchor.
- Cool neutral canvas and white/dark surfaces keep long reading comfortable.
- Cyan is reserved for primary actions, active scope, links, and focus—not decoration.
- Strong one-pixel boundaries replace excessive shadows.
- Modest 8-14 dp corners are modern without turning every container into a pill.
- Sentence-case labels and tabular alignment support rapid scanning.
- Typography is platform-safe and highly legible; source titles may wrap to two lines.

### Dark theme

Dark theme is required for implementation. [`09-ask-ready-dark-tablet.svg`](09-ask-ready-dark-tablet.svg) is the canonical Ask-ready dark reference for the measured physical-tablet viewport (about 1129 × 706 dp). Do not simply invert colors. Keep the navigation region darker than content, raise surface luminance by role, reduce pure-white glare, and measure contrast for body, metadata, outlines, focus, and statuses.

The urgent correction is theme-surface application, not a runtime or navigation rewrite: the app root must paint `background`, navigation must paint `navContainer`, and every text/control surface must use the matching semantic content role. Dark-theme text on an unpainted white platform window is a release-blocking contrast failure. Color literals in the shell must be replaced by semantic roles during implementation.

The broader Ask redesign can follow as a separate slice: constrain the working region, move model setup to Settings, reduce diagnostics prominence, and anchor the composer. Restoring the correct background/surface pairing must not wait for that broader slice.

### Density modes

Use one default density. Do not add a user-facing compact-density preference in v1. Pointer users may get hover affordances and tooltips, but touch targets remain at least 48 dp.

### Theme-independent sizing rules

Light and dark themes use identical measurement, breakpoint, typography, and reflow rules. Theme changes may alter only semantic color/elevation values; they must not move controls, change fixed dimensions, or produce different clipping.

- At 840-1199 dp, use a 96-112 dp labelled rail and center a main working region no wider than 720 dp when no evidence pane is present.
- Apply 16 dp outer margins on compact, 24 dp on medium, and 28-32 dp on expanded windows. The centered max width is a ceiling, not a forced width.
- Scope/mode controls use a minimum 48 dp target and wrap to a second row before labels truncate. Model state is a compact read-only status on Ask; import and selection belong in Settings.
- Empty-state content measures itself from content and available constraints. Do not use a fixed-height viewport around vertically scrollable invitation copy.
- The composer remains reachable without scrolling, uses 96-160 dp at expanded width for an empty/multiline draft, respects IME/system insets, and shares the same 720 dp main-pane maximum.
- At compact height below 480 dp, reduce decorative vertical gaps and example count before hiding status or composer actions. Body content scrolls; the composer/action remains available.
- At 200% font scale, header controls and example actions wrap; no essential label, explanation, or action is clipped or ellipsized into ambiguity.

## Component inventory and contracts

### Application shell

| Component | Required behavior |
| --- | --- |
| `WayfinderAppScaffold` | Owns insets, adaptive navigation, destination state, and global local/offline status. |
| `PrimaryDestinationItem` | Icon plus persistent label; selected state uses container, leading marker on drawer/rail, and semantics. |
| `LocationBar` | Compact breadcrumb or screen title; never becomes a second navigation hierarchy. |
| `OfflineReadyBadge` | Says `Offline · ready`, `Model needed`, or current blocking state; status is not color-only. |
| `KeyboardShortcutStrip` | Expanded windows with hardware keyboard only; hide on touch-only and compact layouts. |

### Ask

| Component | Required behavior |
| --- | --- |
| `QuestionComposer` | Multiline, 48 dp minimum height, clear label, submit icon plus accessible name; draft persists. `/` focuses only when not editing another field. |
| `ScopeChip` | Shows `All N books` or selected scope; opens a labelled selection surface. Scope remains unchanged during a request. |
| `AnswerModeControl` | Lookup, Explain, Summarize, Brainstorm, Compare. Use a menu or segmented treatment only when labels fit. |
| `AskEmptyState` | Short invitation, 2-4 example questions, indexed-book count, no fake conversation. |
| `GenerationProgress` | Stage label: `Searching books`, `Selecting evidence`, `Loading model`, `Writing answer`, or `Checking support`; indeterminate unless real percentage exists. |
| `CancelGeneration` | Visible text action, not icon-only. Must not wait behind the generation mutex. |
| `AnswerCard` | Question, support status, answer prose, inline citation references, fallback label when used. |
| `CitationLink` | Activates matching evidence, moves focus to it, and announces source label. Never use color alone. |
| `EvidencePane` | Cited sources first; retrieved-but-uncited sources appear only under a clearly separate heading if retained. |
| `EvidenceCard` | Source ordinal, title, page/range, excerpt, `Open passage`; source identity persists across reflow. |
| `RecoveryPanel` | Specific state, plain-language explanation, one primary recovery, and optional secondary action. |

### Library and sources

| Component | Required behavior |
| --- | --- |
| `LibraryToolbar` | Search/filter can remain future; `Manage sources` is always clearly available. |
| `SourcebookRow/Card` | Title, system/edition, indexed status, source filename, optional document/chunk metadata; entire row is one coherent target. |
| `SourcebookDetail` | User-facing book metadata and index health. Avoid exposing implementation-only identifiers by default. |
| `SourceFolderCard` | Human-readable folder name/path, permission state, last scan, and explicit `Change folder` / `Scan now`. |
| `ImportProgress` | Stage and counts where reliable; current file name only if safe and meaningful. |
| `ImportSummary` | Imported, unchanged, removed, and failed counts with expandable errors. |
| `AuthoritativeFolderWarning` | Explains that files removed from the selected folder may be removed from Library on scan. Must precede a folder change/scan when destructive pruning is possible. |

### Settings and models

| Component | Required behavior |
| --- | --- |
| `SettingsGroup` | Clear heading and related rows; no card-per-row visual noise. |
| `ModelStatusCard` | Installed/missing/unsupported state, model name, quantization, approximate size/minimum memory, and selected state. |
| `ModelImportAction` | Opens Android's system file picker. Copy names the expected file and that it stays on device. |
| `ModelLoadState` | Loading, ready, error, and compatibility fallback are distinct. No success toast used as the sole confirmation. |

### Session shell

The board describes future information hierarchy only: session name, quick Ask, pinned evidence, and notes. It is not an implementation requirement for v1. If the product cannot persist these objects, omit the destination instead of shipping a nonfunctional shell.

## Screen and state specification

### 1. Ask - ready (default launch)

Artifact: [`01-ask-ready.svg`](01-ask-ready.svg)

- Focus lands on the screen heading for accessibility, not forcibly in the keyboard-opening composer.
- Prominent composer is reachable without scrolling.
- Show current scope, answer mode, selected model label, offline state, and indexed count.
- Example prompts insert editable text; they do not submit automatically.
- With no history, do not show an empty evidence pane on compact. Expanded may show a useful `How answers work` placeholder.

### 2. Ask - generating and cancel

Artifact: [`02-ask-generating.svg`](02-ask-generating.svg)

- Preserve the submitted question above progress.
- Disable duplicate submit and settings that would mutate the in-flight request.
- Keep navigation available, but navigating away must not silently cancel. A running indicator follows the destination state; returning restores progress.
- Cancel is visible, enabled, and separately serviced. After activation: `Stopping…`, then return to Ready with the question available to edit/retry.
- Do not fabricate a percent or token stream if unavailable.

### 3. Ask - cited answer and evidence

Artifact: [`03-ask-cited-answer.svg`](03-ask-cited-answer.svg)

- Answer labels differentiate `Supported answer` and `Cited excerpts fallback`.
- Inline markers map one-to-one to displayed cited evidence.
- Expanded: answer and evidence scroll independently, with coordinated selection.
- Compact: a source summary button such as `2 cited sources` opens Evidence as a full-screen related destination. Back preserves answer scroll and draft.
- `Retrieved but not cited` content, if displayed, is visually and semantically separated from cited support.

### 4. Ask - recovery

Artifact: [`04-ask-recovery-states.svg`](04-ask-recovery-states.svg)

**No books**

- Message: `Add sourcebooks before asking.`
- Primary: `Add sourcebooks`; secondary: `How imports work`.
- Composer may accept a draft but Ask remains disabled with an explanatory supporting label.

**No evidence**

- Message: `I couldn't find enough support in the selected books.`
- Suggest naming a rule, character, place, or exact book term.
- Primary: `Edit question`; secondary: `Change scope`. Never generate an uncited speculative answer.

**Timeout**

- Message: `The local model reached its time limit.`
- If valid cited fallback exists, show it first as useful terminal output.
- Primary: `Try a narrower question`; secondary: `Use grounded excerpts` or `Change model`, according to actual availability.

### 5. Library - list and detail

Artifact: [`05-library-list-detail.svg`](05-library-list-detail.svg)

- Library lists user-facing sourcebooks, even if the current storage model summarizes packs.
- Expanded uses list-detail; compact navigates to full-screen detail.
- Detail shows source identity, system/edition, indexed status, last import, and safe diagnostics counts.
- `Manage sources` is a Library action, not primary navigation.
- Empty state points to source setup and distinguishes `No folder selected` from `Folder contains no valid packs`.

### 6. Manage sources - import

Artifact: [`06-source-management.svg`](06-source-management.svg)

- Current behavior treats the selected SAF folder as authoritative and prunes packs that disappear. The interface must explain this before any action that may remove indexed content.
- Show permissions and folder display name, not a raw URI as the primary label.
- During scan, provide stage, counts if known, and cancellation only if the importer supports safe cancellation. Do not promise it in UI before the contract exists.
- Summary separates imported, unchanged, removed, and failed. Errors identify the affected pack without exposing copyrighted text.

### 7. Session - future shell

Artifact: [`07-session-shell.svg`](07-session-shell.svg)

- Future concept: a session-scoped workspace with quick Ask, pinned evidence, and GM notes.
- Preserve top-level navigation and Wayfinder tokens.
- Production behavior until implemented: destination hidden, or a concise `Planned` screen with no fake controls.

### 8. Settings - local models

Artifact: [`08-settings-models.svg`](08-settings-models.svg)

- Local model setup moves out of the primary Ask content, while Ask retains a compact current-model status/action.
- Show deterministic grounded fallback as a useful available mode, not an error.
- Missing local GGUF: primary `Import model file`; copy states expected compatible model and local-only handling.
- Unsupported device: disable selection with explanation and keep grounded fallback available.
- Replacing a model requires explicit confirmation if it interrupts a loaded model or active request.

## Content and copy guidance

- Prefer verbs tied to the user's task: `Ask`, `Open evidence`, `Manage sources`, `Scan now`, `Import model file`.
- Use `sourcebook` or `book` in user-facing copy. Use `pack` only when naming `.gmnpack` files or troubleshooting import.
- Say `indexed` when content is available to search; do not imply the original PDF is stored or displayed.
- Say `supported` only when citations map to evidence supplied to the answer path.
- Avoid anthropomorphic waiting copy. Prefer `Searching 12 books…` to `I'm thinking…`.
- Never claim privacy solely with an icon. Use concise copy: `Runs on this device. No network required.`
- Error copy pattern: what happened, what remains safe, what action can resolve it.
- Preserve system and edition in ambiguous source titles.

## Accessibility

- Minimum 48 x 48 dp touch targets; avoid adjacent expanded hit regions that overlap.
- Text and meaningful boundaries meet WCAG AA contrast; measure both themes and disabled states.
- Support at least 200% font scale without clipping navigation labels, folder warnings, citation identity, or recovery actions.
- Use real Material components/semantics where possible. Custom cards expose role, label, state, and one primary action.
- Reading order on compact: location/title, status/scope, question/answer, evidence summary, composer. Expanded: navigation landmark, main-pane heading/content, supporting-pane heading/content.
- Pane headings let TalkBack users understand `Answer` and `Evidence` regions.
- Inline citations announce `Citation 1, Bestiary of Cinders, page 142`, not only `one`.
- Progress announces stage changes politely; do not repeatedly announce an indeterminate spinner.
- Cancel receives immediate state feedback.
- Selected navigation, selected book, support status, warning, and error use icon/text plus color.
- Decorative separators are excluded from semantics.

## Keyboard and pointer

Suggested shortcuts apply only when they do not conflict with the platform or text editing:

- `Alt+A`: Ask.
- `Alt+L`: Library.
- `Alt+S`: Session, only when enabled.
- `Ctrl+,`: Settings.
- `/`: Focus question when no text field/menu/dialog currently owns focus.
- `Ctrl+Enter`: Submit question.
- `Esc`: Close menu/dialog/evidence detail; while generating, first press focuses or reveals Cancel rather than cancelling invisibly.

Tab order follows visual task order. Arrow keys navigate radio/segmented groups and list selection where platform conventions support it. Enter/Space activates the focused control. Provide a visible 3 dp focus ring and a distinct hover state; hover never replaces focus or pressed feedback. Source rows use tooltips only for supplemental explanation, never for essential labels.

## Implementation mapping to the current checkout

| Current area | Design target | Required contract/work |
| --- | --- | --- |
| `app/GameMasterNotesApp.kt` | Adaptive shell; Ask default; stable top-level navigation | Change start route and navigation structure in an implementation task. Preserve back stack per destination where practical. |
| `app/AppRoute.kt` | Ask, Library, optional Session, Settings | Remove Home and Packs from top level; add routes only when screens are wired. |
| `feature:assistant/AssistantScreen.kt` | Responsive Ask states and evidence pane | Split the monolith into state components. Existing question/mode/model/messages/evidence data is a baseline; stage, cancellation UI, scope, and explicit timeout/fallback outcome need richer state. |
| `feature:library/LibraryScreen.kt` | Sourcebook list/detail and source-management entry | Current `SourcebookPackSummary` supports the initial list but lacks detail and document-level contracts. Avoid inventing fields in UI. |
| `feature:import/ImportPacksScreen.kt` | Manage Sources | Reframe copy and hierarchy. Existing selected folder, scanning state, summary, and errors map directly. Human-readable folder metadata and warning/confirmation may need contracts. |
| `feature:settings/SettingsScreen.kt` | Settings and local models | Currently placeholder. Model availability/installer logic is owned by Assistant today and should be exposed through a shared state/use case rather than duplicated. |
| `feature:session/SessionScreen.kt` | Future shell | Currently placeholder. Hide or mark planned until persistence and interaction contracts exist. |
| `core:design/GameMasterNotesTheme.kt` | Complete Wayfinder light/dark tokens | Current theme has only a small light palette. Add semantic colors, typography, shapes, focus, dark theme, and reusable components in implementation. |
| `core:data` | Library list/detail identity | Current pack can contain multiple documents. Product decision required on whether Library presents packs, documents/books, or both. |
| `core:ai` / `core:retrieval` | Honest stage/support status | UI must consume explicit runtime stop reason, fallback/validation outcome, cited IDs, and request-correlated stages rather than infer them from strings. |

### Recommended implementation slices

1. **Foundation:** Wayfinder tokens, dark theme, icons, deterministic previews, and shell with Ask default. No screen redesign beyond navigation.
2. **Ask states:** ready, progress/cancel, answer/evidence, and recoveries using existing contracts plus the already-planned explicit runtime outcomes.
3. **Library:** user-facing list/detail and empty states using only real metadata.
4. **Sources:** move import under Library, add authoritative-folder warning, preserve scan behavior.
5. **Settings/models:** extract model setup state from Assistant without duplicating AI ownership.
6. **Adaptive enhancement:** add canonical pane/navigation libraries only after dependency and migration approval.
7. **Session:** implement only after its product/data contracts are approved.

Each slice should include compact and expanded previews/tests plus large-text and dark-theme coverage.

## Intentional deviations from the original Direction 3 board

- The expanded permanent drawer is not mandatory at 840 dp. A labelled rail avoids wasting content space; use the drawer on large widths or when secondary workspace shortcuts exist.
- Recent questions and pinned evidence are omitted until persistence/contracts exist.
- Model import is moved to Settings; Ask keeps status and a route to setup.
- `Session` is explicitly future and may be hidden rather than presenting dead controls.
- The mock does not assume Navigation 3. It specifies pane behavior independent of implementation framework.
- Raw folder URIs are demoted in favor of a display name and permission status.

## Unresolved decisions requiring product or architecture approval

1. **Library unit:** Does one `.gmnpack` represent one user-visible book, or can a pack contain several books that must appear separately? Current schema permits multiple documents.
2. **Folder semantics:** Keep the selected folder authoritative with pruning, or change to an import inbox/copied-library model? This changes warnings and removal behavior.
3. **History persistence:** Are recent questions stored across launches, only in memory, or not in v1?
4. **Pinned evidence:** Is pinning session-scoped, campaign-scoped, or global? Do not ship the drawer entry before this is defined.
5. **Session scope:** What durable entities and lifecycle define a session?
6. **Source scope:** Can Ask scope by pack, document, system, or user collection? The first implementation can support `All indexed books` only.
7. **Dark theme policy:** system default plus manual override, or system-only for v1?
8. **Navigation technology:** preserve Navigation Compose for the first slice or separately migrate to Navigation 3/adaptive scene strategies.
9. **Import cancellation:** current importer contract does not advertise resumable/safe cancellation; UI must not offer it without verification.
10. **Opening original passages:** current app can show indexed excerpts but may not have a supported deep link into the original PDF. `Open passage` means in-app excerpt detail unless architecture adds source opening.

## Acceptance criteria for code generation

### Shell and navigation

- Ask is the cold-launch start destination and no work begins until submission.
- Top-level navigation order and labels are stable across size changes.
- Compact uses a labelled bottom bar; expanded uses labelled rail/drawer according to available width and content needs.
- Destination, draft, request, selection, and scroll state survive resize and rotation.
- Packs/import is reached from Library, not primary navigation.

### Ask

- Ready, generating, cancel/stopping, supported answer, cited fallback, no books, no evidence, timeout, model missing, and generic error are deterministic UI states.
- Cancel remains responsive and returns to an editable/retryable state.
- Displayed cited sources equal the response citation IDs; other retrieved evidence is not labelled as cited.
- Compact evidence is one action away and Back returns to the same answer position.
- Expanded answer/evidence panes meet their minimum widths and scroll independently.
- No automatic generation occurs on composition, navigation, resize, or restore.

### Library and sources

- Empty, populated, selected-detail, scanning, scan-success, unchanged, removed, partial-failure, permission-loss, and empty-folder states are represented.
- Folder authority/pruning consequence is visible before a relevant action.
- Folder display name and permission state are understandable without reading a raw URI.
- Import summary counts and errors match the underlying result object.

### Settings/models

- Ready, missing, loading, selected, unsupported-device, install-failed, and grounded-fallback states are distinct.
- System picker cancellation is harmless and returns to the previous state.
- Ask remains usable with grounded fallback when local GGUF is missing or unsupported.

### Accessibility and input

- Automated semantics tests cover labels, selection, disabled reasons, headings, progress, and citation actions.
- Touch targets are at least 48 dp; contrast is measured; layouts pass 200% font scale without lost actions/content.
- Keyboard-only traversal reaches all actions in predictable order, with visible focus; pointer hover is supplemental.
- TalkBack can navigate answer and evidence headings and understand citation/source relationships.

### Verification matrix

- Deterministic Compose previews/screenshots: phone portrait, phone landscape/compact height, portrait tablet, landscape tablet, split-screen medium width, dark theme, 200% font scale.
- UI tests for submit, cancel, evidence open/back, Library detail, source scan result, model import cancellation, and navigation-state preservation.
- Real-device pass on the target tablet for IME/insets, rotation, keyboard, pointer, TalkBack, long-running cancellation, and offline operation.

## Official references

- [Android: build adaptive navigation](https://developer.android.com/develop/adaptive-apps/guides/build-adaptive-navigation) - navigation bar for compact windows and rail/drawer choices for larger windows.
- [Android: use window size classes](https://developer.android.com/develop/adaptive-apps/guides/use-window-size-classes) - dynamic compact, medium, expanded, large, and extra-large breakpoints.
- [Android: get started with adaptive apps](https://developer.android.com/develop/adaptive-apps/guides/get-started-with-adaptive-apps) - runtime adaptation, posture, and canonical pane behavior.
- [Android: NavigationSuiteScaffold API](https://developer.android.com/reference/kotlin/androidx/compose/material3/adaptive/navigationsuite/NavigationSuiteScaffold) - adaptive navigation component and explicit layout types.
- [Android: Compose accessibility API defaults](https://developer.android.com/develop/ui/compose/accessibility/api-defaults) - semantics defaults and 48 dp minimum interactive target guidance.
- [Android: Compose focus targets](https://developer.android.com/develop/ui/compose/touch-input/focus/focus-target) - Tab/directional focus and interactive focus behavior.
- [WCAG 2.2 Understanding](https://www.w3.org/WAI/WCAG22/understanding/) - contrast, reflow, focus visibility, target size, and consistent navigation.

These references describe platform behavior and accessibility constraints. The navy/cyan palette, information density, copy, and Wayfinder component hierarchy are product design choices.
