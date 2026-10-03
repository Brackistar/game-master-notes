# Offline AI Assistant POC Review

- Created: 2026-09-29
- Last updated: 2026-09-29
- Author/role: Codex, AI/LLM and AI-application specialist
- File: `docs/2026-09-29-ai-assistant-poc-review.md`
- Branch reviewed: `development-alt`
- Scope: first-party Android assistant, retrieval/data layers, pack-builder, prompts, model profiles, llama.cpp JNI runtime, output validation, telemetry, and tests

## Executive decision

The tablet result should be treated as a full RAG-chain failure, not as evidence that the generation timeout alone is too short. The current system can send weakly ranked literal matches through the wrong prompt format to an aggressively quantized 350M model, then accept an answer after only checking its length, citation spelling, and a small amount of word overlap. A longer timeout would let the same chain produce a longer bad answer.

The minimal-change POC path is:

1. Make inference correct and observable: apply the GGUF model's chat template, use the model-card decoding settings, return structured stop/timing data from JNI, and validate answers against stable evidence IDs.
2. Improve retrieval before increasing prompt size: rank FTS candidates, activate the already-packaged vectors with the same query encoder, fuse lexical and semantic ranks, and expand only adjacent/related chunks.
3. Use LFM2.5-350M Q4 as the low-end baseline, not Q2. Gate an LFM2.5-1.2B-Instruct Q4 experiment behind measured device memory and latency. It stays in the existing llama.cpp/LFM architecture while providing a meaningful capability increase.
4. Increase context, output, and time only within measured bounds. Keep deterministic cited excerpts as the terminal fallback.

No current source proves which failure dominated the reported tablet run because no assistant Logcat capture is present. The repository's `debug.log` contains only an unrelated Windows Crashpad access error. The causes below are therefore ranked from direct code evidence, not reconstructed from that exact request.

## Current answer path

```text
question + one-turn follow-up heuristic
  -> FTS4 MATCH (strict AND, then relaxed OR)
  -> first rows by FTS rowid
  -> Kotlin substring scoring and paragraph extraction
  -> at most 4 evidence blocks / about 1,800 characters
  -> hand-built plain prompt
  -> LFM2.5-350M, usually Q2, greedy decode
  -> 1,024-token context / 96 output tokens / 30-second native limit
  -> shallow citation and keyword validation
  -> answer, or deterministic evidence dump
```

The module boundaries are sound for a POC. `RetrievalRepository`, `AiEngine`, `LocalModelRuntime`, Room, and the pack-builder should remain. The improvements fit behind those boundaries; a new framework or agent stack is unnecessary.

## Failure diagnosis

### 1. The active LFM prompt format is incompatible with the model instructions - critical, high confidence

`LocalModelProfiles.kt` assigns `PromptStyle.Plain` to every LFM2.5 profile, and `PromptTemplates.kt` consequently emits ordinary prose ending in `Answer:`. Liquid AI specifies a ChatML-like template beginning with `<|startoftext|>` and role tokens, and its examples use `apply_chat_template(..., add_generation_prompt=True)`. llama.cpp exposes the embedded GGUF chat template through `llama_model_chat_template()` and applies it through `llama_chat_apply_template()`.

Instruction-tuned small models are especially sensitive to role and generation-boundary tokens. A plain completion prompt can cause prompt echo, continuation of evidence, malformed citations, or text that is locally fluent but not an answer. This must be corrected before model-size or timeout conclusions are trusted.

Recommendation: pass structured system/user messages across the runtime boundary and apply the template stored in GGUF metadata in native code. Treat a missing/unsupported template as a model compatibility failure. Keep a profile-specific golden template only as a tested fallback, not as the primary path.

### 2. Retrieval truncates candidates before relevance ranking - critical, high confidence

`SourcebookDao.searchChunks()` returns `0.0 AS rank`, orders by FTS `rowid`, and applies `LIMIT` before Kotlin scoring. With an assistant limit of 4 and multiplier 3, only the first 12 rowid matches are examined. A better passage later in the corpus is invisible. Strict `AND` can have high precision but low recall; relaxed `OR` can admit distractors, and neither branch uses collection-level relevance.

