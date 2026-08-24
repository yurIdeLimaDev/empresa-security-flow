#!/usr/bin/env python3
"""Índice HippoRAG local, sem chave de API, para consulta pelo Codex MCP.

O modelo do Codex formula a resposta. Este módulo só recupera trechos do
corpus permitido. A construção do grafo e os embeddings são determinísticos,
locais e auditáveis; não há chamadas de rede ou envio de conteúdo do projeto.
"""

from __future__ import annotations

import argparse
from contextlib import redirect_stdout
import hashlib
import json
import re
import shutil
import sys
from collections import Counter
from pathlib import Path
from typing import Any

import numpy as np

from build_corpus import CorpusDocument, build_corpus, corpus_fingerprint, project_root

UPSTREAM_COMMIT = "2f52a86dd04e4633703bd2fb3bb6a37683ac3cfb"
BACKEND = "local-deterministic-graph-v1"
INDEX_IDENTITY = f"empresa-security-hipporag-v2:{UPSTREAM_COMMIT}:{BACKEND}"
TOKEN_PATTERN = re.compile(r"[A-Za-zÀ-ÖØ-öø-ÿ_][A-Za-zÀ-ÖØ-öø-ÿ0-9_-]{2,}")
STOPWORDS = frozenset(
    "a ao aos as aonde aquele aquela aqueles aquelas aqui assim com como da das de do dos e em entre esta este estes estas "
    "foi foram ha isso isto ja mas nao nas no nos o os ou para pela pelas pelo pelos por que se sem ser sua suas seu seus "
    "sao tem uma umas um uns na the and for from into not with this that these those are was were be by or of to in on it".split()
)
VECTOR_DIMENSION = 768


def integration_root() -> Path:
    return Path(__file__).resolve().parents[1]


def data_root() -> Path:
    return integration_root() / "data"


def index_root() -> Path:
    return data_root() / "index"


def marker_path() -> Path:
    return data_root() / "project-index.json"


def import_runtime():
    try:
        from hipporag import HippoRAG
        from hipporag.utils.config_utils import BaseConfig
        from hipporag.utils.misc_utils import Chunk
    except ImportError as exc:
        raise SystemExit("Runtime HippoRAG ausente. Execute .\\bootstrap.ps1 antes de iniciar o MCP.") from exc
    return HippoRAG, BaseConfig, Chunk


def tokens(text: str) -> list[str]:
    return [token.casefold() for token in TOKEN_PATTERN.findall(text) if token.casefold() not in STOPWORDS]


class LocalTokenHashEmbedding:
    """Embedding local por hashing de termos e bigramas; não usa rede."""

    def __init__(self, dimension: int = VECTOR_DIMENSION):
        self.dimension = dimension

    def batch_encode(self, texts: list[str] | str, **_kwargs: Any) -> np.ndarray:
        if isinstance(texts, str):
            texts = [texts]
        vectors: list[np.ndarray] = []
        for text in texts:
            vector = np.zeros(self.dimension, dtype=np.float32)
            sequence = tokens(text)
            features = sequence + [f"{left}:{right}" for left, right in zip(sequence, sequence[1:])]
            for feature in features:
                digest = hashlib.blake2b(feature.encode("utf-8"), digest_size=8).digest()
                bucket = int.from_bytes(digest[:4], "big") % self.dimension
                vector[bucket] += 1.0 if digest[4] & 1 else -1.0
            norm = float(np.linalg.norm(vector))
            vectors.append(vector / norm if norm else vector)
        return np.vstack(vectors)

    def close(self) -> None:
        return None


class LocalGraphExtractor:
    """Extrai um grafo lexical limitado e determinístico para o HippoRAG."""

    def infer(self, messages: list[dict[str, Any]] | None = None, **_kwargs: Any):
        content = ""
        if messages:
            last_message = messages[-1]
            raw_content = last_message.get("content", "") if isinstance(last_message, dict) else str(last_message)
            if isinstance(raw_content, list):
                content = " ".join(str(item.get("text", "")) if isinstance(item, dict) else str(item) for item in raw_content)
            else:
                content = str(raw_content)
        frequency = Counter(tokens(content))
        entities = [term for term, _ in frequency.most_common(24)]
        triples: list[list[str]] = []
        for line in content.splitlines():
            line_terms = list(dict.fromkeys(tokens(line)))[:8]
            for left, right in zip(line_terms, line_terms[1:]):
                if left != right:
                    triples.append([left, "relacionado_a", right])
                if len(triples) >= 48:
                    break
            if len(triples) >= 48:
                break
        payload = json.dumps({"named_entities": entities, "triples": triples}, ensure_ascii=False)
        return payload, {"prompt_tokens": 0, "completion_tokens": 0, "total_tokens": 0, "finish_reason": "stop"}, False

    def close(self) -> None:
        return None


# O HippoRAG registra a proveniência pelo módulo da classe. Mantenha-a estável
# tanto ao executar este arquivo diretamente quanto quando o MCP o importa.
LocalTokenHashEmbedding.__module__ = "project_rag"
LocalGraphExtractor.__module__ = "project_rag"


def config(base_config):
    return base_config(
        llm_name=BACKEND,
        embedding_model_name=BACKEND,
        save_dir=str(index_root()),
        synonymy_edge_topk=8,
        retrieval_top_k=8,
        qa_top_k=4,
        preprocess_chunk_max_token_size=1200,
        preprocess_chunk_overlap_token_size=80,
        openie_max_workers=1,
        save_openie=True,
    )


