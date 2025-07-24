package main

import (
	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
)

// MaxFOVRange is the maximum distance in player's field of view.
const MaxFOVRange = 8

// UpdateFOV updates the field of view. It has to be called each time the
// player moves or some terrain change impacts vision.
func (g *Game) UpdateFOV() {
	pp := g.PP()
	rg := gruid.NewRange(-MaxFOVRange, -MaxFOVRange, MaxFOVRange+1, MaxFOVRange+1)
	g.Map.FOV.SetRange(rg.Add(pp).Intersect(g.Map.Terrain.Range()))
	passable := func(p gruid.Point) bool {
		t := g.Map.Terrain.At(p)
		return t != Wall && t != Rubble
	}
	lt := &lighter{g: g}
	g.Map.FOV.VisionMap(lt, pp)
	g.Map.FOVPts = g.Map.FOV.SSCVisionMap(pp, MaxFOVRange, passable, false)
}

// UpdateKnowledge updates player's knowledge based on FOV. It has to be called at
// the end of each turn.
func (g *Game) UpdateKnowledge() {
	pp := g.PP()
	for _, p := range g.Map.FOVPts {
		if paths.DistanceManhattan(p, pp) > MaxFOVRange {
			continue
		}
		cost, ok := g.Map.FOV.At(p)
		if !ok || cost > MaxFOVRange {
			continue
		}
		g.SeeTerrain(p)
	}
	g.SeeEntities()
}

// SeeTerrain handles knowledge changes about a given position during an FOV
// update.
func (g *Game) SeeTerrain(p gruid.Point) {
	t := g.Map.Terrain.At(p)
	if g.Map.KnownTerrain.At(p) != t {
		g.md.auto.aemRebuild = true
		g.Map.KnownTerrain.Set(p, t)
	}
}

// SeeEntities handles knowledge changes about the last known position of
// map entities that are in the field of view.
func (g *Game) SeeEntities() {
	for _, e := range g.NPMapEntities() {
		if g.InFOV(e.P) {
			g.SenseEntity(e, "see")
		} else if g.InFOV(e.KnownP) {
			e.KnownP = InvalidPos
		}
	}
}

// SenseEntity updates known position of an entity to its real one. The given
// verb is used in the log message (one of "see", "sense").
func (g *Game) SenseEntity(e *Entity, verb string) {
	if !e.Seen {
		e.Seen = true
		g.md.auto.mode = noAuto
		switch r := e.Role.(type) {
		case *Actor:
			if r.IsAlive() {
				g.Logf("You %s %s.", verb, One(e.Name))
				if r.DoesAny(MonsNotable) {
					g.StoryLogf("Saw %s (distance: %d)", One(e.Name), g.ppDist(e.P))
				}
			}
		case *Comestible:
			g.LogfStyled("You %s %s.", logNotable, verb, One(e.Name))
			g.StoryLogf("Spotted %s (distance: %d)", One(e.Name), g.ppDist(e.P))
		case *Spirit:
			g.LogfStyled("You %s %s totem.", logNotable, verb, One(e.Name))
			g.StoryLogf("Spotted %s totem (distance: %d)", One(e.Name), g.ppDist(e.P))
		case *CorruptionOrb:
			g.LogStyled("You found the orb of corruption!", logNotable)
			g.StoryLogf("Found the orb of corruption (distance: %d)", g.ppDist(e.P))
		default:
			g.LogfStyled("You %s %s.", logNotable, verb, One(e.Name))
			g.StoryLogf("Discovered %s (distance: %d)", One(e.Name), g.ppDist(e.P))
		}
	}
	e.KnownP = e.P
	if r, ok := e.Role.(*Actor); ok && r.IsDead() {
		e.KnownDead = true
	}
}

func (g *Game) ppDist(p gruid.Point) int {
	return paths.DistanceManhattan(g.PP(), p)
}

