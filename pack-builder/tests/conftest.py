from __future__ import annotations

from pathlib import Path

import pytest


@pytest.fixture()
def synthetic_pdf(tmp_path: Path) -> Path:
    from reportlab.lib.pagesizes import letter
    from reportlab.pdfgen import canvas

    path = tmp_path / "synthetic-sourcebook.pdf"
    pdf = canvas.Canvas(str(path), pagesize=letter)
    pdf.drawString(72, 720, "Chapter One")
    pdf.drawString(72, 700, "The city of Glass Harbor stands beside a silver river.")
    pdf.drawString(72, 680, "Its guilds trade moonlit maps and careful secrets.")
    pdf.showPage()
    pdf.drawString(72, 720, "Chapter Two")
    pdf.drawString(72, 700, "The Crimson Archive records every oath made by nobles.")
    pdf.drawString(72, 680, "A hidden clerk changes one word each winter.")
    pdf.save()
    return path
