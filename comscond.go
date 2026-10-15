// This file implements conditional comestible effects (with or without
// non-conditional variant).

package main

import (
	"fmt"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
)

// Conditional comestible effect, only usable if conditions are met.
type ConditionalComestible interface {
	CanApply(*Game, ID) bool
}

type EffectPolymorphFungus struct{}

func (eff EffectPolymorphFungus) CanApply(g *Game, i ID) bool {
	var n int
	for i := range g.Monsters() {
		ei := g.Entity(i)
		if g.InFOV(ei.P) {
			n++
		}
	}
	return n == 2
}

func (eff EffectPolymorphFungus) Apply(g *Game) bool {
	g.eatMsg(PolymorphFungus.Name(), "")
	// Collect IDs from monsters in view.
	var mids []ID
	for i := range g.Monsters() {
		if g.InFOV(g.Entity(i).P) {
			mids = append(mids, i)
		}
	}
	if len(mids) == 0 {
		g.PutStatus1(PlayerID, g.PlayerActor(), StatusVampirism, DurationVampFung)
		return true
	}
	mks := []ActorKind{BlinkButterfly, ConfusingEye}
	p := g.rand.Perm(len(mids))[:min(len(mids), 2)]
	g.md.PolymorphAnimation(mids, p)
	for i, pi := range p {
		ei := g.Entity(mids[pi])
		ai := ei.Role.(*Actor)
		mki := mks[i]
		e := monster(mki)
		a := e.Role.(*Actor)
		g.Logf("%s变成了%s。", ei.Name, One(e.Name))
		if ai.DoesAny(MonsNotable) {
			g.StoryLogf("将%s变成了%s", One(ei.Name), One(e.Name))
		}
		// Update name and rune only in entity.
		ei.Name = e.Name
		ei.Rune = e.Rune
		// Update actor stats and traits, but keep behavior, statuses,
		// and knowledge info.
		ai.MaxHP = a.MaxHP
		ai.HP = min(a.HP, ai.MaxHP)
		ai.Defense = a.Defense
		ai.Attack = a.Attack
		ai.Kind = a.Kind
		ai.Traits = a.Traits
		switch mki {
		case BlinkButterfly:
			ai.Statuses[StatusLignification] = 0
			ai.Statuses[StatusImbalance] = 0
		case ConfusingEye:
			ai.Statuses[StatusConfusion] = 0
		}
	}
	g.md.PolymorphAnimation(mids, p)
	return true
}

func (eff EffectPolymorphFungus) Desc(g *Game) string {
	if g.HasCondComestibles() {
		return "仅在视野中恰有两只怪物时才能食用。将它们分别变成闪现蝴蝶和迷惑之眼。"
	}
	return fmt.Sprintf("将视野中的一只随机怪物变成闪现蝴蝶；若还有第二只，则将其变成迷惑之眼。\n若视野中没有怪物，则改为获得 @B吸血@N，持续 %d 回合，为吸血突袭做好准备。", DurationVampFung)
}

type EffectRunethornRose struct{}

func (g *Game) anyRuneAt(p gruid.Point) (*RunicTrap, bool) {
	if i, it := g.ItemAt(p); i >= 0 {
		rt, ok := it.(*RunicTrap)
		return rt, ok
	}
	return nil, false
}

func (eff EffectRunethornRose) CanApply(g *Game, i ID) bool {
	_, ok := g.anyRuneAt(g.PP())
	return ok
}

func (eff EffectRunethornRose) Apply(g *Game) bool {
	g.eatMsg(RunethornRose.Name(), "")
	pp := g.PP()
	var t MagicRune // type of rune
	if r, ok := g.anyRuneAt(pp); ok {
		t = r.Rune
	} else {
		t = MagicRune(g.IntN(NRunes))
	}
	size := g.Map.Terrain.Size()
	for i := -1; i <= 1; i += 2 {
		for j := -1; j <= 1; j += 2 {
			q := pp
			for range 2 {
				q = q.Shift(i, j)
				p := q
				for !g.validRunePlace(p) || p == pp {
					p = gruid.Point{g.rand.IntN(size.X), g.rand.IntN(size.Y)}
				}
				i, _ := g.genRunicTrapAt(t, p)
				g.SenseEntity(i, "感知到")
				// Trigger the trap if there's already a
				// monster on top of it.
				if k, ak := g.ActorAt(p); k >= 0 {
					g.TriggerTrap(k, ak)
				}
			}
		}
	}
	return true
}

