# game-master-notes

## Metadata

- File: `README.MD`
- Created: 2026-08-10
- Last updated: 2026-08-28
- User: brackistar

An offline Android lore and notes assistant for game masters running multiple roleplaying systems.

## Overview

`game-master-notes` is being reset around a new goal: a tablet-first Android app that acts as a local source of campaign notes, world lore, and searchable sourcebook knowledge. The app is intended to run on low-requirement Android tablets and remain useful without internet access during prep or live sessions.

The v1 product focuses on three practical jobs:

- Capture and organize campaign notes, NPCs, locations, factions, items, timelines, and tags.
- Import prebuilt sourcebook packs generated from PDFs the user already owns.
- Use local retrieval plus a small on-device AI model to brainstorm, answer lore questions, and support session direction with cited context.

The local model does not need to memorize game systems or campaign lore. That knowledge lives in the app database, sourcebook packs, and user notes. The assistant's job is to work with retrieved context and produce grounded, useful help.

## Target Device Profile

The first target profile is a modest Android tablet rather than a flagship device:

- Android 10 or newer
- Entry-level or midrange ARM CPU
- 4 GB RAM minimum, 6 GB preferred
- 64 GB internal storage minimum, 128 GB preferred
- Optional microSD expansion support
- 8-inch or larger display

This hardware makes offline-first design important. The app should favor small local models, compact indexes, resumable imports, and predictable memory usage over heavyweight agent stacks.

## Planned Features

- Multi-system library for different RPG rulesets and settings
- Campaign and session organization
- Notes for characters, locations, factions, items, events, timelines, and loose ideas
- Sourcebook pack import from local storage or microSD
- Keyword and semantic search across notes, lore, and imported sourcebook chunks
- Session mode with quick note capture, pinned lore, search, and assistant panel
- Grounded local assistant for brainstorming, continuity checks, NPC ideas, rulings support, and session direction
- Model benchmark view for load time, memory use, tokens per second, and answer quality notes

## Sourcebook Packs

Sourcebook PDFs are not committed to this repository and are not shipped with the app.

In v1, PDFs are processed on a local PC using a pack builder CLI. The generated pack contains metadata, extracted text chunks, page or citation references, and embeddings. The Android app imports that pack and stores it locally.

This keeps copyrighted sourcebooks private, avoids slow tablet-side PDF processing, and makes setup practical for large books.

## Local AI Strategy

The app uses an `AiEngine` adapter so different local runtimes and models can be tested without rewriting the product workflow. The current Android slice includes a real local model path through `llama.cpp` and GGUF files, with LFM2.5-350M as the active low-requirement tablet target.

Current model handling:

- The app discovers installed model files in app-private storage.
- The model selector shows only installed compatible LFM2.5 files plus the deterministic grounded fallback.
- Qwen, Gemma, and Phi profiles remain behind the adapter for future testing, but they are not exposed as placeholders in the current UI.
- The deterministic fallback remains available when the local model is missing, too slow, or produces unusable output.

Future benchmark candidates include:

- LFM2.5-350M
- SmolLM2-360M-Instruct
- SmolLM2-135M-Instruct
- MobileLLM and MobileLLM/ParetoQ variants
- Qwen3-0.6B as an upper quality benchmark

The expected sweet spot is around 350M parameters. Larger models may be tested, but v1 should not depend on 1B+ models running comfortably on the target tablet.

## Current Status

This project is in architecture reset and early implementation. The repository currently contains the intended Android offline architecture, implementation tracks, Mermaid diagrams, local VS Code/GitHub-style steering files, the first sourcebook pack builder slice, an Android multi-module app, and CI workflows for both Android and the pack builder.

The Android app now has the first working sourcebook slice: users can choose a device folder containing `.gmnpack` archives, the app persists that folder permission, scans it on startup, indexes pack metadata/documents/chunks in Room, searches chunks with SQLite FTS, and provides an "Ask the Books" assistant backed by cited retrieved chunks.

The active Android navigation is intentionally small:

- Home
- Library
- Packs
- Ask the Books

Current limits are intentional for this slice: vector search from `embeddings.npy` is not active yet, campaign/note retrieval is not active yet, and the local LLM is constrained to short grounded answers over retrieved sourcebook context.

The assistant retrieval path currently uses SQLite FTS plus Kotlin-side relevance filtering. It searches sourcebook chunks, rejects weak partial matches, diversifies results across citations and books, and extracts up to four cited source blocks while preserving multiple relevant paragraphs per block. Short conversational follow-ups reuse the previous user question during retrieval. The local LLM receives a compact multi-block evidence brief and is prompted to synthesize a direct cited answer in a selected lookup, explain, summarize, brainstorm, or compare mode. The deterministic fallback displays separated cited blocks when the LLM is unavailable, too slow, uncited, malformed, or otherwise fails grounded-answer checks. Retrieved source blocks are expandable in the chat UI, and no-result searches return recovery guidance without invoking the model.

