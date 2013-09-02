package main

import (
	"strings"
	"testing"
)

func checkSolve(file string, expected []string, unexpected []string, cacheUpdate bool, t *testing.T) {
	candidates := loadCandidates(cacheUpdate)
	r, err := solveImports(candidates, file)
	if err != nil {
		t.Fatal(err)
	}
	for _, exp := range expected {
		if !strings.Contains(string(r), exp) {
			t.Fatalf("Expected %s in result, got %s", exp, r)
		}
	}
	for _, unexp := range unexpected {
		if strings.Contains(string(r), unexp) {
			t.Fatalf("Unexpected %s in result, got %s", unexp, r)
		}
	}
}

func TestComplexFile(t *testing.T) {
	expected := []string{
		"bytes",
		"fmt",
		"go/ast",
		"go/format",
		"go/token",
		"os",
		"reflect",
		"strconv",
	}
	checkSolve("tests/fix_test.go", expected, []string{}, true, t)
}

func TestSolveAddImports(t *testing.T) {
	expected := []string{
		"\"fmt\"",
		"\"go/ast\"",
		"A comment",
	}
	checkSolve("tests/add_imports.go", expected, []string{}, false, t)
}

func TestRemoveUseless(t *testing.T) {
	unexpected := []string{"\"os\""}
	checkSolve("tests/remove_imports.go", []string{}, unexpected, false, t)
}
