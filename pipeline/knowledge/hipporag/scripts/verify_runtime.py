#!/usr/bin/env python3
"""Verificação sem rede da compatibilidade entre a integração e HippoRAG fixado."""

from __future__ import annotations

import os
import tempfile
from pathlib import Path

os.environ.setdefault("OPENAI_API_KEY", "verification-key-not-used-for-network")

from hipporag.llm.openai_gpt import CacheOpenAI
from hipporag.utils.config_utils import BaseConfig

from project_rag import EMBEDDING_MODEL, MODEL, terra_llm_class


def main() -> int:
    with tempfile.TemporaryDirectory() as directory:
        config = BaseConfig(
            llm_name=MODEL,
            embedding_model_name=EMBEDDING_MODEL,
            embedding_provider="openai",
            temperature=None,
        )
        adapter = terra_llm_class(CacheOpenAI)(
            cache_dir=str(Path(directory) / "cache"),
            global_config=config,
            reasoning_effort="high",
        )
        try:
            params = adapter.llm_config.generate_params
            assert params["model"] == MODEL
            assert params["reasoning_effort"] == "high"
            assert "temperature" not in params
        finally:
            adapter.close()
    print("runtime verification passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
