"""Dump EVERY tier's claims as JSON, so the port copies rather than
retypes.

Retyping is how this milestone shipped seven wrong pins. The port moves
bytes; only prose is rewritten.
"""
import json
import os
import re
import glob

out = {}
for f in sorted(glob.glob('conformance/*/*.t')):
    tier = f.split('/')[1]
    name = os.path.basename(f)
    raw = open(f).read()
    header, _, rest = raw.partition('--- source')
    if not rest:
        continue

    def section(tag):
        if '--- ' + tag not in rest:
            return None
        return re.split(r'\n--- ', rest.split('--- ' + tag, 1)[1])[0].strip('\n')

    src = re.split(r'\n--- ', rest, 1)[0].strip('\n')
    status = re.search(r'(?m)^# STATUS refuses[^\n]*', raw)

    out.setdefault(tier, {})[name] = {
        'header': header,
        'source': src,
        'output': section('expect output'),
        'parses': 'expect parses' in rest,
        'parsent': 'expect parsent' in rest,
        'facts': [l for l in (section('expect tokens') or '').splitlines() if l.strip()],
        'status': status.group(0) if status else None,
    }

path = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'corpus.json')
json.dump(out, open(path, 'w'), indent=1)
print(sum(len(v) for v in out.values()), 'files across', len(out), 'tiers')
