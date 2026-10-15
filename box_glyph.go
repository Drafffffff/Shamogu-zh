//go:build js || sdl

package main

import (
	"image"
	"image/draw"

	"codeberg.org/anaseto/gruid"
)

// Borders are geometry, not font glyphs: each segment reaches its cell edge.
// This keeps corners and joins continuous at both text and dungeon cell sizes.
func boxGlyph(gc gruid.Cell, size gruid.Point) image.Image {
	const up, right, down, left = 1, 2, 4, 8
	segments := map[rune]int{
		'─': left | right, '│': up | down,
		'┌': right | down, '┐': left | down,
		'└': right | up, '┘': left | up,
		'├': up | down | right, '┤': up | down | left,
		'┬': left | right | down, '┴': left | right | up,
		'┼': up | down | left | right,
	}[gc.Rune]
	if segments == 0 {
		return nil
	}
	bg, fg := ColorToRGBA(gc.Style.Bg, false), ColorToRGBA(gc.Style.Fg, true)
	if gc.Style.Attrs&AttrReverse != 0 {
		bg, fg = fg, bg
	}
	im := image.NewRGBA(image.Rect(0, 0, size.X, size.Y))
	draw.Draw(im, im.Bounds(), image.NewUniform(bg), image.Point{}, draw.Src)
	x, y, thickness := size.X/2-1, size.Y/2-1, 3
	fill := func(r image.Rectangle) { draw.Draw(im, r, image.NewUniform(fg), image.Point{}, draw.Src) }
	if segments&up != 0 {
		fill(image.Rect(x, 0, x+thickness, y+thickness))
	}
	if segments&down != 0 {
		fill(image.Rect(x, y, x+thickness, size.Y))
	}
	if segments&left != 0 {
		fill(image.Rect(0, y, x+thickness, y+thickness))
	}
	if segments&right != 0 {
		fill(image.Rect(x, y, size.X, y+thickness))
	}
	return im
}
