package main

// Mod represents the various kinds of advanced mods for the game.
type Mod int

// Those constants describe the available mods.
const (
	ModHeavySteps             Mod = iota // noisy movement
	ModSpiderPhobia                      // fear of spiders (unless berserk); more spiders
	ModMonsterSwarm                      // mod with extra monsters
	ModTerrainFlip                       // mod with flipped foliage/floor cells
	ModNoRecharges                       // mod without spirit recharges
	ModHeavyTelepathy                    // two-way Clarity-like telepathy with heavy-step monsters
	ModSmallInventory                    // inventory with 3 comestibles slots
	ModEnableChallengeSpirits            // dedicated slot for challenge secondary spirits
)

const NMods = 1 + int(ModEnableChallengeSpirits)

var gameMods = []Mod{ModHeavySteps, ModSpiderPhobia, ModMonsterSwarm, ModTerrainFlip, ModNoRecharges,
	ModHeavyTelepathy, ModSmallInventory, ModEnableChallengeSpirits}

func (m Mod) String() string {
	switch m {
	case ModHeavySteps:
		return "Heavy Steps"
	case ModSpiderPhobia:
		return "Spider Phobia"
	case ModMonsterSwarm:
		return "Monster Swarm"
	case ModTerrainFlip:
		return "Terrain Flip"
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
	case ModHeavySteps:
		return "You make noise when you move (about half view range)."
	case ModSpiderPhobia:
		return "You fear spiders. More spiders."
	case ModMonsterSwarm:
		return "Extra monsters."
	case ModTerrainFlip:
		return "Flipped foliage and floor terrains outside of vaults (usually means lots of foliage)."
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
