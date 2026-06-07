package ast

import (
    "io/fs"
    "path/filepath"
    "strings"
)

type WalkConfig struct {
    IncludeTests bool     // include _test.go files
    SkipDirs     []string // additional dirs to skip
}

var defaultSkipDirs = map[string]bool{
	"vendor":       true,
	".git":         true,
	"node_modules": true,
	".idea":        true,
	".vscode":      true,
}

func FindGoFiles(rootDir string, cfg WalkConfig) ([]string, error) {
    skipDirs := make(map[string]bool)
    for k, v := range defaultSkipDirs { skipDirs[k] = v }
    for _, d := range cfg.SkipDirs { skipDirs[d] = true }

    var files []string
    err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
        if err != nil { return nil }
        if d.IsDir() {
            if skipDirs[d.Name()] || strings.HasPrefix(d.Name(), ".") {
                return filepath.SkipDir
            }
            return nil
        }
        if !strings.HasSuffix(path, ".go") { return nil }
        if !cfg.IncludeTests && strings.HasSuffix(path, "_test.go") { return nil }
        files = append(files, path)
        return nil
    })
    return files, err
}