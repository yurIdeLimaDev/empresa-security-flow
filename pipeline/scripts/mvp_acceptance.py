#!/usr/bin/env python3
"""Offline MVP rehearsal and synthetic patch evaluation. Never calls an AI API."""
from __future__ import annotations

import argparse
from datetime import datetime, timezone
import hashlib
import json
import math
import os
from pathlib import Path
import subprocess

PIPELINE = Path(__file__).resolve().parents[1]
GATES = {
    "incomplete_ai_configuration": ["TestMVPMissingGeneratorInputsBlockBeforeNetwork", "TestPatchGenerationDisabledFailsBeforeWorktreeOrNetwork"],
    "authorization": ["TestOnboardingDraftsStayBlockedWithoutAuthorization", "TestMVPFirstCustomerRehearsal"],
    "context_allowlist": ["TestPatchConfigRejectsUnsafeSettings", "TestPatchSourceRejectsUntrackedSymlinkAndOversize"],
    "credential_disclosure": ["TestPatchDisclosureBlocksBeforeTransport", "TestPatchAutomationBlocksCredentialDisclosure"],
    "final_review_and_delivery": ["TestPatchAutomationReachesOnlyFinalHumanReview", "TestMVPFirstCustomerRehearsal", "TestDeliveryPackageIsBoundAndDeterministic"],
    "non_regression_and_bounds": ["TestPatchAutomationRejectsResponsesAndKeepsBEST", "TestMVPSyntheticPatchEvaluation"],
}


def read_json(path: Path, limit=2_000_000):
    if path.is_symlink() or not path.is_file() or path.stat().st_size > limit:
        raise ValueError("input must be a bounded ordinary file")
    def pairs(items):
        result = {}
        for key, value in items:
            if key in result: raise ValueError("duplicate JSON key")
            result[key] = value
        return result
    return json.loads(path.read_text(encoding="utf-8"), object_pairs_hook=pairs,
                      parse_constant=lambda _: (_ for _ in ()).throw(ValueError("nonfinite JSON")))


def write_json(path, value):
    with Path(path).open("x", encoding="utf-8", newline="\n") as stream:
        json.dump(value, stream, ensure_ascii=False, indent=2, allow_nan=False)
        stream.write("\n")


def summarize(poc, pricing=None):
    rows = poc["cases"]
    if len(rows) != 7 or len({r["case"] for r in rows}) != 7:
        raise ValueError("incomplete/duplicate evaluation cases")
    positive = [r for r in rows if r["expected_accept"] is True]
    negative = [r for r in rows if r["expected_accept"] is False]
    summary = {"suite_sha256": poc["suite_sha256"], "mode": poc["mode"],
               "positive_cases": len(positive), "negative_cases": len(negative),
               "correction_acceptance": sum(r["accepted"] is True for r in positive) / len(positive),
               "unsafe_acceptances": sum(r["accepted"] is True for r in negative),
               "expectations_passed": all(r["expectation_met"] is True for r in rows),
               "human_gate_preserved": all(r["human_review_required"] is True for r in rows),
               "bounded": all(r["bounded"] is True for r in rows),
               "calls": sum(r["calls"] for r in rows), "cost_estimate": None,
               "cost_note": "No provider selected or invoked. Unknown price is not zero.",
               "scope": poc["scope"]}
    if pricing is not None:
        numeric = ["input_per_million", "output_per_million", "assumed_input_tokens_per_call", "assumed_output_tokens_per_call"]
        for key in numeric:
            value = pricing[key]
            if type(value) not in (int, float) or not math.isfinite(value) or value < 0: raise ValueError("invalid pricing assumption")
        for key in ("currency", "source", "as_of", "provider_model"):
            if not isinstance(pricing[key], str) or not pricing[key].strip() or len(pricing[key])>300: raise ValueError("missing pricing provenance")
        per_call = (pricing["input_per_million"]*pricing["assumed_input_tokens_per_call"] + pricing["output_per_million"]*pricing["assumed_output_tokens_per_call"])/1_000_000
        summary["cost_estimate"] = {"per_call": per_call, "suite": per_call*summary["calls"], "assumptions": pricing}
        summary["cost_note"] = "Scenario estimate from declared token/rate assumptions, not measured billing; excludes taxes, infra and caching."
    return summary


def compare(current, previous):
    if current["suite_sha256"] != previous["suite_sha256"] or current["scope"] != previous["scope"]:
        raise ValueError("cannot compare different suites/gates")
    # No synthetic score may compensate for an unsafe acceptance or bypass.
    acceptable = (current["unsafe_acceptances"] == 0 and current["bounded"] is True
                  and current["human_gate_preserved"] is True and current["expectations_passed"] is True
                  and current["correction_acceptance"] >= previous["correction_acceptance"])
    return {"comparable": True, "non_regressing": acceptable,
            "correction_acceptance_delta": current["correction_acceptance"]-previous["correction_acceptance"],
            "production_approval": False}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", required=True, type=Path, help="new directory for sanitized reports")
    parser.add_argument("--response", type=Path, help="optional offline JSON with candidate_source only")
    parser.add_argument("--pricing", type=Path, help="optional explicit price/token assumptions, never fetched")
    parser.add_argument("--compare", type=Path, help="previous poc-summary.json from same evaluator")
    args = parser.parse_args()
    output = args.output.absolute()
    if output.exists() or output.is_symlink(): parser.error("output must not exist")
    pricing = read_json(args.pricing) if args.pricing else None
    previous = read_json(args.compare) if args.compare else None
    if args.response: read_json(args.response, 16384)
    output.mkdir(parents=True, mode=0o700)
    env = dict(os.environ)
    env.pop("MVP_POC_RESPONSE", None)
    env["MVP_REPORT_DIR"] = str(output)
    if args.response: env["MVP_POC_RESPONSE"] = str(args.response.resolve())
    pattern = "^(TestMVP|TestPatch|TestOnboardingDrafts|TestDelivery)"
    command = ["go", "test", "./internal/app", "-run", pattern, "-json", "-count=1", "-timeout=10m"]
    events = {}
    try:
        run = subprocess.run(command, cwd=PIPELINE, env=env, capture_output=True, text=True, encoding="utf-8", errors="replace", timeout=660)
        for line in run.stdout.splitlines():
            try: event = json.loads(line)
            except ValueError: continue
            if event.get("Test") and event.get("Action") in ("pass", "fail", "skip"):
                events[event["Test"]] = event["Action"]
        success = run.returncode == 0
    except (subprocess.TimeoutExpired, OSError):
        success = False
    gates = {gate: {"passed": all(events.get(test)=="pass" for test in tests), "tests": tests} for gate,tests in GATES.items()}
    success = success and all(row["passed"] for row in gates.values())
    report = {"schema_version":1,"checked_at":datetime.now(timezone.utc).isoformat(),"status":"passed" if success else "failed", "gates":gates,"tests":events,"remote_services":False,"production_approval":False}
    write_json(output/"gate-matrix.json",report)
    if (output/"poc.json").is_file():
        summary = summarize(read_json(output/"poc.json"),pricing)
        write_json(output/"poc-summary.json",summary)
        if previous:
            comparison=compare(summary,previous); write_json(output/"comparison.json",comparison)
            success=success and comparison["non_regressing"]
    else: success=False
    print(json.dumps({"status":"passed" if success else "failed","reports":str(output),"raw_logs_saved":False}))
    return 0 if success else 1


if __name__ == "__main__":
    raise SystemExit(main())
