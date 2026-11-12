#!/bin/sh
# ABOUTME: Writes the pinned -DDEBUGGING perl's oracle facts over the oracle-graded corpora (spec §9.1) to <outdir>.
# ABOUTME: Refuses a perl that is not the pinned debug build, and a perl5 tree or PerlOnJava not at its pinned commit and clean.
#
# usage: T1_CORPUS=<PerlOnJava>/src/test/resources/unit \
#        sh internal/chart/offline/oracle-facts.sh <perl> <perl5-tree> <outdir>
#
# <outdir> receives in-tree.txt, t1.txt and perlgit.txt; commit them as
# internal/chart/testdata/oracle/. Run from the repository root.
set -eu

[ $# -eq 3 ] || { echo "usage: oracle-facts.sh <perl> <perl5-tree> <outdir>" >&2; exit 2; }
perl=$1 tree=$2 outdir=$3
die() { echo "oracle-facts.sh: $*" >&2; exit 1; }

[ -x "$perl" ] || die "perl $perl is missing or not executable; refusing to read the tree $tree"
commit=$("$perl" -MConfig -e 'print $Config{git_commit_id}') || die "$perl does not run"
case $commit in 94e50866*) ;; *) die "$perl is at commit '$commit', not 94e50866" ;; esac
case " $("$perl" -MConfig -e 'print $Config{ccflags}') " in *" -DDEBUGGING "*) ;; *) die "$perl is not a -DDEBUGGING build" ;; esac

head=$(git -C "$tree" rev-parse HEAD 2>/dev/null) || die "$tree is not a git tree"
case $head in 94e50866*) ;; *) die "$tree is at $head, not 94e50866" ;; esac
[ -z "$(git -C "$tree" status --porcelain)" ] || die "$tree has uncommitted changes; the pinned tree must be clean"

t1=${T1_CORPUS:?set T1_CORPUS to PerlOnJava 73be162b src/test/resources/unit}
t1commit=$(git -C "$t1" rev-parse --short=8 HEAD) || die "T1_CORPUS=$t1 is not a git tree"
[ "$t1commit" = 73be162b ] || die "T1_CORPUS=$t1 is at PerlOnJava $t1commit, not 73be162b"
[ -z "$(git -C "$t1" status --porcelain)" ] || die "T1_CORPUS=$t1 has uncommitted changes; the pinned corpus must be clean"

mkdir -p "$outdir"
cd "$(git rev-parse --show-toplevel)"
go run ./internal/chart/offline/oracle "$perl" "$tree" "$outdir"
