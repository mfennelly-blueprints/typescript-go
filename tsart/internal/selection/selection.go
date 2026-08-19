// Package selection maps editor positions to nodes in a TypeScript AST.
package selection

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/microsoft/typescript-go/internal/ast"
	"github.com/microsoft/typescript-go/internal/astnav"
	"github.com/microsoft/typescript-go/internal/core"
	"github.com/microsoft/typescript-go/internal/parser"
	"github.com/microsoft/typescript-go/internal/scanner"
	"github.com/microsoft/typescript-go/internal/tspath"
)

// Result describes the smallest AST token that contains an editor position.
// Offsets are zero-based byte offsets, matching Neovim's cursor columns.
type Result struct {
	Kind  string
	Start int
	End   int
	Text  string
}

func parseAstForSourceFile(fileName string, sourceText string) P {
	absFileName, err := filepath.Abs(fileName)
	if err != nil {
		return Result{}, err
	}
	absFileName = filepath.ToSlash(absFileName)

	// parse the source file with the tsgo
	// compiler
	sourceFileParsingOptions := ast.SourceFileParseOptions{
		FileName: absFileName,
		Path:     tspath.ToPath(absFileName, "", true),
	}
	core_ScriptKind := core.EnsureScriptKindFromFileName(absFileName)
	ast_SourceFile := parser.ParseSourceFile(sourceFileParsingOptions, sourceText, core_ScriptKind)

	return ast_SourceFile
}

// AstNodeAtPosition parses text as fileName and returns the AST token at the supplied
// one-based line and zero-based byte column. sourceFileUtf8Contents is supplied by Neovim so
// unsaved buffer changes are included in the AST.
func AstNodeAtPosition(fileName string, sourceText string, line int, column int) (Result, error) {
	if line < 1 || column < 0 {
		return Result{}, fmt.Errorf("invalid cursor position %d:%d", line, column)
	}

	ast_SourceFile := parseAstForSourceFile(fileName, sourceText)

	// get the position in the linearized file
	position, err := offsetForPosition(sourceText, line, column)
	if err != nil {
		return Result{}, err
	}
	node := astnav.GetTouchingToken(ast_SourceFile, position)
	if node == nil {
		return Result{}, fmt.Errorf("no AST node at %d:%d", line, column)
	}

	start := scanner.GetTokenPosOfNode(node, ast_SourceFile, false)
	return Result{
		Kind:  node.KindString(),
		Start: start,
		End:   node.End(),
		Text:  sourceText[start:node.End()],
	}, nil
}

// offsetForPosition calculates the offset
// from the position 0, when the 2 dimensional
// file matrix is linearized
func offsetForPosition(text string, line int, column int) (int, error) {
	// take the raw text of the input
	// and convert it into a stream of lines
	// based on <CR> tokens
	lines := strings.Split(text, "\n")
	if line > len(lines) {
		return 0, fmt.Errorf("line %d is outside the buffer", line)
	}
	if column > len(lines[line-1]) {
		return 0, fmt.Errorf("column %d is outside line %d", column, line)
	}

	offset := column
	for i := 0; i < line-1; i++ {
		offset += len(lines[i]) + 1
	}
	return offset, nil
}
