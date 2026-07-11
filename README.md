# design-to-code

> Graph-aware codebase intelligence and automated infrastructure scaffold generation for Go/Gin backends.
> MTech Dissertation, BITS Pilani WILP 2024–2026

![Go](https://img.shields.io/badge/Go-1.22+-00ACD7?style=flat-square)
![Python](https://img.shields.io/badge/Python-3.12+-3776AB?style=flat-square)
![License](https://img.shields.io/badge/license-MIT-green?style=flat-square)
![Status](https://img.shields.io/badge/status-active-brightgreen?style=flat-square)

**Graph-Aware Codebase Intelligence and Automated Infrastructure Scaffold Generation from Architectural Diagrams.**

MTech Dissertation — Kartikey Rai · BITS Pilani WILP · Student ID: 2024AA05563

---

## What this does

A two-component pipeline that takes a Go/Gin backend codebase and produces deployable, validated Infrastructure-as-Code — with no human writing Terraform or Kubernetes YAML by hand, and no external LLM API in the loop.

**Component 1 — Graph Extraction.** Parses any Go codebase with `go/ast` into a queryable property graph: packages, files, functions, types, call edges, import edges, and interface-implementation edges. Framework-agnostic — works on any Go source.

**Component 2 — Schema Extraction, Classification & IaC Generation.** Consumes Component 1's graph for a Gin-based backend application specifically, extracts a canonical architecture schema (routes, ports, database dependencies, environment variables, inter-service calls), classifies the architectural pattern (monolith / microservices / event-driven) with a trained classifier, then deterministically generates Terraform (AWS) and Kubernetes manifests — validated against the real `terraform` and `kubeconform` CLIs.

## Architecture

```
Go/Gin application source
        |
        v
[Component 1]  go/ast parser  -->  graph.json (nodes + edges, framework-agnostic)
        |
        v
[Component 2a] Schema extractor  -->  canonical_architecture.json
        |                             (routes, ports, db deps, env vars,
        |                              resolved cross-service calls)
        v
[Component 2b] Trained classifier (scikit-learn, local, no API)
        |                             --> predicted pattern label
        v
[Component 2c] Deterministic generator (Go text/template)
        |                             --> main.tf + k8s.yaml
        v
[Component 2d] Validator  -->  terraform validate + kubeconform
```

Every stage after the first is a consumer of the previous stage's JSON output, not a re-parse of source -- each artifact (`graph.json`, `schema.json`) is independently inspectable and testable.

## Evaluation

**Component 1** was stress-tested against [`github.com/gin-gonic/gin`](https://github.com/gin-gonic/gin) (the framework itself) to prove the parser scales to real, large codebases.

**Component 2** is evaluated against 3 hand-built canonical sample applications covering the 3 architectural patterns it classifies (`testdata/samples/monolith`, `/microservices`, `/eventdriven`) -- small Gin *applications*, not the framework itself -- including 3/3 real-code classification accuracy and 6/6 real `terraform validate`/`kubeconform` passes.

Full numbers for both components: [`RESULTS.md`](./RESULTS.md).

## Project structure
```
design-to-code/
├── cmd/
│   ├── parser/          # Component 1 CLI: Go source -> graph.json
│   ├── schemadump/      # Dumps schema.json for the 3 canonical samples (classifier validation)
│   └── generate/        # Component 2 end-to-end CLI: source -> validated IaC
├── internal/
│   ├── ast/              # go/ast parsing, call/import/literal-arg extraction
│   ├── graph/             # Graph data model, query API (deps/callers/impacted), JSON export
│   ├── schema/             # Canonical schema extraction, multi-service resolution
│   ├── generator/           # Deterministic Terraform/K8s generation (text/template)
│   └── validator/            # terraform validate / kubeconform wrappers
├── pkg/types/                 # Shared types: graph Node/Edge, ArchitectureSchema
├── python/classifier/           # Trained pattern classifier (scikit-learn)
├── testdata/samples/              # 3 canonical Gin applications (monolith, microservices, event-driven)
├── RESULTS.md                        # Evaluation results, both components
├── RUNBOOK.md                        # Full setup + run instructions
└── go.mod
```

## Quick start

```bash
# Component 1: parse any Go codebase into a graph
go run ./cmd/parser --dir <path-to-go-project> --output graph.json

# Component 2: full pipeline on a Gin application (single service)
go run ./cmd/generate --dir <path-to-gin-app> --out ./iac_output/myapp

# Component 2: full pipeline on a multi-service Gin application
go run ./cmd/generate --dir <path-to-services-parent-dir> --multi-service --out ./iac_output/myapp
```

See [`RUNBOOK.md`](./RUNBOOK.md) for full one-time setup (Python classifier environment, validator CLIs) and every command explained.

## Status

- [x] Go AST parser -- packages, files, functions, structs, interfaces
- [x] Call graph extraction (with literal/identifier argument capture)
- [x] Interface -> implementation edges
- [x] JSON export + query API (deps / callers / impacted)
- [x] Canonical architecture schema extraction (routes, ports, DB, env vars)
- [x] Multi-service extraction with cross-service call resolution
- [x] Trained pattern classifier (monolith / microservices / event-driven) -- no manual labeling, no API
- [x] Deterministic Terraform + Kubernetes generation -- no LLM, no network call
- [x] Validation harness (`terraform validate`, `kubeconform`) -- 6/6 real passes
- [x] End-to-end CLI (`cmd/generate`)

## License

MIT