package main

import (
	"codeberg.org/anaseto/gruid"
)

// RunicTrap represents a triggerable runic trap.
type RunicTrap struct {
	Rune MagicRune // magic rune
}

// MagicRune represents the various kinds of runes used in runic traps.
type MagicRune int

const (
	RuneInert MagicRune = iota
	RuneBerserk
	RuneFire
	RuneLignification
	RunePoison
	RuneWarp
)

// NRunes is the number of non-inert distinct magical runes.
const NRunes = int(RuneWarp)

// String returns the name of the runic trap.
func (mr MagicRune) String() string {
	switch mr {
	case RuneBerserk:
		return "berserk rune"
	case RuneFire:
		return "fire rune"
	case RuneLignification:
		return "lignification rune"
	case RunePoison:
		return "poison rune"
	case RuneWarp:
		return "warp rune"
	default:
		return "inert rune"
	}
}

// Color returns the color of the runic trap.
func (rt RunicTrap) Color() gruid.Color {
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
		return ColorForeground
	}
}

// Desc returns the description of the runic trap.
func (rt RunicTrap) Desc() string {
	switch rt.Rune {
	case RuneBerserk:
		return "Applies Berserk status when stepped on."
	case RuneFire:
		return "Spawns a fire cloud when stepped on."
	case RuneLignification:
		return "Applies Lignification status when stepped on."
	case RunePoison:
		return "Spawns a poison cloud when stepped on."
	case RuneWarp:
		return "Teleports you away when stepped on."
	default:
		return "The already triggered rune became inert."
	}
}

// RunicTraps are items but they cannot be used like other items.
func (rt RunicTrap) Use(g *Game, id ID) bool {
	g.Log("You cannot interact with a trap.")
	return false
}

// NoTrapAt reports whether there's no trap at the given position.
func (g *Game) NoTrapAt(p gruid.Point) bool {
	return g.Map.RuneCache.At(p) <= 0
}

// TriggerTrapAt triggers any runic traps at the given point.
func (g *Game) TriggerTrapAt(p gruid.Point) {
	i := g.Map.RuneCache.At(p)
	if i <= 0 {
		return
	}
	ei := g.Entity(i)
	rt, ok := ei.Role.(*RunicTrap)
	if !ok {
		return
	}
	if j, _ := g.ActorAt(p); j >= 0 {
		if j == PlayerID {
			g.LogfStyled("You trigger the %s.", logSpecial, rt.Rune.String())
			g.StoryLogf("Triggered %s", One(rt.Rune.String()))
		} else if g.InFOV(p) {
			g.LogfStyled("The %s triggers the %s.", logSpecial, g.Entity(j).Name, rt.Rune.String())
		}
	}
	switch rt.Rune {
	case RuneBerserk:
		if i, ai := g.ActorAt(p); i >= 0 {
			g.PutStatus(i, ai, StatusBerserk, DurationBerserkTotem)
		}
	case RuneFire:
		g.FireCloudAt(p)
	case RuneLignification:
		if i, ai := g.ActorAt(p); i >= 0 {
			g.PutStatus(i, ai, StatusLignification, DurationLignification/2)
		}
	case RunePoison:
		g.PoisonCloudAt(p)
	case RuneWarp:
		if i, ai := g.ActorAt(p); i >= 0 {
			g.TeleportActor(i, ai)
		}
	}
	rt.Rune = RuneInert
	ei.Name = RuneInert.String()
	g.Map.RuneCache.SetU(p, 0)
}
