from __future__ import annotations

import io
import json
import zipfile
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path

import numpy as np

from pack_builder.chunking import chunk_pages
from pack_builder.constants import (
    CHUNKS_FILE,
    DEFAULT_MAX_CHARS_PER_CHUNK,
    DOCUMENTS_FILE,
    EMBEDDINGS_FILE,
    EXTRACTION_REPORT_FILE,
    GENERATOR_VERSION,
    MANIFEST_FILE,
    PACK_SCHEMA_VERSION,
    SUSPICIOUS_PAGE_CHAR_THRESHOLD,
)
from pack_builder.embeddings import EmbeddingProvider
from pack_builder.models import ExtractedDocument, SourceChunk
from pack_builder.pdf_extract import PdfExtractor, extract_document, slugify


@dataclass(frozen=True)
class BuildResult:
    manifest: dict[str, object]
    documents: list[dict[str, object]]
    chunks: list[SourceChunk]
    extraction_report: dict[str, object]


def make_pack_id(
    *,
    system: str,
    edition: str,
    title: str,
    source_checksums: list[str],
) -> str:
    import hashlib

    slug = slugify(f"{system}-{edition}-{title}")
    checksum_part = "|".join(sorted(source_checksums))
    digest = hashlib.sha256(
        f"{system}|{edition}|{title}|{checksum_part}".encode("utf-8")
    ).hexdigest()[:12]
    return f"{slug}-{digest}"


def build_extraction_report(
    extractor_name: str,
    documents: list[ExtractedDocument],
    max_chars_per_chunk: int,
) -> dict[str, object]:
    page_lengths: dict[str, list[dict[str, int]]] = {}
    empty_pages: list[dict[str, object]] = []
    suspicious_pages: list[dict[str, object]] = []
    warnings: list[dict[str, object]] = []
    seen_page_texts: dict[str, dict[str, object]] = {}
    duplicate_pages: list[dict[str, object]] = []

    for document in documents:
        lengths: list[dict[str, int]] = []
        for page in document.pages:
            normalized_page_text = " ".join(page.text.split())
            text_len = len(page.text.strip())
            lengths.append({"page": page.page_number, "characters": text_len})
            if text_len == 0:
                empty_pages.append(
                    {"document_id": document.document_id, "page": page.page_number}
                )
            elif text_len < SUSPICIOUS_PAGE_CHAR_THRESHOLD:
                suspicious_pages.append(
                    {
                        "document_id": document.document_id,
                        "page": page.page_number,
                        "characters": text_len,
                    }
                )
            for warning in page.warnings:
                warnings.append(
                    {
                        "document_id": document.document_id,
                        "page": page.page_number,
                        "warning": warning,
                    }
                )
            if normalized_page_text:
                import hashlib

                text_hash = hashlib.sha256(
                    normalized_page_text.encode("utf-8")
                ).hexdigest()
                first_seen = seen_page_texts.get(text_hash)
                if first_seen:
                    duplicate_pages.append(
                        {
                            "document_id": document.document_id,
                            "page": page.page_number,
                            "matches_document_id": first_seen["document_id"],
                            "matches_page": first_seen["page"],
                        }
                    )
                else:
                    seen_page_texts[text_hash] = {
                        "document_id": document.document_id,
                        "page": page.page_number,
                    }
        page_lengths[document.document_id] = lengths

    return {
        "extractor": extractor_name,
        "chunking": {
            "max_chars_per_chunk": max_chars_per_chunk,
            "strategy": "paragraph-aware",
        },
        "page_text_lengths": page_lengths,
        "empty_pages": empty_pages,
        "suspicious_pages": suspicious_pages,
        "duplicate_pages": duplicate_pages,
        "warnings": warnings,
        "errors": [],
    }


def document_metadata(document: ExtractedDocument) -> dict[str, object]:
    text_lengths = [len(page.text.strip()) for page in document.pages]
    return {
        "document_id": document.document_id,
        "source_filename": document.source_filename,
        "source_checksum": document.source_checksum,
        "page_count": document.page_count,
        "extraction_stats": {
            "empty_page_count": sum(1 for length in text_lengths if length == 0),
            "suspicious_page_count": sum(
                1
                for length in text_lengths
                if 0 < length < SUSPICIOUS_PAGE_CHAR_THRESHOLD
            ),
            "total_characters": sum(text_lengths),
        },
    }


def manifest_data(
    *,
    pack_id: str,
    title: str,
    system: str,
    edition: str,
    language: str,
    documents: list[ExtractedDocument],
    extractor_name: str,
    embedding_provider: EmbeddingProvider,
    chunk_count: int,
) -> dict[str, object]:
    return {
        "schema_version": PACK_SCHEMA_VERSION,
        "pack_id": pack_id,
        "title": title,
        "system": system,
        "edition": edition,
        "language": language,
        "source_pdfs": [
            {
                "filename": document.source_filename,
                "checksum": document.source_checksum,
            }
            for document in documents
        ],
        "generator_version": GENERATOR_VERSION,
        "extractor_name": extractor_name,
        "embedding_model_id": embedding_provider.model_id,
        "embedding_dimensions": embedding_provider.dimensions,
        "chunk_count": chunk_count,
        "created_at": datetime.now(timezone.utc).replace(microsecond=0).isoformat(),
    }


