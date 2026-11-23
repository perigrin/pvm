#!/bin/sh
# ABOUTME: Runs bison 3.8.2 on the pinned perly.y and classifies its precedence resolutions with count.pl.
# ABOUTME: Usage: count.sh <perl5 checkout>; refuses any HEAD but 94e50866 or a dirty tree. Prints count.pl's report.
set -eu
pin=94e50866
need=3.8.2
[ $# -eq 1 ] || { echo "usage: count.sh <perl5>" >&2; exit 2; }
head=$(git -C "$1" rev-parse HEAD)
case $head in
$pin*) ;;
*) echo "count.sh: $1 is at $head, not the pinned perly.y revision $pin" >&2; exit 1 ;;
esac
[ -z "$(git -C "$1" status --porcelain)" ] || { echo "count.sh: $1 has changes; the pin needs a clean tree" >&2; exit 1; }
got=$(bison --version | sed -n '1s/.* //p')
[ "$got" = "$need" ] || { echo "count.sh: bison $need is required, found $got" >&2; exit 1; }

here=$(cd "$(dirname "$0")" && pwd)
d=$(mktemp -d)
trap 'rm -rf "$d"' EXIT
cp "$1/perly.y" "$d/perly.y"
# Default reductions hide a state's lookaheads; without them every reduction lists its tokens.
bison -Dlr.default-reduction=accepting --report=itemset,lookaheads,solved \
	--report-file="$d/perly.output" -o "$d/perly.c" "$d/perly.y" 2>/dev/null
perl "$here/count.pl" "$d/perly.output"
