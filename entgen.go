package main

import (
	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
)

// InitPlayer initializes the player entity for the rest of the game.
func (g *Game) InitPlayer() {
	g.Entities[PlayerID] = &Entity{
		Name: "You",
		Rune: '@',
		Role: NewActor(2, 1, 9, Player),
	}
	for i := PrimarySpiritID; i < FirstMapID; i++ {
		g.Entities[i] = emptySlot()
	}
	// g.Entities[PrimarySpiritID+1] = spiritEntity(secondarySpirits[0])
	g.ComputePlayerStats()
}

func emptySlot() *Entity {
	return &Entity{Name: "(empty slot)", P: InvalidPos, Seen: true}
}

// GenEntities generates the entities for a new level, placing them on the map.
func (g *Game) GenEntities(mg *MapGen) {
	// Reset non-player actors.
	clear(g.Entities[FirstMapID+1:])
	g.Entities = g.Entities[:FirstMapID+1]
	g.Map.ActorCache = g.Map.ActorCache.New()
	// Generate new items and actors.
	g.genPlayerOnNewLevel(mg)
	g.genItems(mg)
	g.genMonsters()
}

// genPlayerOnNewLevel places the player in the map with replenished charges and HP.
func (g *Game) genPlayerOnNewLevel(mg *MapGen) {
	p := g.PlayerEntity()
	p.P = mg.RandomVaultsPlace(mg.vaults[:len(mg.vaults)/3], PlaceActor)
	if p.P == InvalidPos {
		panic("invalid player position")
	}
	p.KnownP = p.P
	g.Map.ActorCache.SetU(p.P, PlayerID)
	for _, sp := range g.PlayerSpirits() {
		sp.Charges = sp.MaxCharges[sp.Level]
	}
	pa := g.PlayerActor()
	clear(pa.Statuses)
	pa.HP = pa.MaxHP
}

// spiritInfo gathers information for generating a specific kind of spirit
// entity.
type spiritInfo struct {
	Name         string     // name of the spirit
	MaxCharges   [2]int     // maximum number of charges per level
	Ability      [2]Ability // active ability effect when used
	BonusAttack  [2]int     // attack bonuses per level (may be zero)
	BonusDefense [2]int     // defense bonuses per level (may be zero)
	BonusTraits  [2]Traits  // bonus traits per level
	BonusHP      [2]int     // bonus hp per level (may be zero)
}

// primarySpirits represents the various kinds of primary spirits.
var primarySpirits = []spiritInfo{
	{
		Name:        "Four-Headed Hydra",
		MaxCharges:  [2]int{2, 3},
		Ability:     [2]Ability{EffectFocus{}, EffectFocus{}},
		BonusHP:     [2]int{0, 3},
		BonusTraits: [2]Traits{PatternFourDirs, PatternFourDirs}},
	{
		Name:        "Rampaging Boar",
		MaxCharges:  [2]int{2, 3},
		Ability:     [2]Ability{EffectDig{}, EffectDig{}},
		BonusAttack: [2]int{1, 1},
		BonusTraits: [2]Traits{PatternRampage | Pushing, PatternRampage | PushingUnbalance}},
	{
		Name:         "Jumping Frog",
		MaxCharges:   [2]int{2, 4},
		Ability:      [2]Ability{EffectJump{}, EffectJump{}},
		BonusDefense: [2]int{1, 1},
		BonusTraits:  [2]Traits{PatternCatch, PatternCatchUnbalance}},
	{
		Name:        "Wind Fox",
		MaxCharges:  [2]int{2, 4},
		Ability:     [2]Ability{EffectPushingGale{}, EffectPushingGale{}},
		BonusTraits: [2]Traits{PatternRanged, PatternRangedRecoil}},
	{
		Name:        "Temporal Cat",
		MaxCharges:  [2]int{2, 3},
		Ability:     [2]Ability{EffectTimeStop{}, EffectTimeStop{}},
		BonusTraits: [2]Traits{PatternSwap, PatternSwapDaze}},
}

