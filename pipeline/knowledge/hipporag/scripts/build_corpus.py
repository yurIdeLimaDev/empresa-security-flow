#!/usr/bin/env python3
"""Constrói um corpus auditável e estritamente limitado para o HippoRAG."""

from __future__ import annotations

import argparse
import hashlib
import json
import re
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Iterable

TEXT_EXTENSIONS = {".md", ".json", ".yaml", ".yml", ".go", ".py", ".ps1", ".sh", ".txt"}
MAX_FILE_BYTES = 512 * 1024
MAX_DOCUMENT_CHARS = 80_000
SECRET_PATTERNS = (
    re.compile(r"-----BEGIN (?:[A-Z ]+ )?PRIVATE KEY-----"),
    re.compile(r"\bsk-[A-Za-z0-9_-]{16,}\b"),
    re.compile(r"\bgh[pousr]_[A-Za-z0-9]{20,}\b"),
    re.compile(r"\b(?:AKIA|ASIA)[A-Z0-9]{16}\b"),
)


@dataclass(frozen=True)
class CorpusDocument:
    idx: int
    source: str
    sha256: str
    text: str


def project_root() -> Path:
    return Path(__file__).resolve().parents[4]


def selected_paths(root: Path) -> Iterable[Path]:
    candidates = [
        root / "README.md",
        root / "AGENTS.md",
        root / "docs",
        root / "pipeline" / "README.md",
        root / "pipeline" / "cmd",
        root / "pipeline" / "config",
        root / "pipeline" / "internal",
        root / "pipeline" / "scripts",
        root / "pipeline" / "go.mod",
        root / "pipeline" / "go.sum",
        root / "pipeline" / "tools.lock.json",
        root / "pipeline" / "sbom.cdx.json",
        root / "pipeline" / "knowledge" / "hipporag" / "README.md",
        root / "correcao" / "README.md",
        root / "correcao" / "config",
        root / "correcao" / "docs",
        root / "validacao" / "2026-08-22-fechamento-prontidao-operacional",
        root / "validacao" / "2026-08-22-isolamento-politicas",
    ]
    for candidate in candidates:
        if candidate.is_file():
            yield candidate
        elif candidate.is_dir():
            yield from candidate.rglob("*")


def is_allowed_file(root: Path, path: Path) -> bool:
    if not path.is_file() or path.suffix.lower() not in TEXT_EXTENSIONS:
        return False
    relative = path.relative_to(root).as_posix()
    excluded_parts = {".git", ".runtime", "data", "output", "tool-artifacts", "casos", "runtime", "worktrees", "__pycache__"}
    if any(part in excluded_parts for part in Path(relative).parts):
        return False
    return path.stat().st_size <= MAX_FILE_BYTES


def read_safe_text(root: Path, path: Path) -> tuple[str, str]:
    raw = path.read_bytes()
    try:
        text = raw.decode("utf-8")
    except UnicodeDecodeError as exc:
        raise ValueError(f"arquivo fora de UTF-8: {path.relative_to(root)}") from exc
    for pattern in SECRET_PATTERNS:
        if pattern.search(text):
            raise ValueError(f"possível segredo no corpus: {path.relative_to(root)}")
    relative = path.relative_to(root).as_posix()
    header = f"Fonte versionada: {relative}\n\n"
    return header + text, hashlib.sha256(raw).hexdigest()


def split_document(text: str) -> list[str]:
    """Separa arquivos extensos em limites legíveis, preservando linhas."""
    if len(text) <= MAX_DOCUMENT_CHARS:
        return [text]
    parts: list[str] = []
    current: list[str] = []
    current_length = 0
    for line in text.splitlines(keepends=True):
        if current and current_length + len(line) > MAX_DOCUMENT_CHARS:
            parts.append("".join(current))
            current = []
            current_length = 0
        current.append(line)
        current_length += len(line)
    if current:
        parts.append("".join(current))
    return parts


def build_corpus(root: Path) -> list[CorpusDocument]:
    root = root.resolve()
    paths = sorted({path.resolve() for path in selected_paths(root) if is_allowed_file(root, path)})
    documents: list[CorpusDocument] = []
    for path in paths:
        text, digest = read_safe_text(root, path)
        source = path.relative_to(root).as_posix()
        parts = split_document(text)
        for part_number, part_text in enumerate(parts, start=1):
            part_source = source if len(parts) == 1 else f"{source}#parte-{part_number:04d}"
            documents.append(
                CorpusDocument(
                    idx=len(documents),
                    source=part_source,
                    sha256=digest,
                    text=part_text,
                )
            )
    if not documents:
        raise ValueError("o corpus ficou vazio")
    return documents


def corpus_fingerprint(documents: list[CorpusDocument]) -> str:
    stable = [{"source": item.source, "sha256": item.sha256} for item in documents]
    return hashlib.sha256(json.dumps(stable, ensure_ascii=False, separators=(",", ":")).encode("utf-8")).hexdigest()


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", type=Path, default=project_root())
    parser.add_argument("--output", type=Path)
    args = parser.parse_args()
    documents = build_corpus(args.root)
    payload = {
        "schema_version": 1,
        "project_root": str(args.root.resolve()),
        "fingerprint": corpus_fingerprint(documents),
        "documents": [asdict(item) for item in documents],
    }
    rendered = json.dumps(payload, ensure_ascii=False, indent=2) + "\n"
    if args.output:
        args.output.parent.mkdir(parents=True, exist_ok=True)
        args.output.write_text(rendered, encoding="utf-8")
    print(json.dumps({"documents": len(documents), "fingerprint": payload["fingerprint"]}, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
