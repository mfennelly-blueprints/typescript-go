package main

import (
	"path/filepath"
	"testing"
)

const FILE_NAME = "/Users/mikeyuseblueprintsai.appleaccount.com/projects/cc-services/hoplite/apps/api/src/db-app.ts"
const FUNCTION_SYMBOL_NAME = "singleFlightTask"

// TestTraceExample is intentionally organized as a debugging aid. Run it with:
//
//	cd tsart && go test -v -run TestTraceExample
//
// Useful breakpoints, in order:
//   - compiler.NewProgram                 (file loading starts)
//   - compilerHost.GetSourceFile          (the file is read)
//   - parser.ParseSourceFile              (the AST is built)
//   - program.BindSourceFiles             (binding starts)
//   - binder.BindSourceFile / Binder.bind (symbols are declared)
func TestTraceExample(t *testing.T) {
	fileName := FILE_NAME
	fileName, err := filepath.Abs(fileName)
	if err != nil {
		t.Fatal(err)
	}
	fileName = filepath.ToSlash(fileName)
	program, err := newProgramForTrace(fileName)
	if err != nil {
		t.Fatal(err)
	}
	sourceFile := program.GetSourceFile(fileName)
	if sourceFile == nil {
		t.Fatalf("program did not load %q", fileName)
	}
	if sourceFile.IsBound() {
		t.Fatal("source file was bound before the explicit binding step")
	}
	if got := len(sourceFile.Statements.Nodes); got == 0 {
		t.Fatal("parser produced no top-level statements")
	}

	// This walks the AST and populates SourceFile.Locals (and any nested symbol tables).
	program.BindSourceFiles()
	if !sourceFile.IsBound() {
		t.Fatal("source file was not bound")
	}
	symbol := sourceFile.Locals["singleFlightTask"]
	if symbol == nil {
		t.Fatal("binder did not declare simpleFunction in the source file locals")
	}
}

func TestExtractSubtrees(t *testing.T) {
	fileName := FILE_NAME
	fileName, err := filepath.Abs(fileName)
	if err != nil {
		t.Fatal(err)
	}
	fileName = filepath.ToSlash(fileName)
	sourceFile := NewSourceFileDecoratorImpl(fileName)
	sourceFile.extractSubtrees("singleFlightTask")
	sourceFile.extractSubtrees("createDbBackedApp")
}
