#!/usr/bin/env python3
"""CLI local para indexar e consultar o fluxo com HippoRAG 2."""

from __future__ import annotations

import argparse
import json
import os
import shutil
import sys
from dataclasses import asdict
from pathlib import Path

from build_corpus import CorpusDocument, build_corpus, corpus_fingerprint, project_root

UPSTREAM_COMMIT = "2f52a86dd04e4633703bd2fb3bb6a37683ac3cfb"
MODEL = "gpt-5.6-terra"
EMBEDDING_MODEL = "text-embedding-3-large"
INDEX_IDENTITY = f"empresa-security-hipporag-v1:{UPSTREAM_COMMIT}:{MODEL}:reasoning-high:{EMBEDDING_MODEL}"


def integration_root() -> Path:
    return Path(__file__).resolve().parents[1]


def data_root() -> Path:
    return integration_root() / "data"


def index_root() -> Path:
    return data_root() / "index"


def marker_path() -> Path:
    return data_root() / "project-index.json"


def require_api_key() -> None:
    if not os.environ.get("OPENAI_API_KEY"):
        raise SystemExit("OPENAI_API_KEY não está definida. Defina a chave somente na sessão atual do terminal e tente novamente.")


def import_runtime():
    try:
        from hipporag import HippoRAG
        from hipporag.llm.openai_gpt import CacheOpenAI
        from hipporag.utils.config_utils import BaseConfig
        from hipporag.utils.misc_utils import Chunk
    except ImportError as exc:
        raise SystemExit("Runtime ausente. Execute .\\bootstrap.ps1 antes de indexar ou consultar.") from exc
    return HippoRAG, CacheOpenAI, BaseConfig, Chunk


def terra_llm_class(cache_openai):
    class TerraHighCacheOpenAI(cache_openai):
        def __init__(self, *args, reasoning_effort: str, **kwargs):
            self._reasoning_effort = reasoning_effort
            super().__init__(*args, **kwargs)

        def _init_llm_config(self):
            super()._init_llm_config()
            self.llm_config.generate_params["reasoning_effort"] = self._reasoning_effort

    return TerraHighCacheOpenAI


def config(base_config):
    return base_config(
        llm_name=MODEL,
        embedding_model_name=EMBEDDING_MODEL,
        embedding_provider="openai",
        temperature=None,
        max_new_tokens=1024,
        preprocess_chunk_max_token_size=1200,
        preprocess_chunk_overlap_token_size=120,
        openie_max_workers=2,
        openie_ner_max_tokens=384,
        openie_triple_max_tokens=768,
        retrieval_top_k=12,
        qa_top_k=6,
        save_dir=str(index_root()),
        save_openie=True,
    )


def create_rag():
    require_api_key()
    hipporag, cache_openai, base_config, chunk = import_runtime()
    runtime_config = config(base_config)
    terra_high = terra_llm_class(cache_openai)
    llm = terra_high(
        cache_dir=str(index_root() / "llm_cache"),
        global_config=runtime_config,
        reasoning_effort="high",
        max_retries=runtime_config.max_retry_attempts,
    )
    rag = hipporag(
        global_config=runtime_config,
        extraction_llm=llm,
        qa_llm=llm,
        index_identity=INDEX_IDENTITY,
    )
    return rag, llm, chunk


def write_marker(documents: list[CorpusDocument]) -> None:
    data_root().mkdir(parents=True, exist_ok=True)
    marker_path().write_text(
        json.dumps(
            {
                "schema_version": 1,
                "corpus_fingerprint": corpus_fingerprint(documents),
                "document_count": len(documents),
                "model": MODEL,
                "reasoning_effort": "high",
                "embedding_model": EMBEDDING_MODEL,
                "hipporag_commit": UPSTREAM_COMMIT,
                "index_identity": INDEX_IDENTITY,
            },
            ensure_ascii=False,
            indent=2,
        )
        + "\n",
        encoding="utf-8",
    )


def assert_fresh_index(documents: list[CorpusDocument]) -> None:
    if not marker_path().is_file():
        raise SystemExit("Índice ausente. Execute .\\run.ps1 index primeiro.")
    marker = json.loads(marker_path().read_text(encoding="utf-8"))
    if marker.get("corpus_fingerprint") != corpus_fingerprint(documents):
        raise SystemExit("O índice está desatualizado em relação aos arquivos permitidos. Execute .\\run.ps1 rebuild.")
    if marker.get("index_identity") != INDEX_IDENTITY:
        raise SystemExit("O índice foi criado com configuração incompatível. Execute .\\run.ps1 rebuild.")


def clear_index() -> None:
    target = index_root().resolve()
    allowed = data_root().resolve()
    if target.parent != allowed or target.name != "index":
        raise RuntimeError("destino de reconstrução inválido")
    if target.exists():
        shutil.rmtree(target)
    if marker_path().exists():
        marker_path().unlink()


def index(documents: list[CorpusDocument], rebuild: bool) -> None:
    if index_root().exists() or marker_path().exists():
        if not rebuild:
            raise SystemExit("Já existe um índice. Use .\\run.ps1 rebuild para recriá-lo explicitamente.")
        clear_index()
    rag, llm, chunk = create_rag()
    try:
        chunks = [chunk(content=item.text, source_id=item.source, metadata={"source": item.source, "sha256": item.sha256}) for item in documents]
        rag.index(chunks)
        write_marker(documents)
    finally:
        rag.close()
        llm.close()
    print(json.dumps({"status": "indexed", "documents": len(documents), "fingerprint": corpus_fingerprint(documents)}, ensure_ascii=False))


def ask(documents: list[CorpusDocument], question: str) -> None:
    assert_fresh_index(documents)
    rag, llm, _ = create_rag()
    try:
        retrieved = rag.retrieve([question], num_to_retrieve=6)[0]
        results, _, _ = rag.rag_qa([retrieved])
        answer = results[0]
        print("Resposta:\n" + (answer.answer or "Sem resposta gerada."))
        print("\nFontes recuperadas:")
        metadata = answer.doc_metadata or []
        scores = answer.doc_scores.tolist() if answer.doc_scores is not None else []
        for position, source in enumerate(metadata):
            source_name = source.get("source", source.get("source_id", "fonte sem metadados"))
            score = f" score={scores[position]:.4f}" if position < len(scores) else ""
            print(f"- {source_name}{score}")
    finally:
        rag.close()
        llm.close()


def main() -> int:
    parser = argparse.ArgumentParser()
    subcommands = parser.add_subparsers(dest="command", required=True)
    subcommands.add_parser("corpus")
    subcommands.add_parser("index")
    subcommands.add_parser("rebuild")
    ask_parser = subcommands.add_parser("ask")
    ask_parser.add_argument("question")
    args = parser.parse_args()
    documents = build_corpus(project_root())
    if args.command == "corpus":
        print(json.dumps({"documents": len(documents), "fingerprint": corpus_fingerprint(documents)}, ensure_ascii=False))
    elif args.command == "index":
        index(documents, rebuild=False)
    elif args.command == "rebuild":
        index(documents, rebuild=True)
    else:
        ask(documents, args.question)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
