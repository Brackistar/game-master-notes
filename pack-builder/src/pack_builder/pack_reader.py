from __future__ import annotations

import json
import zipfile
from pathlib import Path

from pack_builder.constants import (
    CHUNKS_FILE,
    EXTRACTION_REPORT_FILE,
    MANIFEST_FILE,
)


def read_json_member(pack_path: Path, member_name: str) -> dict[str, object]:
    with zipfile.ZipFile(pack_path, mode="r") as archive:
        with archive.open(member_name) as handle:
            return json.loads(handle.read().decode("utf-8"))


def read_manifest(pack_path: Path) -> dict[str, object]:
    return read_json_member(pack_path, MANIFEST_FILE)


def read_extraction_report(pack_path: Path) -> dict[str, object]:
    return read_json_member(pack_path, EXTRACTION_REPORT_FILE)


def read_chunks(pack_path: Path) -> list[dict[str, object]]:
    with zipfile.ZipFile(pack_path, mode="r") as archive:
        with archive.open(CHUNKS_FILE) as handle:
            lines = handle.read().decode("utf-8").splitlines()
    return [json.loads(line) for line in lines if line.strip()]