// secondarySpirits represents the various kinds of secondary spirit totems
// that can be found.
var secondarySpirits = []spiritInfo{
	{
		Name:        "Thunder Porcupine",
		MaxCharges:  [2]int{2, 3},
		Ability:     [2]Ability{EffectLightning{}, EffectLightning{}},
		BonusAttack: [2]int{1, 1},
		BonusTraits: [2]Traits{DazingSpines, DazingSpines | ResistanceDaze}},
	{
		Name:        "Barking Hound",
		MaxCharges:  [2]int{2, 4},
		Ability:     [2]Ability{EffectBark{}, EffectBark{}},
		BonusAttack: [2]int{1, 1},
		BonusTraits: [2]Traits{GoodHearing, GoodHearing | ResistanceFear}},
	{
		Name:        "Stinking Skunk",
		MaxCharges:  [2]int{2, 3},
		Ability:     [2]Ability{EffectNoxiousSmell{}, EffectNoxiousSmell{}},
		BonusAttack: [2]int{1, 1},
		BonusTraits: [2]Traits{GoodSmell, GoodSmell | ResistanceConfusion}},
	{
		Name:         "Walking Tree",
		MaxCharges:   [2]int{2, 3},
		Ability:      [2]Ability{EffectLignify{}, EffectLignify{}},
		BonusDefense: [2]int{2, 2},
		BonusTraits:  [2]Traits{VulnerabilityFire, VulnerabilityFire | WoodyLegs}},
	{
		Name:         "Venomous Viper",
		MaxCharges:   [2]int{2, 3},
		Ability:      [2]Ability{EffectPoisonCloud{}, EffectPoisonCloud{}},
		BonusDefense: [2]int{1, 1},
		BonusTraits:  [2]Traits{VenomousMelee, VenomousMelee | ResistancePoison}},
	{
		Name:         "Sprinting Gazelle",
		MaxCharges:   [2]int{2, 3},
		Ability:      [2]Ability{EffectSprint{}, EffectSprint{}},
		BonusDefense: [2]int{1, 1},
		BonusTraits:  [2]Traits{TrailingCloud, TrailingCloud | ResistanceImbalance}},
}

// spiritEntity generates a new spirit entity using the given data.
func spiritEntity(si spiritInfo) *Entity {
	return &Entity{
		Name:   si.Name,
		KnownP: InvalidPos,
		Role: &Spirit{
			Charges:      si.MaxCharges[0],
			MaxCharges:   si.MaxCharges,
			Ability:      si.Ability,
			BonusAttack:  si.BonusAttack,
			BonusDefense: si.BonusDefense,
			BonusTraits:  si.BonusTraits,
			BonusHP:      si.BonusHP}}
}

// itemInfo gathers information for item generation.
type itemInfo struct {
	Name   string
	Effect Effect
}

// ComestibleData provides information about the various kinds of comestibles.
var ComestibleData = []itemInfo{
	{"ambrosia berries", EffectAmbrosiaBerries{}},
	{"clarity leaves", EffectClarityLeaves{}},
	{"berserking flower", EffectBerserkingFlower{}},
	{"foggy-skin onion", EffectFoggySkinOnion{}},
	{"firebreath pepper", EffectFirebreathPepper{}},
	{"lignification fruit", EffectLignificationFruit{}},
	{"teleport mushroom", EffectTeleportMushroom{}},
}

// comestibleEntity generates a new comestible Entity entity using the given
// data.
func comestibleEntity(ci itemInfo) *Entity {
	return &Entity{
		Name:   ci.Name,
		Rune:   '%',
		KnownP: InvalidPos,
		Role:   &Comestible{Effect: ci.Effect}}
}

// MenhirData provides information about the various kinds of menhirs.
var MenhirData = []itemInfo{
	{"mapping menhir", EffectMappingMenhir{}},
	{"teleport menhir", EffectTeleportMenhir{}},
	{"poison menhir", EffectPoisonMenhir{}},
	{"fire menhir", EffectFireMenhir{}},
}

// menhirEntity generates a new menhir Entity entity using the given data.
func menhirEntity(ci itemInfo) *Entity {
	return &Entity{
		Name:   ci.Name,
		Rune:   '&',
		KnownP: InvalidPos,
		Role:   &Menhir{Effect: ci.Effect}}
}

