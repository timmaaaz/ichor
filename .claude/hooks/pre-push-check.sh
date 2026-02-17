#!/bin/bash
# Pre-push hook: blocks git push if build or tests fail.
# Wired via .claude/settings.json PreToolUse hook on Bash commands matching "git push".

INPUT=$(cat)
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // empty')

# Only gate on git push commands
if ! echo "$COMMAND" | grep -qE '^\s*git\s+push'; then
  exit 0
fi

cd "$CLAUDE_PROJECT_DIR" || exit 0

echo "Pre-push check: running go build..." >&2
if ! go build ./... 2>&1; then
  echo "BLOCKED: go build failed. Fix compile errors before pushing." >&2
  exit 2
fi

echo "Pre-push check: running go vet..." >&2
if ! go vet ./... 2>&1; then
  echo "BLOCKED: go vet failed. Fix issues before pushing." >&2
  exit 2
fi

echo "Pre-push check passed." >&2
exit 0
