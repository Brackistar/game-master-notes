from __future__ import annotations

import io
import json
import zipfile
from pathlib import Path

import numpy as np

from pack_builder.constants import (
    CHUNKS_FILE,
    DOCUMENTS_FILE,
    EMBEDDINGS_FILE,
    EXTRACTION_REPORT_FILE,
    MANIFEST_FILE,
)
from pack_builder.validate import validate_pack


def write_minimal_pack(
    path: Path,
    *,
    manifest_dimensions: int = 384,
    embedding_dimensions: int = 384,
    rows: int = 1,
    malformed_chunk: bool = False,
) -> None:
    manifest = {
        "schema_version": "1.0",
        "pack_id": "pack",
        "title": "Title",
        "system": "System",
        "edition": "1e",
        "language": "en",
        "source_pdfs": [{"filename": "book.pdf", "checksum": "abc"}],
        "generator_version": "0.1.0",
        "extractor_name": "pymupdf",
        "embedding_model_id": "deterministic-test-embedding",
        "embedding_dimensions": manifest_dimensions,
        "chunk_count": rows,
        "created_at": "2026-08-10T00:00:00+00:00",
    }
    chunk = {
        "chunk_id": "chunk-1",
        "document_id": "doc",
        "page_start": 1,
        "page_end": 1,
        "citation_label": "doc p. 1",
        "text": "Some text.",
        "char_count": len("Some text."),
        "embedding_row_index": 0,
    }
    chunk_payload = {"chunk_id": "broken"} if malformed_chunk else chunk
    embedding_buffer = io.BytesIO()
    np.save(
        embedding_buffer,
        np.zeros((rows, embedding_dimensions), dtype=np.float32),
        allow_pickle=False,
    )

    with zipfile.ZipFile(path, "w") as archive:
        archive.writestr(MANIFEST_FILE, json.dumps(manifest))
        archive.writestr(DOCUMENTS_FILE, json.dumps({"documents": [{"document_id": "doc"}]}))
        archive.writestr(CHUNKS_FILE, json.dumps(chunk_payload) + "\n")
        archive.writestr(EMBEDDINGS_FILE, embedding_buffer.getvalue())
        archive.writestr(EXTRACTION_REPORT_FILE, json.dumps({"extractor": "pymupdf"}))


def test_validate_accepts_minimal_valid_pack(tmp_path: Path) -> None:
    pack = tmp_path / "valid.gmnpack"
    write_minimal_pack(pack)

    assert validate_pack(pack).ok


def test_validate_rejects_missing_required_files(tmp_path: Path) -> None:
    pack = tmp_path / "missing.gmnpack"
    with zipfile.ZipFile(pack, "w") as archive:
        archive.writestr(MANIFEST_FILE, "{}")

    result = validate_pack(pack)

    assert not result.ok
    assert any("missing required file" in error for error in result.errors)


def test_validate_rejects_bad_embedding_dimensions(tmp_path: Path) -> None:
    pack = tmp_path / "bad-dimensions.gmnpack"
    write_minimal_pack(pack, manifest_dimensions=7, embedding_dimensions=8)

    result = validate_pack(pack)

    assert not result.ok
    assert any("embedding_dimensions" in error for error in result.errors)


def test_validate_rejects_malformed_chunks(tmp_path: Path) -> None:
    pack = tmp_path / "bad-chunk.gmnpack"
    write_minimal_pack(pack, malformed_chunk=True)

    result = validate_pack(pack)

    assert not result.ok
    assert any("missing field" in error for error in result.errors)
