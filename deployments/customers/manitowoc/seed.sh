#!/bin/bash
set -euo pipefail

DB_URL="${1:?Usage: ./seed.sh <database_url> [start_from]}"
START_FROM="${2:-0}"
SEED_DIR="$(dirname "$0")/seed"

# Step 1: Generate Manitowoc-specific SQL from YAML + config
echo "Generating Manitowoc seed data..."
python3 "$(dirname "$0")/generate.py"

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
