package main

// Mod represents the various kinds of advanced mods for the game.
type Mod int

// Those constants describe the available mods.
const (
	ModCorruptedDungeon Mod = iota // mod with lots of unusual and random stuff going on
	ModChallengeSpirits            // dedicated slot for challenge secondary spirits
	ModGluttonyRework              // mod with choice-of-two gluttony rework
	ModNoRecharges                 // mod without spirit recharges

	ModSmallInventory // mod limiting inventory to 3 comestible slots
	ModOrbTracking    // mod making the orb send monsters to you more actively
	ModMonsterRevival // mod making monster occasionally revive

	modCount // dummy mod (for NMods)
)

const NMods = int(modCount)

var gameMods = []Mod{
	// extension mods
	ModCorruptedDungeon, ModChallengeSpirits, ModGluttonyRework,
	// difficulty mods
	ModNoRecharges, ModSmallInventory,
	ModOrbTracking, ModMonsterRevival,
}

func (m Mod) String() string {
	switch m {
	case ModCorruptedDungeon:
		return "Corrupted Dungeon"
	case ModChallengeSpirits:
		return "Challenge Spirits"
	case ModNoRecharges:
		return "No Recharges"
	case ModGluttonyRework:
		return "Gluttony Rework"
	case ModSmallInventory:
		return "Small Inventory"
	case ModOrbTracking:
		return "Orb Tracking"
	case ModMonsterRevival:
		return "Monster Revival"
	default:
		return "(unknown mod)"
	}
}

func (m Mod) Desc() string {
	switch m {
	case ModCorruptedDungeon:
		return "The orb’s influence has scattered twisted surprises through the dungeon. Some might even be good!\n\nRecommended for experienced players that want more unpredictability and variability."
	case ModChallengeSpirits:
		return "Enable extra challenge secondary spirits: Gawalt Monkey, Gardening Lion, Gluttonous Bear, Stomping Elephant, Staring Owl, Runic Chicken.\n\nYou cannot chose more than one regular secondary spirit!"
	case ModNoRecharges:
		return "Spirit ability charges are doubled but don’t recharge when going to the next map level. Use them wisely!"
	case ModGluttonyRework:
		return "Makes the Gluttonous Bear eat in pairs through a choice-of-two menu, without relying on a Gluttony status.\n\nThe Challenge Spirits mod should be enabled too for this mod to take effect."
	case ModSmallInventory:
		return "Inventory can only hold 3 comestibles."
	case ModOrbTracking:
		return "The orb occasionally sends your current position to wandering monsters. You need to keep moving!"
	case ModMonsterRevival:
		return "The orb revives monsters after a random delay. You cannot kill them all!"
	default:
		return "(unknown mod)"
	}
}
