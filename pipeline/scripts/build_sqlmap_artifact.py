#!/usr/bin/env python3
"""Build a deterministic, self-extracting sqlmap Python zip application."""

from __future__ import annotations

import argparse
import os
from pathlib import Path, PurePosixPath
import stat
import zipfile


FIXED_TIMESTAMP = (1980, 1, 1, 0, 0, 0)
IGNORED_PARTS = {".git", "__pycache__", ".pytest_cache"}

LAUNCHER = r'''from __future__ import annotations
import os
from pathlib import Path, PurePosixPath
import runpy
import shutil
import sys
import tempfile
import zipfile

archive = Path(sys.argv[0]).resolve()
with tempfile.TemporaryDirectory(prefix="sqlmap-") as temporary:
    root = Path(temporary)
    with zipfile.ZipFile(archive) as source:
        for member in source.infolist():
            name = PurePosixPath(member.filename)
            if not name.parts or name.parts[0] != "payload" or name.is_absolute() or ".." in name.parts:
                continue
            destination = root.joinpath(*name.parts)
            if member.is_dir():
                destination.mkdir(parents=True, exist_ok=True)
                continue
            destination.parent.mkdir(parents=True, exist_ok=True)
            with source.open(member) as incoming, destination.open("wb") as outgoing:
                shutil.copyfileobj(incoming, outgoing)
    entrypoint = root / "payload" / "sqlmap.py"
    sys.path.insert(0, str(entrypoint.parent))
    sys.argv[0] = str(entrypoint)
    runpy.run_path(str(entrypoint), run_name="__main__")
'''


def zip_info(name: str, mode: int = 0o644) -> zipfile.ZipInfo:
    info = zipfile.ZipInfo(name, FIXED_TIMESTAMP)
    info.compress_type = zipfile.ZIP_DEFLATED
    info.create_system = 3
    info.external_attr = (stat.S_IFREG | mode) << 16
    return info


def source_files(root: Path) -> list[Path]:
    result: list[Path] = []
    for path in root.rglob("*"):
        relative = path.relative_to(root)
        if path.is_symlink():
            raise ValueError(f"symlink não é aceito: {relative}")
        if any(part in IGNORED_PARTS for part in relative.parts):
            continue
        if path.is_file():
            result.append(path)
    return sorted(result, key=lambda item: item.relative_to(root).as_posix())


def build(source: Path, output: Path) -> None:
    if not (source / "sqlmap.py").is_file():
        raise ValueError("source não contém sqlmap.py")
    output.parent.mkdir(parents=True, exist_ok=True)
    temporary = output.with_suffix(output.suffix + ".tmp")
    with zipfile.ZipFile(temporary, "w", compression=zipfile.ZIP_DEFLATED, compresslevel=9) as archive:
        archive.writestr(zip_info("__main__.py"), LAUNCHER.encode("utf-8"))
        for path in source_files(source):
            relative = PurePosixPath("payload", path.relative_to(source).as_posix()).as_posix()
            archive.writestr(zip_info(relative), path.read_bytes())
    os.replace(temporary, output)


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--source", required=True, type=Path)
    parser.add_argument("--output", required=True, type=Path)
    args = parser.parse_args()
    build(args.source.resolve(), args.output.resolve())


if __name__ == "__main__":
    main()
