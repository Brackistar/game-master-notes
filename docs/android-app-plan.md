# Android App Plan

## Metadata

- File: `docs/android-app-plan.md`
- Created: 2026-08-10
- Last updated: 2026-08-11
- User: brackistar

Related diagram: [android-app-plan.mmd](android-app-plan.mmd)

## Purpose

The Android app is the user's session and prep tool. It must work offline on low-requirement Android tablets, import sourcebook packs, store campaign notes, search local lore, and provide manually triggered AI help grounded in local context.

## Product Goals

- Run comfortably on modest Android tablet hardware.
- Provide fast access to notes and lore during sessions.
- Import sourcebook packs from local storage or microSD.
- Search campaign notes and sourcebook chunks.
- Use local AI only when the user asks for it.
- Keep the interface useful before local AI integration is complete.

## Suggested Stack

- Kotlin
- Jetpack Compose
- Android SDK
- Room for SQLite access
- WorkManager for imports and long-running background jobs
- Kotlin coroutines and Flow for asynchronous state
- Navigation Compose for app navigation

## Android Directory

Use `android/` for the native Android app.

The Android project uses these modules:

- `app`
- `core:data`
- `core:domain`
- `core:design`
- `core:importpacks`
- `core:retrieval`
- `core:ai`
- `feature:library`
- `feature:session`
- `feature:import`
- `feature:assistant`
- `feature:home`
- `feature:search`
- `feature:settings`

The `app` module owns the navigation host and depends on feature modules. Feature modules depend on shared core modules, not on each other. Keep cross-feature coordination in `app` or future domain-level use cases.

The current app navigation exposes only the useful first-slice surfaces:

- Home
- Library
- Sourcebook Packs
- Ask the Books

The session, search, and settings feature modules remain in the repository as future placeholders, but they are not part of the current app navigation or `:app` dependencies.

## Main Screens

- Home: indexed pack count and entry points for Library, Packs, and Ask the Books.
- Library: indexed sourcebook packs and basic pack metadata.
- Sourcebook Packs: selected folder, folder picker, manual rescan, import status, and validation errors.
- Ask the Books: question input, deterministic grounded answer, and retrieved citations.

Future screens:

- Session Mode: quick note capture, pinned lore, current scene notes, search, and assistant panel.
- Search: keyword and semantic results with filters and citations.
- Benchmarks: local model load time, memory notes, tokens per second, and qualitative result notes.
- Settings: storage locations, model files, import preferences, and privacy/offline status.

## Session Mode UX

Low-requirement tablet layouts should prioritize repeated live-session actions:

- Capture a note quickly.
- Search a rule or lore detail.
- Pin a relevant character, location, or sourcebook passage.
- Ask for grounded brainstorming.
- Save useful assistant output back into notes.

The session UI should avoid heavy decoration and favor readable, dense panels that can be scanned during play.

## Implementation Phases

### Phase 1: Skeleton

- Create the Android project. Initial multi-module skeleton added under `android/`.
- Add Compose navigation and app theme. Initial app host and shared theme added.
- Add focused first-slice screens for Home, Library, Packs, and Ask the Books.
- Keep Session, Search, and Settings as future modules outside current app navigation.
- Document Gradle build and test commands. Initial commands and Gradle wrapper are in place.

### Phase 2: Notes and Library

- Add Room database and repositories.
- Create systems, campaigns, sessions, notes, entities, and tags.
- Build CRUD screens for notes and campaigns.
- Add basic keyword search over user notes.

### Phase 3: Pack Import

- Add folder picker support for a `.gmnpack` package folder. Implemented with Android Storage Access Framework.
- Validate required pack archive members and manifest/chunk fields before import.
- Import metadata, documents, chunks, citations, and embedding metadata.
- Show scan status and errors.
- Make imports resumable or safely retryable.

### Phase 4: Search and Retrieval UI

- Add FTS-backed sourcebook chunk retrieval. Initial Ask the Books flow uses this.
- Add semantic search once vector storage is available.
- Show mixed results with clear source labels.
- Add context preview before assistant calls.

### Phase 5: Assistant Integration

- Add `AiEngine` interface and deterministic grounded MVP implementation.
- Build assistant UI with cited responses.
- Integrate `llama.cpp` runtime behind the adapter.
- Add model benchmark workflows.

## Testing

- Unit test repositories and use cases.
- Instrumented test database migrations.
- UI tests for note creation, search, import flow, and session mode.
- Manual device tests on low-requirement Android tablets for import time, memory pressure, and responsiveness.

Android CI runs `testDebugUnitTest` and `assembleDebug` for changes under `android/`.

## Risks

- Local model runtime may be slower than expected.
- Android storage permissions around microSD can be awkward.
- Large imports can block or overwhelm memory if not streamed.
- Complex UI can become cramped on smaller tablet displays if panels are too ambitious.

## Suggestions

- Make the non-AI app excellent first.
- Add a fake AI engine early so UI and retrieval can be tested before model integration.
- Keep assistant output saveable as normal notes.
- Design for one active model loaded at a time.
- Add an explicit offline/privacy indicator so the user trusts the app during sessions.