SQLite documents `matchinfo()` for FTS4 relevance scoring and `rank`/BM25 for FTS5. The least disruptive short-term fix is to rank a bounded FTS4 candidate set using `matchinfo()` through a registered ranking function or migrate the database to FTS5 and `ORDER BY rank`. Because the database is version 1 and has no migrations, either choice needs an explicit migration/reindex test; silently replacing the schema is not acceptable.

### 3. Semantic retrieval is packaged but absent - critical, high confidence

The pack-builder writes normalized 384-dimensional `all-MiniLM-L6-v2` vectors and records `embedding_row_index`, but Android skips `embeddings.npy` and stores only its metadata. Consequently, paraphrases, aliases, concepts, and indirect relationships can only succeed when important words overlap literally.

The vector and query encoders must be identical, including tokenizer, truncation, pooling, and normalization. The official model card says `all-MiniLM-L6-v2` is intended for semantic search and truncates input beyond 256 word pieces. Current chunks allow 1,800 characters, so some embeddings may represent only the beginning of a long chunk. Before relying on these vectors, evaluate chunk token lengths and target roughly 120-220 word pieces per embedding unit, preserving page/section provenance. Do not merely raise the chunk size.

For the POC, extract validated vectors to an app-private per-pack file and memory-map or stream them. Brute-force cosine similarity over a modest corpus is simpler and sufficient to establish quality; hide it behind `core:retrieval` so an ANN index can replace it later. Use a compact on-device version of the same MiniLM query encoder. Union lexical top 20 and vector top 20, then fuse ranks and retain approximately 8 candidates before evidence selection.

### 4. Q2 350M is the default quality path - high, high confidence

The UI imports and replaces `lfm2.5-350m-tiny-q2.gguf`, and model discovery puts the Q2 profile first. Q2 is an emergency footprint/speed tradeoff, not a good default for coherent synthesis. Liquid AI describes the 350M model as suitable for extraction and structured output but explicitly not recommended for knowledge-intensive tasks. RAG removes the need for memorized book knowledge, but intent interpretation, comparison, and multi-passage synthesis still require reasoning capacity.

Use this order for experiments:

1. LFM2.5-350M Q4_K_M with the correct template and decoding settings: low-end baseline.
2. LFM2.5-1.2B-Instruct Q4-class GGUF: preferred quality candidate when the target tablet passes memory, latency, and thermal gates.
3. Q2/Q3 350M: explicit compatibility or speed modes, never silently preferred over Q4.

The 1.2B candidate is a minimal architectural change because it uses the same LFM2.5 family, chat format, GGUF path, and llama.cpp runtime. The vendor reports sub-1 GB operation, but Android eligibility must use measured app PSS/native RSS, memory class, low-memory state, and a repeated-query stress test rather than total physical RAM alone.

### 5. Decoding does not follow the model card - high, high confidence

The JNI bridge installs only a greedy sampler. Liquid AI recommends temperature `0.1`, top-k `50`, and repetition penalty `1.05` for LFM2.5. Greedy decoding is reproducible, but it can lock a tiny model into repetitive or malformed continuations. Implement the recommended sampler chain with a fixed seed for evaluation. Keep greedy as an A/B baseline; do not introduce high temperature or broad creative sampling into grounded QA.

### 6. Output budgets conflict with the requested answer shape - high, high confidence

The prompt asks for 2-4 short paragraphs while native generation allows 96 tokens. The 1,024-token context also reserves those 96 tokens and may trim from the beginning of the fully formatted prompt. Native trimming deletes leading tokens without preserving system instructions or message boundaries. That can remove the grounding rules and opening evidence.

Budgeting must happen before template application. Select evidence to fit a measured token budget and reject an oversized prompt; never trim arbitrary leading tokens after tokenization. A small POC answer should be 1-3 concise paragraphs or bullets, not 2-4 paragraphs regardless of evidence.

