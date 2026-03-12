#!/usr/bin/env python3
"""
extract_test_failures.py — run go test -json, write failures to .claude/bugs/open/

Usage:
    ./scripts/extract_test_failures.py ./api/cmd/services/ichor/tests/...
    make test-extract ARGS="./api/cmd/services/ichor/tests/core/contactinfosapi/..."
    go test -json ./... | ./scripts/extract_test_failures.py
"""

import argparse
import collections
import datetime
import json
import os
import pathlib
import re
import subprocess
import sys
from typing import NamedTuple


class TestKey(NamedTuple):
    package: str
    test: str  # empty string for package-level failures


def parse_events(lines) -> dict:
    """
    Parse go test -json event lines, return only failing tests.

    Returns dict of TestKey -> {'output': [str], 'elapsed': float}
    where each value had Action=fail.
    """
    groups = collections.defaultdict(lambda: {"output": [], "failed": False, "elapsed": 0.0})

    for line in lines:
        line = line if isinstance(line, str) else line.decode("utf-8", errors="replace")
        line = line.strip()
        if not line:
            continue
        try:
            ev = json.loads(line)
        except json.JSONDecodeError:
            continue

        package = ev.get("Package", "")
        test = ev.get("Test", "")
        action = ev.get("Action", "")
        key = TestKey(package=package, test=test)

        if action == "output":
            output = ev.get("Output", "")
            if output:
                groups[key]["output"].append(output)
        elif action == "fail":
            groups[key]["failed"] = True
            groups[key]["elapsed"] = ev.get("Elapsed", 0.0)

    return {k: v for k, v in groups.items() if v["failed"]}


# ── Noise patterns ────────────────────────────────────────────────────────────
_NOISE_RES = [re.compile(p) for p in [
    r"dbtest\.go:\d+:.*Name\s*:",
    r"dbtest\.go:\d+:.*HostPort\s*:",
    r"dbtest\.go:\d+:.*Create Database\s*:",
    r"dbtest\.go:\d+:.*Migrate Database\s*:",
    r"dbtest\.go:\d+:.*Seed Database\s*:",
    r"dbtest\.go:\d+:.*Drop Database\s*:",
    r"dbtest\.go:\d+:.*failed to connect",
    r"dbtest\.go:\d+:.*status check",
    r"dbtest\.go:\d+:.*dial error",
    r"dbtest\.go:\d+:.*connection refused",
    r"^=== (RUN|PAUSE|CONT)",
]]

_LOGS_START = re.compile(r"\*{10,}.*LOGS")
_LOGS_END = re.compile(r"\*{10,}.*END")
_RAW_JSON = re.compile(r"^\s*[\[{]")


def filter_noise(lines: list) -> list:
    """
    Remove noise lines from concatenated test output.
    Returns filtered list of strings (no trailing newlines).
    """
    result = []
    in_logs = False
    prev_blank = False

    for raw in lines:
        line = raw.rstrip("\n")

        # LOGS block: skip everything between markers
        if _LOGS_START.search(line):
            in_logs = True
            continue
        if in_logs:
            if _LOGS_END.search(line):
                in_logs = False
            continue

        # Pattern-based noise
        if any(p.search(line) for p in _NOISE_RES):
            continue

        # Raw JSON dumps (no file.go:N: prefix, starts with { or [)
        if _RAW_JSON.match(line):
            continue

        # Collapse consecutive blank lines
        if not line.strip():
            if not prev_blank:
                result.append("")
                prev_blank = True
            continue

        prev_blank = False
        result.append(line)

    # Strip leading/trailing blank lines
    while result and not result[0].strip():
        result.pop(0)
    while result and not result[-1].strip():
        result.pop()

    return result


_SKIP_RE = re.compile(r"^\s*(FAIL|ok |\s*--- FAIL:)")


def is_actionable(lines: list) -> bool:
    """
    Return True if the filtered output contains real failure content
    (not just FAIL/timing/--- FAIL: roll-up lines).
    """
    for line in lines:
        if not line.strip():
            continue
        if _SKIP_RE.match(line):
            continue
        return True
    return False


