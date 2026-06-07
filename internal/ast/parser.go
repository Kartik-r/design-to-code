package ast

import (
    "fmt"
    "go/ast"
    "go/parser"
    "go/token"
    "path/filepath"
    "github.com/Kartik-r/design-to-code/pkg/types"
)

func ParseFile(filePath string) ([]*types.Node, []*types.Edge, error) {
    fset := token.NewFileSet()
    f, err := parser.ParseFile(fset, filePath, nil, parser.AllErrors)
    if err != nil {
        return nil, nil, fmt.Errorf("parsing %s: %w", filePath, err)
    }

    nodes := make([]*types.Node, 0, 32)
    edges := make([]*types.Edge, 0, 64)
    packageName := f.Name.Name
    fileNodeID := "file:" + filePath

    nodes = append(nodes, &types.Node{
        ID: fileNodeID, Type: types.NodeFile,
        Name: filepath.Base(filePath), Package: packageName, FilePath: filePath,
    })

    ast.Inspect(f, func(n ast.Node) bool {
        if n == nil { return false }
        switch x := n.(type) {

        case *ast.FuncDecl:
            nodeID := packageName + "." + x.Name.Name
            nodeType := types.NodeFunction
            meta := map[string]string{}
            if x.Recv != nil && len(x.Recv.List) > 0 {
                nodeType = types.NodeMethod
                recv := extractTypeName(x.Recv.List[0].Type)
                if recv != "" {
                    nodeID = packageName + "." + recv + "." + x.Name.Name
                    meta["receiver"] = recv
                }
            }
            nodes = append(nodes, &types.Node{
                ID: nodeID, Type: nodeType, Name: x.Name.Name,
                Package: packageName, FilePath: filePath, Metadata: meta,
            })
            edges = append(edges, &types.Edge{From: fileNodeID, To: nodeID, Type: types.EdgeContains})

        case *ast.GenDecl:
            for _, spec := range x.Specs {
                ts, ok := spec.(*ast.TypeSpec)
                if !ok { continue }
                nodeID := packageName + "." + ts.Name.Name
                nt := types.NodeStruct
                meta := map[string]string{}
                switch ts.Type.(type) {
                case *ast.StructType:    nt = types.NodeStruct; meta["kind"] = "struct"
                case *ast.InterfaceType: nt = types.NodeInterface; meta["kind"] = "interface"
                default: continue
                }
                nodes = append(nodes, &types.Node{
                    ID: nodeID, Type: nt, Name: ts.Name.Name,
                    Package: packageName, FilePath: filePath, Metadata: meta,
                })
                edges = append(edges, &types.Edge{From: fileNodeID, To: nodeID, Type: types.EdgeContains})
            }
        }
        return true
    })

    for _, imp := range f.Imports {
        if imp.Path == nil { continue }
        importPath := imp.Path.Value[1 : len(imp.Path.Value)-1]
        importID := "pkg:" + importPath
        nodes = append(nodes, &types.Node{ID: importID, Type: types.NodePackage, Name: importPath, Package: importPath})
        edges = append(edges, &types.Edge{From: fileNodeID, To: importID, Type: types.EdgeImports})
    }

    return nodes, edges, nil
}

func extractTypeName(expr ast.Expr) string {
    switch t := expr.(type) {
    case *ast.Ident:     return t.Name
    case *ast.StarExpr:  return extractTypeName(t.X)
    }
    return ""
}