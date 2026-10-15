// This file implements effects for comestibles (except conditional ones),
// abilities, and menhirs.

package main

import (
	"fmt"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
)

// The following constants are used for various effects and statuses.
const (
	DurationBerserk                = 8
	DurationBerserkAfraid          = 4
	DurationBerserkHit             = 4
	DurationBerserkStomp           = 4
	DurationBerserkTotem           = 5
	DurationBerserkTrap            = 5
	DurationClarity                = 5
	DurationConfusionAmbrosia      = 6
	DurationConfusionFireTrap      = 3
	DurationConfusionGluttony      = 6
	DurationConfusionHit           = 6
	DurationConfusionPepperBreath  = 2
	DurationConfusionSkunk         = 5
	DurationConfusionToxins        = 4
	DurationDazeFall               = 2
	DurationDazeJump               = 4
	DurationDazeLightning          = 4
	DurationDazeSpines             = 2
	DurationDazeTeleport           = 4
	DurationFearAmbrosia           = 4
	DurationFearBark               = 4
	DurationFearDeathStare         = 4
	DurationFearFakePortal         = 6
	DurationFearRoar               = 4
	DurationFearUndead             = 3
	DurationFire                   = 3
	DurationFoggySkin              = 10
	DurationImbalanceDragging      = 2
	DurationImbalanceGale          = 5
	DurationImbalanceLignification = 5
	DurationImbalanceSprint        = 5
	DurationImbalanceTailSlap      = 4
	DurationImbalanceTeleport      = 5
	DurationLignification          = 10
	DurationLignificationHit       = 6
	DurationLignificationRoots     = 6
	DurationLignificationSpores    = 6
	DurationLignificationTrap      = 7
	DurationPoisonBerserk          = 5
	DurationPoisonBite             = 4
	DurationPoisonCloud            = 3

	HealAmbrosia          = 8
	HealAmbrosiaHC        = 2
	HealClarityLeaves     = 4
	HealFirebreathPepper  = 2
	HealFoggySkinOnion    = 3
	HealMoonlightLotus    = 2
	HealTeleportLignified = 2
	HealWarpingBean       = 2

	DurationDig       = 5
	DurationDisorient = 6
	DurationGarDis    = 8
	DurationFocus     = 3
	DurationShadow    = 8
	DurationGarden    = 7
	DurationGluttony  = 7
	DurationSprint    = 6
	DurationTimeStop  = 5
	DurationVampirism = 2
	DurationVampFung  = 3
)

// newHPMsg returns a message with new HP after healing an actor for the given
// amount after eating.
func newHPMsg(a *Actor, hp int) string {
	return fmt.Sprintf("%d/%d", max(0, min(a.GetMaxHP(), a.HP+hp)), a.GetMaxHP())
}

// healMsg returns an HP restoring description for the given amount if Healing
// Combat is not enabled.
func (g *Game) healMsg(amount int) string {
	if g.Mod(ModHealingCombat) {
		return ""
	}
	return fmt.Sprintf(" 恢复 %+d 点生命值。", amount)
}

// eatMsg returns an eating message with the given optional new hp info.
func (g *Game) eatMsg(name string, hpstr string) {
	g.Logf("你吃下了%s。", name)
	if hpstr != "" {
		g.StoryLogf("吃下%s（生命值：%s）", One(name), hpstr)
	} else {
		g.StoryLogf("吃下%s", One(name))
	}
}

// ComestibleCount returns the number of comestibles in the inventory.
func (g *Game) ComestibleCount() int {
	var n int
	for range g.PlayerComestibles() {
		n++
	}
	return n
}

// healPlayer restores HP and emits a healing event for spirit conditions
func (g *Game) healPlayer(ai *Actor, amount int) {
	hp := ai.HP
	g.AdjustHP(PlayerID, ai, amount)
	if ai.HP > hp {
		TotemEvents(g, &EventHappened{EvType: EventHeal})
	}
}

type EffectAmbrosiaBerries struct{}

func (eff EffectAmbrosiaBerries) Apply(g *Game) bool {
	pa := g.PlayerActor()
	heal := HealAmbrosia
	if g.Mod(ModHealingCombat) {
		heal = HealAmbrosiaHC
	}
	g.eatMsg(AmbrosiaBerries.Name(), newHPMsg(pa, heal))
	g.healPlayer(pa, heal)
	if g.IntN(2) == 0 {
		g.PutStatus1(PlayerID, pa, StatusConfusion, DurationConfusionAmbrosia)
	} else {
		g.PutStatus1(PlayerID, pa, StatusFear, DurationFearAmbrosia)
	}
	return true
}

func (eff EffectAmbrosiaBerries) Desc(g *Game) string {
	heal := HealAmbrosia
	if g.Mod(ModHealingCombat) {
		heal = HealAmbrosiaHC
	}
	return fmt.Sprintf("恢复 %+d 点生命值。随后你会陷入@O混乱@N或@O恐惧@N，分别持续 %d 或 %d 回合。",
		heal, DurationConfusionAmbrosia, DurationFearAmbrosia)
}

type EffectBerserkingFlower struct{}

func (eff EffectBerserkingFlower) Apply(g *Game) bool {
	pa := g.PlayerActor()
	d := DurationBerserk
	g.eatMsg(BerserkingFlower.Name(), "")
	if pa.Has(StatusClarity) {
		// Eating a berserking flower while clear-minded works but
		// reduces the duration by half (rounding up).
		g.RemoveStatus(PlayerID, pa, StatusClarity)
		d = (d + 1) / 2
	}
	if pa.Has(StatusBerserk) {
		// The berserking flower can be used to renew berserk duration
		// and temporary HP bonus.
		pa.Statuses[StatusBerserk] = max(pa.Statuses[StatusBerserk], d+1)
		bonusHP := pa.HP - max(1, pa.HP-pa.HPBonus())
		g.AdjustHP(PlayerID, pa, pa.HPBonus()-bonusHP)
		return true
	}
	g.PutStatus1(PlayerID, pa, StatusBerserk, d)
	return true
}

