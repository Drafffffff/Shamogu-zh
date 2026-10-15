//go:build sdl

package main

import (
	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/ui"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"golang.org/x/image/font/sfnt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestChineseFontCoverage(t *testing.T) {
	data, _ := fontFiles.ReadFile("fonts/HarmonistPixelSC-Regular.ttf")
	ft, e := sfnt.Parse(data)
	if e != nil {
		t.Fatal(e)
	}
	var buf sfnt.Buffer
	names, _ := filepath.Glob("*.go")
	seen := map[rune]bool{}
	for _, name := range names {
		f, e := parser.ParseFile(token.NewFileSet(), name, nil, 0)
		if e != nil {
			t.Fatal(e)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if x, ok := n.(*ast.BasicLit); ok && x.Kind == token.STRING {
				s, _ := strconv.Unquote(x.Value)
				for _, r := range s {
					if isChineseGlyph(r) && !seen[r] {
						seen[r] = true
						index, e := ft.GlyphIndex(&buf, r)
						if e != nil || index == 0 {
							t.Errorf("missing glyph %U", r)
						}
					}
				}
			}
			return true
		})
	}
	t.Logf("verified %d CJK glyphs", len(seen))
}
func TestChineseRenderScreens(t *testing.T) {
	dir := os.Getenv("SHAMOGU_SCREENSHOTS")
	if dir == "" {
		t.Skip("set SHAMOGU_SCREENSHOTS to export screens")
	}
	os.MkdirAll(dir, 0755)
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	md := &model{gd: gruid.NewGrid(UIWidth, UIHeight), g: &Game{}, targ: &targeting{}}
	md.init()
	render := func(name string) {
		gd := md.Draw()
		img := image.NewRGBA(image.Rect(0, 0, 3840, 2160))
		tm := &monochromeTileManager{}
		size := tm.TileSizeForFrame(gd.Size().X, gd.Size().Y)
		gd.Iter(func(p gruid.Point, c gruid.Cell) {
			draw.Draw(img, image.Rect(p.X*size.X, p.Y*size.Y, (p.X+1)*size.X, (p.Y+1)*size.Y), tm.GetImage(c), image.Point{}, draw.Src)
		})
		f, e := os.Create(filepath.Join(dir, name+".png"))
		if e != nil {
			t.Fatal(e)
		}
		defer f.Close()
		if e := png.Encode(f, img); e != nil {
			t.Fatal(e)
		}
	}
	render("welcome")
	md.openSpiritSelectionMenu(modeNewGame)
	render("spirits")
	md.modSelectionMenu()
	render("mods")
	md.startNewGame(spiritEntity(md.g.Mods, primarySpirits[0]))
	render("game")
	(ActionHelp{}).Handle(md)
	render("help")
	(ActionHelpCombat{}).Handle(md)
	render("basics")
	md.openInventory(-1)
	render("inventory")
	md.mode = modeLoadGame
	render("load")

}

func TestPixelGlyphsAreCrisp(t *testing.T) {
	for _, r := range "汉化鹞牦" {
		gc := gruid.Cell{Rune: r, Style: gruid.Style{Fg: ColorForeground, Bg: ColorBackground}}
		im := chineseGlyph(gc)
		bg := ColorToRGBA(gc.Style.Bg, false)
		fg := ColorToRGBA(gc.Style.Fg, true)
		br, bgch, bb, _ := bg.RGBA()
		fr, fgch, fb, _ := fg.RGBA()
		count := 0
		for y := 0; y < 90; y++ {
			for x := 0; x < 48; x++ {
				rr, gg, bl, _ := im.At(x, y).RGBA()
				if rr == fr && gg == fgch && bl == fb {
					count++
					continue
				}
				if rr != br || gg != bgch || bl != bb {
					t.Fatalf("non-pixel antialiasing for %c at %d,%d", r, x, y)
				}
			}
		}
		if count == 0 {
			t.Fatalf("blank glyph %c", r)
		}
	}
}

func TestMixedScriptInkAlignment(t *testing.T) {
	tm := &monochromeTileManager{}
	ink := func(r rune) image.Rectangle {
		gc := gruid.Cell{Rune: r, Style: gruid.Style{Fg: ColorForeground, Bg: ColorBackground}}
		im := tm.GetImage(gc)
		fg := ColorToRGBA(gc.Style.Fg, true)
		fr, fgc, fb, _ := fg.RGBA()
		bounds := image.Rectangle{}
		for y := 0; y < im.Bounds().Dy(); y++ {
			for x := 0; x < im.Bounds().Dx(); x++ {
				rr, gg, bb, _ := im.At(x, y).RGBA()
				if rr == fr && gg == fgc && bb == fb {
					bounds = bounds.Union(image.Rect(x, y, x+1, y+1))
				}
			}
		}
		return bounds
	}
	for _, large := range []bool{false, true} {
		tm.large = large
		latin := ink('H')
		for _, r := range "汉中化鹞牦" {
			cjk := ink(r)
			if d := latin.Dy() - cjk.Dy(); d < -4 || d > 4 {
				t.Errorf("%c height %d differs from Latin capital height %d", r, cjk.Dy(), latin.Dy())
			}
			if d := latin.Max.Y - cjk.Max.Y; d < -1 || d > 1 {
				t.Errorf("%c ink bottom %d differs from Latin bottom %d", r, cjk.Max.Y, latin.Max.Y)
			}
			t.Logf("%c ink=%v, H ink=%v", r, cjk, latin)
		}
	}
}

