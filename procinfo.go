package main

// MapLevels is the total number of map levels in the dungeon.
const MapLevels = 9

// ProcInfo contains information for procedural generation, like which items
// were already generated, and so on. Note that slices indexed by map level
// start from zero (for map level 1).
type ProcInfo struct {
	Layouts          []MapLayout // map layouts
	Earthquake       int         // level with extra rubble
	FakePortal       []bool      // levels with fake portal
	GuardianTotem1   int         // level with early totem guardian
	GuardianTotem2   int         // level with early totem guardian
	GuardianPortal1  int         // level with early-mid portal guardian
	GuardianPortal2  int         // level with mid-late portal guardian
	WanderingUnique1 int         // level with walking mushroom
	WanderingUnique2 int         // level with noisy imp or teleporting ghost
	MonsEarly        int         // special early level
	MonsMid          int         // special mid level
	MonsMidLate      int         // special mid-late level
	MonsLate         int         // special late level
	Spirits          []int       // spirit index for each dungeon level (-1 if no spirit)
	Menhirs          []int       // menhir-generation data
	MenhirIdx        int         // current index in Menhirs
	Comestibles      []int       // comestible-generation data
	ComestibleIdx    int         // current index in Comestibles
	NComestibles     []int       // number of comestibles per level
	NMenhirs         []int       // number of menhirs per level
}

// InitProcInfo initializes procedural generation starting information.
func (g *Game) InitProcInfo() {
	g.ProcInfo = &ProcInfo{}
	g.layoutsProcGen()
	g.spiritProcGen()
	g.comestiblesProcGen()
	g.menhirsProcGen()
	g.flavoursProcGen()
}

func (g *Game) layoutsProcGen() {
	layouts := []MapLayout{
		CellularAutomataCave, RandomWalkTreeCave, RandomWalkCave,
		CellularAutomataCave, RandomWalkTreeCave, RandomWalkCave,
		MixedAutomataWalkCave, MixedAutomataWalkTreeCave, MixedWalkCaveWalkTreeCave,
	}
	if g.IntN(2) == 0 {
		layouts[3] = MixedAutomataWalkCave
	}
	if g.IntN(2) == 0 {
		layouts[4] = MixedAutomataWalkTreeCave
	}
	if g.IntN(3) == 0 {
		layouts[5] = MixedWalkCaveWalkTreeCave
	}
	g.rand.Shuffle(len(layouts), func(i, j int) {
		layouts[i], layouts[j] = layouts[j], layouts[i]
	})
	// layouts[g.rand.IntN(len(layouts))] = MapLayout(g.rand.IntN(NLayouts))
	g.ProcInfo.Layouts = layouts
}

func (g *Game) flavoursProcGen() {
	gpi := g.ProcInfo
	gpi.Earthquake = 2 + g.IntN(MapLevels-1)
	gpi.FakePortal = make([]bool, MapLevels)
	nfakes := 1 // at least one fake portal per game
	if g.IntN(4) > 0 {
		nfakes++
	}
	if g.IntN(10) == 0 {
		nfakes++
	}
	for range nfakes {
		// Chose random level in 2..8 (included) for malfunctioning
		// portal.
		gpi.FakePortal[1+g.IntN(7)] = true
	}
	randTotemLevel := func(n, m int) int {
		for {
			i := n + g.IntN(m)
			if gpi.Spirits[i-1] >= 0 {
				// There's a totem to guard on that level.
				return i
			}
		}
	}
	gpi.GuardianTotem1 = randTotemLevel(3, 4)
	for gpi.GuardianTotem2 == 0 || gpi.GuardianTotem2 == gpi.GuardianTotem1 {
		gpi.GuardianTotem2 = randTotemLevel(3, 4)
	}
	gpi.GuardianPortal2 = 6 + g.IntN(3)
	for gpi.GuardianPortal1 == 0 || gpi.GuardianPortal1 == gpi.GuardianPortal2 {
		gpi.GuardianPortal1 = 4 + g.IntN(4)
	}
	gpi.WanderingUnique1 = 4 + g.IntN(6)
	gpi.WanderingUnique2 = 4 + g.IntN(6)
	gpi.MonsEarly = g.IntN(4)
	gpi.MonsMid = 4 + g.IntN(3)
	gpi.MonsMidLate = 6 + g.IntN(4)
	if gpi.MonsMidLate == gpi.MonsMid {
		gpi.MonsMidLate++
	}
	gpi.MonsLate = 7 + g.IntN(3)
	// Disable very occasionally some events/flavours, for some
	// extra unpredictability.
	switch g.IntN(10) {
	case 0:
		gpi.Earthquake = 0
	case 1:
		gpi.WanderingUnique1 = 0
	case 2:
		gpi.WanderingUnique2 = 0
	}
}

