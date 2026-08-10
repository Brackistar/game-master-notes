from __future__ import annotations

from pack_builder.pack_archive.schema_contract import pack_schema_contract


def test_schema_contract_lists_required_files_and_fields() -> None:
    contract = pack_schema_contract()

    assert contract["schema_version"] == "1.0"
    assert "manifest.json" in contract["archive_files"]
    assert "pack_id" in contract["manifest_required_fields"]
    assert "chunk_id" in contract["chunk_required_fields"]



