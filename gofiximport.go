package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"io/ioutil"
	"os"
	"path/filepath"
)

var fileFlag = flag.String("file", "", "File to process")
var writeFlag = flag.Bool("w", false, "Write result to source file instead of stdout")
var updateFlag = flag.Bool("u", false, "Force cache update")

func solveUnresolved(f *ast.File, candidates *Candidates,
	imports map[string]string) {
	for _, u := range f.Unresolved {
		knownImport, found := imports[u.Name]
		if found {
			addImport(f, knownImport)
			continue
		}
		pkgs := candidates.Pkgs[u.Name]
		if len(pkgs) == 1 {
			for v := range pkgs {
				addImport(f, v)
			}
		}
	}
}

func removeUnused(f *ast.File) {
	for _, i := range f.Imports {
		impath := importPath(i)
		used := usesImport(f, impath)
		if !used {
			deleteImport(f, impath)
		}
	}
}

func solveImports(candidates *Candidates, file string) ([]byte, error) {
	pkgPath := filepath.Dir(file)
	imports, err := parsePackageImports(pkgPath)
	if err != nil {
		return nil, err
	}
	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	solveUnresolved(f, candidates, imports)
	removeUnused(f)
	return gofmtFile(f)
}

func solveInplace(candidates *Candidates, file string) error {
	r, err := solveImports(candidates, file)
	if err != nil {
		return err
	}
	ioutil.WriteFile(file, r, 0)
	return nil
}

func main() {
	flag.Parse()
	if *fileFlag == "" {
		flag.Usage()
		os.Exit(1)
	}
	candidates := loadCandidates(*updateFlag)
	r, err := solveImports(candidates, *fileFlag)
	if err != nil {
		fmt.Print(err)
        os.Exit(1)
	}
	if *writeFlag {
		ioutil.WriteFile(*fileFlag, r, 0)
	} else {
		fmt.Print(string(r))
	}
}
