# Python Test Failure Extractor Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the broken shell+jq+awk test failure extractor with a correct Python 3 implementation that properly parses `go test -json` output and writes clean per-test `.md` files to `.claude/bugs/open/`.

**Architecture:** Single self-contained Python 3 script. Reads `go test -json` output by either running the test command directly or reading from stdin. Groups JSON events by `(Package, Test)` key using `collections.defaultdict`, applies a noise filter list, and writes one `.md` file per failing test with actionable content only.

**Tech Stack:** Python 3 stdlib only (`json`, `re`, `subprocess`, `pathlib`, `collections`, `argparse`). No pip dependencies. Invoked via `python3 scripts/extract_test_failures.py`.

---

## Chunk 1: Core parsing and noise filter

### Task 1: Write the core event parser

**Files:**
- Create: `scripts/extract_test_failures.py`

The `go test -json` format emits one JSON object per line. Each has `Action`, `Package`, `Test` (absent for package-level), `Output` (for `action=output`), and `Elapsed` (for `action=fail/pass`).

- [ ] **Step 1: Create the script file with data types and parse_events()**

```python
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
```

- [ ] **Step 2: Verify it parses a synthetic JSON stream correctly**

In a Python REPL or quick test, run:
```python
lines = [
    '{"Action":"output","Package":"github.com/foo/bar","Test":"Test_Foo","Output":"    t.go:1: DIFF\\n"}',
    '{"Action":"fail","Package":"github.com/foo/bar","Test":"Test_Foo","Elapsed":0.03}',
    '{"Action":"fail","Package":"github.com/foo/bar","Test":"","Elapsed":0.03}',
]
result = parse_events(lines)
assert TestKey("github.com/foo/bar", "Test_Foo") in result
assert result[TestKey("github.com/foo/bar", "Test_Foo")]["output"] == ["    t.go:1: DIFF\n"]
```

Expected: assertions pass silently.

---

### Task 2: Write the noise filter

**Files:**
- Modify: `scripts/extract_test_failures.py` (append after parse_events)

The noise filter must handle:
1. `dbtest.go` setup lines (Name:, HostPort:, Create/Migrate/Seed/Drop Database:)
2. DB infrastructure errors (connection refused, dial error, failed to connect, status check) — environment failures, not code bugs
3. `=== RUN / PAUSE / CONT` lines
4. LOGS blocks: everything between `***** LOGS` and `***** END` markers (10+ asterisks)
5. Raw JSON dumps: lines that start with `{` or `[` without a `file.go:N:` prefix
6. Collapse consecutive blank lines

- [ ] **Step 3: Append filter_noise() and is_actionable()**

```python
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
```

- [ ] **Step 4: Verify filter_noise() on representative inputs**

```python
# DB infrastructure error → filtered out (leaving nothing actionable)
lines = [
    "    dbtest.go:522: failed to connect to user=postgres: connection refused",
    "--- FAIL: Test_Country (3.54s)",
]
filtered = filter_noise(lines)
assert not is_actionable(filtered), f"Expected not actionable, got: {filtered}"

# Real DIFF output → preserved
lines = [
    "    apitest.go:73: DIFF",
    "    GOT  : 200",
    "    EXP  : 201",
    "--- FAIL: Test_Foo/create-200 (0.03s)",
]
filtered = filter_noise(lines)
assert is_actionable(filtered)
assert any("DIFF" in l for l in filtered)

# Single-line JSON dump → filtered (no in_json state bug)
lines = [
    '{"id":"abc123","name":"test"}',
    "    apitest.go:73: DIFF",
]
filtered = filter_noise(lines)
assert any("DIFF" in l for l in filtered), f"DIFF should survive: {filtered}"
```

Expected: all assertions pass.

---

## Chunk 2: File output and main()

### Task 3: Write output functions and main()

**Files:**
- Modify: `scripts/extract_test_failures.py` (append)

- [ ] **Step 5: Append sanitize_filename(), extract_first_error(), write_failure_file()**

```python
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
```

- [ ] **Step 6: Append main()**

```python
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
```

- [ ] **Step 7: Make the script executable**

```bash
chmod +x scripts/extract_test_failures.py
```

---

### Task 4: Update Makefile and remove old shell script

**Files:**
- Modify: `Makefile` — update `test-extract` target
- Delete: `scripts/extract-test-failures.sh`

- [ ] **Step 8: Update the Makefile target**

Find:
```makefile
test-extract:
	@mkdir -p .claude/bugs/open
	@rm -f .claude/bugs/open/*.md
	@./scripts/extract-test-failures.sh $(ARGS)
```

Replace with:
```makefile
test-extract:
	@python3 scripts/extract_test_failures.py $(ARGS)
```

Note: the Python script handles `mkdir -p` and `rm -f` internally, so the Makefile target becomes a one-liner.

- [ ] **Step 9: Delete the old shell script**

```bash
rm scripts/extract-test-failures.sh
```

- [ ] **Step 10: Verify the script runs without error on a simple package**

```bash
make test-extract ARGS="./foundation/..."
```

Expected: either "No test failures found" or a list of `.md` files written to `.claude/bugs/open/`. No Python tracebacks.

- [ ] **Step 11: Verify piped mode works**

```bash
go test -json -count=1 ./foundation/... | python3 scripts/extract_test_failures.py
```

Expected: same result as Step 10.

- [ ] **Step 12: Commit**

```bash
git add scripts/extract_test_failures.py Makefile
git rm scripts/extract-test-failures.sh
git commit -m "refactor(scripts): replace shell extractor with Python — correct json.loads() parsing"
```

---

## Reference: Known issues this plan fixes

| Shell bug | Python fix |
|-----------|-----------|
| `in_json=1` never resets on single-line `{...}` | No state machine — `_RAW_JSON.match(line)` is stateless |
| `^--- FAIL:` anchor misses indented lines | `_SKIP_RE` uses `^\s*` to match indented lines |
| `\*{10,}` LOGS pattern may not match | Same pattern, but now a compiled regex tested in isolation |
| `printf -- '-'` format flag on macOS | Not applicable — Python f-strings |
| `jq --slurp` loads entire stream into memory | `parse_events()` streams line by line |
| DB infrastructure errors produce empty files | `is_actionable()` skips them cleanly |
