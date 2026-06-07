package ast

import (
	"fmt"
	"log"

	"github.com/Kartik-r/design-to-code/internal/graph"
)

// AnalyzeFile parses a single file and loads all entities into the graph.
func AnalyzeFile(filePath string, g *graph.Graph) error {
	nodes, edges, err := ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("analyzing %s: %w", filePath, err)
	}
	for _, n := range nodes {
		g.AddNode(n)
	}
	for _, e := range edges {
		g.AddEdge(e)
	}
	return nil
}

// AnalyzeDirectory parses an entire Go project and returns a populated graph.
// Files that fail to parse are logged and skipped — one bad file won't abort everything.
func AnalyzeDirectory(rootDir string, cfg WalkConfig) (*graph.Graph, error) {
	files, err := FindGoFiles(rootDir, cfg)
	if err != nil {
		return nil, fmt.Errorf("walking %s: %w", rootDir, err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no .go files found in %s", rootDir)
	}

	g := graph.New()
	for _, filePath := range files {
		if err := AnalyzeFile(filePath, g); err != nil {
			log.Printf("WARN: skipping %s: %v", filePath, err)
		}
	}
	return g, nil
}