func (eff EffectBerserkingFlower) Desc(_ *Game) string {
	return fmt.Sprintf("使你进入@B狂暴@N状态，持续 %d 回合。%s",
		DurationBerserk, statusDesc[StatusBerserk])
}

type EffectLignificationFruit struct{}

func (eff EffectLignificationFruit) Apply(g *Game) bool {
	pa := g.PlayerActor()
	d := DurationLignification
	g.eatMsg(LignificationFruit.Name(), "")
	if pa.Has(StatusFoggySkin) {
		// Eating a lignification fruit while foggy works but reduces
		// the duration by half (rounding up).
		g.RemoveStatus(PlayerID, pa, StatusFoggySkin)
		d = (d + 1) / 2
	}
	if pa.Has(StatusLignification) {
		// The lignification fruit can be used to renew lignification
		// duration and temporary HP bonus.
		pa.Statuses[StatusLignification] = max(pa.Statuses[StatusLignification], d+1)
		bonusHP := pa.HP - max(min(3, pa.HP), pa.HP-pa.HPBonus())
		g.AdjustHP(PlayerID, pa, pa.HPBonus()-bonusHP)
		return true
	}
	g.PutStatus1(PlayerID, pa, StatusLignification, d)
	return true
}

func (eff EffectLignificationFruit) Desc(_ *Game) string {
	return fmt.Sprintf("进入@B木化@N状态，持续 %d 回合。%s",
		DurationLignification, statusDesc[StatusLignification])
}

type EffectClarityLeaves struct{}

func (eff EffectClarityLeaves) Apply(g *Game) bool {
	pa := g.PlayerActor()
	var hpstr string
	switch {
	case g.Mod(ModHealingCombat):
	case pa.Has(StatusBerserk):
		hpstr = newHPWithoutBonus(pa, 1, HealClarityLeaves)
	default:
		hpstr = newHPMsg(pa, HealClarityLeaves)
	}
	extraTurns := pa.Statuses[StatusClarity]
	g.eatMsg(ClarityLeaves.Name(), hpstr)
	g.PutStatus1(PlayerID, pa, StatusClarity, DurationClarity+extraTurns)
	if !g.Mod(ModHealingCombat) {
		g.healPlayer(pa, HealClarityLeaves)
	}
	return true
}

func newHPWithoutBonus(a *Actor, m, hp int) string {
	hpbonus := a.HPBonus()
	var nhp int
	if a.HP < m {
		nhp = a.HP
	} else {
		nhp = max(m, a.HP-hpbonus)
	}
	return fmt.Sprintf("%d/%d", nhp+hp, a.GetMaxHP()-hpbonus)
}

func (eff EffectClarityLeaves) Desc(g *Game) string {
	return fmt.Sprintf("获得@B明晰@N状态，持续 %d 回合。%s%s",
		DurationClarity, statusDesc[StatusClarity], g.healMsg(HealClarityLeaves))
}

type EffectFoggySkinOnion struct{}

func (eff EffectFoggySkinOnion) Apply(g *Game) bool {
	pa := g.PlayerActor()
	var hpstr string
	switch {
	case g.Mod(ModHealingCombat):
	case pa.Has(StatusLignification):
		hpstr = newHPWithoutBonus(pa, 3, HealFoggySkinOnion)
	default:
		hpstr = newHPMsg(pa, HealFoggySkinOnion)
	}
	extraTurns := pa.Statuses[StatusFoggySkin]
	g.eatMsg(FoggySkinOnion.Name(), hpstr)
	g.PutStatus1(PlayerID, pa, StatusFoggySkin, DurationFoggySkin+extraTurns)
	if !g.Mod(ModHealingCombat) {
		g.healPlayer(pa, HealFoggySkinOnion)
	}
	return true
}

func (eff EffectFoggySkinOnion) Desc(g *Game) string {
	return fmt.Sprintf("获得@B雾肤@N状态，持续 %d 回合。%s%s",
		DurationFoggySkin, statusDesc[StatusFoggySkin], g.healMsg(HealFoggySkinOnion))
}

type EffectFirebreathPepper struct{}

func (eff EffectFirebreathPepper) Apply(g *Game) bool {
	pp := g.PP()
	pa := g.PlayerActor()
	var hpstr string
	if !g.Mod(ModHealingCombat) {
		hpstr = newHPMsg(pa, HealFirebreathPepper)
	}
	g.eatMsg(FirebreathPepper.Name(), hpstr)
	g.MakeNoise(pp, NoiseRoar)
	if !g.Mod(ModHealingCombat) {
		g.healPlayer(pa, HealFirebreathPepper)
	}
	g.md.FirebreathAnimation(pp.Add(g.Dir), g.Dir)
	p := pp
	for range MaxFOVRange {
		p = p.Add(g.Dir)
		t := g.Map.Terrain.At(p)
		if !Passable(t) {
			g.ExplosionAt(PlayerID, p)
			return true
		}
		if t == Foliage {
			g.FireCloudAt(p)
		}
		if j, aj := g.ActorAt(p); j >= 0 {
			extraTurns := aj.Statuses[StatusConfusion]
			g.PutStatus(j, aj, StatusConfusion, DurationConfusionPepperBreath+extraTurns)
			extraTurns = aj.Statuses[StatusFire]
			g.PutStatus(j, aj, StatusFire, DurationFire+extraTurns)
		}
	}
	// Leave cloud at the last tile if passable.
	g.NormalCloudAt(p, 2+g.IntN(3))
	return true
}

func (eff EffectFirebreathPepper) Desc(g *Game) string {
	return fmt.Sprintf("向当前方向喷出令人混乱的火焰。使目标陷入@O混乱@N %d 回合并@O着火@N %d 回合。点燃植被，形成持续 6–12 回合的火云，并炸毁墙壁。%s",
		DurationConfusionPepperBreath, DurationFire, g.healMsg(HealFirebreathPepper))
}

type EffectTeleportMushroom struct{}

