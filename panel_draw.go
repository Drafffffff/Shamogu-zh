package main

import (
	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/ui"
	"strings"
)

func textPanelHeight() int {
	if LargeTextPanels {
		return UIHeight
	}
	return UIHeight - 1
}
func menuPanelHeight() int {
	if LargeTextPanels {
		return 10
	}
	return UIHeight - 1
}
func menuPanelWidth() int {
	if LargeTextPanels {
		return TextPanelWidth
	}
	return UIWidth / 2
}
func descriptionPanelWidth() int { return menuPanelWidth() }
func wrapPanelLines(lines []ui.StyledText) []ui.StyledText {
	if !LargeTextPanels {
		return lines
	}
	out := []ui.StyledText{}
	for _, line := range lines {
		out = append(out, line.Format(TextPanelWidth-2).Lines()...)
	}
	return out
}
func drawCentered(gd gruid.Grid, text ui.StyledText, y int) {
	w := text.Size().X
	x := max(0, (gd.Size().X-w)/2)
	text.Draw(gd.Slice(gruid.NewRange(x, y, gd.Size().X, gd.Size().Y)))
}
func introPrompt() ui.StyledText { return ui.Text("按任意键开始探索") }

func (md *model) drawTextPanel() (gruid.Grid, bool) {
	if !LargeTextPanels {
		return gruid.Grid{}, false
	}
	if md.mode == modeIntro {
		gd := md.gd
		gd.Fill(gruid.Cell{Rune: ' '})
		drawCentered(gd, gameTitle().WithStyle(gruid.Style{Fg: ColorMagenta}), 3)
		pic := gruid.NewGrid(23, 10)
		drawGamePicture(pic)
		gd.Slice(gruid.NewRange((UIWidth-23)/2, 6, (UIWidth-23)/2+23, 16)).Copy(pic)
		introText().Format(UIWidth - 8).Draw(gd.Slice(gruid.NewRange(4, 18, UIWidth-4, 23)))
		drawCentered(gd, introPrompt(), 23)
		return gd, true
	}
	gd := gruid.NewGrid(TextPanelWidth, UIHeight)
	switch md.mode {
	case modePager:
		gd.Copy(md.pager.pg.Draw())
	case modeNewGame, modeNewGameMods, modeMenu:
		if md.mode == modeMenu && (md.menu.mode == modeKeysView || md.menu.mode == modeKeysChange) {
			md.textDescription = nil
			md.textDescriptionKey = ""
			md.drawKeySettings()
			gd.Copy(md.menu.keys.Draw())
			break
		}
		y := 0
		if md.mode == modeNewGame {
			drawCentered(gd, gameTitle().WithStyle(gruid.Style{Fg: ColorMagenta}), 0)
			drawCentered(gd, md.spiritTabHint(), 1)
			y = 2
		}
		mgd := md.menu.main.Draw()
		gd.Slice(gruid.NewRange(0, y, TextPanelWidth, UIHeight)).Copy(mgd)
		y += mgd.Size().Y
		lines := []ui.StyledText{}
		if md.desc.Content.Text() != "" && (md.menu.mode != modeEquip || md.drawGroundDesc) {
			if md.desc.Box != nil {
				lines = append(lines, md.desc.Box.Title.WithStyle(gruid.Style{Fg: ColorCyan}))
			}
			lines = append(lines, md.desc.Content.Format(TextPanelWidth-2).Lines()...)
		}
		if md.mode == modeMenu && md.menu.mode == modeEquip {
			lines = append(lines, md.equipPager.ContentTitle().WithStyle(gruid.Style{Fg: ColorCyan}))
			lines = append(lines, md.equipPager.ContentLines()...)
		}
		if len(lines) > 0 && y < UIHeight {
			key := strings.Join(func() []string {
				ss := []string{}
				for _, line := range lines {
					ss = append(ss, line.Text())
				}
				return ss
			}(), "\n")
			key += string(rune(y))
			if md.textDescription == nil || md.textDescriptionKey != key {
				md.textDescription = ui.NewPager(ui.PagerConfig{Grid: gruid.NewGrid(TextPanelWidth, UIHeight-y), Lines: lines, Box: &ui.Box{}})
				md.textDescriptionKey = key
			}
			gd.Slice(gruid.NewRange(0, y, TextPanelWidth, UIHeight)).Copy(md.textDescription.Draw())
		} else {
			md.textDescription = nil
			md.textDescriptionKey = ""
		}
	default:
		return gruid.Grid{}, false
	}
	return gd, true
}

func (md *model) scrollTextDescription(msg gruid.Msg) bool {
	if !LargeTextPanels || md.textDescription == nil {
		return false
	}
	switch m := msg.(type) {
	case gruid.MsgKeyDown:
		if m.Key == gruid.KeyPageDown || m.Key == gruid.KeyPageUp {
			md.textDescription.Update(m)
			return true
		}
	case gruid.MsgMouse:
		if m.Action == gruid.MouseWheelDown || m.Action == gruid.MouseWheelUp {
			m.P = gruid.Point{}
			md.textDescription.Update(m)
			return true
		}
	}
	return false
}
