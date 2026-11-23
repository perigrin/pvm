#!/bin/bash
# usage: run.sh TAG [time]
C=/tmp/claude-0/-home-user/f57559af-f6d7-50b1-b4d1-2c3218e70a25/scratchpad/combined
tag=$1
for i in $(seq 180); do
  l=$(cut -d' ' -f1 /proc/loadavg); n=$(nproc)
  awk -v l=$l -v n=$n 'BEGIN{exit !(l<=n)}' && break
  echo "waiting: load $l > $n" >> $C/env-$tag.txt; sleep 10
done
{ echo "tag $tag"; date -u; uptime; echo "nproc $(nproc)"; uname -r; git -C $C/repo rev-parse HEAD; go version; df -h / | tail -1; } >> $C/env-$tag.txt
cd $C/repo/internal/chart/cmd/ratchet
export T1_CORPUS=/opt/corpora/PerlOnJava/src/test/resources/unit CB_OUT=$C CB_TAG=$tag
[ "$2" = time ] && export CB_TIME=1
$C/ratchet.test -test.run 'TestZZCombined$' -test.v -test.timeout 3h > $C/log-$tag.txt 2>&1
echo "exit $?" >> $C/env-$tag.txt; uptime >> $C/env-$tag.txt