func (eff EffectTeleportMushroom) Apply(g *Game) bool {
	if GluttonyRework {
		// In case the bear teleported or warped first.
		g.UpdateFOV()
		g.UpdateKnowledge()
	}
	pa := g.PlayerActor()
	j := g.TotemID()
	g.eatMsg(TeleportMushroom.Name(), "")
	if !g.TeleportActor(PlayerID, pa, 1) {
		g.Log("木化期间你能抵抗传送。")
		if !g.Mod(ModHealingCombat) {
			g.healPlayer(pa, HealTeleportLignified)
		}
	}
	if j >= 0 {
		g.SenseEntity(j, "感知到")
	}
	return true
}

func (g *Game) teleportStatuses(i ID, ai *Actor, n int) {
	if g.IntN(2) == 0 {
		g.PutStatusN(i, ai, StatusDaze, DurationDazeTeleport, n)
	} else {
		g.PutStatusN(i, ai, StatusImbalance, DurationImbalanceTeleport, n)
	}
}

func (eff EffectTeleportMushroom) Desc(g *Game) string {
	var healMsg string
	if !g.Mod(ModHealingCombat) {
		healMsg = fmt.Sprintf(" 木化时则恢复 %+d 点生命值。", HealTeleportLignified)
	}
	return fmt.Sprintf("将你传送离开，并使你感知所有图腾灵体。随后你会陷入@O眩晕@N或@O失衡@N，分别持续 %d 或 %d 回合。%s",
		DurationDazeTeleport, DurationImbalanceTeleport, healMsg)
}

type EffectMoonlightLotus struct{}

func (eff EffectMoonlightLotus) Apply(g *Game) bool {
	pa := g.PlayerActor()
	pp := g.PP()
	var hpstr string
	if !g.Mod(ModHealingCombat) {
		hpstr = newHPMsg(pa, HealMoonlightLotus)
	}
	g.eatMsg(MoonlightLotus.Name(), hpstr)
	mp := &MappingPath{passable: g.Map.Passable}
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{pp}, MaxFOVRange)
	draw, cost := false, 1
	if g.Map.Terrain.At(pp) == Foliage {
		g.PutStatus1(PlayerID, pa, StatusDisorient, DurationGarDis)
		for _, n := range nodes {
			if n.Cost > cost && draw {
				g.md.AnimationFrameFast()
				draw, cost = false, n.Cost
			}
			switch g.Map.Terrain.At(n.P) {
			case Foliage:
				draw = true
				g.Map.Terrain.Set(n.P, Floor)
				if g.Map.KnownTerrain.At(n.P) == Foliage {
					g.Map.KnownTerrain.Set(n.P, Floor)
				}
			case Wall:
				draw = true
				g.Map.Terrain.Set(n.P, TranslucentWall)
				if g.Map.KnownTerrain.At(n.P) == Wall {
					g.Map.KnownTerrain.Set(n.P, TranslucentWall)
				}
			}
		}
		// Update FOV and player knowledge.
		g.UpdateFOV()
		g.UpdateKnowledge()
	} else {
		g.PutStatus1(PlayerID, pa, StatusGardener, DurationGarDis)
		for _, n := range nodes {
			if n.Cost > cost && draw {
				g.md.AnimationFrameFast()
				draw, cost = false, n.Cost
			}
			switch g.Map.Terrain.At(n.P) {
			case TranslucentWall:
				draw = true
				g.Map.Terrain.Set(n.P, Wall)
				if g.Map.KnownTerrain.At(n.P) == TranslucentWall {
					g.Map.KnownTerrain.Set(n.P, Wall)
				}
			}
		}
		// Updating just FOV is enough, because it becomes smaller.
		g.UpdateFOV()
	}
	return true
}

func (eff EffectMoonlightLotus) Desc(g *Game) string {
	var healmsg string
	if !g.Mod(ModHealingCombat) {
		healmsg = fmt.Sprintf("\n无论如何，都会恢复 %+d 点生命值。", HealMoonlightLotus)
	}
	return fmt.Sprintf("若你站在植被上，则获得@B迷向@N状态，持续 %d 回合，并提高周围能见度，将普通墙壁变为半透明墙、将植被变为地面。\n否则，获得@B园艺师@N状态，持续 %d 回合，并将半透明墙变回普通墙。%s", DurationGarDis, DurationGarDis, healmsg)
}

type EffectFocus struct{}

func (eff EffectFocus) Apply(g *Game) bool {
	pa := g.PlayerActor()
	if pa.Has(StatusDaze) {
		g.Log("眩晕时使用“专注”毫无作用。")
		return false
	}
	if pa.Has(StatusFocus) {
		g.Log("你已处于专注状态，无法再次使用“专注”。")
		return false
	}
	if pa.Has(StatusFear) {
		g.Log("恐惧时无法使用“专注”。")
		return false
	}
	g.PutStatus(PlayerID, pa, StatusFocus, DurationFocus)
	g.instant = true
	return true
}

func (eff EffectFocus) Name() string {
	return "专注"
}

func (eff EffectFocus) Desc(_ *Game) string {
	return fmt.Sprintf("进入@B专注@N状态，持续 %d 回合。%s @G即时生效@N。", DurationFocus, statusDesc[StatusFocus])
}

type EffectDig struct{}

func (eff EffectDig) Apply(g *Game) bool {
	pa := g.PlayerActor()
	if pa.Has(StatusDaze) {
		g.Log("眩晕时使用“掘地”毫无作用。")
		return false
	}
	if pa.Has(StatusDig) {
		g.Log("你已处于掘地状态，无法再次使用“掘地”。")
		return false
	}
	g.PutStatus(PlayerID, pa, StatusDig, DurationDig)
	g.instant = true
	return true
}

func (eff EffectDig) Name() string {
	return "掘地"
}

func (eff EffectDig) Desc(_ *Game) string {
	return fmt.Sprintf("进入@B掘地@N状态，持续 %d 回合。%s @G即时生效@N。",
		DurationDig, statusDesc[StatusDig])
}

type EffectJump struct{}

