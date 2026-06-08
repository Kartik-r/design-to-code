package graph
import ("testing"; "github.com/Kartik-r/design-to-code/pkg/types")

func TestAddAndGetNode(t *testing.T) {
    g := New()
    g.AddNode(&types.Node{ID: "pkg.Foo", Type: types.NodeFunction, Name: "Foo"})
    if g.NodeCount() != 1 { t.Errorf("expected 1, got %d", g.NodeCount()) }
    if g.GetNode("pkg.Foo") == nil { t.Error("node not found") }
}

func TestDuplicateNodeIgnored(t *testing.T) {
    g := New()
    n := &types.Node{ID: "pkg.Foo", Type: types.NodeFunction, Name: "Foo"}
    g.AddNode(n); g.AddNode(n)
    if g.NodeCount() != 1 { t.Errorf("expected 1 after duplicate, got %d", g.NodeCount()) }
}

func TestAddEdge(t *testing.T) {
    g := New()
    g.AddEdge(&types.Edge{From: "pkg.A", To: "pkg.B", Type: types.EdgeCalls})
    if g.EdgeCount() != 1 { t.Errorf("expected 1 edge, got %d", g.EdgeCount()) }
}

func setupQueryGraph() *Graph {
    g := New()
    for _, id := range []string{"pkg.A", "pkg.B", "pkg.C", "pkg.D"} {
        g.AddNode(&types.Node{ID: id, Type: types.NodeFunction, Name: id})
    }
    // A → B → C ← D
    g.AddEdge(&types.Edge{From: "pkg.A", To: "pkg.B", Type: types.EdgeCalls})
    g.AddEdge(&types.Edge{From: "pkg.B", To: "pkg.C", Type: types.EdgeCalls})
    g.AddEdge(&types.Edge{From: "pkg.D", To: "pkg.C", Type: types.EdgeCalls})
    return g
}

func TestGetDependencies(t *testing.T) {
    g := setupQueryGraph()
    deps := g.GetDependencies("pkg.A")
    if len(deps) != 1 || deps[0].ID != "pkg.B" {
        t.Errorf("A should depend only on B, got %v", deps)
    }
}

func TestGetCallers(t *testing.T) {
    g := setupQueryGraph()
    callers := g.GetCallers("pkg.C")
    if len(callers) != 2 { t.Errorf("C should have 2 callers, got %d", len(callers)) }
}

func TestGetImpacted(t *testing.T) {
    g := setupQueryGraph()
    impacted := g.GetImpacted("pkg.C")
    if len(impacted) != 3 {
        t.Errorf("C should impact 3 nodes (A, B, D), got %d", len(impacted))
    }
}

func TestGetImpacted_NoCallers(t *testing.T) {
    g := setupQueryGraph()
    impacted := g.GetImpacted("pkg.A")
    if len(impacted) != 0 { t.Errorf("A has no callers, impacted should be empty, got %d", len(impacted)) }
}