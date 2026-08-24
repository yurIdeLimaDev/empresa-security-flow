#!/usr/bin/env python3
"""Verificação local, sem chave e sem rede, do runtime HippoRAG fixado."""

from __future__ import annotations

from contextlib import redirect_stdout
import sys
import tempfile
from pathlib import Path

from hipporag import HippoRAG
from hipporag.utils.config_utils import BaseConfig
from hipporag.utils.misc_utils import Chunk

from project_rag import BACKEND, INDEX_IDENTITY, LocalGraphExtractor, LocalTokenHashEmbedding


def main() -> int:
    with tempfile.TemporaryDirectory() as directory:
        config = BaseConfig(
            llm_name=BACKEND,
            embedding_model_name=BACKEND,
            save_dir=str(Path(directory) / "index"),
            preprocess_chunk_max_token_size=None,
            retrieval_top_k=1,
            qa_top_k=1,
        )
        embedding = LocalTokenHashEmbedding()
        extractor = LocalGraphExtractor()
        with redirect_stdout(sys.stderr):
            rag = HippoRAG(
                global_config=config,
                embedding_model=embedding,
                extraction_llm=extractor,
                qa_llm=extractor,
                index_identity=INDEX_IDENTITY,
            )
            try:
                rag.index([Chunk("Pipeline 1 produz evidência pública.", source_id="smoke.md")])
                result = rag.retrieve(["Qual pipeline produz evidência?"], num_to_retrieve=1)[0]
                assert len(result.docs) == 1
            finally:
                rag.close()
    print("runtime verification passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