func (eff EffectJump) Apply(g *Game) bool {
	pl, pa := g.Player()
	if !pa.CanMove() {
		g.Log("木化时无法跳跃。")
		return false
	}
	if pa.HP == 1 && pa.Has(StatusPoison) {
		g.Log("中毒且濒死时无法跳跃。")
		return false
	}
	dir, from, to := g.Dir, pl.P, pl.P
	for p := to.Add(dir); g.Map.Passable(p) && g.InFOV(p); p = p.Add(dir) {
		to = p
	}
	for to != from {
		if i, _ := g.ActorAt(to); i < 0 {
			break
		}
		to = to.Sub(dir)
	}
	if to == from {
		g.Log("那个方向没有足够空间跳跃。")
		return false
	}
	g.Log("你跳了起来。")
	g.MoveActor(PlayerID, pa, to, MovJump)
	for p := from.Add(dir); p != to; p = p.Add(dir) {
		if i, ai := g.ActorAt(p); i >= 0 {
			if !g.PlayerFears(i, ai) && !pa.Has(StatusDaze) {
				if pa.Has(StatusSprint) {
					g.PutStatus(i, ai, StatusImbalance, DurationImbalanceSprint)
				} else {
					g.BumpAttackActor(PlayerID, i, pa, ai, AttackPlain, 1, 1)
				}
			}
			g.PutStatus(i, ai, StatusDaze, DurationDazeJump)
		}
	}
	if pa.Has(StatusImbalance) && g.IntN(5) < 2 {
		g.Log("你因失衡而跌倒。")
		duration := DurationDazeFall + paths.DistanceManhattan(from, to)/3
		g.PutStatus1(PlayerID, pa, StatusDaze, duration)
	}
	g.TriggerTrap(PlayerID, pa)
	return true
}

func (eff EffectJump) Name() string {
	return "眩晕跳跃"
}

func (eff EffectJump) Desc(_ *Game) string {
	return fmt.Sprintf("沿当前方向跳至视野内最远的空位。途中会攻击所有怪物，攻击力 +1，并使其@O眩晕@N %d 回合。",
		DurationDazeJump)
}

type EffectPushingGale struct{}

func (eff EffectPushingGale) Apply(g *Game) bool {
	pp := g.PP()
	g.Log("强风将周围的空气推开。")
	g.MakeNoise(pp, NoiseWind)
	for _, dir := range Directions {
		g.pushingGale(pp, dir)
	}
	return true
}

func (g *Game) pushingGale(from, dir gruid.Point) {
	const maxdist = MaxFOVRange
	at := from
	for n, p := 0, at.Add(dir); n < maxdist && g.Map.Passable(p); p = p.Add(dir) {
		g.blowClouds(p, dir)
		at = p
		n++
	}
	pp := g.PP()
	for ; at != from; at = at.Sub(dir) {
		i, ai := g.ActorAt(at)
		if i < 0 {
			continue
		}
		ei := g.Entity(i)
		dmg := 1
		if paths.DistanceManhattan(ei.P, pp) == 1 && g.IntN(3) > 0 {
			dmg++
		}
		g.InflictDamage(i, ai, dmg, AttackGale)
		if ai.IsDead() {
			continue
		}
		if ai.ResistsMove() {
			g.md.HitAnimation(at)
			continue
		}
		g.PutStatus(i, ai, StatusImbalance, DurationImbalanceGale)
		to := at.Add(dir)
		n := 0
		for p := to; g.IsFree(p) && n < 2; p = p.Add(dir) {
			to = p
			n++
		}
		if g.IsFree(to) {
			g.MoveActor(i, ai, to, MovPlain)
		} else {
			g.md.HitAnimation(at)
		}
	}
}

func (g *Game) blowClouds(p, dir gruid.Point) {
	if g.Map.Terrain.At(p) != Foliage || g.Map.Clouds.At(p).Kind != CloudFire {
		// The strong gale dissipates the cloud (except if it's a fire
		// cloud on a dense foliage cell).
		g.RemoveCloudAt(p)
		return
	}
	if p = p.Add(dir); g.Map.Burnable(p) && !g.FireAt(p) {
		// Make fire on foliage propagate in the given direction.
		g.FireCloudAt(p)
	}
}

func (eff EffectPushingGale) Name() string {
	return "推风"
}

func (eff EffectPushingGale) Desc(_ *Game) string {
	return fmt.Sprintf("伤害四个正交方向上的敌人：远程造成 1 点伤害，近战造成 1–2 点。将敌人推开并使其@O失衡@N %d 回合。驱散云雾，但植被火焰会因此蔓延。",
		DurationImbalanceGale)
}

type EffectTimeStop struct{}

func (eff EffectTimeStop) Apply(g *Game) bool {
	pa := g.PlayerActor()
	if pa.Has(StatusTimeStop) {
		g.Log("你已处于止时状态，无法再次使用“止时”。")
		return false
	}
	g.PutStatus(PlayerID, pa, StatusTimeStop, DurationTimeStop)
	g.instant = true
	return true
}

func (eff EffectTimeStop) Name() string {
	return "止时"
}

func (eff EffectTimeStop) Desc(_ *Game) string {
	return fmt.Sprintf("进入@B止时@N状态，持续 %d 回合。%s @G即时生效@N。",
		DurationTimeStop, statusDesc[StatusTimeStop])
}

type EffectTailSlap struct{}

func (eff EffectTailSlap) Apply(g *Game) bool {
	var actors []*Actor
	var ids []ID
	for j, aj := range g.AdjacentActors(g.PP()) {
		actors = append(actors, aj)
		ids = append(ids, j)
	}
	// Shuffle foes so that attack order is not predictable in an
	// artificial way.
	g.rand.Shuffle(len(actors), func(i, j int) {
		actors[i], actors[j] = actors[j], actors[i]
		ids[i], ids[j] = ids[j], ids[i]
	})
	g.Log("你转着圈追自己的尾巴。")
	for k, j := range ids {
		// Tail-slap works even under fear, as it's not really
		// directed. NOTE: we should maybe remove one direction under
		// fear?
		g.BumpAttackActor(PlayerID, j, g.PlayerActor(), actors[k], AttackSlap, 1, 0)
		g.PutStatus(j, actors[k], StatusImbalance, DurationImbalanceTailSlap)
	}
	g.Dir = g.Dir.Mul(-1)
	return true
}

func (eff EffectTailSlap) Name() string {
	return "甩尾"
}

