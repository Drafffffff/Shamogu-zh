// This file defines the Update method for the model.

package main

import (
	"log"
	"strings"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
	"codeberg.org/anaseto/gruid/ui"
)

// Update implements Update() for gruid.Model.
func (md *model) Update(msg gruid.Msg) gruid.Effect {
	if _, ok := msg.(gruid.MsgInit); ok {
		return md.init()
	}
	if _, ok := msg.(gruid.MsgQuit); ok {
		if md.mode != modeQuitting {
			// Save game before quitting.
			if err := md.g.Save(); err != nil {
				log.Printf("退出前保存时出错：%v", err)
			}
		}
		md.mode = modeQuitting
		return gruid.End()
	}
	//log.Printf("msg %+v", msg)
	md.anims.draw = false
	if msg, ok := msg.(msgAnim); ok {
		if int(msg) != md.anims.idx {
			// The animation message is from a finished animation
			// sequence, so we just skip it.
			return nil
		}
		if !md.anims.Done() {
			md.anims.draw = true
			return md.AnimNext()
		}
		md.anims.Finish()
		return nil
	}
	anims := !md.anims.Done()
	if md.interrupt(msg) {
		if anims {
			md.anims.Finish()
			anims = false
		}
		md.StopAuto()
	}
	eff := md.update(msg)
	if cmd := md.AnimCmd(); !anims && cmd != nil {
		return gruid.Batch(eff, cmd)
	}
	return eff
}

func (md *model) interrupt(msg gruid.Msg) bool {
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		return true
	case gruid.MsgMouse:
		return msg.Action != gruid.MouseMove && msg.Action != gruid.MouseRelease
	}
	return false
}

func (md *model) update(msg gruid.Msg) gruid.Effect {
	md.action = ActionNone{}
	md.g.instant = false
	switch md.mode {
	case modeIntro:
		if md.interrupt(msg) {
			md.openSpiritSelectionMenu(modeNewGame)
		}
		return nil
	case modeLoadGame:
		switch msg := msg.(type) {
		case gruid.MsgKeyDown:
			md.updateStatus()
			md.mode = modeNormal
		case gruid.MsgMouse:
			if msg.Action != gruid.MouseMove && msg.Action != gruid.MouseRelease {
				md.updateStatus()
				md.mode = modeNormal
			}
		}
	case modeNewGame:
		return md.updateSpiritSelectionMenu(msg)
	case modeNewGameMods:
		return md.updateModSelectionMenu(msg)
	case modeNormal:
		md.updateNormal(msg)
	case modeCritical:
		if md.more(msg) {
			md.mode = modeNormal
			md.g.Log("好了，小心行事。")
		}
	case modePager:
		return md.updatePager(msg)
	case modeMenu:
		md.updateMenu(msg)
	case modeWizardConfirmation:
		st := md.updateConfirmation(msg)
		md.action = ActionWizardConfirm{State: st}
	case modeUseConfirmation:
		st := md.updateConfirmation(msg)
		switch st {
		case confirmTrue:
			md.g.confirm = true
			md.action = ActionInteract{}
		case confirmFalse:
			md.g.Log("那就别碰它了。")
			md.mode = modeNormal
		}
	case modeQuitConfirmation:
		st := md.updateConfirmation(msg)
		md.action = ActionQuitConfirm{State: st}
	case modeQuitting:
		// Do nothing.
		return nil
	case modeEnd:
		if md.more(msg) {
			md.gameEnded = true
			md.mode = modePager
			md.pager.mode = modePagerEnd
			md.dump(md.g.WriteDump())
		}
	}
	pa := md.g.PlayerActor()
	md.g.UpdateFirePos(PlayerID, pa)
	TotemEvents(md.g, &EventHappened{EvType: EventPlayerTurn})
	xhp := pa.expectedHP()
	eff, done := md.action.Handle(md)
	if done {
		md.endTurn(xhp)
		md.RefreshExamineInfo()
	}
	md.updateStatus()
	return eff
}

// expectedHP returns the expected HP at the end of the turn, assuming no
// external interference or any healing actions.
func (a *Actor) expectedHP() int {
	hp := a.HP
	if a.Statuses[StatusBerserk] == 1 {
		hp = max(1, hp-a.HPBonus())
	}
	if a.Statuses[StatusLignification] == 1 {
		hp = max(min(3, hp), hp-a.HPBonus())
	}
	return hp
}

func (md *model) more(msg gruid.Msg) bool {
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		switch msg.Key {
		case gruid.KeyEscape, gruid.KeySpace:
			return true
		}
	case gruid.MsgMouse:
		if msg.Action == gruid.MouseSecondary ||
			msg.Action == gruid.MouseMain && (msg.P.Y < 2 || msg.P.Y == UIHeight-1) {
			return true
		}
	}
	return false
}