## Roadmap

### Phase 1: Android Foundation

- Create the Android project
- Set up local persistence and core data entities
- Build campaign, session, and note screens
- Document build and test commands

### Phase 2: Sourcebook Packs

- Create the local PC pack builder
- Extract PDF text into citeable chunks
- Generate embeddings and package metadata
- Import packs into the Android app from a user-selected local storage or microSD folder

### Phase 3: Retrieval

- Add keyword search
- Add hybrid local vector search from pack embeddings
- Merge note, lore, and sourcebook results into a context assembler
- Show citations for retrieved source material

### Phase 4: Local Assistant

- Add the `AiEngine` interface
- Integrate the first local runtime path with `llama.cpp`
- Build grounded assistant prompts over retrieved paragraph excerpts
- Add tablet benchmark workflows for candidate models

### Phase 5: Session Polish

- Optimize the low-requirement tablet session layout
- Add pinned lore and fast note capture
- Improve import reliability, search quality, and offline robustness
- Consider optional cloud escalation only after the offline v1 works

## Repository Structure

The repository is still being bootstrapped. Current project source is mostly documentation:

- `README.MD` for product direction, roadmap, and repository map
- `AGENTS.md` for local Codex steering when present in the workspace
- `docs/` for architecture, planning, and Mermaid diagrams
- `.github/instructions/` for local VS Code/GitHub-style steering files

Implementation directories:

- `android/` for the Android application
- `pack-builder/` for the local PC sourcebook pack CLI

The Android implementation has begun as a native multi-module project:

- `android/app/` for the Android application and navigation host
- `android/core/ai/` for the local AI runtime adapter boundary
- `android/core/data/` for the database and repository implementation boundary
- `android/core/design/` for shared Compose theme and UI primitives
- `android/core/domain/` for durable domain identifiers and models
- `android/core/importpacks/` for sourcebook pack import logic
- `android/core/retrieval/` for search and retrieval contracts
- `android/feature/assistant/` for assistant workflows
- `android/feature/home/` for the home entry point
- `android/feature/import/` for sourcebook pack import UI
- `android/feature/library/` for library workflows
- `android/feature/search/` for future search UI
- `android/feature/session/` for future session mode
- `android/feature/settings/` for future settings

The Android Gradle wrapper is committed so builds can run without a separately installed Gradle distribution.

## Android Commands

From `android/`:

```bash
./gradlew assembleDebug
./gradlew testDebugUnitTest
```

On Windows PowerShell:

```powershell
.\gradlew.bat assembleDebug
.\gradlew.bat testDebugUnitTest
```

## Pack Builder Commands

From `pack-builder/`:

```bash
uv venv
uv sync --extra dev
uv run pytest
uv run pack-builder --help
```

The pack builder can also be run through the local virtual environment directly after syncing:

```bash
.\.venv\Scripts\python.exe -m pytest
.\.venv\Scripts\pack-builder.exe --help
```

## Continuous Integration

GitHub Actions workflows live under `.github/workflows/`:

- `android-ci.yml` runs Android unit tests and `assembleDebug` when Android files change.
- `pack-builder-ci.yml` runs the pack-builder test suite on supported Python versions when pack-builder files change.

## Documentation Map

- [Android Offline Architecture](docs/android-offline-architecture.md) with [general architecture diagram](docs/general-architecture.mmd)
- [Development Plan](docs/development-plan.md) with [development roadmap diagram](docs/development-plan.mmd)
- [Android App Plan](docs/android-app-plan.md) with [Android app diagram](docs/android-app-plan.mmd)
- [PC Pack Builder Plan](docs/pc-pack-builder-plan.md) with [pack builder diagram](docs/pc-pack-builder-plan.mmd)
- [Data Layer Plan](docs/data-layer-plan.md) with [data layer diagram](docs/data-layer-plan.mmd)
- [Retrieval and Local AI Plan](docs/retrieval-ai-plan.md) with [retrieval and AI diagram](docs/retrieval-ai-plan.mmd)

## Steering Files

Local steering files live in `.github/instructions/` and use VS Code/GitHub-style `*.instructions.md` frontmatter with `applyTo` patterns. They currently cover project direction, documentation, Android architecture, Compose UI, Android storage/work, Android performance, data layer, retrieval/local AI, code quality, testing quality, refactoring, and pack-builder work.

`AGENTS.md` and `.github/` are currently ignored by `.gitignore`, so they should be treated as local workspace steering unless that ignore policy changes.

## Contributing

Issues, feedback, and ideas are welcome.

This project is being built in public by [`brackistar`](https://github.com/brackistar), starting from architecture and scope decisions before expanding into implementation.

## License

No license has been defined yet for this repository.

Until a license file is added, please treat the project as source-available for review only rather than assuming open-source reuse permissions.
