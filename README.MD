# game-master-notes

## Metadata

- File: `README.MD`
- Created: 2026-08-10
- Last updated: 2026-08-10
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

The app will use an `AiEngine` adapter so different local runtimes and models can be tested without rewriting the product workflow.

Initial benchmark candidates include:

- LFM2.5-350M
- SmolLM2-360M-Instruct
- SmolLM2-135M-Instruct
- MobileLLM and MobileLLM/ParetoQ variants
- Qwen3-0.6B as an upper quality benchmark

The expected sweet spot is around 350M parameters. Larger models may be tested, but v1 should not depend on 1B+ models running comfortably on the target tablet.

## Current Status

This project is in architecture reset and early implementation. The repository currently describes the intended Android offline architecture, implementation tracks, Mermaid diagrams, local VS Code/GitHub-style steering files, the first sourcebook pack builder slice, and an initial multi-module Android app skeleton.

The next step is to define the Android local data model, add Room dependencies and entities, and build the Android sourcebook pack importer against the `.gmnpack` format produced by `pack-builder/`.

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
- Import packs into the Android app from local storage or microSD

### Phase 3: Retrieval

- Add keyword search
- Add local vector search
- Merge note, lore, and sourcebook results into a context assembler
- Show citations for retrieved source material

### Phase 4: Local Assistant

- Add the `AiEngine` interface
- Integrate the first local runtime path
- Build grounded assistant prompts over retrieved context
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
- `android/feature/search/` for search UI
- `android/feature/session/` for session mode
- `android/feature/settings/` for settings

Gradle wrapper files still need to be generated from an Android-capable environment.

## Android Commands

From `android/`, after Gradle wrapper generation:

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
