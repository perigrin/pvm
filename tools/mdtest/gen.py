"""Build topic files for a tier from the extracted .t claims.

Sources, pins, token facts and refusal records are copied VERBATIM out
of `corpus.json`. Only prose is written fresh.

THE PORT MOVES BYTES. Retyping a pin is how this corpus shipped seven
wrong ones in a single milestone, so no generator ever retypes a claim
-- it copies, and `check()` refuses to finish a tier with a file left
behind.

Usage, from the repo root:

    python3 tools/mdtest/extract.py          # .t -> corpus.json
    python3 tools/mdtest/<tier>.py           # corpus.json -> topics
"""
import json
import os

HERE = os.path.dirname(os.path.abspath(__file__))
DATA = json.load(open(os.path.join(HERE, 'corpus.json')))
OUT = 'conformance/mdtest'
FENCE = '```'


def case(tier, name, title, prose):
    """Render one `##` case from the .t file it replaces."""
    d = DATA[tier][name]
    parts = ['## ' + title, '', prose.strip(), '',
             FENCE + 'perl', d['source'], FENCE, '']

    behavior = ['parses: ' + ('yes' if d['parses'] else 'no')]
    if d['status']:
        s = d['status']
        issue = s.split('Issue ', 1)[1].split('.')[0].strip() if 'Issue ' in s else ''
        refusal = s.split('Refusal ', 1)[1].split('.')[0].strip() if 'Refusal ' in s else ''
        behavior.append('refuses: ' + (issue or 'unfiled'))
        if refusal:
            behavior.append('refusal: ' + refusal)
    parts += [FENCE + 'behavior'] + behavior + [FENCE, '']

    if d['output'] is not None:
        parts += [FENCE + 'output', d['output'], FENCE, '']

    if d['facts']:
        parts += [FENCE + 'tokens'] + d['facts'] + [FENCE, '']
    return '\n'.join(parts)


def topic(tier, path, heading, intro, cases):
    """Write one topic file; returns the .t names it covered."""
    body = ['# ' + heading, '', intro.strip(), '']
    for name, title, prose in cases:
        body.append(case(tier, name, title, prose))
    os.makedirs(OUT, exist_ok=True)
    open(os.path.join(OUT, path), 'w').write('\n'.join(body).rstrip() + '\n')
    print('  %-32s %2d cases' % (path, len(cases)))
    return {n for n, _, _ in cases}


def check(tier, covered):
    """Every .t file in the tier must appear in exactly one topic.

    The guard against a silent drop: a port that loses a file loses
    whatever that file claimed, and nothing else would notice until the
    .t files were deleted.
    """
    all_files = set(DATA[tier])
    missing = all_files - covered
    extra = covered - all_files
    if missing:
        raise SystemExit('%s: NOT PORTED: %s' % (tier, sorted(missing)))
    if extra:
        raise SystemExit('%s: unknown files: %s' % (tier, sorted(extra)))
    print('  %s: all %d files ported' % (tier, len(all_files)))
