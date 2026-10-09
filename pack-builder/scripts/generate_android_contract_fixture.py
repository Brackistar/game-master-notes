from __future__ import annotations

from pathlib import Path

import numpy as np

from pack_builder.core_domain.models import SourceChunk
from pack_builder.pack_archive.writer import write_pack_archive


def main() -> None:
    repository = Path(__file__).resolve().parents[2]
    output = repository / "pack-builder" / "tests" / "fixtures" / "v1" / "synthetic.gmnpack"
    chunks = [
        SourceChunk(
            chunk_id="fixture-pack:fixture-doc:0000",
            document_id="fixture-doc",
            page_start=1,
            page_end=1,
            citation_label="Fixture Codex p. 1",
            text="The Azure Bell opens the observatory only at midnight.",
            char_count=58,
            embedding_row_index=0,
        ),
        SourceChunk(
            chunk_id="fixture-pack:fixture-doc:0001",
            document_id="fixture-doc",
            page_start=2,
            page_end=2,
            citation_label="Fixture Codex p. 2",
            text="The keeper carries a silver key marked with seven stars.",
            char_count=58,
            embedding_row_index=1,
        ),
    ]
    embeddings = np.zeros((2, 384), dtype=np.float32)
    embeddings[0, 0] = 1.0
    embeddings[1, 1] = 1.0
    write_pack_archive(
        out_path=output,
        manifest={
            "schema_version": "1.0",
            "pack_id": "fixture-pack",
            "title": "Fixture Codex",
            "system": "Fixture System",
            "edition": "1e",
            "language": "en",
            "source_pdfs": [{"filename": "fixture.pdf", "checksum": "fixture-checksum"}],
            "generator_version": "0.1.0",
            "extractor_name": "fixture",
            "embedding_model_id": "deterministic-fixture-v1",
            "embedding_model_revision": "deterministic-fixture-revision-v1",
            "embedding_dimensions": 384,
            "chunk_count": len(chunks),
            "build_options": {"fixture": True},
            "created_at": "2026-09-30T00:00:00+00:00",
        },
        documents=[
            {
                "document_id": "fixture-doc",
                "source_filename": "fixture.pdf",
                "source_checksum": "fixture-checksum",
                "page_count": 2,
                "extraction_stats": {
                    "empty_page_count": 0,
                    "suspicious_page_count": 0,
                    "total_characters": 116,
                },
            }
        ],
        chunks=chunks,
        embeddings=embeddings,
        extraction_report={"fixture": True, "chunk_count": len(chunks)},
    )
    print(output)


if __name__ == "__main__":
    main()
