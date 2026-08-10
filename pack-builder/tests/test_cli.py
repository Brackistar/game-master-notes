from __future__ import annotations

from pathlib import Path

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
