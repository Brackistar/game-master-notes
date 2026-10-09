# Android Data and Tablet POC Review

## Metadata

- File: `docs/2026-09-29-android-data-poc-review.md`
- Created: 2026-09-29
- Last updated: 2026-09-29
- Author/role: Codex, Android application and data-efficiency specialist
- Branch reviewed: `development-alt`
- Scope: first-party Android application and project documentation
- Excluded: `android/third_party/**`; `llama.cpp` is treated as an external runtime boundary

## Executive Decision

Do not redesign the current module structure for the first proof of concept. Keep Room, FTS4, the existing retrieval interface, `AiEngine`, the `.gmnpack` format, and the JNI boundary. The fastest route to a useful tablet POC is to correct index and citation integrity first, make the current pipeline measurable, then tune model quality and time against a fixed question set. Semantic retrieval should follow as a bounded extension behind `RetrievalRepository`, not as a replacement for FTS.

The tablet's bad answer cannot be reconstructed from the available `debug.log`: it contains only a Windows Crashpad access-denied message and no Android application telemetry. The exact incident cause is therefore unverified. However, the source contains several observed defects that can independently produce weak evidence, permissive answer acceptance, and misleading source cards.

The four highest-priority findings are:

1. Standalone FTS rows are never deleted when a pack is replaced or pruned.
2. FTS candidates are ordered by insertion `rowid`, and every DAO result is assigned rank `0.0`; the candidate limit can therefore exclude better matches before Kotlin reranking.
3. answer-quality validation receives the fully rendered prompt as its `question`, so its relevance check is much easier to pass than intended.
4. the assistant ignores `AiResponse.citationIds` and shows every retrieved result as a citation, whether or not the generated answer cited it.

Increasing only the timeout or output token limit would give the current failure modes more time to run. Correctness and telemetry should land before larger generation budgets.

## Review Basis

### First-party paths inspected

- `android/core/data`: `AppDatabase`, entities, DAO, repository, and repository tests
- `android/core/importpacks`: SAF folder scan, archive parsing, entity mapping, and import transaction path
- `android/core/retrieval`: retrieval request/result contracts
- `android/core/ai`: model profiles, prompt templates, evidence builder, answer guard, model selector, Kotlin/JNI runtime, and native bridge
- `android/feature/assistant`: query planning, ViewModel orchestration, timeout handling, chat rendering, and citations
- `android/feature/import`: pack scan UI and lifecycle
- `android/app`: dependency assembly
- Current README, architecture/planning documents, session recap, local tablet Logcat helper, and first-party audit report

### Authoritative references

