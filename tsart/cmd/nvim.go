package cmd

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/tsart/internal/selection"
	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
	"github.com/spf13/cobra"
)

type nvimSelection struct {
	FileName string `eval:"expand('%:p')"`
	Text     string `eval:"join(getline(1, '$'), \"\\n\")"`
	Line     int    `eval:"line('.')"`
	Column   int    `eval:"col('.') - 1"`
}

// nvimVisualSelection is evaluated when :TsartAnnotateSelection is invoked
// from Visual mode. Neovim records the inclusive endpoints in the '< and '>
// marks, including after Visual mode has been left to run the command.
type nvimVisualSelection struct {
	Text        string `eval:"join(getline(1, '$'), \"\\n\")"`
	StartLine   int    `eval:"line(\"'<\")"`
	StartColumn int    `eval:"col(\"'<\")"`
	EndLine     int    `eval:"line(\"'>\")"`
	EndColumn   int    `eval:"col(\"'>\")"`
	Mode        string `eval:"visualmode()"`
}

var nvimCmd = &cobra.Command{
	Use:                "nvim",
	Short:              "Run tsart as a Neovim remote-plugin host",
	Args:               cobra.NoArgs,
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	Run:                func(_ *cobra.Command, _ []string) { RunNvimPlugin() },
}

// RunNvimPlugin starts the remote-plugin host. It is exported so main can
// hand its arguments straight to the go-client manifest parser.
func RunNvimPlugin() {
	plugin.Main(registerNvimHandlers)
}

func registerNvimHandlers(p *plugin.Plugin) error {
	p.HandleCommand(&plugin.CommandOptions{Name: "TsartSelectAST", Eval: "*"}, selectASTAtCursor)
	p.HandleCommand(&plugin.CommandOptions{Name: "TsartHighlightAST", Eval: "*"}, highlightASTAtCursor)
	p.HandleCommand(&plugin.CommandOptions{Name: "TsartClearASTHighlight"}, clearASTHighlight)
	// A Visual-mode ':' command is prefixed with the '<,'> line range.  Accept
	// it even though character-accurate bounds come from the Visual marks below.
	p.HandleCommand(&plugin.CommandOptions{Name: "TsartAnnotateSelection", NArgs: "+", Range: ".", Eval: "*"}, annotateVisualSelection)
	return nil
}

func selectASTAtCursor(v *nvim.Nvim, selected *nvimSelection) error {
	result, err := selection.At(selected.FileName, selected.Text, selected.Line, selected.Column)
	if err != nil {
		return err
	}
	message := fmt.Sprintf("tsart AST: %s %q [%d,%d)", result.Kind, result.Text, result.Start, result.End)
	return v.Echo([]nvim.TextChunk{{Text: message}}, true, map[string]interface{}{})
}

// highlightASTAtCursor highlights the smallest AST token containing the
// cursor. The highlight is kept in its own namespace, so each invocation
// replaces the previous tsart highlight without disturbing other plugins.
func highlightASTAtCursor(v *nvim.Nvim, selected *nvimSelection) error {
	result, err := selection.At(selected.FileName, selected.Text, selected.Line, selected.Column)
	if err != nil {
		return err
	}

	startRow, startCol, err := bufferPositionAtOffset(selected.Text, result.Start)
	if err != nil {
		return err
	}
	endRow, endCol, err := bufferPositionAtOffset(selected.Text, result.End)
	if err != nil {
		return err
	}
	buffer, err := v.CurrentBuffer()
	if err != nil {
		return err
	}
	namespace, err := v.CreateNamespace("tsart-ast-highlight")
	if err != nil {
		return err
	}
	if err := v.ClearBufferNamespace(buffer, namespace, 0, -1); err != nil {
		return err
	}
	_, err = v.SetBufferExtmark(buffer, namespace, startRow, startCol, map[string]interface{}{
		"end_row":  endRow,
		"end_col":  endCol,
		"hl_group": "Visual",
		"priority": 200,
	})
	if err != nil {
		return err
	}
	message := fmt.Sprintf("tsart highlighted: %s %q", result.Kind, result.Text)
	return v.Echo([]nvim.TextChunk{{Text: message}}, true, map[string]interface{}{})
}

