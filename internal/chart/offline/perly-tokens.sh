#!/bin/sh
# ABOUTME: Prints the terminals perly.y declares, one per line, in first-declaration order, unfiltered.
# ABOUTME: Offline only: make test never runs it. Usage: perly-tokens.sh <perl5-tree>; refuses any HEAD but 94e50866.
set -eu
pin=94e50866
[ $# -eq 1 ] || { echo "usage: perly-tokens.sh <perl5-tree>" >&2; exit 2; }
head=$(git -C "$1" rev-parse HEAD)
case $head in
$pin*) ;;
*) echo "perly-tokens.sh: $1 is at $head, not the pinned perly.y revision $pin" >&2; exit 1 ;;
esac
[ -z "$(git -C "$1" status --porcelain)" ] || { echo "perly-tokens.sh: $1 has uncommitted changes; the pinned tree must be clean" >&2; exit 1; }
# The declaration section ends at the first %%; <type> tags are dropped.
awk '
/^%%/ { exit }
/^%(token|left|right|nonassoc)([ \t]|$)/ {
	sub(/^%[a-z]+/, ""); gsub(/<[A-Za-z_0-9]+>/, " ")
	for (i = 1; i <= NF; i++) if (!seen[$i]++) print $i
}' "$1/perly.y"