// genItems generates items for the new level.
func (g *Game) genItems(mg *MapGen) {
	// Portal.
	if g.Map.Level < MapLevels {
		portal := &Entity{Name: "magic portal", Rune: '>', Role: &Portal{}}
		portal.P = mg.RandomVaultsPlace(mg.vaults[len(mg.vaults)/2:], PlaceStatic)
		g.Map.Portal = portal.P
		g.AddEntity(portal)
		if g.ProcInfo.FakePortal == g.Map.Level {
			portal := &Entity{Name: "magic portal", Rune: '>', Role: &Portal{Fake: true}}
			portal.P = mg.RandomVaultsPlace(mg.vaults[len(mg.vaults)/2:], PlaceStatic)
			g.AddEntity(portal)
		}
	} else {
		orb := &Entity{Name: "Orb of Corruption", Rune: '☼', Role: &CorruptionOrb{}}
		orb.P = mg.RandomVaultsPlace(mg.vaults[(2*len(mg.vaults))/3:], PlaceStatic)
		g.Map.Orb = orb.P
		g.AddEntity(orb)
	}
	// Menhirs.
	nmenhirs := g.ProcInfo.NMenhirs[g.Map.Level-1]
	for range nmenhirs {
		me := menhirEntity(MenhirData[g.NextMenhirIdx()])
		me.P = mg.RandomPlace(PlaceStatic)
		g.AddEntity(me)
	}
	// Totemic spirit.
	if spiritIndex := g.ProcInfo.Spirits[g.Map.Level-1]; spiritIndex >= 0 {
		sp := spiritEntity(secondarySpirits[spiritIndex])
		spr := sp.Role.(*Spirit)
		spr.Charges = spr.MaxCharges[spr.Level]
		sp.Rune = '!'
		sp.P = mg.RandomPlace(PlaceItem)
		g.Map.Totem = sp.P
		g.AddEntity(sp)
	}
	// Comestibles.
	nitems := g.ProcInfo.NComestibles[g.Map.Level-1]
	for i := range nitems {
		co := comestibleEntity(ComestibleData[g.NextComestibleIdx()])
		if i <= (nitems+1)/2 {
			co.P = mg.RandomPlace(PlaceItem)
		} else {
			co.P = g.randomFreeItemFloor()
		}
		g.AddEntity(co)
	}
}

// randomFreeItemFloor returns a random map floor position that is not occuppied by
// an item.
func (g *Game) randomFreeItemFloor() gruid.Point {
loop:
	for {
		p := g.RandomPassable()
		for i := range g.ItemEntities() {
			if g.Entity(i).P == p {
				continue loop
			}
		}
		// We replace foliage or rubble with floor for putting the
		// item. We could allow items on non-floor cells, but for
		// clarity, it's better this way.
		g.Map.Terrain.Set(p, Floor)
		return p
	}
}

// monsterInfo gathers information for generating a specific kind of monster.
type monsterInfo struct {
	Name    string // name of the monster
	R       rune   // rune (for display)
	Attack  int    // actor's attack
	Defense int    // actor's defense
	HP      int    // actor's max HP
	Traits  Traits // actor's traits
}

// monsterKind represents the various kinds of monsters.
type monsterKind int

const (
	AcidMound monsterKind = iota
	BarkingHound
	BerserkingSpider
	BlinkButterfly
	BurningPhoenix
	ConfusingEye
	CrazyDruid
	DisgustingRat
	EarthDragon
	ExplodingNadre
	FearsomeLich
	FireLama
	FourHeadedHydra
	LashingFrog
	MadOctopode
	RampagingBoar
	TemporalCat
	ThunderPorcupine
	UndeadKnight
	VenomousViper
	WalkingTree
	WindFox
)