func TestContinuousBoxBorders(t *testing.T) {
	for _, large := range []bool{false, true} {
		tm := &monochromeTileManager{large: large}
		sz := tm.TileSize()
		gd := gruid.NewGrid(3, 3)
		st := gruid.Style{Fg: ColorForeground, Bg: ColorBackground}
		ui.Box{Style: st}.Draw(gd)
		im := image.NewRGBA(image.Rect(0, 0, 3*sz.X, 3*sz.Y))
		gd.Iter(func(p gruid.Point, c gruid.Cell) {
			draw.Draw(im, image.Rect(p.X*sz.X, p.Y*sz.Y, (p.X+1)*sz.X, (p.Y+1)*sz.Y), tm.GetImage(c), image.Point{}, draw.Src)
		})
		fg := ColorToRGBA(st.Fg, true)
		assertInk := func(x, y int) {
			if im.At(x, y) != fg {
				t.Fatalf("border gap at %d,%d, large=%v", x, y, large)
			}
		}
		left, right, top, bottom := sz.X/2, 2*sz.X+sz.X/2, sz.Y/2, 2*sz.Y+sz.Y/2
		for y := top; y <= bottom; y++ {
			assertInk(left, y)
			assertInk(right, y)
		}
		for x := left; x <= right; x++ {
			assertInk(x, top)
			assertInk(x, bottom)
		}
	}
}

func TestLargePanelDescriptionScrollsWithoutClipping(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	md := &model{gd: gruid.NewGrid(UIWidth, UIHeight), g: &Game{}, targ: &targeting{}}
	md.init()
	md.startNewGame(spiritEntity(md.g.Mods, primarySpirits[0]))
	md.openInventory(-1)
	lines := ""
	for i := 0; i < 60; i++ {
		lines += fmt.Sprintf("row-%02d\n", i)
	}
	md.desc.Content = ui.Text(lines)
	gd, ok := md.drawTextPanel()
	if !ok || gd.Size().X != 48 {
		t.Fatal("large panel not used")
	}
	if md.textDescription == nil || md.textDescription.Lines() != 61 {
		t.Fatal("description lines lost")
	}
	for i := 0; i < 60; i++ {
		md.scrollTextDescription(gruid.MsgKeyDown{Key: gruid.KeyPageDown})
	}
	if md.textDescription.View().Max.Y != 61 {
		t.Fatalf("end of long description not reachable: %v", md.textDescription.View())
	}
	md.mode = modeNormal
	if md.Draw().Size().X != 80 {
		t.Fatal("returning to map lost columns")
	}
}

func TestDynamicItemDescriptionsAcrossMods(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	md := &model{gd: gruid.NewGrid(UIWidth, UIHeight), g: &Game{}, targ: &targeting{}}
	md.init()
	g := md.g
	variants := [][]bool{make([]bool, NMods), make([]bool, NMods)}
	for i := range variants[1] {
		variants[1][i] = true
	}
	for i := range NMods {
		m := make([]bool, NMods)
		m[i] = true
		variants = append(variants, m)
	}
	count := 0
	check := func(name, text string) {
		t.Helper()
		if strings.Contains(text, "%!") {
			t.Fatalf("broken format in %s: %s", name, text)
		}
		if ui.Text(text).WithMarkups(Markups).Format(TextPanelWidth-2).Size().X > TextPanelWidth-2 {
			t.Fatalf("description overflows: %s", name)
		}
		count++
	}
	spirits := append(append(append([]spiritInfo{}, primarySpirits...), secondarySpirits...), challengeSpirits...)
	for _, mods := range variants {
		g.Mods = mods
		g.Init(spiritEntity(mods, primarySpirits[0]))
		for _, info := range spirits {
			e := spiritEntity(mods, info)
			sp := e.Role.(*Spirit)
			check(e.Name, sp.Desc(g))
			check(e.Name, sp.UpgradeDesc(g, e))
			sp.Level = 1
			check(e.Name, sp.Desc(g))
		}
		for ck := range ComestibleData {
			e := comestibleEntity(comestibleKind(ck))
			check(e.Name, e.Role.(*Comestible).Desc(g))
		}
		for _, info := range MenhirData {
			e := menhirEntity(info)
			check(e.Name, e.Role.(*Menhir).Desc(g))
		}
	}
	t.Logf("checked %d dynamic descriptions", count)
}
