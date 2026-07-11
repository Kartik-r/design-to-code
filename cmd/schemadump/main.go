// Command schemadump extracts the canonical architecture schema for each
// of the 3 canonical sample apps under testdata/samples and writes them
// to disk as JSON. Useful for:
//   - feeding python/classifier/evaluate_on_real.py real, held-out schemas
//   - manually inspecting extractor output while debugging
//
// Run from the repo root:
//
//	go run ./cmd/schemadump
//
// Writes to ./schema_dumps/{monolith,microservices,eventdriven}.json
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	astpkg "github.com/Kartik-r/design-to-code/internal/ast"
	"github.com/Kartik-r/design-to-code/internal/schema"
	"github.com/Kartik-r/design-to-code/pkg/types"
)

const outDir = "schema_dumps"

func main() {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "creating output dir: %v\n", err)
		os.Exit(1)
	}

	// monolith is a single-service directory, extracted directly (no
	// classifier call needed — ExtractSchema always labels it "monolith").
	g, err := astpkg.AnalyzeDirectory("./testdata/samples/monolith", astpkg.WalkConfig{})
	if err != nil {
		fail("parsing monolith sample", err)
	}
	mono := schema.ExtractSchema(g)
	dump("monolith.json", mono)

	// microservices and eventdriven are multi-service directories; pass ""
	// so the trained classifier predicts the pattern label itself.
	micro, err := schema.ExtractMultiService("./testdata/samples/microservices", "")
	if err != nil {
		fail("extracting microservices sample", err)
	}
	dump("microservices.json", micro)

	evt, err := schema.ExtractMultiService("./testdata/samples/eventdriven", "")
	if err != nil {
		fail("extracting eventdriven sample", err)
	}
	dump("eventdriven.json", evt)

	fmt.Printf("Wrote 3 schema files to ./%s/\n", outDir)
}

func dump(filename string, s *types.ArchitectureSchema) {
	out, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		fail("marshaling "+filename, err)
	}
	path := filepath.Join(outDir, filename)
	if err := os.WriteFile(path, out, 0644); err != nil {
		fail("writing "+path, err)
	}
	fmt.Printf("  %s (pattern: %s, %d service(s))\n", path, s.Pattern, len(s.Services))
}

func fail(context string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", context, err)
	os.Exit(1)
}