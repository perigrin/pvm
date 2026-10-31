#!/bin/sh
# ABOUTME: Builds referee-prec and referee-noprec, bison parsers of the pinned perly.y, into an output directory.
# ABOUTME: Usage: build.sh <perl5 checkout> <outdir>; refuses any perl5 HEAD but 94e50866. Needs bison and cc.
set -eu
pin=94e50866
[ $# -eq 2 ] || { echo "usage: build.sh <perl5> <outdir>" >&2; exit 2; }
head=$(git -C "$1" rev-parse HEAD)
case $head in
$pin*) ;;
*) echo "build.sh: $1 is at $head, not the pinned perly.y revision $pin" >&2; exit 1 ;;
esac
here=$(cd "$(dirname "$0")" && pwd)
mkdir -p "$2"
for v in prec noprec; do
	perl "$here/strip.pl" $v <"$1/perly.y" >"$2/referee-$v.y"
	bison -Wno-conflicts-sr -Wno-conflicts-rr -Wno-other -o "$2/referee-$v.c" "$2/referee-$v.y"
	cc -O1 -o "$2/referee-$v" "$2/referee-$v.c"
done
