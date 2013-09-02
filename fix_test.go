package main

import (
	"go/parser"
	"strings"
	"testing"
)

func TestAddImport(t *testing.T) {
	src := `
    package p

    // Toto
    const c = 1.0
    var X = f(3.14)*2 + c
    `
	f, _ := parser.ParseFile(fset, "src.go", src, parser.ParseComments)

	addImport(f, "fmt")
	addImport(f, "os")

	r, _ := gofmtFile(f)
	if !strings.Contains(string(r), "\"fmt\"") {
		t.Fatalf("Expected fmt import to be present, got\n%s", r)
	}
	if !strings.Contains(string(r), "\"os\"") {
		t.Fatalf("Expected os import to be present, got\n%s", r)
	}
	if !strings.Contains(string(r), "Toto") {
		t.Fatalf("Expected comment to be present, got\n%s", r)
	}
}

func TestHijackEmptyImportGroup(t *testing.T) {
	src := `
    package p

    import ()

    // Toto
    const c = 1.0
    var X = f(3.14)*2 + c
    `
	f, _ := parser.ParseFile(fset, "src.go", src, parser.ParseComments)

	addImport(f, "fmt")

	r, _ := gofmtFile(f)
	if strings.Count(string(r), "import") > 1 {
		t.Fatalf("Expected only one import group, got\n%s", r)
	}
}
