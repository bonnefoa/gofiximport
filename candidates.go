package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var fset = token.NewFileSet()
var gofiximportDir = os.ExpandEnv("$HOME/.gofiximport")
var cachePath = filepath.Join(gofiximportDir, "cache")

var gopaths []string

type pathContext struct {
	candidates *Candidates
	visited    map[string]bool
}

// Candidates are elligible package
type Candidates struct {
	Pkgs       map[string]map[string]bool
	Ts         int64
	updateTime time.Time
}

func init() {
	gopaths = strings.Split(os.Getenv("GOPATH"), ":")
	r := runtime.GOROOT()
	if r != "" {
		gopaths = append(gopaths, filepath.Join(r, "src", "pkg"))
	}
}

func filterGoSources(f os.FileInfo) bool {
	return strings.HasSuffix(f.Name(), ".go") &&
		!strings.HasSuffix(f.Name(), "_test.go")
}

func fullPackageFromPath(path string) string {
	for _, gopath := range gopaths {
		path = strings.Replace(path, gopath, "", -1)
	}
	path = strings.Replace(path, "src/", "", -1)
	path = strings.Replace(path, "pkg/", "", -1)
	path = strings.Trim(path, "/")
	return path
}

func packageShortName(packageName string) string {
	splitted := strings.Split(packageName, "/")
	return splitted[len(splitted)-1]
}

func parsePackageImports(packagePath string) (imports map[string]string, err error) {
	imports = map[string]string{}
	p, err := parsePackage(packagePath, parser.ImportsOnly)
	if err != nil {
		return
	}
	for _, f := range p.Files {
		for _, imp := range f.Imports {
			importPath := importPath(imp)
			shortName := packageShortName(importPath)
			imports[shortName] = importPath
			if imp.Name != nil {
				imports[imp.Name.Name] = importPath
			}
		}
	}
	return
}

func parsePackage(packagePath string,
	mode parser.Mode) (pkg *ast.Package, err error) {
	pkgs, err := parser.ParseDir(fset, packagePath, filterGoSources, mode)
	if err != nil {
		return
	}
	for _, v := range pkgs {
		pkg = v
		ast.PackageExports(v)
		return
	}
	return
}

func processPackage(path string, candidates *Candidates) {
	f, err := parsePackage(path, 0)
	if err != nil {
		if *verboseFlag {
			fmt.Printf("Error on package parse, ignoring file. %q\n", err)
		}
		return
	}
	if f == nil {
		return
	}
	canonicalPkg := filepath.Base(path)
	fullPackage := fullPackageFromPath(path)
	if c := candidates.Pkgs[canonicalPkg]; c == nil {
		candidates.Pkgs[canonicalPkg] = map[string]bool{fullPackage: true}
	} else {
		candidates.Pkgs[canonicalPkg][fullPackage] = true
	}
}

func walkFun(ctx interface{}, path string, info os.FileInfo, err error) error {
	context := ctx.(*pathContext)
	candidates := context.candidates
	visited := context.visited
	if info.ModTime().Before(candidates.updateTime) {
		return filepath.SkipDir
	}
	if _, found := visited[path]; found {
		return filepath.SkipDir
	}
	visited[path] = true
	isSym := info.Mode()&os.ModeSymlink > 0
	if !info.IsDir() && !isSym {
		return nil
	}
	basePath := filepath.Base(path)
	if strings.HasPrefix(basePath, ".") {
		return filepath.SkipDir
	}
	// Resolve symlink
	if isSym {
		symPath, err := os.Readlink(path)
		if err != nil {
			return nil
		}
		if !filepath.IsAbs(symPath) {
			symPath = filepath.Join(path, symPath)
		}
		_, err = os.Open(symPath)
		if err != nil {
			return nil
		}
		return PathWalk(context, symPath, walkFun)
	}
	processPackage(path, candidates)
	return nil
}

func updateCandidates() *Candidates {
	visited := map[string]bool{}
	candidates := &Candidates{Pkgs: map[string]map[string]bool{}, Ts: 0}
	context := &pathContext{candidates, visited}

	for _, path := range gopaths {
		PathWalk(context, path, walkFun)
	}
	saveCandidates(candidates)
	return candidates
}

func loadCandidates(cacheUpdate bool) *Candidates {
	f, err := os.Open(cachePath)
	if err != nil || cacheUpdate {
		return updateCandidates()
	}
	candidates := &Candidates{}
	decoder := json.NewDecoder(f)
	decoder.Decode(candidates)
	candidates.updateTime = time.Unix(candidates.Ts, 0)

	// Checking most recent modification
	t := mostRecentModification(gopaths)
	if t.After(candidates.updateTime) {
		return updateCandidates()
	}

	return candidates
}

func saveCandidates(candidates *Candidates) {
	os.Mkdir(gofiximportDir, 0777)
	f, err := os.Create(cachePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	encoder := json.NewEncoder(f)
	updateTime := time.Now()
	candidates.updateTime = updateTime
	candidates.Ts = updateTime.Unix()
	encoder.Encode(candidates)
}
