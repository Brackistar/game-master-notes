# Development Plan

## Metadata

- File: `docs/development-plan.md`
- Created: 2026-08-10
- Last updated: 2026-08-11
- User: brackistar

Related diagram: [development-plan.mmd](development-plan.mmd)

## Purpose

This document is the implementation map for the Android offline reset. It splits the work into independent tracks that can be planned, built, and tested without losing the whole-product shape.

The major development tracks are:

- Local PC data conversion application
- Android application
- Data layer
- Retrieval and local AI

Each track has its own plan document:

- [PC Pack Builder Plan](pc-pack-builder-plan.md)
- [Android App Plan](android-app-plan.md)
- [Data Layer Plan](data-layer-plan.md)
- [Retrieval and Local AI Plan](retrieval-ai-plan.md)

The general architecture diagram lives in [general-architecture.mmd](general-architecture.mmd).

Related local steering files live in `.github/instructions/`, including project, documentation, code quality, testing quality, refactoring, Android, pack-builder, data layer, and retrieval/local AI instructions.

## Suggested Build Order

1. Define the sourcebook pack contract first.
2. Build a minimal PC pack builder that can produce one importable test pack.
3. Add the local Room database to the Android project skeleton.
4. Implement pack import into SQLite tables without AI.
5. Add keyword search and citation display.
6. Add vector search with a small test embedding set.
7. Add the retrieval context assembler.
8. Add the `AiEngine` adapter and a fake local AI implementation for UI testing.
9. Integrate `llama.cpp` after the non-AI product path works.
10. Benchmark real models on low-requirement Android tablets before choosing defaults.

This order keeps risk visible. The app becomes useful as a searchable offline library before local generation is fully solved.

Current implementation note: steps 1-5 have an initial working path for sourcebook packs. The Android app can select a package folder, import `.gmnpack` metadata/documents/chunks into Room, retrieve chunks with SQLite FTS, and answer through a deterministic grounded MVP assistant. Vector search, campaign notes, and `llama.cpp` are still future work.

## Milestones

### Milestone 1: Documentation and Contracts

- Finalize the v1 sourcebook pack manifest.
- Define Android entities and repository boundaries.
- Define acceptance test scenarios for import, search, retrieval, and local assistant use.
- Keep all copyrighted sourcebook content out of the repo.

### Milestone 2: Minimal Offline Library

- Extend the initial Android skeleton.
- Add Room database and migrations.
- Add sourcebook packs, source documents, and source chunks. Initial sourcebook tables are implemented.
- Add systems, campaigns, sessions, notes, and tags later.
- Add manual note creation and basic search.

### Milestone 3: Sourcebook Pack Import

- Build the PC CLI to extract PDF text and write a pack archive.
- Import pack metadata, documents, chunks, citations, and embedding metadata into Android.
- Use a selected package folder and rescan it on app start.
- Show imported sourcebook packs in the Library.

### Milestone 4: Retrieval

- Add SQLite FTS for sourcebook chunks first; add notes later.
- Add local vector search.
- Merge keyword and semantic results.
- Build context bundles with citations.

### Milestone 5: Local Assistant

- Add the `AiEngine` interface.
- Create a deterministic grounded MVP engine for app development and tests.
- Integrate `llama.cpp` as the first real engine.
- Add a benchmark screen and record model suitability on the target tablet.

## Cross-Track Decisions

- The pack builder owns PDF extraction and embedding generation.
- The Android app owns import validation, local storage, search, retrieval, and assistant UX.
- The data layer is the app memory. Models should not be treated as a source of truth.
- The local AI path must remain manually triggered in v1.
- Cloud AI is out of v1, except as a future optional extension point.

## Acceptance Criteria

- A user can build a sourcebook pack from a PDF they own.
- A user can copy that pack to the tablet and import it.
- A user can create campaign notes and search them alongside sourcebook chunks.
- A user can ask for session help and receive an answer grounded in retrieved context.
- The app remains useful with internet disabled.
- Benchmarks on low-requirement Android tablets guide model choice instead of assumptions.
