package main

import (
	"fmt"
	"strings"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
)

// Actor holds data relevant to any kind of actor (player or monster).
type Actor struct {
	HP        int         // health points
	MaxHP     int         // maximum health points
	Defense   int         // defense (defensive potential)
	Attack    int         // attack (offensive potential)
	FirePos   gruid.Point // position at start of turn while on Fire
	KnownDead bool        // whether it's a monster known to be dead
	Statuses  Statuses    // statuses (confused, etc.)
	Kind      ActorKind   // actor kind (player or some kind of monster)
	Traits    Traits      // passives traits
	Behavior  *Behavior   // behavior component (monsters only)
}

// NewActor returns a new actor with the given combat stats and traits.
func NewActor(attack, defense, hp int, ak ActorKind, t Traits) *Actor {
	a := &Actor{
		Attack:   attack,
		Defense:  defense,
		HP:       hp,
		MaxHP:    hp,
		Kind:     ak,
		Statuses: make(Statuses, NStatuses),
		Traits:   t,
	}
	return a
}

// DoesAny reports whether the player has any trait in t.
func (a *Actor) DoesAny(t Traits) bool {
	return a.Traits.Any(t)
}

// DoesAll reports whether the player has all traits in t.
func (a *Actor) DoesAll(t Traits) bool {
	return a.Traits.All(t)
}

// Has reports whether the player has status st.
func (a *Actor) Has(st Status) bool {
	return a.Statuses.Has(st)
}

// Is reports whether the actor is of the given kind.
func (a *Actor) Is(ak ActorKind) bool {
	return a.Kind == ak
}

// CanMove reports whether the actor isn't under any status that prevents
// movement.
func (a *Actor) CanMove() bool {
	return !a.Has(StatusLignification) || a.DoesAny(WoodyLegs)
}

// ResistsMove reports whether the actor resists against forced movement. This
// includes lignified actors that can still move due to woodylegs as well as
// walking trees.
func (a *Actor) ResistsMove() bool {
	return a.Has(StatusLignification) || a.Is(WalkingTree)
}

// DoesChaosAttack reports whether the actor has a multi-effect chaotic attack.
func (a *Actor) DoesChaosAttack() bool {
	return !NoChaos && a.Is(ChaosMegabat)
}

// IsAlive reports whether the actor is still alive.
func (a *Actor) IsAlive() bool {
	return a.HP > 0
}

// IsDead reports whether the actor is dead.
func (a *Actor) IsDead() bool {
	return a.HP <= 0
}

// AdjustHP heals an actor by a given amount (or hurts if negative). The result
// is guaranteed to be in [1,a.MaxHP].
func (g *Game) AdjustHP(i ID, ai *Actor, amount int) {
	g.AdjustHPWithMin(i, ai, amount, 1)
}

// AdjustHPWithMin heals an actor by a given amount (or hurts if negative). The
// result is guaranteed to be in [min(m,ai.HP,maxHP), maxHP] where maxHP is
// ai.GetMaxHP().
func (g *Game) AdjustHPWithMin(i ID, ai *Actor, amount int, m int) {
	if ai.IsDead() {
		// Should not happen: dead monsters remain dead.
		return
	}
	hp, maxHP := ai.HP, ai.GetMaxHP()
	ai.HP = min(maxHP, max(min(m, hp), hp+amount))
	if i == PlayerID {
		if ai.HP > hp {
			g.Logf("你感觉好多了（+%d 点生命值）。", ai.HP-hp)
		} else if ai.HP < hp {
			g.Logf("你感到虚弱（-%d 点生命值）。", hp-ai.HP)
		} else if amount >= 0 {
			g.Log("你感觉没有变化（+0 点生命值）。")
		}
	}
}

// UpdateFirePos updates the FirePos field of the given actor.
func (g *Game) UpdateFirePos(i ID, ai *Actor) {
	if ai.Has(StatusFire) {
		ai.FirePos = g.Entity(i).P
	} else {
		ai.FirePos = InvalidPos
	}
}

// Traits represents special caracteristics of an actor as a bitset.
type Traits uint64

// Any reports whether any of the traits is in the set.
func (t Traits) Any(of Traits) bool {
	return t&of != 0
}

// All reports whether all of the traits are in the set.
func (t Traits) All(of Traits) bool {
	return t&of == of
}

