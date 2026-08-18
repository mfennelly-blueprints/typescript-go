package main

import (
	ast "github.com/microsoft/typescript-go/internal/ast"
	compiler "github.com/microsoft/typescript-go/internal/compiler"
	core "github.com/microsoft/typescript-go/internal/core"
	tsoptions "github.com/microsoft/typescript-go/internal/tsoptions"
	osvfs "github.com/microsoft/typescript-go/internal/vfs/osvfs"
	"github.com/sirupsen/logrus"
	"path/filepath"

	nvimGoClient "github.com/neovim/go-client/nvim"
)

var clipboard [][]byte

type FileDecorator struct {
	compilerProgram *compiler.Program
}

func NewFileDecorator(fileName string, compilerAst *ast.Node) *FileDecorator {
	// get the fileName
	fileName = filepath.ToSlash(fileName)
	compilerProgram, err := newProgramFromFilePath(fileName)
	if err != nil {
		logrus.Fatal(err)
	}
	return &FileDecorator{
		compilerProgram: compilerProgram,
	}
}

func newProgramFromFilePath(fileName string) (*compiler.Program, error) {
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

func newAstFromSymbol(symbolName string, fileName string) {
	program, err := newProgramFromFilePath(fileName)
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
}

func getSymbolFromAst(symbolName string, sourceFile *ast.SourceFile) *ast.Symbol {

	if symbol := sourceFile.Locals[symbolName]; symbol == nil {
		return nil
	} else {
		logrus.Infof("AST statements: %d; simpleFunction declarations: %d", len(sourceFile.Statements.Nodes), len(symbol.Declarations))
		return symbol
	}
}

// GoCut: delete the given range and stash it
func cut(v *nvimGoClient.Nvim, r [2]int) error {
	b, err := v.CurrentBuffer()
	if err != nil {
		return err
	}
	start, end := r[0]-1, r[1] // 0-indexed, end-exclusive
	lines, err := v.BufferLines(b, start, end, false)
	if err != nil {
		return err
	}
	clipboard = lines
	return v.SetBufferLines(b, start, end, false, nil)
}

// GoPaste: insert stashed lines below the cursor
func paste(v *nvimGoClient.Nvim) error {
	b, err := v.CurrentBuffer()
	if err != nil {
		return err
	}
	win, err := v.CurrentWindow()
	if err != nil {
		return err
	}
	pos, err := v.WindowCursor(win)
	if err != nil {
		return err
	}
	row := pos[0]
	return v.SetBufferLines(b, row, row, false, clipboard)
}
