// This file defines most actions available in the game.

package main

import (
	"cmp"
	"fmt"
	"log"
	"maps"
	"runtime"
	"slices"
	"strings"
	"time"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
	"codeberg.org/anaseto/gruid/ui"
)

// Action represents types that describe and handle a game action, often the
// last UI Action performed.
type Action interface {
	// Handle processes an action and returns possibly an effect along with
	// a boolean that reports whether the action ends the player's game
	// turn.
	Handle(*model) (gruid.Effect, bool)
}

// ActionDesc is a named action with a description.
type ActionDesc interface {
	Action
	String() string
	Desc(*model) string
}

// updateActionDesc updates the description label for the given described action.
func (md *model) updateActionDesc(a ActionDesc) {
	l := md.desc
	stt := ui.StyledText{}.WithMarkups(Markups)
	l.Box = &ui.Box{Title: ui.Text(a.String())}
	l.Content = stt.WithText(a.Desc(md)).Format(descriptionPanelWidth() - 2)
}

// ActionNone does nothing.
type ActionNone struct{}

func (a ActionNone) Handle(md *model) (gruid.Effect, bool) {
	return nil, false
}

// ActionWait waits for a turn.
type ActionWait struct{}

func (a ActionWait) Handle(md *model) (gruid.Effect, bool) {
	md.g.stopSprint()
	md.g.Stats.Waits++
	return nil, true
}

func (g *Game) stopSprint() {
	if pa := g.PlayerActor(); pa.Has(StatusSprint) {
		pa.Statuses[StatusSprint] = 0
		g.LogStyled("你停止冲刺。", logStatusEnd)
	}
}

func (a ActionWait) String() string {
	return "等待一回合"
}

// ActionBump moves the player to a given position and updates FOV information,
// or attacks if there is a monster.
type ActionBump struct {
	Delta gruid.Point
}

func (a ActionBump) Handle(md *model) (eff gruid.Effect, done bool) {
	return nil, md.g.PlayerBump(a.Delta)
}

func (a ActionBump) String() string {
	var text string
	switch a.Delta {
	case gruid.Point{-1, 0}:
		text = "向西移动"
	case gruid.Point{1, 0}:
		text = "向东移动"
	case gruid.Point{0, -1}:
		text = "向北移动"
	case gruid.Point{0, 1}:
		text = "向南移动"
	}
	return text
}

// ActionRun performs automatic movement in a given direction.
type ActionRun struct {
	Delta gruid.Point
}

func (a ActionRun) Handle(md *model) (gruid.Effect, bool) {
	g := md.g
	if g.DangerInFOV() {
		g.Log("无法奔跑：视野中有危险！")
		return nil, false
	}
	if pa := g.PlayerActor(); pa.Has(StatusPoison) {
		g.Log("中毒时无法奔跑！")
		return nil, false
	}
	p := g.PP().Add(a.Delta)
	if !g.PlayerPassableNoTraps(p) {
		g.Log("无法朝那个方向奔跑。")
		return nil, false
	}
	bump := ActionBump(a)
	eff, done := bump.Handle(md)
	if !done {
		return eff, false
	}
	md.UpdateRun(a.Delta)
	md.targ.CancelExamine()
	return md.UpdateAutoMode(eff, autoRun), true
}

func (a ActionRun) String() string {
	var text string
	switch a.Delta {
	case gruid.Point{-1, 0}:
		text = "向西奔跑"
	case gruid.Point{1, 0}:
		text = "向东奔跑"
	case gruid.Point{0, -1}:
		text = "向北奔跑"
	case gruid.Point{0, 1}:
		text = "向南奔跑"
	}
	return text
}

// ActionTravel starts travelling to the currently examined target.
type ActionTravel struct{}

func (a ActionTravel) Handle(md *model) (gruid.Effect, bool) {
	tp := md.targ.p
	if tp == InvalidPos {
		return nil, false
	}
	g := md.g
	pp := g.PP()
	if tp == pp {
		// wait a turn
		return nil, true
	}
	path := md.auto.path
	if len(path) <= 1 || path[len(path)-1] != tp || path[0] != pp {
		path = g.PlayerPath(pp, tp)
	}
	if len(path) <= 1 {
		return nil, false
	}
	fif := g.DangerInFOV() || g.PlayerActor().Has(StatusPoison)
	next := path[1]
	bump := ActionBump{Delta: next.Sub(pp)}
	eff, done := bump.Handle(md)
	if !done {
		return eff, false
	}
	if g.PP() == next {
		md.auto.path = path[1:]
	}
	if fif {
		// do not start auto-travel if there was a foe in view.
		return eff, true
	}
	md.targ.CancelExamine()
	return md.UpdateAutoMode(eff, autoTravel), true
}

func (a ActionTravel) String() string {
	return "前往目标"
}

// ActionAutoExplore starts automatic exploration.
type ActionAutoExplore struct{}

func (a ActionAutoExplore) Handle(md *model) (gruid.Effect, bool) {
	g := md.g
	if g.DangerInFOV() {
		g.Log("无法自动探索：视野中有危险！")
		return nil, false
	}
	if pa := g.PlayerActor(); pa.Has(StatusPoison) {
		g.Log("中毒时无法自动探索！")
		return nil, false
	}
	md.auto.aemRebuild = true
	n, ok := md.UpdateAutoExplore()
	if !ok {
		return nil, false
	}
	bump := ActionBump{Delta: n.Sub(g.PP())}
	eff, done := bump.Handle(md)
	if !done {
		return eff, false
	}
	md.targ.CancelExamine()
	return md.UpdateAutoMode(eff, autoExplore), true
}

func (a ActionAutoExplore) String() string {
	return "自动探索"
}

// ActionCursorBump moves the cursor.
type ActionCursorBump struct {
	Delta gruid.Point
}

func (a ActionCursorBump) Handle(md *model) (gruid.Effect, bool) {
	np := md.targ.p.Add(a.Delta)
	md.Examine(np)
	return nil, false
}

func (a ActionCursorBump) String() string {
	var text string
	switch a.Delta {
	case gruid.Point{-1, 0}:
		text = "光标向西移动"
	case gruid.Point{1, 0}:
		text = "光标向东移动"
	case gruid.Point{0, -1}:
		text = "光标向北移动"
	case gruid.Point{0, 1}:
		text = "光标向南移动"
	}
	return text
}

// ActionCursorRun moves the cursor (fast).
type ActionCursorRun struct {
	Delta gruid.Point
}

func (a ActionCursorRun) Handle(md *model) (gruid.Effect, bool) {
	np := md.targ.p.Add(a.Delta.Mul(4))
	if np.X < 0 {
		np.X = 0
	}
	if np.X >= MapWidth {
		np.X = MapWidth - 1
	}
	if np.Y < 0 {
		np.Y = 0
	}
	if np.Y >= MapHeight {
		np.Y = MapHeight - 1
	}
	md.Examine(np)
	return nil, false
}

func (a ActionCursorRun) String() string {
	var text string
	switch a.Delta {
	case gruid.Point{-1, 0}:
		text = "光标快速向西移动"
	case gruid.Point{1, 0}:
		text = "光标快速向东移动"
	case gruid.Point{0, -1}:
		text = "光标快速向北移动"
	case gruid.Point{0, 1}:
		text = "光标快速向南移动"
	}
	return text
}

// ActionNextMonster examines next monster (sorted by distance to player).
type ActionNextMonster struct{}

func (a ActionNextMonster) Handle(md *model) (gruid.Effect, bool) {
	g := md.g
	if g.Wizard.Mode == WizardRevealTerrain {
		return nil, false
	}
	me := g.getMonsterTargIDs()
	if len(me) == 0 {
		return nil, false
	}
	g.sortEntityTargIDs(me)
	i := g.getMonsterTargIDIndex(me)
	i = (i + 1) % len(me)
	var p gruid.Point
	if g.Wizard.Mode == WizardReveal {
		p = g.Entity(me[i]).P
	} else {
		p = g.Entity(me[i]).KnownP
	}
	md.targ.kb = true
	md.Examine(p)
	return nil, false
}

// getMonsterTargIDs returns the ids of monsters that can be examined.
func (g *Game) getMonsterTargIDs() []ID {
	var me []ID
	for i, ei := range g.NPMapEntities() {
		ai, ok := ei.Role.(*Actor)
		if !ok {
			continue
		}
		var b bool // whether monster is examinable
		if g.Wizard.Mode == WizardReveal {
			b = ai.IsAlive()
		} else {
			b = ei.KnownP != InvalidPos && !ai.KnownDead
		}
		if b {
			me = append(me, i)
		}
	}
	return me
}

// sortEntityTargIDs sorts the ids of monsters that can be examined by
// distance to the player.
func (g *Game) sortEntityTargIDs(me []ID) {
	pp := g.PP()
	slices.SortFunc(me, func(i, j ID) int {
		pi, pj := g.Entity(i).KnownP, g.Entity(j).KnownP
		if g.Wizard.Mode.Reveal() {
			pi, pj = g.Entity(i).P, g.Entity(j).P
		}
		return cmp.Compare(paths.DistanceManhattan(pp, pi), paths.DistanceManhattan(pp, pj))
	})
}

