from __future__ import annotations

import hashlib
from typing import Protocol

import numpy as np

from pack_builder.constants import (
    DEFAULT_EMBEDDING_DIMENSIONS,
    DEFAULT_EMBEDDING_MODEL_ID,
)


class EmbeddingProvider(Protocol):
    model_id: str
    dimensions: int

    def encode(self, texts: list[str]) -> np.ndarray:
        """Return one float32 vector per input text."""


class SentenceTransformerEmbeddingProvider:
    def __init__(self, model_id: str = DEFAULT_EMBEDDING_MODEL_ID) -> None:
        from sentence_transformers import SentenceTransformer

        self.model_id = model_id
        self._model = SentenceTransformer(model_id)
        dimension = sentence_transformer_dimensions(self._model)
        self.dimensions = int(dimension or DEFAULT_EMBEDDING_DIMENSIONS)

    def encode(self, texts: list[str]) -> np.ndarray:
        embeddings = self._model.encode(
            texts,
            convert_to_numpy=True,
            normalize_embeddings=True,
            show_progress_bar=False,
        )
        return np.asarray(embeddings, dtype=np.float32)


def sentence_transformer_dimensions(model: object) -> int | None:
    get_dimension = getattr(model, "get_embedding_dimension", None)
    if callable(get_dimension):
        return get_dimension()

    get_legacy_dimension = getattr(model, "get_sentence_embedding_dimension", None)
    if callable(get_legacy_dimension):
        return get_legacy_dimension()

    return None


class DeterministicEmbeddingProvider:
    """Offline test provider that preserves the pack contract without model files."""

    model_id = "deterministic-test-embedding"

    def __init__(self, dimensions: int = DEFAULT_EMBEDDING_DIMENSIONS) -> None:
        self.dimensions = dimensions

    def encode(self, texts: list[str]) -> np.ndarray:
        rows: list[np.ndarray] = []
        for text in texts:
            seed = int.from_bytes(hashlib.sha256(text.encode("utf-8")).digest()[:8], "big")
            generator = np.random.default_rng(seed)
            vector = generator.normal(size=self.dimensions).astype(np.float32)
            norm = np.linalg.norm(vector)
            if norm:
                vector = vector / norm
            rows.append(vector)
        if not rows:
            return np.empty((0, self.dimensions), dtype=np.float32)
        return np.vstack(rows).astype(np.float32)


def get_embedding_provider(
    provider_name: str,
    model_id: str = DEFAULT_EMBEDDING_MODEL_ID,
) -> EmbeddingProvider:
    normalized = provider_name.lower()
    if normalized == "sentence-transformers":
        return SentenceTransformerEmbeddingProvider(model_id=model_id)
    if normalized == "deterministic":
        return DeterministicEmbeddingProvider()
    raise ValueError(f"unknown embedding provider: {provider_name}")