- Android recommends that data-layer suspend functions be main-safe and that classes performing blocking or CPU-heavy work select an appropriate, preferably injected, dispatcher: [Best practices for coroutines in Android](https://developer.android.com/kotlin/coroutines/coroutines-best-practices).
- Room supports asynchronous one-shot DAO calls through `suspend` and observable reads through `Flow`: [Write asynchronous DAO queries](https://developer.android.com/training/data-storage/room/async-queries).
- Room's `@Fts4` maps an entity to an SQLite FTS4 virtual table: [Room Fts4 reference](https://developer.android.com/reference/androidx/room/Fts4).
- SQLite documents `matchinfo()` as the FTS3/FTS4 source of metrics for relevance filtering and sorting; plain document/row order is not relevance order: [SQLite FTS3 and FTS4 extensions](https://www.sqlite.org/fts3.html#matchinfo).
- Logcat entries include timestamp, process/thread identifiers, tag, package, priority, and message; these fields support a correlated diagnostic format: [View logs with Logcat](https://developer.android.com/studio/debug/logcat).
- WorkManager is appropriate only if pack import must reliably continue after the screen or process leaves the foreground; ordinary cancellable coroutines remain appropriate for visible, user-scoped work: [Android task scheduling](https://developer.android.com/develop/background-work/background-tasks/persistent).
- Android Macrobenchmark runs end-user workflows on a physical device and emits JSON plus traces. It is suitable after functional telemetry stabilizes, not as a substitute for per-request AI metrics: [Write a Macrobenchmark](https://developer.android.com/topic/performance/benchmarking/macrobenchmark-overview).

## Current POC Data Path

```text
.gmnpack via SAF
  -> ContentResolverPackImporter
  -> in-memory JSON/chunk/entity lists
  -> one Room replace transaction
  -> source_chunks + standalone source_chunks_fts
  -> FTS MATCH, rowid-ordered bounded candidates
  -> Kotlin paragraph/sentence scoring
  -> up to four RetrievalResult values
  -> EvidenceBriefBuilder
  -> model-specific prompt template
  -> llama.cpp generation or deterministic fallback
  -> chat answer plus all retrieved source cards
```

This is a sensible POC shape. The problems are in lifecycle integrity, ranking, contract fidelity, and observability rather than the high-level boundaries.

## Observed Facts

### P0 - FTS lifecycle is inconsistent with pack lifecycle

`SourceChunkFtsEntity` is a standalone FTS4 table with no Room foreign key or external-content relationship. `replaceImportedPack()` deletes the pack and cascades its document/chunk rows, but it never deletes FTS rows. The prune paths similarly delete only `sourcebook_packs`. Reimport and prune therefore leave stale FTS rows behind.

Effects:

- Repeated reimports grow the FTS table indefinitely.
- Old and new FTS rows can point at the same current `chunkId`, producing duplicate joined candidates.
- Stale rows can consume the DAO's small candidate window before current relevant rows are considered.
- If chunk IDs are not globally unique across packs, joining only on `chunkId` can associate a match with the wrong current chunk.

The current replacement unit test searches only for newly inserted text, so it does not detect stale old-term matches, duplicate candidate pressure, or FTS-row growth.

Minimal correction:

- Add explicit FTS deletion by `packId` inside the same replacement transaction before deleting/inserting the pack.
- Delete FTS rows for every pruned pack; do not rely on a cascade that does not exist.
- Define and validate the chunk-ID contract: either globally unique IDs or joins/keys that include `packId`.
- Add an integrity query/test asserting that active FTS rows map one-to-one to active chunks after import, reimport, and prune.

### P0 - The retrieval window is not relevance-ranked

`SourcebookDao.searchChunks()` returns `0.0 AS rank`, orders by FTS `rowid`, and applies `LIMIT` before application scoring. `SourcebookRepository` requests only `resultLimit * 3`, which is 12 candidates for the assistant's four results. Kotlin then counts whether each normalized term occurs at least once; it does not consider phrase proximity, frequency, rarity, title/system fields, or source quality.

This means the final sorter can only rerank the first insertion-ordered matches. A strong result outside that window is invisible. SQLite explicitly provides FTS4 `matchinfo()` metrics for relevance sorting, although integrating a custom rank function through Room may be more change than the POC needs.

Minimal correction for the POC:

- First remove stale FTS rows.
- Raise the bounded candidate pool to a measured fixed cap such as 32 or 64 for a modest corpus.
- Return candidate diagnostics and rerank off the main thread using transparent signals: exact phrase, all-term coverage, term frequency, compact proximity, and title/system match.
- Preserve strict and relaxed retrieval as separate logged stages. Do not silently replace an empty strict result with relaxed results without recording the mode.
- Keep FTS4 for now. Consider `matchinfo()` only if the bounded Kotlin reranker is too slow or inaccurate in the target corpus.

### P0 - Answer validation checks the rendered prompt, not the user question

`LocalModelAiEngine.generate()` replaces `AiRequest.prompt` with the full model prompt. `LlamaCppLocalModelRuntime.generate()` later calls both `EvidenceBriefBuilder.build(question = request.prompt, ...)` and `validateGroundedAnswer(request.prompt, generated, ...)`. The validator's supposed question terms consequently include instructions and evidence text.

An answer can pass the relevance check by repeating almost any evidence or instruction term even if it does not answer the user's intent. The guard also verifies citation-label membership, not whether the cited passage supports each claim. That is useful syntax validation, but it is not semantic grounding validation.

Minimal correction:

- Keep `userQuestion`, `evidence`, and `renderedPrompt` as distinct values through generation.
- Validate question coverage against the original question only.
- Build the evidence brief once. Do not reparse the numbered brief in the runtime; reparsing currently drops continuation lines that do not begin with a numbered citation.
- Record a quality result with explicit reason codes and checks, then fall back when direct-question coverage or citation mapping fails.

### P0 - Displayed citations do not follow the response contract

The `AiResponse` contract contains `citationIds`, but `AssistantViewModel` discards them and creates `ChatMessage(..., citations = results)`. The UI therefore labels all retrieved source blocks as answer citations. A model may cite one source, cite no source, or time out, while the screen still displays every retrieved result beneath the answer.

Additionally, the repository ignores the imported chunk's `citationLabel` and reconstructs labels from pack title and page range. That can erase document-level distinctions in multi-document packs.

Minimal correction:

- Map response citation IDs back to selected evidence items and show only those as citations.
- For deterministic fallback, show exactly the evidence excerpts rendered by the fallback.
- For timeout/error text, show no claim citations; offer a separate expandable "Retrieved evidence" section if useful.
- Carry an explicit citation object containing `sourceId`, `packId`, `documentId`, imported label, page range, and excerpt. Keep display label separate from identity.

### P1 - Timeouts and cancellation are not ready for a larger budget

Current local limits are fixed globally at 1,024 context tokens, 96 output tokens, and a 30-second native deadline. The ViewModel wraps model load plus generation in a 40-second `withTimeoutOrNull`.

Observed constraints:

- Cold model load consumes part of the outer ask budget.
- A blocking JNI call is not made cooperative merely by coroutine timeout. The native deadline is currently the effective protection.
- `aiEngine.cancel()` is invoked only after the outer timeout returns; if JNI is still blocking, this is too late to serve as the watchdog that ends it.
- The runtime cancellation method is correctly outside the generation mutex, but the UI exposes no Cancel action.
- The native bridge reports prompt tokens, generated tokens, and stop reason, but those events do not share an assistant request ID.

Minimal correction before increasing time:

- Separate cold-load and generation timing.
- Add a watchdog coroutine that can call native cancel while generation is in progress, plus a visible Cancel button.
- Make budgets model/profile and answer-mode specific instead of global constants.
- Start tablet experiments around 1,536-2,048 context tokens, 160-192 output tokens, a 60-second native deadline, and a 70-75-second end-to-end ceiling. These are experiment bounds, not defaults, until measured on the target device.
- Preserve prompt instructions and the highest-ranked evidence if native context trimming occurs. The current native trim removes tokens from the beginning, where instructions and early evidence live.

### P1 - The installed model path favors speed over reasoning capability

The UI-visible installed-model list is limited to the LFM2.5 350M family. The first/default profile is the 350M Q2 file. Q4/Q3 profiles exist, and Gemma/Qwen/Phi profiles and engines exist in code, but `availableModels()` only enumerates the LFM family, so the larger profiles are not reachable through the current selector.

This establishes a likely quality ceiling: more time can reduce truncation, but it cannot turn a heavily quantized 350M model into a strong relational reasoner. Model suitability must be decided by the AI specialist's benchmark, not by timeout tuning alone.

Minimal Android-side support:

- Keep the adapter and profile model.
- Prefer installed Q4 over Q2 as the first quality comparison on devices that can run it.
- Allow profile-defined context/output/deadline/thread values and log the exact active profile.
- Expose additional already-defined profiles only after file validation, device-memory gates, prompt-template verification, and tablet benchmarks.

### P1 - Existing logs cannot explain answer quality

Current logs cover model discovery/load, high-level request timing, retrieval count/citation labels, evidence character counts, native tokenization/decode, generation stop reason, and fallback reason. This is a useful start.

Missing diagnostic links:

- No collision-resistant request/run ID spans ViewModel, repository, evidence builder, Kotlin runtime, JNI, and native logs.
- No normalized retrieval terms, strict/relaxed mode, candidate count before filtering, candidate source IDs, lexical scores, or selected-evidence mapping.
- No prompt-trim boolean at the assistant layer and no record of which evidence survived trimming.
- No generated tokens-per-second summary, cold/warm status, app build ID, device profile, pack fingerprints, or database/index counts.
- Pack import has no phase timing, peak batch size, row counts, or per-pack outcome logs.
- The local VS Code helper streams only the current app PID and does not persist a complete timestamped test artifact. Logcat is a circular buffer, so a reproducible capture workflow is required.

Minimal diagnostic event shape:

```text
requestId, stage, elapsedMs, outcome,
buildId, deviceModel, androidVersion,
modelId, quantization, coldLoad, contextLimit, outputLimit, deadlineMs, threads,
queryHash, followUp, retrievalMode, normalizedTermCount,
candidateCount, selectedSourceIds, selectedScores, evidenceChars, estimatedOrNativePromptTokens,
promptTrimmed, generatedTokens, tokensPerSecond, stopReason,
qualityUsable, qualityReason, responseCitationIds, displayedSourceIds
```

Do not log sourcebook text or the raw user question by default. For a deliberate local debug build, an opt-in evidence preview can log bounded redacted excerpts.

For each physical-tablet test, save one artifact containing device/build/model/pack metadata, the full app process logs, and relevant system warnings from before model load through answer render. The current `debug.log` should not be treated as such an artifact.

### P1 - Repository post-processing inherits the caller context

Room's suspend DAO call is asynchronous, but paragraph splitting, regex work, scoring, sorting, and diversification occur in `SourcebookRepository.search()` after the DAO returns. The caller is `viewModelScope`, whose default dispatcher is Main. The work is bounded today but should be main-safe before candidate limits and reranking grow.

Minimal correction:

- Inject a `CoroutineDispatcher` into `SourcebookRepository` and run CPU ranking/excerpt assembly on `Dispatchers.Default`.
- Keep Room DAO work asynchronous as it is.
- Inject the dispatcher in tests, following Android's coroutine guidance.

### P2 - Import is bounded by bytes but still memory-heavy and weakly diagnosed

The importer correctly owns `Dispatchers.IO`, bounds readable archive members, uses a transaction for replacement, and avoids reading `embeddings.npy`. However, an allowed `chunks.jsonl` is materialized as a byte array, a list of JSON objects, a list of chunk entities, and a list of FTS entities before bulk insertion. On a 4 GB or low-RAM tablet this can create avoidable peaks.

The importer also requires `embeddings.npy` but does not validate or import its rows. `inspectPack()` reports `embeddingCount` by reading `chunk_count`, not by inspecting embeddings. This must be corrected before semantic retrieval is trusted.

Minimal correction:

- Add import telemetry first: archive size, declared counts, parsed counts, batch count, elapsed time, and failure phase.
- Enforce maximum archive entries, chunk count, JSONL line bytes, total expanded bytes, and per-pack text characters.
- Stream JSONL into bounded Room insert batches when real packs approach the current memory ceiling.
- Validate embedding dtype, dimensions, row count, finite values, and row-to-`embeddingRowIndex` mapping before enabling vectors.
- Use WorkManager only if imports must survive navigation/process death. For the first foreground POC, a cancellable screen-scoped import with progress is the smaller change.

### P2 - Chat state and tests are not yet long-session safe

The chat stores an unbounded message list, and each assistant message retains full retrieval snippets. The UI renders the whole conversation in a vertically scrolling `Column`. This is acceptable for a short demonstration but not a multi-hour game session.

Current tests cover basic Room replacement/search, excerpt extraction, prompt formatting, simple citation rejection, model selection, and two follow-up-planner cases. They do not cover FTS cleanup, ranking order, citation projection, timeout/cancel races, original-question validation, cold/warm generation, real pack import limits, or physical-device quality.

Minimal correction:

- Cap in-memory history for the POC and use a lazy list before long-session testing.
- Extract a small `AskBooksInteractor` inside `feature:assistant` so retrieval/generation/citation behavior can be unit-tested without Compose. No new module or framework is required.

## Hypotheses About the Failed Tablet Answer

These are hypotheses because no failed-run Android logs or exact question/answer/evidence set is available.

| Hypothesis | Supporting evidence | How to confirm |
| --- | --- | --- |
| Weak or wrong candidates reached the model | rowid ordering, constant DAO rank, tiny candidate window, FTS-only retrieval | Log every candidate source ID/score and manually grade Recall@4 for the failed question |
| Reimport history polluted the FTS candidate set | FTS rows are not deleted on replace/prune | Compare active chunk count with FTS row count; reset DB and rerun the same question |
| The Q2 350M model lacked reasoning capacity | default visible local profile is 350M Q2 | Run identical evidence/prompt across deterministic fallback, 350M Q2, 350M Q4, and one memory-safe larger instruct model |
| The answer was truncated by token/time limits | 96 output tokens and 30-second native deadline | Inspect native `stopReason`, generated token count, and output ending |
| A bad answer incorrectly passed the guard | validator uses the rendered prompt as the question | Replay the captured output in a unit test using original-question versus rendered-prompt validation |
| Citations looked authoritative despite not being used | UI attaches all retrieval results and ignores response citation IDs | Compare generated citation IDs with displayed source IDs |
| Prompt-template mismatch harmed output | LFM uses a hand-built plain template | AI specialist verifies the exact GGUF/model template and replays a fixed prompt through the reference CLI and Android runtime |

## Minimal-Change Plan of Action

### Phase 0 - Make the current result trustworthy

1. Delete FTS rows transactionally on pack replace and prune; add integrity tests.
2. Preserve the original user question through the AI request and validate against it.
3. Map response citation IDs to source objects; separate citations from merely retrieved evidence.
4. Preserve imported citation labels and explicit source identity.
5. Add a request ID and structured stage events across retrieval, generation, fallback, and display.

Exit condition: the same database state and question produce traceable candidates, evidence, answer, and displayed citations, with no stale index rows.

### Phase 1 - Improve lexical retrieval before adding architecture

1. Move reranking/excerpt work to an injected Default dispatcher.
2. Expand the bounded candidate pool and add deterministic phrase, coverage, frequency, and proximity scoring.
3. Add strict/relaxed retrieval diagnostics and a golden question set.
4. Add an evidence preview in debug/test UX so a tester can judge retrieval separately from generation.

Exit condition: lexical and exact-name questions meet the retrieval acceptance targets below on the real imported pack.

### Phase 2 - Increase usable model capability and time safely

1. Add active watchdog cancellation and a Cancel button.
2. Move context, output, deadline, and thread budgets into model profiles.
3. Benchmark Q2 versus Q4 and one memory-safe larger model with identical evidence.
4. Increase budgets only for profiles that meet latency, memory, and cancellation targets.
5. Record cold load, prompt evaluation, generation, total time, stop reason, and tokens/second.

Exit condition: one profile consistently beats deterministic evidence presentation on usefulness without violating responsiveness or grounding targets.

### Phase 3 - Add bounded hybrid retrieval

1. Validate and import `embeddings.npy` metadata and rows in bounded form.
2. Embed the query locally with the exact compatible embedding model.
3. Union top vector candidates with top FTS candidates, deduplicate by source ID, and rerank transparently.
4. Send only the final small evidence set to the generator.

Exit condition: relationship/paraphrase questions improve materially over FTS-only retrieval while exact-name performance does not regress.

## POC Acceptance Criteria

### Data integrity

- After initial import, replacement, and prune, every active chunk has exactly one active FTS row and no FTS row points outside the active chunk set.
- Searching a term that existed only in the replaced pack version returns no result.
- A retrieval response contains no duplicate `sourceId` values.
- Every displayed citation resolves to the exact active pack/document/chunk/page record used by the answer.
- Importing a malformed, oversized, or internally inconsistent pack fails without replacing the previously valid pack.

### Retrieval quality

Create a versioned golden set of at least 40 real questions: 15 exact/lexical, 10 paraphrase, 10 relationship or multi-hop, and 5 deliberately unanswerable.

- Exact/lexical Recall@4: at least 90%.
- Paraphrase Recall@4: at least 75% after hybrid retrieval; record the FTS-only baseline first.
- Relationship/multi-hop Recall@4: at least 65% after hybrid retrieval.
- Unanswerable questions: at least 80% return an insufficient-evidence response rather than a confident answer.
- Human-graded irrelevant evidence blocks: no more than 10% of selected blocks.

### Answer and citation quality

- 100% of displayed answer citations are present in `AiResponse.citationIds` or are explicitly labeled "Retrieved evidence" rather than citations.
- 100% of citation IDs resolve to evidence actually supplied to the model.
- At least 80% of answerable golden questions receive a human score of 3 or higher on a 0-4 usefulness rubric.
- At least 90% of factual claims in accepted answers are supported by one displayed cited passage during manual review.
- No accepted answer passes only because it overlaps with prompt instructions; automated tests cover this regression.

### Tablet performance and control

- Warm retrieval p50 below 500 ms and p95 below 1,500 ms for the target corpus.
- UI remains responsive during import, model load, retrieval, and generation; no ANR occurs in the matrix.
- User cancellation stops native work and returns the UI to idle within 2 seconds at the next bounded native check.
- A configured deadline ends generation and returns a grounded fallback or explicit timeout state within 5 seconds of the deadline.
- Warm end-to-end answer p50 at or below 45 seconds and p95 at or below 75 seconds for the selected POC model; cold load is measured separately.
- Twenty consecutive warm questions complete without process death, runaway FTS growth, or unbounded chat-memory growth.

### Diagnostics

- Every test question has one request ID present in all stage logs.
- A saved test artifact is sufficient to reconstruct: build/device/model, pack state, query plan, candidates and scores, selected evidence, prompt/token budgets, native stop reason, quality decision, answer citation IDs, displayed source IDs, and stage timings.
- Logs contain no raw copyrighted source text or raw user question unless a deliberate local debug opt-in is enabled.

## Test Matrix

| Area | Case | Expected evidence | Test level |
| --- | --- | --- | --- |
| FTS lifecycle | Import, replace changed text, replace again, prune | one FTS row per active chunk; old-only terms absent | Room/Robolectric |
| Identity | Two packs with potentially colliding chunk IDs | rejection or unambiguous pack-qualified mapping | Room/import integration |
| Ranking | Relevant chunk inserted after many weaker matches | correct source appears in top four | repository unit/integration |
| Query modes | strict hit, strict miss plus relaxed hit, empty terms | logged mode and deterministic result | repository unit |
| Follow-up | pronoun/short follow-up, unrelated new question | only true follow-up includes previous question | planner/interactor unit |
| Evidence | multi-paragraph and multiple-source excerpts | no text lost during one-time assembly | AI unit |
| Validation | nonsense with valid citation, prompt-word echo, unsupported citation | rejected with specific reason | AI unit |
| Citation display | one cited result among four retrieved | one citation plus optional separately labeled evidence | ViewModel unit/Compose UI |
| Timeout | cold load, prompt decode timeout, generation timeout | distinct phase/stop reason and bounded return | interactor plus device |
| Cancellation | cancel during load, prompt decode, token loop | UI idle; native stop observed | device instrumentation |
| Model comparison | fallback, 350M Q2, 350M Q4, larger gated profile | same question/evidence and persisted metrics | physical tablet |
| Pack bounds | oversized member, long JSONL line, excessive chunk count, corrupt embeddings | previous pack preserved; precise error phase | importer integration |
| Memory | largest valid pack and 20-question chat | bounded peak and no process death | physical tablet/profiler |
| Semantic retrieval | paraphrase and relationship questions | hybrid improves golden-set Recall@4 | retrieval integration |
| Offline | airplane mode throughout import/search/answer | no network dependency or stall | physical tablet |

## Recommended Diagnostic Runbook for the Next Tablet Session

1. Record app commit/build ID, device model, Android version, available RAM, selected model file name/size/quantization, and imported pack fingerprints.
2. Clear Logcat immediately before launch, start the app, and persist app plus relevant system logs to a timestamped host file rather than only streaming the PID view.
3. Ask each golden question once with deterministic fallback, once with the current local model, and once after any profile/budget change.
4. Save the exact question, expected source IDs, retrieved source IDs/scores, selected evidence IDs, answer, response citation IDs, displayed IDs, quality reason, stop reason, and timings in a local test worksheet.
5. Repeat warm runs without reloading the model; report cold load separately.
6. Reset/reimport the database for one comparison run. If quality changes, inspect FTS integrity before attributing the difference to the model.
7. Keep the device on stable power and note thermal state. Do not compare emulator timing with physical-tablet timing; Android's benchmark guidance explicitly discourages emulator performance conclusions.

## Architecture Guidance to the Coordinating Architect

Keep the current module boundaries. The POC needs small contract improvements, not a new framework:

- `core:data`: own FTS lifecycle, lexical candidates, integrity checks, and main-safe ranking.
- `core:retrieval`: extend result identity and later add a hybrid implementation behind the existing interface.
- `core:importpacks`: validate source and vector invariants; stream only when measured pack sizes require it.
- `core:ai`: preserve structured request fields, own profile budgets, native watchdog/cancel, quality results, and model metrics.
- `feature:assistant`: add a small interactor, render cited versus retrieved sources honestly, and expose progress/cancel.

Avoid introducing a DI framework, a new database, an agent framework, or an on-device LLM reranker for the first POC. The current interfaces are sufficient for the proposed corrections and for a later FTS-plus-vector union.

## Final Priority Order

1. FTS cleanup and source identity integrity.
2. Original-question validation and exact citation projection.
3. Correlated, persisted tablet diagnostics.
4. Bounded lexical reranking and golden retrieval tests.
5. Safe cancellation and profile-specific time/token/context budgets.
6. Q2/Q4/larger-model benchmark on identical evidence.
7. Bounded vector import and hybrid retrieval.
8. Streaming import and long-session UI optimizations when measurements justify them.