// ResetKnowledge initializes player's knowledge of map terrain.
func (g *Game) ResetKnowledge() {
	g.Map.KnownTerrain.Fill(Unknown)
	for _, e := range g.NPMapEntities() {
		e.KnownP = InvalidPos
	}
}

// InFOV returns true if p is in the player's field of view.
func (g *Game) InFOV(p gruid.Point) bool {
	cost, ok := g.Map.FOV.At(p)
	return ok && cost <= MaxFOVRange && g.Map.FOV.Visible(p)
}

// DangerInFOV returns true if there is a foe in the player's field of view.
func (g *Game) DangerInFOV() bool {
	for i, ai := range g.Monsters() {
		if ai.IsAlive() && g.InFOV(g.Entity(i).P) {
			return true
		}
	}
	return g.dangerousCloudInFOV()
}

// DangerInProximity returns true if there is a sensed foe in the player's
// proximity (either from seen in FOV or from footstep noise).
func (g *Game) DangerInProximity() bool {
	for i, ai := range g.Monsters() {
		ei := g.Entity(i)
		if ai.IsAlive() && (g.InFOV(ei.P) || ei.Noise) {
			return true
		}
	}
	return g.dangerousCloudInFOV()
}

func (g *Game) dangerousCloudInFOV() bool {
	for _, cl := range g.Map.Clouds.Clouds {
		if g.InFOV(cl.P) && cl.Kind != CloudFog {
			return true
		}
	}
	return false
}

// lighter implements rl.Lighter
type lighter struct {
	g *Game
}

func (lt *lighter) MaxCost(src gruid.Point) int {
	return MaxFOVRange + 1
}

func (lt *lighter) Cost(src, from, to gruid.Point) int {
	g := lt.g
	wallcost := lt.MaxCost(src)
	// no terrain cost on origin
	if src == from {
		// specific diagonal costs
		opaque, hard := g.opaqueDiagonal(from, to)
		if opaque {
			return wallcost
		}
		if hard {
			return wallcost - 1
		}
		return paths.DistanceManhattan(to, from)
	}
	// from terrain specific costs
	t := g.Map.Terrain.AtU(from)
	if t == Wall || t == Rubble {
		return wallcost
	}
	if _, ok := g.Map.Clouds.At(from); ok {
		return wallcost
	}
	// specific diagonal costs
	opaque, hard := g.opaqueDiagonal(from, to)
	if opaque {
		return wallcost
	}
	if t == Foliage {
		return wallcost + paths.DistanceManhattan(to, from) - 3
	}
	if hard {
		cost := wallcost - paths.DistanceManhattan(from, src) - 1
		if cost < 1 {
			cost = 1
		}
		return cost
	}
	return paths.DistanceManhattan(to, from)
}

func (g *Game) opaqueDiagonal(from, to gruid.Point) (opaque, hard bool) {
	// The state uses cardinal movement only, so two diagonal walls should,
	// for example, block line of sight. This is in contrast with the main
	// mechanics of the line of sight algorithm, which for gameplay reasons
	// allows diagonals for light rays in normal circumstances.
	var ps [2]gruid.Point
	delta := to.Sub(from)
	switch delta {
	case gruid.Point{1, -1}:
		ps[0] = from.Shift(1, 0)
		ps[1] = from.Shift(0, -1)
	case gruid.Point{-1, -1}:
		ps[0] = from.Shift(-1, 0)
		ps[1] = from.Shift(0, -1)
	case gruid.Point{-1, 1}:
		ps[0] = from.Shift(-1, 0)
		ps[1] = from.Shift(0, 1)
	case gruid.Point{1, 1}:
		ps[0] = from.Shift(1, 0)
		ps[1] = from.Shift(0, 1)
	default:
		return false, false
	}
	opaque = true
	hard = true
	for _, p := range ps {
		if _, ok := g.Map.Clouds.At(p); ok {
			continue
		}
		switch g.Map.Terrain.AtU(p) {
		case Wall, Rubble:
		case Foliage:
			opaque = false
		default:
			return false, false
		}
	}
	return opaque, hard
}