func (eff EffectTailSlap) Desc(_ *Game) string {
	return fmt.Sprintf("转身甩尾攻击相邻敌人，造成的有效伤害 +1，并使其@O失衡@N %d 回合。",
		DurationImbalanceTailSlap)
}

type EffectVampirism struct{ Duration int }

func (eff EffectVampirism) Apply(g *Game) bool {
	pa := g.PlayerActor()
	if pa.Has(StatusDaze) {
		g.Log("眩晕时使用“吸血”毫无作用。")
		return false
	}
	if pa.Has(StatusVampirism) {
		g.Log("你已处于吸血状态，无法再次使用“吸血”。")
		return false
	}
	if pa.Has(StatusFear) {
		g.Log("恐惧时无法使用“专注”。")
		return false
	}
	g.PutStatus(PlayerID, pa, StatusVampirism, eff.Duration)
	g.instant = true
	return true
}

func (eff EffectVampirism) Name() string {
	return "吸血"
}

func (eff EffectVampirism) Desc(_ *Game) string {
	return fmt.Sprintf("进入@B吸血@N状态，持续 %d 回合。%s @G即时生效@N。", eff.Duration, statusDesc[StatusVampirism])
}

type EffectLightning struct{}

func (eff EffectLightning) Apply(g *Game) bool {
	const (
		maxdist = 3
		selfDmg = 1
	)
	pa := g.PlayerActor()
	if pa.HP <= 1 {
		g.Log("你伤势过重，无法使用“闪电”。")
		return false
	}
	mp := &MapPath{passable: func(q gruid.Point) bool {
		if !inMap(q) {
			return false
		}
		i, _ := g.ActorAt(q)
		return i >= 0
	}}
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{g.PP()}, maxdist)
	done := false
	lightning := func() {
		g.md.LightningAnimation(nodes)
		g.Logf("你释放静电（-%d 点生命值）。", selfDmg)
		g.MakeNoise(g.PP(), NoiseLightning)
		g.InflictDamage(PlayerID, pa, selfDmg, AttackOther)
		done = true
	}
	for _, n := range nodes {
		i, ai := g.ActorAt(n.P)
		if i < 0 || i == PlayerID {
			continue
		}
		if !done {
			lightning()
		}
		dmg := maxdist + 1 - n.Cost
		if g.InflictDamage(i, ai, dmg, AttackLightning) > 0 {
			g.PutStatus(i, ai, StatusDaze, DurationDazeLightning)
		}
	}
	if !done && pa.Has(StatusDaze) && pa.DoesAny(ResistanceDaze) {
		lightning()
	}
	if !done {
		g.Log("没有合适的闪电目标。")
	}
	return done
}

func (eff EffectLightning) Name() string {
	return "闪电"
}

func (eff EffectLightning) Desc(_ *Game) string {
	return fmt.Sprintf("对所有相邻的一串怪物造成 1–3 点伤害，伤害随距离递减。受影响的怪物会@O眩晕@N %d 回合。你也会受到 1 点伤害。",
		DurationDazeLightning)
}

type EffectBark struct{}

func (eff EffectBark) Apply(g *Game) bool {
	const maxdist = MaxFOVRange
	mp := &MapPath{passable: g.Map.Passable}
	g.PR.BreadthFirstMap(mp, []gruid.Point{g.PP()}, maxdist)
	g.Log("你吠叫。")
	pp := g.PP()
	g.md.BarkAnimation(pp)
	g.MakeNoise(pp, NoiseBark)
	for i, ai := range g.Monsters() {
		ei := g.Entity(i)
		if g.PR.BreadthFirstMapAt(ei.P) <= maxdist {
			g.PutStatus(i, ai, StatusFear, DurationFearBark)
		}
	}
	return true
}

func (eff EffectBark) Name() string {
	return "吠叫"
}

func (eff EffectBark) Desc(_ *Game) string {
	return fmt.Sprintf("恐吓距离你 %d 格内的所有怪物，持续 %d 回合。",
		MaxFOVRange, DurationFearBark)
}

type EffectNoxiousSmell struct{}

func (eff EffectNoxiousSmell) Apply(g *Game) bool {
	const maxdist = MaxFOVRange
	g.Log("你散发出令人混乱的恶臭。")
	g.confusingSmell(PlayerID, maxdist, DurationConfusionSkunk)
	return true
}

func (g *Game) confusingSmell(i ID, maxdist, duration int) {
	p := g.Entity(i).P
	mp := &MapPath{passable: g.Map.Passable}
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{p}, maxdist)
	g.md.FloodAnimation(nodes, ColorGreen)
	for j, aj := range g.Actors() {
		if j == i {
			continue
		}
		ej := g.Entity(j)
		if g.PR.BreadthFirstMapAt(ej.P) <= maxdist {
			g.PutStatus(j, aj, StatusConfusion, duration)
		}
	}
}

func (eff EffectNoxiousSmell) Name() string {
	return "恶臭"
}

func (eff EffectNoxiousSmell) Desc(_ *Game) string {
	return fmt.Sprintf("使距离你 %d 格内的怪物陷入混乱，持续 %d 回合。",
		MaxFOVRange, DurationConfusionSkunk)
}

type EffectLignify struct{}

func (eff EffectLignify) Apply(g *Game) bool {
	done := false
	for i, ai := range g.Monsters() {
		ei := g.Entity(i)
		if !g.InFOV(ei.P) {
			continue
		}
		if ai.DoesAny(ImmunityLignification) || ai.Has(StatusLignification) {
			// Double-check before so that the "Growing roots"
			// message appears first and only if the status will
			// apply.
			continue
		}
		if !done {
			g.Log("生长的根须使敌人木化。")
			done = true
		}
		g.PutStatus(i, ai, StatusLignification, DurationLignificationRoots)
	}
	if done {
		mp := &MapPath{passable: inMap}
		nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{g.PP()}, MaxFOVRange)
		g.md.FloodAnimation(nodes, ColorYellow)
	} else {
		g.Log("视野内没有可木化的敌人。")
	}
	return done
}

func (eff EffectLignify) Name() string {
	return "木化"
}

func (eff EffectLignify) Desc(_ *Game) string {
	return fmt.Sprintf("使视野内的怪物木化 %d 回合。",
		DurationLignificationRoots)
}