def create_rag():
    hipporag, base_config, chunk = import_runtime()
    runtime_config = config(base_config)
    embedding = LocalTokenHashEmbedding()
    extractor = LocalGraphExtractor()
    rag = hipporag(
        global_config=runtime_config,
        embedding_model=embedding,
        extraction_llm=extractor,
        qa_llm=extractor,
        index_identity=INDEX_IDENTITY,
    )
    return rag, chunk


def marker(documents: list[CorpusDocument]) -> dict[str, Any]:
    return {
        "schema_version": 2,
        "corpus_fingerprint": corpus_fingerprint(documents),
        "document_count": len(documents),
        "backend": BACKEND,
        "network": "none",
        "hipporag_commit": UPSTREAM_COMMIT,
        "index_identity": INDEX_IDENTITY,
    }


def write_marker(documents: list[CorpusDocument]) -> None:
    data_root().mkdir(parents=True, exist_ok=True)
    marker_path().write_text(json.dumps(marker(documents), ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def index_is_current(documents: list[CorpusDocument]) -> bool:
    if not marker_path().is_file() or not index_root().is_dir():
        return False
    try:
        saved = json.loads(marker_path().read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return False
    expected = marker(documents)
    return all(saved.get(key) == value for key, value in expected.items())


def clear_index() -> None:
    target = index_root().resolve()
    allowed = data_root().resolve()
    if target.parent != allowed or target.name != "index":
        raise RuntimeError("destino de reconstrução inválido")
    if target.exists():
        shutil.rmtree(target)
    if marker_path().exists():
        marker_path().unlink()


def build_index(documents: list[CorpusDocument], rebuild: bool = False) -> dict[str, Any]:
    if index_is_current(documents):
        return {"status": "current", "documents": len(documents), "fingerprint": corpus_fingerprint(documents)}
    if index_root().exists() or marker_path().exists():
        if not rebuild:
            raise RuntimeError("índice incompatível; solicite reconstrução explícita")
        clear_index()
    with redirect_stdout(sys.stderr):
        rag, chunk = create_rag()
        try:
            chunks = [chunk(content=item.text, source_id=item.source, metadata={"source": item.source, "sha256": item.sha256}) for item in documents]
            rag.index(chunks)
            write_marker(documents)
        finally:
            rag.close()
    return {"status": "indexed", "documents": len(documents), "fingerprint": corpus_fingerprint(documents)}


def ensure_current_index() -> tuple[list[CorpusDocument], bool]:
    documents = build_corpus(project_root())
    rebuilt = not index_is_current(documents)
    if rebuilt:
        build_index(documents, rebuild=True)
    return documents, rebuilt


def excerpt(text: str, query: str, limit: int = 1400) -> str:
    candidates = tokens(query)
    lowered = text.casefold()
    positions = [lowered.find(candidate) for candidate in candidates if lowered.find(candidate) >= 0]
    start = max(0, (min(positions) if positions else 0) - limit // 4)
    end = min(len(text), start + limit)
    rendered = text[start:end].strip()
    return ("…" if start else "") + rendered + ("…" if end < len(text) else "")


def search(question: str, top_k: int = 4) -> dict[str, Any]:
    if not question or not question.strip():
        raise ValueError("a pergunta não pode estar vazia")
    top_k = max(1, min(int(top_k), 6))
    documents, rebuilt = ensure_current_index()
    with redirect_stdout(sys.stderr):
        rag, _ = create_rag()
        try:
            retrieved = rag.retrieve([question], num_to_retrieve=top_k)[0]
            scores = retrieved.doc_scores.tolist() if retrieved.doc_scores is not None else []
            snippets = []
            for position, document in enumerate(retrieved.docs):
                metadata = retrieved.doc_metadata[position] if position < len(retrieved.doc_metadata) else {}
                source = metadata.get("source", metadata.get("source_id", "fonte sem metadados"))
                snippets.append({"source": source, "score": round(float(scores[position]), 5) if position < len(scores) else None, "excerpt": excerpt(str(document), question)})
        finally:
            rag.close()
    return {
        "question": question,
        "index_rebuilt": rebuilt,
        "corpus_fingerprint": corpus_fingerprint(documents),
        "network": "none",
        "sources": snippets,
        "instruction": "Responda apenas com base nos trechos. Se estiverem insuficientes, diga isso e abra somente a fonte indicada necessária para confirmar.",
    }


def status() -> dict[str, Any]:
    documents = build_corpus(project_root())
    return {
        "current": index_is_current(documents),
        "corpus_documents": len(documents),
        "corpus_fingerprint": corpus_fingerprint(documents),
        "backend": BACKEND,
        "network": "none",
        "automatic_refresh": "cada consulta MCP reconstrói o índice se o corpus permitido mudou",
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    commands = parser.add_subparsers(dest="command", required=True)
    commands.add_parser("corpus")
    commands.add_parser("index")
    commands.add_parser("rebuild")
    search_parser = commands.add_parser("search")
    search_parser.add_argument("question")
    search_parser.add_argument("--top-k", type=int, default=4)
    args = parser.parse_args()
    if args.command == "corpus":
        print(json.dumps(status(), ensure_ascii=False))
    elif args.command == "index":
        documents = build_corpus(project_root())
        print(json.dumps(build_index(documents), ensure_ascii=False))
    elif args.command == "rebuild":
        documents = build_corpus(project_root())
        clear_index()
        print(json.dumps(build_index(documents, rebuild=True), ensure_ascii=False))
    else:
        print(json.dumps(search(args.question, args.top_k), ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
