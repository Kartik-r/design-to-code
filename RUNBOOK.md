# Design-to-Code — Runbook

One-time setup and every-time run commands for Component 1 and Component 2.
Run everything from the repo root (`design-to-code/`) unless a step says otherwise.

---

## Part A — Component 1 (Graph Extraction)

Works on any Go codebase, framework-agnostic.

### A1. One-time setup
Nothing extra beyond the Go toolchain.
```bash
go build ./...
```

### A2. Parse any Go codebase into a graph
```bash
go run ./cmd/parser --dir <path-to-go-project> --output graph.json --verbose
```
See `RESULTS.md` for the gin-gonic/gin benchmark run this way.

### A3. Run Component 1's tests
```bash
go test ./internal/ast/... ./internal/graph/...
```

---

## Part B — One-time setup for Component 2 (Python classifier environment)

Component 2's pattern classifier is a trained Python model, shelled out to by Go. Set this up once per machine.

```bash
cd python/classifier
python3 -m venv venv
source venv/bin/activate
pip install scikit-learn pandas joblib
python3 generate_training_data.py   # -> training_data.csv (600 synthetic samples)
python3 train.py                    # -> pattern_classifier.joblib + prints accuracy/confusion matrix
cd ../..
```

**Every new terminal session**, before running anything from Component 2 (Python scripts *or* the Go CLI, since Go shells out to Python), reactivate the venv:
```bash
cd python/classifier && source venv/bin/activate && cd ../..
```
You'll see `(venv)` in your prompt when active. If it's not active when you run `go run ./cmd/generate`, Go will hit your system Python (no scikit-learn installed) and the classifier call will fail with `ModuleNotFoundError`.

---

## Part C — One-time setup for IaC validation

```bash
brew tap hashicorp/tap
brew install hashicorp/tap/terraform
brew install kubeconform
```
(Terraform isn't in Homebrew's main repo anymore due to their license change — it needs HashiCorp's own tap.)

Verify both:
```bash
terraform -v
kubeconform -v
```

---

## Part D — Running the full pipeline (Component 1 + Component 2, end to end)

This is the command you'll actually use for demos. One binary, everything chained: parse → extract schema → classify pattern → generate Terraform + K8s → validate.

**Single-service application (monolith pattern):**
```bash
go run ./cmd/generate --dir <path-to-gin-app> --out ./iac_output/myapp
```

**Multi-service application (microservices / event-driven patterns)** — `--dir` should point at a parent directory whose immediate subdirectories are each one service:
```bash
go run ./cmd/generate --dir <path-to-services-parent-dir> --multi-service --out ./iac_output/myapp
```

**On the 3 canonical sample apps** (reproduces `RESULTS2.md`):
```bash
go run ./cmd/generate --dir ./testdata/samples/monolith --out ./iac_output/monolith
go run ./cmd/generate --dir ./testdata/samples/microservices --multi-service --out ./iac_output/microservices
go run ./cmd/generate --dir ./testdata/samples/eventdriven --multi-service --out ./iac_output/eventdriven
```
Each writes `schema.json`, `main.tf`, and `k8s.yaml` into its `--out` directory, then prints PASS/FAIL from both validators.

**Useful flags:**
- `--schema-only` — stop after schema extraction + classification, skip generation/validation. Fast way to sanity-check classification on a new codebase.
- `--skip-validate` — generate files but skip running the validators (e.g. if `terraform`/`kubeconform` aren't installed on this machine).

### D1. Regenerate the held-out schema fixtures (for classifier validation)
```bash
go run ./cmd/schemadump
```
Writes `schema_dumps/{monolith,microservices,eventdriven}.json`. Then, with the venv active:
```bash
cd python/classifier
python3 evaluate_on_real.py ../../schema_dumps/monolith.json ../../schema_dumps/microservices.json ../../schema_dumps/eventdriven.json
```

### D2. Run all of Component 2's Go tests
```bash
go test ./internal/schema/... ./internal/generator/... ./internal/validator/... -v
```

---

## Part E — Trying it on a different repo (not the canonical samples)

Anyone cloning this project who wants to try it on their own Gin codebase, not just the bundled samples, do this:

```bash
# 1. Get the target codebase locally -- this is NOT part of this repo,
#    just a normal local clone anywhere on your machine.
git clone https://github.com/someuser/some-gin-api ~/some-gin-api

# 2. (Optional) Sanity-check Component 1 can parse it on its own first --
#    useful if something goes wrong later and you want to isolate whether
#    the problem is in parsing (Component 1) or extraction (Component 2).
go run ./cmd/parser --dir ~/some-gin-api --output /tmp/check_graph.json --verbose

# 3. Run the full Component 2 pipeline against it
go run ./cmd/generate --dir ~/some-gin-api --out ./iac_output/some-gin-api
```

If the target repo is structured as multiple services (a parent folder containing several subfolders, each its own `package main`), use `--multi-service` and point `--dir` at the *parent* folder, not at one service subfolder:
```bash
go run ./cmd/generate --dir ~/some-multi-service-repo --multi-service --out ./iac_output/some-multi-service-repo
```

**Caveat:** extraction rules in `internal/schema/extractor.go` are written specifically for Gin's calling conventions (`router.GET(...)`, `router.Run(":8080")`, `sql.Open(...)`, `os.Getenv(...)`). A Go app using a different HTTP framework will still parse fine at the Component 1 stage, but Component 2's schema extraction won't recognize its route/server-start calls, and you'll get an empty or incomplete schema back.

---

## Quick reference — daily startup sequence
```bash
cd design-to-code
cd python/classifier && source venv/bin/activate && cd ../..
go build ./...
```