### 7. Timeout state is lost and increasing the UI timeout alone has little effect - high, high confidence

There are two limits: 30 seconds inside native generation and 40 seconds around load plus generation in the ViewModel. Raising only the outer value leaves the native limit unchanged. Native generation returns partial text when its loop reaches the timeout, but Kotlin receives only a string, so a partial timed-out answer can be accepted as normal. Prompt-evaluation timeout throws while generation timeout returns partial output. The semantics are inconsistent.

Return a structured native result containing prompt tokens, generated tokens, stop reason, prompt-evaluation time, first-token latency, generation time, and token rates. Treat `timeout`, `cancelled`, decode error, context overflow, and EOG as distinct outcomes. A timed-out partial answer may be shown only after the same grounding validation and with an explicit incomplete marker; otherwise use the deterministic fallback.

### 8. Grounding validation can approve nonsense - high, high confidence

`LocalModelAiEngine` replaces `AiRequest.prompt` with the entire formatted prompt before `LlamaCppLocalModelRuntime` calls `validateGroundedAnswer()`. The validator therefore checks output coverage against terms from instructions and evidence as well as the original question. It requires only one recognized citation somewhere and does not verify that each claim is supported. `extractCitationIds()` also reads at most the first bracket pair per line. Finally, the UI attaches every retrieved result to the message regardless of which sources the model actually cited.

Use stable short evidence IDs (`E1`-`E4`) in the model prompt and map them to source IDs/citation labels in the UI. Require every factual sentence or bullet to end with one or more valid evidence IDs. Reject unknown IDs, uncited factual sentences, prompt/control-token leakage, excessive repetition, and claims with no meaningful token/number/name overlap with the cited evidence. This is not full entailment, but it is a deterministic, inspectable POC guard. Preserve the original user question separately from the rendered prompt.

### 9. Relationship-aware retrieval has no data representation - medium, high confidence

The current chunk schema contains document, page range, text, and embedding row only. It lacks section path, aliases, entity mentions, adjacency links, and typed relationships. The one-turn query planner only concatenates a previous question for a few surface patterns.

Do not build a full knowledge graph for the first POC. After hybrid retrieval, add cheap expansion from existing data: neighboring chunks in the same document, overlapping page ranges, repeated named phrases, and same-section chunks once section metadata exists. For a later backward-compatible pack schema revision, add optional `section_path`, `entity_names`, `aliases`, and relation records with source chunk IDs. Correlation should remain retrieval work; do not ask the 350M generator to scan the corpus or invent a graph.

### 10. Tests prove mechanics, not answer quality - medium, high confidence

Current tests cover literal FTS behavior, paragraph extraction, evidence formatting, citation rejection, model selection, follow-up concatenation, chunking, and embedding shape. They do not execute a real GGUF, compare formatted token sequences, inspect stop reasons, test semantic retrieval, test late-row relevance, detect unsupported cited claims, or score end-to-end answers. All existing focused tests pass, but they cannot catch the reported failure class.

## Bounded POC settings

These are starting ranges to benchmark, not hard-coded promises.

| Control | Current | POC default | POC upper bound | Guidance |
|---|---:|---:|---:|---|
| Model | LFM2.5-350M Q2 | 350M Q4_K_M | 1.2B Instruct Q4-class if gated | Increase capability by measured model profile, not filename alone. |
| Context | 1,024 tokens | 1,536 tokens | 2,048 tokens | Do not jump to the model's 32K maximum on a 4 GB tablet. |
| Output | 96 tokens | 144 tokens | 192 tokens | Enough for a concise grounded answer; longer output raises latency and hallucination exposure. |
| Native generation deadline | 30 s | 60 s | 90 s in "Useful and slow" mode | Separate model load, prompt evaluation, and decode timing. |
| Outer request deadline | 40 s | native deadline + 10-15 s | 105 s | It must exceed, not replace, the native deadline. |
| Evidence blocks | 4 | 3-4 | 5 only if token budget permits | More passages can reduce small-model focus. |
| Fused candidate pool | 12 rowid matches | lexical 20 + vector 20 -> fused 8 | 50 per retriever for offline evaluation | Candidate recall grows before prompt size does. |
| Threads | 1-2 | 2 | 3-4 only after device benchmark | More threads can increase heat/contention and hurt sustained latency. |
| Prompt batch | 32 | 64 | 128 if memory is stable | Benchmark prompt tokens/s and native RSS. |
| Sampling | greedy | temp 0.1, top-k 50, repeat 1.05, fixed seed | remain low-temperature | Do not increase creativity for book-grounded answers. |