// getMonsterTargIDIndex returns the id of any monster currently being
// examined, or -1 if none.
func (g *Game) getMonsterTargIDIndex(me []ID) int {
	if g.Wizard.Mode == WizardReveal {
		if id, _ := g.ActorAt(g.md.targ.p); id >= 0 {
			return slices.Index(me, id)
		}
		return -1
	}
	if id, _ := g.KnownActorAt(g.md.targ.p); id >= 0 {
		return slices.Index(me, id)
	}
	return -1
}

func (a ActionNextMonster) String() string {
	return "查看下一个怪物"
}

// ActionPreviousMonster examines previous monster (sorted by distance to
// player).
type ActionPreviousMonster struct{}

func (a ActionPreviousMonster) Handle(md *model) (gruid.Effect, bool) {
	g := md.g
	if g.Wizard.Mode == WizardRevealTerrain {
		return nil, false
	}
	me := g.getMonsterTargIDs()
	if len(me) == 0 {
		return nil, false
	}
	g.sortEntityTargIDs(me)
	i := g.getMonsterTargIDIndex(me)
	if i == -1 {
		i = 0
	}
	i--
	if i < 0 {
		i += len(me)
	}
	var p gruid.Point
	if g.Wizard.Mode == WizardReveal {
		p = g.Entity(me[i]).P
	} else {
		p = g.Entity(me[i]).KnownP
	}
	md.targ.kb = true
	md.Examine(p)
	return nil, false
}

func (a ActionPreviousMonster) String() string {
	return "查看上一个怪物"
}

// ActionNextItem examines next food item (sorted by distance to player).
type ActionNextItem struct{ Type itemType }

// itemType represents various kinds of targetable item kinds.
type itemType int

const (
	itemComestible itemType = iota
	itemTotem
	itemMenhir
	itemPortal // includes CorruptionOrb too
	itemRune
)

func item2type(it Item) itemType {
	switch it.(type) {
	case *Comestible:
		return itemComestible
	case *Spirit:
		return itemTotem
	case *EmptyTotem:
		return itemTotem
	case *Menhir:
		return itemMenhir
	case *Portal:
		return itemPortal
	case *CorruptionOrb:
		return itemPortal
	case *RunicTrap:
		return itemRune
	default:
		panic("意外的物品")
	}
}

func (a ActionNextItem) Handle(md *model) (gruid.Effect, bool) {
	g := md.g
	items := g.getItemTargIDs(a.Type)
	if len(items) == 0 {
		return nil, false
	}
	g.sortEntityTargIDs(items)
	i := g.getItemTargIDIndex(items, a.Type)
	i = (i + 1) % len(items)
	var p gruid.Point
	if g.Wizard.Mode.Reveal() {
		p = g.Entity(items[i]).P
	} else {
		p = g.Entity(items[i]).KnownP
	}
	if p == InvalidPos {
		return nil, false
	}
	md.targ.kb = true
	md.Examine(p)
	return nil, false
}

// getItemTargIDs returns the ids of items that can be examined.
func (g *Game) getItemTargIDs(ty itemType) []ID {
	var items []ID
	for i, it := range g.ItemEntities() {
		if item2type(it) != ty {
			continue
		}
		if g.Entity(i).KnownP != InvalidPos || g.Wizard.Mode.Reveal() {
			items = append(items, i)
		}
	}
	return items
}

// getItemTargIDIndex returns the id of any monster currently being
// examined, or -1 if none.
func (g *Game) getItemTargIDIndex(items []ID, ty itemType) int {
	if g.Wizard.Mode.Reveal() {
		if id, it := g.ItemAt(g.md.targ.p); id >= 0 {
			if item2type(it) == ty {
				return slices.Index(items, id)
			}
		}
		return -1
	}
	if id, it := g.KnownItemAt(g.md.targ.p); id >= 0 {
		if item2type(it) == ty {
			return slices.Index(items, id)
		}
	}
	return -1
}

func (a ActionNextItem) String() string {
	switch a.Type {
	case itemComestible:
		return "查看下一个已知食物"
	case itemTotem:
		return "查看下一个已知灵体"
	case itemMenhir:
		return "查看下一个已知石碑"
	case itemPortal:
		// Also works for the orb of corruption.
		return "查看下一个已知传送门"
	case itemRune:
		return "查看下一个已知符文陷阱"
	default:
		// Should not happen.
		return "查看下一个已知道具"
	}
}

// ActionExamine examines a target screen position (relative position ignoring
// log at the top, so starting at the map).
type ActionExamine struct {
	Target gruid.Point
}

func (a ActionExamine) Handle(md *model) (gruid.Effect, bool) {
	p := a.Target
	if inMap(p) {
		md.Examine(p)
	} else {
		md.targ.CancelExamine()
	}
	return nil, false
}

// ActionExamineModeToggle enters keyboard targeting mode.
type ActionExamineModeToggle struct{}

func (a ActionExamineModeToggle) Handle(md *model) (gruid.Effect, bool) {
	if !md.targ.kb {
		md.targ.kb = true
		md.Examine(md.g.PP())
		return nil, false
	}
	md.targ.CancelExamine()
	return nil, false
}

func (a ActionExamineModeToggle) String() string {
	return "查看（键盘模式）"
}

// ActionScroll represents a vertical scrolling action (usable for scrolling
// examine description boxes).
type ActionScroll struct {
	Delta gruid.Point
}

func (a ActionScroll) Handle(md *model) (gruid.Effect, bool) {
	switch a.Delta {
	case gruid.Point{0, 1}:
		if md.targ.p != InvalidPos {
			md.targ.scroll = 0
		}
	case gruid.Point{0, -1}:
		if md.targ.p != InvalidPos {
			md.targ.scroll++
		}
	}
	return nil, false
}

func (a ActionScroll) String() string {
	switch a.Delta {
	case gruid.Point{0, 1}:
		return "向上滚动说明框"
	default:
		return "向下滚动说明框"
	}
}

// ActionInteract interacts with items on current cell.
type ActionInteract struct{}

func (a ActionInteract) Handle(md *model) (gruid.Effect, bool) {
	g := md.g
	md.mode = modeNormal
	pa := g.PlayerActor()
	if pa.Has(StatusDaze) {
		g.logDaze()
		return nil, false
	}
	pp := g.PP()
	for i, ei := range g.NPMapEntities() {
		if ei.P != pp {
			continue
		}
		switch it := ei.Role.(type) {
		case *Comestible:
			return nil, equipItem(md, i)
		case *Spirit:
			return nil, equipItem(md, i)
		case Item:
			// Interacting with non-equippable means using them.
			return nil, it.Use(g, i)
		}
	}
	g.Log("没有可互动的目标。")
	return nil, false
}

func (a ActionInteract) String() string {
	return "互动（装备/激活）"
}

func equipItem(md *model, i ID) bool {
	g := md.g
	ei := g.Entity(i)
	spirit := false // whether the to-be-equipped item is a spirit
	var from, to ID
	switch ei.Role.(type) {
	case *Spirit:
		spirit = true
		from, to = ID(0), ID(NSpirits)
	case *Comestible:
		if g.WarnLeave(EventPickup) {
			g.md.mode = modeUseConfirmation
			return false
		}
		for i, e := range g.Entities[NSpirits:g.InventoryEnd()] {
			if !e.IsItem() {
				// Equip on first empty slot.
				return g.EquipItemAt(ID(i + NSpirits))
			}
		}
		from, to = NSpirits, FirstMapID
	}
	entries := []ui.MenuEntry{}
	r := 'a'
	for j, ej := range g.InventoryItems() {
		entries = appendMenuHeader(entries, j)
		var entry ui.MenuEntry
		entry.Disabled = j < from || j >= to
		if !entry.Disabled {
			rj, ok := ej.Role.(*Spirit)
			if ok && rj.Level >= 1 {
				// Cannot upgrade a spirit more than once.
				entry.Disabled = true
			}
			if entry.Disabled && j == from {
				from++
			}
		}
		if entry.Disabled {
			entry.Text = ui.Textf("  - %s", ej.Name).WithStyle(gruid.Style{Fg: ColorForegroundSecondary})
		} else {
			entry.Text = ui.Textf("%c - %s", r, ej.Text())
			entry.Keys = []gruid.Key{gruid.Key(r)}
		}
		entries = append(entries, entry)
		r = nextRuneKey(r)
	}
	if spirit && from == to {
		// NOTE: At some point I considered providing another kind of
		// benefit when all spirits were already upgraded, but I
		// decided against it to not encourage perfectionism.
		g.Log("没有可升级的灵体，也没有空槽位。")
		return false
	}
	altBgEntries(entries)
	md.menu.main.SetBox(&ui.Box{Title: ui.Text("物品栏（装备）").WithStyle(gruid.Style{}.WithFg(ColorYellow))})
	md.menu.main.SetEntries(entries)
	md.mode = modeMenu
	md.menu.mode = modeEquip
	ej := g.Entity(from)
	md.menu.main.SetActiveInvokable(0)
	_, ok := ej.Role.(*Spirit)
	md.drawGroundDesc = !ok
	md.updateEquipItemDesc(ei, ej)
	return false
}

