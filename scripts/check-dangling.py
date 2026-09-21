#!/usr/bin/env python3
# ABOUTME: Reports which surviving git-zhi issues reference ids about to be deleted.
# ABOUTME: Run before deleting refs/zhi issue refs by hand; a lost ref breaks the graph.

"""Find dependency edges that would dangle after deleting a set of issues.

git-zhi stores each issue as its own ref with blocked_by/blocks edges
embedded, so deleting a ref leaves any edge naming it pointing at nothing.
This reads every issue ref directly rather than through `git-zhi list`,
which only reports open issues.

A dangling edge is not cosmetic: `git-zhi list` drops the affected issue
from the graph with a `references nonexistent upstream` notice, so the
issue stops appearing in dependency order.

Usage:  check-dangling.py <doomed-ids-file>
"""

import json
import subprocess
import sys


def issue_edges(ref):
    """Return (state, blocked_by + blocks) for one issue ref.

    An issue is `issue.md`: YAML frontmatter between `---` fences, then the
    body. Only three fields are needed here, and the edge lists are plain
    `- <uuid>` items, so this reads them directly rather than adding a YAML
    dependency for a one-off check.
    """
    blob = subprocess.run(
        ["git", "show", f"{ref}:issue.md"],
        capture_output=True, text=True,
    )
    if blob.returncode != 0:
        return None, []

    lines = blob.stdout.split("\n")
    if not lines or lines[0].strip() != "---":
        return None, []

    state, deps, field = None, [], None
    for line in lines[1:]:
        if line.strip() == "---":
            break
        if line.startswith("state:"):
            state = line.split(":", 1)[1].strip()
        elif line.rstrip() in ("blocked_by:", "blocks:"):
            field = line.rstrip()
        elif line.startswith("- ") and field:
            deps.append(line[2:].strip())
        elif line and not line.startswith((" ", "-")):
            field = None
    return state, deps


def main():
    doomed = {line.strip() for line in open(sys.argv[1]) if line.strip()}

    refs = subprocess.run(
        ["git", "for-each-ref", "refs/zhi/_/issues/", "--format=%(refname)"],
        capture_output=True, text=True, check=True,
    ).stdout.split()

    dangling = []
    unreadable = []
    for ref in refs:
        uuid = ref.rsplit("/", 1)[-1]
        if uuid in doomed:
            continue
        state, deps = issue_edges(ref)
        if state is None:
            unreadable.append(uuid)
            continue
        for dep in deps:
            if dep in doomed:
                dangling.append((uuid[:8], state, dep[:8]))

    print(f"surviving refs checked: {len(refs) - len(doomed)}")
    print(f"unreadable (schema differs): {len(unreadable)}")
    print(f"edges that would dangle: {len(dangling)}")
    for d in dangling:
        print("   ", d)


if __name__ == "__main__":
    main()
