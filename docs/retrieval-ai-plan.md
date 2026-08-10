# Retrieval and Local AI Plan

## Metadata

- File: `docs/retrieval-ai-plan.md`
- Created: 2026-08-10
- Last updated: 2026-08-10
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

Initial implementations:

- `FakeAiEngine` for UI and tests.
- `LlamaCppAiEngine` for `llama.cpp` on Android.

Potential later implementations:

- MLC LLM.
- LiteRT.
- LEAP.
- Optional cloud engine outside v1.

## Model Strategy

The first runtime target is `llama.cpp` with GGUF models.

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
- Use fake AI responses for deterministic UI tests.
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