// ActionInventory opens the inventory menu.
type ActionInventory struct{}

func (a ActionInventory) Handle(md *model) (gruid.Effect, bool) {
	return md.openInventory(-1)
}

// openInventory opens the inventory. The entry corresponding to the given ID
// is marked as already selected and disabled (used by GluttonyRework).
func (md *model) openInventory(bound ID) (gruid.Effect, bool) {
	entries := []ui.MenuEntry{}
	r := 'a'
	for i, ei := range md.g.InventoryItems() {
		entries = appendMenuHeader(entries, i)
		var entry ui.MenuEntry
		if bound == i {
			entry.Text = ui.Textf("* - %s", ei.Name)
			entry.Disabled = true
		} else if ei.IsItem() && (bound == -1 || IsComestible(ei.Item())) {
			entry.Text = ui.Textf("%c - %s", r, ei.TextWarn(md.g, -1)) // ID only matters if on the ground
			entry.Keys = []gruid.Key{gruid.Key(r)}
		} else {
			entry.Text = ui.Textf("  - %s", ei.Name).WithStyle(gruid.Style{Fg: ColorForegroundSecondary})
			entry.Disabled = true
		}
		entries = append(entries, entry)
		r = nextRuneKey(r)
	}
	if md.g.Mod(ModSmallInventory) {
		// Ensure we get the same letter as usual for any ground
		// comestibles.
		r += 2
	}
	if id, it := md.g.ItemAt(md.g.PP()); id >= 0 {
		if IsComestible(it) {
			entries = appendMenuHeader(entries, InventorySize)
			var entry ui.MenuEntry
			if id != bound {
				entry.Text = ui.Textf("%c - %s", r, md.g.Entity(id).TextWarn(md.g, id))
				entry.Keys = []gruid.Key{gruid.Key(r)}
			} else {
				entry.Text = ui.Textf("* - %s", md.g.Entity(id).Name)
				entry.Disabled = true
			}
			entries = append(entries, entry)
		}
	}
	altBgEntries(entries)
	if bound == -1 {
		md.menu.main.SetBox(&ui.Box{Title: ui.Text("物品栏（使用）").WithStyle(gruid.Style{Fg: ColorYellow})})
	} else {
		md.menu.main.SetBox(&ui.Box{Title: ui.Text("请选择其他食物！").WithStyle(gruid.Style{Fg: ColorYellow})})
	}
	md.menu.main.SetEntries(entries)
	md.mode = modeMenu
	if bound == -1 {
		md.menu.mode = modeInventory
	} else {
		md.menu.mode = modeInventoryBound
	}
	md.menu.bound = bound
	md.menu.main.SetActiveInvokable(0)
	var e *Entity
	if bound == -1 {
		e = md.g.Entity(0)
	} else {
		e = md.g.Entity(md.inventoryComestibleID(md.menu.main.Active()))
	}
	md.updateItemDesc(e)
	return nil, false
}

func nextRuneKey(r rune) rune {
	for {
		r++
		switch r {
		case 'h', 'l', 'j', 'k':
		default:
			return r
		}
	}
}

func appendMenuHeader(entries []ui.MenuEntry, i ID) []ui.MenuEntry {
	hstyle := gruid.Style{}.WithFg(ColorCyan)
	switch i {
	case 0:
		entries = append(entries, ui.MenuEntry{
			Text:     ui.Text("主灵体").WithStyle(hstyle),
			Disabled: true,
		})
	case 1:
		entries = append(entries, ui.MenuEntry{
			Text:     ui.Text("副灵体").WithStyle(hstyle),
			Disabled: true,
		})
	case NSpirits:
		entries = append(entries, ui.MenuEntry{
			Text:     ui.Text("食物").WithStyle(hstyle),
			Disabled: true,
		})
	case NSpirits + NComestibles:
		entries = append(entries, ui.MenuEntry{
			Text:     ui.Text("地上的食物").WithStyle(hstyle),
			Disabled: true,
		})
	}
	return entries
}

// altBgEntries updates entries to use alternate background color for entries
// of odd index. It uses appropriate markups for each entry.
func altBgEntries(entries []ui.MenuEntry) {
	for i := range entries {
		if i%2 == 1 {
			t := entries[i].Text
			st := t.Style().WithBg(ColorBackgroundSecondary)
			entries[i].Text = t.WithStyle(st).WithMarkups(MarkupsAltBg)
		} else {
			entries[i].Text = entries[i].Text.WithMarkups(Markups)
		}
	}
}

// updateItemDesc updates the description label with text for the given entity.
func (md *model) updateItemDesc(e *Entity) {
	l := md.desc
	stt := ui.StyledText{}.WithMarkups(Markups)
	l.Box = &ui.Box{Title: ui.Text(e.Name)}
	switch it := e.Role.(type) {
	case *Spirit:
		if it.Level == 1 {
			l.Box.Title = ui.Textf("%s（已升级）", e.Name)
		}
		l.Content = stt.WithText(it.Desc(md.g)).Format(descriptionPanelWidth() - 3)
	case Item:
		l.Content = stt.WithText(it.Desc(md.g)).Format(descriptionPanelWidth() - 3)
	default:
		// Should not happen.
		l.Content = stt.WithText(e.Name).Format(descriptionPanelWidth() - 3)
	}
}

func (a ActionInventory) String() string {
	return "打开物品栏"
}

func (a ActionInventory) Desc(_ *model) string {
	return "查看你的灵体和食物。选择灵体会消耗一次充能；选择食物则会将其吃掉。"
}

// ActionUseItem uses an item.
type ActionUseItem struct {
	ID ID
}

func (a ActionUseItem) Handle(md *model) (gruid.Effect, bool) {
	md.mode = modeNormal
	return nil, md.g.useItem(a.ID)
}

// useItem attempts to use the entity with given ID if it's an item. It reports
// whether the attempt was a success.
func (g *Game) useItem(id ID) bool {
	e := g.Entity(id)
	it, ok := e.Role.(Item)
	if !ok {
		return false
	}
	return it.Use(g, id)
}

// ActionBindItem chooses the first comestible for GluttonyRework.
type ActionBindItem struct {
	ID ID
}

func (a ActionBindItem) Handle(md *model) (gruid.Effect, bool) {
	return md.openInventory(a.ID)
}

// ActionUseTwoItems uses two comestibles for GluttonyRework.
type ActionUseTwoItems struct {
	ID0 ID
	ID1 ID
}

func (a ActionUseTwoItems) Handle(md *model) (gruid.Effect, bool) {
	md.mode = modeNormal
	ok := md.g.useItem(a.ID0) && md.g.useItem(a.ID1)
	if !ok {
		log.Print("错误：未能同时吃下两份食物。")
	}
	return nil, ok
}

// updateEquipItemDesc updates the label for equipping ei on ej.
func (md *model) updateEquipItemDesc(ei, ej *Entity) {
	g := md.g
	stt := ui.StyledText{}.WithMarkups(Markups)
	switch rj := ej.Role.(type) {
	case *Spirit:
		md.desc.Content = stt // empty (not drawn)
		md.equipPager.SetBox(&ui.Box{Title: stt.WithTextf("%s（升级）", ej.Name)})
		md.equipPager.SetLines(stt.WithText(rj.UpgradeDesc(g, ei)).Format(descriptionPanelWidth() - 3).Lines())
	case *Comestible:
		md.desc.Box = &ui.Box{Title: stt.WithTextf("%s（替换）", ej.Name)}
		md.desc.Content = stt.WithText(rj.Desc(g)).Format(descriptionPanelWidth() - 3)
		if ri, ok := ei.Role.(*Comestible); ok {
			md.equipPager.SetBox(&ui.Box{Title: stt.WithTextf("%s（地面）", ei.Name)})
			md.equipPager.SetLines(stt.WithText(ri.Desc(g)).Format(descriptionPanelWidth() - 3).Lines())
		}
	default:
		// Empty slot: should only happen with a spirit.
		if ri, ok := ei.Role.(*Spirit); ok {
			md.desc.Box = &ui.Box{Title: stt.WithText(ej.Name)}
			s := "@M可将地上的灵体装备到此空槽位。@N"
			md.desc.Content = stt.WithText(s).Format(descriptionPanelWidth() - 3)
			md.equipPager.SetBox(&ui.Box{Title: stt.WithTextf("%s（地面）", ei.Name)})
			md.equipPager.SetLines(stt.WithText(ri.Desc(g)).Format(descriptionPanelWidth() - 3).Lines())
		}
	}
}

// ActionEquipItemAt equips an item.
type ActionEquipItemAt struct {
	ID ID
}

func (a ActionEquipItemAt) Handle(md *model) (gruid.Effect, bool) {
	done := md.g.EquipItemAt(a.ID)
	if !done {
		md.mode = modeNormal
	}
	return nil, done
}