func clearASTHighlight(v *nvim.Nvim) error {
	buffer, err := v.CurrentBuffer()
	if err != nil {
		return err
	}
	namespace, err := v.CreateNamespace("tsart-ast-highlight")
	if err != nil {
		return err
	}
	return v.ClearBufferNamespace(buffer, namespace, 0, -1)
}

// annotateVisualSelection highlights the active Visual selection and displays
// its comment at the end of the selection. The extmark tracks edits to the
// buffer, so the highlight and comment move with the annotated text.
func annotateVisualSelection(v *nvim.Nvim, args []string, _ [2]int, selected *nvimVisualSelection) error {
	if selected.Mode == "\x16" {
		return fmt.Errorf("blockwise Visual selections are not supported")
	}
	startRow, startCol, endRow, endCol, err := visualSelectionRange(selected)
	if err != nil {
		return err
	}
	buffer, err := v.CurrentBuffer()
	if err != nil {
		return err
	}
	namespace, err := v.CreateNamespace("tsart-comments")
	if err != nil {
		return err
	}
	comment := strings.Join(args, " ")
	_, err = v.SetBufferExtmark(buffer, namespace, startRow, startCol, map[string]interface{}{
		"end_row":  endRow,
		"end_col":  endCol,
		"hl_group": "IncSearch",
		"priority": 200,
		"virt_text": []interface{}{
			[]interface{}{fmt.Sprintf("  💬 %s", comment), "Comment"},
		},
		"virt_text_pos": "eol",
	})
	if err != nil {
		return err
	}
	return v.Echo([]nvim.TextChunk{{Text: "tsart annotation added"}}, true, map[string]interface{}{})
}

// visualSelectionRange converts Neovim's one-based, inclusive Visual marks
// into the zero-based, end-exclusive range expected by an extmark.
func visualSelectionRange(selected *nvimVisualSelection) (startRow, startCol, endRow, endCol int, err error) {
	if selected.StartLine < 1 || selected.EndLine < selected.StartLine || selected.StartColumn < 1 || selected.EndColumn < 1 {
		return 0, 0, 0, 0, fmt.Errorf("no Visual selection is available")
	}
	lines := strings.Split(selected.Text, "\n")
	if selected.EndLine > len(lines) {
		return 0, 0, 0, 0, fmt.Errorf("Visual selection is outside the buffer")
	}
	startRow, startCol = selected.StartLine-1, selected.StartColumn-1
	endRow = selected.EndLine - 1
	if selected.Mode == "V" {
		return startRow, 0, endRow, len(lines[endRow]), nil
	}
	endStart := selected.EndColumn - 1
	if endStart >= len(lines[endRow]) {
		return 0, 0, 0, 0, fmt.Errorf("Visual selection is outside the buffer")
	}
	_, width := utf8.DecodeRuneInString(lines[endRow][endStart:])
	if width == 0 || (width == 1 && lines[endRow][endStart] >= utf8.RuneSelf) {
		return 0, 0, 0, 0, fmt.Errorf("Visual selection ends inside an invalid UTF-8 character")
	}
	return startRow, startCol, endRow, endStart + width, nil
}

// bufferPositionAtOffset converts a zero-based byte offset to Neovim's
// zero-based row and byte column coordinates.
func bufferPositionAtOffset(text string, offset int) (row, column int, err error) {
	if offset < 0 || offset > len(text) {
		return 0, 0, fmt.Errorf("offset %d is outside the buffer", offset)
	}
	before := text[:offset]
	row = strings.Count(before, "\n")
	if lastNewline := strings.LastIndex(before, "\n"); lastNewline >= 0 {
		return row, len(before) - lastNewline - 1, nil
	}
	return row, len(before), nil
}

func init() {
	rootCmd.AddCommand(nvimCmd)
}
