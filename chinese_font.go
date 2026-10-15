//go:build js || sdl

package main

import (
	"embed"
	"image"
	"image/color"
	"image/draw"
	"unicode"

	"codeberg.org/anaseto/gruid"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

//go:embed fonts/HarmonistPixelSC-Regular.ttf
var fontFiles embed.FS
var chineseFace font.Face
var largeChineseFace font.Face

func isChineseGlyph(r rune) bool {
	return unicode.Is(unicode.Han, r) || unicode.Is(unicode.Hiragana, r) || unicode.Is(unicode.Katakana, r) || r >= 0x3000 && r <= 0x303f || r >= 0xff00 && r <= 0xffef
}

func initChineseFont() {
	if chineseFace != nil {
		return
	}
	data, err := fontFiles.ReadFile("fonts/HarmonistPixelSC-Regular.ttf")
	if err != nil {
		panic(err)
	}
	ft, err := opentype.Parse(data)
	if err != nil {
		panic(err)
	}
	chineseFace, err = opentype.NewFace(ft, &opentype.FaceOptions{Size: 48, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		panic(err)
	}
	largeChineseFace, err = opentype.NewFace(ft, &opentype.FaceOptions{Size: 72, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		panic(err)
	}
}

func chineseGlyph(gc gruid.Cell) image.Image {
	return pixelTextGlyph(gc, false)
}

func pixelTextGlyph(gc gruid.Cell, large bool) image.Image {
	initChineseFont()
	width, bottom, baseline := 48, 63, 55
	face := chineseFace
	if large {
		width, bottom, baseline = 80, 74, 62
		face = largeChineseFace
	}
	img := image.NewRGBA(image.Rect(0, 0, width, 90))
	bg, fg := ColorToRGBA(gc.Style.Bg, false), ColorToRGBA(gc.Style.Fg, true)
	if gc.Style.Attrs&AttrReverse != 0 {
		fg, bg = bg, fg
	}
	draw.Draw(img, img.Bounds(), image.NewUniform(bg), image.Point{}, draw.Src)
	d := font.Drawer{Dst: img, Src: image.NewUniform(fg), Face: face}
	advance := d.MeasureString(string(gc.Rune))
	// CJK ink is 44 px high; native Latin capitals are 48 px. Align their
	// visible bottoms at y=63 (the pixel font's descent contributes 8 px).
	d.Dot = fixed.Point26_6{X: (fixed.I(width) - advance) / 2, Y: fixed.I(baseline)}
	if unicode.Is(unicode.Han, gc.Rune) || unicode.Is(unicode.Hiragana, gc.Rune) || unicode.Is(unicode.Katakana, gc.Rune) {
		bounds, _ := font.BoundString(face, string(gc.Rune))
		// Center the visible ink, including locally supplied fallback glyphs
		// whose bearings differ from the original font's CJK glyphs.
		d.Dot.X = fixed.I((width-(bounds.Max.X-bounds.Min.X).Ceil())/2 - bounds.Min.X.Floor())
		d.Dot.Y = fixed.I(bottom - bounds.Max.Y.Ceil())
	}
	d.DrawString(string(gc.Rune))
	return img
}

func scaleNativeTile(src *image.RGBA, bg color.Color) image.Image {
	return scaleNativePixels(src, bg, false)
}

func scaleNativePixels(src *image.RGBA, bg color.Color, large bool) image.Image {
	width, scale, left, top := 48, 3, 0, 9
	if large {
		width, scale, left, top = 80, 4, 8, 2
	}
	dst := image.NewRGBA(image.Rect(0, 0, width, 90))
	// Integer enlargement preserves the original pixel-art proportions. The extra
	// vertical margin gives exactly 16:9 for the 80 by 24 grid, without stretching.
	for y := 0; y < 90; y++ {
		for x := 0; x < width; x++ {
			sy := (y - top) / scale
			if y < top || sy >= 24 || x < left || x >= left+16*scale {
				dst.Set(x, y, bg)
				continue
			}
			dst.Set(x, y, src.At((x-left)/scale, sy))
		}
	}
	return dst
}
