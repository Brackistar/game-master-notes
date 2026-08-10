from __future__ import annotations

from pack_builder.core_domain.models import (
    ExtractedDocument,
    ExtractedPage,
    replace_document_pages,
)

FRONT_MATTER_TERMS = [
    "credits",
    "writers:",
    "developer:",
    "editors:",
    "art director:",
    "copyright",
    "all rights reserved",
    "no part of this publication",
]


def remove_front_matter_pages(
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
            if page.page_number <= max_page and is_front_matter_page(page.text):
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


def is_front_matter_page(text: str) -> bool:
    normalized = text.lower()
    matches = sum(1 for term in FRONT_MATTER_TERMS if term in normalized)
    return matches >= 2