// Extra Traits. NOTE: there shouldn't be more than 64 (remaining free: 12).
//
// Mostly to be used for player traits and traits that may be shared by more
// than a single actor kind (but currently various traits are specific to
// single monsters).
const (
	PatternCatch = 1 << iota // unbalancing and catching attack (frog, octopode)

	PatternDragging     // dragging attack pattern (crocodile players, alligator)
	PatternFourDirs     // four-directional attack in 4 dirs (hydra, undead knight)
	PatternRampage      // rampaging (boar, phoenix)
	PatternRanged       // ranged attack (eye, llama, lich)
	PatternRangedRecoil // ranged attack with recoil (wind fox)
	PatternSneaky       // sneaky attack pattern (vampiric bat, chaos megabat)
	PatternSwap         // swapping-attack (crazy druid)
	PatternSwapDaze     // dazing swapping-attack (temporal cat)

	BadHearing      // worse hearing of noise (footsteps)
	BurningHits     // hits may burn
	DazingSpines    // body covered in dazing spines (porcupine)
	Dazzling        // zebra's camouflage
	Elephanty       // rat phobia, difficult rotation when facing a wall, immune to berserk after-effects
	Gawalt          // wall-jumping, weakened hits, menhir sensing and hiding
	Gluttony        // gluttony status after eating
	GoodHearing     // better hearing of noise (footsteps)
	GoodSmell       // smell food from a distance
	NocturnalFlying // flying a bit above the ground like an owl
	PushingCharge   // unbalancing pushing when charging (boar)
	RunicChicken    // doesn't trigger traps, cackles loudly on sighting a portal or totem
	ScaryRoar       // intimidating roar on first sight
	TrailingCloud   // leave dust behind after moving (on floor tiles)
	VenomousMelee   // (non-ranged) attacks may poison foes
	WoodyLegs       // can move while lignified

	VulnerabilityFire // increases duration of fire status

	ResistanceConfusion // halves duration of confusion status
	ResistanceDaze      // allows spirit invocation while dazed
	ResistanceFear      // halves duration of fear status
	ResistanceFire      // delays burning
	ResistanceImbalance // halves duration of imbalance status
	ResistancePoison    // halves duration of poison status

	ImmunityConfusion     // immune to confusion
	ImmunityDaze          // immune to daze status effect
	ImmunityFear          // immune to fear
	ImmunityFire          // immune to fire
	ImmunityImbalance     // immune to imbalance
	ImmunityLignification // immune to lignification
	ImmunityPoison        // immune to poison

	// Monster-only traits shared by various monsters.

	MonsBerserking     // hits may make you berserk (spiders, crazy druid)
	MonsConfusion      // confusing attack (floating eyes, walking mushroom)
	MonsDig            // dig walls when hunting (boars, dragons)
	MonsExplodingDeath // explode on death (nadres, golems)
	MonsFear           // attack may frighten (lich, knight)
	MonsIgnoreDefense  // attack ignores defense (acid mound, lich)
	MonsScales         // scales provide extra protection from ranged attacks (viper, hydra, nadre, dragon, alligator)

	// NOTE: we don't do descriptions for the traits below (beyond the log
	// messages about noise), so they could be moved to another struct
	// field to make place for more traits if we ever need it.

	MonsCreep          // creep noise (instead of footsteps)
	MonsHeavyFootsteps // heavy footsteps
	MonsLightFootsteps // light footsteps
	MonsSilent         // silent movement (spiders, wraith)
	MonsWingFlap       // wing flapping noise (instead of footsteps)

	MonsNotable // notable monster (mention in timeline)

	NoTraits = 0
)

// TraitDesc returns a suitable trait description for the given actor kind and
// extra traits. For monsters, it includes the description of some
// monster-specific traits.
func (g *Game) TraitDesc(ak ActorKind, t Traits) string {
	desc := []string{}

	// Attack patterns.
	desc = traitDesc(desc, t, PatternCatch, "可远程抓取敌人并使其@O失衡@N")
	desc = traitDesc(desc, t, PatternDragging, "拖拽并使敌人@O失衡@N的攻击")
	desc = traitDesc(desc, t, PatternFourDirs, "四向攻击")
	desc = traitDesc(desc, t, PatternRampage, "狂暴冲锋")
	desc = traitDesc(desc, t, PatternRanged, "远程攻击")
	desc = traitDesc(desc, t, PatternRangedRecoil, "带后退效果的远程攻击")
	desc = traitDesc(desc, t, PatternSneaky, "潜袭式攻击")
	desc = traitDesc(desc, t, PatternSwap, "无声扭曲空间的远程攻击")
	desc = traitDesc(desc, t, PatternSwapDaze, "无声扭曲空间并可能使敌人@O眩晕@N的远程攻击")
	switch ak {
	case Player:
		// Player crocodile gets slow rotation, too.
		switch {
		case t.Any(PatternDragging):
			desc = append(desc, "转身缓慢（无法边转身边攻击）")
		case t.Any(PatternFourDirs):
			desc = append(desc, "每多一个敌人，攻击力便提高")
		case t.Any(PatternSneaky):
			desc = append(desc, "令人混乱的冲锋")
		}
	case DraggingAlligator:
		desc = append(desc, "令人混乱的撕咬")
	case ChaosMegabat:
		if !NoChaos {
			desc = append(desc, "混沌撕咬（效果随机）")
		}
		desc = append(desc, "死亡时会发出叫声")
	}

	// Extra traits.
	desc = traitDesc(desc, t, BurningHits, "命中时可能@O点燃@N敌人")
	desc = traitDesc(desc, t, DazingSpines, "近战攻击者可能被尖刺@O眩晕@N")
	desc = traitDesc(desc, t, Dazzling, "炫目（将攻击转向你身后的敌人，但容易被发现）")
	desc = traitDesc(desc, t, Elephanty, "害怕老鼠；面墙时转向困难；在死路中不会被察觉；免疫狂暴后遗症")
	desc = traitDesc(desc, t, Gawalt, "蹬墙跳移动与攻击；命中伤害降低；可感知并藏匿于石碑处")
	if g.Mod(ModTotemConditions) {
		desc = traitDesc(desc, t, Gluttony, "贪食（必须成对进食，但忽略其他进食条件）")
	} else {
		desc = traitDesc(desc, t, Gluttony, "贪食（必须成对进食）")
	}
	desc = traitDesc(desc, t, GoodSmell, "能闻到远处的食物")
	desc = traitDesc(desc, t, NocturnalFlying, "夜行（视野范围较小），可飞越植被")
	desc = traitDesc(desc, t, PushingCharge, "冲锋会推开敌人并使其@O失衡@N")
	desc = traitDesc(desc, t, RunicChicken, "能感知符文但不会触发；看见传送门或图腾时会大声咯咯笑；容易陷入混乱（持续时间更长）")
	desc = traitDesc(desc, t, ScaryRoar, "初次发现时会发出恐怖咆哮")
	desc = traitDesc(desc, t, TrailingCloud, "移动时会留下尘土")
	desc = traitDesc(desc, t, VenomousMelee, "@O带毒@N的近战攻击")
	desc = traitDesc(desc, t, WoodyLegs, "木质腿（木化时仍可移动）")
	switch {
	case t.Any(GoodHearing) && !t.Any(BadHearing):
		desc = append(desc, "听觉敏锐")
	case !t.Any(GoodHearing) && t.Any(BadHearing):
		desc = append(desc, "听觉迟钝")
	}

	// Extra monster-specific traits.
	switch ak {
	case BarkingHound:
		desc = append(desc, "发现你时会吠叫并使你@O恐惧@N")
	case BlinkButterfly:
		desc = append(desc, "命中会使你闪烁传送")
	case BurningPhoenix:
		desc = append(desc, "失衡易伤（被点燃的概率降低）")
	case ChaosMegabat:
		// Megabat hunts toward sound when out of view.
		desc = append(desc, "会循声追猎")
	case EarthDragon:
		desc = append(desc, "命中时可能推开敌人并使其@O失衡@N")
	case FireLlama:
		desc = append(desc, "会从远处喷吐@O火焰@N")
	case FourHeadedHydra:
		desc = append(desc, "四个头会同时攻击")
	case HungryRat:
		desc = append(desc, "能循着气味追猎你；若在其视野内进食，它会陷入狂暴")
	case NoisyImp:
		desc = append(desc, "会为你弹吉他，但害怕你")
	case WalkingMushroom:
		desc = append(desc, "发现你时会释放使目标@B木化@N的孢子")
	case WalkingTree:
		desc = append(desc, "命中可能使你@B木化@N；攻击无视木化的防御加成；木质腿；永久木化")
	case WarpingWraith:
		desc = append(desc, "命中会将你传送走")
	}
	desc = traitDesc(desc, t, MonsBerserking, "命中时可能使你短暂@B狂暴@N")
	desc = traitDesc(desc, t, MonsConfusion, "命中时可能使你@O混乱@N")
	desc = traitDesc(desc, t, MonsDig, "追猎时会挖穿墙壁")
	desc = traitDesc(desc, t, MonsExplodingDeath, "死亡时会爆炸")
	desc = traitDesc(desc, t, MonsFear, "命中时可能使你@O恐惧@N")
	desc = traitDesc(desc, t, MonsIgnoreDefense, "攻击无视防御")
	desc = traitDesc(desc, t, MonsScales, "鳞片（对远程攻击防御 +1）")

	// Vulnerabilities, Resistances, Immunities.
	desc = traitDesc(desc, t, VulnerabilityFire, "容易受@O火焰@N影响（持续时间更长）")
	desc = traitDesc(desc, t, ResistanceConfusion, "抵抗@O混乱@N（持续时间减半）")
	desc = traitDesc(desc, t, ResistanceDaze, "抵抗@O眩晕@N（眩晕时仍可召唤灵体）")
	desc = traitDesc(desc, t, ResistanceFear, "抵抗@O恐惧@N（持续时间减半）")
	desc = traitDesc(desc, t, ResistanceFire, "抵抗@O火焰@N（延迟燃烧）")
	desc = traitDesc(desc, t, ResistanceImbalance, "抵抗@O失衡@N（持续时间减半）")
	desc = traitDesc(desc, t, ResistancePoison, "抵抗@O中毒@N（持续时间减半）")
	desc = traitDesc(desc, t, ImmunityConfusion, "免疫@O混乱@N")
	desc = traitDesc(desc, t, ImmunityDaze, "免疫@O眩晕@N")
	desc = traitDesc(desc, t, ImmunityFear, "免疫@O恐惧@N")
	desc = traitDesc(desc, t, ImmunityFire, "免疫@O火焰@N")
	desc = traitDesc(desc, t, ImmunityImbalance, "免疫@O失衡@N")
	desc = traitDesc(desc, t, ImmunityLignification, "免疫@B木化@N")
	desc = traitDesc(desc, t, ImmunityPoison, "免疫@O中毒@N")

	return strings.Join(desc, ", ")
}

