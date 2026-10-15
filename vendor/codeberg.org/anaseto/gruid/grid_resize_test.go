package gruid

import "testing"

func TestResizeRefreshesPreservedCells(t *testing.T) {
	app := &App{}
	for _, width := range []int{80, 48, 80} {
		gd := NewGrid(width, 24)
		gd.Fill(Cell{Rune: 'A'})
		frame := app.computeFrame(gd, false)
		if len(frame.Cells) != width*24 {
			t.Fatalf("width %d: got %d cells, want full repaint of %d", width, len(frame.Cells), width*24)
		}
	}
}
