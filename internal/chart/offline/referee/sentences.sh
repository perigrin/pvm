#!/bin/sh
# ABOUTME: Generates the precedence referee's fixtures from bison's resolutions of the pinned perly.y into an output directory.
# ABOUTME: Usage: sentences.sh <perl5 checkout> <outdir>; refuses any HEAD but 94e50866 or a dirty tree. Needs bison 3.8.2 and cc.
set -eu
pin=94e50866
need=3.8.2
[ $# -eq 2 ] || { echo "usage: sentences.sh <perl5> <outdir>" >&2; exit 2; }
head=$(git -C "$1" rev-parse HEAD)
case $head in
$pin*) ;;
*) echo "sentences.sh: $1 is at $head, not the pinned perly.y revision $pin" >&2; exit 1 ;;
esac
[ -z "$(git -C "$1" status --porcelain)" ] || { echo "sentences.sh: $1 has changes; the pin needs a clean tree" >&2; exit 1; }
got=$(bison --version | sed -n '1s/.* //p')
[ "$got" = "$need" ] || { echo "sentences.sh: bison $need is required, found $got" >&2; exit 1; }

here=$(cd "$(dirname "$0")" && pwd)
d=$(mktemp -d)
trap 'rm -rf "$d"' EXIT
mkdir -p "$2"

# bison -v alone no longer reports resolutions; --report=solved adds them.
cp "$1/perly.y" "$d/perly.y"
bison -v --report=solved -o "$d/perly.c" "$d/perly.y"
grep 'Conflict between rule' "$d/perly.output" >"$2/bison-conflicts.txt"

sh "$here/build.sh" "$1" "$d"
cc -O1 -DYYDEBUG=1 -o "$d/referee-prec-trace" "$d/referee-prec.c"
for v in prec noprec; do
	bison -Wno-conflicts-sr -Wno-conflicts-rr -Wno-other --report=itemset,solved \
		--report-file="$d/$v.output" -o "$d/report-$v.c" "$d/referee-$v.y"
done
perl "$here/sentences.pl" "$d" "$2"