// ActionMenu open main game menu.
type ActionMenu struct{}

// menuActions represents the various entries in the main menu: they should
// have a corresponding entry in menuKeys.
var menuActions = []Action{
	ActionInteract{},
	ActionInventory{},
	ActionViewMessages{},
	ActionHelp{},
	ActionDump{},
	ActionConfig{},
	ActionSaveQuit{},
	ActionQuit{},
	ActionWizardNextLevel{},
}

var menuKeys = []rune{'e', 'i', 'm', '?', '#', 'C', 'S', 'Q', '»'}

func init() {
	if len(menuActions) != len(menuKeys) {
		panic("菜单操作与按键数量不匹配")
	}
}

func (a ActionMenu) Handle(md *model) (gruid.Effect, bool) {
	const nwizard = 1 // number of wizard actions
	md.menuActions = menuActions
	if !md.g.Wizard.Extra {
		md.menuActions = md.menuActions[:len(md.menuActions)-nwizard]
	}
	hstyle := gruid.Style{}.WithFg(ColorCyan)
	entries := []ui.MenuEntry{}
	for i, it := range md.menuActions {
		r := menuKeys[i]
		switch r {
		case 'e':
			entries = append(entries, ui.MenuEntry{
				Text:     ui.Text("游戏操作").WithStyle(hstyle),
				Disabled: true,
			})
		case 'm':
			entries = append(entries, ui.MenuEntry{
				Text:     ui.Text("游戏信息").WithStyle(hstyle),
				Disabled: true,
			})
		case 'C':
			entries = append(entries, ui.MenuEntry{
				Text:     ui.Text("其他操作").WithStyle(hstyle),
				Disabled: true,
			})
		case '>':
			entries = append(entries, ui.MenuEntry{
				Text:     ui.Text("向导模式操作").WithStyle(hstyle),
				Disabled: true,
			})
		}
		disabled := false
		if r == 'e' {
			// The first action (interact) isn't always available.
			if j, _ := md.g.ItemAt(md.g.PP()); j < 0 {
				disabled = true
				md.menuActions = md.menuActions[1:]
			}
		}
		if !disabled {
			entries = append(entries, ui.MenuEntry{
				Text: ui.Textf("%c - %s", r, it),
				Keys: []gruid.Key{gruid.Key(r)},
			})
		}
	}
	altBgEntries(entries)
	md.menu.main.SetBox(&ui.Box{Title: ui.Text("菜单").WithStyle(gruid.Style{}.WithFg(ColorYellow))})
	md.menu.main.SetEntries(entries)
	md.menu.main.SetActiveInvokable(0)
	md.updateMenuActionDesc(0)
	md.mode = modeMenu
	md.menu.mode = modeGameMenu
	return nil, false
}

func (a ActionMenu) String() string {
	return "菜单"
}

// ActionConfig opens settings menu.
type ActionConfig struct{}

var configActions = []Action{
	ActionSetKeys{},
	ActionToggleDarkLight{},
	ActionToggleExtraWarnings{},
}

var configKeys = []rune{':', 'c', 'w', 't'}

func (a ActionConfig) Handle(md *model) (gruid.Effect, bool) {
	md.openConfig(true)
	return nil, false
}

func (md *model) openConfig(reset bool) {
	md.menuActions = configActions
	entries := []ui.MenuEntry{}
	for i, it := range md.menuActions {
		r := configKeys[i]
		entries = append(entries, ui.MenuEntry{
			Text: ui.Textf("%c - %s", r, it),
			Keys: []gruid.Key{gruid.Key(r)},
		})
	}
	altBgEntries(entries)
	md.menu.main.SetBox(&ui.Box{Title: ui.Text("设置").WithStyle(gruid.Style{}.WithFg(ColorYellow))})
	md.menu.main.SetEntries(entries)
	if reset {
		// Reset active entry.
		md.menu.main.SetActiveInvokable(0)
	}
	if idx := md.menu.main.ActiveInvokable(); idx >= 0 {
		md.updateMenuActionDesc(idx)
	}
	md.mode = modeMenu
	md.menu.mode = modeConfigMenu
}

func (a ActionConfig) String() string {
	return "设置选项"
}

func (a ActionConfig) Desc(_ *model) string {
	return "打开设置菜单，调整多种选项。"
}

// ActionToggleDarkLight toggles dark/light mode.
type ActionToggleDarkLight struct{}

func (a ActionToggleDarkLight) Handle(md *model) (gruid.Effect, bool) {
	GameConfig.DarkColors = !GameConfig.DarkColors
	err := SaveConfig()
	if err != nil {
		log.Printf("保存设置失败：%v", err)
	}
	clearCache()
	eff := gruid.Cmd(func() gruid.Msg { return gruid.MsgScreen{} })
	md.openConfig(false)
	return eff, false
}

func (a ActionToggleDarkLight) String() string {
	if GameConfig.DarkColors {
		return "切换为浅色主题"
	}
	return "切换为深色主题"
}

func (a ActionToggleDarkLight) Desc(_ *model) string {
	if GameConfig.DarkColors {
		return "切换此选项即可改用浅色主题。\n\n@S当前设置：深色主题。@N"
	}
	return "切换此选项即可改用深色主题。\n\n@S当前设置：浅色主题。@N"
}

// ActionToggleExtraWarnings enables/disables extra warnings.
type ActionToggleExtraWarnings struct{}

func (a ActionToggleExtraWarnings) Handle(md *model) (gruid.Effect, bool) {
	GameConfig.WarningsExtra = !GameConfig.WarningsExtra
	err := SaveConfig()
	if err != nil {
		log.Printf("保存设置失败：%v", err)
	}
	clearCache()
	eff := gruid.Cmd(func() gruid.Msg { return gruid.MsgScreen{} })
	md.openConfig(false)
	return eff, false
}

func (a ActionToggleExtraWarnings) String() string {
	if GameConfig.WarningsExtra {
		return "关闭额外警告"
	}
	return "开启额外警告"
}

func (a ActionToggleExtraWarnings) Desc(_ *model) string {
	var verb, current string
	if GameConfig.WarningsExtra {
		verb, current = "关闭", "已开启"
	} else {
		verb, current = "开启", "已关闭"
	}
	return fmt.Sprintf("切换此选项可%s额外确认警告。\n\n受到@O火焰@N或@O中毒@N影响时会触发警告；某些状态（如@B狂暴@N或@B木化@N）只剩一回合时也会触发。\n\n@S当前设置：%s。@N", verb, current)
}

// ActionSetKeys opens the keymap settings.
type ActionSetKeys struct{}

// configurableAction represents actions that may be configurable so need a
// textual representation in the UI.
type configurableAction interface {
	fmt.Stringer
	Action
}

var configurableKeyActions = [...]configurableAction{
	ActionBump{Delta: gruid.Point{-1, 0}},
	ActionBump{Delta: gruid.Point{1, 0}},
	ActionBump{Delta: gruid.Point{0, -1}},
	ActionBump{Delta: gruid.Point{0, 1}},
	ActionRun{Delta: gruid.Point{-1, 0}},
	ActionRun{Delta: gruid.Point{1, 0}},
	ActionRun{Delta: gruid.Point{0, -1}},
	ActionRun{Delta: gruid.Point{0, 1}},
	ActionWait{},
	ActionCursorBump{Delta: gruid.Point{-1, 0}},
	ActionCursorBump{Delta: gruid.Point{1, 0}},
	ActionCursorBump{Delta: gruid.Point{0, -1}},
	ActionCursorBump{Delta: gruid.Point{0, 1}},
	ActionCursorRun{Delta: gruid.Point{-1, 0}},
	ActionCursorRun{Delta: gruid.Point{1, 0}},
	ActionCursorRun{Delta: gruid.Point{0, -1}},
	ActionCursorRun{Delta: gruid.Point{0, 1}},
	ActionInteract{},
	ActionInventory{},
	ActionViewMessages{},
	ActionExamineModeToggle{},
	ActionNextItem{itemComestible},
	ActionNextItem{itemTotem},
	ActionNextItem{itemMenhir},
	ActionNextItem{itemPortal},
	ActionNextItem{itemRune},
	ActionNextMonster{},
	ActionPreviousMonster{},
	ActionTravel{},
	ActionAutoExplore{},
	ActionMenu{},
	ActionHelp{},
	ActionConfig{},
	ActionSaveQuit{},
	ActionQuit{},
	ActionWizard{},
}

func (a ActionSetKeys) Handle(md *model) (gruid.Effect, bool) {
	entries := []ui.MenuEntry{}
	r := 'a'
	for _, it := range configurableKeyActions {
		desc := it.String()
		desc = fmt.Sprintf(" %-36s %s", desc, md.keysForAction(it))
		entries = append(entries, ui.MenuEntry{
			Text: ui.Text(desc),
		})
		r = nextRuneKey(r)
	}
	altBgEntries(entries)
	md.setKeyBindingsTitle()
	md.menu.keys.SetEntries(entries)
	md.menu.main.SetActiveInvokable(0)
	md.mode = modeMenu
	md.menu.mode = modeKeysView
	return nil, false
}