// MonsData provides information about the various kinds of monsters.
var MonsData = []monsterInfo{
	// 		  {name, rune, attack, defense, hp, traits}
	AcidMound:        {"acid mound", 'a', 2, 0, 5, MonsIgnoreDefense | MonsImmunityImbalance | MonsCreep},
	BarkingHound:     {"barking hound", 'h', 3, 0, 4, MonsBarking | GoodHearing | MonsImmunityFear},
	BerserkingSpider: {"berserking spider", 's', 2, 1, 3, MonsBerserking | MonsSilent},
	BlinkButterfly:   {"blinking butterfly", 'b', 2, 3, 2, MonsBlink | MonsImmunityImbalance | MonsImmunityLignification | MonsWingFlap},
	BurningPhoenix:   {"burning phoenix", 'P', 3, 1, 5, PatternRampage | MonsBurn | MonsImmunityFire | MonsImmunityLignification | MonsWingFlap | MonsNotable},
	ConfusingEye:     {"confusing eye", 'e', 2, 0, 2, PatternRanged | MonsConfusion | MonsImmunityConfusion | MonsLightFootsteps},
	CrazyDruid:       {"crazy druid", 'C', 3, 1, 4, PatternSwap | MonsBurn | MonsBerserking | MonsNotable},
	DisgustingRat:    {"disgusting rat", 'r', 2, 0, 3, MonsNausea | MonsLightFootsteps},
	EarthDragon:      {"earth dragon", 'D', 4, 2, 6, PushingUnbalance | MonsDig | MonsImmunityFear | MonsHeavyFootsteps | MonsNotable},
	ExplodingNadre:   {"exploding nadre", 'n', 2, 2, 1, MonsExplodingDeath},
	FearsomeLich:     {"fearsome lich", 'L', 2, 2, 5, PatternRanged | MonsFear | MonsIgnoreDefense | MonsImmunityPoison | MonsNotable},
	FireLama:         {"fire lama", 'l', 2, 0, 4, PatternRanged | MonsSpitFire},
	FourHeadedHydra:  {"four-headed hydra", 'H', 2, 0, 6, PatternFourDirs | MonsFourHeaded | MonsHeavyFootsteps | MonsNotable},
	MadOctopode:      {"mad octopode", 'O', 3, 2, 5, PatternCatchUnbalance | MonsImmunityImbalance | VenomousMelee | MonsCreep | MonsNotable},
	LashingFrog:      {"lashing frog", 'F', 2, 1, 4, PatternCatchUnbalance},
	RampagingBoar:    {"rampaging boar", 'B', 3, 0, 4, PatternRampage | PushingUnbalance | MonsDig},
	TemporalCat:      {"temporal cat", 'c', 2, 0, 4, PatternSwapDaze | MonsLightFootsteps},
	ThunderPorcupine: {"thunder porcupine", 'p', 2, 0, 3, DazingSpines | MonsImmunityDaze | MonsLightFootsteps},
	UndeadKnight:     {"undead knight", 'K', 3, 3, 4, PatternFourDirs | MonsFear | MonsImmunityPoison | MonsNotable},
	VenomousViper:    {"venomous viper", 'v', 2, 1, 4, VenomousMelee | MonsImmunityPoison | MonsCreep},
	WalkingTree:      {"walking tree", 'T', 2, 4, 5, MonsLignify | MonsImmunityImbalance | MonsImmunityPoison | MonsImmunityLignification | VulnerabilityFire | MonsHeavyFootsteps | MonsNotable},
	WindFox:          {"wind fox", 'f', 2, 0, 4, PatternRangedRecoil | MonsLightFootsteps},
}

// monsEarly represents early level monsters.
var monsEarly = []monsterKind{
	BerserkingSpider,
	DisgustingRat,
	ConfusingEye,
	ThunderPorcupine,
}

// monsMid represents monters that appear throughout the game (but less often
// early).
var monsMid = []monsterKind{
	AcidMound,
	BarkingHound,
	BlinkButterfly,
	ExplodingNadre,
	FireLama,
	LashingFrog,
	RampagingBoar,
	TemporalCat,
	VenomousViper,
	WindFox,
}

// monsLate represents dangerous monsters that appear mostly in mid-to-late
// levels.
var monsLate = []monsterKind{
	BurningPhoenix,
	EarthDragon,
	FourHeadedHydra,
	WalkingTree,
}

