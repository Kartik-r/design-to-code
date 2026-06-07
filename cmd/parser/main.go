package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	astpkg "github.com/Kartik-r/design-to-code/internal/ast"
	"github.com/Kartik-r/design-to-code/internal/graph"
	"github.com/Kartik-r/design-to-code/pkg/types"
)

func main() {
	dir     := flag.String("dir", ".", "Go project directory to analyze")
	output  := flag.String("output", "", "Output file path for JSON graph (e.g. graph.json)")
	query   := flag.String("query", "", "Query: deps:<nodeID>  callers:<nodeID>  impacted:<nodeID>")
	verbose := flag.Bool("verbose", false, "Print every node and its type")
	tests   := flag.Bool("tests", false, "Include _test.go files in analysis")

	flag.Parse()

	fmt.Println("Design-to-Code Parser")
	fmt.Printf("Target: %s\n\n", *dir)

	g, err := runAnalysis(*dir, astpkg.WalkConfig{IncludeTests: *tests})
	if err != nil {
		log.Fatalf("Analysis failed: %v", err)
	}

	fmt.Println(g.Summary())
	printBreakdown(g)

	if *output != "" {
		if err := g.WriteJSON(*output); err != nil {
			log.Fatalf("Failed to write JSON: %v", err)
		}
		fmt.Printf("\nGraph written to: %s\n", *output)
	}

	if *query != "" {
		runQuery(g, *query)
	}

	if *verbose {
		printAllNodes(g)
	}
}

func runAnalysis(target string, cfg astpkg.WalkConfig) (*graph.Graph, error) {
	info, err := os.Stat(target)
	if err != nil {
		return nil, fmt.Errorf("cannot access %s: %w", target, err)
	}

	if info.IsDir() {
		return astpkg.AnalyzeDirectory(target, cfg)
	}

	nodes, edges, err := astpkg.ParseFile(target)
	if err != nil {
		return nil, err
	}
	g := graph.New()
	for _, n := range nodes {
		g.AddNode(n)
	}
	for _, e := range edges {
		g.AddEdge(e)
	}
	return g, nil
}

func runQuery(g *graph.Graph, query string) {
	parts := strings.SplitN(query, ":", 2)
	if len(parts) != 2 {
		fmt.Println("\nInvalid query format.")
		fmt.Println("Usage: --query deps:<nodeID>")
		fmt.Println("       --query callers:<nodeID>")
		fmt.Println("       --query impacted:<nodeID>")
		return
	}

	queryType := parts[0]
	nodeID    := parts[1]

	fmt.Printf("\nQuery [%s] on node: %s\n", queryType, nodeID)
	fmt.Println(strings.Repeat("-", 50))

	if !g.HasNode(nodeID) {
		fmt.Printf("Node '%s' not found.\n", nodeID)
		fmt.Println("Tip: node IDs follow the pattern  package.FunctionName")
		fmt.Println("     or  package.ReceiverType.MethodName  for methods")
		return
	}

	var results []*types.Node
	switch queryType {
	case "deps":
		results = g.GetDependencies(nodeID)
		fmt.Printf("Direct dependencies (%d):\n", len(results))
		for _, n := range results {
			fmt.Printf("  → [%-12s] %s\n", n.Type, n.ID)
		}
	case "callers":
		results = g.GetCallers(nodeID)
		fmt.Printf("Direct callers (%d):\n", len(results))
		for _, n := range results {
			fmt.Printf("  ← [%-12s] %s\n", n.Type, n.ID)
		}
	case "impacted":
		results = g.GetImpacted(nodeID)
		fmt.Printf("Transitively impacted if this changes (%d):\n", len(results))
		for _, n := range results {
			fmt.Printf("  ⚡ [%-12s] %s\n", n.Type, n.ID)
		}
	default:
		fmt.Printf("Unknown query type '%s'. Valid: deps, callers, impacted\n", queryType)
	}
}

func printBreakdown(g *graph.Graph) {
	fmt.Println("\nNode breakdown:")
	for nodeType, count := range g.NodeCountByType() {
		fmt.Printf("  %-14s %d\n", nodeType, count)
	}
}

func printAllNodes(g *graph.Graph) {
	fmt.Printf("\nAll nodes (%d):\n", g.NodeCount())
	for _, n := range g.GetAllNodes() {
		fmt.Printf("  [%-12s] %s\n", n.Type, n.ID)
	}
}