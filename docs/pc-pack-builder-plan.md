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

## Future Directory

Use `pack-builder/` when implementation begins.

Suggested structure:

- `pack_builder/cli.py`
- `pack_builder/pdf_extract.py`
- `pack_builder/chunking.py`
- `pack_builder/embeddings.py`
- `pack_builder/pack_writer.py`
- `pack_builder/validate.py`
- `tests/`

## V1 Pack Archive Layout

Use a ZIP archive with this internal shape:

- `manifest.json`
- `documents.json`
- `chunks.jsonl`
- `embeddings.bin` or `embeddings.npy`
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
- `validate`: verify manifest, chunk count, embedding count, and required fields.

Example shape:

```bash
pack-builder build --system "Mage" --edition "20th" --title "Core Rulebook" --out mage-core.gmnpack book.pdf
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
- Use tiny synthetic PDFs in the repo for tests, not copyrighted sourcebooks.
- Manually test with real owned books outside the repo.

## Risks

- PDFs with scanned images may need OCR later.
- Rulebooks with complex tables may extract poorly.
- Embedding model choice affects Android search compatibility.
- Very large packs may stress tablet import time and storage.

## Suggested Extensions

- OCR pipeline for scanned books.
- Extraction profiles per publisher or PDF style.
- Deduplication for repeated headers, footers, and legal text.
- Optional image/table extraction metadata.
- Pack signing or checksums for integrity.
