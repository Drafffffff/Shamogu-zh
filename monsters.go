package main

import (
	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
)

// HandleMonsterTurn handles AI for a monster actor with given ID. The ID
// should then correspond to a non-player actor.
func (g *Game) HandleMonsterTurn(i ID, ai *Actor) {
	beh := ai.Behavior
	if beh.SkipTurn {
		beh.SkipTurn = false
		return
	}
	if ai.Has(StatusDaze) {
		return
	}
	ei := g.Entity(i)
	if ai.Has(StatusConfusion) && g.monsterDiscord(i, ai) {
		return
	}
	if g.InFOV(ei.P) && g.MonsterUpdateTarget(i, ai, g.PP()) {
		// A wandering monster spends one turn noticing the player.
		return
	}
	if beh.Target == ei.P || beh.Target == InvalidPos {
		// The monster's target has either been reached or is invalid:
		// we spend a turn chosing a new target.
		from, maxdist := ei.P, unreachable
		if beh.Guard != InvalidPos {
			maxdist = MaxFOVRange - 2
			from = beh.Guard
		} else if beh.State == Hunting {
			maxdist = (3 * MaxFOVRange) / 2
		}
		beh.Target = g.RandomPassableNearby(from, maxdist)
		beh.State = Wandering
		if maxdist > MaxFOVRange && g.IntN(3) == 0 || maxdist <= MaxFOVRange && g.IntN(10) == 0 {
			g.CallToCommonTarget(i, ai)
		}
		return
	}
	if beh.State == Wandering && len(beh.Path) >= 2 && beh.Path[0] == ei.P && beh.Path[len(beh.Path)-1] == ai.Behavior.Target {
		// The monster is wandering and is still on the previously
		// computed path, so we try to continue on the cached path.
		g.MonsterBumpTo(i, ai, beh.Path[1])
		return
	}
	// Monster is hunting or was moved from its old path position,
	// so we recompute the current best path.
	beh.Path = g.MonPath(i, ai, ei.P, beh.Target)
	if beh.State != Hunting {
		// A wandering monster spends a turn updating its path.
		return
	}
	if len(beh.Path) >= 2 {
		g.MonsterBumpTo(i, ai, beh.Path[1])
	}
}

// monsterDiscord makes a confused monster possibly attack a random adjacent
// monster, instead of their normal action. It reports whether such an attack
// was performed.
func (g *Game) monsterDiscord(i ID, ai *Actor) bool {
	if ai.Has(StatusFear) {
		return false
	}
	ei := g.Entity(i)
	var ids []ID
	var actors []*Actor
	for p := range Neighbors(ei.P) {
		if j, aj := g.ActorAt(p); j >= 0 {
			ids = append(ids, j)
			actors = append(actors, aj)
		}
	}
	if len(ids) == 0 {
		return false
	}
	k := g.IntN(len(ids))
	g.BumpAttackActor(i, ids[k], ai, actors[k], AttackPlain, 0)
	return true
}

// MonsterBumpTo handles movement or attack for a monster actor. The position
// should be adjacent to the monster's (the next position in the movement
// path).
func (g *Game) MonsterBumpTo(i ID, ai *Actor, to gruid.Point) {
	if ai.Has(StatusFear) && g.MonsterFlee(i, ai) {
		return
	}
	ei := g.Entity(i)
	pl := g.PlayerEntity()
	if AdjacentPoints(ei.P, pl.P) {
		// Attack player.
		g.BumpAttackActor(i, PlayerID, ai, pl.Actor(), AttackPlain, 0)
		return
	}
	if g.InFOV(ei.P) && g.ActorInRange(ei.P, pl.P) == PlayerID {
		pa := g.PlayerActor()
		switch {
		case ai.DoesAny(PatternRanged|PatternRangedRecoil) ||
			ai.DoesAny(PatternCatch|PatternCatchUnbalance) && !pa.ResistsMove():
			g.md.RangedAttackAnimation(ai, ei.P, pl.P)
			g.BumpAttackActor(i, PlayerID, ai, pa, AttackRanged, 0)
			return
		case ai.DoesAny(PatternSwap|PatternSwapDaze) && ai.CanMove() && !pa.ResistsMove():
			g.md.RangedAttackAnimation(ai, pl.P, pl.P)
			g.BumpAttackActor(i, PlayerID, ai, pa, AttackRanged, 0)
			return
		case ai.DoesAny(PatternRampage) && ai.CanMove():
			dir := toDir(pl.P.Sub(ei.P))
			to := pl.P.Sub(dir)
			from := ei.P
			g.BumpMoveActor(i, ai, to)
			g.md.MoveAnimation(from, to)
			g.BumpAttackActor(i, PlayerID, ai, pa, AttackCharge, 0)
			return
		}
	}
	if !ai.CanMove() {
		return
	}
	j, aj := g.ActorAt(to)
	if j < 0 {
		// Move and then try charge.
		delta := to.Sub(ei.P)
		nto := to.Add(delta)
		g.BumpMoveActor(i, ai, to)
		if ai.DoesAny(PatternFourDirs) {
			g.FourDirectionalAttack(i, ai, AttackCharge)
		} else if j, aj := g.ActorAt(nto); j == PlayerID {
			g.BumpAttackActor(i, j, ai, aj, AttackCharge, 0)
		}
		return
	}
	// Next position is obstructed by another monster.
	if g.InFOV(ei.P) {
		return
	}
	// Outside of player's LOS, we simplify logic: they may lose a turn or
	// swap positions (if both want to exchange their positions). This is
	// cheaper and avoids the need for complex logic to prevent blockages.
	if len(aj.Behavior.Path) < 2 {
		return
	}
	if !aj.Behavior.SkipTurn && (aj.Behavior.Path[1] == ei.P || g.IntN(5) == 0) {
		// Swap monster positions and skip next turn.
		g.BumpMoveActor(i, ai, to)
		aj.Behavior.SkipTurn = true
	}
}

