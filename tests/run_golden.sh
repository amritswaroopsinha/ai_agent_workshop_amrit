#!/usr/bin/env bash
# Golden tests: diff mytools against real bedtools.
# Usage: ./tests/run_golden.sh
set -uo pipefail

MYTOOLS=${MYTOOLS:-mytools}     # override to test a different build
DATA=$(dirname "$0")/../data
tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
pass=0; fail=0

bedtools sort -i "$DATA/a.bed" > "$tmp/sorted_a.bed"

# check <name> -- <args...>
#   runs "$MYTOOLS <args>" and "bedtools <args>", diffs them
check() {
  local name=$1; shift; shift        # drop the literal --
  "$MYTOOLS" "$@" > "$tmp/got"  2>"$tmp/got.err"
  local got_rc=$?
  bedtools   "$@" > "$tmp/want" 2>/dev/null
  local want_rc=$?

  if [[ $got_rc -ne $want_rc ]]; then
    echo "FAIL $name (exit $got_rc, bedtools gave $want_rc)"
    sed 's/^/      /' "$tmp/got.err" | head -3
    (( fail++ )); return
  fi
  if diff -q "$tmp/want" "$tmp/got" >/dev/null; then
    echo "ok   $name"; (( pass++ ))
  else
    echo "FAIL $name"
    diff -u "$tmp/want" "$tmp/got" | sed 's/^/      /' | head -20
    (( fail++ ))
  fi
}

# check_stdin <name> <stdin-file> -- <args...>
#   same as check(), but feeds stdin-file to both tools' stdin
check_stdin() {
  local name=$1 input=$2; shift; shift; shift
  "$MYTOOLS" "$@" < "$input" > "$tmp/got"  2>"$tmp/got.err"
  local got_rc=$?
  bedtools   "$@" < "$input" > "$tmp/want" 2>/dev/null
  local want_rc=$?

  if [[ $got_rc -ne $want_rc ]]; then
    echo "FAIL $name (exit $got_rc, bedtools gave $want_rc)"
    sed 's/^/      /' "$tmp/got.err" | head -3
    (( fail++ )); return
  fi
  if diff -q "$tmp/want" "$tmp/got" >/dev/null; then
    echo "ok   $name"; (( pass++ ))
  else
    echo "FAIL $name"
    diff -u "$tmp/want" "$tmp/got" | sed 's/^/      /' | head -20
    (( fail++ ))
  fi
}

check "sort a.bed"                    -- sort -i "$DATA/a.bed"
check_stdin "sort stdin" "$DATA/a.bed" -- sort -i -

check "merge (sorted a.bed)"                    -- merge -i "$tmp/sorted_a.bed"
check "merge -d 10 (sorted a.bed)"              -- merge -d 10 -i "$tmp/sorted_a.bed"
check_stdin "merge stdin" "$tmp/sorted_a.bed"   -- merge -i -

check "intersect default"                          -- intersect -a "$DATA/a.bed" -b "$DATA/b.bed"
check "intersect -u"                               -- intersect -a "$DATA/a.bed" -b "$DATA/b.bed" -u
check "intersect -v"                               -- intersect -a "$DATA/a.bed" -b "$DATA/b.bed" -v
check "intersect -wa"                              -- intersect -a "$DATA/a.bed" -b "$DATA/b.bed" -wa
check_stdin "intersect -a stdin" "$DATA/a.bed"     -- intersect -a - -b "$DATA/b.bed"

check "subtract"                               -- subtract -a "$DATA/a.bed" -b "$DATA/b.bed"
check_stdin "subtract -a stdin" "$DATA/a.bed"  -- subtract -a - -b "$DATA/b.bed"

echo "---"
echo "$pass passed, $fail failed"
[[ $fail -eq 0 ]]
