#!/bin/sh
# Early V2 speed reading (perigrin, 2026-11-27, "Take a quiet reading now"): design §7's protocol,
# interleaved B A B A B A in one session. Usage: drive.sh <after-commit> <out-dir>
# Before is e3432c0f (pu with V1 after the baseline retake). Each side is built once, then run in turn.
set -u
after=$1
out=$2
repo=/home/user/pvm
corpus=/opt/corpora/PerlOnJava/src/test/resources/unit
mkdir -p "$out"
before=$(git -C "$repo" rev-parse e3432c0f) || exit 2
after=$(git -C "$repo" rev-parse "$after") || exit 2
for side in before after; do
	eval c=\$$side
	wt=$out/wt-$side
	[ -d "$wt" ] || git -C "$repo" worktree add -q --detach "$wt" "$c" || exit 2
	(cd "$wt" && go build -o "$out/ratchet-$side" ./internal/chart/cmd/ratchet) || exit 2
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
	(cd "$out/wt-$side" && T1_CORPUS=$corpus "$out/ratchet-$side") >"$out/run$n-$side.stdout" 2>"$out/run$n-$side.stderr"
	rc=$?
	end=$(date -u +%Y-%m-%dT%H:%M:%SZ)
	git -C "$out/wt-$side" checkout -q -- test/e2e/.perl-version 2>/dev/null
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
