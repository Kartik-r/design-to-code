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