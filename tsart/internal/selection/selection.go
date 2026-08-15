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

// At parses text as fileName and returns the AST token at the supplied
// one-based line and zero-based byte column. text is supplied by Neovim so
// unsaved buffer changes are included in the AST.
func At(fileName, text string, line, column int) (Result, error) {
	if line < 1 || column < 0 {
		return Result{}, fmt.Errorf("invalid cursor position %d:%d", line, column)
	}

	position, err := offsetForPosition(text, line, column)
	if err != nil {
		return Result{}, err
	}
	absFileName, err := filepath.Abs(fileName)
	if err != nil {
		return Result{}, err
	}
	absFileName = filepath.ToSlash(absFileName)
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: absFileName,
		Path:     tspath.ToPath(absFileName, "", true),
	}, text, core.EnsureScriptKindFromFileName(absFileName))
	node := astnav.GetTouchingToken(file, position)
	if node == nil {
		return Result{}, fmt.Errorf("no AST node at %d:%d", line, column)
	}

	start := scanner.GetTokenPosOfNode(node, file, false)
	return Result{
		Kind:  node.KindString(),
		Start: start,
		End:   node.End(),
		Text:  text[start:node.End()],
	}, nil
}

func offsetForPosition(text string, line, column int) (int, error) {
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
