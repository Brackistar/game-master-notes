# Android Offline Architecture

## Metadata

- File: `docs/android-offline-architecture.md`
- Created: 2026-08-10
- Last updated: 2026-08-10
- User: brackistar

Related diagram: [general-architecture.mmd](general-architecture.mmd)

## Purpose

This document defines the v1 architecture for `game-master-notes` after the Android reset. The app is an offline lore, sourcebook, and notes assistant for game masters using a low-requirement Android tablet.

The design goal is not to run a large general assistant on tablet hardware. The design goal is to combine well-structured local data, retrieval, citations, and small local models so the app can help during prep and live sessions without internet access.

## Android App Modules

The first implementation should use a native Android structure with Kotlin and Jetpack Compose.

- App shell: navigation, theming, settings, permissions, and import entry points.
- Library: systems, sourcebooks, campaigns, sessions, notes, NPCs, locations, factions, items, timelines, and tags.
- Session mode: quick note capture, current scene notes, pinned lore, search, and assistant panel for low-requirement tablet layouts.
- Storage: Room entities, DAOs, migrations, and local file references for imported packs and model files.
- Search: SQLite FTS queries, local vector search, and result ranking.
- Retrieval: context assembly from campaign notes, user lore, and sourcebook chunks.
- AI: `AiEngine` adapter, prompt construction, streaming responses when available, and benchmark reporting.

Keep modules small at first. The important boundary is between product workflows, local storage/search, pack import, and AI runtime integration.

## Sourcebook Pack Format

Sourcebook packs are generated on a local PC from PDFs the user already owns. The Android app imports generated packs, not raw project-bundled sourcebook text.

A v1 pack should include:

- Pack metadata: pack id, title, game system, edition, source PDF name, generator version, created timestamp, language, and checksum.
- Document structure: sections, headings, page ranges, and optional table of contents data when extraction can find it.
- Text chunks: stable chunk ids, normalized text, page references, section references, and source citation labels.
- Embeddings: one vector per searchable chunk, using the selected local embedding model and recorded dimensions.
- Import manifest: schema version, chunk count, embedding count, and compatibility flags.

The repository must not contain copyrighted PDF contents or generated packs from commercial books unless the user explicitly adds private local artifacts outside project source.

## Pack Builder CLI

The pack builder should be a local PC Python CLI in a future `pack-builder/` directory.

Responsibilities:

- Accept one or more user-owned PDFs as input.
- Extract text with page boundaries preserved.
- Normalize whitespace and split text into citeable chunks.
- Generate embeddings.
- Write a portable pack archive that the Android app can import.
- Record enough metadata to rebuild or debug the pack later.

The first version should optimize for reliable text extraction and citations over perfect layout recovery.

## Retrieval Pipeline

Assistant requests and search requests should share a retrieval pipeline.

1. Normalize the user query.
2. Search SQLite FTS for exact and keyword matches.
3. Search the local vector index for semantic matches.
4. Merge and rank results from notes, lore entities, session records, and sourcebook chunks.
5. Build a compact context bundle with citations and source labels.
6. Pass the context bundle to the assistant when the user asks for AI help.
7. Require the assistant UI to show cited notes or sourcebook chunks when retrieved context was used.

The database is the app's memory. The local model should answer from retrieved material instead of pretending to know sourcebooks or campaign continuity from its own weights.

## Local AI Adapter Strategy

The app should define an `AiEngine` adapter before integrating any specific runtime. The adapter should support:

- Model discovery from app-private storage or imported model directories.
- Model metadata inspection.
- Load, unload, and active model status.
- Prompt execution with cancellation.
- Streaming tokens when supported.
- Benchmark collection for load time, memory use, tokens per second, context length, and qualitative notes.

The first runtime target is `llama.cpp` on Android with GGUF models. Other runtimes such as MLC LLM, LiteRT, or LEAP can be evaluated later behind the same adapter.

Initial benchmark candidates:

- LFM2.5-350M
- SmolLM2-360M-Instruct
- SmolLM2-135M-Instruct
- MobileLLM and MobileLLM/ParetoQ variants
- Qwen3-0.6B as an upper quality benchmark

V1 should treat 350M-class models as the practical design center. Models above 1B parameters should be considered experiments, not product requirements, on the target tablet.

## Tablet Constraints

The target Android tablet profile has useful storage but limited CPU and RAM for local AI.

Design implications:

- Keep imports resumable and avoid long blocking UI work.
- Prefer compact chunks and indexes over large in-memory structures.
- Load only one local generation model at a time.
- Keep context windows modest and retrieval bundles focused.
- Make assistant actions manually triggered, not constantly running in the background.
- Store model files and large packs in locations that can support microSD workflows where Android permissions allow it.
- Benchmark real performance on device before promoting any model as the default.

## Acceptance Scenarios

The architecture is ready for implementation when it supports these scenarios:

- The user imports a prebuilt sourcebook pack from local storage or microSD.
- The user creates campaign notes and links them to system/sourcebook material.
- The user searches by exact terms and semantic meaning.
- The user asks for session brainstorming and receives an answer grounded in retrieved lore.
- The app remains useful with internet disabled.
- Benchmarks capture local model load time, memory use, tokens per second, and answer usefulness.