func (md *model) setKeyBindingsTitle() {
	if md.UsesDefaultKeybindings() {
		md.menu.keys.SetBox(&ui.Box{Title: ui.Text("按键绑定（默认）").WithStyle(gruid.Style{}.WithFg(ColorYellow))})
	} else {
		md.menu.keys.SetBox(&ui.Box{Title: ui.Text("按键绑定（自定义）").WithStyle(gruid.Style{}.WithFg(ColorYellow))})
	}
}

func (md *model) keysForAction(a Action) string {
	keys := []gruid.Key{}
	for k, action := range md.keysNormal {
		if a == action && !k.In(keys) {
			keys = append(keys, k)
		}
	}
	if KbTargetingMode(a) {
		for k, action := range md.keysTarget {
			if a == action && !k.In(keys) {
				keys = append(keys, k)
			}
		}
	}
	ks := make([]string, len(keys))
	for i := range ks {
		ki := string(keys[i])
		if ki == " " {
			ki = "空格键"
		}
		ks[i] = ki
	}
	slices.Sort(ks)
	return strings.Join(ks, " 或 ")
}

func (a ActionSetKeys) String() string {
	return "查看/自定义按键绑定"
}

func (a ActionSetKeys) Desc(md *model) string {
	var s string
	if md.UsesDefaultKeybindings() {
		s = "\n\n@S当前使用默认按键绑定。@N"
	} else {
		s = "\n\n@S当前使用自定义按键绑定。@N"
	}
	return fmt.Sprintf("打开菜单以查看和自定义按键绑定。%s", s)
}

// UsesDefaultKeybindings reports whether default keybindings are currently in
// use.
func (md *model) UsesDefaultKeybindings() bool {
	return maps.Equal(md.keysNormal, DefaultKeysNormal) &&
		maps.Equal(md.keysTarget, DefaultKeysTarget)
}

// ActionViewMessages opens the log message viewer.
type ActionViewMessages struct{}

func (a ActionViewMessages) Handle(md *model) (gruid.Effect, bool) {
	if len(md.pager.lines) > 0 {
		md.pager.lines = md.pager.lines[:len(md.pager.lines)-1]
	}
	for _, e := range md.g.Logs.Entries[len(md.pager.lines):] {
		md.pager.lines = append(md.pager.lines, md.pager.markup.WithText(e.MText))
	}
	md.pager.pg.SetLines(wrapPanelLines(md.pager.lines))
	md.pager.pg.SetCursor(gruid.Point{0, len(md.pager.lines)})
	md.pager.pg.SetBox(&ui.Box{Title: ui.Text("消息").WithStyle(gruid.Style{}.WithFg(ColorYellow))})
	md.mode = modePager
	md.pager.mode = modePagerNormal
	return nil, false
}

func (a ActionViewMessages) String() string {
	return "查看消息"
}

func (a ActionViewMessages) Desc(_ *model) string {
	return "打开分页查看器，浏览之前的消息记录。支持翻页、鼠标滚动及常见分页快捷键。"
}

// ActionDump writes a dump with game statistics.
type ActionDump struct{}

func (a ActionDump) Handle(md *model) (gruid.Effect, bool) {
	msg, err := md.g.WriteDump()
	if err != nil {
		log.Printf("写入统计文件失败：%v", err)
		md.g.LogStyled("写入统计信息失败。", logError)
	} else {
		md.g.Log(msg)
	}
	md.mode = modePager
	md.pager.mode = modePagerNormal
	dump := md.g.DumpInGame()
	stts := ui.Text(strings.TrimRight(dump, "\n")).WithMarkups(Markups).Lines()
	md.pager.pg.SetLines(wrapPanelLines(stts))
	md.pager.pg.SetCursor(gruid.Point{0, 0})
	md.pager.pg.SetBox(&ui.Box{Title: ui.Text("游戏总结").WithStyle(gruid.Style{}.WithFg(ColorYellow))})
	return nil, false
}

func (a ActionDump) String() string {
	return "游戏总结"
}

func (a ActionDump) Desc(_ *model) string {
	if runtime.GOOS == "js" {
		return "将游戏统计信息写在下方，也会在游戏内分页查看器中打开总结。"
	}
	return "将游戏统计信息写入数据目录中的 dump.txt 文件，也会在游戏内分页查看器中打开总结。"
}

// ActionSaveQuit asks for quitting the game after saving.
type ActionSaveQuit struct{}

func (a ActionSaveQuit) Handle(md *model) (gruid.Effect, bool) {
	if err := md.g.Save(); err != nil {
		log.Printf("保存游戏失败：%v", err)
		return nil, false
	}
	md.mode = modeQuitting
	return gruid.End(), false
}

func (a ActionSaveQuit) String() string {
	return "保存并退出"
}

func (a ActionSaveQuit) Desc(_ *model) string {
	return "保存当前进度并退出游戏。下次启动时将直接从此处继续。"
}

// ActionQuit asks for quitting the game, without saving.
type ActionQuit struct{}

func (a ActionQuit) Handle(md *model) (gruid.Effect, bool) {
	md.mode = modeQuitConfirmation
	md.g.LogStyled("确定不保存并退出吗？[y/N]", logConfirm)
	return nil, false
}

func (a ActionQuit) String() string {
	return "退出（不保存）"
}

func (a ActionQuit) Desc(_ *model) string {
	return "删除本轮游戏的所有存档并退出。"
}

// ActionQuitConfirm quits the game.
type ActionQuitConfirm struct {
	State confirm
}

func (a ActionQuitConfirm) Handle(md *model) (gruid.Effect, bool) {
	switch a.State {
	case confirmTrue:
		md.mode = modeQuitting
		err := RemoveSaveFile()
		if err != nil {
			log.Printf("删除存档文件失败：%v", err)
		}
		RemoveReplay()
		return gruid.End(), false
	case confirmFalse:
		md.g.Log("那就继续游戏吧。")
		md.mode = modeNormal
	}
	return nil, false
}

// Wizard collects information relative to wizard mode.
type Wizard struct {
	Mode          WizardMode // Currently active wizard mode
	Resurrections int        // Number of resurrections
	Extra         bool       // Whether to enable extra wizard cheats
}

// WizardMode describes the possible wizard modes.
type WizardMode int

const (
	WizardNone          WizardMode = iota
	WizardNormal                   // non-permadeath mode with normal visibility
	WizardReveal                   // reveal map and monsters
	WizardRevealTerrain            // reveal map (without monsters)
)

func (wm WizardMode) Reveal() bool {
	return wm == WizardReveal || wm == WizardRevealTerrain
}

// ActionWizard enters wizard-mode confirmation. When already in wizard mode,
// cycle to next wizard visibility mode. The first wizard mode is just a
// non-permadeath mode, without extra cheats (other than resurrection).
type ActionWizard struct{}

func (a ActionWizard) Handle(md *model) (gruid.Effect, bool) {
	switch md.g.Wizard.Mode {
	case WizardNone:
		md.g.LogStyled("确定进入向导模式吗？（非永久死亡模式）[y/N]", logConfirm)
		md.mode = modeWizardConfirmation
	case WizardNormal:
		if !md.g.Wizard.Extra {
			md.g.LogStyled("确定启用额外作弊功能吗？（显示地图）[y/N]", logConfirm)
			md.mode = modeWizardConfirmation
		} else {
			md.g.Wizard.Mode = WizardReveal
			md.targ.CancelExamine()
		}
	case WizardReveal:
		md.g.Wizard.Mode = WizardRevealTerrain
		md.targ.CancelExamine()
	case WizardRevealTerrain:
		md.g.Wizard.Mode = WizardNormal
		md.targ.CancelExamine()
	}
	return nil, false
}

func (a ActionWizard) String() string {
	return "进入向导模式（不可逆）"
}

// ActionWizardConfirm enters wizard-mode.
type ActionWizardConfirm struct {
	State confirm
}

func (a ActionWizardConfirm) Handle(md *model) (gruid.Effect, bool) {
	switch a.State {
	case confirmTrue:
		if md.g.Wizard.Mode == WizardNone {
			md.g.Wizard.Mode = WizardNormal
			md.g.LogStyled("向导模式已启用。", logSpecial)
			md.g.StoryLog("已进入向导模式")
		} else {
			md.g.Wizard.Extra = true
			md.g.Wizard.Mode = WizardReveal
			md.targ.CancelExamine()
			md.g.LogStyled("额外作弊功能已启用。", logSpecial)
			md.g.StoryLog("已启用向导模式额外作弊功能！")
		}
		md.mode = modeNormal
		return nil, false
	case confirmFalse:
		md.mode = modeNormal
		if md.g.Wizard.Mode == WizardNone {
			md.g.Log("那就正常继续游戏吧。")
		} else {
			md.g.Log("那就不显示地图吧。")
		}
	}
	return nil, false
}

