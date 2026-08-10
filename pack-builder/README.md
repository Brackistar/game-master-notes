# PC Pack Builder

## Metadata

- File: `pack-builder/README.md`
- Created: 2026-08-10
- Last updated: 2026-08-10
- User: brackistar

Local CLI for converting user-owned PDFs into `.gmnpack` archives that the Android app can import.

Generated packs and source PDFs are private local artifacts. Do not commit commercial PDFs, extracted sourcebook text, or generated commercial packs.

## Commands

From this directory:

```bash
uv run pack-builder build --system "Example System" --edition "1e" --title "Example Book" --out example.gmnpack book.pdf
uv run pack-builder inspect example.gmnpack
uv run pack-builder validate example.gmnpack
uv run --extra dev pytest
```

The default build path uses `sentence-transformers/all-MiniLM-L6-v2` and writes 384-dimensional embeddings. Tests use a deterministic local embedding provider so they do not need internet access or model files.

## Build Controls

Useful build options:

```bash
uv run pack-builder build --force --max-chars-per-chunk 1200 --report-out report.json --verbose --system "Example System" --edition "1e" --title "Example Book" --out example.gmnpack book.pdf
uv run pack-builder build --chunk-overlap-chars 200 --no-clean-text --no-deduplicate-chunks --system "Example System" --edition "1e" --title "Example Book" --out example.gmnpack book.pdf
uv run pack-builder build --dry-run --report-out report.json --system "Example System" --edition "1e" --title "Example Book" --out example.gmnpack book.pdf
uv run pack-builder build --config build-config.json
uv run pack-builder report example.gmnpack
uv run pack-builder sample-chunks --limit 5 example.gmnpack
uv run pack-builder inspect --json example.gmnpack
uv run pack-builder validate --json example.gmnpack
```

- `--force` overwrites an existing output pack.
- `--dry-run` extracts and chunks without generating embeddings or writing a pack.
- `--max-chars-per-chunk` controls retrieval chunk size.
- `--chunk-overlap-chars` prepends trailing context from the previous chunk.
- `--clean-text/--no-clean-text` toggles repeated-line removal and hyphen repair.
- `--deduplicate-chunks/--no-deduplicate-chunks` toggles duplicate chunk removal.
- `--report-out` writes the extraction quality report as standalone JSON.
- `--verbose` prints empty, suspicious, duplicate, and warning counts.
- `--json` makes `inspect` and `validate` machine-readable.
- `--config` reads repeatable build settings from a JSON file.

Example `build-config.json`:

```json
{
  "pdfs": ["C:/path/to/book.pdf"],
  "system": "Example System",
  "edition": "1e",
  "title": "Example Book",
  "out": "C:/path/to/example.gmnpack",
  "extractor": "pymupdf",
  "embedding_provider": "sentence-transformers",
  "embedding_model": "sentence-transformers/all-MiniLM-L6-v2",
  "max_chars_per_chunk": 1200,
  "chunk_overlap_chars": 200,
  "clean_text": true,
  "deduplicate_chunks": true,
  "report_out": "C:/path/to/report.json"
}
```

CLI flags override config file values.

## Quality Inspection

- `report` prints extraction quality metrics from `extraction-report.json`.
- `report --json` prints the full report.
- `sample-chunks` prints a few chunk text samples for manual review.
- `sample-chunks --contains "term"` filters sampled chunks by text.

## Archive Layout

Each `.gmnpack` is a ZIP archive containing:

- `manifest.json`
- `documents.json`
- `chunks.jsonl`
- `embeddings.npy`
- `extraction-report.json`

The extraction report records chunking settings, cleanup actions, chunk quality actions, per-page text lengths, empty pages, suspiciously short pages, duplicate normalized page text, warnings, and errors.
