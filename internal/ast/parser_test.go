package ast

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Kartik-r/design-to-code/pkg/types"
)

// testdataPath finds the absolute path to a testdata file.
// Uses runtime.Caller so tests work regardless of which directory you run them from.
func testdataPath(t *testing.T, filename string) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine current file path")
	}
	// parser_test.go is at internal/ast/parser_test.go
	// so go up two levels to reach project root
	root := filepath.Join(filepath.Dir(currentFile), "..", "..")
	return filepath.Join(root, "testdata", filename)
}

func TestParseFile_FindsAllEntities(t *testing.T) {
	path := testdataPath(t, "sample.go")
	nodes, edges, err := ParseFile(path)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) == 0 {
		t.Fatal("expected nodes, got none")
	}
	if len(edges) == 0 {
		t.Fatal("expected edges, got none")
	}

	// Check all expected entities are present
	names := make(map[string]bool)
	for _, n := range nodes {
		names[n.Name] = true
	}

	for _, want := range []string{"FormatUser", "ProcessUsers", "NewUserService", "User", "Repository"} {
		if !names[want] {
			t.Errorf("expected to find node named %q but it was missing", want)
		}
	}
}

func TestParseFile_ExtractsImports(t *testing.T) {
	path := testdataPath(t, "sample.go")
	nodes, _, err := ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	count := 0
	for _, n := range nodes {
		if n.Type == types.NodePackage {
			count++
		}
	}

	// sample.go imports "fmt" and "strings" — so 2 package nodes expected
	if count != 2 {
		t.Errorf("expected 2 import nodes, got %d", count)
	}
}

func TestParseFile_MethodHasReceiver(t *testing.T) {
	path := testdataPath(t, "sample.go")
	nodes, _, err := ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, n := range nodes {
		if n.Type == types.NodeMethod && n.Name == "GetUser" {
			if n.Metadata["receiver"] != "UserService" {
				t.Errorf("expected receiver=UserService, got %q", n.Metadata["receiver"])
			}
			return // found it, test passes
		}
	}
	t.Error("method GetUser not found in parsed nodes")
}