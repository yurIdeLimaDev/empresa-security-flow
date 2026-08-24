#!/usr/bin/env python3
"""Servidor MCP stdio que expõe o índice local do projeto ao Codex."""

from __future__ import annotations

import json
import sys
import traceback
from typing import Any

from project_rag import search, status

SERVER_INFO = {"name": "project_knowledge", "version": "1.0.0"}
PROTOCOL_VERSION = "2025-03-26"

if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8", errors="strict")
if hasattr(sys.stdin, "reconfigure"):
    sys.stdin.reconfigure(encoding="utf-8", errors="strict")
if hasattr(sys.stderr, "reconfigure"):
    sys.stderr.reconfigure(encoding="utf-8", errors="backslashreplace")


def send(message: dict[str, Any]) -> None:
    sys.stdout.write(json.dumps(message, ensure_ascii=False, separators=(",", ":")) + "\n")
    sys.stdout.flush()


def tool_result(payload: dict[str, Any]) -> dict[str, Any]:
    rendered = json.dumps(payload, ensure_ascii=False, indent=2)
    return {"content": [{"type": "text", "text": rendered}], "structuredContent": payload}


TOOLS = [
    {
        "name": "search",
        "description": "Consulta o conhecimento versionado do projeto pelo índice HippoRAG local. Sempre use antes de varrer arquivos para perguntas sobre arquitetura, políticas, fluxo, ferramentas, evidências ou decisões. Retorna no máximo seis trechos e suas fontes. Não acessa rede.",
        "inputSchema": {
            "type": "object",
            "properties": {
                "query": {"type": "string", "description": "Pergunta específica sobre o projeto."},
                "top_k": {"type": "integer", "minimum": 1, "maximum": 6, "default": 4, "description": "Quantidade máxima de trechos a retornar."},
            },
            "required": ["query"],
            "additionalProperties": False,
        },
    },
    {
        "name": "status",
        "description": "Informa o estado e o fingerprint do índice local. Não acessa rede.",
        "inputSchema": {"type": "object", "properties": {}, "additionalProperties": False},
    },
]


def handle(request: dict[str, Any]) -> dict[str, Any] | None:
    method = request.get("method")
    request_id = request.get("id")
    if method == "notifications/initialized":
        return None
    if method == "initialize":
        result = {
            "protocolVersion": PROTOCOL_VERSION,
            "capabilities": {"tools": {"listChanged": False}},
            "serverInfo": SERVER_INFO,
            "instructions": "Use search antes de consultar arquivos do projeto. O índice é reconstruído automaticamente quando o corpus permitido muda.",
        }
    elif method == "tools/list":
        result = {"tools": TOOLS}
    elif method == "tools/call":
        params = request.get("params") or {}
        name = params.get("name")
        arguments = params.get("arguments") or {}
        if name == "search":
            result = tool_result(search(str(arguments.get("query", "")), int(arguments.get("top_k", 4))))
        elif name == "status":
            result = tool_result(status())
        else:
            raise ValueError(f"ferramenta desconhecida: {name}")
    else:
        return {"jsonrpc": "2.0", "id": request_id, "error": {"code": -32601, "message": f"método desconhecido: {method}"}}
    return {"jsonrpc": "2.0", "id": request_id, "result": result}


def main() -> int:
    for raw in sys.stdin:
        try:
            request = json.loads(raw)
            response = handle(request)
            if response is not None:
                send(response)
        except Exception as exc:  # Mantém o protocolo vivo e não vaza stack trace ao modelo.
            request_id = request.get("id") if "request" in locals() and isinstance(request, dict) else None
            if request_id is not None:
                send({"jsonrpc": "2.0", "id": request_id, "error": {"code": -32603, "message": f"falha no índice local: {exc}"}})
            print(traceback.format_exc(), file=sys.stderr, flush=True)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
