---
title: TestVerifySymlinks races other make test runs through the shared temp dir
state: done
urgency: normal
milestone: v0.1
created: 2026-10-20T02:10:43.358355035Z
updated: 2026-10-20T02:22:30.403872144Z
sessions:
- start_sha: 88ef2c6b3fc39140ad4569332b89522a03eb0481
  end_sha: 88ef2c6b3fc39140ad4569332b89522a03eb0481
  commits: 0
  started_at: 2026-10-20T02:22:30.238965178Z
  ended_at: 2026-10-20T02:22:30.403872144Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-20T02:22:30.238965178Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-20T02:22:30.403872144Z
---

internal/cli/symlinks_test.go: TestVerifySymlinks and TestCreateSymlinks both write test_binary and its component symlinks into the shared os.TempDir() rather than t.TempDir(), so two make test runs on one machine race: one run's TestCreateSymlinks leaves pvm/pvx/pm/psc links that the other's TestVerifySymlinks finds ("Expected all symlinks to be missing, but some were found"). Seen 2026-10-20 failing a core-module-types merge while parallel workers ran make test. Fix: use t.TempDir() in both tests.
