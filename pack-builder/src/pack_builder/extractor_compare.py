from __future__ import annotations

from pathlib import Path

from pack_builder.pdf_extract import extract_document, get_extractor

EXTRACTORS_TO_COMPARE = ["pymupdf", "pymupdf-layout", "pdfplumber"]


def compare_extractors(pdf_path: Path) -> dict[str, object]:
    documents = {
        name: extract_document(pdf_path, get_extractor(name))
        for name in EXTRACTORS_TO_COMPARE
    }
    page_count = max(document.page_count for document in documents.values())
    pages = [
        compare_page(index, {name: document.pages for name, document in documents.items()})
        for index in range(1, page_count + 1)
    ]
    return {
        "pdf": str(pdf_path),
        "extractors": EXTRACTORS_TO_COMPARE,
        "page_count": page_count,
        "pages": pages,
        "summary": summarize_pages(pages),
    }


def compare_page(index: int, pages_by_extractor: dict[str, list]) -> dict[str, object]:
    row: dict[str, object] = {"page": index}
    for name, pages in pages_by_extractor.items():
        row[f"{name}_characters"] = len(page_text(pages, index).strip())
    row["character_delta"] = int(row["pymupdf_characters"]) - int(
        row["pdfplumber_characters"]
    )
    row["layout_delta"] = int(row["pymupdf-layout_characters"]) - int(
        row["pymupdf_characters"]
    )
    return row


def page_text(pages, page_number: int) -> str:
    if page_number > len(pages):
        return ""
    return pages[page_number - 1].text


def summarize_pages(pages: list[dict[str, object]]) -> dict[str, object]:
    summary = {
        f"{name}_characters": sum(int(page[f"{name}_characters"]) for page in pages)
        for name in EXTRACTORS_TO_COMPARE
    }
    summary["character_delta"] = (
        summary["pymupdf_characters"] - summary["pdfplumber_characters"]
    )
    summary["layout_delta"] = (
        summary["pymupdf-layout_characters"] - summary["pymupdf_characters"]
    )
    return summary
