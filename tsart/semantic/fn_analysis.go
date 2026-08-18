package main

import (
	fmt "fmt"
	filepath "path/filepath"

	ast "github.com/microsoft/typescript-go/internal/ast"
	compiler "github.com/microsoft/typescript-go/internal/compiler"
	core "github.com/microsoft/typescript-go/internal/core"
	tsoptions "github.com/microsoft/typescript-go/internal/tsoptions"
	osvfs "github.com/microsoft/typescript-go/internal/vfs/osvfs"
	logrus "github.com/sirupsen/logrus"
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

func newSourceFile(fileName string) *ast.SourceFile {
	program, err := newProgramForTrace(fileName)
	if err != nil {
		logrus.Fatal(err)
	}
	sourceFile := program.GetSourceFile(fileName)
	if sourceFile == nil {
		logrus.Fatalf("program did not load %q", fileName)
	}
	if sourceFile.IsBound() {
		logrus.Fatal("source file was bound before the explicit binding step")
	}
	if got := len(sourceFile.Statements.Nodes); got == 0 {
		logrus.Fatal("parser produced no top-level statements")
	}

	// This walks the AST and populates SourceFile.Locals (and any nested symbol tables).
	program.BindSourceFiles()
	if !sourceFile.IsBound() {
		logrus.Fatal("source file was not bound")
	}

	return sourceFile
}

type SourceFileDecoratorImpl struct {
	pSourceFile *ast.SourceFile
}

func NewSourceFileDecoratorImpl(fileName string) *SourceFileDecoratorImpl {
	pSourceFile := newSourceFile(fileName)
	return &SourceFileDecoratorImpl{
		pSourceFile: pSourceFile,
	}
}

func (sf *SourceFileDecoratorImpl) extractSubtrees(functionName string) error {
	symbol := sf.pSourceFile.Locals[functionName]
	if symbol == nil {
		logrus.Fatal("binder did not declare simpleFunction in the source file locals")
	}
	for _, fnDec := range symbol.Declarations {
		if fnDec.Kind != ast.KindFunctionDeclaration {
			return fmt.Errorf("node is not a function declaration")
		}

		for _, local := range getLocalsForFn(fnDec) {
			printKind(local)
		}

		for _, ret := range getLZeroNodesFromFn(fnDec) {
			printKind(ret)
			sf.printNodeSource(ret)
		}
	}
	return nil
}

func printKind(n *ast.Node) {
	logrus.Infof("kind: %s", n.KindString())
}

func getLocalsForFn(functionDeclaration *ast.Node) []*ast.Node {
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

func getLZeroNodesFromFn(fn *ast.Node) []*ast.Node {
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

func (sf *SourceFileDecoratorImpl) printNodeSource(n *ast.Node) {

	text := sf.pSourceFile.Text()
	start, end := n.Pos(), n.End()
	if start < 0 || end < start || end > len(text) {
		logrus.Errorf("%s has no usable source range: [%d:%d]", n.Kind, start, end)
		return
	}

	logrus.Infof("%s:\n%s", n.Kind, text[start:end])
}
