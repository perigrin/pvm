---
title: A .pmt container return type keeps its element type, once the paper defines containers
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-06T00:35:06.123503181Z
updated: 2026-10-16T18:48:32.396269577Z
---

## Context

- paths: `internal/types/declared.go` (`types.Signature`), `internal/parse/decl.go` (`typeExpr`, `returnType`)
- commands: `make test`

A `.tpm` return type that is a container drops its element: `sub f (Str $x) List[Str];` records `Returns: List`, because `types.Signature` has no field for a return element (built by 01a10177-49d9). Also provisional, by the `ponytail:` comment on `typeExpr`: a union of containers ORs their elements, and any name may take brackets (`Scalar[Str]`). The parameter side records `Param.Element`. What a container means is the types paper's question (issue 01a0fa46-a60d-77f6-8462-e9b69f67db6d, parametric containers); this issue waits for its answer rather than inventing one.

## Prerequisites

- [ ] Types paper: parametric containers, tuples, lvalue lists, and multi and context-indexed functions

## Steps

- [ ] When the paper defines containers, write failing test `TestTpmContainerReturnTypeKeepsElement`; record the return element as the paper's model says; `make test`; commit

## Acceptance criteria

- [ ] a container return type keeps its element (`go test ./internal/parse/ -run '^TestTpmContainerReturnTypeKeepsElement$' -count=1 -v | grep -q '^--- PASS: TestTpmContainerReturnTypeKeepsElement'`)
