from __future__ import annotations

from enum import Enum
from pathlib import Path
from typing import Annotated

import typer
from rich.console import Console
from rich.table import Table

from pack_builder.constants import DEFAULT_EMBEDDING_MODEL_ID, DEFAULT_EXTRACTOR
from pack_builder.embeddings import get_embedding_provider
from pack_builder.pack_writer import build_pack
from pack_builder.pdf_extract import get_extractor
from pack_builder.validate import validate_pack


class ExtractorName(str, Enum):
    pymupdf = "pymupdf"
    pdfplumber = "pdfplumber"


class EmbeddingProviderName(str, Enum):
    sentence_transformers = "sentence-transformers"
    deterministic = "deterministic"


app = typer.Typer(no_args_is_help=True)
console = Console()


@app.command()
def build(
    pdfs: Annotated[
        list[Path],
        typer.Argument(
            exists=True,
            file_okay=True,
            dir_okay=False,
            readable=True,
            help="One or more user-owned PDFs to package.",
        ),
    ],
    system: Annotated[str, typer.Option("--system", help="Game system name.")],
    edition: Annotated[str, typer.Option("--edition", help="Game system edition.")],
    title: Annotated[str, typer.Option("--title", help="Pack title.")],
    out: Annotated[Path, typer.Option("--out", help="Output .gmnpack path.")],
    extractor: Annotated[
        ExtractorName,
        typer.Option("--extractor", help="PDF extraction adapter."),
    ] = ExtractorName(DEFAULT_EXTRACTOR),
    language: Annotated[str, typer.Option("--language", help="Pack language code.")] = "en",
    embedding_provider: Annotated[
        EmbeddingProviderName,
        typer.Option(
            "--embedding-provider",
            help="Embedding provider. Deterministic is for tests and local fixtures.",
        ),
    ] = EmbeddingProviderName.sentence_transformers,
    embedding_model: Annotated[
        str,
        typer.Option("--embedding-model", help="Sentence Transformers model id."),
    ] = DEFAULT_EMBEDDING_MODEL_ID,
) -> None:
    """Build a .gmnpack archive from one or more PDFs."""
    try:
        pdf_extractor = get_extractor(extractor.value)
        embeddings = get_embedding_provider(embedding_provider.value, embedding_model)
        manifest = build_pack(
            pdf_paths=pdfs,
            out_path=out,
            title=title,
            system=system,
            edition=edition,
            language=language,
            extractor=pdf_extractor,
            embedding_provider=embeddings,
        )
    except Exception as exc:
        console.print(f"[red]Build failed:[/red] {exc}")
        raise typer.Exit(code=1) from exc

    result = validate_pack(out)
    if not result.ok:
        console.print("[red]Pack was written but failed validation:[/red]")
        for error in result.errors:
            console.print(f"- {error}")
        raise typer.Exit(code=1)

    console.print(f"[green]Wrote[/green] {out}")
    console.print(
        f"Pack {manifest['pack_id']} with {manifest['chunk_count']} chunks "
        f"and {manifest['embedding_dimensions']}d embeddings."
    )


@app.command()
def inspect(
    pack: Annotated[
        Path,
        typer.Argument(exists=True, file_okay=True, dir_okay=False, readable=True),
    ],
) -> None:
    """Print pack metadata and validation status."""
    result = validate_pack(pack)
    if not result.manifest:
        console.print("[red]Could not read manifest.[/red]")
        for error in result.errors:
            console.print(f"- {error}")
        raise typer.Exit(code=1)

    manifest = result.manifest
    table = Table(title=str(pack))
    table.add_column("Field")
    table.add_column("Value")
    for field_name in [
        "pack_id",
        "title",
        "system",
        "edition",
        "language",
        "schema_version",
        "extractor_name",
        "embedding_model_id",
        "embedding_dimensions",
        "chunk_count",
        "created_at",
    ]:
        table.add_row(field_name, str(manifest.get(field_name, "")))
    console.print(table)

    if result.ok:
        console.print("[green]Validation: ok[/green]")
    else:
        console.print("[red]Validation: failed[/red]")
        for error in result.errors:
            console.print(f"- {error}")
        raise typer.Exit(code=1)


@app.command()
def validate(
    pack: Annotated[
        Path,
        typer.Argument(exists=True, file_okay=True, dir_okay=False, readable=True),
    ],
) -> None:
    """Validate a .gmnpack archive."""
    result = validate_pack(pack)
    if result.ok:
        console.print("[green]Pack is valid.[/green]")
        for warning in result.warnings:
            console.print(f"[yellow]Warning:[/yellow] {warning}")
        return

    console.print("[red]Pack is invalid.[/red]")
    for error in result.errors:
        console.print(f"- {error}")
    raise typer.Exit(code=1)


if __name__ == "__main__":
    app()
