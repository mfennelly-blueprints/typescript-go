package main

import (
	"path/filepath"
	"testing"

	compiler "github.com/microsoft/typescript-go/internal/compiler"
	core "github.com/microsoft/typescript-go/internal/core"
	tsoptions "github.com/microsoft/typescript-go/internal/tsoptions"
	osvfs "github.com/microsoft/typescript-go/internal/vfs/osvfs"
)

func newProgramForTrace(fileName string) (*compiler.Program, error) {
	fileName, err := filepath.Abs(fileName)
	if err != nil {
		return nil, err
	}
	fileName = filepath.ToSlash(fileName)
	compilerOptions := &core.CompilerOptions{NoLib: core.TSTrue}
	config := &tsoptions.ParsedCommandLine{
		ParsedConfig: &core.ParsedOptions{
			FileNames:       []string{fileName},
			CompilerOptions: compilerOptions,
		},
	}
	host := compiler.NewCompilerHost(filepath.Dir(fileName), osvfs.FS(), "", nil, nil)

	// NewProgram loads example.ts and constructs its SourceFile AST. It does not bind it yet.
	return compiler.NewProgram(compiler.ProgramOptions{Config: config, Host: host}), nil
}

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
	fileName := filepath.Join("test-data", "example.ts")
	fileName, err := filepath.Abs(fileName)
	if err != nil {
		t.Fatal(err)
	}
	fileName = filepath.ToSlash(fileName)
	symbolName := "simpleFunction"
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
	if symbol := sourceFile.Locals[symbolName]; symbol == nil {
		t.Fatal("binder did not declare simpleFunction in the source file locals")
	} else {
		t.Logf("AST statements: %d; simpleFunction declarations: %d", len(sourceFile.Statements.Nodes), len(symbol.Declarations))
	}
}
