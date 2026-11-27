#!/bin/sh
# V2 verdict reading (design §7): the speed protocol interleaved B A B A B A in one quiet session.
# Usage, from the worktree root: sh docs/assessments/chart-perly-w4/verdict/drive.sh [after-commit]
# Before is e3432c0f (pu with V1 after the baseline retake); after defaults to pu with every V2 code PR.
# Each side is checked out and built once outside the repository, then run in turn; the run files land here.
set -u
cd "$(git rev-parse --show-toplevel)" || exit 2
out=docs/assessments/chart-perly-w4/verdict
scratch=/tmp/claude-0/-home-user/f57559af-f6d7-50b1-b4d1-2c3218e70a25/scratchpad/v2/verdict-wt
corpus=/opt/corpora/PerlOnJava/src/test/resources/unit
before=$(git rev-parse --verify e3432c0f^{commit}) || exit 2
after=$(git rev-parse --verify "${1:-1477098790ad6668b402eb0c1d0ee67586e18fd6}^{commit}") || exit 2
[ "$after" != "$before" ] || { echo "after is before" >&2; exit 2; }
if git grep -q '\.Constrain(' "$after" -- internal/chart/forest/forest.go; then
	echo "forest.go at $after still calls Constrain" >&2
	exit 2
fi
# A reading is taken once: never overwrite a kept run.
if ls "$out"/run* >/dev/null 2>&1; then
	echo "$out already holds runs; refusing to overwrite" >&2
	exit 2
fi
mkdir -p "$out" "$scratch" || exit 2
cleanup() {
	for side in before after; do
		[ -d "$scratch/wt-$side" ] && git worktree remove --force "$scratch/wt-$side"
	done
	rm -rf "$scratch"
}
trap cleanup EXIT
for side in before after; do
	eval c=\$$side
	git worktree add -q --detach "$scratch/wt-$side" "$c" || exit 2
	(cd "$scratch/wt-$side" && go build -o "$scratch/ratchet-$side" ./internal/chart/cmd/ratchet) || exit 2
done
# The ratchet reads its corpora relative to the module root, so each side runs from its own checkout.
n=0
for side in before after before after before after; do
	n=$((n + 1))
	eval c=\$$side
	waited=0
	while [ "$(cut -d' ' -f1 /proc/loadavg | awk '{print ($1 < 1.0)}')" != 1 ] && [ $waited -lt 1800 ]; do
		sleep 15
		waited=$((waited + 15))
	done
	la=$(cat /proc/loadavg)
	start=$(date -u +%Y-%m-%dT%H:%M:%SZ)
	(cd "$scratch/wt-$side" && T1_CORPUS=$corpus "$scratch/ratchet-$side") >"$out/run$n-$side.stdout" 2>"$out/run$n-$side.stderr"
	rc=$?
	end=$(date -u +%Y-%m-%dT%H:%M:%SZ)
	git -C "$scratch/wt-$side" checkout -q -- test/e2e/.perl-version 2>/dev/null
	cat >"$out/run$n-$side.meta" <<EOF
run: $n
state: $side
commit: $c
loadavg-before: $la
load1: ${la%% *}
cores: $(nproc)
kernel: $(uname -r)
waited-for-quiet-s: $waited
start: $start
end: $end
exit: $rc
EOF
done
echo done
