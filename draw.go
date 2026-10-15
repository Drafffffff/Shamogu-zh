// This file defines the Draw method for the model.

package main

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/ui"
)

// RenderOrder is a type representing the priority of an entity rendering.
type RenderOrder int

// Those constants represent distinct kinds of rendering priorities. In case
// two entities are at a given position, only the one with the highest priority
// gets displayed.
const (
	RONone RenderOrder = iota
	ROItem
	ROActor
)

// RenderOrder returns the rendering priority of an entity.
func (g *Game) RenderOrder(i ID) (ro RenderOrder) {
	e := g.Entity(i)
	if i == PlayerID {
		return ROActor
	}
	if a, ok := e.Role.(*Actor); ok && !a.KnownDead {
		return ROActor
	}
	if e.IsItem() && i >= FirstMapID {
		return ROItem
	}
	return RONone
}

// RenderOrderWizard returns the rendering priority of an entity for the wizard
// monster revealing mode.
func (g *Game) RenderOrderWizard(i ID) (ro RenderOrder) {
	e := g.Entity(i)
	if i == PlayerID {
		return ROActor
	}
	if a, ok := e.Role.(*Actor); ok && a.IsAlive() {
		return ROActor
	}
	if e.IsItem() && i >= FirstMapID {
		return ROItem
	}
	return RONone
}

// Draw implements Draw() for gruid.Model.
func (md *model) Draw() gruid.Grid {
	if !md.anims.Done() {
		return md.drawAnimationFrame()
	}
	if gd, ok := md.drawTextPanel(); ok {
		return gd
	}
	md.gd.Fill(gruid.Cell{Rune: ' '})
	switch md.mode {
	case modeQuitting:
		return md.gd.Slice(gruid.Range{})
	case modePager:
		if md.pager.mode == modePagerEnd {
			md.gd.Copy(md.pager.pg.Draw())
			return md.gd
		}
	case modeNewGame:
		md.drawSpiritSelection()
		return md.gd
	case modeNewGameMods:
		md.drawModSelection()
		return md.gd
	case modeLoadGame:
		md.drawLoadGameScreen()
		return md.gd
	}
	// Log drawing.
	md.log.Content = md.DrawLog()
	md.log.Draw(md.gd.Slice(md.gd.Range().Lines(0, 2)))
	// Map drawing.
	if md.g.win {
		md.drawWin(md.gd.Slice(md.gd.Range().Shift(0, 2, 0, -1)))
	} else {
		md.drawMap(md.gd.Slice(md.gd.Range().Shift(0, 2, 0, -1)))
	}
	// Some extra widgets may appear in some modes: they're drawn over log
	// lines and map.
	switch md.mode {
	case modeNormal:
		if md.status.focus {
			md.drawStatusDesc()
		} else if md.targ.p != InvalidPos {
			// Target description label drawing (over log lines and map).
			md.drawTargInfo()
		}
	case modePager:
		pgd := md.pager.pg.Draw()
		rg := md.pager.pg.View()
		sz := pgd.Size()
		if rg.Min.Y > 0 {
			pgd.Set(gruid.Point{sz.X - 1, 1}, gruid.Cell{Rune: '↑'})
		}
		if rg.Max.Y < md.pager.pg.Lines() {
			pgd.Set(gruid.Point{sz.X - 1, sz.Y - 2}, gruid.Cell{Rune: '↓'})
		}
		md.gd.Copy(pgd)
	case modeMenu:
		switch md.menu.mode {
		case modeInventory, modeInventoryBound, modeEquip:
			md.drawInventory()
		case modeGameMenu, modeConfigMenu, modeHelpMenu:
			md.gd.Copy(md.menu.main.Draw())
			if md.desc.Content.Text() != "" {
				md.desc.Draw(md.gd.Slice(md.gd.Range().Columns(UIWidth/2, UIWidth)))
			}
		case modeKeysView, modeKeysChange:
			md.drawKeySettings()
		}
	}
	// We draw the status line last: it should always be visible and no
	// other widgets should ever need that space.
	md.gd.Slice(md.gd.Range().Line(UIHeight-1).Shift(1, 0, 0, 0)).Copy(md.status.menu.Draw())
	if pages := md.status.menu.Pages(); pages.X > 0 {
		// When the status bar cannot hold all the content, draw arrows
		// at the start and/or end of the line.
		page := md.status.menu.Page()
		if page.X < pages.X {
			md.gd.Set(gruid.Point{UIWidth - 1, UIHeight - 1}, gruid.Cell{Rune: '>'})
		}
		if page.X > 0 {
			md.gd.Set(gruid.Point{0, UIHeight - 1}, gruid.Cell{Rune: '<'})
		}
	}
	return md.gd
}