type EffectPoisonCloud struct{}

func (eff EffectPoisonCloud) Apply(g *Game) bool {
	p := g.PP()
	behind := p.Sub(g.Dir)
	done := false
	for q := range g.Map.PassableNeighbors(p) {
		if q == behind {
			continue
		}
		nq := q.Add(q.Sub(p))
		if g.IsFree(q) && g.IsFree(nq) {
			q = nq
		}
		done = true
		g.PoisonCloudAt(q)
	}
	if done {
		g.Log("你向周围喷出毒雾。")
	} else {
		g.Log("没有足够空间喷吐毒雾。")
	}
	return done
}

func (eff EffectPoisonCloud) Name() string {
	return "毒云"
}

func (eff EffectPoisonCloud) Desc(_ *Game) string {
	return "向前方及两侧最多两格外的空位喷出毒云，持续 7–11 回合。"
}

type EffectSprint struct{}

func (eff EffectSprint) Apply(g *Game) bool {
	pa := g.PlayerActor()
	if pa.Has(StatusDaze) {
		g.Log("眩晕时使用“冲刺”毫无作用。")
		return false
	}
	if pa.Has(StatusSprint) {
		g.Log("你已处于冲刺状态，无法再次使用“冲刺”。")
		return false
	}
	if !pa.CanMove() {
		g.Log("木化时无法使用“冲刺”。")
		return false
	}
	g.PutStatus(PlayerID, pa, StatusSprint, DurationSprint)
	g.instant = true
	return true
}

func (eff EffectSprint) Name() string {
	return "冲刺"
}

func (eff EffectSprint) Desc(_ *Game) string {
	return fmt.Sprintf("进入@B冲刺@N状态，持续 %d 回合。%s @G即时生效@N。",
		DurationSprint, statusDesc[StatusSprint])
}

type EffectFireRetreat struct{}

func (eff EffectFireRetreat) Apply(g *Game) bool {
	pl, pa := g.Player()
	if !pa.CanMove() {
		g.Log("木化时无法撤退。")
		return false
	}
	if pa.HP == 1 && pa.Has(StatusPoison) {
		g.Log("中毒且濒死时无法撤退。")
		return false
	}
	from := pl.P
	p, to := from, from
	for range 2 {
		p = p.Sub(g.Dir)
		if !g.InFOV(p) {
			break
		}
		if !g.Map.Passable(p) && !g.Dig(PlayerID, pa, p) {
			break
		}
		if i, _ := g.ActorAt(p); i >= 0 {
			break
		}
		to = p
	}
	if from == to {
		g.Log("没有空间撤退。")
		return false
	}
	g.Log("你向后撤退。")
	g.MoveActor(PlayerID, pa, to, MovJump)
	g.FireCloudAt(from)
	// We used MovJump for showing the fire cloud before triggering any
	// traps.
	g.TriggerTrap(PlayerID, pa)
	return true
}

func (eff EffectFireRetreat) Name() string {
	return "烈焰撤退"
}

func (eff EffectFireRetreat) Desc(_ *Game) string {
	return "向后移动最多两格，并留下持续 6–12 回合的火云。"
}

type EffectShadows struct{}

func (eff EffectShadows) Apply(g *Game) bool {
	pa := g.PlayerActor()
	g.PutStatus(PlayerID, pa, StatusShadow, DurationShadow+pa.Statuses[StatusShadow])
	g.instant = true
	return true
}

func (eff EffectShadows) Name() string {
	return "暗影"
}

func (eff EffectShadows) Desc(_ *Game) string {
	return fmt.Sprintf("@B暗影@N状态还会持续 %d 回合。%s @G即时生效@N。",
		DurationShadow, statusDesc[StatusShadow])
}

type EffectSnack struct{}

func (eff EffectSnack) Apply(g *Game) bool {
	pa := g.PlayerActor()
	if pa.Has(StatusLignification) {
		g.Log("木化时无法觅食。")
		return false
	}
	pp := g.PP()
	var i ID = -1
	var ci Item
	for j, cj := range g.Comestibles() {
		if i < 0 || paths.DistanceManhattan(pp, g.Entity(j).P) < paths.DistanceManhattan(pp, g.Entity(i).P) {
			i = j
			ci = cj
		}
	}
	if i < 0 {
		g.Log("地图上已没有食物。")
		return false
	}
	g.SenseEntity(i, "感知到")
	g.Log("你传送去快速觅食！")
	to := g.Entity(i).P
	g.MoveActor(PlayerID, pa, to, MovTeleport)
	g.UpdateFOV()
	g.UpdateKnowledge()
	g.snack = true
	defer func() { g.snack = false }()
	return ci.Use(g, i)
}

func (eff EffectSnack) Name() string {
	return "觅食"
}

func (eff EffectSnack) Desc(g *Game) string {
	desc := "若附近有食物，你会传送到最近的一份并将其吃掉。"
	if GluttonyRework {
		return desc
	}
	return desc + "不会触发@B贪食@N。"
}

type EffectGarden struct{}

func (eff EffectGarden) Apply(g *Game) bool {
	pa := g.PlayerActor()
	if pa.Has(StatusGardener) {
		g.Log("你已处于园艺师状态，无法再次使用“园艺”。")
		return false
	}
	g.PutStatus(PlayerID, pa, StatusGardener, DurationGarden)
	g.instant = true
	return true
}

func (eff EffectGarden) Name() string {
	return "园艺"
}

func (eff EffectGarden) Desc(_ *Game) string {
	return fmt.Sprintf("进入@B园艺师@N状态，持续 %d 回合。%s @G即时生效@N。",
		DurationGarden, statusDesc[StatusGardener])
}

type EffectStomp struct{}

