#!/usr/bin/env bash
# scripts/check-determinism.sh
#
# Verifies that `make reseed-frontend` produces byte-identical seed data
# across runs for the tables that P1 (productbus.SeedCreate) stabilises.
# Excludes columns that drift via wall-clock time (created_date /
# updated_date), since those use time.Now() inside the seed code path.
#
# Tables checked (post-P1):
#   - products.products                (id, sku, brand_id, ...)
#   - inventory.label_catalog          (id, code, type, entity_ref, payload_json)
#
# Tables explicitly NOT checked (still drift via uuid.New() in their own
# seed funcs — out of scope for P1, candidates for future work):
#   - products.cost_history, inventory.inventory_items,
#     inventory.serial_numbers, sales.order_line_items, ...
#
# WHEN TO ADD A TABLE TO THE SNAPSHOTS LIST:
#   You should add an entry to the SNAPSHOTS array below whenever any of
#   these signals appear in a PR or its review:
#
#     1. A new SeedCreate (or equivalent) method is added to a *bus
#        package, mirroring labelbus.SeedCreate / productbus.SeedCreate.
#     2. A TestSeed* function is migrated from `uuid.New()` to deriving
#        its primary key from `seedid.Stable("<entity>:<key>")`. The
#        primary key column is now byte-stable across reseeds and
#        should be diff-checked here.
#     3. A FK column from a downstream table to a stabilized primary
#        key is itself derived from the parent's stable key (rather
#        than queried-then-cached at seed time). When this happens,
#        the downstream table can also be added.
#
#   How to add: append a pipe-delimited entry following the existing
#   format: <label>|<schema.table>|<comma-separated stable cols>|<order-by>
#   Always project columns explicitly — never use `SELECT *` because
#   `created_date` and any future timestamp/sequence columns will drift
#   via wall-clock time and produce false-positive failures.
#
#   How to remove: if a table is intentionally non-deterministic and
#   the comment block above lists it as a candidate, but the team has
#   decided not to stabilize it, delete the corresponding entry below
#   AND update this header comment to remove it from the candidate list.
#
# Usage:
#   ./scripts/check-determinism.sh                  # uses default DSN
#   ./scripts/check-determinism.sh "postgresql://user:pass@host:port/db"
#
# Exit codes:
#   0  all checked tables are byte-identical across reseeds
#   1  at least one table drifted
#   2  prerequisites missing (psql, make, expected env)
#   3  make reseed-frontend failed (build, migrations, or seed error)

set -euo pipefail

DSN="${1:-postgresql://postgres:postgres@localhost:5432/postgres}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
WORK_DIR="$(mktemp -d -t ichor-determinism.XXXXXX)"
trap 'rm -rf "$WORK_DIR"' EXIT

require_tool () {
	command -v "$1" >/dev/null 2>&1 || {
		echo "ERROR: required tool '$1' not on PATH" >&2
		exit 2
	}
}
require_tool psql
require_tool make
require_tool diff

# Each entry: <label>|<schema.table>|<projected stable columns>|<order-by>
SNAPSHOTS=(
	"products|products.products|id, sku, brand_id, category_id, name, description, model_number, upc_code, status, is_active, is_perishable, handling_instructions, units_per_case, tracking_type, inventory_type|sku"
	"label_catalog|inventory.label_catalog|id, code, type, entity_ref, payload_json|code"
)

dump_snapshot () {
	local run_label="$1"
	local out_dir="$WORK_DIR/$run_label"
	mkdir -p "$out_dir"
	for entry in "${SNAPSHOTS[@]}"; do
		IFS='|' read -r label table cols order_by <<<"$entry"
		local out_file="$out_dir/${label}.tsv"
		# \COPY runs client-side and writes to STDOUT; --no-psqlrc avoids
		# user-specific format settings polluting the output.
		psql "$DSN" --no-psqlrc -At <<-SQL > "$out_file"
		\COPY (SELECT $cols FROM $table ORDER BY $order_by) TO STDOUT WITH (NULL '\N')
		SQL
		echo "  captured: $label ($(wc -l < "$out_file") rows)"
	done
}

cd "$REPO_ROOT"

echo "==> Run 1: make reseed-frontend"
if ! make reseed-frontend; then
	echo "ERROR: 'make reseed-frontend' (Run 1) failed. Check build, migrations, or seed logic." >&2
	exit 3
fi
echo "==> Snapshot 1"
dump_snapshot run1

echo "==> Run 2: make reseed-frontend"
if ! make reseed-frontend; then
	echo "ERROR: 'make reseed-frontend' (Run 2) failed. Check build, migrations, or seed logic." >&2
	exit 3
fi
echo "==> Snapshot 2"
dump_snapshot run2

echo
echo "==> Diffing snapshots"
DRIFT=0
for entry in "${SNAPSHOTS[@]}"; do
	IFS='|' read -r label _ _ _ <<<"$entry"
	if diff -u "$WORK_DIR/run1/${label}.tsv" "$WORK_DIR/run2/${label}.tsv" > "$WORK_DIR/${label}.diff"; then
		echo "  STABLE: $label"
	else
		DRIFT=1
		echo "  DRIFT:  $label"
		echo "    --- diff (first 40 lines) ---"
		head -40 "$WORK_DIR/${label}.diff" | sed 's/^/    /'
		echo "    --- end diff ---"
		# Preserve full diff for later inspection.
		cp "$WORK_DIR/${label}.diff" "$REPO_ROOT/check-determinism-${label}.diff"
		echo "    full diff saved: check-determinism-${label}.diff"
	fi
done

if [ $DRIFT -ne 0 ]; then
	echo
	echo "FAIL: at least one table drifted between reseeds. See diff files in repo root." >&2
	exit 1
fi
echo
echo "OK: all checked tables are byte-identical across reseeds"