// ActionWizardNextLevel is a debug cheat to quickly go to next level. Only
// accessible after revealing map.
type ActionWizardNextLevel struct{}

func (a ActionWizardNextLevel) Handle(md *model) (gruid.Effect, bool) {
	g := md.g
	if !g.Wizard.Extra {
		return nil, false
	}
	if g.Map.Level == MapLevels {
		g.Log("你已经到达最后一层，重新开始！")
		*g = Game{rand: g.rand, Mods: g.Mods, md: md, Wizard: g.Wizard}
		spe := spiritEntity(g.Mods, primarySpirits[g.IntN(len(primarySpirits))])
		md.targ.CancelExamine()
		md.startNewGame(spe)
		return nil, false
	}
	md.g.StoryLog("向导模式：使用“前往下一层”作弊功能")
	g.LogStyled("你使用了“前往下一层”作弊功能。", logSpecial)
	g.NextLevel()
	md.mode = modeNormal
	return nil, false
}

func (a ActionWizardNextLevel) String() string {
	return "前往下一层"
}

// actionAuto handles automatic movement of all kinds. It is triggered when
// receiving msgAuto messages. The value of thoses messages is used to ensure
// that we only use it on a specific turn, or discard it, which should be
// normally the case, but better be safe than sorry.
type actionAuto struct {
	msg msgAuto
}

type msgAuto int

func (a actionAuto) Handle(md *model) (gruid.Effect, bool) {
	g := md.g
	if int(a.msg) != g.Turn {
		return nil, false
	}
	switch md.auto.mode {
	case autoRun:
		if !md.KeepRunning() {
			md.StopAuto()
			return nil, false
		}
		bump := ActionBump{Delta: md.auto.delta}
		eff, done := bump.Handle(md)
		if !done {
			md.StopAuto()
			return eff, false
		}
		md.UpdateRun(md.auto.delta)
		return md.UpdateAutoMode(eff, autoRun), true
	case autoTravel:
		if !md.KeepTraveling() {
			md.StopAuto()
			return nil, false
		}
		next := md.auto.path[1]
		bump := ActionBump{Delta: next.Sub(g.PP())}
		eff, done := bump.Handle(md)
		if !done {
			md.StopAuto()
			return eff, false
		}
		if g.PP() == next {
			md.auto.path = md.auto.path[1:]
		}
		return md.UpdateAutoMode(eff, autoTravel), true
	case autoExplore:
		n, ok := md.UpdateAutoExplore()
		if !ok {
			md.StopAuto()
			return nil, false
		}
		bump := ActionBump{Delta: n.Sub(g.PP())}
		eff, done := bump.Handle(md)
		if !done {
			md.StopAuto()
			return eff, false
		}
		return md.UpdateAutoMode(eff, autoExplore), true
	}
	return nil, false
}

func (md *model) autoCmd() gruid.Effect {
	n := md.g.Turn
	return gruid.Cmd(func() gruid.Msg {
		t := time.NewTimer(AnimDurShort)
		<-t.C
		return msgAuto(n + 1)
	})
}

func (a actionAuto) String() string {
	return "自动行动消息处理"
}

// KbTargetingMode reports whether the action is used in keyboard targeting mode.
func KbTargetingMode(a Action) bool {
	switch a.(type) {
	case ActionCursorBump, ActionCursorRun, ActionTravel:
		return true
	}
	return false
}

// ActionHelp enters a help menu.
type ActionHelp struct{}

// helpActions represents the various entries in the main help: they should
// have a corresponding entry in helpKeys.
var helpActions = []Action{
	ActionHelpDefaultKeys{},
	ActionHelpCombat{},
	ActionHelpStealth{},
	ActionHelpItems{},
	ActionHelpStatuses{},
	ActionHelpMods{},
	ActionHelpTips{},
}

var helpKeys = []rune{'?', 'c', 's', 'i', 'S', 'm', 't'}

func init() {
	if len(helpActions) != len(helpKeys) {
		panic("帮助操作与按键列表长度不匹配")
	}
}

func (a ActionHelp) Handle(md *model) (gruid.Effect, bool) {
	md.menuActions = helpActions
	entries := []ui.MenuEntry{}
	for i, it := range md.menuActions {
		r := helpKeys[i]
		entries = append(entries, ui.MenuEntry{
			Text: ui.Textf("%c - %s", r, it),
			Keys: []gruid.Key{gruid.Key(r)},
		})
	}
	altBgEntries(entries)
	md.menu.main.SetBox(&ui.Box{Title: ui.Text("帮助菜单").WithStyle(gruid.Style{}.WithFg(ColorYellow))})
	md.menu.main.SetEntries(entries)
	md.menu.main.SetActiveInvokable(0)
	md.updateMenuActionDesc(0)
	md.mode = modeMenu
	md.menu.mode = modeHelpMenu
	return nil, false
}

func (a ActionHelp) String() string {
	return "帮助"
}

func (a ActionHelp) Desc(_ *model) string {
	return "打开帮助菜单，查看各主题说明。"
}

// ActionHelpDefaultKeys enters a help menu.
type ActionHelpDefaultKeys struct{}

func (a ActionHelpDefaultKeys) Handle(md *model) (gruid.Effect, bool) {
	md.KeysHelp()
	return nil, false
}

func (a ActionHelpDefaultKeys) String() string {
	return "按键绑定（默认）"
}

func (a ActionHelpDefaultKeys) Desc(_ *model) string {
	return "显示一页简明说明，列出大多数默认按键绑定。"
}

func (md *model) updateKeysDescription(title string, actions []string) {
	md.pager.mode = modePagerNormal
	md.mode = modePager
	title = fmt.Sprintf(" %s ", title)
	md.pager.pg.SetBox(&ui.Box{Title: ui.Text(title).WithStyle(gruid.Style{}.WithFg(ColorYellow))})
	lines := []ui.StyledText{}
	for i := 0; i < len(actions)-1; i += 2 {
		stt := ui.StyledText{}
		if actions[i+1] != "" {
			stt = stt.WithTextf(" %-36s %s", actions[i], actions[i+1])
		} else {
			stt = stt.WithTextf(" %s ", actions[i]).WithStyle(gruid.Style{}.WithFg(ColorCyan))
		}
		if i%4 == 2 {
			stt = stt.WithStyle(stt.Style().WithBg(ColorBackgroundSecondary))
		}
		lines = append(lines, stt)
	}
	md.pager.pg.SetLines(wrapPanelLines(lines))
	md.pager.pg.SetCursor(gruid.Point{0, 0})
}

func (md *model) KeysHelp() {
	runKeys := "Shift+方向键或HJKL"
	if runtime.GOOS == "js" {
		runKeys = "HJKL"
	}
	entries := []string{
		"基本操作", "",
		"移动或攻击", "方向键、hjkl 或鼠标左键",
		"等待一回合", "“.”、回车键，或在 @ 上单击鼠标左键",
		"互动（装备/激活/…）", "e",
		"使用物品栏道具", "i",
		"切换键盘查看模式", "x",
		"打开菜单", "空格键或鼠标右键",
		"关闭菜单、物品栏…", "空格键、ESC 或在窗口外单击鼠标左键",
		"进阶操作", "",
		"查看之前的消息", "m",
		"上下滚动说明框", "u/d、PageUp/PageDown 或鼠标滚轮",
		"查看下一个传送门/石碑/符文", "> & =",
		"查看下一个图腾/食物", "! %",
		"查看上一个/下一个怪物", "- +",
		"奔跑（沿指定方向自动移动）", runKeys,
		"旅行（自动前往目标地点）", "键盘查看模式下按“.”或回车键",
		"自动探索（请谨慎使用）", "o",
		"其他常用操作", "",
		"保存并退出", "S",
		"退出（不保存）", "Q",
	}
	md.updateKeysDescription("按键绑定（默认）", entries)
}

// helpTopic opens the pager with the given title and text content.
func (md *model) helpTopic(title, content string) {
	text := ui.Text(content).WithMarkups(Markups).Format(TextPanelWidth - 2)
	stts := text.Lines()
	md.pager.pg.SetLines(wrapPanelLines(stts))
	md.pager.pg.SetCursor(gruid.Point{0, 0})
	md.mode = modePager
	md.pager.mode = modePagerNormal
	md.pager.pg.SetBox(&ui.Box{Title: ui.Text(title).WithStyle(gruid.Style{}.WithFg(ColorYellow))})
}

// ActionHelpCombat opens the log message viewer.
type ActionHelpCombat struct{}

func (a ActionHelpCombat) Handle(md *model) (gruid.Effect, bool) {
	md.helpTopic("战斗（帮助）", combatHelpText())
	return nil, false
}

func (a ActionHelpCombat) String() string {
	return "战斗"
}

func (a ActionHelpCombat) Desc(_ *model) string {
	return "介绍战斗机制，包括攻击模式和命中效果。"
}

