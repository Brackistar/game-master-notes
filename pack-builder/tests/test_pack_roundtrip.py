from __future__ import annotations

import io
import zipfile
from pathlib import Path

import numpy as np

from pack_builder.embedding_generation.embeddings import DeterministicEmbeddingProvider
from pack_builder.pack_archive.writer import build_pack
from pack_builder.pdf_extraction.extract import get_extractor
from pack_builder.pack_archive.validate import validate_pack


def test_build_pack_writes_valid_archive(synthetic_pdf: Path, tmp_path: Path) -> None:
    out = tmp_path / "synthetic.gmnpack"

    build_result = build_pack(
        pdf_paths=[synthetic_pdf],
        out_path=out,
        title="Synthetic Book",
        system="Test System",
        edition="1e",
        language="en",
        extractor=get_extractor("pymupdf"),
        embedding_provider=DeterministicEmbeddingProvider(),
    )
    manifest = build_result.manifest

    assert out.exists()
    assert manifest["chunk_count"] >= 1
    assert validate_pack(out).ok
    assert build_result.extraction_report["chunking"]["max_chars_per_chunk"] == 1800

    with zipfile.ZipFile(out) as archive:
        assert {
            "manifest.json",
            "documents.json",
            "chunks.jsonl",
            "embeddings.npy",
            "extraction-report.json",
        } <= set(archive.namelist())
        embeddings = np.load(io.BytesIO(archive.read("embeddings.npy")), allow_pickle=False)

    assert embeddings.shape[0] == manifest["chunk_count"]
    assert embeddings.shape[1] == 384
    assert manifest["build_options"]["extractor"] == "pymupdf"
    assert "timing" in build_result.extraction_report


def test_build_pack_respects_custom_chunk_size(
    synthetic_pdf: Path, tmp_path: Path
) -> None:
    out = tmp_path / "small-chunks.gmnpack"

    build_result = build_pack(
        pdf_paths=[synthetic_pdf],
        out_path=out,
        title="Synthetic Book",
        system="Test System",
        edition="1e",
        language="en",
        extractor=get_extractor("pymupdf"),
        embedding_provider=DeterministicEmbeddingProvider(),
        max_chars_per_chunk=80,
    )

    assert build_result.manifest["chunk_count"] > 1
    assert build_result.extraction_report["chunking"]["max_chars_per_chunk"] == 80


def test_build_pack_records_cleanup_and_chunk_quality(
    synthetic_pdf: Path, tmp_path: Path
) -> None:
    out = tmp_path / "quality.gmnpack"

    build_result = build_pack(
        pdf_paths=[synthetic_pdf],
        out_path=out,
        title="Synthetic Book",
        system="Test System",
        edition="1e",
        language="en",
        extractor=get_extractor("pymupdf"),
        embedding_provider=DeterministicEmbeddingProvider(),
        chunk_overlap_chars=20,
    )

    assert build_result.extraction_report["cleanup"]["enabled"] is True
    assert build_result.extraction_report["front_matter_cleanup"]["enabled"] is True
    assert build_result.extraction_report["toc_cleanup"]["enabled"] is True
    assert build_result.extraction_report["chunk_quality"]["overlap_chars"] == 20
    assert "advanced_extraction" in build_result.extraction_report