// confirm describes the possible results a confirmation update on message.
type confirm int

const (
	confirmTrue confirm = iota
	confirmFalse
	confirmPass
)

func (md *model) updateConfirmation(msg gruid.Msg) confirm {
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		if msg.Key == "y" || msg.Key == "Y" {
			return confirmTrue
		}
		return confirmFalse
	case gruid.MsgMouse:
		if msg.Action == gruid.MouseSecondary ||
			msg.Action == gruid.MouseMain && (msg.P.Y < 2 || msg.P.Y == UIHeight-1) {
			return confirmFalse
		}
	}
	return confirmPass
}

func (md *model) updateNormal(msg gruid.Msg) {
	switch msg := msg.(type) {
	case msgAuto:
		md.action = actionAuto{msg: msg}
	case gruid.MsgKeyDown:
		md.updateKeyDown(msg)
	case gruid.MsgMouse:
		md.updateMouse(msg)
	}
}

func (md *model) updateKeyDown(msg gruid.MsgKeyDown) {
	md.status.focus = false
	if !md.targ.kb {
		if inMap(md.targ.p) {
			if msg.Key != gruid.KeyPageDown && msg.Key != gruid.KeyPageUp {
				md.targ.CancelExamine()
			}
		} else if msg.Key == gruid.KeyPageDown || msg.Key == gruid.KeyPageUp {
			md.status.menu.Update(msg)
			return
		}
	}
	a := md.keysNormal[msg.Key]
	if md.targ.kb {
		b := md.keysTarget[msg.Key]
		if b != nil {
			a = b
		}
	}
	shift := msg.Mod&gruid.ModShift != 0
	if shift && !msg.Key.IsRune() {
		switch b := a.(type) {
		case ActionBump:
			a = ActionRun(b)
		case ActionCursorBump:
			a = ActionCursorRun(b)
		}
	}
	if a != nil {
		md.action = a
	}
}

func (md *model) updateMouse(msg gruid.MsgMouse) {
	if msg.P.Y == UIHeight-1 {
		md.updateStatusMouse(msg)
		return
	}
	md.status.focus = false
	md.status.menu.SetActive(0)
	p := msg.P.Add(gruid.Point{0, -2}) // relative position ignoring log
	switch msg.Action {
	case gruid.MouseWheelUp:
		md.action = ActionScroll{Delta: gruid.Point{0, 1}}
	case gruid.MouseWheelDown:
		md.action = ActionScroll{Delta: gruid.Point{0, -1}}
	case gruid.MouseMove:
		if md.auto.mode == noAuto && !md.targ.kb {
			md.action = ActionExamine{Target: p}
		}
	case gruid.MouseSecondary:
		if !md.targ.kb {
			md.targ.CancelExamine()
		}
		md.action = ActionMenu{}
	case gruid.MouseMain:
		if !inMap(p) {
			if p.Y >= -2 && p.Y <= 0 {
				md.action = ActionViewMessages{}
			}
			return
		}
		pp := md.g.PP()
		if md.targ.p != p {
			md.Examine(p)
			return
		}
		if paths.DistanceManhattan(p, pp) == 1 {
			md.action = ActionBump{Delta: p.Sub(pp)}
		} else {
			md.action = ActionTravel{}
		}
	}
}