func traitDesc(desc []string, t1, t2 Traits, s string) []string {
	if t1.All(t2) {
		return append(desc, s)
	}
	return desc
}

// Status describes different kind of statuses.
type Status int

const (
	StatusBerserk Status = iota
	StatusClarity
	StatusConfusion
	StatusDaze
	StatusDig
	StatusDisorient
	StatusFear
	StatusFire
	StatusFocus
	StatusFoggySkin
	StatusGardener
	StatusGluttony
	StatusImbalance
	StatusLignification
	StatusPoison
	StatusShadow
	StatusSprint
	StatusTimeStop
	StatusVampirism
)

// Bad reports whether the status is inherently bad.
func (st Status) Bad() bool {
	switch st {
	case StatusConfusion, StatusDaze, StatusFear, StatusFire,
		StatusImbalance, StatusPoison:
		return true
	default:
		return false
	}
}

// NStatuses is the number of different kind of statuses.
const NStatuses = StatusVampirism + 1

var statusAbbrevation = []string{
	StatusBerserk:       "狂",
	StatusClarity:       "明",
	StatusConfusion:     "乱",
	StatusDaze:          "晕",
	StatusDig:           "掘",
	StatusDisorient:     "迷",
	StatusFear:          "惧",
	StatusFire:          "火",
	StatusFocus:         "专",
	StatusGardener:      "园",
	StatusGluttony:      "贪",
	StatusImbalance:     "失",
	StatusFoggySkin:     "雾",
	StatusLignification: "木",
	StatusPoison:        "毒",
	StatusShadow:        "影",
	StatusSprint:        "冲",
	StatusTimeStop:      "停",
	StatusVampirism:     "吸",
}

func (st Status) Abbr() string {
	return statusAbbrevation[st]
}

// StatusAbbr is like Abbr but accounts for actor-specific resistances.
func (g *Game) StatusAbbr(ai *Actor, st Status) string {
	switch {
	case st == StatusLignification && ai.DoesAny(WoodyLegs) ||
		st == StatusDaze && ai.DoesAny(ResistanceDaze):
		return statusAbbrevation[st] + "*"
	default:
		return statusAbbrevation[st]
	}
}

