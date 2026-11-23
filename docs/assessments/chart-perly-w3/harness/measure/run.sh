#!/bin/bash
# usage: run.sh TAG [profile]
M=/tmp/claude-0/-home-user/f57559af-f6d7-50b1-b4d1-2c3218e70a25/scratchpad/measure
tag=$1
for i in $(seq 180); do
  l=$(cut -d' ' -f1 /proc/loadavg); n=$(nproc)
  awk -v l=$l -v n=$n 'BEGIN{exit !(l<=n)}' && break
  echo "waiting: load $l > $n" >> $M/env-$tag.txt; sleep 10
done
{ echo "tag $tag"; date -u; uptime; echo "nproc $(nproc)"; uname -r; git -C $M/repo rev-parse HEAD; df -h / | tail -1; } >> $M/env-$tag.txt
cd $M/repo/internal/chart/cmd/ratchet
export T1_CORPUS=/opt/corpora/PerlOnJava/src/test/resources/unit MEASURE_OUT=$M MEASURE_TAG=$tag
if [ "$2" = profile ]; then
  MEASURE_PROFILE=1 $M/ratchet.test -test.run 'TestZZMeasure$' -test.v -test.timeout 2h -test.cpuprofile $M/constrain.pprof > $M/log-$tag.txt 2>&1
else
  $M/ratchet.test -test.run 'TestZZMeasure$' -test.v -test.timeout 2h > $M/log-$tag.txt 2>&1
fi
echo "exit $?" >> $M/env-$tag.txt; uptime >> $M/env-$tag.txt
