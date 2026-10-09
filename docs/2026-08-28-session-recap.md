# Session Recap: Android Local AI and Retrieval

## Metadata

- File: `docs/2026-08-28-session-recap.md`
- Created: 2026-08-28
- Last updated: 2026-08-28
- User: brackistar

## Summary

The last development session moved the Android app from a deterministic-only sourcebook assistant toward a working offline local-AI slice. The app now supports installed LFM2.5 GGUF model discovery, manual model import, llama.cpp-backed generation, deterministic fallback behavior, and stricter sourcebook retrieval.

## Implemented State

- Added a llama.cpp Android native bridge under `core:ai`.
- Added `LlamaCppLocalModelRuntime`, `ModelSelectingAiEngine`, model profiles, device capability checks, and model-file import support.
- Focused the current model UI on installed LFM2.5 350M GGUF files plus the deterministic grounded fallback.
- Hid missing Qwen, Gemma, Phi, and other future model placeholders from the current selector.
- Added Logcat telemetry for assistant requests, retrieval, model loading, prompt evaluation, generation, timeouts, and output-quality fallback decisions.
- Tuned native llama.cpp execution for the target tablet: `arm64-v8a`, 16 KB page-size compatibility, optimized native build flags, greedy sampling, compact context, and bounded answer budgets.
- Added output-quality guards so short, malformed, citationless, or timed-out model output falls back to readable grounded excerpts.
- Improved sourcebook pack scanning to find `.gmnpack` files in nearby subfolders.
- Hardened pack archive reads with bounded member sizes.
- Tightened retrieval to reject weak partial matches.
- Changed retrieved evidence from raw chunk prefixes or isolated sentences to useful paragraph-sized cited excerpts.

## Current Retrieval Format

Assistant retrieval currently uses SQLite FTS over imported sourcebook chunks. The user question is normalized into meaningful terms, common filler words are removed, and multi-term questions use stricter matching to avoid irrelevant partial hits.

Matching chunks are then processed in Kotlin:

- Prefer the best matching paragraph when paragraph breaks are present.
- Fall back to a small sentence window around the best sentence when paragraph breaks are not available.
- Clip excerpts cleanly at word boundaries.
- Preserve normalized citation labels in the form `Book Title, pp. start-end`.

The prompt evidence format is:

```text
Evidence:
1. [Book Title, pp. 10-11] Useful paragraph excerpt.

2. [Book Title, pp. 42-43] Another useful paragraph excerpt.
```

The deterministic fallback renders the same evidence as readable separated excerpts:

```text
I found these relevant passages in the loaded books:

[Book Title, pp. 10-11]
Useful paragraph excerpt.

[Book Title, pp. 42-43]
Another useful paragraph excerpt.
```

## Current Local AI Format

The local AI path routes assistant questions through the selected `AiEngine`. LFM2.5 currently uses a plain prompt template rather than a guessed chat-template wrapper:

```text
Answer only from the evidence. Be clear and human. If evidence is weak, say so. Cite sources like [Book, p. 1].

Evidence:
1. [Book Title, pp. 10-11] Useful paragraph excerpt.

Question: user question
Answer:
```

This keeps the model in the answer path while limiting prompt-evaluation cost on low-RAM tablets.

## Validation

Recent validation commands passed:

```powershell
.\gradlew.bat :core:data:testDebugUnitTest
.\gradlew.bat :core:ai:testDebugUnitTest :feature:assistant:compileDebugKotlin
.\gradlew.bat :core:ai:externalNativeBuildDebug :feature:assistant:compileDebugKotlin
```

Windows builds may print Kotlin daemon `AccessDeniedException` warnings and continue through Gradle's fallback compiler path.

## Open Next Steps

- Add hybrid retrieval with local vectors from `embeddings.npy`.
- Add sourcebook embedding storage and vector similarity search behind `core:retrieval`.
- Combine strict FTS candidates with semantic vector candidates before evidence extraction.
- Optionally let the local LFM model rerank only a very small candidate set.
- Add campaign notes and lore entity retrieval.
- Add model benchmark UX for load time, prompt-evaluation time, generation speed, timeout rate, and answer quality.
