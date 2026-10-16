---
title: shapedCallReturns duplicates SelectShaped's UnknownCtx join
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-16T22:39:42.118785718Z
updated: 2026-10-16T22:39:42.118785718Z
---

`shapedCallReturns` (internal/infer/infer.go:1006-1024) loops over Scalar, List and Void and joins the non-Failed results by hand; since da0a (merge 7c74da24) `types.SelectShaped(sigs, unknown, sh, UnknownCtx)` returns exactly that join. Collapse the inner context loop to one SelectShaped call. `shapedOperatorCandidates` (962-981) keeps its loop because Taking and Answering return nil for UnknownCtx. Cleanup, not a defect: found by da0a's review gate (confidence 85), 2026-10-16. Precision and every infer golden must be unchanged.