var statusName = []string{
	StatusBerserk:       "狂暴",
	StatusClarity:       "明晰",
	StatusConfusion:     "混乱",
	StatusDaze:          "眩晕",
	StatusDig:           "掘",
	StatusDisorient:     "迷向",
	StatusFear:          "恐惧",
	StatusFire:          "着火",
	StatusFocus:         "专注",
	StatusFoggySkin:     "雾肤",
	StatusGardener:      "园艺师",
	StatusGluttony:      "贪食",
	StatusImbalance:     "失衡",
	StatusLignification: "木化",
	StatusPoison:        "中毒",
	StatusShadow:        "暗影",
	StatusSprint:        "冲刺",
	StatusTimeStop:      "止时",
	StatusVampirism:     "吸血",
}

func (st Status) Name() string {
	return statusName[st]
}

var statusAdjective = []string{
	StatusBerserk:       "狂暴",
	StatusClarity:       "明晰",
	StatusConfusion:     "混乱",
	StatusDaze:          "眩晕",
	StatusDig:           "掘地",
	StatusDisorient:     "迷向",
	StatusFear:          "恐惧",
	StatusFire:          "燃烧",
	StatusFocus:         "专注",
	StatusFoggySkin:     "雾肤",
	StatusGardener:      "园艺",
	StatusGluttony:      "贪食",
	StatusImbalance:     "失衡",
	StatusLignification: "木化",
	StatusPoison:        "中毒",
	StatusShadow:        "暗影",
	StatusSprint:        "冲刺",
	StatusTimeStop:      "止时",
	StatusVampirism:     "吸血",
}

func (st Status) String() string {
	return statusAdjective[st]
}

var statusDesc = []string{
	StatusBerserk: fmt.Sprintf(
		"攻击力 %+d，造成的有效伤害 +1，并获得临时生命值加成。抵抗@O恐惧@N。效果结束后会@O中毒@N %d 回合。",
		BerserkAttackBonus, DurationPoisonBerserk),
	StatusClarity:   "抵抗@O混乱@N、@O眩晕@N和@B狂暴@N。使你能够感知附近的怪物（范围为视野的 2 倍）。",
	StatusConfusion: "使用灵体能力会使你受到 1 点伤害。新获得的@O失衡@N、@O眩晕@N和@O恐惧@N效果持续时间加倍。",
	StatusDaze:      "在受到伤害或效果结束前，除等待或进食外无法行动。",
	StatusDig:       "允许撞入墙中。冲锋时有效伤害 +1，并保证攻击会推开敌人（近战也有效）。",
	StatusDisorient: "使视野内所有追猎中的怪物朝与你当前方向相反的方向移动。它们途中可能误伤其他怪物。迷向的怪物不会进行四向冲锋。",
	StatusFear:      "无法攻击敌人或朝敌人移动。",
	StatusFire:      fmt.Sprintf("停留原地或走入火云会使你受到 %d 点伤害。首次@O着火@N时还会立即额外受到一次伤害。", FireDamage),
	StatusFocus:     "用四个头从正面发起攻击。",
	StatusFoggySkin: "向两侧和身后释放雾气。抵抗@O火焰@N和@B木化@N。",
	StatusGardener:  "每回合在半径 2 格内生长植被。",
	StatusGluttony:  fmt.Sprintf("若效果结束前没有进食，你会不消耗回合地吃下一份随机食物。若没有食物，则会陷入混乱 %d 回合。", DurationConfusionGluttony),
	StatusImbalance: fmt.Sprintf("攻击力减半。若同时处于混乱状态，则防御力 %+d。", ConfusedImbalanceDefenseBonus),
	StatusLignification: fmt.Sprintf(
		"防御力 %+d，受到的攻击伤害最多为 1，并获得临时生命值加成。抵抗@O中毒@N和@O失衡@N，延缓燃烧，但无法移动。效果结束时，你和相邻生物会@O失衡@N %d 回合。",
		LignificationDefenseBonus, DurationImbalanceLignification),
	StatusPoison: fmt.Sprintf("移动会使你受到 %d 点伤害。效果结束时会散发臭味毒素，使其他生物混乱 %d 回合。",
		PoisonDamage, DurationConfusionToxins),
	StatusShadow:   "和谐暗影会让非追猎状态的怪物无法察觉你，除非你攻击它们。你可以瞬间穿过半透明墙。战斗不会发出声音，命中敌人还会让追猎中的怪物跟丢你。",
	StatusSprint:   "你可以沿当前方向以三倍速度移动到可见地点，横向则为两倍速度。还可跳过敌人并使其失衡。等待会结束此状态。",
	StatusTimeStop: "效果结束前，除你之外所有生物的时间都会冻结；影响你的状态仍会继续计时。",
	StatusVampirism: fmt.Sprintf("攻击必定命中，并按造成的伤害恢复生命值。攻击力 %+d。",
		VampirismAttackBonus),
}

// Desc returns a description of the status, without extra details, as shown in
// ability and comestible descriptions.
func (st Status) Desc() string {
	return statusDesc[st]
}

