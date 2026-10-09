# Retrieval and Local AI Plan

## Metadata

- File: `docs/retrieval-ai-plan.md`
- Created: 2026-08-10
- Last updated: 2026-08-28
- User: brackistar

Related diagram: [retrieval-ai-plan.mmd](retrieval-ai-plan.mmd)

## Purpose

Retrieval and local AI turn the app from a static notebook into a grounded session assistant. The retrieval pipeline finds relevant notes, lore, and sourcebook chunks. The local model uses that context to brainstorm, summarize, and help direct sessions.

## Product Goals

- Answer from local data, not from model memory.
- Combine keyword and semantic search.
- Show citations when sourcebook chunks or notes are used.
- Keep local AI manually triggered.
- Support multiple local model candidates through an adapter.
- Benchmark real performance on low-requirement Android tablets.

## Retrieval Pipeline

The pipeline should be shared by search and assistant workflows:

1. Normalize the query.
2. Run SQLite FTS search.
3. Run semantic vector search.
4. Merge and rank results.
5. Apply campaign, system, sourcebook, and tag filters.
6. Build a compact context bundle.
7. Pass the context bundle to the assistant or render it in search results.

Current Android status:

- Sourcebook chunk retrieval is implemented with SQLite FTS.
- Query normalization keeps meaningful terms and removes common filler words.
- FTS matching is intentionally strict for multi-term questions to avoid weak partial matches.
- Retrieved chunks are post-filtered in Kotlin and converted into up to four cited source blocks, preserving multiple relevant paragraphs within each block.
- If paragraph breaks are unavailable, retrieval falls back to the best matching sentence with nearby context.
- Natural-language follow-up questions reuse the previous user question during retrieval, while the original wording remains visible in chat.
- The assistant supports lookup, explanation, summarization, brainstorming, and comparison modes; empty retrieval produces a recovery message without loading or invoking the local model.
- Vector search is still planned; `embeddings.npy` is imported only as metadata today.

## Ranking Inputs

Ranking should consider:

- Keyword score.
- Vector similarity.
- Current campaign/session relevance.
- Pinned lore.
- Recency for user notes.
- Source type, such as note, lore entity, or sourcebook chunk.
- Citation quality, such as page availability.

V1 ranking can be simple and transparent. Avoid complex ranking until there is real usage data.

## Context Bundle

The context bundle passed to local AI should include:

- User request.
- Active campaign and session names.
- Retrieved notes.
- Retrieved lore entities.
- Retrieved sourcebook chunks.
- Citation labels and source ids.
- Style or instruction hints, if the user defines them later.

The assistant prompt should explicitly say when context is insufficient and should avoid inventing citations.

The current assistant evidence brief is compact and citation-first:

```text
Evidence:
1. [Book Title, pp. 10-11] Useful paragraph excerpt.

2. [Book Title, pp. 42-43] Another useful paragraph excerpt.
```

The deterministic fallback renders the same evidence as separated cited excerpts for readability.

Generated answers are accepted only when they contain usable prose, supported citations, and evidence-related wording. Otherwise the deterministic cited evidence response is shown. Chat answers keep their source blocks expandable so the user can inspect the exact retrieved text without leaving the conversation.

## AiEngine Adapter

Define an `AiEngine` interface before integrating a real runtime.

Required capabilities:

- List available models.
- Read model metadata.
- Load one model.
- Unload current model.
- Generate a response from a prompt.
- Cancel an active generation.
- Stream tokens when supported.
- Report runtime status and benchmark metrics.

Current implementations:

- `GroundedMvpAiEngine` for deterministic cited fallback answers.
- `ModelSelectingAiEngine` for discovering installed compatible local models.
- `LlamaCppLocalModelRuntime` for `llama.cpp` on Android.

Potential later implementations:

- MLC LLM.
- LiteRT.
- LEAP.
- Optional cloud engine outside v1.

## Model Strategy

The first runtime target is `llama.cpp` with GGUF models and is implemented for `arm64-v8a` devices. The current UI focuses on installed LFM2.5-350M files and hides missing model placeholders. Qwen, Gemma, and Phi profiles remain available in the compatibility layer for later experiments.

Benchmark candidates:

- LFM2.5-350M
- SmolLM2-360M-Instruct
- SmolLM2-135M-Instruct
- MobileLLM and MobileLLM/ParetoQ variants
- Qwen3-0.6B as an upper quality benchmark

The expected design center is 350M-class local models. The product should not depend on 1B+ models for v1.

## Assistant Workflows

V1 assistant workflows:

- Brainstorm scene complications from active session context.
- Suggest NPC motives from notes and sourcebook chunks.
- Summarize retrieved lore.
- Find continuity conflicts.
- Draft a quick description, rumor, clue, or consequence.
- Help interpret retrieved rules text without claiming authority beyond citations.

Assistant output should be saveable as a note.

## Benchmarking

Benchmark runs should record:

- Device model.
- Android version.
- Model id and file size.
- Quantization.
- Load time.
- Prompt token estimate.
- Output token estimate.
- Tokens per second.
- Peak or observed memory notes when available.
- User quality rating.
- Freeform notes about usefulness.

## Testing

- Unit test query normalization.
- Unit test result merging and ranking.
- Unit test context bundle construction.
- Test assistant prompt generation with fixed retrieved inputs.
- Use deterministic fallback responses for reliable tests.
- Manually test real local models on the target tablet.

## Risks

- Small models may be weak without carefully assembled context.
- Long sourcebook passages can overflow practical context limits.
- Local generation may block UI if cancellation and threading are wrong.
- Model file management can become confusing for users.

## Suggestions

- Build retrieval first and local AI second.
- Always show the retrieved context before or beside assistant answers.
- Make "save to notes" a primary action.
- Keep prompts short and heavily grounded.
- Use benchmark results to choose defaults, not popularity lists.
