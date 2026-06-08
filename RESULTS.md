# Evaluation Results — gin-gonic/gin

## Run Info
- Date: June 7, 2026
- Tool version: Component 1 — Go AST Graph Engine (with call graph)
- Gin commit: d75fcd4c9ab260e5225de590f1f0f8c0e0e12d11

## Metrics
- Total nodes: 715
- Total edges: 2,286
- Analysis time: 0.436s (real) for 59 files, ~150k LOC

## Edge breakdown
- Before CALLS edges (structural only): 825
- After CALLS edges: 2,286
- CALLS edges extracted: 1,461

## Node breakdown
- FILE:       59
- PACKAGE:    60
- INTERFACE:  14
- STRUCT:     62
- METHOD:     374
- FUNCTION:   146

## Query Results

### deps:gin.Default — what does gin.Default call?
Direct dependencies (4):
- gin.debugPrintWARNINGDefault
- gin.New
- gin.Logger
- gin.Recovery

### callers:gin.New — who calls gin.New?
Direct callers (3):
- file:/tmp/gin/gin.go  (file-level contains edge)
- gin.Default
- gin.CreateTestContext

### impacted:gin.New — what breaks if gin.New changes?
Transitively impacted (4):
- file:/tmp/gin/gin.go
- gin.Default
- gin.CreateTestContext
- file:/tmp/gin/test_helpers.go

## Correctness Verification
Query result for deps:gin.Default was manually verified against
gin-gonic/gin source (gin.go). The function calls exactly:
debugPrintWARNINGDefault(), New(), Logger(), Recovery() — all 4
correctly identified by the tool.

## Notes
- FILE nodes appear in callers/impacted results because the graph
  includes CONTAINS edges (file → function). This is expected behaviour.
  For clean function-only results, filter by node type in future work.
- Call resolution is best-effort (name-based). Full type resolution
  via go/types is documented as future work.