// Details returns extra description details about the status that appear in
// help and status mouse hover.
func (st Status) Details() string {
	var details string
	switch st {
	case StatusBerserk:
		details = "失去额外生命值后，生命值不会低于 1。食用狂暴花可刷新持续时间和额外生命值。"
	case StatusClarity:
		details = "仅玩家可获得，可通过食用明晰叶获得。食用狂暴花可移除此效果，并获得一半持续时间的@B狂暴@N。"
	case StatusDaze:
		details = fmt.Sprintf("拥有眩晕抗性的玩家仍可召唤灵体；对他们而言，该状态显示为 @O%s*@N。", StatusDaze.Abbr())
	case StatusConfusion:
		details = "混乱的怪物在与其他怪物相邻时也会攻击它们。有些怪物使用挖掘、喷火、吠叫或推撞等特殊行动时，甚至会伤到自己或受到影响。"
	case StatusDisorient:
		details = "仅玩家可获得，可通过炫目斑马灵体或食用月光莲获得。"
	case StatusFear:
		details = "怪物面对你时会逃跑，但仍可能从视野外发动伏击冲锋。回合结束时，若被困且恐惧的生物无路可逃，就会陷入狂暴。木化且恐惧时受到攻击也会使你狂暴。"
	case StatusFire:
		details = "升级火蜥蜴后获得火焰抗性，不会受到额外的即时伤害。"
	case StatusFocus:
		details = "会结束@B冲刺@N。仅玩家可获得，可通过四头九头蛇灵体获得。"
	case StatusFoggySkin:
		details = "仅玩家可获得，可通过食用雾肤洋葱获得。食用木化果可移除此效果，并获得一半持续时间的@B木化@N。"
	case StatusGardener:
		details = "仅玩家可获得，可通过园艺狮灵体或食用月光莲获得。"
	case StatusGluttony:
		details = "仅玩家可获得，可通过贪食熊灵体获得。"
	case StatusImbalance:
		details = "失衡时冲刺或跳跃有时会使你@O眩晕@N倒地。"
	case StatusLignification:
		details = fmt.Sprintf("失去额外生命值后，玩家生命值不会低于 3；怪物则与@B狂暴@N一样，最低为 1。食用木化果可刷新持续时间和额外生命值。此状态会结束@B冲刺@N。“木质腿”生物不会扎根，因此没有延迟燃烧保护，但木化时仍可移动，代价是额外减少一回合持续时间；其状态标记为@B%s*@N，结束时也不会@O失衡@N，因为没有根须收回。", StatusLignification.Abbr())
	case StatusPoison:
		details = "若有怪物在中毒后存活，请小心它们可能陷入混乱！"
	case StatusShadow:
		details = "注意，攻击游荡中的怪物会使其照常开始追猎你。仅玩家可获得，可通过加瓦尔特猴灵体或食用隐匿蜗牛获得。"
	case StatusSprint:
		details = "无法进行普通攻击。仅玩家可获得，可通过疾驰瞪羚灵体获得。"
	case StatusTimeStop:
		details = "仅玩家可获得，可通过时序猫灵体获得。"
	case StatusVampirism:
		details = "仅玩家可获得，可通过吸血蝙蝠灵体或食用变形菌获得。"
	}
	return details
}

// Statuses maps ongoing statuses to their remaining turns. Slices of that type should be created with NStatuses elements.
type Statuses []int

// Put puts on a particular status for a given number of turns (if greater than current one).
func (sts Statuses) Put(st Status, turns int) bool {
	if turns <= sts[st] {
		return false
	}
	sts[st] = turns
	return true
}

// Has reports whether the given status is ongoing.
func (sts Statuses) Has(st Status) bool {
	return sts[st] > 0
}

// PutStatus1 is PutStatusN(..., 1).  It increases final duration of turns by
// one: meant for statuses that are acquired during the actor's current turn
// and don't have any effect until after the end of the turn.
func (g *Game) PutStatus1(i ID, ai *Actor, st Status, turns int) bool {
	return g.PutStatusN(i, ai, st, turns, 1)
}

// PutStatus is PutStatusN(..., false).
func (g *Game) PutStatus(i ID, ai *Actor, st Status, turns int) bool {
	return g.PutStatusN(i, ai, st, turns, 0)
}