Knobs that should not be blindly increased: context to 32K, evidence count, chunk size, thread count, temperature, timeout without a native stop reason, or model size without PSS/thermal measurements. None of these repairs poor retrieval or an incorrect chat template.

## Staged minimal-change implementation plan

### Stage 0 - Reproduce and instrument

- Add a request/run ID shared across UI, retrieval, prompt building, Kotlin runtime, and JNI.
- Capture the exact model profile, GGUF metadata/template availability, pack fingerprints, retrieval ranks, evidence IDs, token counts, sampler settings, stop reason, and answer-validator reason.
- Add a user-triggered diagnostics export. Default logs must not include question text or copyrighted excerpts; use hashes, stable source/chunk IDs, lengths, and scores. An explicit developer option may include redacted text for local-only troubleshooting.
- Keep a bounded rotating JSONL journal in app-private storage, for example three 1 MB files or the last 100 requests. Logcat remains a live view, not the only evidence store.
- Re-run the failed question on the tablet with Q2 and Q4, correct/incorrect templates, and identical retrieved evidence. This isolates retrieval, formatting, quantization, and time.

Exit condition: one exported record can reconstruct every decision in the answer loop without exposing book text by default.

### Stage 1 - Correct generation without changing retrieval architecture

- Apply the GGUF chat template through llama.cpp and add a native golden test against the official LFM2.5 template.
- Preserve `originalQuestion`, structured evidence, and rendered prompt as separate values.
- Implement structured generation results and the recommended low-temperature sampler chain.
- Replace generated full citation labels with `E1`-style IDs and strengthen deterministic validation.
- Make 350M Q4 the recommended baseline; expose Q2 as a clearly labeled low-quality compatibility mode.
- Raise budgets only to the default row in the table above, then benchmark.

Exit condition: zero prompt echoes/control tokens and zero invalid citations across the evaluation set; timed-out output can never appear as a normal complete answer.

### Stage 2 - Hybrid retrieval

- Rank lexical candidates before limiting them. Prefer FTS5 `rank`/BM25 if Room/database compatibility is verified; otherwise implement tested FTS4 `matchinfo()` ranking first.
- Import and validate `embeddings.npy`: dtype, dimensions, row count, finite values, row-index uniqueness, and model ID.
- Add the matching MiniLM query encoder behind a small interface in `core:retrieval` or `core:ai`; do not couple it to Compose.
- Retrieve lexical and vector candidates independently, fuse by reciprocal rank fusion initially, then tune a weighted score only after labeled examples exist. Research shows fusion choice is data-dependent; RRF is a starting baseline, not a universal optimum.
- Expand top candidates by immediate document neighbors and deduplicate overlapping evidence. Rerank to 3-4 final blocks with lexical coverage, cosine score, proximity, source diversity, and token cost.
- Rebuild POC packs with embedding-sized chunks or embedding subunits that stay below MiniLM's 256-word-piece truncation limit.

Exit condition: paraphrase and relationship-query recall meets the metrics below without sending additional unrelated text to the LLM.

### Stage 3 - Capability profile and relationship metadata

- Add an LFM2.5-1.2B-Instruct Q4-class profile using the same runtime and template path.
- Gate availability using measured memory class, current low-memory state, native model-load success, and a benchmark, not total RAM alone.
- Add optional section/entity/alias/relation metadata in a backward-compatible pack schema revision only after hybrid retrieval's residual failures are classified.
- Keep deterministic query expansion and relation traversal bounded to one or two hops. The final LLM receives only selected evidence.

