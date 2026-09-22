#!/usr/bin/env python3
"""Reproduce non-invasive local checks; save only status, counts and public SBOM."""
from __future__ import annotations

import argparse
from datetime import datetime, timezone
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys

from mvp_acceptance import write_json

ROOT=Path(__file__).resolve().parents[2]


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output",required=True,type=Path)
    parser.add_argument("--bash",help="optional exact Bash executable (Git Bash works for syntax only)")
    args=parser.parse_args()
    destination=args.output.absolute()
    if destination.exists() or destination.is_symlink():parser.error("output must be new")
    destination.mkdir(parents=True,mode=0o700)
    npm=shutil.which("npm.cmd" if os.name=="nt" else "npm")
    results=[]

    def run(name,command,cwd,timeout=660,input_text=None):
        try:
            result=subprocess.run(command,cwd=cwd,capture_output=True,text=True,encoding="utf-8",errors="replace",timeout=timeout,input=input_text)
            row={"check":name,"status":"passed" if result.returncode==0 else "failed","exit_code":result.returncode}
            results.append(row)
            print(json.dumps(row),flush=True)
            return result.stdout
        except (OSError,subprocess.TimeoutExpired):
            results.append({"check":name,"status":"failed","reason":"unavailable_or_timeout"})
            return ""

    pipeline=ROOT/"pipeline";landing=ROOT/"landing-page"
    run("go_mod_verify",["go","mod","verify"],pipeline)
    run("go_vet",["go","vet","./..."],pipeline)
    run("go_tests_race",["go","test","-race","./...","-count=1","-timeout=10m"],pipeline)
    run("supply_chain_sbom",["go","run","./cmd/pipeline","supply-chain","--lock","tools.lock.json","--sbom","sbom.cdx.json","--strict"],pipeline)
    run("python_tests",[sys.executable,"-m","unittest","discover","-s","scripts/tests","-p","test_*.py","-v"],pipeline)
    docs=run("documentation",[sys.executable,"scripts/check_docs.py"],pipeline)
    try: write_json(destination/"documents.json",json.loads(docs))
    except ValueError: pass
    run("corpus_tests",[sys.executable,"-m","unittest","knowledge/hipporag/tests/test_build_corpus.py","-v"],pipeline)
    run("corpus_build",[sys.executable,"knowledge/hipporag/scripts/build_corpus.py"],pipeline)
    bash=args.bash or shutil.which("bash")
    for script in sorted((pipeline/"scripts").glob("*.sh")):
        if bash:
            run("syntax_"+script.name,[bash,"-n"],pipeline,input_text=script.read_text(encoding="utf-8"))
        else: results.append({"check":"syntax_"+script.name,"status":"not_run"})
    if npm:
        run("landing_lint",[npm,"run","lint"],landing)
        run("landing_types",[npm,"exec","--","tsc","--noEmit"],landing)
        run("landing_build_tests",[npm,"test"],landing)
        audit=run("npm_audit_all",[npm,"audit","--json"],landing)
        try:
            parsed=json.loads(audit)
            write_json(destination/"npm-audit.json",{"counts":parsed["metadata"]["vulnerabilities"],"includes_development":True})
        except (ValueError,KeyError): pass
        sbom=run("npm_sbom",[npm,"sbom","--sbom-format","cyclonedx"],landing)
        try:
            parsed=json.loads(sbom)
            # Keep only the public dependency inventory, not machine metadata.
            write_json(destination/"landing-sbom.cdx.json",{k:parsed[k] for k in ("bomFormat","specVersion","version","components")})
        except (ValueError,KeyError): results.append({"check":"sbom_contract","status":"failed"})
    else:results.append({"check":"npm","status":"not_run"})
    passed=all(r["status"]=="passed" for r in results)
    write_json(destination/"checks.json",{"schema_version":1,"checked_at":datetime.now(timezone.utc).isoformat(),"status":"passed" if passed else "failed","checks":results,"raw_logs_saved":False,"production_approval":False})
    return int(not passed)


if __name__=="__main__":raise SystemExit(main())
