#!/bin/sh
# ABOUTME: Builds each toy grammar here with bison and writes its fixture, <toy>.txt, into an output directory.
# ABOUTME: Usage: build.sh <outdir>; needs bison and cc. The fixtures are internal/chart/testdata/precedence/toys/.
set -eu
[ $# -eq 1 ] || { echo "usage: build.sh <outdir>" >&2; exit 2; }
here=$(cd "$(dirname "$0")" && pwd)
mkdir -p "$1"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
for y in "$here"/*.y; do
	n=$(basename "$y" .y)
	perl "$here/toys.pl" bison <"$y" >"$tmp/$n.y"
	# An unresolved conflict fails the build: every conflict a toy has is resolved by precedence.
	bison -Werror=conflicts-sr -Werror=conflicts-rr -o "$tmp/$n.c" "$tmp/$n.y"
	cc -O1 -o "$tmp/$n" "$tmp/$n.c"
	{
		echo "# $n: written by build.sh from toys/$n.y; do not edit."
		perl "$here/toys.pl" header <"$y"
		perl "$here/toys.pl" sentences <"$y" | "$tmp/$n"
	} >"$1/$n.txt"
done
