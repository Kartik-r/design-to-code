package graph

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Kartik-r/design-to-code/pkg/types"
)

type graphJSON struct {
	NodeCount int           `json:"node_count"`
	EdgeCount int           `json:"edge_count"`
	Nodes     []*types.Node `json:"nodes"`
	Edges     []*types.Edge `json:"edges"`
}

func (g *Graph) ToJSON() ([]byte, error) {
	return json.MarshalIndent(graphJSON{
		NodeCount: g.NodeCount(), EdgeCount: g.EdgeCount(),
		Nodes: g.GetAllNodes(), Edges: g.GetAllEdges(),
	}, "", "  ")
}

func (g *Graph) WriteJSON(outputPath string) error {
	data, err := g.ToJSON()
	if err != nil {
		return fmt.Errorf("serializing: %w", err)
	}
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", outputPath, err)
	}
	return nil
}