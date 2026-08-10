from __future__ import annotations

from pathlib import Path
import json

from typer.testing import CliRunner

from pack_builder.cli import app


def test_cli_build_inspect_validate(synthetic_pdf: Path, tmp_path: Path) -> None:
    runner = CliRunner()
    out = tmp_path / "cli-built.gmnpack"

    build_result = runner.invoke(
        app,
        [
            "build",
            "--system",
            "Test System",
            "--edition",
            "1e",
            "--title",
            "Synthetic Book",
            "--out",
            str(out),
            "--extractor",
            "pymupdf",
            "--embedding-provider",
            "deterministic",
            str(synthetic_pdf),
        ],
    )

    assert build_result.exit_code == 0, build_result.output
    assert out.exists()

    inspect_result = runner.invoke(app, ["inspect", str(out)])
    assert inspect_result.exit_code == 0, inspect_result.output
    assert "Synthetic Book" in inspect_result.output

    validate_result = runner.invoke(app, ["validate", str(out)])
    assert validate_result.exit_code == 0, validate_result.output
    assert "Pack is valid" in validate_result.output


def test_cli_rejects_overwrite_without_force(
    synthetic_pdf: Path, tmp_path: Path
) -> None:
    runner = CliRunner()
    out = tmp_path / "existing.gmnpack"
    out.write_text("already here", encoding="utf-8")

    result = runner.invoke(
        app,
        [
            "build",
            "--system",
            "Test System",
            "--edition",
            "1e",
            "--title",
            "Synthetic Book",
            "--out",
            str(out),
            "--embedding-provider",
            "deterministic",
            str(synthetic_pdf),
        ],
    )

    assert result.exit_code == 1
    assert "already exists" in result.output
    assert out.read_text(encoding="utf-8") == "already here"


def test_cli_dry_run_does_not_write_pack_and_exports_report(
    synthetic_pdf: Path, tmp_path: Path
) -> None:
    runner = CliRunner()
    out = tmp_path / "dry-run.gmnpack"
    report = tmp_path / "report.json"

    result = runner.invoke(
        app,
        [
            "build",
            "--system",
            "Test System",
            "--edition",
            "1e",
            "--title",
            "Synthetic Book",
            "--out",
            str(out),
            "--embedding-provider",
            "deterministic",
            "--dry-run",
            "--report-out",
            str(report),
            "--max-chars-per-chunk",
            "200",
            str(synthetic_pdf),
        ],
    )

    assert result.exit_code == 0, result.output
    assert "Dry run complete" in result.output
    assert not out.exists()
    payload = json.loads(report.read_text(encoding="utf-8"))
    assert payload["chunking"]["max_chars_per_chunk"] == 200


def test_cli_json_output_for_inspect_and_validate(
    synthetic_pdf: Path, tmp_path: Path
) -> None:
    runner = CliRunner()
    out = tmp_path / "json-output.gmnpack"

    build_result = runner.invoke(
        app,
        [
            "build",
            "--system",
            "Test System",
            "--edition",
            "1e",
            "--title",
            "Synthetic Book",
            "--out",
            str(out),
            "--embedding-provider",
            "deterministic",
            str(synthetic_pdf),
        ],
    )
    assert build_result.exit_code == 0, build_result.output

    inspect_result = runner.invoke(app, ["inspect", "--json", str(out)])
    assert inspect_result.exit_code == 0, inspect_result.output
    inspect_payload = json.loads(inspect_result.output)
    assert inspect_payload["ok"] is True
    assert inspect_payload["manifest"]["title"] == "Synthetic Book"

    validate_result = runner.invoke(app, ["validate", "--json", str(out)])
    assert validate_result.exit_code == 0, validate_result.output
    validate_payload = json.loads(validate_result.output)
    assert validate_payload["ok"] is True