func (md *model) updateStatusMouse(msg gruid.MsgMouse) {
	msg.P.Y = 0
	msg.P.X = max(0, msg.P.X-1)
	if !msg.P.In(md.status.menu.Bounds()) || md.targ.kb {
		md.status.focus = false
		if md.targ.kb {
			return
		}
		md.targ.CancelExamine()
		switch msg.Action {
		case gruid.MouseWheelDown, gruid.MouseWheelUp:
			op := msg.P
			msg.P.X = 0
			md.status.menu.Update(msg)
			msg.P = op
			if !msg.P.In(md.status.menu.Bounds()) {
				return
			}
			// After scrolling, mouse is within the status bar, so
			// an extra update is needed.
			msg.Action = gruid.MouseMove
		default:
			return
		}
	}
	page := md.status.menu.Page()
	md.status.menu.Update(msg)
	update := !md.status.focus
	switch md.status.menu.Action() {
	case ui.MenuMove:
		if page != md.status.menu.Page() {
			if !msg.P.In(md.status.menu.Bounds()) {
				// After scrolling, mouse is outside the status bar.
				md.status.focus = false
				return
			}
		}
		update = true
	case ui.MenuInvoke:
		i := statusEntry(md.status.menu.Active())
		switch i {
		case statusMenu:
			md.status.focus = false
			md.action = ActionMenu{}
			return
		default:
			update = true
		}
	}
	if update {
		md.targ.CancelExamine()
		const statusIndex = statusDefense + 1
		i := statusEntry(md.status.menu.Active())
		md.status.focus = false
		switch {
		case i == statusLevel:
			md.status.desc.Box = &ui.Box{Title: ui.Text("层数")}
			md.status.desc.SetText("当前地图层数。")
			md.status.focus = true
		case i == statusTurns:
			md.status.desc.Box = &ui.Box{Title: ui.Text("回合")}
			md.status.desc.SetText("游戏开始后经过的回合数。")
			md.status.focus = true
		case i == statusMenu:
			md.status.desc.Box = &ui.Box{Title: ui.Text("菜单")}
			md.status.desc.SetText("点击打开菜单，或按空格键、M 键。")
			md.status.focus = true
		case i == statusDirection:
			md.status.desc.Box = &ui.Box{Title: ui.Text("方向")}
			md.status.desc.SetText("上次攻击或移动的方向。")
			md.status.focus = true
		case i == statusHP:
			md.status.desc.Box = &ui.Box{Title: ui.Text("生命")}
			md.status.desc.SetText("你的生命值。")
			md.status.focus = true
		case i == statusAttack:
			md.status.desc.Box = &ui.Box{Title: ui.Text("攻击")}
			md.status.desc.SetText("你的攻击属性。")
			md.status.focus = true
		case i == statusDefense:
			md.status.desc.Box = &ui.Box{Title: ui.Text("防御")}
			md.status.desc.SetText("你的防御属性。")
			md.status.focus = true
		case i >= statusIndex:
			i := md.status.menu.Active() - int(statusIndex)
			var st Status = -1
			var n int
			pa := md.g.PlayerActor()
			for j, turns := range pa.Statuses {
				if turns <= 0 {
					continue
				}
				if n == i {
					st = Status(j)
					break
				}
				n++
			}
			if st < 0 {
				if !pa.IsDead() && !md.g.win {
					md.status.desc.Box = &ui.Box{Title: ui.Text("所在格")}
					md.status.desc.SetText("玩家所在的位置。")
					md.status.focus = true
				}
				break
			}
			turns := pa.Statuses[st]
			title := ui.StyledText{}.WithMarkups(Markups).
				WithTextf("@%c%s@N（剩余 %s）",
				statusColor(st, turns), st.Name(), CountNoun("回合", turns))
			md.status.desc.Box = &ui.Box{Title: title}
			desc := st.Desc()
			if details := st.Details(); details != "" {
				desc += " " + details
			}
			md.status.desc.SetText(desc)
			md.status.desc.Content = md.status.desc.Content.Format(UIWidth - 2)
			md.status.focus = true
		}
	}
}

// endTurn finalizes player's turn and runs other events until next player
// turn. The given argument contains the expected HP at the end of the turn.
func (md *model) endTurn(xhp int) {
	if md.mode != modeCritical {
		md.mode = modeNormal
	}
	g := md.g
	g.EndTurn()
	pa := g.PlayerActor()
	if pa.DoesAny(GoodSmell) {
		g.SmellFood()
	}
	if pa.DoesAny(RunicChicken) {
		g.SenseItems(itemRune)
	}
	if pa.DoesAny(Gawalt) {
		g.SenseItems(itemMenhir)
	}
	hp := pa.HP
	if hp <= 0 {
		md.death()
		return
	}
	if hp <= HPCritical && xhp > HPCritical {
		// Warn only if HP went below critical point when it was
		// expected to remain higher.
		g.WarningPrompt()
			g.LogStyled("*** 生命值危急 ***", logConfirm)
	}
	if md.mode == modeCritical {
		md.logConfirmContinue()
	}
	g.Logs.NextTick = g.Logs.Index
}

func (md *model) death() {
	g := md.g
	g.StoryLog("死亡")
	g.LogStyled("你死了……", logSpecial)
	g.LevelStats()
	md.logConfirmContinue()
	md.mode = modeEnd
}

func (md *model) logConfirmContinue() {
	md.g.LogStyled("[空格键/ESC 继续]", logConfirm)
	md.targ.CancelExamine()
}

func (md *model) dump(msg string, err error) {
	s := md.g.DumpEndSummary(false)
	lines := strings.Split(s, "\n")
	stts := []ui.StyledText{}
	for _, l := range lines {
		stts = append(stts, ui.Text(l))
	}
	var details string
	if err != nil {
		log.Printf("写入统计记录时出错：%v", err)
		details = "统计记录写入失败。"
	} else {
		details = msg
	}
	stts = append(stts, ui.Text(details))
	md.pager.pg.SetBox(&ui.Box{Title: ui.Text("游戏总结").WithStyle(gruid.Style{Fg: ColorYellow})})
	md.pager.pg.SetLines(stts)
	md.pager.pg.SetCursor(gruid.Point{0, 0})
}
