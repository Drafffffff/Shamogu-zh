package main

import (
	"fmt"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
)

// The following constants are used for various effects and statuses.
const (
	DurationBerserk                = 8
	DurationBerserkTotem           = 5
	DurationConfusion              = 7
	DurationConfusionAmbrosia      = 7
	DurationConfusionToxins        = 5
	DurationDazeJump               = 5
	DurationDazeLightning          = 3
	DurationDazePenalty            = 7
	DurationDazeSpines             = 3
	DurationFearBark               = 4
	DurationMonsFear               = 3
	DurationFire                   = 3
	DurationFoggySkin              = 12
	DurationImbalanceGale          = 5
	DurationImbalanceLignification = 6
	DurationImbalanceSprint        = 5
	DurationLignification          = 12
	DurationNauseaRat              = 5
	DurationNauseaTeleport         = 7
	DurationPoison                 = 4
	DurationPoisonBerserk          = 6

	HealAmbrosia         = 8
	HealClarityLeaves    = 4
	HealFirebreathPepper = 2
	HealFoggySkinOnion   = 2

	DurationDig      = 6
	DurationFocus    = 3
	DurationSprint   = 7
	DurationTimeStop = 4
)

func newHP(a *Actor, hp int) int {
	return max(0, min(a.GetMaxHP(), a.HP+hp))
}

func (g *Game) eatMsg(name string, hp int) {
	g.Logf("You eat the %s.", name)
	pa := g.PlayerActor()
	var nausea string
	if pa.Has(StatusNausea) {
		nausea = "while nauseous "
	}
	if hp != 0 {
		g.StoryLogf("Ate %s %s(HP: %d/%d)", One(name), nausea, newHP(pa, hp), pa.GetMaxHP())
	} else {
		g.StoryLogf("Ate %s %s", One(name), nausea)
	}
}

type EffectAmbrosiaBerries struct{}

func (eff EffectAmbrosiaBerries) Apply(g *Game) bool {
	pa := g.PlayerActor()
	if pa.HP == pa.GetMaxHP() && pa.Has(StatusConfusion) {
		g.Log("You’re already confused and at full health.")
		return false
	}
	g.eatMsg("ambrosia berries", HealAmbrosia)
	g.AdjustHP(PlayerID, pa, HealAmbrosia)
	g.PutStatus(PlayerID, pa, StatusConfusion, DurationConfusionAmbrosia)
	return true
}

func (eff EffectAmbrosiaBerries) Desc() string {
	return fmt.Sprintf("Restores %+d HP, but @Oconfuses@N you for %d turns.",
		HealAmbrosia, DurationConfusionAmbrosia)
}

type EffectBerserkingFlower struct{}

func (eff EffectBerserkingFlower) Apply(g *Game) bool {
	a := g.PlayerActor()
	if a.Has(StatusBerserk) {
		g.Log("You’re already berserk.")
		return false
	}
	g.eatMsg("berserking flower", 0)
	g.PutStatus(PlayerID, a, StatusBerserk, DurationBerserk)
	return true
}

func (eff EffectBerserkingFlower) Desc() string {
	return fmt.Sprintf("Makes you @BBerserk@N for %d turns. %s",
		DurationBerserk, statusDesc[StatusBerserk])
}

type EffectLignificationFruit struct{}

func (eff EffectLignificationFruit) Apply(g *Game) bool {
	a := g.PlayerActor()
	if a.Has(StatusLignification) {
		g.Log("You’re already lignified.")
		return false
	}
	if a.Has(StatusFoggySkin) {
		g.Log("You cannot lignify yourself while smoky.")
		return false
	}
	g.eatMsg("lignification fruit", 0)
	g.PutStatus(PlayerID, a, StatusLignification, DurationLignification)
	return true
}

func (eff EffectLignificationFruit) Desc() string {
	return fmt.Sprintf("@BLignifies@N you for %d turns. %s",
		DurationLignification, statusDesc[StatusLignification])
}

type EffectClarityLeaves struct{}

func (eff EffectClarityLeaves) Apply(g *Game) bool {
	pa := g.PlayerActor()
	g.eatMsg("clarity leaves", HealClarityLeaves)
	g.AdjustHP(PlayerID, pa, HealClarityLeaves)
	g.RemoveStatus(PlayerID, pa, StatusConfusion)
	pp := g.PP()
	sense := false
	for i := range g.Monsters() {
		ei := g.Entity(i)
		if paths.DistanceManhattan(ei.P, pp) < 3*MaxFOVRange {
			if ei.KnownP != ei.P {
				sense = true
			}
			ei.KnownP = ei.P
		}
	}
	if sense {
		g.Log("You sensed some monsters.")
	}
	return true
}

