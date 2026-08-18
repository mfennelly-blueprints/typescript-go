package main

import (
	"path/filepath"
	"testing"

	ast "github.com/microsoft/typescript-go/internal/ast"
	compiler "github.com/microsoft/typescript-go/internal/compiler"
	core "github.com/microsoft/typescript-go/internal/core"
	tsoptions "github.com/microsoft/typescript-go/internal/tsoptions"
	osvfs "github.com/microsoft/typescript-go/internal/vfs/osvfs"
)

const FILE_NAME = "/Users/mikeyuseblueprintsai.appleaccount.com/projects/cc-services/hoplite/apps/api/src/db-app.ts"
const FUNCTION_NAME = "createDbBackedApp"

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
	extractSubtrees(t, sourceFile, symbol)
}

func extractSubtrees(t *testing.T, sourceFile *ast.SourceFile, symbol *ast.Symbol) {
	for _, declaration := range symbol.Declarations {
		if declaration == nil {
			continue
		}
		t.Log("\n\n\n")
		t.Log("locals.........")
		for _, local := range getLocals(declaration) {
			printKind(t, local)
		}

		t.Log("\n\n\n")
		t.Log("return.........")
		for _, ret := range getLZeroReturnsFromFunction(declaration) {
			printKind(t, ret)
		}
	}
}

func printKind(t *testing.T, n *ast.Node) {
	t.Logf("kind: %s", n.KindString())
}

func getLocals(functionDeclaration *ast.Node) []*ast.Node {
	// loop over all node-local declarations, printing the
	// value declaration for each
	var returns []*ast.Node
	for _, local := range functionDeclaration.Locals() {
		if local.ValueDeclaration != nil {
			returns = append(returns, local.ValueDeclaration)
		}
	}
	return returns
}

func getLZeroReturnsFromFunction(fn *ast.Node) []*ast.Node {
	body := fn.Body()
	if body == nil {
		return nil
	}

	var returns []*ast.Node
	for _, statement := range body.Statements() {
		if statement.Kind == ast.KindReturnStatement {
			returns = append(returns, statement)
		}
	}
	return returns
}

func printNodeSource(t *testing.T, sourceFile *ast.SourceFile, n *ast.Node) {
	t.Helper()

	text := sourceFile.Text()
	start, end := n.Pos(), n.End()
	if start < 0 || end < start || end > len(text) {
		t.Logf("%s has no usable source range: [%d:%d]", n.Kind, start, end)
		return
	}

	t.Logf("%s:\n%s", n.Kind, text[start:end])
}