func (md *model) drawAnimationFrame() gruid.Grid {
	if !md.anims.draw {
		return md.gd.Slice(gruid.Range{})
	}
	gd := md.gd
	for _, fc := range md.anims.frames[0].Cells {
		gd.Set(fc.P, fc.Cell)
	}
	md.anims.frames = md.anims.frames[1:]
	return gd
}

func (md *model) drawWin(gd gruid.Grid) {
	gd.Fill(gruid.Cell{
		Rune:  '♣',
		Style: gruid.Style{Fg: ColorGreen, Bg: ColorBackground, Attrs: AttrInMap}})
	textgd := gd.Slice(gruid.NewRange(6, 7, 6+62, 7+5))
	textgd.Fill(gruid.Cell{Rune: ' '})
	stt := ui.StyledText{}.WithMarkups(Markups)
	stt.WithText(`你摧毁了腐化之球！ ` +
		`地城正在消失，你感到自己被传送回了山中！ `).
		Format(60).Draw(textgd.Slice(textgd.Range().Shift(1, 1, -1, -1)))
}

func (md *model) drawMap(gd gruid.Grid) {
	g := md.g
	switch g.Wizard.Mode {
	case WizardReveal:
		md.drawMapWizardReveal(gd)
		return
	case WizardRevealTerrain:
		md.drawMapWizardRevealTerrain(gd)
		return
	}
	// Terrain drawing.
	for p := range g.Map.KnownTerrain.Points() {
		r := g.Map.KnownTerrainRuneAt(p)
		g.drawTerrainRuneAt(gd, r, p)
	}
	for _, cl := range g.Map.Clouds.Clouds {
		if !g.InFOV(cl.P) {
			continue
		}
		drawCloud(gd, cl)
	}
	// Entity drawing (with rendering order).
	SortedIDs := make([]ID, len(g.Entities))
	for i := range SortedIDs {
		SortedIDs[i] = ID(i)
	}
	slices.SortFunc(SortedIDs, func(i, j ID) int {
		return cmp.Compare(g.RenderOrder(i), g.RenderOrder(j))
	})
	for _, i := range SortedIDs {
		o := g.RenderOrder(i)
		if o == RONone {
			continue
		}
		e := md.g.Entity(i)
		if o == ROItem && g.CloudAt(e.P) && g.InFOV(e.P) {
			// Favor clouds over items, but in case of normal
			// clouds, show the color of the item.
			if g.Map.Clouds.At(e.P).Kind == CloudNormal {
				c := gd.At(e.P)
				c.Style.Fg = e.Color()
				gd.Set(e.P, c)
			}
			continue
		}
		if e.Noise && !g.InFOV(e.P) {
			c := gd.At(e.P)
			c.Rune = '♫'
			c.Style.Fg = noiseColor(e)
			gd.Set(e.P, c)
			if e.KnownP == e.P {
				continue
			}
		}
		p := e.KnownP
		if p == InvalidPos {
			continue
		}
		g.drawEntityAt(gd, e, p)
	}
	// Draw non-movement noise.
	for at := range g.Map.Noise {
		if !g.InFOV(at) {
			c := gd.At(at)
			c.Rune = '♪'
			c.Style.Fg = ColorCyan
			gd.Set(at, c)
		}
	}
	// Travel path highlighting.
	md.drawTravelPath(gd)
}

func (g *Game) sortedIDsByRenderOrder() []ID {
	sortedIDs := make([]ID, len(g.Entities))
	for i := range sortedIDs {
		sortedIDs[i] = ID(i)
	}
	slices.SortFunc(sortedIDs, func(i, j ID) int {
		return cmp.Compare(g.RenderOrder(i), g.RenderOrder(j))
	})
	return sortedIDs
}

