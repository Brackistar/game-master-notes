from __future__ import annotations

from pack_builder.constants import (
    CHUNKS_FILE,
    DOCUMENTS_FILE,
    EMBEDDINGS_FILE,
    EXTRACTION_REPORT_FILE,
    MANIFEST_FILE,
    PACK_SCHEMA_VERSION,
)


def pack_schema_contract() -> dict[str, object]:
    return {
        "schema_version": PACK_SCHEMA_VERSION,
        "archive_files": [
            MANIFEST_FILE,
            DOCUMENTS_FILE,
            CHUNKS_FILE,
            EMBEDDINGS_FILE,
            EXTRACTION_REPORT_FILE,
        ],
        "manifest_required_fields": [
            "schema_version",
            "pack_id",
            "title",
            "system",
            "edition",
            "language",
            "source_pdfs",
            "generator_version",
            "extractor_name",
            "embedding_model_id",
            "embedding_dimensions",
            "chunk_count",
            "created_at",
        ],
        "chunk_required_fields": [
            "chunk_id",
            "document_id",
            "page_start",
            "page_end",
            "citation_label",
            "text",
            "char_count",
            "embedding_row_index",
        ],
        "embedding_format": "numpy .npy float32 matrix",
    }