func (eff EffectStomp) Apply(g *Game) bool {
	pp := g.PP()
	mp := &MapPath{passable: inMap}
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{pp}, 2)
	g.Log("你重踏地面。")
	g.MakeNoise(pp, NoiseStomp)
	for _, n := range nodes {
		g.stompDigAt(n.P)
	}
	pa := g.PlayerActor()
	d := DurationBerserkStomp
	if pa.Has(StatusClarity) {
		// Using stomp while clear-minded works but reduces the
		// duration by half (rounding up).
		g.RemoveStatus(PlayerID, pa, StatusClarity)
		d = (d + 1) / 2
	}
	if pa.Has(StatusBerserk) {
		// Stomp renews berserk duration (if lower) and temporary HP bonus.
		pa.Statuses[StatusBerserk] = max(pa.Statuses[StatusBerserk], d+1)
		bonusHP := pa.HP - max(1, pa.HP-pa.HPBonus())
		g.AdjustHP(PlayerID, pa, pa.HPBonus()-bonusHP)
	} else {
		g.PutStatus1(PlayerID, pa, StatusBerserk, d)
	}
	return true
}

// stompDig is similar to dig but with player-specific code only and without
// the noise.
func (g *Game) stompDigAt(at gruid.Point) {
	t := g.Map.Terrain.At(at)
	if Passable(t) {
		return
	}
	g.BreakWall(at)
	if t == TranslucentWall {
		g.PoisonCloudAt(at)
	}
	// XXX: generate dust like for the earth menhir?
}

func (eff EffectStomp) Name() string {
	return "重踏"
}

func (eff EffectStomp) Desc(_ *Game) string {
	return fmt.Sprintf("摧毁半径两格内的墙壁，并使你进入@B狂暴@N状态 %d 回合。",
		DurationBerserkStomp)
}

type EffectDeathStare struct{}

func (eff EffectDeathStare) Apply(g *Game) bool {
	pl, pa := g.Player()
	if pa.Has(StatusFear) {
		g.Log("恐惧时无法使用“死亡凝视”。")
		return false
	}
	if j, aj := g.RangedTargetInDir(pl.P, g.Dir); j >= 0 {
		ej := g.Entity(j)
		g.Logf("你向%s投去致命的凝视。", ej.Name)
		g.md.RangedBeamAnimation(pl.P, ej.P, ColorViolet)
		g.InflictDamage(j, aj, aj.HP, AttackOther)
		g.PutStatus1(PlayerID, pa, StatusFear, DurationFearDeathStare)
		return true
	}
	g.Log("前方没有可见敌人，无法使用“死亡凝视”。")
	return false
}

// DeathStareTarget returns the first visible actor target in the given
// direction. Works even through translucent walls.
func (g *Game) DeathStareTarget(from, dir gruid.Point) (ID, *Actor) {
	for p := from.Add(dir); g.InFOV(p); p = p.Add(dir) {
		if i, ai := g.ActorAt(p); i >= 0 {
			return i, ai
		}
	}
	return -1, nil
}

func (eff EffectDeathStare) Name() string {
	return "死亡凝视"
}

func (eff EffectDeathStare) Desc(_ *Game) string {
	return fmt.Sprintf("杀死前方第一个可见敌人。作为反噬，你会陷入@O恐惧@N %d 回合。",
		DurationFearDeathStare)
}

type EffectLayRune struct{}

func (eff EffectLayRune) Apply(g *Game) bool {
	pp := g.PP()
	if g.Map.Terrain.At(pp) != Floor {
		g.Log("无法在非地面格上布置符文。")
		return false
	}
	if i, _ := g.ItemAt(pp); i >= 0 {
		g.Log("无法在被占据的格子上布置符文。")
		return false
	}
	r := MagicRune(g.IntN(NRunes))
	_, e := g.genRunicTrapAt(r, pp)
	e.Seen = true
	g.Logf("你布置了%s。", r)
	g.MakeNoise(pp, NoiseCackle)
	return true
}

func (eff EffectLayRune) Name() string {
	return "布置符文"
}

func (eff EffectLayRune) Desc(_ *Game) string {
	return "在当前地面格上布置随机符文陷阱。"
}

type EffectDisorient struct{}

func (eff EffectDisorient) Apply(g *Game) bool {
	pa := g.PlayerActor()
	g.PutStatus(PlayerID, pa, StatusDisorient, DurationDisorient+pa.Statuses[StatusDisorient])
	g.instant = true
	return true
}

func (eff EffectDisorient) Name() string {
	return "迷向"
}

func (eff EffectDisorient) Desc(_ *Game) string {
	return fmt.Sprintf("进入@B迷向@N状态，持续 %d 回合。%s @G即时生效@N。", DurationDisorient, statusDesc[StatusDisorient])
}

func (g *Game) menhirMsg(name string) {
	g.Logf("你激活了%s。", name)
}

type EffectEarthMenhir struct{}

func (eff EffectEarthMenhir) Apply(g *Game) bool {
	const maxdist = MaxFOVRange + 4 // "earthquake" distance
	g.menhirMsg("土之石碑")
	pp := g.PP()
	g.MakeNoise(pp, NoiseEarthMenhir)
	mp := &MappingPath{passable: g.Map.NoWallAt}
	// Destroy nearby walls.
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{pp}, maxdist)
	draw, cost := false, 1
	for _, n := range nodes {
		if n.Cost > cost && draw {
			g.md.AnimationFrameFast()
			draw, cost = false, n.Cost
		}
		t := g.Map.Terrain.At(n.P)
		if t != Wall || n.Cost > maxdist || !g.Map.PassableNeighbor(n.P) {
			continue
		}
		g.earthMenhirDig(n.P)
		if g.Map.KnownTerrain.At(n.P) == Wall {
			g.Map.KnownTerrain.Set(n.P, Rubble)
		}
		draw = true
	}
	g.UpdateFOV()
	g.md.AnimationFrameFast()
	// Reveal walls connected to interior walls (ignoring translucent
	// ones).
	draw, cost = false, 1
	nodes = g.PR.BreadthFirstMap(mp, []gruid.Point{pp}, unreachable)
	for _, n := range nodes {
		if n.Cost > cost && draw {
			g.md.AnimationFrameFast()
			draw, cost = false, n.Cost
		}
		t := g.Map.Terrain.At(n.P)
		if t != Wall {
			continue
		}
		if g.Map.KnownTerrain.At(n.P) != Wall && g.Map.hasInteriorNeighbor(n.P) {
			g.Map.KnownTerrain.Set(n.P, Wall)
			draw = true
		}
	}
	return true
}