func (eff EffectClarityLeaves) Desc() string {
	return fmt.Sprintf("Removes @OConfusion@N and makes you sense nearby monsters. Restores %+d HP.",
		HealClarityLeaves)
}

type EffectFoggySkinOnion struct{ Beam bool }

func (eff EffectFoggySkinOnion) Apply(g *Game) bool {
	pa := g.PlayerActor()
	if pa.Has(StatusFoggySkin) {
		g.Log("Your skin is already smoky.")
		return false
	}
	g.eatMsg("foggy-skin onion", HealFoggySkinOnion)
	g.AdjustHP(PlayerID, pa, HealFoggySkinOnion)
	g.PutStatus(PlayerID, pa, StatusFoggySkin, DurationFoggySkin)
	return true
}

func (eff EffectFoggySkinOnion) Desc() string {
	return fmt.Sprintf("Get @BFoggy-Skin@N for %d turns. %s Restores %+d HP.",
		DurationFoggySkin, statusDesc[StatusFoggySkin], FoggySkinDefenseBonus)
}

type EffectFirebreathPepper struct{ Beam bool }

func (eff EffectFirebreathPepper) Apply(g *Game) bool {
	pp := g.PP()
	p := pp.Add(g.Dir)
	pa := g.PlayerActor()
	done := false
	eat := func() {
		done = true
		g.eatMsg("firebreath pepper", HealFirebreathPepper)
		g.MakeNoise(pp, NoiseRoar)
		g.AdjustHP(PlayerID, pa, HealFirebreathPepper)
		g.md.FirebreathAnimation(p, g.Dir)
		pa.Statuses[StatusDaze] = 0
	}
	for i := 0; i < MaxFOVRange; i++ {
		if !done && !g.InFOV(p) {
			eat()
		}
		t := g.Map.Terrain.At(p)
		if !Passable(t) {
			break
		}
		if t == Foliage {
			if !done {
				eat()
			}
			g.FireCloudAt(p)
		}
		if j, aj := g.ActorAt(p); j >= 0 {
			if !done {
				eat()
			}
			g.PutStatus(j, aj, StatusFire, DurationFire)
		}
		p = p.Add(g.Dir)
	}
	if !done {
		if pa.Has(StatusDaze) || pa.HP < pa.GetMaxHP() {
			eat()
		} else {
			g.Log("Breathing fire in the current direction is useless.")
		}
	}
	return done
}

func (eff EffectFirebreathPepper) Desc() string {
	return fmt.Sprintf("Makes you breathe a fire beam in the current direction doing %d damage and inflicting the @OFire@N status. Burns foliage, producing fire clouds lasting 4-12 turns. Removes @ODaze@N. Restores %+d HP.",
		FireDamage, HealFirebreathPepper)
}

type EffectTeleportMushroom struct{}

func (eff EffectTeleportMushroom) Apply(g *Game) bool {
	pa := g.PlayerActor()
	if !g.TeleportActor(PlayerID, pa) {
		g.Log("You cannot teleport while lignified.")
		return false
	}
	g.eatMsg("teleport mushroom", 0)
	g.PutStatus(PlayerID, pa, StatusNausea, DurationNauseaTeleport)
	if i, _ := g.Spirit(); i >= 0 {
		if ei := g.Entity(i); paths.DistanceManhattan(ei.P, g.PP()) < 4*MaxFOVRange {
			g.SenseEntity(ei, "sense")
		}
	}
	return true
}

func (eff EffectTeleportMushroom) Desc() string {
	return "Teleports you away. Gives @ONausea@N but makes you sense any nearby totemic spirit."
}

type EffectFocus struct{}

func (eff EffectFocus) Apply(g *Game) bool {
	a := g.PlayerActor()
	if a.Has(StatusFocus) {
		g.Log("You’re already focusing your heads.")
		return false
	}
	g.PutStatus(PlayerID, a, StatusFocus, DurationFocus)
	g.instant = true
	return true
}

func (eff EffectFocus) Name() string {
	return "focus"
}

