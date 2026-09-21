#!/usr/bin/env python3
# ABOUTME: Renders open git-zhi issues to a markdown archive that can reconstitute them.
# ABOUTME: Written to defer M1/M2 while the conformance corpus is built, not to delete them.
"""Archive open chain issues.

The archive is the MECHANISM OF REVERSAL, not a record that something was
lost: these issues come back after the corpus is written. So it preserves
everything needed to recreate them -- id, title, milestone, state, the
dependency edges, and the full body including measurements and acceptance
criteria -- rather than a summary.

Usage:  git-zhi list --format json | archive-chain.py > archive.md
"""

import json
import sys
from collections import defaultdict


def short(uuid):
    """The 8-char prefix git-zhi displays."""
    return uuid.split("-")[0]


def main():
    issues = json.load(sys.stdin)["issues"]
    by_ms = defaultdict(list)
    for i in issues:
        by_ms[i.get("milestone", "(no milestone)")].append(i)

    titles = {i["id"]: i["title"] for i in issues}

    out = []
    w = out.append

    w("# Deferred chain: M1 and M2 as they stood before the corpus")
    w("")
    w("**Status:** archived 2026-09-21, to be reconstituted after the")
    w("conformance corpus is written. This is a DEFERRAL, not a deletion.")
    w("")
    w("The corpus rewrites what these milestones were measuring against, so")
    w("they are dropped now and rebuilt against the corpus later. This file")
    w("holds everything needed to recreate them: ids, dependency edges, and")
    w("full bodies with their measurements.")
    w("")
    w("Completed work is NOT here -- M0's 8 and M1's 32 done issues stay in")
    w("the chain as history. Only open work was dropped.")
    w("")

    total = len(issues)
    w(f"**{total} issues**, by milestone:")
    w("")
    for ms in sorted(by_ms):
        states = defaultdict(int)
        for i in by_ms[ms]:
            states[i["state"]] += 1
        detail = ", ".join(f"{n} {s}" for s, n in sorted(states.items()))
        w(f"- `{ms}` — {len(by_ms[ms])} ({detail})")
    w("")

    for ms in sorted(by_ms):
        w(f"## {ms}")
        w("")
        for i in sorted(by_ms[ms], key=lambda x: x["created"]):
            w(f"### {short(i['id'])} — {i['title']}")
            w("")
            w(f"- **id:** `{i['id']}`")
            w(f"- **state:** {i['state']}")
            w(f"- **urgency:** {i.get('urgency', 'normal')}")
            w(f"- **created:** {i['created'][:10]}")
            for field, label in (("blocked_by", "blocked by"), ("blocks", "blocks")):
                for dep in i.get(field) or []:
                    name = titles.get(dep, "(outside this archive)")
                    w(f"- **{label}:** `{short(dep)}` — {name}")
            w("")
            body = (i.get("body") or "").rstrip()
            if body:
                w(body)
                w("")

    sys.stdout.write("\n".join(out) + "\n")


if __name__ == "__main__":
    main()
