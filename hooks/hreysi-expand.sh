#!/bin/sh
# hreysi: ambient expansion trigger (Claude Code SessionEnd / Stop hook).
#
# When there are commits since the last expansion, ask an agent to narrate them
# into today's buildlog entry, using the session transcript as lived experience.
# Best-effort and non-fatal: it never blocks or fails a session.
set -e

input=$(cat 2>/dev/null || true)

# Hook input is JSON on stdin (cwd, transcript_path). Parse with python3.
field() {
  printf '%s' "$input" | python3 -c "import sys,json; print(json.load(sys.stdin).get('$1',''))" 2>/dev/null || echo ""
}
cwd=$(field cwd)
[ -n "$cwd" ] && cd "$cwd" 2>/dev/null || true

root=$(git rev-parse --show-toplevel 2>/dev/null) || exit 0
head=$(git rev-parse HEAD 2>/dev/null) || exit 0

mark="$root/buildlog/.hreysi/last_expanded"
last=$(cat "$mark" 2>/dev/null || echo "")
[ "$head" = "$last" ] && exit 0   # gate: nothing new since last expansion

transcript=$(field transcript_path)

# Preferred: hand it to a headless agent that runs the `expand` skill against the
# transcript. If the CLI isn't available, drop a pending marker for next session.
if command -v claude >/dev/null 2>&1; then
  claude -p "Run the hreysi 'expand' skill for the repo at $root. There are commits since the last expansion. Use the session transcript at ${transcript:-(none)} as the lived experience. Write ## The Journey and ## Improvements into today's buildlog/ entry (never touch ## Commits), then run: git -C $root rev-parse HEAD > $mark" >/dev/null 2>&1 || true
else
  mkdir -p "$root/buildlog/.hreysi"
  printf '%s\n' "$head" > "$root/buildlog/.hreysi/pending_expansion"
fi
exit 0
