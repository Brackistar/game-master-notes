from __future__ import annotations

from pack_builder.embedding_generation.embeddings import sentence_transformer_dimensions


def test_sentence_transformer_dimensions_prefers_new_method() -> None:
    class FakeModel:
        def get_embedding_dimension(self) -> int:
            return 384

        def get_sentence_embedding_dimension(self) -> int:
            raise AssertionError("deprecated method should not be called")

    assert sentence_transformer_dimensions(FakeModel()) == 384


def test_sentence_transformer_dimensions_falls_back_to_legacy_method() -> None:
    class FakeModel:
        def get_sentence_embedding_dimension(self) -> int:
            return 128

    assert sentence_transformer_dimensions(FakeModel()) == 128



