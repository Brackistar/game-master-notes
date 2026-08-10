# Android App

## Metadata

- File: `android/README.md`
- Created: 2026-08-10
- Last updated: 2026-08-10
- User: brackistar

This directory contains the native Android app for `game-master-notes`.

The first implementation uses a native Android multi-module structure so product workflows, shared foundations, storage, retrieval, pack import, and AI runtime boundaries stay explicit from the start.

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
- `feature:session` - tablet session mode.
- `feature:import` - sourcebook pack import UI.
- `feature:search` - keyword and semantic search UI.
- `feature:assistant` - grounded assistant UI.
- `feature:settings` - storage, model, and privacy settings.

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

The Gradle wrapper files still need to be generated from an Android-capable environment.
