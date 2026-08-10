# PC Pack Builder Plan

## Metadata

- File: `docs/pc-pack-builder-plan.md`
- Created: 2026-08-10
- Last updated: 2026-08-10
- User: brackistar

Related diagram: [pc-pack-builder-plan.mmd](pc-pack-builder-plan.mmd)

## Purpose

The PC pack builder converts user-owned sourcebook PDFs into portable sourcebook packs that the Android app can import. It runs on the user's local computer because PDF extraction and embedding generation are more practical there than on low-requirement Android tablets.

The builder must never require committing PDFs, extracted book text, or commercial generated packs to the repository.

## Product Goals

- Accept one or more local PDFs.
- Extract text with page references preserved.
- Create citeable text chunks.
- Generate embeddings for semantic search.
- Produce an importable pack archive with manifest, chunks, citations, and embeddings.
- Produce logs and validation output so bad PDF extraction can be diagnosed.

## Suggested Stack

- Language: Python
- CLI framework: Typer or argparse
- PDF extraction: start with PyMuPDF or pdfplumber, then benchmark both on real books
- Data files inside pack: JSON for manifest and metadata, JSONL for chunks, binary or NumPy-compatible format for embeddings
- Archive format: ZIP with a stable internal layout
- Tests: pytest

## Implementation Status

The first implementation slice now lives in `pack-builder/`.

Current structure:

- `src/pack_builder/build_config.py`
- `src/pack_builder/cli.py`
- `src/pack_builder/chunking.py`
- `src/pack_builder/chunk_quality.py`
- `src/pack_builder/embeddings.py`
- `src/pack_builder/extraction_diagnostics.py`
- `src/pack_builder/extraction_report.py`
- `src/pack_builder/extractor_compare.py`
- `src/pack_builder/pdf_extract.py`
- `src/pack_builder/pack_pipeline.py`
- `src/pack_builder/pack_reader.py`
- `src/pack_builder/pack_writer.py`
- `src/pack_builder/schema_contract.py`
- `src/pack_builder/text_cleanup.py`
- `src/pack_builder/toc_cleanup.py`
- `src/pack_builder/validate.py`
- `tests/`

The current code keeps the CLI, pack preparation pipeline, archive writing, and extraction report generation in separate modules to keep file size and cognitive complexity under control.

Implemented commands:

- `build --system --edition --title --out --extractor [pymupdf|pdfplumber] <pdf...>`
- `build --config <json>`
- `build --dry-run`
- `build --force`
- `build --chunk-overlap-chars <int>`
- `build --max-chars-per-chunk <int>`
- `build --no-clean-text`
- `build --keep-toc-pages`
- `build --toc-max-page <int>`
- `build --no-deduplicate-chunks`
- `build --report-out <json>`
- `inspect <pack>`
- `inspect --json <pack>`
- `report <pack>`
- `report --json <pack>`
- `compare-extractors <pdf>`
- `compare-extractors --json <pdf>`
- `schema`
- `schema --json`
- `sample-chunks <pack>`
- `sample-chunks --contains <text> --limit <n> <pack>`
- `validate <pack>`
- `validate --json <pack>`

Implemented quality controls:

- Output overwrite protection unless `--force` is passed.
- Dry run extraction and chunk preview without embedding generation.
- Config-file builds with CLI flag override precedence.
- Extraction report export and report inspection.
- Sample chunk inspection for real-world PDF review.
- Repeated-line cleanup for common headers and footers.
- Hyphenated line repair.
- Early table-of-contents cleanup with a max-page guard, including dense inline TOC text extracted from complex PDFs.
- Optional chunk overlap for retrieval continuity.
- Duplicate normalized chunk removal.
- Duplicate normalized page detection.
- OCR-needed page diagnostics for image-only extraction results.
- Table-shaped and multi-column-shaped text diagnostics.
- Extractor comparison between PyMuPDF and pdfplumber.
- Schema contract output for Android importer planning.
- Password-protected PDF handling is intentionally excluded.
- Focused config validation for invalid JSON, bad `pdfs`, missing required values, missing PDFs, and too-small chunks.

