package types

// NodeType identifies the kind of code entity a node represents
type NodeType string

const (
    NodePackage   NodeType = "PACKAGE"
    NodeFile      NodeType = "FILE"
    NodeFunction  NodeType = "FUNCTION"
    NodeMethod    NodeType = "METHOD"
    NodeStruct    NodeType = "STRUCT"
    NodeInterface NodeType = "INTERFACE"
)

// EdgeType identifies the kind of relationship an edge represents
type EdgeType string

const (
    EdgeCalls      EdgeType = "CALLS"
    EdgeImports    EdgeType = "IMPORTS"
    EdgeImplements EdgeType = "IMPLEMENTS"
    EdgeContains   EdgeType = "CONTAINS"
    EdgeEmbeds     EdgeType = "EMBEDS"
)

// Node represents a code entity in the graph (function, type, file, package)
type Node struct {
    ID       string            `json:"id"`
    Type     NodeType          `json:"type"`
    Name     string            `json:"name"`
    Package  string            `json:"package"`
    FilePath string            `json:"file_path"`
    Metadata map[string]string `json:"metadata,omitempty"`
}

// Edge represents a directed relationship between two nodes.
// Args captures literal (string/int/bool) arguments passed on a CALLS edge,
// e.g. the route path in c.GET("/users/:id", h) or the port in r.Run(":8080").
// Non-literal arguments (variables, expressions) are omitted — this stays a
// generic capability of the parser, not gin-specific logic.
type Edge struct {
    From string   `json:"from"`
    To   string   `json:"to"`
    Type EdgeType `json:"type"`
    Args []string `json:"args,omitempty"`
}