func (md *model) drawMapWizardReveal(gd gruid.Grid) {
	g := md.g
	// Terrain drawing.
	for p, t := range g.Map.Terrain.All() {
		r := g.Map.TerrainRuneAt(p, t)
		g.drawTerrainRuneAt(gd, r, p)
	}
	for _, cl := range g.Map.Clouds.Clouds {
		drawCloud(gd, cl)
	}
	// Entity drawing (with rendering order).
	sortedIDs := g.sortedIDsByRenderOrder()
	for _, i := range sortedIDs {
		o := g.RenderOrderWizard(i)
		if o == RONone {
			continue
		}
		e := md.g.Entity(i)
		switch {
		case o == ROActor:
			if e.Actor().IsDead() {
				continue
			}
		case o == ROItem && g.CloudAt(e.P):
			// Favor clouds over items, but in case of normal
			// clouds, show the color of the item.
			if g.Map.Clouds.At(e.P).Kind == CloudNormal {
				c := gd.At(e.P)
				c.Style.Fg = e.Color()
				gd.Set(e.P, c)
			}
			continue
		}
		g.drawEntityAt(gd, e, e.P)
	}
	// Travel path highlighting.
	md.drawTravelPath(gd)
}

func (md *model) drawMapWizardRevealTerrain(gd gruid.Grid) {
	g := md.g
	// Terrain drawing.
	for p, t := range g.Map.Terrain.All() {
		r := g.Map.TerrainRuneAt(p, t)
		g.drawTerrainRuneAt(gd, r, p)
	}
	// Entity drawing (with rendering order).
	sortedIDs := g.sortedIDsByRenderOrder()
	for _, i := range sortedIDs {
		if g.RenderOrder(i) != ROItem {
			continue
		}
		e := md.g.Entity(i)
		g.drawEntityAt(gd, e, e.P)
	}
	it := gd.Iterator()
	for it.Next() {
		c := it.Cell()
		if c.Style.Fg == ColorForegroundSecondary {
			c.Style.Fg = ColorForeground
		}
		if c.Rune == ' ' {
			c.Style.Bg = ColorBackground
		} else {
			c.Style.Bg = ColorBackgroundSecondary
		}
		it.SetCell(c)
	}
	// Target highlighting.
	if p := md.targ.p; p != InvalidPos {
		c := gd.At(p)
		c.Style.Attrs |= AttrReverse
		gd.Set(p, c)
	}
}

func (g *Game) drawTerrainRuneAt(gd gruid.Grid, r rune, p gruid.Point) {
	c := gruid.Cell{Rune: r}
	if g.InFOV(p) {
		c.Style.Bg = ColorBackgroundSecondary
	} else {
		if c.Style.Fg == ColorForeground {
			c.Style.Fg = ColorForegroundSecondary
		}
	}
	c.Style.Attrs = AttrInMap
	if r == '#' || r == '◊' {
		c.Style.Attrs |= AttrBold
	}
	gd.Set(p, c)
}

func drawCloud(gd gruid.Grid, cl Cloud) {
	c := gd.At(cl.P)
	c.Rune = '≡'
	c.Style.Fg = cl.Kind.Color()
	gd.Set(cl.P, c)
}

func (md *model) drawTravelPath(gd gruid.Grid) {
	if md.targ.p != InvalidPos {
		for _, p := range md.auto.path {
			c := gd.At(p)
			c.Style.Attrs |= AttrReverse
			gd.Set(p, c)
		}
	}
}

func (g *Game) drawEntityAt(gd gruid.Grid, e *Entity, p gruid.Point) {
	c := gd.At(p)
	c.Rune = e.Rune
	if !e.IsActor() || g.InFOV(p) ||
		g.PlayerActor().Has(StatusClarity) && SensingRange(g.PP(), p) || g.Wizard.Mode.Reveal() {
		c.Style.Fg = e.Color()
	} else {
		c.Style.Fg = ColorForeground
	}
	gd.Set(p, c)
}

func (md *model) drawStatusDesc() {
	rg := md.status.menu.ActiveBounds()
	x := (rg.Min.X + rg.Max.X) / 2
	sz := md.status.desc.Content.Size()
	w, h := sz.X, sz.Y
	x -= w / 2
	if x+w > UIWidth-2 {
		x = UIWidth - w - 2
	}
	if x < 0 {
		x = 0
	}
	md.status.desc.Draw(md.gd.Slice(md.gd.Range().Lines(UIHeight-3-h, UIHeight-1).Shift(x, 0, 0, 0)))
}

