package main

// Mod represents the various kinds of advanced mods for the game.
type Mod int

// Those constants describe the available mods.
const (
	ModNoisy                  Mod = iota // noisy movement and uncontrollable screams
	ModSpiderPhobia                      // fear of spiders (unless berserk); more spiders
	ModMonsterSwarm                      // mod with extra monsters
	ModExtraGuardians                    // mod with extra regular guardians
	ModExtraFoliage                      // mod with flipped foliage/floor cells
	ModNoRecharges                       // mod without spirit recharges
	ModHeavyTelepathy                    // two-way Clarity-like telepathy with heavy-step monsters
	ModSmallInventory                    // inventory with 3 comestibles slots
	ModEnableChallengeSpirits            // dedicated slot for challenge secondary spirits
)

const NMods = 1 + int(ModEnableChallengeSpirits)

var gameMods = []Mod{ModNoisy, ModSpiderPhobia, ModMonsterSwarm, ModExtraGuardians, ModExtraFoliage, ModNoRecharges,
	ModHeavyTelepathy, ModSmallInventory, ModEnableChallengeSpirits}

func (m Mod) String() string {
	switch m {
	case ModNoisy:
		return "Noisy"
	case ModSpiderPhobia:
		return "Spider Phobia"
	case ModMonsterSwarm:
		return "Monster Swarm"
	case ModExtraGuardians:
		return "Extra Guardians"
	case ModExtraFoliage:
		return "Extra foliage"
	case ModNoRecharges:
		return "No Recharges"
	case ModHeavyTelepathy:
		return "Heavy Telepathy (NYI)"
	case ModSmallInventory:
		return "Small Inventory (NYI)"
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
		return "You fear spiders. More spiders."
	case ModMonsterSwarm:
		return "Extra monsters."
	case ModExtraGuardians:
		return "Some portals now have extra regular monsters as guardians."
	case ModExtraFoliage:
		return "Extra foliage."
	case ModNoRecharges:
		return "Spirit ability charges are doubled but don't recharge when going to next floor."
	case ModHeavyTelepathy:
		return "You get Clarity-like telepathy with heavy-footstep monsters."
	case ModSmallInventory:
		return "Only 3 comestible inventory slots."
	case ModEnableChallengeSpirits:
		return "Replace one secondary spirit slot with a special challenge one."
	default:
		return "(unknown mod)"
	}
}