func (g *Game) validRunePlace(p gruid.Point) bool {
	m := g.Map
	if m.Terrain.At(p) != Floor || m.RuneCache.At(p) > 0 {
		return false
	}
	i, _ := g.ItemAt(p)
	return i < 0
}

func (eff EffectRunethornRose) Desc(g *Game) string {
	if !g.HasCondComestibles() {
		return "生成 8 个同类符文：在地面上每个对角方向最多放置两个；若空间不足，则在地图上随机放置。\n符文类型随机决定；若你正站在陷阱上，则使用该陷阱的符文类型。"
	}
	return "仅在你站在符文陷阱上时才能食用。生成 8 个同类符文：在地面上每个对角方向最多放置两个；若空间不足，则在地图上随机放置。"
}

type EffectTunnelingAcorn struct{}

func (eff EffectTunnelingAcorn) CanApply(g *Game, i ID) bool {
	pp := g.PP()
	n := pp.Add(g.Dir)
	return inMap(n) && !g.Map.Passable(n) && g.Map.AdjacentNonPassableCount(pp) >= 3
}

func (eff EffectTunnelingAcorn) Apply(g *Game) bool {
	g.eatMsg(TunnelingAcorn.Name(), "")
	p := g.PP().Add(g.Dir)
	from := p
	for g.Map.Passable(p) {
		// Skip over floor tiles for non-conditional variant, and
		// reveal terrain on the path.
		p = p.Add(g.Dir)
		g.SeeTerrain(p)
		if i, _ := g.ItemAt(p); i >= 0 {
			g.SenseEntity(i, "感知到")
		}
	}
	for inMap(p) && !g.Map.Passable(p) {
		g.disintegrateWall(p)
		p = p.Add(g.Dir)
	}
	if from != p {
		g.md.BeamAnimation(from, p)
	}
	return true
}

func (g *Game) disintegrateWall(at gruid.Point) {
	g.Stats.Digs++
	TotemEvents(g, &EventHappened{EvType: EventWallBreak})
	g.Map.Terrain.Set(at, Floor)
	g.Map.KnownTerrain.Set(at, Floor) // reveal tunnel
	g.NormalCloudAt(at, 10+g.IntN(30))
}

func (eff EffectTunnelingAcorn) Desc(g *Game) string {
	if !g.HasCondComestibles() {
		return "发射一道光束，静默摧毁前方遇到的墙壁，直到遇到非墙地形为止，并显露沿途地形。"
	}
	return "仅在你紧邻三面墙且面朝其中一面时才能食用。静默摧毁前方遇到的所有墙壁，直到遇到非墙地形为止。"
}

type EffectVanishingSnail struct{}

func (eff EffectVanishingSnail) CanApply(g *Game, i ID) bool {
	count := 0
	for _, a := range g.Monsters() {
		if a.Behavior.State == Hunting {
			count++
		}
	}
	return count >= 4
}

func (eff EffectVanishingSnail) Apply(g *Game) bool {
	g.eatMsg(VanishingSnail.Name(), "")
	for i, ai := range g.Monsters() {
		ei := g.Entity(i)
		if ai.Behavior.State == Hunting {
			// Make monster lose track of you and wander again to a
			// nearby location, as if hit by a player with Shadow.
			ai.Behavior.State = Wandering
			ai.Behavior.Target = g.RandomPassableWithin(ei.P, MaxFOVRange)
		}
	}
	pa := g.PlayerActor()
	for p := range g.Map.PassableNeighbors(g.PP()) {
		if k, ak := g.ActorAt(p); k >= 0 && !ak.ResistsMove() {
			ek := g.Entity(k)
			if to, ok := g.BlinkPos(ek.P); ok {
				g.MoveActor(k, ak, to, MovTeleport)
			}
		}
	}
	g.PutStatus1(PlayerID, pa, StatusShadow, DurationShadow)
	return true
}

func (eff EffectVanishingSnail) Desc(g *Game) string {
	cond := "仅在至少有四只怪物正在追猎你时才能食用，包括视野外的怪物。"
	if !g.HasCondComestibles() {
		cond = ""
	}
	return cond + fmt.Sprintf(
		"使所有怪物失去对你的追踪，并将身旁的怪物闪现带走。获得 @B暗影@N，持续 %d 回合。",
		DurationShadow)
}

type EffectWarpingBean struct{}

