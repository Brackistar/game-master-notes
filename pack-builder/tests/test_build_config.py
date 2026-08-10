from __future__ import annotations

import json
from pathlib import Path

import pytest

from pack_builder.build_config import BuildOverrides, load_build_config, resolve_build_options


def write_config(path: Path, payload: object) -> None:
    path.write_text(json.dumps(payload), encoding="utf-8")


def base_config(synthetic_pdf: Path, out: Path) -> dict[str, object]:
    return {
        "pdfs": [str(synthetic_pdf)],
        "system": "Config System",
        "edition": "1e",
        "title": "Config Book",
        "out": str(out),
        "embedding_provider": "deterministic",
    }


def test_load_build_config_rejects_invalid_json(tmp_path: Path) -> None:
    config = tmp_path / "bad.json"
    config.write_text("{bad json", encoding="utf-8")

    with pytest.raises(ValueError, match="not valid JSON"):
        load_build_config(config)


def test_load_build_config_rejects_non_object_json(tmp_path: Path) -> None:
    config = tmp_path / "list.json"
    write_config(config, ["not", "an", "object"])

    with pytest.raises(ValueError, match="must be a JSON object"):
        load_build_config(config)


def test_resolve_build_options_rejects_bad_pdfs_field(tmp_path: Path) -> None:
    config = tmp_path / "bad-pdfs.json"
    write_config(
        config,
        {
            "pdfs": "book.pdf",
            "system": "System",
            "edition": "1e",
            "title": "Book",
            "out": str(tmp_path / "out.gmnpack"),
        },
    )

    with pytest.raises(ValueError, match="'pdfs' must be a list"):
        resolve_build_options(config, BuildOverrides())


def test_resolve_build_options_reports_missing_required_values(
    synthetic_pdf: Path, tmp_path: Path
) -> None:
    config = tmp_path / "missing-title.json"
    write_config(
        config,
        {
            "pdfs": [str(synthetic_pdf)],
            "system": "System",
            "edition": "1e",
            "out": str(tmp_path / "out.gmnpack"),
        },
    )

    with pytest.raises(ValueError, match="missing required values: title"):
        resolve_build_options(config, BuildOverrides())


def test_resolve_build_options_rejects_missing_pdf(tmp_path: Path) -> None:
    config = tmp_path / "missing-pdf.json"
    write_config(
        config,
        {
            "pdfs": [str(tmp_path / "missing.pdf")],
            "system": "System",
            "edition": "1e",
            "title": "Book",
            "out": str(tmp_path / "out.gmnpack"),
        },
    )

    with pytest.raises(ValueError, match="PDF does not exist"):
        resolve_build_options(config, BuildOverrides())


def test_resolve_build_options_rejects_too_small_chunk_size(
    synthetic_pdf: Path, tmp_path: Path
) -> None:
    config = tmp_path / "small-chunks.json"
    payload = base_config(synthetic_pdf, tmp_path / "out.gmnpack")
    payload["max_chars_per_chunk"] = 199
    write_config(config, payload)

    with pytest.raises(ValueError, match="must be at least 200"):
        resolve_build_options(config, BuildOverrides())


def test_resolve_build_options_rejects_invalid_overlap(
    synthetic_pdf: Path, tmp_path: Path
) -> None:
    config = tmp_path / "bad-overlap.json"
    payload = base_config(synthetic_pdf, tmp_path / "out.gmnpack")
    payload["max_chars_per_chunk"] = 240
    payload["chunk_overlap_chars"] = 240
    write_config(config, payload)

    with pytest.raises(ValueError, match="smaller than chunk size"):
        resolve_build_options(config, BuildOverrides())


def test_resolve_build_options_rejects_negative_toc_max_page(
    synthetic_pdf: Path, tmp_path: Path
) -> None:
    config = tmp_path / "bad-toc.json"
    payload = base_config(synthetic_pdf, tmp_path / "out.gmnpack")
    payload["toc_max_page"] = -1
    write_config(config, payload)

    with pytest.raises(ValueError, match="must be zero or greater"):
        resolve_build_options(config, BuildOverrides())


def test_resolve_build_options_uses_cli_overrides(
    synthetic_pdf: Path, tmp_path: Path
) -> None:
    config = tmp_path / "build.json"
    config_out = tmp_path / "config.gmnpack"
    override_out = tmp_path / "override.gmnpack"
    write_config(config, base_config(synthetic_pdf, config_out))

    options = resolve_build_options(
        config,
        BuildOverrides(
            title="Override Book",
            out=override_out,
            max_chars_per_chunk=240,
        ),
    )

    assert options.title == "Override Book"
    assert options.out_path == override_out
    assert options.max_chars_per_chunk == 240
