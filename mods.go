package main

// Mod represents the various kinds of advanced mods for the game.
type Mod int

// Those constants describe the available mods.
const (
	ModNoisy                  Mod = iota // noisy movement and uncontrollable screams
	ModSpiderPhobia                      // fear of spiders (unless berserk); more spiders
	ModExtraMonsters                     // mod with extra monsters
	ModExtraGuardians                    // mod with extra regular guardians
	ModCorruptedDungeon                  // mod with lots of unusual and random stuff going on
	ModNoRecharges                       // mod without spirit recharges
	ModGluttony                          // uncontrollable extra eating
	ModGawalt                            // gawalt-like wall jumping
	ModHeavyTelepathy                    // two-way Clarity-like telepathy with heavy-step monsters
	ModEnableChallengeSpirits            // dedicated slot for challenge secondary spirits
)

const NMods = 1 + int(ModEnableChallengeSpirits)

var gameMods = []Mod{ModNoisy, ModSpiderPhobia, ModExtraMonsters, ModExtraGuardians, ModCorruptedDungeon, ModNoRecharges,
	ModGluttony, ModGawalt, ModHeavyTelepathy, ModEnableChallengeSpirits}

func (m Mod) String() string {
	switch m {
	case ModNoisy:
		return "Noisy"
	case ModSpiderPhobia:
		return "Spider Phobia"
	case ModExtraMonsters:
		return "Extra Monsters"
	case ModExtraGuardians:
		return "Extra Guardians"
	case ModCorruptedDungeon:
		return "Corrupted Dungeon"
	case ModNoRecharges:
		return "No Recharges"
	case ModGluttony:
		return "Gluttony"
	case ModGawalt:
		return "Gawalt Monkey"
	case ModHeavyTelepathy:
		return "Heavy Telepathy (NYI)"
	case ModEnableChallengeSpirits:
		return "Challenge Spirits (NYI)"
	default:
		return "(unknown mod)"
	}
}

func (m Mod) Desc() string {
	switch m {
	case ModNoisy:
		return "You make noise when you move (about half view range). You sometimes uncontrollably scream at your foes frighteningly."
	case ModSpiderPhobia:
		return "You fear spiders. Spiders are more common."
	case ModExtraMonsters:
		return "More wandering monsters than usual."
	case ModExtraGuardians:
		return "Some extra regular monsters as guardians."
	case ModCorruptedDungeon:
		return "The orb's influence has scattered twisted surprises through the land."
	case ModNoRecharges:
		return "Spirit ability charges are doubled but don't recharge when going to next floor."
	case ModHeavyTelepathy:
		return "You get Clarity-like telepathy with heavy-footstep monsters."
	case ModGluttony:
		return "When eating, you sometimes eat an extra random comestible on the same turn. You have +2 HP."
	case ModGawalt:
		return "You can perform 2-tile jumps by propulsing against walls! You attack any monster on the jump path. When attacking, damage is reduced by one if greater than one."
	case ModEnableChallengeSpirits:
		return "Replace one secondary spirit slot with a special challenge one."
	default:
		return "(unknown mod)"
	}
}