// Markups contains the styling markup-characters we use for StyledText with
// default background. Lowercase variants use map tiles.
var Markups = map[rune]gruid.Style{
	'S': {Fg: ColorForegroundSecondary},
	'B': {Fg: ColorBlue},
	'C': {Fg: ColorCyan},
	'G': {Fg: ColorGreen},
	'M': {Fg: ColorMagenta},
	'O': {Fg: ColorOrange},
	'R': {Fg: ColorRed},
	'V': {Fg: ColorViolet},
	'Y': {Fg: ColorYellow},
	's': {Fg: ColorForegroundSecondary, Attrs: AttrInMap},
	'b': {Fg: ColorBlue, Attrs: AttrInMap},
	'c': {Fg: ColorCyan, Attrs: AttrInMap},
	'g': {Fg: ColorGreen, Attrs: AttrInMap},
	'm': {Fg: ColorMagenta, Attrs: AttrInMap},
	'o': {Fg: ColorOrange, Attrs: AttrInMap},
	'r': {Fg: ColorRed, Attrs: AttrInMap},
	'v': {Fg: ColorViolet, Attrs: AttrInMap},
	'y': {Fg: ColorYellow, Attrs: AttrInMap},
}

// MarkupsAltBg is similar to Markups but it uses secondary background.
var MarkupsAltBg = maps.Clone(Markups)

func init() {
	for r, v := range MarkupsAltBg {
		v.Bg = ColorBackgroundSecondary
		MarkupsAltBg[r] = v
	}
}

func (md *model) drawTargInfo() {
	g := md.g
	p := gruid.Point{}
	if md.targ.p.X < MapWidth/2 {
		p.X += MapWidth/2 + 1
	}
	info := md.targ.info
	n := md.targ.scroll
	y := 2
	stt := ui.StyledText{}.WithMarkups(Markups)
	formatBox := func(title, s string, fg gruid.Color) {
		if n > 0 {
			n--
			return
		}
		md.desc.Content = stt.WithText(s).Format(MapWidth/2 - 3)
		if md.desc.Content.Size().Y+y > MapHeight {
			md.desc.Box = &ui.Box{Title: stt.WithText(title).WithStyle(gruid.Style{Fg: fg}),
				Footer: ui.NewStyledText("滚动/向下翻页查看更多……", gruid.Style{Fg: ColorMagenta})}
		} else {
			md.desc.Box = &ui.Box{Title: stt.WithText(title).WithStyle(gruid.Style{Fg: fg})}
		}
		y += md.desc.Draw(md.gd.Slice(gruid.NewRange(0, y, MapWidth/2-1, 2+MapHeight).Add(p))).Size().Y
	}

	features := []string{}
	switch {
	case g.Wizard.Mode.Reveal():
		features = append(features, TerrainName(g.Map.Terrain.At(md.targ.p)))
	default:
		features = append(features, TerrainName(g.Map.KnownTerrain.At(md.targ.p)))
	}
	if info.cloud != NoCloud {
		clstr := info.cloud.String()
		features = append(features, fmt.Sprintf("@%c%s@N", info.cloud.ColorRune(), clstr))
	}
	if !info.sees && !info.unknown {
		features = append(features, "已发现")
	} else if info.unknown {
		features = append(features, "未探索")
	}
	t := features[0]
	if len(features) > 1 {
		t += " (" + strings.Join(features[1:], ", ") + ")"
	}
	var fg gruid.Color
	desc := ""
	switch {
	case g.Wizard.Mode.Reveal():
		desc = TerrainDesc(g.Map.Terrain.At(md.targ.p))
	default:
		desc = TerrainDesc(g.Map.KnownTerrain.At(md.targ.p))
	}
	if info.cloud != NoCloud {
		desc += "\n" + info.cloud.Desc()
	}
	if noise, ok := g.Map.Noise[md.targ.p]; ok && !g.InFOV(md.targ.p) {
		msg := "噪音"
		switch noise {
		case NoiseBark:
			msg = "吠叫"
		case NoiseCombat:
			msg = "撞击"
		case NoiseDig:
			msg = "落石"
		case NoiseExplosion:
			msg = "爆炸"
		case NoiseEarthMenhir:
			msg = "嘈杂回声" // should not happen (player-only now)
		case NoiseFakePortal:
			msg = "诡异声响" // should not happen (player-only now)
		case NoiseLightning:
			msg = "火花" // should not happen (player-only now)
		case NoiseRoar:
			msg = "咆哮" // should not happen (player-only now)
		case NoiseWind:
			msg = "哨声" // should not happen (player-only now)
		}
		formatBox("噪音", fmt.Sprintf("你听见附近传来%s。", msg), ColorCyan)
	}
	formatBox(t, desc, fg)
	for _, id := range info.entities {
		if int(id) >= len(md.g.Entities) {
			// Should not happen, but we keep this check in case
			// examination info doesn't get updated correctly
			// between levels.
			continue
		}
		e := g.Entity(id)
		if !info.sees && e.Noise && e.P == md.targ.p && !g.Wizard.Mode.Reveal() {
			noise := "脚步声"
			if a, ok := e.Role.(*Actor); ok {
				noise = monsterNoiseStringStyled(a)
			}
			formatBox("怪物发出的声音", fmt.Sprintf("你听见附近传来%s。", noise), ColorForeground)
			continue
		}
		if e.KnownP == InvalidPos && !g.Wizard.Mode.Reveal() {
			continue
		}
		var sb strings.Builder
		name := e.Name
		cl := e.Color()
		switch r := e.Role.(type) {
		case *Actor:
			switch {
			case g.Wizard.Mode == WizardRevealTerrain:
				continue
			case info.sees ||
				g.PlayerActor().Has(StatusClarity) && SensingRange(g.PP(), e.P) ||
				g.Wizard.Mode == WizardReveal:
				fmt.Fprintf(&sb, "生命：%s 攻击：%s 防御：%s",
					r.fmtHP(),
					fmtAD(r.GetAttack(), r.Attack), fmtAD(r.GetDefense(), r.Defense))
				for i, turns := range r.Statuses {
					if turns <= 0 {
						continue
					}
					st := Status(i)
					fmt.Fprintf(&sb, " @%c%s(%d)@N ", statusColor(st, turns), g.StatusAbbr(r, st), turns)
				}
				if s := g.MindStateString(r); s != "" {
					name += fmt.Sprintf(" (%s)", s)
				}
			default:
				cl = ColorForeground
				fmt.Fprintf(&sb, "生命：?/%d 攻击：%d 防御：%d", r.MaxHP, r.Attack, r.Defense)
			}
			sb.WriteByte('\n')
			fmt.Fprintf(&sb, "@C特性：@N %s。", g.TraitDesc(r.Kind, r.Traits))
		case *Spirit:
			sb.WriteString(r.ExamineDesc(g))
		case *Menhir:
			sb.WriteString(r.Desc(g))
			if r.Used {
				name += "（失效）"
			}
		case *RunicTrap:
			sb.WriteString(r.Desc(g))
			if r.KnownUsed || r.Used && g.Wizard.Mode.Reveal() {
				name += "（失效）"
			}
		case *Comestible:
			sb.WriteString(r.ExamineDesc(g))
		case Item:
			sb.WriteString(r.Desc(g))
		}
		formatBox(name, sb.String(), cl)
	}
}