func (g *Game) spiritProcGen() {
	// spirits represents a permutation of indices (corresponding to
	// secondary spirits).
	spirits := make([]int, len(secondarySpirits))
	for i := range spirits {
		spirits[i] = i
	}
	g.rand.Shuffle(len(spirits), func(i, j int) {
		spirits[i], spirits[j] = spirits[j], spirits[i]
	})
	// mapSpirits maps dungeon levels to the index of the totemic spirit
	// they contain.
	mapSpirits := make([]int, MapLevels)
	for i := range mapSpirits {
		// We use -1 to represent the absence of totemic spirit in a
		// level.
		mapSpirits[i] = -1
	}
	// Games generate at most 6 totems among possible ones.
	copy(mapSpirits, spirits[:min(6, len(spirits))])
	// We shuffle over more levels to possibly leave early ones without
	// spirit.
	s := mapSpirits[1:8]
	g.rand.Shuffle(len(s), func(i, j int) {
		s[i], s[j] = s[j], s[i]
	})
	g.ProcInfo.Spirits = mapSpirits
}

// menhirProcGen creates a random sequence of menhir kinds that still ensures
// that all menhirs appear at least once in a while.
func (g *Game) menhirProcGen() {
	g.ProcInfo.Menhirs = g.genRandomIndices(len(MenhirData))
}

// genRandomIndices produces 2*l not-too-random indices up to length l.
func (g *Game) genRandomIndices(l int) []int {
	indices := make([]int, 2*l)
	for i := range indices {
		indices[i] = i % l
	}
	for i := range l {
		indices[i] = g.rand.IntN(l)
	}
	g.rand.Shuffle(len(indices), func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})
	return indices
}

// NextMenhirIdx returns the next menhir kind (index of menhirData) to be
// generated.
func (g *Game) NextMenhirIdx() int {
	gpi := g.ProcInfo
	if gpi.MenhirIdx >= len(gpi.Menhirs) {
		gpi.MenhirIdx = 0
		g.menhirProcGen()
	}
	i := gpi.Menhirs[gpi.MenhirIdx]
	gpi.MenhirIdx++
	return i
}

// comestibleProcGen creates a random sequence of comestibles kinds that still
// ensures that all comestibles appear at least once in a while.
func (g *Game) comestibleProcGen() {
	g.ProcInfo.Comestibles = g.genRandomIndices(len(ComestibleData))
}

// NextComestibleIdx returns the next comestible kind (index of
// comestibleData) to be generated.
func (g *Game) NextComestibleIdx() int {
	gpi := g.ProcInfo
	if gpi.ComestibleIdx >= len(gpi.Comestibles) {
		gpi.ComestibleIdx = 0
		g.comestibleProcGen()
	}
	i := gpi.Comestibles[gpi.ComestibleIdx]
	gpi.ComestibleIdx++
	return i
}

func (g *Game) comestiblesProcGen() {
	g.ProcInfo.NComestibles = []int{
		2, 2, 3, 3, 4,
		4, 5, 5, 6,
	}
	g.ProcInfo.NComestibles[g.IntN(2)]++
	g.ProcInfo.NComestibles[2+g.IntN(3)]++
	g.ProcInfo.NComestibles[5+g.IntN(2)]++
	g.ProcInfo.NComestibles[7+g.IntN(2)]++
	g.ProcInfo.NComestibles[1+g.IntN(8)]++
	g.ProcInfo.NComestibles[1+g.IntN(8)]--
}

func (g *Game) menhirsProcGen() {
	// Maximum of 4 in any level.
	g.ProcInfo.NMenhirs = []int{
		1, 1, 1, 1, 2,
		2, 2, 2, 2,
	}
	// Extra menhirs.
	g.ProcInfo.NMenhirs[2+g.IntN(MapLevels-2)]++ // levels 3-9
	g.ProcInfo.NMenhirs[g.IntN(2)]++             // levels 1-2
	g.ProcInfo.NMenhirs[2+g.IntN(2)]++           // levels 3-4
	g.ProcInfo.NMenhirs[5+g.IntN(2)]++           // levels 6-7
	g.ProcInfo.NMenhirs[7+g.IntN(2)]++           // levels 8-9
	for i, n := range g.ProcInfo.NMenhirs {
		// Never generate more than 3 menhirs on a given level.
		g.ProcInfo.NMenhirs[i] = min(3, n)
	}
}
