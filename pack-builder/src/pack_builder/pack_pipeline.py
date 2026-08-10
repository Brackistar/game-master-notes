from __future__ import annotations

import hashlib
from dataclasses import dataclass
from pathlib import Path

from pack_builder.chunking import chunk_pages
from pack_builder.chunk_quality import improve_chunks
from pack_builder.models import ExtractedDocument, SourceChunk
from pack_builder.pdf_extract import PdfExtractor, extract_document, slugify
from pack_builder.text_cleanup import clean_documents
from pack_builder.toc_cleanup import remove_toc_pages as remove_toc_pages_from_documents


@dataclass(frozen=True)
class PreparedPackContent:
    documents: list[ExtractedDocument]
    pack_id: str
    chunks: list[SourceChunk]
    cleanup_report: dict[str, object]
    toc_report: dict[str, object]
    chunk_quality_report: dict[str, object]


def prepare_pack_content(
    *,
    pdf_paths: list[Path],
    title: str,
    system: str,
    edition: str,
    extractor: PdfExtractor,
    max_chars_per_chunk: int,
    clean_text: bool,
    remove_toc_pages: bool,
    toc_max_page: int,
    deduplicate: bool,
    chunk_overlap_chars: int,
) -> PreparedPackContent:
    documents = [extract_document(pdf_path, extractor) for pdf_path in pdf_paths]
    documents, cleanup_report = maybe_clean_documents(documents, clean_text)
    documents, toc_report = maybe_remove_toc_pages(
        documents,
        remove_toc_pages=remove_toc_pages,
        toc_max_page=toc_max_page,
    )
    pack_id = make_content_pack_id(system, edition, title, documents)
    chunks = chunk_documents(
        pack_id=pack_id,
        documents=documents,
        max_chars_per_chunk=max_chars_per_chunk,
    )
    chunks, chunk_quality_report = improve_chunks(
        chunks,
        overlap_chars=chunk_overlap_chars,
        deduplicate=deduplicate,
    )
    return PreparedPackContent(
        documents=documents,
        pack_id=pack_id,
        chunks=chunks,
        cleanup_report=cleanup_report,
        toc_report=toc_report,
        chunk_quality_report=chunk_quality_report,
    )


def maybe_clean_documents(
    documents: list[ExtractedDocument],
    enabled: bool,
) -> tuple[list[ExtractedDocument], dict[str, object]]:
    if not enabled:
        return documents, {"enabled": False}
    return clean_documents(documents)


def maybe_remove_toc_pages(
    documents: list[ExtractedDocument],
    *,
    remove_toc_pages: bool,
    toc_max_page: int,
) -> tuple[list[ExtractedDocument], dict[str, object]]:
    if not remove_toc_pages:
        return documents, {"enabled": False}
    return remove_toc_pages_from_documents(documents, max_page=toc_max_page)


def make_content_pack_id(
    system: str,
    edition: str,
    title: str,
    documents: list[ExtractedDocument],
) -> str:
    return make_pack_id(
        system=system,
        edition=edition,
        title=title,
        source_checksums=[document.source_checksum for document in documents],
    )


def make_pack_id(
    *,
    system: str,
    edition: str,
    title: str,
    source_checksums: list[str],
) -> str:
    slug = slugify(f"{system}-{edition}-{title}")
    checksum_part = "|".join(sorted(source_checksums))
    digest = hashlib.sha256(
        f"{system}|{edition}|{title}|{checksum_part}".encode("utf-8")
    ).hexdigest()[:12]
    return f"{slug}-{digest}"


def chunk_documents(
    *,
    pack_id: str,
    documents: list[ExtractedDocument],
    max_chars_per_chunk: int,
) -> list[SourceChunk]:
    chunks: list[SourceChunk] = []
    for document in documents:
        chunks.extend(
            reindex_chunks(
                chunks=chunk_pages(
                    pack_id=pack_id,
                    document_id=document.document_id,
                    pages=document.pages,
                    max_chars=max_chars_per_chunk,
                ),
                row_offset=len(chunks),
            )
        )
    return chunks


def reindex_chunks(chunks: list[SourceChunk], row_offset: int) -> list[SourceChunk]:
    return [
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
        for index, chunk in enumerate(chunks)
    ]
