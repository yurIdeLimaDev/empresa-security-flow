"""Exercita indexação e recuperação HippoRAG sem chave, rede ou modelo remoto."""

from __future__ import annotations

import hashlib
import tempfile
import unittest
from pathlib import Path

import numpy as np

from hipporag import HippoRAG
from hipporag.utils.config_utils import BaseConfig
from hipporag.utils.misc_utils import Chunk


class FakeEmbedding:
    def batch_encode(self, texts, **_kwargs):
        if isinstance(texts, str):
            texts = [texts]
        vectors = []
        for text in texts:
            digest = hashlib.sha256(text.encode("utf-8")).digest()
            vector = np.asarray([(byte + 1) / 255 for byte in digest[:16]], dtype=np.float32)
            vectors.append(vector / np.linalg.norm(vector))
        return np.vstack(vectors)

    def close(self):
        pass


class FakeLLM:
    def infer(self, messages=None, **_kwargs):
        if not messages:
            raise ValueError("messages é obrigatório")
        return (
            '{"named_entities":["Pipeline 1","evidência"],"triples":[["Pipeline 1","produz","evidência"]]}',
            {"prompt_tokens": 0, "completion_tokens": 0, "total_tokens": 0, "finish_reason": "stop"},
            False,
        )

    def close(self):
        pass


class HippoRAGOfflineSmokeTests(unittest.TestCase):
    def test_index_and_retrieve_without_network(self):
        with tempfile.TemporaryDirectory() as directory:
            config = BaseConfig(
                llm_name="offline-smoke",
                embedding_model_name="offline-smoke",
                save_dir=str(Path(directory) / "index"),
                synonymy_edge_topk=2,
                retrieval_top_k=2,
                qa_top_k=1,
                preprocess_chunk_max_token_size=None,
            )
            embedding = FakeEmbedding()
            llm = FakeLLM()
            rag = HippoRAG(
                global_config=config,
                embedding_model=embedding,
                extraction_llm=llm,
                qa_llm=llm,
                index_identity="offline-smoke-v1",
            )
            try:
                rag.index(
                    [
                        Chunk("Pipeline 1 produz evidência pública.", source_id="docs/pipeline.md"),
                        Chunk("Pipeline 2 depende de autorização documentada.", source_id="docs/policy.md"),
                    ]
                )
                result = rag.retrieve(["Qual pipeline produz evidência?"], num_to_retrieve=1)[0]
                self.assertEqual(len(result.docs), 1)
                self.assertEqual(len(result.doc_metadata), 1)
                self.assertIn("source_id", result.doc_metadata[0])
            finally:
                rag.close()


if __name__ == "__main__":
    unittest.main()
