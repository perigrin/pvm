#!/bin/sh
# ABOUTME: Prints perly.y's grammar alternatives as bison 3.8.2 numbers them: perly.y:<yyrline> <lhs>: <rhs>.
# ABOUTME: Offline only: make test and the ratchet never run it. Usage: perly-alternatives.sh <perl5-tree>.
set -eu

NEED=3.8.2

if [ $# -ne 1 ]; then
	echo "usage: perly-alternatives.sh <perl5-tree>" >&2
	exit 2
fi
# Another bison may number alternatives or word its report differently.
got=$(bison --version | sed -n '1s/.* //p')
if [ "$got" != "$NEED" ]; then
	echo "perly-alternatives.sh: bison $NEED is required, found $got" >&2
	exit 1
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
bison -t -v -o "$tmp/perly.c" "$1/perly.y"

# yyrline is indexed from 1 ($accept) where the report counts from 0, so rule N
# of the report is yyrline[N+1], the line where its right-hand side begins.
# $accept and the mid-rule rules (@N, $@N) are left out as rules of their own.
awk '
FNR == NR {
	if (/yyrline\[\] =/) { on = 1; next }
	if (on && /^};/) on = 0
	if (on) { gsub(/[{} ]/, ""); n = split($0, a, ","); for (i = 1; i <= n; i++) if (a[i] != "") line[c++] = a[i] }
	next
}
/^Grammar/ { g = 1; next }
/^Terminals/ { g = 0 }
g && $1 ~ /^[0-9]+$/ {
	n = $1
	if ($2 == "|") { rhs = $0; sub(/^ *[0-9]+ *\| */, "", rhs) }
	else { lhs = $2; rhs = $0; sub(/^ *[0-9]+ [^ ]+ */, "", rhs) }
	if (lhs !~ /^[$@]/) print "perly.y:" line[n + 1] " " lhs " " rhs
}' "$tmp/perly.c" "$tmp/perly.output"
