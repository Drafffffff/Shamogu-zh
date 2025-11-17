package main

// Mod represents the various kinds of advanced mods for the game.
type Mod int

// Those constants describe the available mods.
const (
	ModCorruptedDungeon Mod = iota // mod with lots of unusual and random stuff going on
	ModChallengeSpirits            // dedicated slot for challenge secondary spirits
	ModNoRecharges                 // mod without spirit recharges
	ModGluttonyRework              // mod with choice-of-two gluttony rework

	modCount // dummy mod (for NMods)
)

const NMods = int(modCount)

var gameMods = []Mod{
	ModCorruptedDungeon, ModChallengeSpirits, ModNoRecharges,
	ModGluttonyRework}

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
	default:
		return "(unknown mod)"
	}
}
