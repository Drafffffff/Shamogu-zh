package main

import (
	"math/rand/v2"
	"slices"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
	"codeberg.org/anaseto/gruid/rl"
)

// MapPath implements the paths.Pather interface and is used to provide pathing
// information in map generation.
type MapPath struct {
	passable func(gruid.Point) bool
	nbs      paths.Neighbors
}

func (mp *MapPath) Neighbors(p gruid.Point) []gruid.Point {
	return mp.nbs.Cardinal(p, mp.passable)
}

// MappingPath implements the paths.Pather interface and is used to provide pathing
// information in magic mapping and for connected components (filling interior
// walls).
type MappingPath struct {
	passable func(p gruid.Point) bool
	nbs      paths.Neighbors
}

func (mp *MappingPath) Neighbors(p gruid.Point) []gruid.Point {
	if !mp.passable(p) {
		// We don't want to mark interior walls as explored.
		return nil
	}
	return mp.nbs.Cardinal(p, inMap)
}

// PlayerPath returns a player path between two map positions.
func (g *Game) PlayerPath(from, to gruid.Point) []gruid.Point {
	path := []gruid.Point{}
	path = g.PR.JPSPath(path, from, to, g.PlayerPassableFunc(), false)
	if len(path) == 0 {
		return nil
	}
	return path
}

// PlayerPassableFunc returns a passability function suited for computing a
// player path.
func (g *Game) PlayerPassableFunc() func(gruid.Point) bool {
	knownAt := g.Map.KnownTerrain.At
	if g.PlayerActor().Statuses.Has(StatusDig) {
		return func(p gruid.Point) bool {
			return inMap(p) && knownAt(p) != Unknown
		}
	}
	passable := g.Map.Passable
	return func(p gruid.Point) bool {
		return passable(p) && knownAt(p) != Unknown
	}
}

// PlayerPassable reports whether a given position is passable for the player.
func (g *Game) PlayerPassable(p gruid.Point) bool {
	return g.PlayerPassableFunc()(p)
}

// MonPath implements paths.Astar for monster movement.
type MonPath struct {
	passable func(gruid.Point) bool // map passability for given monster
	nbs      paths.Neighbors
	actorAt  func(gruid.Point) bool // whether there's an actor at a given position
	pp       gruid.Point            // player's position
	rand     *rand.Rand
	confused bool // confused monsters won't move around other ones
}

func (aip *MonPath) Neighbors(p gruid.Point) []gruid.Point {
	nbs := aip.nbs.Cardinal(p, aip.passable)
	// Some cheap shuffling so that monster movement is not unnaturally
	// predictable.
	if aip.rand.IntN(2) == 0 {
		slices.Reverse(nbs)
	}
	return nbs
}

func (aip *MonPath) Cost(from, to gruid.Point) int {
	if !aip.confused && aip.actorAt(to) && to != aip.pp {
		return 5
	}
	return 1
}

func (aip *MonPath) Estimation(from, to gruid.Point) int {
	return paths.DistanceManhattan(from, to)
}

// MonPath returns a path between two positions for a monster actor.
func (g *Game) MonPath(i ID, ai *Actor, from, to gruid.Point) []gruid.Point {
	pp := g.PP()
	aip := &MonPath{
		actorAt:  g.Map.BoolCache.At,
		pp:       pp,
		passable: g.Map.Passable,
		rand:     g.rand,
		confused: ai.Has(StatusConfusion)}
	var path []gruid.Point
	ei := g.Entity(i)
	beh := ei.Role.(*Actor).Behavior
	if beh.State != Hunting && paths.DistanceManhattan(from, pp) > MaxFOVRange &&
		paths.DistanceManhattan(from, to) > MaxFOVRange {
		path = g.PR.JPSPath(beh.Path, from, to, aip.passable, false)
	}
	if len(path) == 0 {
		// We use a boolean cache marking actor positions for faster
		// cost computations during AstarPath.
		g.Map.BoolCache = g.Map.BoolCache.New()
		for _, ne := range g.Entities {
			if ne.IsAlive() {
				g.Map.BoolCache.Set(ne.P, true)
			}
		}
		if beh.State == Hunting && ai.DoesAny(MonsDig) {
			aip.passable = func(p gruid.Point) bool { return true }
		}
		path = g.PR.AstarPath(aip, from, to)
	}
	if len(path) == 0 {
		return nil
	}
	return path
}

// tunnelPath computes tunnel paths in map generation.
type tunnelPath struct {
	mg  *MapGen
	nbs paths.Neighbors
}

func (tp *tunnelPath) Neighbors(p gruid.Point) []gruid.Point {
	return tp.nbs.Cardinal(p, inMap)
}

func (tp *tunnelPath) Cost(from, to gruid.Point) int {
	if tp.mg.vault.At(from) && !tp.mg.tunnel.At(from) {
		return 50
	}
	cost := 1
	t := tp.mg.terrain.At(from)
	if tp.mg.vault.At(from) {
		cost += 7
	} else if !tp.mg.tunnel.At(from) && t != Floor {
		cost++
	}
	if Passable(t) {
		return cost
	}
	wc := countWalls(tp.mg.terrain, from, 1, true)
	return cost + 8 - wc
}

func countWalls(gd rl.Grid, p gruid.Point, radius int, countOut bool) int {
	count := 0
	rg := gruid.Range{
		gruid.Point{p.X - radius, p.Y - radius},
		gruid.Point{p.X + radius + 1, p.Y + radius + 1},
	}
	if countOut {
		osize := rg.Size()
		rg = rg.Intersect(gd.Range())
		size := rg.Size()
		count += osize.X*osize.Y - size.X*size.Y
	} else {
		rg = rg.Intersect(gd.Range())
	}
	gd = gd.Slice(rg)
	count += gd.Count(rl.Cell(Wall))
	return count
}

func (tp *tunnelPath) Estimation(from, to gruid.Point) int {
	return paths.DistanceManhattan(from, to)
}
