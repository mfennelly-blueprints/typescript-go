package cmd

import "testing"

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
