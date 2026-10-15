package ui

import (
	"codeberg.org/anaseto/gruid"
	"testing"
)

func TestReplayVariableGridForwardAndBackward(t *testing.T) {
	rp := &Replay{grid: gruid.NewGrid(80, 24)}
	makeFrame := func(w int, r rune) gruid.Frame {
		frame := gruid.Frame{Width: w, Height: 24}
		for y := 0; y < 24; y++ {
			for x := 0; x < w; x++ {
				frame.Cells = append(frame.Cells, gruid.FrameCell{P: gruid.Point{X: x, Y: y}, Cell: gruid.Cell{Rune: r}})
			}
		}
		return frame
	}
	rp.frames = []gruid.Frame{makeFrame(80, 'a'), makeFrame(48, 'b'), makeFrame(80, 'c')}
	for i, w := range []int{80, 48, 80} {
		rp.fidx = i + 1
		rp.next()
		if rp.grid.Size().X != w || rp.grid.At(gruid.Point{X: w - 1, Y: 23}).Rune != rune('a'+i) {
			t.Fatalf("forward frame %d incorrect", i)
		}
	}
	for i, w := range []int{48, 80} {
		rp.previous()
		if rp.grid.Size().X != w || rp.grid.At(gruid.Point{X: w - 1, Y: 23}).Rune != rune('b'-i) {
			t.Fatalf("backward frame %d incorrect", i)
		}
	}
	rp.fidx = 2
	rp.next()
	if rp.grid.Size().X != 48 || rp.grid.At(gruid.Point{X: 47, Y: 23}).Rune != 'b' {
		t.Fatal("resumed playback incorrect")
	}
}