def sanitize_filename(package: str, test: str) -> str:
    """Derive a safe .md filename from package path and test name."""
    pkg_short = package.split("/")[-1] if package else "unknown"
    if not test:
        test_clean = "package-failure"
    else:
        # Replace subtest separator / with _, then sanitize remaining chars
        test_clean = re.sub(r"[^a-zA-Z0-9_-]", "-", test.replace("/", "_"))
    return f"{pkg_short}_{test_clean}.md"


def extract_first_error(lines: list) -> str:
    """Extract the first meaningful error line for the SUMMARY.md index."""
    signals = ["DIFF", "Error", "panic:", "Should receive", "expected",
               "status code", "cannot", "undefined", "unexpected error"]
    for line in lines:
        if any(s in line for s in signals):
            return line.strip()[:100]
    for line in lines:
        if line.strip():
            return line.strip()[:100]
    return "(see file for details)"


def write_failure_file(path: pathlib.Path, package: str, test: str,
                       elapsed: float, filtered_lines: list) -> None:
    """Write a single .md failure file."""
    title = f"Test Failure: {test}" if test else f"Package Failure: {package}"
    body = "\n".join(filtered_lines)
    path.write_text(
        f"# {title}\n\n"
        f"- **Package**: `{package}`\n"
        f"- **Duration**: {elapsed}s\n"
        f"\n## Failure Output\n\n"
        f"```\n{body}\n```\n",
        encoding="utf-8",
    )


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Extract go test -json failures to .claude/bugs/open/"
    )
    parser.add_argument("packages", nargs="*", help="Go package paths (passed to go test)")
    parser.add_argument("--output-dir", default=".claude/bugs/open",
                        help="Directory to write .md files (default: .claude/bugs/open)")
    args = parser.parse_args()

    output_dir = pathlib.Path(args.output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)

    # Clear existing .md files from a prior run
    for f in output_dir.glob("*.md"):
        f.unlink()

    if args.packages:
        cmd = ["go", "test", "-json", "-count=1"] + args.packages
        print(f"Running: CGO_ENABLED=0 {' '.join(cmd)}", flush=True)
        env = {**os.environ, "CGO_ENABLED": "0"}
        proc = subprocess.run(cmd, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, env=env)
        lines = proc.stdout.decode("utf-8", errors="replace").splitlines(keepends=True)
    else:
        # Accept piped input: go test -json ./... | python3 scripts/extract_test_failures.py
        lines = sys.stdin

    failures = parse_events(lines)

    if not failures:
        print("No test failures found. All tests passed.")
        return

    print(f"Found {len(failures)} failing test(s). Writing to {output_dir}/", flush=True)

    summary_entries = []
    written = 0
    packages_seen: set = set()

    for key, data in failures.items():
        filtered = filter_noise(data["output"])

        if not is_actionable(filtered):
            continue

        filename = sanitize_filename(key.package, key.test)
        filepath = output_dir / filename
        first_error = extract_first_error(filtered)

        write_failure_file(filepath, key.package, key.test, data["elapsed"], filtered)

        summary_entries.append(f"- {filename} — {first_error}")
        packages_seen.add(key.package)
        written += 1
        print(f"  wrote: {filename}", flush=True)

    # Write SUMMARY.md
    timestamp = datetime.datetime.utcnow().strftime("%Y-%m-%dT%H:%M:%SZ")
    summary_lines = "\n".join(summary_entries)
    (output_dir / "SUMMARY.md").write_text(
        f"# Test Failures Summary\n"
        f"Generated: {timestamp}\n"
        f"Total: {written} failures across {len(packages_seen)} packages\n\n"
        f"## Failures\n\n"
        f"{summary_lines}\n",
        encoding="utf-8",
    )

    print(f"\nSummary: {written} failures across {len(packages_seen)} packages"
          f" → {output_dir}/SUMMARY.md")


if __name__ == "__main__":
    main()
