from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class TextBlock:
    x0: float
    y0: float
    x1: float
    y1: float
    text: str


def order_text_blocks(blocks: list[TextBlock], page_width: float) -> str:
    usable_blocks = [block for block in blocks if block.text.strip()]
    if not usable_blocks:
        return ""
    columns = split_columns(usable_blocks, page_width)
    return "\n\n".join(block.text.strip() for column in columns for block in column)


def split_columns(
    blocks: list[TextBlock],
    page_width: float,
) -> list[list[TextBlock]]:
    if not looks_like_two_columns(blocks, page_width):
        return [sort_top_to_bottom(blocks)]

    midpoint = page_width / 2
    full_width = [block for block in blocks if spans_midpoint(block, midpoint)]
    column_blocks = [block for block in blocks if block not in full_width]
    left = [block for block in column_blocks if block_center(block) < midpoint]
    right = [block for block in column_blocks if block_center(block) >= midpoint]
    return merge_full_width_blocks(
        full_width,
        [sort_top_to_bottom(left), sort_top_to_bottom(right)],
    )


def looks_like_two_columns(blocks: list[TextBlock], page_width: float) -> bool:
    if len(blocks) < 6:
        return False
    midpoint = page_width / 2
    left = [block for block in blocks if block_center(block) < midpoint]
    right = [block for block in blocks if block_center(block) >= midpoint]
    if len(left) < 3 or len(right) < 3:
        return False
    return not all(spans_midpoint(block, midpoint) for block in blocks)


def merge_full_width_blocks(
    full_width: list[TextBlock],
    columns: list[list[TextBlock]],
) -> list[list[TextBlock]]:
    if not full_width:
        return columns
    leading = [
        block
        for block in sort_top_to_bottom(full_width)
        if all(not column or block.y1 <= column[0].y0 for column in columns)
    ]
    trailing = [block for block in full_width if block not in leading]
    merged: list[list[TextBlock]] = []
    if leading:
        merged.append(leading)
    merged.extend(columns)
    if trailing:
        merged.append(sort_top_to_bottom(trailing))
    return [column for column in merged if column]


def sort_top_to_bottom(blocks: list[TextBlock]) -> list[TextBlock]:
    return sorted(blocks, key=lambda block: (round(block.y0, 1), round(block.x0, 1)))


def block_center(block: TextBlock) -> float:
    return (block.x0 + block.x1) / 2


def spans_midpoint(block: TextBlock, midpoint: float) -> bool:
    return block.x0 < midpoint < block.x1