func (eff EffectFocus) Desc() string {
	return fmt.Sprintf("@BFocus@N for %d turns. %s @GInstant@N.", DurationFocus, statusDesc[StatusFocus])
}

type EffectDig struct{}

func (eff EffectDig) Apply(g *Game) bool {
	a := g.PlayerActor()
	if a.Has(StatusDig) {
		g.Log("You’re already on a destructive charge.")
		return false
	}
	g.PutStatus(PlayerID, a, StatusDig, DurationDig)
	g.instant = true
	return true
}

func (eff EffectDig) Name() string {
	return "dig"
}

func (eff EffectDig) Desc() string {
	return fmt.Sprintf("@BDig@N for %d turns. %s @GInstant@N.",
		DurationDig, statusDesc[StatusDig])
}

type EffectJump struct{}

func (eff EffectJump) Apply(g *Game) bool {
	pl, pa := g.Player()
	if !pa.CanMove() {
		g.Log("You cannot jump while lignified.")
		return true
	}
	if pa.HP == 1 && pa.Has(StatusPoison) {
		g.Log("You cannot jump while poisoned and almost dead.")
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
		g.Log("There’s no room to jump in that direction.")
		return false
	}
	g.md.TeleportAnimation(from, to, true)
	g.BumpMoveActor(PlayerID, pa, to)
	g.Log("You jump.")
	for p := from.Add(dir); p != to; p = p.Add(dir) {
		if i, ai := g.ActorAt(p); i >= 0 {
			if !pa.Has(StatusFear) && !pa.Has(StatusDaze) {
				g.BumpAttackActor(PlayerID, i, pa, ai, AttackPlain, 1)
			}
			g.PutStatus(i, ai, StatusDaze, DurationDazeJump)
		}
	}
	return true
}

func (eff EffectJump) Name() string {
	return "dazing jump"
}

func (eff EffectJump) Desc() string {
	return fmt.Sprintf("Jump in the current direction to the furthest free position in view. You will attack any monsters on the path with +1 attack and @Odaze@N them for %d turns.",
		DurationDazeJump)
}

type EffectPushingGale struct{}

func (eff EffectPushingGale) Apply(g *Game) bool {
	pp := g.PP()
	g.Log("A strong gale pushes the air away from you.")
	g.MakeNoise(pp, NoiseWind)
	for _, dir := range directions {
		g.pushingGale(pp, dir)
	}
	return true
}

func (eff EffectPushingGale) Name() string {
	return "pushing gale"
}

func (eff EffectPushingGale) Desc() string {
	return fmt.Sprintf("Hurt foes in the four cardinal directions for 1-2 damage, pushing them away and leaving them @Ounbalanced@N for %d turns.",
		DurationImbalanceGale)
}

var directions = []gruid.Point{
	{1, 0},
	{0, -1},
	{-1, 0},
	{0, 1},
}

func (g *Game) pushingGale(from, dir gruid.Point) {
	const maxdist = MaxFOVRange - 1
	at := from
	for n, p := 0, at.Add(dir); n < maxdist && g.Map.Passable(p); p = p.Add(dir) {
		at = p
		n++
	}
	for ; at != from; at = at.Sub(dir) {
		i, ai := g.ActorAt(at)
		if i < 0 {
			continue
		}
		ei := g.Entity(i)
		hp := ai.HP
		dmg := min(hp, 1+g.IntN(2))
		g.InflictDamage(i, ai, dmg, AttackGale)
		if ai.IsDead() {
			continue
		}
		g.MonsterUpdateTarget(i, ai, g.PP()) // alert monster toward wind origin
		g.PutStatus(i, ai, StatusImbalance, DurationImbalanceGale)
		to := at.Add(dir)
		n := 0
		for p := to; g.IsFree(p) && n < 2; p = p.Add(dir) {
			to = p
			n++
		}
		if g.IsFree(to) {
			g.MoveActor(i, ai, to)
			g.md.MoveAnimation(ei.P, to)
			// XXX: gale animation: we could collect positions
			// instead first, so that we can do all directions at
			// once in the animation.
		}
	}
}

type EffectTimeStop struct{}

func (eff EffectTimeStop) Apply(g *Game) bool {
	a := g.PlayerActor()
	if a.Has(StatusTimeStop) {
		g.Log("You already froze time.")
		return false
	}
	g.PutStatus(PlayerID, a, StatusTimeStop, DurationTimeStop)
	g.instant = true
	return true
}

