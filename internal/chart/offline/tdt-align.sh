#!/bin/sh
# ABOUTME: Captures every waypoint-1 corpus file's -DT stream with tdt-capture.sh, then aligns them to internal/lexer tokens.
# ABOUTME: Writes $TDT_FIXTURES/<file>.tdt and testdata/tdt/sample.tdt; offline. Usage: TDT_FIXTURES=<dir> T1_CORPUS=<dir> tdt-align.sh <perl>.
set -eu

perl=${1:?usage: TDT_FIXTURES=<dir> T1_CORPUS=<PerlOnJava unit dir> tdt-align.sh <debugging-perl>}
t1=${T1_CORPUS:?set T1_CORPUS to PerlOnJava 73be162b src/test/resources/unit}
# The corpus fixtures are too large to commit (perigrin: "Don't commit fixtures"); they live outside the tree.
out=${TDT_FIXTURES:?set TDT_FIXTURES to the directory the corpus fixtures go to}
top=$(git rev-parse --show-toplevel)
# Absolute before any cd, and outside the repository: the aligner clears the
# *.tdt files under it (the committed sample.tdt among them, inside the tree).
# It is resolved through its nearest existing parent and created only once accepted.
refuse() {
	echo "tdt-align.sh: TDT_FIXTURES=$out is inside the repository or holds it; name a directory outside it" >&2
	exit 2
}
case $out in /*) ;; *) out=$PWD/$out ;; esac
base=$out rest=
while [ ! -d "$base" ]; do
	base=${base%/}
	rest=/${base##*/}$rest
	base=$(dirname "$base")
done
# A . or .. below the existing parent would resolve only once created.
case "$rest/" in */./* | */../*) refuse ;; esac
out=$(cd "$base" && pwd -P)
out=${out%/}$rest
out=${out:-/}
topreal=$(cd "$top" && pwd -P)
case "$topreal/" in "${out%/}/"*) refuse ;; esac
case "${out%/}/" in "$topreal/"*) refuse ;; esac
mkdir -p "$out"
here=$top/internal/chart/offline
cap=$(mktemp -d)
trap 'rm -rf "$cap"' EXIT

# capture <cwd> <PERL5LIB> <path from cwd> <key>: one stream, run as label.sh runs perl.
# Exit 1 (refused, timeout) and 2 with a `# truncated` line are streams; any other exit 2 stops the run.
capture() {
	mkdir -p "$cap/$(dirname "$4")"
	s=0
	(cd "$1" && PERL5LIB=$2 sh "$here/tdt-capture.sh" "$perl" "$3") >"$cap/$4" || s=$?
	if [ "$s" -eq 2 ] && ! tail -n 1 "$cap/$4" | grep -q '^# truncated: '; then
		echo "tdt-align.sh: $4: capture failed" >&2
		exit 2
	fi
}

start=$(date +%s)
cd "$top"
sample=internal/chart/offline/testdata/tdt-sample.pl
capture "$top" "" "$sample" "$sample"
for f in internal/chart/testdata/guacamole/parses/*.pl internal/chart/testdata/guacamole/parsent/*.pl \
	internal/chart/testdata/std5k.pl internal/psc/testdata/latency/edit5k.pl; do
	capture "$top" internal/chart/testdata/guacamole/stubs "$f" "$f"
done
[ -z "$(git -C "$t1" status --porcelain -- .)" ] || { echo "tdt-align.sh: T1_CORPUS=$t1 has uncommitted changes; the pinned corpus must be clean" >&2; exit 1; }
t1top=$(git -C "$t1" rev-parse --show-toplevel)
t1root=$(git -C "$t1" rev-parse --show-prefix)
for f in $(cd "$t1" && ls -- *.t); do
	capture "$t1top" "" "${t1root%/}/$f" "T1/$f"
done
echo "tdt-align.sh: captured in $(($(date +%s) - start)) s"
go run ./internal/chart/offline/tdt/align "$cap" "$t1" "$out"
# The sample's alignment is committed: the golden one TestAlignSample compares against.
mv "$out/$sample.tdt" internal/chart/testdata/tdt/sample.tdt