// hasInteriorNeighbor reports whether there's a earth-menhir revelable wall at the
// given position (wall adjacent to an interior wall).
func (m *Map) hasInteriorNeighbor(p gruid.Point) bool {
	for _, dir := range Directions {
		if m.interiorWallAt(p.Add(dir)) {
			return true
		}
	}
	return false
}

// interiorWallAt reports whether there's an interior wall at the given
// position (plain wall not connected to any non-wall tile).
func (m *Map) interiorWallAt(p gruid.Point) bool {
	if m.Terrain.At(p) != Wall {
		return false
	}
	for q := range Neighbors(p) {
		if m.Terrain.At(q) != Wall {
			return false
		}
	}
	return true
}

func (g *Game) BreakWall(at gruid.Point) {
	g.Stats.Digs++
	TotemEvents(g, &EventHappened{EvType: EventWallBreak})
	g.Map.Terrain.Set(at, Rubble)
}

// earthMenhirDig is similar to dig but without the noise and extra behaviors
// (because we want a single noise source at the "earthquake" centre).
func (g *Game) earthMenhirDig(at gruid.Point) {
	g.BreakWall(at)
	for q := range Neighbors(at) {
		switch g.Map.Terrain.At(q) {
		case Floor, Foliage:
			g.NormalCloudAt(q, 2+g.IntN(3))
		}
		// We only put a cloud on floor and foliage tiles but not
		// rubble ones, for clarity, as it doesn't change visibility,
		// so it's almost the same (except for the very rare case of a
		// poison cloud on rubble that would be replaced by a normal
		// one).
	}
}

func (eff EffectEarthMenhir) Desc(_ *Game) string {
	return "发出响亮的回声，震碎附近的墙壁，并揭示与内部墙壁相邻的地图边界。"
}

type EffectWarpingMenhir struct{}

func (eff EffectWarpingMenhir) Apply(g *Game) bool {
	g.menhirMsg("传送石碑")
	// Teleport nearby monsters away.
	mp := &MapPath{passable: g.Map.NoWallAt}
	pp := g.PP()
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{pp}, MaxFOVRange)
	for _, n := range nodes {
		if n.P == pp {
			continue
		}
		if i, ai := g.ActorAt(n.P); i > 0 {
			g.TeleportActor(i, ai, 0)
		}
	}
	// Reveal portal or orb of corruption.
	for i, ei := range g.NPMapEntities() {
		switch ei.Role.(type) {
		case *Portal:
			g.SenseEntity(i, "感知到")
		case *CorruptionOrb:
			g.SenseEntity(i, "感知到")
		}
	}
	if g.IntN(2+(MapLevels-g.Map.Level)/2) == 0 {
		// Dungeon sends a warping wraith to investigate: it first goes
		// to the menhir, then back to the portal.
		q := g.Map.Orb
		if q == InvalidPos {
			q = g.Map.Portal
		}
		_, a := g.genMonsterGuardian(WarpingWraith, q)
		a.Behavior.Target = pp
	}
	return true
}

func (eff EffectWarpingMenhir) Desc(_ *Game) string {
	return fmt.Sprintf("将距离你 %d 格内的怪物传送走，并揭示所有传送门的位置。地城可能会察觉到你并派遣幽灵调查；你所在层数越深，发生概率越高。", MaxFOVRange)
}

type EffectPoisonMenhir struct{}

func (eff EffectPoisonMenhir) Apply(g *Game) bool {
	g.menhirMsg("毒之石碑")
	// Spawn poison clouds randomly.
	mp := &MapPath{passable: g.Map.NoWallAt}
	pp := g.PP()
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{pp}, MaxFOVRange)
	g.rand.Shuffle(len(nodes), func(i, j int) {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	})
	for _, n := range nodes[:min(max(3, len(nodes)/4), len(nodes))] {
		if n.P == pp {
			continue
		}
		g.PoisonCloudAt(n.P)
	}
	// Reveal translucent walls (which contain poison gas).
	for p, t := range g.Map.Terrain.All() {
		if t == TranslucentWall {
			g.Map.KnownTerrain.Set(p, t)
		} else if g.Map.KnownTerrain.At(p) == TranslucentWall {
			g.Map.KnownTerrain.Set(p, t)
		}
	}
	return true
}

func (eff EffectPoisonMenhir) Desc(_ *Game) string {
	return fmt.Sprintf("在距离你 %d 格内生成大量毒云，并揭示半透明墙壁。", MaxFOVRange)
}

type EffectFireMenhir struct{}

func (eff EffectFireMenhir) Apply(g *Game) bool {
	g.menhirMsg("火之石碑")
	// Spawn fire clouds randomly.
	mp := &MapPath{passable: g.Map.NoWallAt}
	pp := g.PP()
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{pp}, MaxFOVRange)
	g.rand.Shuffle(len(nodes), func(i, j int) {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	})
	for _, n := range nodes[:min(max(3, len(nodes)/3), len(nodes))] {
		if n.P == pp {
			continue
		}
		g.FireCloudAt(n.P)
	}
	// Reveal foliage.
	// Extra FOV update first to avoid info leak during animation. Just FOV
	// is enough, because it becomes strictly smaller.
	g.UpdateFOV()
	g.revealFoliage()
	return true
}

func (g *Game) revealFoliage() {
	mp := &MapPath{passable: g.Map.Passable}
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{g.PP()}, unreachable)
	var draw bool
	cost := 1
	for _, n := range nodes {
		if n.Cost > cost && draw {
			g.md.AnimationFrameFast()
			draw = false
			cost = n.Cost
		}
		t := g.Map.Terrain.At(n.P)
		if t != Foliage {
			continue
		}
		if g.Map.KnownTerrain.At(n.P) != t && g.IntN(5) < 2 {
			g.Map.KnownTerrain.Set(n.P, t)
			draw = true
		}
	}
}

func (eff EffectFireMenhir) Desc(_ *Game) string {
	return fmt.Sprintf("在距离你 %d 格内生成大量火云，每团持续 6–12 回合，并部分揭示植被。", MaxFOVRange)
}
