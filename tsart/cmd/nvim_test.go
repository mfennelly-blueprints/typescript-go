package cmd

import (
	"strings"
	"testing"
)

func TestBufferPositionAtOffset(t *testing.T) {
	text := "const cafe = 1\n\u00e9x\n"
	tests := []struct {
		offset     int
		wantRow    int
		wantColumn int
	}{
		{offset: 0, wantRow: 0, wantColumn: 0},
		{offset: 10, wantRow: 0, wantColumn: 10},
		{offset: 15, wantRow: 1, wantColumn: 0},
		{offset: 17, wantRow: 1, wantColumn: 2}, // \u00e9 occupies two UTF-8 bytes.
		{offset: len(text), wantRow: 2, wantColumn: 0},
	}

	for _, test := range tests {
		row, column, err := bufferPositionAtOffset(text, test.offset)
		if err != nil {
			t.Fatalf("bufferPositionAtOffset(%d) returned error: %v", test.offset, err)
		}
		if row != test.wantRow || column != test.wantColumn {
			t.Errorf("bufferPositionAtOffset(%d) = (%d, %d), want (%d, %d)", test.offset, row, column, test.wantRow, test.wantColumn)
		}
	}
}

func TestBufferPositionAtOffsetRejectsInvalidOffsets(t *testing.T) {
	if _, _, err := bufferPositionAtOffset("x", -1); err == nil {
		t.Fatal("negative offset was accepted")
	}
	if _, _, err := bufferPositionAtOffset("x", 2); err == nil {
		t.Fatal("offset beyond the buffer was accepted")
	}
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
	if err != nil {
		t.Fatalf("visualSelectionRange returned error: %v", err)
	}
	if startRow != 0 || startCol != 6 || endRow != 1 || endCol != 12 {
		t.Errorf("visualSelectionRange = (%d, %d, %d, %d), want (0, 6, 1, 12)", startRow, startCol, endRow, endCol)
	}
}

func TestVisualSelectionRangeLinewise(t *testing.T) {
	selected := &nvimVisualSelection{Text: "one\ntwo\nthree", StartLine: 1, StartColumn: 1, EndLine: 2, EndColumn: 3, Mode: "V"}
	startRow, startCol, endRow, endCol, err := visualSelectionRange(selected)
	if err != nil {
		t.Fatalf("visualSelectionRange returned error: %v", err)
	}
	if startRow != 0 || startCol != 0 || endRow != 1 || endCol != 3 {
		t.Errorf("visualSelectionRange = (%d, %d, %d, %d), want (0, 0, 1, 3)", startRow, startCol, endRow, endCol)
	}
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
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("visualSelectionRange error = %v, want text %q", err, test.want)
			}
		})
	}
}
