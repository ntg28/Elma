package main

import (
    "os"
    "fmt"
    "errors"
    "go/types"
    "go/token"
    "elma/gen"
    "golang.org/x/tools/go/packages"
)

type ElmaImporter struct {
    Root string
    Pkgs []*packages.Package
}

const MAX_FILE_SIZE = 1024 * 8

func resolvePackage(pkg *packages.Package, imp ElmaImporter) (*types.Package, error) {
    fset := token.NewFileSet()
    tcfg := types.Config{
        Importer: imp,
        IgnoreFuncBodies: true,
    }

    for _, file := range pkg.GoFiles {
        fset.AddFile(file, fset.Base(), MAX_FILE_SIZE)
    }

    return tcfg.Check(pkg.ID, fset, pkg.Syntax, nil)
}

func (imp ElmaImporter) Import(path string) (*types.Package, error) {
    for _, pkg := range imp.Pkgs {
        if pkg.ID == imp.Root + "/" + path {
            return resolvePackage(pkg, imp)
        }
    }
    return nil, errors.New("package not found")
}

func main() {
    if len(os.Args) < 2 {
        fmt.Fprintf(os.Stderr, "ERROR: missing path argument\n")
        os.Exit(1)
    }

    cfg := &packages.Config {
        Mode: packages.NeedFiles | packages.NeedSyntax,
    }

    path := os.Args[1]
    path_pattern := path + "/..."

    src, err := packages.Load(cfg, path_pattern)
    if err != nil {
        fmt.Fprintf(os.Stderr, "%v\n", err)
        os.Exit(1)
    }

    lib, err := packages.Load(cfg, "./lib/...")
    if err != nil {
        fmt.Fprintf(os.Stderr, "%v\n", err)
        os.Exit(1)
    }

    all := []*packages.Package{}

    for _, pkg := range lib { all = append(all, pkg) }
    for _, pkg := range src { all = append(all, pkg) }

    fset := token.NewFileSet()
    tcfg := types.Config{
        Importer: ElmaImporter{"elma", all},
    }

    fset.AddFile(path + "/main.go", fset.Base(), MAX_FILE_SIZE)

    _, e := tcfg.Check("src", fset, src[0].Syntax, nil)
    if e != nil {
        fmt.Fprintf(os.Stderr, "%v\n", e)
        os.Exit(1)
    }

    g := gen.Gen{
        Pkgs: all,
        Binds: map[string]string{},
    }

    out := g.GenPkg(src[0])

    os.Remove(path + "/main.js")

    outfile, err := os.OpenFile(path + "/main.js", os.O_RDWR|os.O_CREATE, 0644)
    if err != nil {
        fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
        os.Exit(1)
    }

    outfile.WriteString(out)
    outfile.Close()
}
