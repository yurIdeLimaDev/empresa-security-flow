#!/usr/bin/env python3
"""Validate existing host/restore evidence, export only sanitized decisions. No provisioning."""
from __future__ import annotations

import argparse
from datetime import datetime, timezone
import hashlib
import json
from pathlib import Path
import re

from mvp_acceptance import read_json, write_json


def ordinary(root, name):
    path = root / name
    if path.resolve()!=path.absolute() or not path.is_file() or path.stat().st_size>64*1024*1024:
        raise ValueError("missing, linked or oversized evidence")
    if not path.resolve().is_relative_to(root.resolve()): raise ValueError("evidence escaped root")
    return path


def digest(path):
    with path.open("rb") as stream: return hashlib.file_digest(stream,"sha256").hexdigest()


def pairs(path):
    result={}
    for line in path.read_text(encoding="utf-8").splitlines():
        key,sep,value=line.partition("=")
        if not sep or key in result: raise ValueError("invalid evidence key/value")
        result[key]=value
    return result


def fresh(value, now, hours):
    observed=datetime.fromisoformat(value.replace("Z","+00:00"))
    if observed.tzinfo is None: return False
    age=(now-observed).total_seconds()
    return -300 <= age <= hours*3600


def evaluate(root, lock, backup, restore, runner_hash, now=None, max_age=24):
    now=now or datetime.now(timezone.utc)
    root=root.absolute()
    sums=ordinary(root,"SHA256SUMS")
    recorded={}
    for line in sums.read_text(encoding="utf-8").splitlines():
        match=re.fullmatch(r"([a-f0-9]{64})  (?:\./)?([^\r\n]+)",line)
        if not match: raise ValueError("invalid checksums")
        expected,name=match.groups()
        if name in recorded or len(recorded)>=10000 or "\\" in name or ".." in Path(name).parts or Path(name).is_absolute(): raise ValueError("unsafe checksum path")
        if digest(ordinary(root,name))!=expected: raise ValueError("evidence hash mismatch")
        recorded[name]=expected
    required=["host-facts.txt","host-drift/host-drift.txt","supply-chain-report.json","runtime-readiness.json","isolation/isolation-smoke.json"]
    if not set(required)<=recorded.keys(): raise ValueError("incomplete preflight evidence")
    facts=pairs(ordinary(root,required[0])); drift=pairs(ordinary(root,required[1]))
    supply=read_json(ordinary(root,required[2])); runtime=read_json(ordinary(root,required[3])); isolation=read_json(ordinary(root,required[4]))
    expected={tool["name"]:tool for tool in lock["tools"]}
    checks=runtime.get("checks",[])
    tools={row["tool"]:row for row in checks}
    runtime_ok=len(tools)==len(checks) and set(tools)==set(expected)
    for name,tool in expected.items():
        row=tools.get(name,{})
        if tool["execution_class"]=="automatic":
            image=tool["container"]["image"]
            if "@sha256:" not in image: image += "@"+tool["container"]["digest"]
            runtime_ok=runtime_ok and row.get("status")=="passed" and row.get("image")==image
        else: runtime_ok=runtime_ok and row.get("status")=="restricted"
    decisions={
        "manifest_integrity":True,
        "runner_pin":bool(re.fullmatch(r"[a-f0-9]{64}",runner_hash)) and facts.get("runner_sha256")==runner_hash,
        "fresh_preflight":fresh(facts["checked_at"],now,max_age) and fresh(drift["checked_at"],now,max_age) and fresh(runtime["checked_at"],now,max_age) and fresh(supply["checked_at"],now,max_age),
        "host_drift":all(drift.get(k)==v for k,v in {"status":"passed","docker_user_chain":"present","docker_service":"active","security_updates":"enabled","ssh_key_only":"passed","dedicated_roots":"passed","reference_image":"present_by_digest"}.items()),
        "supply_chain":supply.get("approved") is True and len(supply.get("checks",[]))==len(expected) and {r.get("tool") for r in supply["checks"]}==set(expected) and all(r.get("approved") is True for r in supply["checks"]),
        "all_locked_runtimes":runtime.get("approved") is True and runtime_ok,
        "firewall_canaries_proxy":isolation.get("canary")=="negative-blocked-and-positive-allowed" and isolation.get("proxy_request")=="allowed-pinned-request-logged",
        "restore_bound_to_backup":restore.get("verified") is True and type(restore.get("entry_count")) is int and restore["entry_count"]>0 and restore["entry_count"]==backup.get("entry_count") and bool(re.fullmatch(r"[a-f0-9]{64}",restore.get("backup_sha256",""))) and restore["backup_sha256"]==backup.get("ciphertext_sha256"),
        "fresh_restore":fresh(backup["created_at"],now,max_age) and fresh(restore["restored_at"],now,max_age),
    }
    return {"schema_version":1,"checked_at":now.isoformat(),"status":"passed" if all(decisions.values()) else "blocked","checks":decisions,"input_hashes":{name:recorded[name] for name in required},"raw_evidence_included":False,"production_approval":False,"limit":"Integrity of supplied operator evidence, not remote attestation or a substitute for customer authorization."}


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--preflight",type=Path,required=True)
    parser.add_argument("--backup-receipt",type=Path,required=True)
    parser.add_argument("--restore-receipt",type=Path,required=True)
    parser.add_argument("--runner-sha256",required=True)
    parser.add_argument("--lock",type=Path,default=Path(__file__).resolve().parents[1]/"tools.lock.json")
    parser.add_argument("--output",type=Path,required=True)
    args=parser.parse_args()
    if args.output.exists() or args.output.is_symlink(): parser.error("output must be new")
    try:
        report=evaluate(args.preflight,read_json(args.lock),read_json(args.backup_receipt),read_json(args.restore_receipt),args.runner_sha256)
    except (ValueError,KeyError,TypeError,OSError):
        report={"schema_version":1,"status":"blocked","reason":"missing_invalid_or_inconsistent_evidence","production_approval":False}
    write_json(args.output,report)
    print(json.dumps({"status":report["status"],"raw_evidence_included":False}))
    return int(report["status"]!="passed")


if __name__=="__main__": raise SystemExit(main())
