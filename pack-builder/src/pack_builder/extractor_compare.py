from __future__ import annotations

from pathlib import Path

from pack_builder.pdf_extract import extract_document, get_extractor


def compare_extractors(pdf_path: Path) -> dict[str, object]:
    pymupdf = extract_document(pdf_path, get_extractor("pymupdf"))
    pdfplumber = extract_document(pdf_path, get_extractor("pdfplumber"))
    page_count = max(pymupdf.page_count, pdfplumber.page_count)
    pages = [
        compare_page(index, pymupdf.pages, pdfplumber.pages)
        for index in range(1, page_count + 1)
    ]
    return {
        "pdf": str(pdf_path),
        "extractors": ["pymupdf", "pdfplumber"],
        "page_count": page_count,
        "pages": pages,
        "summary": summarize_pages(pages),
    }


def compare_page(index: int, pymupdf_pages, pdfplumber_pages) -> dict[str, object]:
    pymupdf_text = page_text(pymupdf_pages, index)
    pdfplumber_text = page_text(pdfplumber_pages, index)
    return {
        "page": index,
        "pymupdf_characters": len(pymupdf_text.strip()),
        "pdfplumber_characters": len(pdfplumber_text.strip()),
        "character_delta": len(pymupdf_text.strip()) - len(pdfplumber_text.strip()),
    }


def page_text(pages, page_number: int) -> str:
    if page_number > len(pages):
        return ""
    return pages[page_number - 1].text


def summarize_pages(pages: list[dict[str, object]]) -> dict[str, object]:
    pymupdf_total = sum(int(page["pymupdf_characters"]) for page in pages)
    pdfplumber_total = sum(int(page["pdfplumber_characters"]) for page in pages)
    return {
        "pymupdf_characters": pymupdf_total,
        "pdfplumber_characters": pdfplumber_total,
        "character_delta": pymupdf_total - pdfplumber_total,
    }
