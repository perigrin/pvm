"""Print every case source in a tier, for acceptance criteria to grep.

    python3 tools/mdtest/tiersrc.py 04_operators | grep -q '+='

Acceptance criteria used to ask this of the corpus with

    awk '/^--- source/{s=1;next} /^--- expect/{s=0} s' conformance/04_operators/*.t

which knew that a case was a `.t` file, that a tier was a directory of
them, and how a section was delimited. All three changed with the topic
format. The QUESTION did not: "does this tier's source contain X".

A tier is found by the `**Tier NN name.**` line a topic declares, the
same way the op-budget lint finds it -- declared rather than derived,
because the optimiser erases the construct a case is about.
"""
import glob
import re
import sys

if len(sys.argv) != 2:
    sys.exit('usage: tiersrc.py <tier>   e.g. 04_operators')

want = sys.argv[1]
for path in sorted(glob.glob('conformance/mdtest/*.md')):
    raw = open(path).read()
    m = re.search(r'(?m)^\*\*Tier (\d\d) (\w+)', raw)
    if not m or m.group(1) + '_' + m.group(2) != want:
        continue
    for block in re.finditer(r'(?m)^```perl\n(.*?)\n```', raw, re.S):
        print(block.group(1))