// PutStatusN adds a status on an actor for a given number of turns. It
// adjusts final duration of turns by n.
func (g *Game) PutStatusN(i ID, ai *Actor, st Status, turns int, n int) bool {
	if ai.IsDead() {
		return false
	}
	active := ai.Has(st) // whether the status is already active
	switch st {
	case StatusBerserk:
		if active || ai.Has(StatusClarity) {
			return false
		}
	case StatusFire:
		if ai.Has(StatusFoggySkin) || ai.DoesAny(ImmunityFire) {
			return false
		}
		if ai.DoesAny(VulnerabilityFire) {
			turns += 2
		}
		if active || ai.DoesAny(ResistanceFire) || ai.Has(StatusLignification) && !ai.DoesAny(WoodyLegs) {
			// If Fire is already active or we have fire resistance
			// or Lignification without woody legs, we skip extra
			// damage at onset.
			break
		}
		g.InflictDamage(i, ai, FireDamage, AttackFireCatch)
		if ai.IsDead() {
			TotemEvents(g, &EventWithStatus{EvType: EventStatus, EvEntity: g.Entity(i), EvStatus: st})
			return false
		}
	case StatusFear:
		if ai.Has(StatusBerserk) || ai.DoesAny(ImmunityFear) || ai.Is(NoisyImp) {
			return false
		}
		if ai.DoesAny(ResistanceFear) {
			turns = (turns + 1) / 2
		}
		if ai.Has(StatusConfusion) {
			turns *= 2
		}
	case StatusDaze:
		if ai.Has(StatusClarity) || ai.DoesAny(ImmunityDaze) {
			return false
		}
		if ai.Has(StatusConfusion) {
			turns *= 2
		}
	case StatusConfusion:
		if ai.Has(StatusClarity) || ai.DoesAny(ImmunityConfusion) {
			return false
		}
		if ai.DoesAny(RunicChicken) {
			turns += (turns + 1) / 2
		}
		if ai.DoesAny(ResistanceConfusion) {
			turns = (turns + 1) / 2
		}
	case StatusImbalance:
		if ai.Has(StatusLignification) || ai.DoesAny(ImmunityImbalance) {
			return false
		}
		if ai.DoesAny(ResistanceImbalance) {
			turns = (turns + 1) / 2
		}
		if ai.Has(StatusConfusion) {
			turns *= 2
		}
	case StatusLignification:
		if ai.DoesAny(ImmunityLignification) || active || ai.Has(StatusFoggySkin) {
			return false
		}
	case StatusPoison:
		if ai.Has(StatusLignification) || ai.DoesAny(ImmunityPoison) {
			return false
		}
		if ai.DoesAny(ResistancePoison) {
			turns = (turns + 1) / 2
		}
	}
	if turns <= 0 {
		return false
	}
	turns += n
	animateStatus := func() {
		if i == PlayerID && !active {
			// Alert the player with a short animation when a new status
			// becomes active.
			if st.Bad() {
				g.md.BadStatusAnimation()
			} else {
				g.md.GoodStatusAnimation()
			}
		}
	}
	if n >= 1 {
		// We animate before getting the status, as n >= 1 currently
		// means the status is gained during the actor's turn (n == 1),
		// so showing it after would briefly show duration+1, which is
		// a bit unexpected.
		animateStatus()
	}
	if !ai.Statuses.Put(st, turns) {
		return false
	}
	if i == PlayerID {
		g.Stats.Statuses[st]++
		style := logNotable
		if st.Bad() {
			g.StoryLogf("获得%s（%s）", st, fmtTurns(turns, n))
			style = logHurtPlayer
			if !active && GameConfig.WarningsExtra && (st == StatusFire || st == StatusPoison) {
				g.WarningPrompt()
			}
		} else if st == StatusLignification || st == StatusBerserk || st == StatusFoggySkin {
			g.StoryLogf("变为%s（%s）", st, fmtTurns(turns, n))
		}
		g.LogfStyled("你处于%s状态（%s）。", style, st, fmtTurns(turns, n))
		g.md.StopAuto()
	} else if ei := g.Entity(i); g.InFOV(ei.P) {
		g.Logf("%s变为%s（%s）。", ei.Name, st, fmtTurns(turns, n))
	}
	TotemEvents(g, &EventWithStatus{EvType: EventStatus, EvEntity: g.Entity(i), EvStatus: st})
	switch st {
	case StatusBerserk:
		g.AdjustHP(i, ai, ai.HPBonus())
		g.RemoveStatus(i, ai, StatusFear)
	case StatusClarity:
		g.RemoveStatus(i, ai, StatusConfusion)
		g.RemoveStatus(i, ai, StatusDaze)
		g.RemoveStatus(i, ai, StatusBerserk)
		g.senseMonsters()
	case StatusLignification:
		g.AdjustHP(i, ai, ai.HPBonus())
		g.RemoveStatus(i, ai, StatusPoison)
		g.RemoveStatus(i, ai, StatusFoggySkin)
		g.RemoveStatus(i, ai, StatusImbalance)
		g.RemoveStatus(i, ai, StatusSprint)
	case StatusFoggySkin:
		g.RemoveStatus(i, ai, StatusLignification)
		g.RemoveStatus(i, ai, StatusFire)
	case StatusSprint:
		g.RemoveStatus(i, ai, StatusFocus)
	case StatusFocus:
		g.RemoveStatus(i, ai, StatusSprint)
	}
	if n == 0 {
		animateStatus()
	}
	return true
}

func fmtTurns(turns int, n int) string {
	turns -= n
	return CountNoun("回合", turns)
}

// HPBonus returns the amount of HP for berserk or lignification.
func (a *Actor) HPBonus() int {
	bonus := (2 * a.MaxHP) / 3
	return max(1, bonus)
}

// RemoveStatus removes a status on an actor.
func (g *Game) RemoveStatus(i ID, ai *Actor, st Status) {
	if ai.IsDead() {
		return
	}
	if ai.Statuses[st] <= 0 {
		return
	}
	ai.Statuses[st] = 0
	if i == PlayerID {
		g.LogfStyled("你不再处于%s状态。", logStatusEnd, Status(st))
		g.md.StopAuto()
	}
	switch st {
	case StatusPoison:
		dist := MaxFOVRange
		if i == PlayerID {
			g.Log("你排出带有恶臭、令人混乱的毒素。")
		} else if ei := g.Entity(i); g.InFOV(ei.P) {
			g.Logf("%s排出了带有恶臭、令人混乱的毒素。", ei.Name)
		}
		g.confusingSmell(i, dist, DurationConfusionToxins)
	case StatusBerserk:
		g.AdjustHP(i, ai, -ai.HPBonus())
		if !ai.DoesAny(Elephanty) {
			mode := g.md.mode
			g.PutStatus(i, ai, StatusPoison, DurationPoisonBerserk)
			// No warning from Berserk's Poison (got warned on
			// previous turn already).
			g.md.mode = mode
		}
	case StatusLignification:
		floor := 3
		if i != PlayerID {
			floor = 1
		}
		g.AdjustHPWithMin(i, ai, -ai.HPBonus(), floor)
		if !ai.DoesAny(WoodyLegs) {
			g.PutStatus(i, ai, StatusImbalance, DurationImbalanceLignification)
		}
		p := g.Entity(i).P
		for j, aj := range g.AdjacentActors(p) {
			if !ai.DoesAny(WoodyLegs) {
				g.PutStatus(j, aj, StatusImbalance, DurationImbalanceLignification)
			}
		}
	}
}

// randomComestible returns a random comestible among inventory and ground
// comestibles.
func (g *Game) randomComestible() (ID, *Comestible) {
	ids := g.ComestibleIDs()
	if len(ids) == 0 {
		return -1, nil
	}
	id := ids[g.IntN(len(ids))]
	return id, g.Entity(id).Role.(*Comestible)
}

// RememberStatusTurns updates the last turn the player spent with each status
// at the start of turn.
func (g *Game) RememberStatusTurns() {
	pa := g.PlayerActor()
	for st, turns := range pa.Statuses {
		if turns <= 0 {
			continue
		}
		g.StatusTurn[st] = g.Turn
	}
}

