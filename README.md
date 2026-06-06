# design-to-code
Go AST-based codebase intelligence and diagram-to-IaC scaffold generator · MTech Dissertation, BITS Pilani WILP 2024-2026

Graph-Aware Codebase Intelligence and Automated Infrastructure Scaffold Generation from Architectural Diagrams.
MTech Dissertation — Kartikey Rai · BITS Pilani WILP · Student ID: 2024AA05563
What this does
Component 1 — Parses a Go codebase using go/ast and builds a code graph: packages, files, functions, types, call edges, import edges, and interface-implementation edges.
Component 2 — Accepts an architectural diagram as input and generates Terraform + Kubernetes IaC scaffolds for canonical cloud-native patterns.
Evaluation corpus
github.com/gin-gonic/gin — used as the open-source Go codebase for parser evaluation.
Run
go run ./cmd/graphparser --path ./corpus/gin --output graph.json
Roadmap
[ ] Go AST parser — packages, files, functions
[ ] Call graph extraction
[ ] Interface → implementation edges
[ ] JSON / GraphML export
[ ] Diagram-to-IaC generator (Component 2)
