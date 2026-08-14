#!/usr/bin/env sh
set -eu

binary=${1:-./bin/reclaimit}
if [ ! -x "$binary" ]; then
  printf 'reclaimit binary is not executable: %s\n' "$binary" >&2
  printf 'Build it first with: go build -o ./bin/reclaimit ./cmd/reclaimit\n' >&2
  exit 1
fi

demo_root=$(mktemp -d "${TMPDIR:-/tmp}/reclaimit-demo.XXXXXX")
cleanup() {
  rm -rf -- "$demo_root"
}
trap cleanup EXIT HUP INT TERM

mkdir -p \
  "$demo_root/web-app/.git" \
  "$demo_root/web-app/node_modules/pkg" \
  "$demo_root/python-api/.git" \
  "$demo_root/python-api/.venv/lib"
dd if=/dev/zero of="$demo_root/web-app/node_modules/pkg/index.js" bs=1048576 count=12 2>/dev/null
dd if=/dev/zero of="$demo_root/python-api/.venv/lib/site.py" bs=1048576 count=8 2>/dev/null

printf '\n1/3 Analyze the synthetic workspace\n\n'
"$binary" analyze --root "$demo_root"

if [ -t 0 ] && [ -t 1 ]; then
  printf '\n2/3 Review candidates in the TUI; press q to continue\n\n'
  "$binary" tui --root "$demo_root"
else
  printf '\n2/3 TUI skipped because this terminal is not interactive\n'
fi

printf '\n3/3 Preview node_modules cleanup without deleting it\n\n'
"$binary" clean \
  --root "$demo_root" \
  --include-category node-modules \
  --dry-run
