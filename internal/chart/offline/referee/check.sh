#!/bin/sh
# ABOUTME: Builds the bison referees into a temporary directory and checks every committed case's verdict.
# ABOUTME: Usage: check.sh <perl5 checkout at 94e50866>; offline only, never run by make test or the ratchet.
set -eu
here=$(cd "$(dirname "$0")" && pwd)
d=$(mktemp -d)
trap 'rm -rf "$d"' EXIT
sh "$here/build.sh" "$1" "$d"

n=0 bad=0
while read -r name mode prec noprec tokens; do
	case $name in '' | '#'*) continue ;; esac
	flag=
	[ "$mode" = prefix ] && flag=-p
	for v in prec noprec; do
		eval want=\$$v
		got=$(printf '%s\n' $tokens | "$d/referee-$v" $flag)
		[ "$got" = "$want" ] || { echo "check.sh: $name: referee-$v gave $got, want $want" >&2; bad=1; }
	done
	n=$((n + 1))
done <"$here/testdata/cases.txt"
[ "$n" -gt 0 ] || { echo "check.sh: no cases" >&2; exit 1; }
[ "$bad" -eq 0 ] || exit 1
echo "check.sh: $n cases agree"