const spiritSelectionMargin = 5

func (md *model) drawSpiritSelection() {
	menugd := md.menu.main.Draw()
	gdslice := md.gd.Slice(gruid.NewRange(0, spiritSelectionMargin, UIWidth, UIHeight))
	gdslice.Copy(menugd)
	md.desc.Draw(gdslice.Slice(gdslice.Range().Columns(UIWidth/2, UIWidth)))
	ui.Textf("萨满山灵守护者 · Shamogu %s", Version).WithStyle(gruid.Style{}.WithFg(ColorMagenta)).
		Draw(md.gd.Slice(gruid.NewRange(-25+UIWidth/2, 3, UIWidth, UIHeight)))
	y := 8 + len(primarySpirits)
	n := md.g.modCount()
	var tabtext string
	if n > 0 {
		tabtext = fmt.Sprintf(`@G（按 TAB 选择模组，已启用 @V%d@G 个）@N`, n)
	} else {
		tabtext = `@G（按 TAB 选择模组）@N`
	}
	ui.Text(tabtext).WithMarkups(Markups).
		Draw(md.gd.Slice(gruid.NewRange(0, y, UIWidth/2, y+1)))
	y = 10 + len(primarySpirits)
	ui.Text(`山中万物正在遭受腐化。野兽逐渐失去理智，变得富有攻击性。这些异变始于一座地城传送门突然 ` +
		`出现在山顶之后。作为山之守护者，你决定深入地城，找出并摧毁腐化的源头……愿好运与你同在！`).Format(72).
		Draw(md.gd.Slice(gruid.NewRange(4, y, UIWidth-4, UIHeight)))
}

func (md *model) drawModSelection() {
	menugd := md.menu.main.Draw()
	md.gd.Copy(menugd)
	md.desc.Draw(md.gd.Slice(md.gd.Range().Columns(UIWidth/2, UIWidth)))
}