Implemented archive layout:

- `manifest.json`
- `documents.json`
- `chunks.jsonl`
- `embeddings.npy`
- `extraction-report.json`

## V1 Pack Archive Layout

Use a ZIP archive with this internal shape:

- `manifest.json`
- `documents.json`
- `chunks.jsonl`
- `embeddings.npy`
- `extraction-report.json`

The manifest should include:

- Pack schema version
- Pack id
- Title
- Game system
- Edition
- Language
- Source PDF filename
- Source PDF checksum
- Generator version
- Embedding model id
- Embedding dimensions
- Chunk count
- Created timestamp

Each chunk should include:

- Stable chunk id
- Document id
- Section or heading label when known
- Page start
- Page end
- Citation label
- Normalized text
- Token or character count
- Embedding row index

## CLI Commands

Initial commands:

- `build`: convert PDF files into a pack archive.
- `inspect`: print pack metadata and extraction stats.
- `report`: print extraction quality metrics.
- `compare-extractors`: compare PyMuPDF and pdfplumber output for one PDF.
- `schema`: print the current `.gmnpack` contract.
- `sample-chunks`: inspect representative chunks for manual quality review.
- `validate`: verify manifest, chunk count, embedding count, and required fields.

Example shape:

```bash
pack-builder build --system "Mage" --edition "20th" --title "Core Rulebook" --out mage-core.gmnpack book.pdf
pack-builder build --config mage-core-build.json
pack-builder report mage-core.gmnpack
pack-builder compare-extractors book.pdf
pack-builder schema --json
pack-builder sample-chunks --limit 5 mage-core.gmnpack
pack-builder inspect mage-core.gmnpack
pack-builder validate mage-core.gmnpack
```

## Implementation Phases

### Phase 1: Text Extraction

- Read a PDF page by page.
- Preserve page numbers.
- Normalize whitespace.
- Emit extraction reports with empty pages, suspiciously short pages, and extraction errors.

### Phase 2: Chunking

- Split text into chunks that are small enough for retrieval context.
- Preserve page ranges and section labels where available.
- Prefer paragraph-aware splitting before falling back to character limits.
- Keep chunk ids deterministic from pack id, document id, page range, and chunk index.

### Phase 3: Embeddings

- Generate one embedding per chunk.
- Record model id and dimensions in the manifest.
- Fail validation if embedding count and chunk count differ.
- Keep the embedding model replaceable so Android does not hard-code one provider.

### Phase 4: Pack Writing

- Write manifest, document metadata, chunks, embeddings, and report into one ZIP archive.
- Validate the archive immediately after writing.
- Make pack generation repeatable for the same input and settings when possible.

## Testing

- Unit test page extraction normalization.
- Unit test chunk boundaries and citation labels.
- Unit test manifest validation.
- Unit test pack read/write round trips.
- Unit test build config validation and CLI override behavior.
- Unit test report and sample-chunk CLI output.
- Unit test repeated-line cleanup and hyphen repair.
- Unit test table-of-contents cleanup.
- Unit test chunk overlap and duplicate chunk removal.
- Unit test advanced extraction diagnostics.
- Unit test schema contract output.
- Unit test extractor comparison CLI output.
- Use tiny synthetic PDFs in the repo for tests, not copyrighted sourcebooks.
- Manually test with real owned books outside the repo.

## Risks

- PDFs with scanned images may need OCR later.
- Password-protected PDFs are not handled by design.
- Rulebooks with complex tables may extract poorly.
- Embedding model choice affects Android search compatibility.
- Very large packs may stress tablet import time and storage.

## Suggested Extensions

- OCR pipeline for scanned books.
- Extraction profiles per publisher or PDF style.
- Deduplication for repeated headers, footers, and legal text.
- Optional image/table extraction metadata.
- Pack signing or checksums for integrity.