// ResistStatusRoll reports whether the actor got lucky and was granted
// resistance to the given status. The intention is similar as with lucky rolls
// when at low HP: make it so the player feels unlucky less often. It's only
// used for close re-applications due to some chance-based on-hit effects;
// deterministic applications shouldn't use such a resist roll.
func (g *Game) ResistStatusRoll(i ID, st Status) bool {
	if i != PlayerID {
		return false
	}
	switch g.Turn - g.StatusTurn[st] {
	case 0:
		// A decent chance to resist re-application of
		// statuses if they're already active or just expired
		// on the same turn.
		// In particular, this reduces the likeliness of an
		// annoying immediate re-application of Daze or a
		// helpful Berserk or Lignification.
		return g.IntN(2) == 0
	case 1:
		switch st {
		case StatusDaze, StatusFire, StatusImbalance:
			// A decent chance to resist re-application of some
			// statuses that can feel annoying if repeated on the
			// turn following expiration without a pause. In
			// particular, this make consecutive fire or imbalance
			// dances less likely.
			return g.IntN(2) == 0
		}
	}
	return false
}

// ProgressStatus reduces the number of remaining turns for a status by one.
func (g *Game) ProgressStatus(i ID, ai *Actor, st Status) {
	turns := ai.Statuses[st]
	if turns <= 1 {
		g.RemoveStatus(i, ai, st)
		return
	}
	ai.Statuses[st]--
	if turns > 2 || i != PlayerID || !GameConfig.WarningsExtra {
		return
	}
	switch st {
	case StatusBerserk, StatusLignification, StatusDisorient, StatusGluttony, StatusShadow:
		// For statuses where that matters, warn when there's
		// only one turn remaining after progression.
		g.WarningPrompt()
	}
}

// Behavior holds simple Behavior data for a monster.
type Behavior struct {
	Path     []gruid.Point // path to destination
	Guard    []gruid.Point // position(s) to guard (empty if not)
	Target   gruid.Point   // Current goal (if not static)
	State    Mindstate     // Wandering, Hunting, ...
	SkipTurn bool          // skip turn (like after swapped positions with another monster)
}

// Mindstate represents the kind of monster behavioral states.
type Mindstate int

// Those constants represent all the possible monster mindstates.
const (
	Wandering Mindstate = iota
	Hunting
)

// MindStateString returns a text describing a monster's main state.
func (g *Game) MindStateString(a *Actor) string {
	beh := a.Behavior
	if beh == nil {
		return ""
	}
	switch beh.State {
	case Hunting:
		return "追猎中"
	default:
		if g.Mod(ModCorruptedDungeon) {
			return ""
		}
		if len(beh.Guard) > 0 {
			return "守卫中"
		}
		return "游荡中"
	}
}

// HandleActorTurn handles turn status updates for an actor with given ID, as
// well as monster behavior for non-player actors.
func (g *Game) HandleActorTurn(i ID, ai *Actor) {
	ei := g.Entity(i)
	if ai.IsDead() {
		return
	}
	if ai.Behavior != nil {
		g.UpdateFirePos(i, ai)
		g.HandleMonsterTurn(i, ai)
	}
	if ai.Has(StatusFire) && (ei.P == ai.FirePos || ai.FirePos != InvalidPos && g.FireAt(ei.P)) {
		g.InflictDamage(i, ai, FireDamage, AttackFire)
	}
	if ai.IsDead() {
		// The actor died due to fire or poison (the latter only for a
		// monster, as it would have happened earlier for the player).
		return
	}
	g.handleCornered(i, ai)
	if i == PlayerID {
		// Gluttony is handled separately from other statuses, because
		// it may lead to eating and may hence mess with other
		// statuses, so it should happen before the progression of
		// other statuses, as well as before some special effects like
		// exuding fog.
		g.handleGluttony()
		if ai.IsDead() {
			return
		}
		if ai.Has(StatusFoggySkin) {
			g.exudeFog()
		}
		if ai.Has(StatusGardener) {
			g.garden()
		}
	}
	// NOTE: We make statuses progress at the end of the turn. That means
	// bonus HP from expiring Berserk or Lignification stills protects
	// against Fire or Poison damage happening during the player's turn but
	// not against any damage inflicted by monsters acting later.
	// While not perfect, the alternative is problematic in other ways
	// (many monster statuses would only be partially relevant on their
	// last turn, making it confusing for the player from an UI point of
	// view).
	for _, st := range sortedStatuses {
		g.ProgressStatus(i, ai, st)
	}
}

// handleGluttony handles progression of the Gluttony status.
func (g *Game) handleGluttony() {
	if GluttonyRework {
		return
	}
	pa := g.PlayerActor()
	switch {
	case pa.Statuses[StatusGluttony] <= 0:
	case pa.Statuses[StatusGluttony] == 1:
		// On Gluttony expiration, the Gluttonous Bear cannot control
		// itself anymore and eats a random comestible, or becomes
		// confused if it cannot find one.
		if id, co := g.randomComestible(); id >= 0 {
			switch co.Effect.(type) {
			case EffectTeleportMushroom, EffectWarpingBean:
				g.UpdateFOV()
				g.UpdateKnowledge()
			}
			g.LogStyled("你再也无法忍受饥饿了！", logStatusEnd)
			co.Use(g, id)
		} else {
			pa.Statuses[StatusGluttony] = 0
			g.LogStyled("你四处寻找，却找不到任何食物！", logStatusEnd)
			g.PutStatus1(PlayerID, pa, StatusConfusion, DurationConfusionGluttony)
		}
	default:
		pa.Statuses[StatusGluttony]--
	}
}

