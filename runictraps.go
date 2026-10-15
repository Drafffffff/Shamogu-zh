package main

import (
	"fmt"

	"codeberg.org/anaseto/gruid"
)

// RunicTrap represents a triggerable runic trap.
type RunicTrap struct {
	Used      bool      // whether the rune has already been triggered
	KnownUsed bool      // whether the rune is known to have already been triggered
	Rune      MagicRune // magic rune
}

// MagicRune represents the various kinds of runes used in runic traps.
type MagicRune int

const (
	RuneBerserk MagicRune = iota
	RuneFire
	RuneLignification
	RunePoison
	RuneWarp
)

// NRunes is the number of non-inert distinct magical runes.
const NRunes = 1 + int(RuneWarp)

// String returns the name of the runic trap.
func (mr MagicRune) String() string {
	switch mr {
	case RuneBerserk:
		return "狂暴符文"
	case RuneFire:
		return "火焰符文"
	case RuneLignification:
		return "木化符文"
	case RunePoison:
		return "毒素符文"
	case RuneWarp:
		return "传送符文"
	default:
		// Should not happen.
		return "未知符文"
	}
}

// Color returns the color of the runic trap.
func (rt *RunicTrap) Color() gruid.Color {
	if rt.KnownUsed {
		return ColorForeground
	}
	switch rt.Rune {
	case RuneBerserk:
		return ColorMagenta
	case RuneFire:
		return ColorRed
	case RuneLignification:
		return ColorOrange
	case RunePoison:
		return ColorGreen
	case RuneWarp:
		return ColorViolet
	default:
		// Should not happen.
		return ColorForeground
	}
}

// Desc returns the description of the runic trap.
func (rt *RunicTrap) Desc(_ *Game) string {
	if rt.KnownUsed {
		return "符文已经触发过了。"
	}
	switch rt.Rune {
	case RuneBerserk:
		return fmt.Sprintf("踩上时获得 @B狂暴@N 状态，持续 %d 回合；若已处于狂暴状态则不会触发。", DurationBerserkTrap)
	case RuneFire:
		return fmt.Sprintf("踩上时陷入 @O迷向@N，持续 %d 回合，并生成持续 6 至 12 回合的火焰云雾。",
			DurationConfusionFireTrap)
	case RuneLignification:
		return fmt.Sprintf("踩上时获得 @B木化@N 状态，持续 %d 回合；若已处于木化状态则不会触发。", DurationLignificationTrap)
	case RunePoison:
		return "踩上时生成持续 7 至 11 回合的毒雾。"
	case RuneWarp:
		return "将踩上去的生物传送走。"
	default:
		// Should not happen.
		return "未知效果。"
	}
}

// RunicTraps are items but they cannot be used like other items.
func (rt *RunicTrap) Use(g *Game, id ID) bool {
	g.Log("你无法与陷阱互动。")
	return false
}

// NoTrapAt reports whether there's no trap at the given position.
func (g *Game) NoTrapAt(p gruid.Point) bool {
	return g.Map.RuneCache.At(p) <= 0
}

// TriggerTrap makes the given actor trigger a runic trap at its current
// position, if any.
func (g *Game) TriggerTrap(i ID, ai *Actor) {
	if ai.DoesAny(RunicChicken) {
		return
	}
	ei := g.Entity(i)
	pi := ei.P
	rt, ok := g.RunicTrapAt(pi)
	if !ok || rt.Used {
		return
	}
	switch {
	case rt.Rune == RuneLignification && (ai.Has(StatusLignification) || ai.Is(WalkingTree)):
		// Do not trigger lignification traps while lignified.
		return
	case rt.Rune == RuneBerserk && ai.Has(StatusBerserk):
		// Do not trigger berserk traps while berserk.
		return
	case rt.Rune == RuneWarp && i == PlayerID:
		// Update FOV before teleporting.
		g.UpdateFOV()
		g.UpdateKnowledge()
	}
	switch {
	case i == PlayerID:
		g.LogfStyled("你触发了%s。", logSpecial, rt.Rune.String())
		g.StoryLogf("触发了%s", One(rt.Rune.String()))
		g.Stats.PlayerTriggers++
		rt.KnownUsed = true
	case g.InFOV(pi):
		g.LogfStyled("%s触发了%s。", logSpecial, ei.Name, rt.Rune.String())
		rt.KnownUsed = true
	case SensingRange(pi, g.PP()) && g.PlayerActor().DoesAny(RunicChicken):
		g.LogfStyled("%s触发了。", logSpecial, rt.Rune.String())
		rt.KnownUsed = true
	}
	if i != PlayerID {
		g.Stats.MonsterTriggers++
	}
	g.Stats.MapTriggeredTraps[g.Map.Level-1]++
	rt.Used = true
	g.Map.RuneCache.SetU(pi, 0)
	// Apply runic effect.
	switch rt.Rune {
	case RuneBerserk:
		g.PutStatus1(i, ai, StatusBerserk, DurationBerserkTrap)
	case RuneFire:
		g.PutStatus1(i, ai, StatusConfusion, DurationConfusionFireTrap)
		g.FireCloudAt(pi)
	case RuneLignification:
		g.PutStatus1(i, ai, StatusLignification, DurationLignificationTrap)
	case RunePoison:
		g.PoisonCloudAt(pi)
	case RuneWarp:
		g.TeleportActor(i, ai, 1)
	}
}

// RunicTrapAt returns any runic trap at the given position.
func (g *Game) RunicTrapAt(at gruid.Point) (*RunicTrap, bool) {
	i := g.Map.RuneCache.At(at)
	if i <= 0 {
		return nil, false
	}
	ei := g.Entity(i)
	rt, ok := ei.Role.(*RunicTrap)
	return rt, ok
}
