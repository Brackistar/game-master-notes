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

## Archive Layout

Each `.gmnpack` is a ZIP archive containing:

- `manifest.json`
- `documents.json`
- `chunks.jsonl`
- `embeddings.npy`
- `extraction-report.json`