// genMonsters generates monster entities for the current level.
func (g *Game) genMonsters() {
	if g.Map.Orb != InvalidPos {
		g.genMonsterGuardian(FearsomeLich, g.Map.Orb)
		g.genMonsterGuardian(UndeadKnight, g.Map.Orb)
	}
	if g.ProcInfo.GuardianMid == g.Map.Level {
		g.genMonsterGuardian(MadOctopode, g.Map.Portal)
	}
	if g.ProcInfo.GuardianEarly == g.Map.Level {
		g.genMonsterGuardian(CrazyDruid, g.Map.Totem)
	}
	// number of early, mid, late monsters.
	nE, nM, nL := nMonsE[g.Map.Level-1], nMonsM[g.Map.Level-1], nMonsL[g.Map.Level-1]
	switch g.Map.Level {
	case g.ProcInfo.MonsEarly:
		// Special level with many early monsters of same kind.
		nE += 2
		nM--
		mk := monsEarly[g.IntN(len(monsEarly))]
		for range nE - 2 {
			g.genMonster(mk)
		}
		nE = 2 // two remaining
	case g.ProcInfo.MonsMid:
		// Special level with many mid monsters of same kind.
		nE -= 2
		nM += 2
		mk := monsMid[g.IntN(len(monsMid))]
		for range nM - 2 {
			g.genMonster(mk)
		}
		nM = 2 // two remaining
	case g.ProcInfo.MonsLate:
		// Special level with many late monsters of same kind.
		nE -= 2
		nM -= 2
		nL += 3
		mk := monsLate[g.IntN(len(monsLate))]
		for range nL - 2 {
			g.genMonster(mk)
		}
		nL = 2 // two remaining
	}
	for range nE {
		g.genMonster(monsEarly[g.IntN(len(monsEarly))])
	}
	for range nM {
		g.genMonster(monsMid[g.IntN(len(monsMid))])
	}
	for range nL {
		g.genMonster(monsLate[g.IntN(len(monsLate))])
	}
}

var nMonsE = [MapLevels]int{3, 5, 6, 6, 5, 5, 5, 4, 4}
var nMonsM = [MapLevels]int{2, 3, 4, 6, 8, 10, 10, 10, 10}
var nMonsL = [MapLevels]int{0, 0, 1, 1, 2, 2, 3, 4, 5}

// genMonster spawns a monster of the given kind on a random free floor.
func (g *Game) genMonster(mk monsterKind) {
	p := g.PlayerEntity().P
	q := g.randomFreeFloor()
	// XXX: use walkable-path distance?
	for paths.DistanceManhattan(p, q) <= MaxFOVRange+1 {
		q = g.randomFreeFloor()
	}
	g.genMonsterAt(mk, q)
}

func (g *Game) genMonsterAt(mk monsterKind, p gruid.Point) *Entity {
	mons := monster(MonsData[mk])
	mons.P = p
	g.Map.ActorCache.SetU(mons.P, ID(len(g.Entities)))
	g.AddEntity(mons)
	return mons
}

// genMonsterGardian spawns a gardian monster of the given kind around a
// certain position.
func (g *Game) genMonsterGuardian(mk monsterKind, at gruid.Point) {
	p := g.randomFreeNearby(at, 5)
	mons := g.genMonsterAt(mk, p)
	a := mons.Actor()
	a.Behavior.Guard = at
}

// monster generates a new monster entity using the given data.
func monster(mi monsterInfo) *Entity {
	return &Entity{
		Name:   mi.Name,
		Rune:   mi.R,
		KnownP: InvalidPos,
		Role:   NewActor(mi.Attack, mi.Defense, mi.HP, mi.Traits),
	}
}

// randomFreeFloor returns a random map floor position that is not occuppied by
// an actor.
func (g *Game) randomFreeFloor() gruid.Point {
loop:
	for {
		p := g.RandomPassable()
		if i, _ := g.ActorAt(p); i >= 0 {
			continue loop
		}
		return p
	}
}

// randomFreeNearby returns a random map floor position that is not occuppied
// by an actor near the given position.
func (g *Game) randomFreeNearby(at gruid.Point, maxdist int) gruid.Point {
	n := 0
loop:
	for {
		if n > 1000 {
			return g.randomFreeFloor()
		}
		p := g.RandomPassableWithin(at, maxdist)
		if i, _ := g.ActorAt(p); i >= 0 {
			n++
			continue loop
		}
		return p
	}
}