Exit condition: the 1.2B profile produces a statistically meaningful quality gain on the same retrieved evidence and remains inside tablet latency, memory, and thermal gates.

## Telemetry contract

One event record per stage should contain:

- Request: run ID, timestamp, app/build version, answer mode, question hash/length/language, follow-up flag.
- Device: ABI, Android version, low-RAM flag, memory class, available memory, PSS/native RSS before load, after load, and peak when practical, battery/thermal status.
- Model: exact profile, model file size and content fingerprint, quantization, GGUF template hash/name, context, batch, threads, seed, and sampler chain.
- Retrieval: normalized-query hash, retriever name, elapsed time, candidate count, each candidate's stable chunk ID, lexical/vector rank and score, fusion rank, expansion reason, and final evidence ID/token count.
- Generation: prompt tokens before/after budgeting, whether trimming was attempted, prompt-evaluation milliseconds and tokens/s, first-token latency, generated tokens, decode tokens/s, total milliseconds, and stop reason.
- Validation: cited evidence IDs, rejection/fallback reason, repetition score, unsupported-sentence count, answer length, and whether the displayed sources exactly match cited IDs.
- Outcome: success/fallback/no-evidence/error/cancel, user rating when provided, and diagnostics schema version.

Never log full questions, prompts, answers, model paths, or sourcebook excerpts by default. Logs should be useful through stable identifiers and fingerprints; content capture must be explicit, local, bounded, and easy to delete.

## Synthetic evaluation corpus

Create two small, original `.gmnpack` books plus distractor passages. No commercial text is needed.

Suggested corpus:

- `Ashfall Field Guide`: factions (Ember Wardens, Glass Cartographers), places (Cinder Gate, Moonwell), rules (stress, rest, ward keys), aliases, numeric limits, and an explicit exception.
- `Tides of Veyra`: characters and organizations linked through two-hop relationships, two editions with a changed rule, and similarly named distractors.
- Include headings, adjacent chunks, relevant passages after many earlier rowids, paraphrases with no important shared nouns, conflicting statements with edition labels, and passages that do not answer a plausible question.

Required cases:

1. Exact lookup: direct term and one source.
2. Paraphrase: "recover strain" when the book says "remove stress after a full rest."
3. Alias: a title or nickname maps to the canonical entity.
4. Two-hop relation: who controls a key used to enter a named place.
5. Neighbor expansion: definition and exception occur in adjacent chunks.
6. Compare: two factions across separate passages.
7. Numeric rule: preserve numbers, units, and exceptions.
8. Negation: distinguish "cannot" from "can."
9. Edition conflict: answer from the requested edition and surface the conflict.
10. Follow-up: resolve "What about the second one?" from the previous question.
11. Unanswerable: decline with no invented fact or citation.
12. Prompt-injection-like source text: book prose that says to ignore instructions must remain inert evidence.
13. Distractor overload: relevant chunk appears after at least 20 lexical matches.
14. Timeout/cancel: deterministic fallback and responsive cancellation.

Each case should declare acceptable source chunk IDs, forbidden distractors, required facts, forbidden claims, expected answerability, and answer-mode expectations. Run retrieval independently from generation so model and retriever regressions are distinguishable.

## Acceptance gates for the first usable POC

Use at least 50 synthetic questions, including at least 15 paraphrase/relationship cases and 10 unanswerable/adversarial cases.

### Retrieval

- Recall@4: at least 90% on literal cases, 80% on paraphrase/relationship cases, and 85% overall.
- nDCG@4: at least 0.80 overall.
- Required evidence appears in the fused top 8 for at least 95% of answerable cases.
- Unanswerable queries do not produce a high-confidence evidence bundle in at least 90% of cases.

### Answer quality

- 100% of rendered citations map to retrieved evidence and the UI shows only actually cited evidence plus a separately labeled "other retrieved sources" section if desired.
- At least 95% of factual sentences are supported by their cited evidence under deterministic checks plus human review.
- At least 85% exact required-fact coverage on lookup/numeric cases and 80% human-rated useful/correct on compare/explain cases.
- Zero accepted prompt echoes, control-token leaks, obvious word salad, or unsupported edition/number claims.
- 100% deterministic fallback on invalid citation, timeout, cancellation, context overflow, or failed validation.

