# Android App

## Metadata

- File: `android/README.md`
- Created: 2026-08-10
- Last updated: 2026-08-11
- User: brackistar

This directory contains the native Android app for `game-master-notes`.

The first implementation uses a native Android multi-module structure so product workflows, shared foundations, storage, retrieval, pack import, and AI runtime boundaries stay explicit from the start.

## Current Working Slice

- Users choose a folder through Android's Storage Access Framework.
- The app persists read access to that folder and scans immediate `.gmnpack` children on startup.
- Imported packs are indexed into Room tables for packs, documents, chunks, and FTS search rows.
- Library and Home show indexed pack state.
- Ask the Books retrieves chunks with SQLite FTS and answers with a deterministic grounded MVP responder plus citations.

Deferred from this slice: vector search over `embeddings.npy`, campaign/note retrieval, and the real `llama.cpp`/GGUF runtime.

## Structure

- `app/` - Android application module and navigation host.
- `core:design` - Compose theme and shared UI styling.
- `core:domain` - durable domain identifiers and models.
- `core:data` - Room database and repository implementation boundary.
- `core:importpacks` - sourcebook pack inspection and import boundary.
- `core:retrieval` - search and context retrieval boundary.
- `core:ai` - local AI runtime adapter boundary.
- `feature:home` - home and entry workflows.
- `feature:library` - systems, campaigns, notes, lore, and tags.
- `feature:import` - sourcebook pack import UI.
- `feature:assistant` - grounded assistant UI.

The repository still contains placeholder modules for future session, search, and settings work, but they are not exposed in the current app navigation.

## Commands

From this directory:

```bash
./gradlew assembleDebug
./gradlew testDebugUnitTest
```

On Windows PowerShell:

```powershell
.\gradlew.bat assembleDebug
.\gradlew.bat testDebugUnitTest
```

The Gradle wrapper is committed in this directory. Local Android SDK paths belong in `local.properties`, which is intentionally ignored.
