---
title: TestVerifySymlinks races other make test runs through the shared temp dir
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-20T02:10:43.358355035Z
updated: 2026-10-20T02:10:43.358355035Z
---

internal/cli/symlinks_test.go: TestVerifySymlinks and TestCreateSymlinks both write test_binary and its component symlinks into the shared os.TempDir() rather than t.TempDir(), so two make test runs on one machine race: one run's TestCreateSymlinks leaves pvm/pvx/pm/psc links that the other's TestVerifySymlinks finds ("Expected all symlinks to be missing, but some were found"). Seen 2026-10-20 failing a core-module-types merge while parallel workers ran make test. Fix: use t.TempDir() in both tests.
