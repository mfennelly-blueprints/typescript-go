package selection

import "testing"

func TestAtUsesTheLiveBufferText(t *testing.T) {
	text := "function renamed(): void {\n  renamed()\n}\n"

	got, err := At("example.ts", text, 2, 3)
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != "KindIdentifier" || got.Text != "renamed" {
		t.Fatalf("At() = %#v, want the renamed identifier", got)
	}
	if got.Start != 29 || got.End != 36 {
		t.Fatalf("At() range = %d:%d, want 29:36", got.Start, got.End)
	}
}

func TestAtRejectsPositionsOutsideTheBuffer(t *testing.T) {
	if _, err := At("example.ts", "const x = 1\n", 3, 0); err == nil {
		t.Fatal("At() succeeded for a line outside the buffer")
	}
}
