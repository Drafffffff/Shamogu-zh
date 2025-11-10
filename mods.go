package main

// Mod represents the various kinds of advanced mods for the game.
type Mod int

// Those constants describe the available mods.
const (
	ModCorruptedDungeon Mod = iota // mod with lots of unusual and random stuff going on
	ModChallengeSpirits            // dedicated slot for challenge secondary spirits
	ModNoRecharges                 // mod without spirit recharges

	ModCombatGrowl // uncontrollable growling during combat
	ModHeavySteps  // noisy movement

)

const NMods = 1 + int(ModHeavySteps)

var gameMods = []Mod{
	ModCorruptedDungeon, ModChallengeSpirits, ModNoRecharges,
	ModCombatGrowl, ModHeavySteps}

func (m Mod) String() string {
	switch m {
	case ModCorruptedDungeon:
		return "Corrupted Dungeon"
	case ModChallengeSpirits:
		return "Challenge Spirits"
	case ModNoRecharges:
		return "No Recharges"
	case ModCombatGrowl:
		return "Combat Growl"
	case ModHeavySteps:
		return "Heavy Steps"
	default:
		return "(unknown mod)"
	}
}

func (m Mod) Desc() string {
	switch m {
	case ModCorruptedDungeon:
		return "The orb’s influence has scattered twisted surprises through the dungeon. Some might even be good!\n\nRecommended for experienced players that want more unpredictability and variability."
	case ModChallengeSpirits:
		return "Enable extra challenge secondary spirits: Gawalt Monkey, Gardening Lion, Gluttonous Bear.\n\nYou cannot chose more than one regular secondary spirit, and the first totem is always a challenge one!"
	case ModNoRecharges:
		return "Spirit ability charges are doubled but don’t recharge when going to the next map level. Use them wisely!"
	case ModCombatGrowl:
		return "You sometimes growl during combat, briefly scaring foes (experimental)."
	case ModHeavySteps:
		return "You frequently make noise when you move (experimental)."
	default:
		return "(unknown mod)"
	}
}