// sortedStatuses contains statuses sorted for progression dependencies,
// so that it's easier to get durations right in a consistent way. NOTE: it
// allows Poison to expire before being applied again too (which may or may not
// make more sense than just prolonging it, though it's a quite rare
// edge-case).
var sortedStatuses = []Status{
	StatusClarity,
	StatusConfusion,
	StatusDaze,
	StatusDig,
	StatusDisorient,
	StatusFear,
	StatusFire,
	StatusFocus,
	StatusFoggySkin,
	StatusGardener,
	StatusImbalance,
	StatusPoison,
	StatusShadow,
	StatusSprint,
	StatusTimeStop,
	StatusVampirism,

	StatusBerserk,       // after Poison
	StatusLignification, // after Imbalance & Berserk
}

// exudeFog implements the fog for the player-only status effect Foggy-Skin.
func (g *Game) exudeFog() {
	pp := g.PP()
	for q := range g.Map.PassableNeighbors(pp) {
		if q == pp.Add(g.Dir) {
			// NOTE: this assumes exudeFog is only called for the
			// player for now (currently foggy skin status is
			// player-only).
			continue
		}
		g.NormalCloudAt(q, 2+g.IntN(3))
	}
}

// garden implements gardening for the player-only status effect Gardener.
func (g *Game) garden() {
	pp := g.PP()
	mp := &MapPath{passable: g.Map.Passable}
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{pp}, 2)
	for _, n := range nodes {
		if i, _ := g.ItemAt(n.P); i >= 0 {
			continue
		}
		if g.Map.Terrain.At(n.P) == Floor {
			g.Map.Terrain.Set(n.P, Foliage)
		}
	}
}

// handleCornered makes cornered and afraid actors become berserk.
func (g *Game) handleCornered(i ID, ai *Actor) {
	switch {
	case ai.Has(StatusClarity):
		// Players under clarity cannot become Berserk.
		return
	case ai.Has(StatusFear) || ai.Is(NoisyImp) && !ai.Has(StatusBerserk):
		// Actor is afraid or is playing music while not berserk: it
		// may become berserk if cornered.
	case i == PlayerID && ai.DoesAny(Elephanty):
		g.handleCorneredWithRatPhobia()
		return
	default:
		// The actor isn't afraid of anything, so it cannot feel
		// cornered.
		return
	}
	ei := g.Entity(i)
	from := ei.P
	if !g.InFOV(from) {
		return
	}
	for _, dir := range Directions {
		p := from.Add(dir)
		if !inMap(p) {
			continue
		}
		if !g.Map.Passable(p) && !ai.Has(StatusDig) && !ai.DoesAny(MonsDig) {
			continue
		}
		if j, _ := g.ActorAt(p); j >= 0 {
			continue
		}
		j, _ := g.RangedTargetInDir(from, dir)
		if j < 0 {
			// One direction is free of any actors: the afraid
			// actor is not cornered yet.
			return
		}
		if i != PlayerID && j != PlayerID {
			// The afraid monster is not cornered yet: one of the
			// directions has free space and no player.
			return
		}
	}
	// The actor is afraid and cornered: become berserk for a short while.
	if i == PlayerID {
		g.Log("你感到无路可退。")
		g.Stats.Cornered++
	} else if g.InFOV(from) {
		g.LogfStyled("恐惧的%s感到无路可退。", logSpecial, ei.Name)
	}
	g.PutStatus1(i, ai, StatusBerserk, DurationBerserkAfraid)
}

func (g *Game) handleCorneredWithRatPhobia() {
	from, pa := g.PP(), g.PlayerActor()
	for _, dir := range Directions {
		p := from.Add(dir)
		if !inMap(p) {
			continue
		}
		if !g.Map.Passable(p) && !pa.Has(StatusDig) && !pa.DoesAny(MonsDig) {
			continue
		}
		if j, aj := g.ActorAt(p); j >= 0 && g.PlayerFears(j, aj) {
			continue
		}
		if j, aj := g.RangedTargetInDir(from, dir); j < 0 || !g.PlayerFears(j, aj) {
			// One direction is free of any fearsome actors: the
			// player is not cornered yet.
			return
		}
	}
	// You feel cornered: become berserk for a short while.
	g.Log("你感到无路可退。")
	g.Stats.Cornered++
	g.PutStatus1(PlayerID, pa, StatusBerserk, DurationBerserkAfraid)
}

// TeleportActor teleports the given actor, adjusting final duration of debuff
// statuses by n.
func (g *Game) TeleportActor(i ID, ai *Actor, n int) bool {
	if ai.ResistsMove() {
		return false
	}
	from := g.Entity(i).P
	to := g.TeleportPoint(from)
	if i == PlayerID && g.Map.Orb != InvalidPos && g.IntN(2) == 0 {
		// On last level, give the orb a chance to drive the player
		// away from it, to make teleport-based strategies less
		// reliable (but still viable enough!).
		//
		// Result: teleport to furthest point from the orb among 3
		// random possible teleport points.
		for range 2 {
			if nto := g.TeleportPoint(from); paths.DistanceManhattan(nto, g.Map.Orb) > paths.DistanceManhattan(to, g.Map.Orb) {
				to = nto
			}
		}
		g.LogStyled("一股奇异能量干扰了你的传送！", logSpecial)
	}
	g.MoveActor(i, ai, to, MovTeleport)
	g.teleportStatuses(i, ai, n)
	if i == PlayerID {
		TotemEvents(g, &EventHappened{EvType: EventTeleport})
	}
	return true
}

// TeleportPoint returns a suitable teleport destination for the current given
// position. In particular, it ensures that the destination is floor, rubble or
// foliage, unoccupied by any other actor or trap, and not too close to the
// origin.
func (g *Game) TeleportPoint(from gruid.Point) gruid.Point {
	const mindist = 2 * MaxFOVRange
	for {
		p := g.RandomPassableWithoutTrap()
		if paths.DistanceManhattan(from, p) < mindist {
			continue
		}
		if i, _ := g.ActorAt(p); i >= 0 {
			continue
		}
		return p
	}
}
