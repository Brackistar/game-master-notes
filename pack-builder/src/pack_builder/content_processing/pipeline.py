from __future__ import annotations

import hashlib
from dataclasses import dataclass
from pathlib import Path

from pack_builder.content_processing.chunking import Paragraph, chunk_pages, chunk_paragraphs, split_page_paragraphs
from pack_builder.content_processing.content_classification import (
    ContentLabel,
    MIN_CONFIDENT_CHARACTERS,
    MIN_CONFIDENT_PARAGRAPHS,
    classify_paragraph,
    summarize_classifications,
)
from pack_builder.content_processing.chunk_quality import improve_chunks
from pack_builder.content_processing.front_matter_cleanup import (
    remove_front_matter_pages,
)
from pack_builder.core_domain.models import ExtractedDocument, SourceChunk
from pack_builder.pdf_extraction.extract import PdfExtractor, extract_document, slugify
from pack_builder.content_processing.text_cleanup import clean_documents
from pack_builder.content_processing.toc_cleanup import (
    remove_toc_pages as remove_toc_pages_from_documents,
)


@dataclass(frozen=True)
class PreparedPackContent:
    documents: list[ExtractedDocument]
    pack_id: str
    chunks: list[SourceChunk]
    cleanup_report: dict[str, object]
    front_matter_report: dict[str, object]
    toc_report: dict[str, object]
    chunk_quality_report: dict[str, object]


@dataclass(frozen=True)
class PreparedSplitPackContent:
    categories: dict[ContentLabel, PreparedPackContent]
    classification_report: dict[str, object]
    category_decisions: dict[str, dict[str, object]]


def prepare_pack_content(
    *,
    pdf_paths: list[Path],
    title: str,
    system: str,
    edition: str,
    extractor: PdfExtractor,
    max_chars_per_chunk: int,
    clean_text: bool,
    remove_front_matter: bool,
    front_matter_max_page: int,
    remove_toc_pages: bool,
    toc_max_page: int,
    deduplicate: bool,
    chunk_overlap_chars: int,
) -> PreparedPackContent:
    documents = [extract_document(pdf_path, extractor) for pdf_path in pdf_paths]
    documents, cleanup_report = maybe_clean_documents(documents, clean_text)
    documents, front_matter_report = maybe_remove_front_matter(
        documents,
        remove_front_matter=remove_front_matter,
        front_matter_max_page=front_matter_max_page,
    )
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
        front_matter_report=front_matter_report,
        toc_report=toc_report,
        chunk_quality_report=chunk_quality_report,
    )


def prepare_split_pack_content(
    *,
    pdf_paths: list[Path],
    title: str,
    system: str,
    edition: str,
    language: str,
    extractor: PdfExtractor,
    max_chars_per_chunk: int,
    clean_text: bool,
    remove_front_matter: bool,
    front_matter_max_page: int,
    remove_toc_pages: bool,
    toc_max_page: int,
    deduplicate: bool,
    chunk_overlap_chars: int,
) -> PreparedSplitPackContent:
    documents = [extract_document(pdf_path, extractor) for pdf_path in pdf_paths]
    documents, cleanup_report = maybe_clean_documents(documents, clean_text)
    documents, front_matter_report = maybe_remove_front_matter(
        documents,
        remove_front_matter=remove_front_matter,
        front_matter_max_page=front_matter_max_page,
    )
    documents, toc_report = maybe_remove_toc_pages(
        documents,
        remove_toc_pages=remove_toc_pages,
        toc_max_page=toc_max_page,
    )

    category_paragraphs: dict[ContentLabel, dict[str, list[Paragraph]]] = {
        ContentLabel.RULESET: {},
        ContentLabel.NARRATIVE: {},
    }
    classified_rows: list[tuple[str, object]] = []
    for document in documents:
        for page in document.pages:
            for paragraph in split_page_paragraphs(page):
                result = classify_paragraph(paragraph.text, language)
                classified_rows.append((paragraph.text, result))
                destinations = (
                    (ContentLabel.RULESET, ContentLabel.NARRATIVE)
                    if result.label is ContentLabel.AMBIGUOUS
                    else (result.label,)
                )
                for destination in destinations:
                    category_paragraphs[destination].setdefault(document.document_id, []).append(paragraph)

    classification_report = summarize_classifications(classified_rows, language=language)
    confident_characters = classification_report["confident_characters"]
    confident_paragraphs = classification_report["confident_paragraphs"]
    categories: dict[ContentLabel, PreparedPackContent] = {}
    decisions: dict[str, dict[str, object]] = {}
    for label, suffix in ((ContentLabel.RULESET, "Ruleset"), (ContentLabel.NARRATIVE, "Narrative")):
        category_title = f"{title} - {suffix}"
        pack_id = make_content_pack_id(system, edition, category_title, documents)
        chunks = chunk_document_paragraphs(
            pack_id=pack_id,
            documents=documents,
            paragraphs_by_document=category_paragraphs[label],
            max_chars_per_chunk=max_chars_per_chunk,
        )
        chunks, quality_report = improve_chunks(
            chunks,
            overlap_chars=chunk_overlap_chars,
            deduplicate=deduplicate,
        )
        qualifies = (
            int(confident_characters[label.value]) >= MIN_CONFIDENT_CHARACTERS
            and int(confident_paragraphs[label.value]) >= MIN_CONFIDENT_PARAGRAPHS
        )
        decisions[label.value] = {
            "qualifies": qualifies,
            "confident_characters": confident_characters[label.value],
            "confident_paragraphs": confident_paragraphs[label.value],
            "chunk_count": len(chunks),
            "reason": "qualified" if qualifies else "insufficient confident content",
        }
        categories[label] = PreparedPackContent(
            documents=documents,
            pack_id=pack_id,
            chunks=chunks,
            cleanup_report=cleanup_report,
            front_matter_report=front_matter_report,
            toc_report=toc_report,
            chunk_quality_report=quality_report,
        )
    return PreparedSplitPackContent(categories, classification_report, decisions)


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


def maybe_remove_front_matter(
    documents: list[ExtractedDocument],
    *,
    remove_front_matter: bool,
    front_matter_max_page: int,
) -> tuple[list[ExtractedDocument], dict[str, object]]:
    if not remove_front_matter:
        return documents, {"enabled": False}
    return remove_front_matter_pages(documents, max_page=front_matter_max_page)


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


def chunk_document_paragraphs(
    *,
    pack_id: str,
    documents: list[ExtractedDocument],
    paragraphs_by_document: dict[str, list[Paragraph]],
    max_chars_per_chunk: int,
) -> list[SourceChunk]:
    chunks: list[SourceChunk] = []
    for document in documents:
        document_chunks = chunk_paragraphs(
            pack_id=pack_id,
            document_id=document.document_id,
            paragraphs=paragraphs_by_document.get(document.document_id, []),
            max_chars=max_chars_per_chunk,
        )
        chunks.extend(reindex_chunks(document_chunks, len(chunks)))
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



