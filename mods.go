package main

import "fmt"

// Mod represents the various kinds of mods for the game.
type Mod int

// Those constants describe the available mods.
const (
	ModCorruptedDungeon Mod = iota // mod with lots of unusual and random stuff going on
	ModAdvancedSpirits             // dedicated slot for challenge secondary spirits
	ModTotemConditions             // totems come with Curse/Until conditions

	ModGluttonyRework // mod with choice-of-two gluttony rework

	ModSmallInventory // mod limiting inventory to 3 comestible slots
	ModNoRecharges    // mod without spirit recharges
	ModHealingCombat  // combat-based healing system instead of comestible-based

	modCount // dummy mod (for NMods)
)

// Number of mods.
const NMods = int(modCount)

// Position of first mod in each mod section.
const (
	FirstExpansion = ModCorruptedDungeon
	FirstSmallMod  = ModGluttonyRework
	FirstChallenge = ModSmallInventory
)

// GameMods represents the available game mods.
var GameMods = []Mod{
	// Expansions
	ModCorruptedDungeon, ModAdvancedSpirits, ModTotemConditions,
	// Small Mods
	ModGluttonyRework,
	// Challenges
	ModSmallInventory,
	ModNoRecharges, ModHealingCombat,
}

func (m Mod) String() string {
	switch m {
	case ModCorruptedDungeon:
		return "Corrupted Dungeon"
	case ModAdvancedSpirits:
		return "Advanced Spirits"
	case ModTotemConditions:
		return "Totem Conditions"
	case ModGluttonyRework:
		return "Gluttony Rework"
	case ModSmallInventory:
		return "Small Inventory"
	case ModNoRecharges:
		return "No Recharges"
	case ModHealingCombat:
		return "Healing Combat"
	default:
		return "(unknown mod)"
	}
}

func (m Mod) Desc() string {
	switch m {
	case ModCorruptedDungeon:
		return "The orb’s influence has scattered twisted surprises through the dungeon. Some might even be good!\n\n@SRecommended for players that want more unpredictability and variability.@N"
	case ModAdvancedSpirits:
		return "Enable extra advanced secondary spirits: Dazzling Zebra, Gardening Lion, Gawalt Monkey, Gluttonous Bear, Runic Chicken, Staring Owl, Stomping Elephant.\n\nAdvanced secondaries always have strong points but also serious and quirky drawbacks.\n\n@SRecommended for players that want more build variety.@N"
	case ModTotemConditions:
		return "Read the fine print: each spirit will curse you to play by its rules or lose its benefits. You may also find comestibles that aren’t so easy to eat!\n\nThink ahead when choosing spirits; the run will have one fewer empty totem to expand your options.\n\n@SRecommended for players that want a fun challenge and tricky totem choices.@N"
	case ModGluttonyRework:
		return "Makes the Gluttonous Bear eat in pairs through a choice-of-two menu, without relying on a Gluttony status.\n\nThe Advanced Spirits mod should be enabled too for this mod to take effect.\n\n@SRecommended for players that don’t like the default Bear with a Gluttony status.@N"
	case ModSmallInventory:
		return "Inventory can only hold 3 comestibles. Choose them with care!"
	case ModNoRecharges:
		return "Spirit ability charges are doubled but don’t recharge when going to the next map level. Use them wisely!"
	case ModHealingCombat:
		return fmt.Sprintf("Healing happens through combat. When a monster dies, you may heal for 1 HP, with higher chance at low HP. However, comestibles don’t provide healing anymore, except for ambrosia berries still healing %d HP.\n\nVampiric Bat players start with an extra “vampirism” charge and get another one at map level 5, instead of healing on monster death. Less frequent healing but more active control over it!\n\n@SRecommended for players that want to try an alternate experience with a single source of healing.@N", HealAmbrosiaHC)
	default:
		return "(unknown mod)"
	}
}

// HasVampirism reports whether the player has a vampirism ability.
func (g *Game) HasVampirism() bool {
	sp := g.Entity(0).Role.(*Spirit)
	_, ok := sp.GetAbility().(EffectVampirism)
	return ok
}