func (md *model) drawLoadGameScreen() {
	gd := md.gd
	drawCentered(gd, gameTitle().WithStyle(gruid.Style{Fg: ColorMagenta}), 4)
	picture := gruid.NewGrid(23, 10)
	drawGamePicture(picture)
	gd.Slice(gruid.NewRange((UIWidth-23)/2, 7, (UIWidth-23)/2+23, 17)).Copy(picture)
	drawCentered(gd, loadPrompt(), 19)
}

func drawGamePicture(gd gruid.Grid) {
	markups := maps.Clone(Markups)
	for i, st := range markups {
		markups[i] = st.WithAttrs(AttrInMap)
	}
	st := gruid.Style{}.WithAttrs(AttrInMap)
	markups['w'] = st.WithFg(ColorForeground).WithBg(ColorBackgroundSecondary).WithAttrs(AttrBold | AttrInMap)
	markups['W'] = st.WithFg(ColorForegroundSecondary).WithBg(ColorBackground).WithAttrs(AttrBold | AttrInMap)
	markups['l'] = st.WithFg(ColorForeground).WithBg(ColorBackgroundSecondary)
	markups['y'] = st.WithFg(ColorYellow).WithBg(ColorBackgroundSecondary)
	markups['c'] = st.WithFg(ColorCyan).WithBg(ColorBackgroundSecondary)
	markups['b'] = st.WithFg(ColorBlue).WithBg(ColorBackgroundSecondary)
	markups['r'] = st.WithFg(ColorRed).WithBg(ColorBackgroundSecondary)
	markups['d'] = st.WithFg(ColorForegroundSecondary).WithBg(ColorBackground)
	markups['t'] = gruid.Style{}.WithFg(ColorGreen)
	stt := ui.StyledText{}.WithMarkups(markups)
	text := `@t───────────────────────
@l""@d...@w#@W##########@d...@W#@d^@W##
@d"@l.@w◊@d.@l@rF@w#@t SHAMOGU @W#@d.....^@W#
@d.@l^@c&@w#@l..@w#####@W####@d..@W##@d....
@l^..@b@@@l.....@rC@l..@d....@W###@d..."
@l..^@w#@l.@rc@l.@w◊@c>@w#@d^..@W#####@d.@Y%@d.""
@d.@O♫@d"@W#@w#@l...@w◊@y!@d..""@W###@d..""""
@d."""@W#@w#@l...@d..."""@W#@d..""""@W#
@d"""""@W#@w#@l.@d.........."""@W##
@t───────────────────────
`
	stt.WithText(text).Draw(gd)
}

func (md *model) drawInventory() {
	menugd := md.menu.main.Draw()
	md.gd.Copy(menugd)
	var y int
	if md.desc.Content.Text() != "" && (md.menu.mode != modeEquip || md.drawGroundDesc) {
		descgd := md.desc.Draw(md.gd.Slice(md.gd.Range().Columns(UIWidth/2, UIWidth)))
		y = descgd.Size().Y
	}
	if md.menu.mode != modeEquip {
		return
	}
	// We draw the description of the item on the ground below the other
	// one (if any).
	md.gd.Slice(gruid.NewRange(UIWidth/2, y, UIWidth, UIHeight-1)).Copy(md.equipPager.Draw())
}

func (md *model) drawKeySettings() {
	gd := md.menu.keys.Draw()
	max := gd.Size()
	t := ui.Text("（R）重置 （Enter）修改").WithStyle(gruid.Style{}.WithFg(ColorCyan))
	if md.menu.mode == modeKeysChange {
		t = t.WithText(" 按下新按键…… ")
	}
	t.Draw(gd.Slice(gd.Range().Line(max.Y-1).Shift(2, 0, 0, 0)))
	md.gd.Copy(gd)
}

func gameTitle() ui.StyledText { return ui.Textf("萨满山灵守护者 · Shamogu %s", Version) }
func introText() ui.StyledText {
	return ui.Text(`山中万物正在遭受腐化。野兽逐渐失去理智，变得富有攻击性。这些异变始于一座地城传送门突然 ` +
		`出现在山顶之后。作为山之守护者，你决定深入地城，找出并摧毁腐化的源头……愿好运与你同在！`)
}
func loadPrompt() ui.StyledText { return ui.Text("—按任意键载入存档—") }

func (md *model) spiritTabHint() ui.StyledText {
	n := md.g.modCount()
	var tabtext string
	if n > 0 {
		tabtext = fmt.Sprintf(`@G（按 TAB 选择模组，已启用 @V%d@G 个）@N`, n)
	} else {
		tabtext = `@G（按 TAB 选择模组）@N`
	}
	return ui.Text(tabtext).WithMarkups(Markups)
}
