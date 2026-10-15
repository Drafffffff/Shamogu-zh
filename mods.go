package main

import "fmt"

// Mod represents the various kinds of mods for the game.
type Mod int

// Those constants describe the available mods.
const (
	ModAdvancedSpirits  Mod = iota // dedicated slot for challenge secondary spirits
	ModCorruptedDungeon            // mod with lots of unusual and random stuff going on
	ModTotemConditions             // totems come with Curse/Until conditions

	ModHealingCombat  // combat-based healing system instead of comestible-based
	ModNoRecharges    // mod without spirit recharges
	ModSmallInventory // mod limiting inventory to 3 comestible slots

	modCount // dummy mod (for NMods)
)

// Number of mods.
const NMods = int(modCount)

// Position of first mod in each mod section.
const (
	FirstExpansion = ModAdvancedSpirits
	FirstChallenge = ModHealingCombat
)

// GameMods represents the available game mods.
var GameMods = []Mod{
	// Expansions
	ModAdvancedSpirits, ModCorruptedDungeon, ModTotemConditions,
	// Challenges
	ModHealingCombat,
	ModNoRecharges,
	ModSmallInventory,
}

func (m Mod) String() string {
	switch m {
	case ModAdvancedSpirits:
		return "进阶灵体"
	case ModCorruptedDungeon:
		return "腐化地城"
	case ModTotemConditions:
		return "图腾诅咒"
	case ModHealingCombat:
		return "战斗疗愈"
	case ModNoRecharges:
		return "不再充能"
	case ModSmallInventory:
		return "小型物品栏"
	default:
		return "（未知模组）"
	}
}

func (m Mod) Desc() string {
	switch m {
	case ModAdvancedSpirits:
		return "新增进阶副灵体：炫目斑马、园艺狮、加瓦尔特猴、贪食熊、符文鸡、凝视猫头鹰、重踏象。\n\n进阶副灵体各有强项，也有显著且古怪的缺点。\n\n@S推荐给希望尝试更多搭配的玩家。@N"
	case ModCorruptedDungeon:
		return "魔法球的影响在地城中散播了扭曲的惊喜，其中有些甚至可能是好事！\n\n@S推荐给熟悉基础游戏、希望增加变化与不可预测性的玩家。@N"
	case ModTotemConditions:
		return "仔细阅读说明：每个灵体都会要求你遵守它的规则，否则便会失去其能力。你还可能发现一些不容易吃下的食物！\n\n选择灵体时要深思熟虑；每局可用于扩展搭配的空图腾会少一个。\n\n@S推荐给希望体验条件严苛、图腾选择棘手的玩家。@N"
	case ModHealingCombat:
		return fmt.Sprintf("通过战斗恢复生命。怪物死亡时，你有机会恢复 1 点生命值；生命值越低，触发概率越高。但食物不再恢复生命，仙馔莓除外，它仍可恢复 %d 点生命值。\n\n吸血蝙蝠玩家开局会额外获得一次“吸血”充能，并在第 5 层再获得一次，不再通过击杀怪物回血。回血机会更少，但更容易主动掌控！\n\n@S推荐给想体验单一回血来源的玩家。@N", HealAmbrosiaHC)
	case ModNoRecharges:
		return "灵体能力的充能数量翻倍，但前往下一层时不会恢复。请谨慎使用！"
	case ModSmallInventory:
		return "物品栏最多只能存放 3 份食物，请谨慎选择！"
	default:
		return "（未知模组）"
	}
}

// HasVampirism reports whether the given entity has a vampirism ability.
func (e *Entity) HasVampirism() bool {
	if sp, ok := e.Role.(*Spirit); ok {
		_, ok := sp.GetAbility().(EffectVampirism)
		return ok
	}
	return false
}