func (eff EffectTimeStop) Name() string {
	return "stop time"
}

func (eff EffectTimeStop) Desc() string {
	return fmt.Sprintf("@BStop Time@N for %d turns. %s @GInstant@N.",
		DurationTimeStop, statusDesc[StatusTimeStop])
}

type EffectLightning struct{}

func (eff EffectLightning) Apply(g *Game) bool {
	const (
		maxdist = 3
		selfDmg = 1
	)
	pa := g.PlayerActor()
	if pa.HP <= 1 {
		g.Log("You’re too injured for using lightning.")
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
	for _, n := range nodes {
		i, ai := g.ActorAt(n.P)
		if i < 0 || i == PlayerID {
			continue
		}
		if !done {
			g.md.LightningAnimation(nodes)
			g.Logf("You emit a static discharge (-%d HP).", selfDmg)
			g.MakeNoise(g.PP(), NoiseLightning)
			g.InflictDamage(PlayerID, pa, selfDmg, AttackOther)
			done = true
		}
		dmg := maxdist + 1 - n.Cost
		if g.InflictDamage(i, ai, dmg, AttackLightning) > 0 {
			g.PutStatus(i, ai, StatusDaze, DurationDazeLightning)
		}
	}
	if !done {
		g.Log("No suitable targets for lightning.")
	}
	return done
}

func (eff EffectLightning) Name() string {
	return "lightning"
}

func (eff EffectLightning) Desc() string {
	return fmt.Sprintf("Inflict 1-3 damage on any chain of monsters adjacent to you. Damage decreases with distance. Affected monsters are @Odazed@N for %d turns. Hurts you too for 1 HP.",
		DurationDazeLightning)
}

type EffectBark struct{}

func (eff EffectBark) Apply(g *Game) bool {
	const maxdist = MaxFOVRange
	mp := &MappingPath{passable: g.Map.Passable}
	g.PR.BreadthFirstMap(mp, []gruid.Point{g.PP()}, maxdist)
	g.Log("You bark.")
	g.MakeNoise(g.PP(), NoiseBark)
	for i, ai := range g.Monsters() {
		ei := g.Entity(i)
		if g.PR.BreadthFirstMapAt(ei.P) <= maxdist {
			g.PutStatus(i, ai, StatusFear, DurationFearBark)
		}
	}
	return true
}

func (eff EffectBark) Name() string {
	return "bark"
}

func (eff EffectBark) Desc() string {
	return fmt.Sprintf("Frighten all monsters in a %d-radius distance for %d turns.",
		MaxFOVRange, DurationFearBark)
}

type EffectNoxiousSmell struct{}

func (eff EffectNoxiousSmell) Apply(g *Game) bool {
	const maxdist = MaxFOVRange
	g.Log("You emit a confusing noxious smell.")
	g.confusingSmell(PlayerID, maxdist, DurationConfusion)
	return true
}

func (g *Game) confusingSmell(i ID, maxdist, duration int) {
	p := g.Entity(i).P
	mp := &MappingPath{passable: g.Map.Passable}
	g.PR.BreadthFirstMap(mp, []gruid.Point{p}, maxdist)
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
	return "noxious smell"
}

func (eff EffectNoxiousSmell) Desc() string {
	return fmt.Sprintf("Confuse monsters in a %d-radius distance for %d turns.",
		MaxFOVRange, DurationConfusion)
}

type EffectLignify struct{}

func (eff EffectLignify) Apply(g *Game) bool {
	done := false
	for i, ai := range g.Monsters() {
		ei := g.Entity(i)
		if !g.InFOV(ei.P) {
			continue
		}
		if g.PutStatus(i, ai, StatusLignification, DurationLignification/2) {
			done = true
		}
	}
	if done {
		g.Log("Growing roots lignified your foes.")
	} else {
		g.Log("No lignifiable foes in view.")
	}
	return done
}

func (eff EffectLignify) Name() string {
	return "lignify"
}

func (eff EffectLignify) Desc() string {
	return fmt.Sprintf("Lignify monsters in view for %d turns.",
		DurationLignification/2)
}

type EffectPoisonCloud struct{}

func (eff EffectPoisonCloud) Apply(g *Game) bool {
	p := g.PP()
	done := false
	for q := range g.Map.PassableNeighbors(p) {
		if q == p.Sub(g.Dir) {
			continue
		}
		done = true
		g.PoisonCloudAt(q)
	}
	if done {
		g.Log("You spit poisonous fumes around.")
	} else {
		g.Log("There’s no room to spit poison.")
	}
	return done
}

func (eff EffectPoisonCloud) Name() string {
	return "poison cloud"
}

func (eff EffectPoisonCloud) Desc() string {
	return "Spit poisonous clouds on the sides and in front of you, lasting 6-10 turns."
}

type EffectSprint struct{}

func (eff EffectSprint) Apply(g *Game) bool {
	a := g.PlayerActor()
	if a.Has(StatusSprint) {
		g.Log("You’re already sprinting.")
		return false
	}
	if !a.CanMove() {
		g.Log("You cannot sprint while lignified.")
		return false
	}
	g.PutStatus(PlayerID, a, StatusSprint, DurationSprint)
	g.instant = true
	return true
}

func (eff EffectSprint) Name() string {
	return "sprint"
}

func (eff EffectSprint) Desc() string {
	return fmt.Sprintf("@BSprint@N for %d turns. %s @GInstant@N. Canceled by waiting.",
		DurationSprint, statusDesc[StatusSprint])
}

func (g *Game) menhirMsg(name string) {
	g.Logf("You activate the %s.", name)
}

type EffectMappingMenhir struct{}

func (eff EffectMappingMenhir) Apply(g *Game) bool {
	const maxdist = MaxFOVRange * 4
	g.menhirMsg("mapping menhir")
	g.MakeNoise(g.PP(), NoiseMapping)
	mp := &MappingPath{passable: func(p gruid.Point) bool { return g.Map.Terrain.At(p) != Wall }}
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{g.PP()}, maxdist)
	var draw bool
	cost := 1
	for _, n := range nodes {
		if n.Cost > cost && draw {
			g.md.AnimationFrame()
			draw = false
			cost = n.Cost
		}
		t := g.Map.Terrain.At(n.P)
		if g.Map.KnownTerrain.At(n.P) != t {
			g.Map.KnownTerrain.Set(n.P, t)
			draw = true
		}
	}
	for i := range g.ItemEntities() {
		ei := g.Entity(i)
		if g.PR.BreadthFirstMapAt(ei.P) <= maxdist {
			g.SenseEntity(ei, "sense")
		}
	}
	return true
}