### Tablet operation

- Zero ANRs, crashes, or low-memory kills in 20 consecutive mixed requests.
- Cancel acknowledgement reaches native code within 1 second and generation stops at the next decode checkpoint.
- Warm answer latency on the target tablet: median at most 45 seconds and p95 at most 90 seconds in "Useful and slow" mode.
- Peak process PSS remains below 75% of the device's app memory limit, with at least 20% available-memory headroom after model load; the exact limit is device-measured.
- Five consecutive requests do not show unbounded latency growth, memory growth, or severe thermal throttling.

Promote the 1.2B model only if it improves useful/correct answer rate by at least 10 percentage points over 350M Q4 on identical evidence, without violating the operational gates.

## Verification performed for this review

- Reviewed first-party code and tests under `android/core/ai`, `android/core/data`, `android/core/retrieval`, `android/core/importpacks`, `android/feature/assistant`, and `pack-builder`; excluded `android/third_party/**` source.
- Confirmed branch `development-alt` and preserved all existing modified/untracked files.
- Android focused tests passed: `:core:data:testDebugUnitTest`, `:core:ai:testDebugUnitTest`, and `:feature:assistant:testDebugUnitTest` (`BUILD SUCCESSFUL`).
- Focused pack-builder tests passed: 6 tests for chunking, embeddings, and schema contract. Pytest emitted only a cache-write permission warning.
- No real-GGUF or tablet end-to-end evaluation was available in this review, so performance settings remain benchmark gates rather than confirmed device values.

## Primary references

- Liquid AI, [LFM2.5-350M model card](https://huggingface.co/LiquidAI/LFM2.5-350M) - intended use, context, generation settings, GGUF support, and ChatML-like template.
- Liquid AI, [LFM2.5-1.2B-Instruct model card](https://huggingface.co/LiquidAI/LFM2.5-1.2B-Instruct) and [official GGUF repository](https://huggingface.co/LiquidAI/LFM2.5-1.2B-Instruct-GGUF) - capability candidate, template, and llama.cpp format.
- llama.cpp, [official simple chat example](https://github.com/ggml-org/llama.cpp/blob/master/examples/simple-chat/simple-chat.cpp) - reading and applying the model's embedded chat template.
- llama.cpp, [server documentation](https://github.com/ggml-org/llama.cpp/blob/master/tools/server/README.md) - sampler configuration and timing/token reporting capabilities.
- SQLite, [FTS3/FTS4 documentation](https://www.sqlite.org/fts3.html) - `matchinfo()` relevance data and ranking guidance.
- SQLite, [FTS5 documentation](https://www.sqlite.org/fts5.html) - BM25 and `ORDER BY rank` behavior.
- Android Developers, [Room FTS4 reference](https://developer.android.com/reference/androidx/room/Fts4) and [Room entity/FTS guidance](https://developer.android.com/training/data-storage/room/defining-data) - current and available Room FTS mappings.
- Sentence Transformers, [all-MiniLM-L6-v2 model card](https://huggingface.co/sentence-transformers/all-MiniLM-L6-v2) - semantic-search use, 384 dimensions, pooling expectations, and 256-word-piece truncation.
- Chen et al., [An Analysis of Fusion Functions for Hybrid Retrieval](https://arxiv.org/abs/2210.11934) - lexical/semantic complementarity and the need to evaluate fusion choices.
- Lewis et al., [Retrieval-Augmented Generation for Knowledge-Intensive NLP Tasks](https://arxiv.org/abs/2005.11401) - original retrieval-plus-generation framing.
- Android Developers, [memory management overview](https://developer.android.com/topic/performance/memory-overview) and [memory fundamentals](https://developer.android.com/topic/performance/memory/guide/concepts) - app memory limits, PSS, and the low-memory performance cliff.