func combatHelpText() string {
	return strings.TrimSpace(`
@Y回合制系统@N

Shamogu 是一款回合制游戏。每个行动者每回合行动一次：玩家先行动，然后是怪物。怪物行动顺序无法预测；除动画期间外，你通常只会看到最终状态，感觉它们像是同时行动。

@Y战斗伤害@N

战斗伤害受攻击力（A）和防御力（D）影响。最大伤害为攻击力 A 与 3 中的较小值。攻击力越高，攻击时造成较高伤害的概率越大；防御力越高，防御时受到较高伤害的概率越低。伤害为 0 表示未命中。

处于@B专注@N状态的玩家和四头九头蛇每次攻击都会进行四次伤害判定，造成四倍伤害。

@B狂暴@N或@B掘地@N等特殊效果可使有效总攻击伤害再增加 1，并无视防御及伤害上限。

战斗伤害会扣除行动者的生命值（HP）。生命值归零即死亡。生命值不会自动恢复，因此怪物不会恢复已损失的生命。玩家可以通过食用食物或穿过传送门恢复生命。

@Y当前方向@N

状态栏中的箭头表示你上次移动或攻击的方向。某些效果和能力会使用此方向，例如食用吐火椒、冲刺或跳跃。当前方向是玩家专属的机制。

@Y攻击模式@N

Shamogu 中的玩家和怪物都可能使用多种攻击模式。

@C普通@N攻击是部分怪物使用的基础攻击。它们会攻击相邻目标，也可能从相邻位置发起@M正面冲锋@N。

@C四向@N攻击会影响四个相邻方向上的敌人。冲锋时，则会攻击新位置周围的敌人。因此，横向移动无法躲开相邻四向攻击怪物的冲锋！同时对付多个敌人时，四头九头蛇玩家的攻击属性会随敌人数增加。

@C狂奔@N攻击可从视野范围内任意距离发起冲锋。冲锋攻击力 +1，但木化敌人抵抗强制移动。

@C推撞@N攻击会将敌人推开一格并使其失衡。若敌人身后还有其他敌人，则会改为额外攻击后者。狂奔野猪及其灵体从两格以外冲锋时会推撞；距离超过四格时，@O失衡@N持续时间加倍。地龙和使用@B掘地@N的玩家也可通过近战推撞。

@C远程@N攻击无需移动即可攻击视野范围内的目标。

@C抓取@N攻击属于远程攻击，@M成功命中@N时会抓住敌人并将其拉到身边，使其失衡。从远处抓取时攻击力 +1；距离超过四格时，@O失衡@N持续时间加倍。

@C后撤@N是风狐怪物及其灵体的额外攻击效果。从四格内发起攻击后，远程攻击可能使其后退一格；近战攻击则可能后退两格并制造更大的风声。

@C空间扭曲@N攻击属于远程攻击，@M命中时会交换双方位置@N，未命中时则正常移动。落点相邻的敌人会被闪烁传送走。此类攻击会反转你的当前方向，也会交换云雾位置，但不会改变地形。

@C拖拽@N攻击属于近战攻击，会将敌人向后拖动并使其失衡。拖拽鳄鱼和旋转鳄鱼玩家会使用此攻击。

@C潜袭@N由两个阶段组成。远程时相当于狂奔攻击；近战时则先普通攻击，再迅速后退两格。巨蝠和吸血蝙蝠玩家会使用潜袭攻击。

@M注意：@N鳄鱼的拖拽和蝙蝠的后退不会在你@O中毒@N或它们会把你移入可见危险火焰时触发；若你已经身处@O火焰@N中，或受@B雾肤@N保护，则不受此限制。野猪的推撞冲锋和风狐的后撤没有安全限制，不过后撤属于非自愿移动，不受中毒影响。使用@B掘地@N进行近战推撞时，若需要移动玩家则受中毒限制；但若将怪物撞上墙壁或其他怪物，失衡和穿刺效果仍会触发。

@Y命中效果@N

多种怪物的攻击命中玩家后可能施加状态效果。

有些效果命中后必定触发，例如火焰羊驼从远处喷火，或闪烁蝴蝶使你闪烁传送。

说明中的“可能”表示命中后有一定概率施加状态。多数情况下，概率随伤害增加。凤凰的点燃概率还会受失衡影响，因为它们的大翅膀难以重新平衡。猫的空间扭曲攻击造成眩晕的概率则随距离变化：距离越远，概率越低、持续时间越长。

后续命中可刷新负面状态的持续时间，但不能刷新@B狂暴@N或@B木化@N等正面状态。感谢灵体祖先的庇佑，玩家受到的后续命中状态施加概率会减半，直到状态结束后的回合；对@O眩晕@N、@O火焰@N和@O失衡@N，此保护还会额外延长一回合。
`)
}

// ActionHelpStealth opens the log message viewer.
type ActionHelpStealth struct{}

func (a ActionHelpStealth) Handle(md *model) (gruid.Effect, bool) {
	md.helpTopic("潜行（帮助）", stealthHelpText())
	return nil, false
}

func (a ActionHelpStealth) String() string {
	return "潜行"
}

func (a ActionHelpStealth) Desc(_ *model) string {
	return "介绍潜行机制，包括怪物心智状态与噪音。"
}

func stealthHelpText() string {
	return strings.TrimSpace(`
Shamogu 中，潜行十分重要。怪物行为主要受两种机制影响：是否进入视野，以及噪音。

@Y怪物心智状态@N

怪物可能处于@O游荡@N、@O守卫@N或@R追猎@N状态。守卫怪物会在某个地城区域附近巡逻，通常守护着图腾或传送门。游荡怪物会走得更远，可能前往地图上的任意位置，但倾向于附近有趣的地点。追猎怪物通常会前往最后一次看见你的地点；失去你的踪迹后，它们会转为游荡或守卫，并短暂搜寻附近区域，之后恢复原有行为。

进入视野时，游荡和守卫怪物通常会花一回合注意到玩家并转为追猎状态，但也有例外。例如，炫目斑马玩家会立刻被察觉。饥饿鼠从远处闻到你，或巨蝠听到你时，也会立即开始追猎，即使它们仍在视野之外。此外，怪物有时会从视野外发起伏击冲锋，例如从碎石后冲出，攻击后立即开始追猎。

@Y噪音@N

战斗及某些其他行动会产生噪音，游戏日志会用拟声词表示。部分噪音只能传到几格外，另一些最远可传至视野范围的两倍。噪音可能只是轻轻的“砰！”，也可能是响亮的“咆哮！”或“咔！”、麻烦的“轰隆！”，甚至更令人警觉的“叮铃！”。怪物会前往调查噪音来源。

Shamogu 还会用视觉符号显示视野外怪物发出的声音：@R脚步声@N、@O轻微脚步声@N、@G窸窣声@N、@C拍翼声@N和@M沉重的脚步声@N。怪物听不见你的脚步声，或者至少它们不在意这么轻微的声音！

拥有“听觉敏锐”特性的玩家和怪物能听到更远处的声音。拥有“听觉迟钝”特性的玩家听不清@G窸窣声@N和@C拍翼声@N，但仍能正常听见其他声音。

@Y时限@N

在同一层停留超过一千回合后，地城核心会发现你的位置，并让怪物在下次行动时追猎你；还会派来一对新的烈焰魔像。别磨蹭太久！
`)
}

// ActionHelpItems opens the log message viewer.
type ActionHelpItems struct{}

func (a ActionHelpItems) Handle(md *model) (gruid.Effect, bool) {
	md.helpTopic("物品（帮助）", itemHelpText())
	return nil, false
}

func (a ActionHelpItems) String() string {
	return "物品"
}

func (a ActionHelpItems) Desc(_ *model) string {
	return "概述游戏中的各类物品。"
}

func itemHelpText() string {
	return strings.TrimSpace(`
@C食物@N每层都会出现。站在食物上时，按@M互动键@N（e）即可拾取。物品栏已满时，你需要选择一份食物留在地上。打开@M物品栏@N（i）即可吃掉已装备的食物或地上的任意食物。

@C图腾@N每层出现一次。按@M互动键@N（e）可以选择其中的灵体作为新的副灵体，或用于升级现有灵体。之后打开@M物品栏@N（i）即可使用已选灵体。注意：看见你选择灵体的怪物会陷入狂暴！

@C石碑@N每层都会出现。按@M互动键@N（e）即可激活。

@C传送门@N每层出现一次。按@M互动键@N（e）即可前往下一层，并补满灵体充能和生命值。偶尔还会遇到额外的故障传送门，激活后会吓跑附近的怪物。

@C符文@N是每层都会出现的固定陷阱。玩家或怪物踩上去时会触发。游荡中的怪物会避开符文，追猎中的怪物则不会。
`)
}

// ActionHelpMods opens the log message viewer.
type ActionHelpMods struct{}