// MonsterUpdateTarget makes an ai entity aware about a new target position.
// It returns true if the monster's mind state changed.
func (g *Game) MonsterUpdateTarget(i ID, ai *Actor, at gruid.Point) bool {
	ei, beh := g.Entity(i), ai.Behavior
	if beh.State == Hunting {
		if g.InFOV(ei.P) {
			// If player's in FOV, then PP gets priority.
			beh.Target = g.PP()
		} else {
			beh.Target = at
		}
		return false
	}
	if g.InFOV(ei.P) && at == g.PP() {
		g.Logf("The %s notices you.", ei.Name)
		beh.State = Hunting
		if ai.DoesAny(MonsBarking) {
			g.Logf("The %s barks.", ei.Name)
			g.MakeNoise(ei.P, NoiseBark)
			g.PutStatus(PlayerID, g.PlayerActor(), StatusFear, DurationFearBark)
			if ai.Has(StatusConfusion) {
				g.LogfStyled("The confused %s bit its tongue while barking!", logHurtMons, g.Entity(i).Name)
				g.InflictDamage(i, ai, ConfusionDamage, AttackOther)
			}
		}
	}
	beh.Target = at
	return beh.State == Hunting
}

// RandomPassableNearby returns a random free passable point at a maximal distance
// maxdist from the given point (for small distances). Long distances are
// ignored and simply produce a random point somewhat biased toward origin.
func (g *Game) RandomPassableNearby(from gruid.Point, maxdist int) gruid.Point {
	if maxdist <= 2*MaxFOVRange {
		return g.RandomPassableClosePath(from, maxdist)
	}
	// For longer distances, maxdist is ignored. We simply generate a
	// random target, with some bias toward targets that are close.
	p := g.RandomPassable()
	q := g.RandomPassable()
	if paths.DistanceManhattan(from, p) > paths.DistanceManhattan(from, q) {
		return q
	}
	return p
}

// RandomPassableClosePath returns a random free passable point at a maximal
// passable path distance maxdist from the given point.
func (g *Game) RandomPassableClosePath(from gruid.Point, maxdist int) gruid.Point {
	dij := &MapPath{passable: g.Map.Passable}
	nodes := g.PR.BreadthFirstMap(dij, []gruid.Point{from}, maxdist)
	if len(nodes) == 0 {
		// Should not happen.
		return from
	}
	return nodes[g.IntN(len(nodes))].P
}

// CallToCommonTarget makes various nearby monsters chose the same target.  The
// idea is to make monsters occasionally naturally form small groups in a
// simple way (without requiring more complex herd mechanics).
func (g *Game) CallToCommonTarget(i ID, ai *Actor) {
	const maxdist = 6
	n := 0
	p := g.Entity(i).P
	for j, aj := range g.Monsters() {
		if j == i || aj.Behavior.State == Hunting || n >= 2 ||
			paths.DistanceManhattan(p, g.Entity(j).P) > maxdist ||
			aj.Behavior.Guard != InvalidPos {
			continue
		}
		aj.Behavior.Target = ai.Behavior.Target
		n++
	}
}

// MonsterFlee handles movement for an afraid hunting monster. When they're in
// a straight direction from the player, they'll flee randomly in a direction
// that avoids them (if any is free).
func (g *Game) MonsterFlee(i ID, ai *Actor) bool {
	ei, pl := g.Entity(i), g.PlayerEntity()
	if !g.InFOV(ei.P) || g.ActorInRange(ei.P, pl.P) != PlayerID {
		return false
	}
	var nbs paths.Neighbors
	ps := nbs.Cardinal(ei.P, func(p gruid.Point) bool {
		if j, _ := g.ActorAt(p); j >= 0 || !g.Map.Passable(p) {
			return false
		}
		return paths.DistanceManhattan(p, pl.P) > paths.DistanceManhattan(ei.P, p)
	})
	if len(ps) == 0 {
		// No free positions: lose a turn doing nothing.
		return true
	}
	to := ps[g.IntN(len(ps))]
	g.BumpMoveActor(i, ai, to)
	return true
}
