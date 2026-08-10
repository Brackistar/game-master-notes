from __future__ import annotations

import hashlib
import re
from pathlib import Path
from typing import Protocol

from pack_builder.layout_order import TextBlock, order_text_blocks
from pack_builder.models import ExtractedDocument, ExtractedPage


class PdfExtractor(Protocol):
    name: str

    def extract_pages(self, pdf_path: Path) -> list[ExtractedPage]:
        """Extract text from a PDF while preserving 1-based page numbers."""


class PyMuPdfExtractor:
    name = "pymupdf"

    def extract_pages(self, pdf_path: Path) -> list[ExtractedPage]:
        import pymupdf

        pages: list[ExtractedPage] = []
        with pymupdf.open(pdf_path) as document:
            for index, page in enumerate(document, start=1):
                try:
                    text = page.get_text("text", sort=True) or ""
                    pages.append(ExtractedPage(page_number=index, text=text))
                except Exception as exc:  # pragma: no cover - depends on damaged PDFs.
                    pages.append(
                        ExtractedPage(
                            page_number=index,
                            text="",
                            warnings=[f"page extraction failed: {exc}"],
                        )
                    )
        return pages


class PyMuPdfLayoutExtractor:
    name = "pymupdf-layout"

    def extract_pages(self, pdf_path: Path) -> list[ExtractedPage]:
        import pymupdf

        pages: list[ExtractedPage] = []
        with pymupdf.open(pdf_path) as document:
            for index, page in enumerate(document, start=1):
                try:
                    blocks = pymupdf_text_blocks(page)
                    text = order_text_blocks(blocks, page.rect.width)
                    pages.append(ExtractedPage(page_number=index, text=text))
                except Exception as exc:  # pragma: no cover - depends on damaged PDFs.
                    pages.append(
                        ExtractedPage(
                            page_number=index,
                            text="",
                            warnings=[f"layout extraction failed: {exc}"],
                        )
                    )
        return pages


class PdfPlumberExtractor:
    name = "pdfplumber"

    def extract_pages(self, pdf_path: Path) -> list[ExtractedPage]:
        import pdfplumber

        pages: list[ExtractedPage] = []
        with pdfplumber.open(pdf_path) as document:
            for index, page in enumerate(document.pages, start=1):
                try:
                    text = page.extract_text(layout=True) or ""
                    pages.append(ExtractedPage(page_number=index, text=text))
                except Exception as exc:  # pragma: no cover - depends on damaged PDFs.
                    pages.append(
                        ExtractedPage(
                            page_number=index,
                            text="",
                            warnings=[f"page extraction failed: {exc}"],
                        )
                    )
        return pages


def get_extractor(name: str) -> PdfExtractor:
    normalized = name.lower()
    if normalized == PyMuPdfExtractor.name:
        return PyMuPdfExtractor()
    if normalized == PyMuPdfLayoutExtractor.name:
        return PyMuPdfLayoutExtractor()
    if normalized == PdfPlumberExtractor.name:
        return PdfPlumberExtractor()
    raise ValueError(f"unknown PDF extractor: {name}")


def pymupdf_text_blocks(page) -> list[TextBlock]:
    blocks: list[TextBlock] = []
    for raw_block in page.get_text("blocks", sort=False):
        if len(raw_block) < 5:
            continue
        x0, y0, x1, y1, text = raw_block[:5]
        blocks.append(TextBlock(float(x0), float(y0), float(x1), float(y1), str(text)))
    return blocks


def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for block in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def slugify(value: str) -> str:
    slug = re.sub(r"[^a-zA-Z0-9]+", "-", value.strip().lower()).strip("-")
    return slug or "document"


def document_id_for(path: Path, checksum: str) -> str:
    return f"{slugify(path.stem)}-{checksum[:12]}"


def extract_document(pdf_path: Path, extractor: PdfExtractor) -> ExtractedDocument:
    resolved = pdf_path.resolve()
    checksum = sha256_file(resolved)
    pages = extractor.extract_pages(resolved)
    return ExtractedDocument(
        document_id=document_id_for(resolved, checksum),
        source_path=resolved,
        source_filename=resolved.name,
        source_checksum=checksum,
        page_count=len(pages),
        pages=pages,
    )
