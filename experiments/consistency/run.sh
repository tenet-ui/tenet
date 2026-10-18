#!/usr/bin/env bash
# Consistency experiment runner (slice 6). Prompts and conditions: briefs.md and the spec.
set -euo pipefail

src=${BASH_SOURCE[0]}
case $src in */*) src_dir=${src%/*} ;; *) src_dir=. ;; esac
here=$(cd "$src_dir" && pwd)
repo=$(cd "$here/../.." && pwd)
briefs="$here/briefs.md"
out_dir=${OUT_DIR:-$here/outputs}
MODEL=claude-sonnet-5-5
# --bare is not used: it skips keychain reads, so a subscription login fails ("Not logged in"). See REPORT.md limits.
BARE_FLAGS=()

# section <heading>: prints the body of "## <heading>" up to the next "## " heading.
section() {
  awk -v h="## $1" '$0 == h {on = 1; next} /^## / {on = 0} on' "$briefs" | sed '/./,$!d'
}

brief_ids() {
  sed -n 's/^## brief: //p' "$briefs"
}

# build_prompt <plain|tools> <brief-id>
build_prompt() {
  section "brief: $2"
  echo
  if [ "$1" = tools ]; then
    section addendum-tools
    echo
  fi
  section contract
}

if [ "${1:-}" = --print-prompt ]; then
  build_prompt "$2" "$3"
  exit 0
fi

dry=0
force=0
only=""
while [ $# -gt 0 ]; do
  case $1 in
    --dry-run) dry=1 ;;
    --force) force=1 ;;
    --only) only=$2; shift ;;
    *) echo "unknown argument: $1" >&2; exit 2 ;;
  esac
  shift
done

selected() { [ -z "$only" ] || [ "$only" = "$1" ]; }

if [ "$dry" = 0 ] && ! command -v claude >/dev/null 2>&1; then
  echo "claude CLI not found on PATH" >&2
  exit 1
fi

if [ -n "$only" ]; then
  ids=$(brief_ids)
  if ! grep -qxF -- "$only" <<< "$ids"; then
    echo "unknown brief id: $only" >&2
    exit 2
  fi
fi

if [ "$dry" = 1 ]; then
  for cond in plain tools; do
    for id in $(brief_ids); do
      if selected "$id"; then
        echo "would run: $cond $id"
      fi
    done
  done
  exit 0
fi

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
(cd "$repo" && go build -o "$work/tenet" ./cmd/tenet)
cat > "$work/mcp.json" <<EOF
{"mcpServers":{"tenet":{"command":"$work/tenet","args":["--dir","$repo/packs/stripe-reference"]}}}
EOF
echo '{"mcpServers":{}}' > "$work/none.json"
mkdir -p "$out_dir" && out_dir=$(cd "$out_dir" && pwd)
cd "$work"

# run_one <plain|tools> <brief-id>
run_one() {
  local cond=$1 id=$2 dir="$out_dir/$1" attempt err out prompt ok=0
  local cfg=()
  mkdir -p "$dir" || return 1
  if [ "$force" = 0 ] && [ -f "$dir/$id.html" ]; then
    echo "skip $cond $id (already has output)" >&2
    return 0
  fi
  if [ "$cond" = tools ]; then
    cfg=(--mcp-config "$work/mcp.json" --allowedTools "mcp__tenet__*")
  else
    cfg=(--mcp-config "$work/none.json")
  fi
  prompt=$(build_prompt "$cond" "$id")
  err=""
  for attempt in 1 2; do
    err=""
    rm -f "$work/candidate.html"
    if out=$(printf '%s' "$prompt" | claude -p "${BARE_FLAGS[@]}" --strict-mcp-config "${cfg[@]}" --model "$MODEL" --no-session-persistence --max-budget-usd 2 --tools="" 2>"$work/err.txt"); then
      printf '%s\n' "$out" | sed '/^```/d' > "$work/candidate.html"
      if grep -qi '<html' "$work/candidate.html"; then
        ok=1
        break
      fi
      err="reply had no <html> element"
    else
      err=$(cat "$work/err.txt")
      [ -n "$err" ] || err="claude exited non-zero"
    fi
  done
  if [ "$ok" = 1 ]; then
    mv "$work/candidate.html" "$dir/$id.html" || return 1
    rm -f "$dir/$id.failed"
    echo "ok $cond $id" >&2
  else
    rm -f "$dir/$id.html"
    printf '%s\n' "$err" > "$dir/$id.failed" || return 1
    echo "FAILED $cond $id: $err" >&2
    return 1
  fi
}

failures=0
for cond in plain tools; do
  for id in $(brief_ids); do
    if selected "$id"; then
      run_one "$cond" "$id" || failures=$((failures + 1))
    fi
  done
done
echo "done: $failures failed" >&2
[ "$failures" -eq 0 ]
