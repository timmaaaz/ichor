#!/bin/bash
set -euo pipefail

DB_URL="${1:?Usage: ./seed.sh <database_url> [start_from]}"
START_FROM="${2:-0}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SEED_DIR="$SCRIPT_DIR/seed"

# Add libpq to PATH if installed via brew (provides psql)
if [ -d "/opt/homebrew/opt/libpq/bin" ]; then
  export PATH="/opt/homebrew/opt/libpq/bin:$PATH"
fi

# Use venv Python if available, otherwise system Python
PYTHON="${SCRIPT_DIR}/.venv/bin/python3"
if [ ! -x "$PYTHON" ]; then
  PYTHON="python3"
fi

# Step 1: Generate Manitowoc-specific SQL from YAML + config
echo "Generating Manitowoc seed data..."
"$PYTHON" "$SCRIPT_DIR/generate.py"

# Step 2: Run SQL files in numeric order, optionally skipping already-applied files
for f in "$SEED_DIR"/*.sql; do
  FILE_NUM=$(basename "$f" | grep -o '^[0-9]*')
  if [ "$((10#$FILE_NUM))" -lt "$((10#$START_FROM))" ]; then
    echo "Skipping $(basename "$f") (already applied)"
    continue
  fi
  echo "Running $(basename "$f")..."
  psql "$DB_URL" -f "$f" --set ON_ERROR_STOP=on
done

echo "Manitowoc seed complete."
