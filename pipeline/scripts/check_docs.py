#!/usr/bin/env python3
"""Check UTF-8, headings and inline local links in selected current documents.

No network, file edits, anchor validation, legal approval or private case reads.
Markdown examples in code fences/inline code are deliberately ignored.
"""

from __future__ import annotations

import argparse
import json
import re
from pathlib import Path
from urllib.parse import unquote, urlsplit


CURRENT_DOCUMENTS = (
    "README.md",
    "pipeline/README.md",
    "correcao/README.md",
    "correcao/docs/GERACAO_PATCHES_SEM_PROVEDOR.md",
    "correcao/docs/FLUXO_DE_CORRECAO_SEGURA.md",
    "docs/ARQUITETURA_IMPLEMENTADA.md",
    "docs/DIAGRAMAS_MERMAID.md",
    "docs/LANDING_COMERCIAL_E_OPERACAO_PUBLICA.md",
    "docs/ROADMAP_TECNOLOGIA_NEGOCIOS_2026-09-11.md",
    "docs/VALIDACAO_LOCAL_MVP.md",
    "docs/ESTADO_ATUAL.md",
    "correcao/avaliacao/README.md",
    "deploy/linux/ACEITACAO.md",
    "deploy/linux/README.md",
    "negocio/README.md",
    "negocio/GUIA_DE_EXECUCAO_DO_TITULAR_PARA_MVP.md",
)
FLOW_DOCUMENTS = (
    "README.md",
    "pipeline/README.md",
    "correcao/README.md",
    "correcao/docs/GERACAO_PATCHES_SEM_PROVEDOR.md",
    "correcao/docs/FLUXO_DE_CORRECAO_SEGURA.md",
    "docs/ARQUITETURA_IMPLEMENTADA.md",
    "docs/DIAGRAMAS_MERMAID.md",
    "docs/ESTADO_FLUXO.md",
    "deploy/linux/ACEITACAO.md",
    "deploy/linux/README.md",
)
LINK = re.compile(r"!?\[[^\]\n]*\]\(\s*(?:<([^>\n]+)>|([^\s)]+))(?:\s+\"[^\"]*\")?\s*\)")


def prose_lines(text: str):
    fence = None
    for number, line in enumerate(text.splitlines(), 1):
        marker = re.match(r"^\s*(`{3,}|~{3,})", line)
        if marker:
            current = marker.group(1)
            if fence is None:
                fence = current
            elif current[0] == fence[0] and len(current) >= len(fence):
                fence = None
            continue
        if fence is None:
            yield number, re.sub(r"(`+).*?\1", "", line)


def check_document(root: Path, path: Path) -> list[str]:
    relative = path.relative_to(root).as_posix()
    if path.is_symlink() or not path.is_file() or not path.resolve().is_relative_to(root):
        return [f"{relative}: missing document or symbolic link"]
    try:
        text = path.read_text(encoding="utf-8", errors="strict")
    except UnicodeError:
        return [f"{relative}: invalid UTF-8"]
    errors = []
    if "\ufffd" in text:
        errors.append(f"{relative}: replacement character in text")
    if not text.startswith("# "):
        errors.append(f"{relative}: missing main heading")
    for number, line in prose_lines(text):
        # Reference-style definitions are paths too; remote citations remain offline.
        definition = re.match(r'^\s*\[(?!\^)[^\]]+\]:\s*(?:<([^>]+)>|(\S+))', line)
        if definition:
            target = definition.group(1) or definition.group(2)
            parts = urlsplit(target)
            if not parts.scheme and not parts.netloc and parts.path:
                destination = (path.parent / unquote(parts.path)).resolve()
                if not destination.is_relative_to(root) or not destination.exists():
                    errors.append(f"{relative}:{number}: unresolved reference link: {target}")
        for match in LINK.finditer(line):
            target = match.group(1) or match.group(2)
            parts = urlsplit(target)
            if parts.scheme or parts.netloc or not parts.path:
                continue
            destination = (path.parent / unquote(parts.path)).resolve()
            if not destination.is_relative_to(root) or not destination.exists():
                errors.append(f"{relative}:{number}: unresolved local link: {target}")
    return errors


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[2])
    parser.add_argument("--profile", choices=("all", "flow"), default="all")
    args = parser.parse_args()
    root = args.root.resolve()
    paths = [root / name for name in (CURRENT_DOCUMENTS if args.profile == "all" else FLOW_DOCUMENTS)]
    if args.profile == "all":
        # Keep filled client dossiers out of the selection. Only versioned models.
        legal = root / "negocio/juridico"
        if not legal.is_dir() or legal.is_symlink():
            print(json.dumps({"status": "failed", "errors": ["missing canonical legal models"]}))
            return 1
        paths.extend(sorted(legal.glob("*.md")))
    # Explicit public documentation directories only: never client cases,
    # runtime evidence, hidden sync repositories or generated corpora.
    directories = ("docs", "docs/runbooks", "correcao/docs", "correcao/avaliacao", "deploy/linux")
    if args.profile == "all":
        directories += ("negocio", "negocio/oferta")
    for directory in directories:
        folder = root / directory
        if folder.is_symlink():
            print(json.dumps({"status":"failed","errors":["linked documentation directory"]}))
            return 1
        paths.extend(sorted(folder.glob("*.md")))
    # The engine may also be published separately from the landing repository.
    if (root / "landing-page").is_dir():
        paths.extend(root / name for name in ("landing-page/README.md", "landing-page/DEPLOYMENT.md", "landing-page/PORTAL_CLIENTE.md"))
    paths = sorted(set(paths))
    errors = [error for path in paths for error in check_document(root, path)]
    print(json.dumps({"status": "failed" if errors else "passed", "documents": len(paths),
                      "errors": errors, "checks": ["utf8", "heading", "inline_local_link_paths", "reference_definition_paths"],
                      "legal_approval": False}, ensure_ascii=False, indent=2))
    return int(bool(errors))


if __name__ == "__main__":
    raise SystemExit(main())