func (eff EffectWarpingBean) CanApply(g *Game, i ID) bool {
	s := g.PlayerActor().Statuses
	return s.Has(StatusLignification) || s.Has(StatusDaze) || s.Has(StatusPoison)
}

func (eff EffectWarpingBean) Apply(g *Game) bool {
	if GluttonyRework {
		// In case the bear teleported or warped first.
		g.UpdateFOV()
		g.UpdateKnowledge()
	}
	pa := g.PlayerActor()
	if pa.Has(StatusBerserk) {
		pa.HP = max(1, pa.HP-pa.HPBonus())
	}
	if pa.Has(StatusLignification) {
		pa.HP = max(min(3, pa.HP), pa.HP-pa.HPBonus())
	}
	for i := range pa.Statuses {
		pa.Statuses[i] = 0
	}
	var hpstr string
	if !g.Mod(ModHealingCombat) {
		hpstr = newHPMsg(pa, HealWarpingBean)
	}
	g.eatMsg(WarpingBean.Name(), hpstr)
	if !g.Mod(ModHealingCombat) {
		g.healPlayer(pa, HealFirebreathPepper)
	}
	pp := g.PP()
	ps := g.BlinkPosSlice(pp)
	to := pp
	for range 6 {
		p := ps[g.IntN(len(ps))]
		if paths.DistanceManhattan(to, pp) < paths.DistanceManhattan(p, pp) {
			to = p
		}
	}
	if to != pp {
		g.MoveActor(PlayerID, pa, to, MovTeleport)
	}
	return true
}

func (eff EffectWarpingBean) Desc(g *Game) string {
	cond := "仅在你处于木化、眩晕或中毒状态时才能食用。"
	if !g.HasCondComestibles() {
		cond = ""
	}
	return fmt.Sprintf("%s清除你的所有状态，且不会产生通常的后遗效果；随后将你传送到视野内的随机位置，远处格子的概率更高。%s",
		cond, g.healMsg(HealWarpingBean))
}

type EffectFortressOyster struct{}

func (eff EffectFortressOyster) CanApply(g *Game, i ID) bool {
	count := 0
	for range g.AdjacentActors(g.PP()) {
		count++
	}
	return count >= 3
}

func (eff EffectFortressOyster) Apply(g *Game) bool {
	const d = 5 // common status duration
	pa := g.PlayerActor()
	g.eatMsg(FortressOyster.Name(), "")
	g.RemoveStatus(PlayerID, pa, StatusClarity)
	if pa.Has(StatusBerserk) {
		pa.Statuses[StatusBerserk] = max(pa.Statuses[StatusBerserk], d+1)
		bonusHP := pa.HP - max(1, pa.HP-pa.HPBonus())
		g.AdjustHP(PlayerID, pa, pa.HPBonus()-bonusHP)
	} else {
		g.PutStatus1(PlayerID, pa, StatusBerserk, d)
	}
	g.RemoveStatus(PlayerID, pa, StatusFoggySkin)
	if pa.Has(StatusLignification) {
		// We handle lignification after berserk, because at low HP it
		// can result in higher final bonuses due to the HP floor of 3.
		pa.Statuses[StatusLignification] = max(pa.Statuses[StatusLignification], d+1)
		bonusHP := pa.HP - max(min(3, pa.HP), pa.HP-pa.HPBonus())
		g.AdjustHP(PlayerID, pa, pa.HPBonus()-bonusHP)
	} else {
		g.PutStatus1(PlayerID, pa, StatusLignification, d)
	}
	g.PutStatus1(PlayerID, pa, StatusVampirism, d)
	return true
}

func (eff EffectFortressOyster) Desc(g *Game) string {
	cond := "仅在你至少紧邻三只怪物时才能食用。"
	if !g.HasCondComestibles() {
		cond = ""
	}
	return cond + "获得 @B狂暴@N、@B木化@N和@B吸血@N，持续 5 回合，并移除不兼容的状态；如适用，还会刷新持续时间和临时生命值加成。"
}

type EffectSeeingPotato struct{}

func (eff EffectSeeingPotato) CanApply(g *Game, i ID) bool {
	j, _ := g.ItemAt(g.PP())
	return i >= 0 && i == j
}