func (a ActionHelpMods) Handle(md *model) (gruid.Effect, bool) {
	title := "模组（帮助）"
	n := md.g.modCount()
	if n == 0 {
		md.helpTopic(title, "你尚未启用任何扩展或挑战模组，当前游玩的是基础游戏。\n\n可在模组选项菜单启用模组；在新游戏菜单按 TAB 或空格键即可打开该菜单。")
		return nil, false
	}
	var sb strings.Builder
	if n == 1 {
		sb.WriteString("你已启用以下模组：")
	} else {
		fmt.Fprintf(&sb, "你已启用以下 %d 个模组：", n)
	}
	for _, m := range GameMods {
		if !md.g.Mod(m) {
			continue
		}
		fmt.Fprintf(&sb, "\n\n@Y%s@N\n\n%s", m.String(), m.Desc())
		if s := modExtraInfo(m); s != "" {
			sb.WriteString("\n\n@C补充说明@N\n\n")
			sb.WriteString(s)
		}
	}
	md.helpTopic(title, sb.String())
	return nil, false
}

func modExtraInfo(m Mod) string {
	switch m {
	case ModCorruptedDungeon:
		return `启用“腐化地城”后，怪物的游荡或守卫行为会被隐藏。大多数时候，它们的职责与基础游戏相同，但你偶尔需要根据实际行动猜测其行为。

游戏中可能出现多种主题楼层，出现概率各不相同，内容也可能相当出人意料！

有时会发生地图腐化事件。尤其要注意，整层可能同时生成大量持续较久的雾气。不过，事件也会减少地图上的植物和瓦砾，因此视野会先暂时缩小，随后逐步扩大。`
	case ModAdvancedSpirits:
		return `部分高级副灵体特性较难理解或运用。

斑马的“炫目”特性会把指向你的攻击转移给你另一侧紧邻的怪物，因此需要采用特别的站位策略。此外，斑马会让怪物立刻发现你，不会让怪物损失一个回合；探索时要小心突然遭到远程攻击。

狮子会在初次看见怪物时咆哮，但若你先在视野外通过@B清明@N感知到怪物，就能避免咆哮。恐惧持续 4 至 5 回合，距离较近时更可能持续 4 回合，因为危险越近，怪物越快克服恐惧。

獾会使基础伤害超过 2 的攻击伤害降低 1 点，也会使造成 2 点基础伤害的攻击中约三分之二降低 1 点伤害，但不会影响命中率。

熊会忽略食用稀有食物的特殊要求。

大象处于死路，或獾站在石碑格上时，游荡怪物不会注意到你，但追猎中的怪物仍能看见你。

鸡不会触发陷阱，规则看似简单；但由于游荡怪物会避开陷阱，这也会带来一些微妙的潜行影响。

@M注意：@N若未与“图腾条件”同时启用，此模组会让普通灵体在游戏前期比高级灵体更稀有，使你更难跳过高级灵体。`
	case ModTotemConditions:
		return `选择灵体时请多加留意：升级时，某些图腾诅咒可能替换已有诅咒；甚至任何升级都可能施加诅咒（包括主灵体升级）！仔细观察惩罚的性质：限制越困难，失败时的惩罚往往越轻。

不必太担心失去一两个灵体：每局都会额外生成一个非空图腾，因此多出的图腾比完成构筑所需数量多两个。此外，灵体离开时，时间会冻结数回合，让你有喘息和调整的机会。

要求“完成整层”的诅咒不包括你获得诅咒所在的楼层。“一回合移动超过一格”的诅咒只检查整回合结束后的最终位置，因此要小心青蛙从远处追上你，或蝴蝶将你闪现带走！

@C条件食物@N是只能在特定情况下食用的特殊食物。有些是此模组独有的（牡蛎、土豆、栗子），另一些在基础游戏中也有普通版本（橡子、豆子、菌类、玫瑰、蜗牛）。条件食物比基础游戏中的稀有食物更常见，尤其在游戏后期。

@M注意：@N与“高级灵体”同时启用时，条件食物在游戏前期不再比普通食物更常见，而是有时会带来较轻的诅咒，使其更有吸引力。此外，启用“无充能”时，部分诅咒会被禁用或调整以维持平衡。`
	default:
		return ""
	}
}

func (a ActionHelpMods) String() string {
	return "模组"
}

func (a ActionHelpMods) Desc(_ *model) string {
	return "列出已启用的模组及其说明。"
}

// ActionHelpStatuses opens the log message viewer.
type ActionHelpStatuses struct{}

func (a ActionHelpStatuses) Handle(md *model) (gruid.Effect, bool) {
	md.helpTopic("状态（帮助）", statusesHelpText(md.g))
	return nil, false
}

func (a ActionHelpStatuses) String() string {
	return "状态"
}

func (a ActionHelpStatuses) Desc(_ *model) string {
	return "详细说明各种状态效果及颜色标记的含义。"
}

func statusesHelpText(g *Game) string {
	var sb strings.Builder
	sb.WriteString(`@Y颜色说明@N

状态栏和说明中，正面状态以@B蓝色@N显示，负面状态以@O橙色@N显示。状态会持续若干回合，回合数标在括号中。状态在你的回合结束、怪物行动之前到期，因此有时需要留意状态即将结束。持续的最后一回合中，若正面状态在你和怪物行动时均有效，则显示为@C青色@N；若只在你的回合生效，则显示为@V紫色@N；若在你的回合不生效，则显示为@S灰色@N。

地图上怪物的颜色可能随其当前状态改变。默认情况下，无状态的游荡怪物为@O橙色@N，正在追猎的怪物为@R红色@N。@M洋红色@N表示正面状态；@Y黄色@N表示同时具有正面和负面状态；@G绿色@N表示负面状态；@C青色@N表示具有两种或更多负面状态。`)
	sb.WriteString("\n\n")
	sb.WriteString(`@Y状态列表@N`)
	sb.WriteString("\n\n")
	pa := g.PlayerActor()
	for i, desc := range statusDesc {
		st := Status(i)
		if st == StatusGluttony && GluttonyRework {
			continue
		}
		color := statusColor(st, 42)
		fmt.Fprintf(&sb, "@%c%s@N（@%c%s@N）。", color, st.Name(), color, g.StatusAbbr(pa, st))
		sb.WriteString(desc)
		details := st.Details()
		if details != "" {
			sb.WriteRune(' ')
			sb.WriteString(details)
		}
		sb.WriteString("\n\n")
	}
	return sb.String()
}

// ActionHelpTips opens the log message viewer.
type ActionHelpTips struct{}

func (a ActionHelpTips) Handle(md *model) (gruid.Effect, bool) {
	md.helpTopic("提示（帮助）", tipsHelpText())
	return nil, false
}

func (a ActionHelpTips) String() string {
	return "提示"
}

func (a ActionHelpTips) Desc(_ *model) string {
	return "为新玩家提供一些实用提示。"
}

func tipsHelpText() string {
	return strings.TrimSpace(`
@Y一般提示@N

Shamogu 是一款@M回合制@N游戏：遇到危险时不必慌张，可以从容思考！

经常检查物品栏，看看灵体能力或食物能否帮你应对当前局面。通常越早使用越好！

浪费灵体充能和食物固然可惜，但带着它们死去更糟！

多数物品和能力都有多种效果与用途。例如，吠叫或冻结时间既可用于攻击，也可用于逃跑；冻结时间还可以拖到棘手状态结束。

@Y战斗提示@N

尽量先于怪物出手，争取@M先攻@N。尤其要避免走到远程或狂暴怪物的正前方，必要时可以曲折移动。别忘了，即使普通近战怪物也能从一格外正面冲锋！

熟悉各类怪物的特点。检查怪物可以查看其攻击、防御和特殊特性。

攻击值为 N 的怪物，单次最多造成 N 与 3 中较小的伤害；燃烧凤凰和远程喷火羊驼可能使你着火，四头九头蛇每次攻击则会进行四次伤害判定。

有些远程怪物，例如喷火羊驼或抽打青蛙，远程攻击比近战更危险。即使你也有远程攻击，有时仍应设法贴近它们！

许多怪物拥有特殊攻击。例如酸液丘会无视防御，毒蛇可能使你中毒。

@Y潜行提示@N

杀光所有怪物并无额外好处。食物充足、不再需要图腾，或已经找到图腾后，可以不探索地图剩余部分，直接通过传送门前往下一层。漏掉一两个图腾也没关系。

怪物看不见你后通常很快就会放弃搜索；利用植物和瓦砾摆脱追踪。不过要小心会凭气味追猎你的老鼠。

留意@M脚步声@N：穿过茂密植物时，控制脚步声有助于避免不必要的遭遇。

@Y配置提示@N

如果你常被@O火焰@N或@O毒素@N打个措手不及，可以在配置中启用额外警告。状态@B狂暴@N或@B木化@N即将结束时，这些警告也会提醒你。

@Y向导模式@N

建议通过复盘并享受每次失败，逐步提升技巧来认识游戏。如果你因此感到沮丧，可以在放弃前尝试启用@M向导模式@N（W）。无需开启其他作弊功能，它就像一种“冒险模式”，每次死亡都会让你复活。复活记录在游戏时间线中，你可以查看消耗了几条命，并继续尝试，直到能在默认的永久死亡模式中获胜。

@Y挑战提示@N

如果你觉得 Shamogu 太简单，可以在新游戏菜单按 TAB 或空格键，试试各种扩展和挑战模组！
`)
}
