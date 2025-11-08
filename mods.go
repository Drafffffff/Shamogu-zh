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

	ModExtraMonsters  // mod with extra monsters (XXX: maybe remove)
	ModExtraGuardians // mod with extra regular guardians (XXX: maybe remove)

)

const NMods = 1 + int(ModExtraGuardians)

var gameMods = []Mod{
	ModCorruptedDungeon, ModChallengeSpirits, ModNoRecharges,
	ModCombatGrowl, ModHeavySteps,
	ModExtraMonsters, ModExtraGuardians}

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
	case ModExtraMonsters:
		return "Extra Monsters"
	case ModExtraGuardians:
		return "Extra Guardians"
	default:
		return "(unknown mod)"
	}
}

func (m Mod) Desc() string {
	switch m {
	case ModCorruptedDungeon:
		return "The orb's influence has scattered twisted surprises through the dungeon. Some might even be good! Recommended for players that want more unpredictability and variability between playthroughs."
	case ModChallengeSpirits:
		return "Enable extra challenge secondary spirits: Gawalt Monkey, Gardening Lion, Gluttonous Bear. You cannot chose more than one regular secondary spirit, and the first totem is always a challenge one!"
	case ModNoRecharges:
		return "Spirit ability charges are doubled but don't recharge when going to next floor. Use them wisely!"
	// case ModScaryRoar:
	// 	return "You roar at monsters on first sight to scare them off."
	case ModCombatGrowl:
		return "You sometimes growl during combat, briefly scaring foes."
	case ModHeavySteps:
		return "You frequently make noise when you move."
	// case ModGluttony:
	// 	return "When eating, you sometimes eat an extra random comestible on the same turn. You have +2 HP."
	// case ModGawalt:
	// 	return "You can perform 2-tile jumps by propulsing against walls! You attack any monster on the jump path. When attacking, damage is reduced by one if greater than one."
	case ModExtraMonsters:
		return "More wandering monsters than usual."
	case ModExtraGuardians:
		return "Some extra regular monsters as guardians."
	default:
		return "(unknown mod)"
	}
}
