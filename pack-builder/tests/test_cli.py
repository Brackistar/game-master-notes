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


def test_cli_build_from_json_config(synthetic_pdf: Path, tmp_path: Path) -> None:
    runner = CliRunner()
    out = tmp_path / "config-built.gmnpack"
    config = tmp_path / "build-config.json"
    config.write_text(
        json.dumps(
            {
                "pdfs": [str(synthetic_pdf)],
                "system": "Config System",
                "edition": "2e",
                "title": "Configured Book",
                "out": str(out),
                "embedding_provider": "deterministic",
                "max_chars_per_chunk": 200,
            }
        ),
        encoding="utf-8",
    )

    result = runner.invoke(app, ["build", "--config", str(config)])

    assert result.exit_code == 0, result.output
    assert out.exists()
    inspect_result = runner.invoke(app, ["inspect", "--json", str(out)])
    payload = json.loads(inspect_result.output)
    assert payload["manifest"]["title"] == "Configured Book"
    assert payload["manifest"]["system"] == "Config System"


def test_cli_flags_override_json_config(synthetic_pdf: Path, tmp_path: Path) -> None:
    runner = CliRunner()
    config_out = tmp_path / "config-built.gmnpack"
    override_out = tmp_path / "override-built.gmnpack"
    config = tmp_path / "build-config.json"
    config.write_text(
        json.dumps(
            {
                "pdfs": [str(synthetic_pdf)],
                "system": "Config System",
                "edition": "2e",
                "title": "Configured Book",
                "out": str(config_out),
                "embedding_provider": "deterministic",
            }
        ),
        encoding="utf-8",
    )

    result = runner.invoke(
        app,
        [
            "build",
            "--config",
            str(config),
            "--title",
            "Override Book",
            "--out",
            str(override_out),
        ],
    )

    assert result.exit_code == 0, result.output
    assert override_out.exists()
    assert not config_out.exists()
    inspect_result = runner.invoke(app, ["inspect", "--json", str(override_out)])
    payload = json.loads(inspect_result.output)
    assert payload["manifest"]["title"] == "Override Book"


def test_cli_report_and_sample_chunks(synthetic_pdf: Path, tmp_path: Path) -> None:
    runner = CliRunner()
    out = tmp_path / "sample.gmnpack"
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

    report_result = runner.invoke(app, ["report", "--json", str(out)])
    assert report_result.exit_code == 0, report_result.output
    report_payload = json.loads(report_result.output)
    assert report_payload["extractor"] == "pymupdf"

    sample_result = runner.invoke(
        app, ["sample-chunks", "--json", "--limit", "1", str(out)]
    )
    assert sample_result.exit_code == 0, sample_result.output
    sample_payload = json.loads(sample_result.output)
    assert sample_payload["count"] == 1
    assert "citation_label" in sample_payload["chunks"][0]


def test_cli_schema_outputs_contract() -> None:
    runner = CliRunner()

    result = runner.invoke(app, ["schema", "--json"])

    assert result.exit_code == 0, result.output
    payload = json.loads(result.output)
    assert payload["schema_version"] == "1.0"
    assert "chunks.jsonl" in payload["archive_files"]


def test_cli_compare_extractors_outputs_summary(
    synthetic_pdf: Path,
) -> None:
    runner = CliRunner()

    result = runner.invoke(
        app,
        ["compare-extractors", "--json", str(synthetic_pdf)],
    )

    assert result.exit_code == 0, result.output
    payload = json.loads(result.output)
    assert payload["extractors"] == ["pymupdf", "pdfplumber"]
    assert payload["page_count"] == 2
    assert "character_delta" in payload["summary"]
