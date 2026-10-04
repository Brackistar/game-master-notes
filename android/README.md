# Android App

## Metadata

- File: `android/README.md`
- Created: 2026-08-10
- Last updated: 2026-10-03
- User: brackistar

This directory contains the native Android app for `game-master-notes`.

The first implementation uses a native Android multi-module structure so product workflows, shared foundations, storage, retrieval, pack import, and AI runtime boundaries stay explicit from the start.

## Current Working Slice

- Users choose a folder through Android's Storage Access Framework.
- The app persists read access to that folder and scans immediate `.gmnpack` children on startup.
- Imported packs are indexed into Room tables for packs, documents, chunks, and FTS search rows.
- Library and Home show indexed pack state.
- Ask the Books retrieves chunks with SQLite FTS, extracts a compact cited evidence brief, and exposes answer-style and model selectors.
- The assistant checks device RAM, supported ABI, and installed model files on startup. For now, the UI only shows installed LFM2.5 350M GGUF models plus the deterministic grounded fallback.
- Qwen, Gemma, and Phi adapter profiles remain in the `core:ai` compatibility layer for future work, but they are not exposed in the current selector.

The real local model path now uses llama.cpp through the `core:ai` native bridge. The app packages an `arm64-v8a` native runtime and discovers supported GGUF model files under the app private files directory at `files/models/`:

- `lfm2.5-350m-tiny-q2.gguf`
- `lfm2.5-350m-tiny-q3.gguf`
- `lfm2.5-350m-fast-q4.gguf`
- `lfm2.5-350m-q4.gguf`

Only installed files appear in the selector; missing placeholder rows are intentionally hidden. Deferred from this slice: vector search over `embeddings.npy`, campaign/note retrieval, direct model downloads, and broader model-file management. GGUF model files are intentionally not committed.

The local LLM path is optimized for tablets by using hybrid RAG:

1. SQLite FTS retrieves candidate sourcebook chunks for assistant questions.
2. Kotlin-side relevance filtering rejects weak partial matches.
3. A deterministic evidence extractor selects up to 4 cited source blocks, preserving up to 3 relevant paragraphs per block and falling back to a nearby sentence window when source text has no paragraph breaks.
4. The local LLM receives a compact multi-block evidence brief and is prompted to synthesize a direct 2-4 paragraph answer with citations instead of returning a single fragment.
5. If the local model output is too short, malformed, citationless, or times out, the deterministic grounded responder displays the cited excerpts directly.

The assistant now keeps the preceding user question in mind for short follow-ups such as "What about the second type?" while preserving the user's original wording in the conversation. It supports lookup, explain, summarize, brainstorm, and compare modes. When retrieval finds nothing, the app explains how to reformulate the request and skips model loading. Generated answers must contain usable prose and citations supported by the retrieved evidence; source blocks remain expandable below each answer for inspection.

This keeps an LLM in the answer path while reducing prompt size, prompt-evaluation time, and hallucination risk. Vector search from pack embeddings is the next retrieval upgrade, not part of the current Android runtime path.

## Importing AI Models

The app supports manual LFM2.5 GGUF model import from the Ask the Books screen:

1. Download a compatible GGUF model file to the Android device.
2. Open **Ask the Books**.
3. Tap **Import LFM2.5 350M Q2 GGUF**.
4. Select the downloaded `.gguf` file.
5. The app copies it into its private `files/models/` directory as `lfm2.5-350m-tiny-q2.gguf`.

Recommended test downloads:

- LFM2.5 350M: `LiquidAI/LFM2.5-350M-GGUF`, file `LFM2.5-350M-Q4_K_M.gguf`.
  https://huggingface.co/LiquidAI/LFM2.5-350M-GGUF
- For the slowest tablets, test `mradermacher/LFM2.5-350M-GGUF`, file `LFM2.5-350M.Q2_K.gguf` or `LFM2.5-350M.Q3_K_S.gguf`.
  https://huggingface.co/mradermacher/LFM2.5-350M-GGUF
  Direct Q2 download: https://huggingface.co/mradermacher/LFM2.5-350M-GGUF/resolve/main/LFM2.5-350M.Q2_K.gguf
  Direct Q3_K_S download: https://huggingface.co/mradermacher/LFM2.5-350M-GGUF/resolve/main/LFM2.5-350M.Q3_K_S.gguf
- Import the Q2 file first. The compatibility layer can recognize additional LFM2.5 quantization slots later, but current app testing should center on Q2.

The app stores imported files internally as:

- `lfm2.5-350m-tiny-q2.gguf`

Do not commit downloaded model files. Check each model license before redistribution or publishing builds that automate downloads.

## Structure

- `app/` - Android application module and navigation host.
- `core:design` - Compose theme and shared UI styling.
- `core:domain` - durable domain identifiers and models.
- `core:data` - Room database and repository implementation boundary.
- `core:importpacks` - sourcebook pack inspection and import boundary.
- `core:retrieval` - search and context retrieval boundary.
- `core:ai` - local AI runtime adapter boundary, device capability checks, LFM2.5/Qwen3/Gemma3/Phi4 model profiles, model availability, and the llama.cpp JNI bridge.
- `feature:home` - home and entry workflows.
- `feature:library` - systems, campaigns, notes, lore, and tags.
- `feature:import` - sourcebook pack import UI.
- `feature:assistant` - grounded assistant UI. It depends on the `core:retrieval` interface instead of the concrete Room repository.

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

The Android build pins its toolchain and dependency versions in `gradle/libs.versions.toml`: Gradle 8.14.3, Android Gradle Plugin 8.13.2, Kotlin 2.3.21, the matching Compose compiler plugin, KSP 2.3.12, and Room 2.8.5. This deliberately uses the explicit Kotlin Android plugin rather than AGP 9 built-in Kotlin while validating the Room/KSP toolchain. CI and local builds use JDK 17. The Android Gradle Plugin, Kotlin, KSP, Room, Gradle wrapper, SDK, NDK, and CMake versions must be upgraded as a tested set; independently selecting each latest version is not a compatibility strategy.

CI installs Android API 35, NDK 28.2.13676358, and CMake 3.22.1, then runs unit tests and assembles the debug APK with Gradle deprecation warnings treated as errors. Local builds require those same SDK packages and the llama.cpp submodule. Native packaging is restricted to `arm64-v8a`; 32-bit Android devices only see the deterministic fallback model.
