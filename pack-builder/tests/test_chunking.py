from __future__ import annotations

from pack_builder.chunking import (
    chunk_pages,
    citation_label,
    deterministic_chunk_id,
    normalize_text,
)
from pack_builder.models import ExtractedPage


def test_normalize_text_collapses_whitespace_without_losing_paragraphs() -> None:
    raw = "  First   line\r\nsecond line\n\n\n\tThird   line  "

    assert normalize_text(raw) == "First line\nsecond line\n\nThird line"


def test_chunk_pages_is_paragraph_aware_and_keeps_page_ranges() -> None:
    pages = [
        ExtractedPage(1, "First paragraph.\n\nSecond paragraph."),
        ExtractedPage(2, "Third paragraph."),
    ]

    chunks = chunk_pages(
        pack_id="pack-a",
        document_id="doc-a",
        pages=pages,
        max_chars=45,
    )

    assert [chunk.page_start for chunk in chunks] == [1, 2]
    assert [chunk.page_end for chunk in chunks] == [1, 2]
    assert chunks[0].text == "First paragraph.\n\nSecond paragraph."
    assert chunks[1].text == "Third paragraph."


def test_ids_and_citation_labels_are_deterministic() -> None:
    first = deterministic_chunk_id("pack", "doc", 1, 2, 0)
    second = deterministic_chunk_id("pack", "doc", 1, 2, 0)

    assert first == second
    assert first.startswith("chunk-")
    assert citation_label("doc", 3, 3) == "doc p. 3"
    assert citation_label("doc", 3, 5) == "doc pp. 3-5"