func (eff EffectMappingMenhir) Desc() string {
	return "Maps the surrounding area by producing a noisy echoing sound."
}

type EffectTeleportMenhir struct{}

func (eff EffectTeleportMenhir) Apply(g *Game) bool {
	g.menhirMsg("teleport menhir")
	mp := &MappingPath{passable: g.Map.Passable}
	pp := g.PP()
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{pp}, MaxFOVRange)
	g.rand.Shuffle(len(nodes), func(i, j int) {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	})
	for _, n := range nodes {
		if n.P == pp {
			continue
		}
		if i, ai := g.ActorAt(n.P); i > 0 {
			g.TeleportActor(i, ai)
		}
	}
	return true
}

func (eff EffectTeleportMenhir) Desc() string {
	return fmt.Sprintf("Teleports away monsters in a %d-radius distance.", MaxFOVRange)
}

type EffectPoisonMenhir struct{}

func (eff EffectPoisonMenhir) Apply(g *Game) bool {
	g.menhirMsg("poison menhir")
	mp := &MappingPath{passable: g.Map.Passable}
	pp := g.PP()
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{pp}, MaxFOVRange)
	g.rand.Shuffle(len(nodes), func(i, j int) {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	})
	for _, n := range nodes[:min(max(3, len(nodes)/3), len(nodes))] {
		if n.P == pp {
			continue
		}
		g.PoisonCloudAt(n.P)
	}
	return true
}

func (eff EffectPoisonMenhir) Desc() string {
	return fmt.Sprintf("Makes many poisonous clouds appear in a %d-radius distance.", MaxFOVRange)
}

type EffectFireMenhir struct{}

func (eff EffectFireMenhir) Apply(g *Game) bool {
	g.menhirMsg("fire menhir")
	mp := &MappingPath{passable: g.Map.Passable}
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
	return true
}

func (eff EffectFireMenhir) Desc() string {
	return fmt.Sprintf("Makes many fire clouds appear in a %d-radius distance, each one lasting 4-12 turns.", MaxFOVRange)
}
