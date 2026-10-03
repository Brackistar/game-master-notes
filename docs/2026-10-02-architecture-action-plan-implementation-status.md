# Architecture Action Plan Implementation Status

## Metadata

- Created: 2026-10-02
- Last updated: 2026-10-02
- User: Brackistar
- File: `docs/2026-10-02-architecture-action-plan-implementation-status.md`
- Source plan: `docs/2026-09-29-poc-architecture-action-plan.md`
- Scope: first-party Android and pack-builder code; vendored `android/third_party/**` excluded

## Completed automated work

### Work package 0 - Correctness baseline

- FTS rows are replaced and pruned transactionally with their owning pack rows.
- Duplicate and cross-pack chunk identity collisions fail closed.
- Schema `1.0`, manifest counts, source identity, citation projection, and absence-based pruning have regression coverage.
- A real pack-builder-produced `.gmnpack` fixture exercises the Android contract.

### Work package 1 - Correlated diagnostics

- Request IDs span planning, retrieval, evidence selection, model load, native generation, validation, display, and completion.
- The app-private JSONL journal is bounded by file size, rotation count, and request count.
- Export and clear are explicit user actions. Default events contain hashes, IDs, ranks, counts, timings, and terminal reasons, not source or question text.

### Work package 2 - Prompt and runtime boundary

- Structured system/user messages cross JNI and the GGUF-embedded chat template is applied natively.
- Evidence is labeled `E1` through `E4` and mapped back to stable source IDs after answer validation.
- Context overflow fails closed; prompt truncation was removed.
- Temperature, top-k, repetition penalty, seed, context, batch, output, and deadline are profile-owned.
- EOG, max tokens, timeout, cancellation, context overflow, unsupported template, prompt decode error, and native error are distinct outcomes.
- The assistant exposes cancellation through the native atomic cancel path.

### Work package 3 - Lexical retrieval

- Strict and relaxed FTS candidates use a bounded 256-row window.
- Deterministic phrase, term coverage, frequency, proximity, metadata, and structural ranking runs on `Dispatchers.Default`.
- Result deduplication and diversity are stable, including a regression where the best hit occurs after the old row cutoff.

### Work package 4 - Hybrid retrieval foundation

- Android validates NPY float32 layout, shape, finite values, and normalized rows before pack replacement.
- Pack-builder validation enforces the same finite and normalization contract.
- Matching encoder/store model IDs and dimensions are mandatory.
- Bounded brute-force cosine search, lexical/vector top-20 retrieval, and deterministic Reciprocal Rank Fusion are implemented and tested.
- ONNX Runtime Mobile 1.30.0 runs a bundled ARM64 quantized `all-MiniLM-L6-v2` encoder on CPU.
- The model revision, vocabulary, tokenizer length, pooling, normalization, and checksums are pinned and checked at installation.
- Pack manifests now carry the immutable embedding revision. Older unversioned packs remain available to lexical search but are excluded from semantic retrieval until rebuilt.
- Validated NPY matrices are atomically installed as app-private per-pack sidecars and repaired on the next scan if a filesystem commit fails.
- The assistant now uses hybrid retrieval in production with privacy-preserving lexical, vector, and fused rank diagnostics. Encoder failures fall back to lexical retrieval.

### Work packages 5 and 6 - Experiments and relationships

- LFM2.5 350M Q4 remains the baseline; Q2/Q3 remain compatibility profiles.
- LFM2.5 1.2B Q4 is present only as an explicitly labeled experiment.
- Bounded adjacency, overlapping-page, and repeated-name relationship expansion is implemented as a pure tested policy.
- The richer relationship schema, ANN index, and knowledge graph remain deferred as directed by the plan.

### Evaluation infrastructure

- A deterministic 50-question synthetic corpus covers literal, paraphrase, relationship, numeric, adjacent-evidence, unanswerable, and adversarial cases.
- Recall and reciprocal-rank metric helpers are tested.
- Existing AppContainer construction remains the test seam; no DI or mocking framework was added.

## Remaining external gates

These are not software-complete claims and must remain open before promoting hybrid retrieval or a larger model:

1. Run the 50-case corpus on the TCL tablet against real imported lexical and hybrid indexes and record Recall@4 and nDCG@4. The corpus, metric primitives, production encoder, and hybrid path exist; measured gate results do not.
2. Run physical-tablet cold/warm latency, PSS/native RSS, repeated-query thermal, cancellation, timeout, and malformed-model checks with the exact GGUF files. No local JVM test can substitute for these measurements.
3. Promote the 1.2B experiment only if it improves useful/correct answers by at least 10 percentage points and passes every device gate.

## Automated verification

Run from `android/`:

```powershell
.\gradlew.bat testDebugUnitTest
.\gradlew.bat :core:ai:externalNativeBuildDebug
```

Run from `pack-builder/`:

```powershell
.\.venv\Scripts\python.exe -m pytest
```

Verified on 2026-10-02: Android unit tests passed, the arm64 native build passed, and all 54 pack-builder tests passed. Physical-device acceptance was intentionally not claimed.
