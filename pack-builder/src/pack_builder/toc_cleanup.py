from __future__ import annotations

import re

from pack_builder.models import ExtractedDocument, ExtractedPage, replace_document_pages

MIN_TOC_LINES = 5
MIN_INLINE_TOC_REFS = 8
PAGE_REF_PATTERN = re.compile(r"\b[\w'’][\w'’:\- ]{2,80}?\s+\d{1,3}\b")


def remove_toc_pages(
    documents: list[ExtractedDocument],
    *,
    max_page: int,
) -> tuple[list[ExtractedDocument], dict[str, object]]:
    cleaned_documents: list[ExtractedDocument] = []
    removed_pages: dict[str, list[int]] = {}

    for document in documents:
        kept_pages: list[ExtractedPage] = []
        removed: list[int] = []
        for page in document.pages:
            if page.page_number <= max_page and is_toc_page(page.text):
                removed.append(page.page_number)
                continue
            kept_pages.append(page)
        removed_pages[document.document_id] = removed
        cleaned_documents.append(replace_document_pages(document, kept_pages))

    return cleaned_documents, {
        "enabled": True,
        "max_page": max_page,
        "removed_pages": removed_pages,
    }


def is_toc_page(text: str) -> bool:
    normalized = text.lower()
    if "table of contents" in normalized:
        return True
    lines = [line for line in text.splitlines() if line.strip()]
    return has_dense_toc_lines(lines) or count_inline_page_refs(text) >= MIN_INLINE_TOC_REFS


def has_dense_toc_lines(lines: list[str]) -> bool:
    if len(lines) < MIN_TOC_LINES:
        return False
    toc_like = sum(1 for line in lines if looks_like_toc_line(line))
    return toc_like >= max(MIN_TOC_LINES, len(lines) // 2)


def looks_like_toc_line(line: str) -> bool:
    stripped = line.strip()
    if len(stripped) < 6:
        return False
    return bool(re.search(r"\b\d{1,3}\b\s*$", stripped))


def count_inline_page_refs(text: str) -> int:
    return len(PAGE_REF_PATTERN.findall(text))