func (eff EffectSeeingPotato) Apply(g *Game) bool {
	g.eatMsg("看见了马铃薯的茎", "")
	const maxdist = 2 * MaxFOVRange
	mp := &MappingPath{passable: func(p gruid.Point) bool { return inMap(p) }}
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{g.PP()}, maxdist)
	var draw bool
	cost := 1
	for _, n := range nodes {
		if n.Cost > cost && draw {
			g.md.AnimationFrameFast()
			draw = false
			cost = n.Cost
		}
		t := g.Map.Terrain.At(n.P)
		if t != Floor {
			continue
		}
		if g.Map.KnownTerrain.At(n.P) != t {
			g.Map.KnownTerrain.Set(n.P, t)
			draw = true
		}
	}
	pp := g.PP()
	for i, ei := range g.NPMapEntities() {
		p := ei.P
		if paths.DistanceManhattan(pp, p) <= maxdist && g.Map.Terrain.At(p) == Floor {
			g.SenseEntity(i, "感知到")
		}
	}
	return true
}

func (eff EffectSeeingPotato) Desc(g *Game) string {
	cond := "只能从地上拾取食用，不能从物品栏中食用。"
	if !g.HasCondComestibles() {
		cond = ""
	}
	return cond + "感知周围地面格及其上的所有实体（范围为视野的 2 倍）。"
}

type EffectTrickyChestnut struct{}

func (eff EffectTrickyChestnut) CanApply(g *Game, i ID) bool {
	if i, it := g.ItemAt(g.PP()); i >= 0 {
		switch o := it.(type) {
		case *Portal:
			return o.Used
		case *Spirit:
			return true
		}
	}
	return false
}

func (eff EffectTrickyChestnut) Apply(g *Game) bool {
	g.eatMsg(TrickyChestnut.Name(), "")
	pp := g.PP()
	var tot *Spirit
	if i, it := g.ItemAt(pp); i >= 0 {
		switch o := it.(type) {
		case *Portal:
			o.Fake = false
			o.Used = false
			g.LogStyled("传送门修好了！", logSpecial)
			return true
		case *Spirit:
			tot = o
		}
	}
	// Recharge spirits by default.
	g.LogStyled("灵体充能已恢复！", logSpecial)
	if tot != nil { // turn totem to empty if present
		g.Entity(g.TotemID()).P = InvalidPos
		et := g.Entity(g.genEmptyTotemAt(pp))
		et.Seen = true
		for _, sp := range g.PlayerSpirits() {
			if g.Mod(ModNoRecharges) {
				sp.Charges = min(sp.Charges+2, sp.GetMaxCharges())
			} else {
				sp.Charges = sp.GetMaxCharges()
			}
		}
	} else {
		// Bear outside of conditions.
		for _, sp := range g.PlayerSpirits() {
			sp.Charges = min(sp.Charges+1, sp.GetMaxCharges())
		}
	}
	return true
}

func (eff EffectTrickyChestnut) Desc(g *Game) string {
	nocond := !g.HasCondComestibles()
	recharges := "站在已充能的图腾上时食用，可清空图腾并恢复所有灵体的充能。"
	if g.Mod(ModNoRecharges) {
		recharges = "站在已充能的图腾上时食用，可清空图腾并为每个灵体恢复两点充能。"
		if nocond {
			recharges = "若你站在已充能的图腾上，则清空图腾并为每个灵体恢复两点充能。"
		}
	} else if nocond {
		recharges = "若你站在已充能的图腾上，则清空图腾并恢复所有灵体的充能。"
	}
	if nocond {
		return fmt.Sprintf("若你站在故障传送门上，可使其恢复运作。\n%s\n否则，每个灵体恢复一点充能。",
			recharges)
	}
	return "站在故障传送门上时食用，可使其恢复运作。\n" + recharges
}

// CantEatMsg returns a message suitable for when you try to eat a conditional
// comestible when the necessary conditions aren't fulfilled.
func CantEatMsg(ef Effect) string {
	switch ef.(type) {
	case EffectVanishingSnail, EffectWarpingBean:
		return "你拿不起来！"
	case EffectFortressOyster, EffectTunnelingAcorn, EffectTrickyChestnut:
		return "它打不开！"
	case EffectRunethornRose:
		return "你无法穿过它的荆棘！"
	case EffectSeeingPotato:
		return "必须先把它种在地上！"
	case EffectPolymorphFungus:
		return "仅当视野中恰有两只怪物时才能食用！"
	default:
		panic("意外的不可食用食物")
	}
}

// HasCondComestibles reports whether conditional comestibles are enabled.
func (g *Game) HasCondComestibles() bool {
	return g.Mod(ModTotemConditions) && !g.PlayerActor().DoesAny(Gluttony)
}
