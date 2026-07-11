# Design-to-Code — Runbook

One-time setup and every-time run commands for Component 1 and Component 2.
Run everything from the repo root (`design-to-code/`) unless a step says otherwise.

---

## Part A — Component 1 (Graph Extraction)

### A1. One-time setup
Nothing extra — Component 1 is pure Go, no external dependencies beyond the Go toolchain.

```bash
go build ./...
```
Confirms everything compiles. Run this any time after pulling changes.

### A2. Run the parser on any Go codebase
```bash
go run ./cmd/parser --dir <path-to-go-project> --output graph.json --verbose
```
Parses the target directory's `.go` files into a node/edge graph (`graph.json`). This is Component 1's deliverable — already evaluated (see RESULTS.md, gin/gonic benchmark).

### A3. Run Component 1's tests
```bash
go test ./internal/ast/... ./internal/graph/...
```
Confirms the parser and graph package still behave correctly. Run after any change to `internal/ast` or `internal/graph`.

---

## Part B — Component 2 (Schema Extraction + Classification)

### B1. One-time setup — Python environment
```bash
cd python/classifier
python3 -m venv venv
source venv/bin/activate
pip install scikit-learn pandas joblib
```
Creates an isolated Python environment so you don't fight macOS's system-Python restrictions. Do this once per machine.

**Every new terminal session**, before running any Python step below, reactivate it:
```bash
cd python/classifier
source venv/bin/activate
```
You'll see `(venv)` in your prompt when active. This also matters when running Go — see B5.

### B2. One-time setup — train the classifier
Inside `python/classifier`, venv active:
```bash
python3 generate_training_data.py
```
Synthesizes 600 labeled feature vectors (200 each: monolith, microservices, event-driven) into `training_data.csv`. Synthetic data used only to teach the model each pattern's general shape — not your real sample apps.

```bash
python3 train.py
```
Trains a logistic regression classifier, prints accuracy + confusion matrix, saves the trained model to `pattern_classifier.joblib`. Re-run only if you change `features.py` or `generate_training_data.py` — the saved `.joblib` file is what Go uses at runtime.

### B3. Validate the classifier against real code
From the repo root:
```bash
go run ./cmd/schemadump
```
Runs Component 1's parser + Component 2's extractor against the 3 canonical sample apps (`testdata/samples/monolith`, `/microservices`, `/eventdriven`), writing real schema output to `schema_dumps/*.json`.

Then, back in `python/classifier` (venv active):
```bash
python3 evaluate_on_real.py ../../schema_dumps/monolith.json ../../schema_dumps/microservices.json ../../schema_dumps/eventdriven.json
```
Checks whether the trained classifier correctly labels real, held-out extracted schemas. This is the actual evaluation number for viva.

### B4. Run Component 2's Go tests
```bash
go test ./internal/schema/... ./internal/generator/...
```
Regression check for the extractor (and the currently-unused generator package, pending cleanup).

### B5. Run the full pipeline end-to-end (pending CLI wiring)
Once the CLI flag is wired:
```bash
go run ./cmd/parser --dir <gin-repo-path> --generate-iac
```
Calls, in order: AST parse to graph, graph to schema extraction, Python classifier subprocess, IaC template generation, then validation. **The venv must be active in this same terminal** — Go's `exec.Command("python3", ...)` uses whatever `python3` is on `PATH` at that moment.

---

## Quick reference — daily startup sequence
```bash
cd design-to-code
cd python/classifier && source venv/bin/activate && cd ../..
go build ./...
```