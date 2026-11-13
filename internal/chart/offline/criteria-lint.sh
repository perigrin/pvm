#!/bin/sh
# ABOUTME: Lints a chain's acceptance criteria: a line with no command, a non-guard command already green, a pin on a file a later issue edits.
# ABOUTME: Reads `git zhi issue list --format json` on stdin, needs jq, runs each non-guard command in the current directory; needs no perl.
#
# usage: git zhi issue list --all --format json |
#            jq '[.[] | select(.milestone=="<m>")]' |
#            sh internal/chart/offline/criteria-lint.sh
#
# A criterion is a checkbox line under `## Acceptance Criteria` (both its
# Positive and Negative Scenarios); its command is the last paren-wrapped
# backtick span on the line. Issues in state done or cancelled are skipped
# (their criteria are green by construction) but still count as later issues.
# Prints one line per defect, `<id prefix> "<title>": <defect>: <criterion>`,
# and exits 1 if there are any, 0 if none.
set -eu

die() { echo "criteria-lint.sh: $*" >&2; exit 2; }
command -v jq >/dev/null || die "jq is not on PATH"

in=$(mktemp)
script=$(mktemp)
trap 'rm -f "$in" "$script"' EXIT
cat >"$in"
jq -e 'type == "array"' "$in" >/dev/null 2>&1 ||
    die "stdin is not a JSON array of issues (git zhi issue list --format json)"

# shortcut: a pin is recognised by its command's verb (a hash, cmp/diff, a
# commit log or count), not by what it asserts; widen the pattern when a pin
# in another shape gets past it.
jq -r '
  def criteria: reduce ((.body // "") | split("\n"))[] as $l ({in: false, out: []};
      if ($l | test("^## ")) then .in = ($l | test("^## Acceptance Criteria\\s*$"))
      elif .in and ($l | test("^\\s*- \\[[ xX]\\] ")) then .out += [$l | sub("^\\s*- \\[[ xX]\\] "; "")]
      else . end) | .out;
  def command: [match("\\(`([^`]*)`\\)"; "g")] | if length == 0 then null else last.captures[0].string end;
  def paths: [(.body // "") | split("\n")[] | select(test("^\\s*- paths:")) | match("`([^`]+)`"; "g").captures[0].string];
  def name: "\(.id[0:13]) \"\(.title)\"";
  (map({key: .id, value: (.blocks // [])}) | from_entries) as $blocks
  | (map({key: .id, value: paths}) | from_entries) as $paths
  | def later($id): ($blocks[$id] // []) as $d | ($d + ([$d[] | later(.)[]])) | unique;
  .[] | select(.state != "done" and .state != "cancelled") | . as $i
  | criteria[] as $c | ($c | command) as $cmd
  | if $cmd == null then "say \("\($i | name): no command: \($c)" | @sh)"
    else
      (if ($c | startswith("guard:")) then empty
       else "check \("\($i | name): passes on the current tree: \($c)" | @sh) \($cmd | @sh)" end),
      (select($cmd | test("sha[0-9]*sum|md5sum|cksum|hash-object|\\bcmp\\b|\\bdiff\\b|git log|rev-list"))
       | later($i.id)[] as $l | ($paths[$l] // [])[] as $p | select($cmd | contains($p))
       | "say \("\($i | name): pins \($p), which \($l[0:13]) edits: \($c)" | @sh)")
    end
' "$in" >"$script"

found=0
say() { echo "$1"; found=1; }
check() { if (sh -c "$2") </dev/null >/dev/null 2>&1; then say "$1"; fi; }
. "$script"
exit "$found"
