package cmd

import (
	"fmt"
	"strings"

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
