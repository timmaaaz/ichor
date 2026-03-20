#!/bin/bash
set -euo pipefail

DB_URL="${1:?Usage: ./seed.sh <database_url> [start_from]}"
START_FROM="${2:-0}"
SEED_DIR="$(cd "$(dirname "$0")/seed" && pwd)"

# Add libpq to PATH if installed via brew (provides psql)
if [ -d "/opt/homebrew/opt/libpq/bin" ]; then
  export PATH="/opt/homebrew/opt/libpq/bin:$PATH"
fi

# Run SQL files in numeric order, optionally skipping already-applied files
for f in "$SEED_DIR"/*.sql; do
  FILE_NUM=$(basename "$f" | grep -o '^[0-9]*')
  if [ "$((10#$FILE_NUM))" -lt "$((10#$START_FROM))" ]; then
    echo "Skipping $(basename "$f") (already applied)"
    continue
  fi
  echo "Running $(basename "$f")..."
  psql -q "$DB_URL" -f "$f" --set ON_ERROR_STOP=on
done

echo "Manitowoc seed complete."