def write_pack_archive(
    *,
    out_path: Path,
    manifest: dict[str, object],
    documents: list[dict[str, object]],
    chunks: list[SourceChunk],
    embeddings: np.ndarray,
    extraction_report: dict[str, object],
) -> None:
    out_path.parent.mkdir(parents=True, exist_ok=True)
    with zipfile.ZipFile(out_path, mode="w", compression=zipfile.ZIP_DEFLATED) as archive:
        archive.writestr(MANIFEST_FILE, json.dumps(manifest, indent=2) + "\n")
        archive.writestr(
            DOCUMENTS_FILE, json.dumps({"documents": documents}, indent=2) + "\n"
        )
        archive.writestr(
            CHUNKS_FILE,
            "".join(json.dumps(chunk.to_json()) + "\n" for chunk in chunks),
        )
        embedding_buffer = io.BytesIO()
        np.save(embedding_buffer, embeddings.astype(np.float32), allow_pickle=False)
        archive.writestr(EMBEDDINGS_FILE, embedding_buffer.getvalue())
        archive.writestr(
            EXTRACTION_REPORT_FILE, json.dumps(extraction_report, indent=2) + "\n"
        )


def build_pack(
    *,
    pdf_paths: list[Path],
    out_path: Path,
    title: str,
    system: str,
    edition: str,
    language: str,
    extractor: PdfExtractor,
    embedding_provider: EmbeddingProvider,
    max_chars_per_chunk: int = DEFAULT_MAX_CHARS_PER_CHUNK,
) -> BuildResult:
    documents = [extract_document(pdf_path, extractor) for pdf_path in pdf_paths]
    pack_id = make_pack_id(
        system=system,
        edition=edition,
        title=title,
        source_checksums=[document.source_checksum for document in documents],
    )

    chunks: list[SourceChunk] = []
    for document in documents:
        document_chunks = chunk_pages(
            pack_id=pack_id,
            document_id=document.document_id,
            pages=document.pages,
            max_chars=max_chars_per_chunk,
        )
        row_offset = len(chunks)
        chunks.extend(
            SourceChunk(
                chunk_id=chunk.chunk_id,
                document_id=chunk.document_id,
                page_start=chunk.page_start,
                page_end=chunk.page_end,
                citation_label=chunk.citation_label,
                text=chunk.text,
                char_count=chunk.char_count,
                embedding_row_index=row_offset + index,
            )
            for index, chunk in enumerate(document_chunks)
        )

    if not chunks:
        raise ValueError("no text chunks were created from the provided PDFs")

    embeddings = embedding_provider.encode([chunk.text for chunk in chunks])
    if embeddings.shape != (len(chunks), embedding_provider.dimensions):
        raise ValueError(
            "embedding provider returned shape "
            f"{embeddings.shape}, expected {(len(chunks), embedding_provider.dimensions)}"
        )

    manifest = manifest_data(
        pack_id=pack_id,
        title=title,
        system=system,
        edition=edition,
        language=language,
        documents=documents,
        extractor_name=extractor.name,
        embedding_provider=embedding_provider,
        chunk_count=len(chunks),
    )
    document_rows = [document_metadata(document) for document in documents]
    report = build_extraction_report(extractor.name, documents, max_chars_per_chunk)

    write_pack_archive(
        out_path=out_path,
        manifest=manifest,
        documents=document_rows,
        chunks=chunks,
        embeddings=embeddings,
        extraction_report=report,
    )
    return BuildResult(
        manifest=manifest,
        documents=document_rows,
        chunks=chunks,
        extraction_report=report,
    )


def preview_pack(
    *,
    pdf_paths: list[Path],
    title: str,
    system: str,
    edition: str,
    language: str,
    extractor: PdfExtractor,
    max_chars_per_chunk: int = DEFAULT_MAX_CHARS_PER_CHUNK,
) -> BuildResult:
    documents = [extract_document(pdf_path, extractor) for pdf_path in pdf_paths]
    pack_id = make_pack_id(
        system=system,
        edition=edition,
        title=title,
        source_checksums=[document.source_checksum for document in documents],
    )
    chunks: list[SourceChunk] = []
    for document in documents:
        row_offset = len(chunks)
        document_chunks = chunk_pages(
            pack_id=pack_id,
            document_id=document.document_id,
            pages=document.pages,
            max_chars=max_chars_per_chunk,
        )
        chunks.extend(
            SourceChunk(
                chunk_id=chunk.chunk_id,
                document_id=chunk.document_id,
                page_start=chunk.page_start,
                page_end=chunk.page_end,
                citation_label=chunk.citation_label,
                text=chunk.text,
                char_count=chunk.char_count,
                embedding_row_index=row_offset + index,
            )
            for index, chunk in enumerate(document_chunks)
        )
    report = build_extraction_report(extractor.name, documents, max_chars_per_chunk)
    manifest = {
        "schema_version": PACK_SCHEMA_VERSION,
        "pack_id": pack_id,
        "title": title,
        "system": system,
        "edition": edition,
        "language": language,
        "source_pdfs": [
            {"filename": document.source_filename, "checksum": document.source_checksum}
            for document in documents
        ],
        "generator_version": GENERATOR_VERSION,
        "extractor_name": extractor.name,
        "embedding_model_id": None,
        "embedding_dimensions": None,
        "chunk_count": len(chunks),
        "created_at": None,
        "dry_run": True,
    }
    return BuildResult(
        manifest=manifest,
        documents=[document_metadata(document) for document in documents],
        chunks=chunks,
        extraction_report=report,
    )
