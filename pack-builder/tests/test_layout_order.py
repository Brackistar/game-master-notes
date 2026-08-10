from __future__ import annotations

from pack_builder.pdf_extraction.layout_order import TextBlock, order_text_blocks


def block(x0: float, y0: float, x1: float, y1: float, text: str) -> TextBlock:
    return TextBlock(x0=x0, y0=y0, x1=x1, y1=y1, text=text)


def test_order_text_blocks_keeps_single_column_top_to_bottom() -> None:
    text = order_text_blocks(
        [
            block(20, 80, 200, 100, "second"),
            block(20, 20, 200, 40, "first"),
        ],
        page_width=400,
    )

    assert text == "first\n\nsecond"


def test_order_text_blocks_reads_left_column_before_right_column() -> None:
    blocks = [
        block(250, 20, 380, 40, "right one"),
        block(20, 80, 170, 100, "left two"),
        block(250, 80, 380, 100, "right two"),
        block(20, 20, 170, 40, "left one"),
        block(20, 140, 170, 160, "left three"),
        block(250, 140, 380, 160, "right three"),
    ]

    text = order_text_blocks(blocks, page_width=400)

    assert text == (
        "left one\n\nleft two\n\nleft three\n\n"
        "right one\n\nright two\n\nright three"
    )


def test_order_text_blocks_keeps_heading_before_columns() -> None:
    blocks = [
        block(20, 10, 380, 30, "Chapter Title"),
        block(20, 60, 170, 80, "left one"),
        block(250, 60, 380, 80, "right one"),
        block(20, 100, 170, 120, "left two"),
        block(250, 100, 380, 120, "right two"),
        block(20, 140, 170, 160, "left three"),
        block(250, 140, 380, 160, "right three"),
    ]

    text = order_text_blocks(blocks, page_width=400)

    assert text.startswith("Chapter Title\n\nleft one")



