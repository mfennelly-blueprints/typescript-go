package cmd

import (
	"os/exec"
	"testing"

	"github.com/neovim/go-client/nvim"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	EXAMPLE_FILE = "/tmp/example.ts"
)

func getExampleSelection() *nvimSelection {
	return &nvimSelection{
		FileName: EXAMPLE_FILE,
		Text:     "function renamed(): void {\n  renamed()\n}\n",
		Line:     2,
		Column:   3,
	}
}

func newTestNvim(t *testing.T) *nvim.Nvim {
	t.Helper()

	if _, err := exec.LookPath(nvim.BinaryName); err != nil {
		t.Skipf("%s is not installed: %v", nvim.BinaryName, err)
	}

	childProcOpts := nvim.ChildProcessArgs(
		"-u", "NONE", "-n", "-i", "NONE", "--embed", "--headless",
	)

	headlessNvimClient, err := nvim.NewChildProcess(childProcOpts)
	require.NoError(t, err, "starting Neovim")

	t.Cleanup(func() {
		// End the embedded session explicitly; Close then releases the RPC pipes.
		_ = headlessNvimClient.Command("qall!")
		_ = headlessNvimClient.Close()
	})

	return headlessNvimClient
}

func TestSelectASTAtCursorEchoesSelection(t *testing.T) {
	headlessNvimClient := newTestNvim(t)

	// example selection mocks the scenario
	// where  a user has their cursor on  a given
	// file position, and sends a request payload
	// which is routed to selectAstAtCursor
	exampleSelection := getExampleSelection()
	require.NoError(t, selectASTAtCursor(headlessNvimClient, exampleSelection))

	messages, err := headlessNvimClient.Exec("messages", true)
	require.NoError(t, err, "reading Neovim messages")

	const want = `tsart AST: KindIdentifier "renamed" [29,36)`
	assert.Contains(t, messages, want, "echoed message")
}

func TestBufferPositionAtOffset(t *testing.T) {
	text := "const cafe = 1\n\u00e9x\n"
	tests := []struct {
		name       string
		offset     int
		wantRow    int
		wantColumn int
	}{
		{name: "start", offset: 0, wantRow: 0, wantColumn: 0},
		{name: "first line", offset: 10, wantRow: 0, wantColumn: 10},
		{name: "second line", offset: 15, wantRow: 1, wantColumn: 0},
		{name: "multibyte character", offset: 17, wantRow: 1, wantColumn: 2}, // \u00e9 occupies two UTF-8 bytes.
		{name: "end", offset: len(text), wantRow: 2, wantColumn: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			row, column, err := bufferPositionAtOffset(text, test.offset)
			require.NoError(t, err)
			assert.Equal(t, test.wantRow, row)
			assert.Equal(t, test.wantColumn, column)
		})
	}
}

func TestBufferPositionAtOffsetRejectsInvalidOffsets(t *testing.T) {
	_, _, err := bufferPositionAtOffset("x", -1)
	require.Error(t, err, "negative offset should be rejected")

	_, _, err = bufferPositionAtOffset("x", 2)
	require.Error(t, err, "offset beyond the buffer should be rejected")
}

func TestVisualSelectionRange(t *testing.T) {
	selected := &nvimVisualSelection{
		Text:        "const café = 1\nreturn café",
		StartLine:   1,
		StartColumn: 7,
		EndLine:     2,
		EndColumn:   11,
		Mode:        "v",
	}
	startRow, startCol, endRow, endCol, err := visualSelectionRange(selected)
	require.NoError(t, err)
	assert.Equal(t, 0, startRow)
	assert.Equal(t, 6, startCol)
	assert.Equal(t, 1, endRow)
	assert.Equal(t, 12, endCol)
}

func TestVisualSelectionRangeLinewise(t *testing.T) {
	selected := &nvimVisualSelection{Text: "one\ntwo\nthree", StartLine: 1, StartColumn: 1, EndLine: 2, EndColumn: 3, Mode: "V"}
	startRow, startCol, endRow, endCol, err := visualSelectionRange(selected)
	require.NoError(t, err)
	assert.Equal(t, 0, startRow)
	assert.Equal(t, 0, startCol)
	assert.Equal(t, 1, endRow)
	assert.Equal(t, 3, endCol)
}

func TestVisualSelectionRangeReportsOutOfBoundsDetails(t *testing.T) {
	tests := []struct {
		name     string
		selected nvimVisualSelection
		want     string
	}{
		{
			name:     "line beyond buffer",
			selected: nvimVisualSelection{Text: "one", StartLine: 1, StartColumn: 1, EndLine: 2, EndColumn: 1, Mode: "v"},
			want:     "ends at line 2, but the buffer has 1 lines",
		},
		{
			name:     "column beyond line",
			selected: nvimVisualSelection{Text: "one", StartLine: 1, StartColumn: 1, EndLine: 1, EndColumn: 4, Mode: "v"},
			want:     "ends at byte column 4 on line 1, but that line has 3 bytes",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, _, _, err := visualSelectionRange(&test.selected)
			require.Error(t, err)
			assert.ErrorContains(t, err, test.want)
		})
	}
}
