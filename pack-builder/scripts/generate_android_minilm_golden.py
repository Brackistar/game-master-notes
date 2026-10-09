from __future__ import annotations

import json
from pathlib import Path

from sentence_transformers import SentenceTransformer

MODEL_ID = "sentence-transformers/all-MiniLM-L6-v2"
REVISION = "1110a243fdf4706b3f48f1d95db1a4f5529b4d41"
TEXT = "The Silver Ladder guards the hidden library."


def main() -> None:
    repository = Path(__file__).resolve().parents[2]
    output = (
        repository
        / "android/core/retrieval/src/androidTest/assets/minilm_golden.json"
    )
    output.parent.mkdir(parents=True, exist_ok=True)
    model = SentenceTransformer(MODEL_ID, revision=REVISION)
    vector = model.encode([TEXT], normalize_embeddings=True)[0]
    output.write_text(
        json.dumps(
            {
                "model_id": MODEL_ID,
                "revision": REVISION,
                "text": TEXT,
                "vector": [round(float(value), 8) for value in vector],
            }
        ),
        encoding="utf-8",
    )
    print(output)


if __name__ == "__main__":
    main()
