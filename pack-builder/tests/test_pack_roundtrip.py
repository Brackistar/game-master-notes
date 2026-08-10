from __future__ import annotations

import io
import zipfile
from pathlib import Path

import numpy as np

from pack_builder.embeddings import DeterministicEmbeddingProvider
from pack_builder.pack_writer import build_pack
from pack_builder.pdf_extract import get_extractor
from pack_builder.validate import validate_pack


def test_build_pack_writes_valid_archive(synthetic_pdf: Path, tmp_path: Path) -> None:
    out = tmp_path / "synthetic.gmnpack"

    manifest = build_pack(
        pdf_paths=[synthetic_pdf],
        out_path=out,
        title="Synthetic Book",
        system="Test System",
        edition="1e",
        language="en",
        extractor=get_extractor("pymupdf"),
        embedding_provider=DeterministicEmbeddingProvider(),
    )

    assert out.exists()
    assert manifest["chunk_count"] >= 1
    assert validate_pack(out).ok

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
