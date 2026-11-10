#!/bin/sh
# ABOUTME: Load reproduction for lex-overhead: two chart ratchets started together beside one busy loop per CPU.
# ABOUTME: Prints every lex-overhead reading and the load; exits 1 on a reading above 1.3, under two readings, or no load raised.
#
# A close gate (git zhi milestone edit --state complete, and verify) runs its
# criteria one at a time: observed on git-zhi 0.7.2, three issues' six
# criteria ran strictly back to back. So the load a ratchet meets at a close
# is the machine's, not the gate's: other workers' builds and tests, and the
# ratchet's own parallel phases. This applies that load as one busy loop per
# CPU beside two ratchets, started together so both finish within 270 s.
# Run it from the repository root.

ratchets=2
tmp=$(mktemp -d)
hogs=
trap 'kill $hogs 2>/dev/null; rm -rf "$tmp"' EXIT INT TERM
load1() { cut -d' ' -f1 /proc/loadavg; }

before=$(load1)
for _ in $(seq "$(nproc)"); do
	sh -c 'while :; do :; done' &
	hogs="$hogs $!"
done
(while :; do load1 >>"$tmp/load"; sleep 2; done) &
hogs="$hogs $!"

pids=
for i in $(seq "$ratchets"); do
	make -s chart-ratchet >"$tmp/out$i" 2>"$tmp/err$i" &
	pids="$pids $!"
done
for p in $pids; do wait "$p"; done
load1 >>"$tmp/load"
kill $hogs 2>/dev/null
during=$(sort -g "$tmp/load" | tail -1)

n=0
bad=0
for i in $(seq "$ratchets"); do
	grep '^chart-ratchet: lex-overhead-\(std5k\|edit5k\) measurement' "$tmp/err$i"
	r=$(awk '$2 == "lex-overhead" { print $3 }' "$tmp/out$i")
	echo "loadcheck: ratchet $i lex-overhead ${r:-missing}"
	case $r in
	'' | *[!0-9.]*) continue ;;
	esac
	n=$((n + 1))
	awk -v r="$r" 'BEGIN { exit !(r > 1.3) }' && bad=1
done
echo "loadcheck: readings $n load-before $before load-during $during"
[ "$bad" = 0 ] && [ "$n" -ge 2 ] && awk -v a="$before" -v b="$during" 'BEGIN { exit !(b >= a + 1) }'
