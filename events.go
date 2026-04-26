package main

import (
	"fmt"
	"strings"

	"codeberg.org/anaseto/gruid"
)

// Event represents an occurrence that might trigger a spirit condition
type Event interface {
	Type()  EventType
}

// EventType categorizes condition events
type EventType uint64

func (t EventType) Any(of EventType) bool {
	return t&of != 0
}

func (g *Game) CheckEvent(ev EventType) bool {
	return ev.Any(g.EventFilter)
}

// UpdateEventFilter re-computes which event types affect current spirits
func (g *Game) UpdateEventFilter() {
	g.EventFilter = 0
	for _, e := range g.Entities[:NSpirits] {
		if sp, ok := e.Role.(*Spirit); ok {
			g.EventFilter |= sp.Condition.EventTypes()
		}
	}
}

const (
	EventMenhir EventType = 1 << iota // menhir activated
	EventTotem                        // totem used
	EventFakePortal                   // portal malfunctioned
	EventNewLevel                     // portal worked
	EventPlayRune                     // player triggered a rune
	EventMonsRune                     // monster triggered a rune

	EventPickup                       // picked up a comestible
	EventComestible                   // ate a comestible
	EventInvoke                       // invoked a spirit
	EventHPGain                       // player gained HP
	EventTeleport                     // player teleported
	EventDamage                       // player takes staus-based damage

	EventBeginStatus                  // actor status started
	EventMonsDeath                    // monster died

	EventWallBreak                    // wall broken
)

// EventWatcher
type EventWatcher interface {
	EventTypes()  EventType
	CheckEvent(Event) bool
	Desc()  string
}

// EventHappened is used for an EventType with no extra information
type EventHappened struct {
	EvType EventType
}

func (ev *EventHappened) Type() EventType {
	return ev.EvType
}

type WatchEvent struct {
	Type  EventType
}

func (w *WatchEvent) Desc() string {
	switch w.Type {
	case EventMenhir:      return "you activate a Menhir"
	case EventTotem:       return "you use a totem"
	case EventFakePortal:  return "a portal you use malfunctions"
	case EventNewLevel:    return "you reach the next level"
	case EventPlayRune:    return "you trigger a rune"
	case EventMonsRune:    return "a monster triggers a rune"
	case EventPickup:      return "you pick up a comestible"
	case EventComestible:  return "you eat a comestible"
	case EventInvoke:      return "you invoke a spirit"
	case EventHPGain:      return "you gain HP"
	case EventTeleport:    return "you teleport"
	case EventDamage:      return "you are hurt by fire or poison"
	case EventMonsDeath:   return "a monster dies"
	case EventWallBreak:   return "a wall breaks"
	default: return "(unknown)" // includes some generic events that shouldn't be used
	}
}

func (w *WatchEvent) EventTypes() EventType {
	return w.Type
}

func (w *WatchEvent) CheckEvent(e Event) bool {
	return e.Type().Any(w.Type)
}

// EventWithEntity carries the entity relevant to the event
type EventWithEntity struct {
	EvType   EventType
	EvEntity *Entity
}

func (ev *EventWithEntity) Type() EventType {
	return ev.EvType
}

// EntityMatcher filters entities
type EntityMatcher interface {
	Match(*Entity) bool
	Desc() string
}

type MatchEither struct {
	A EntityMatcher
	B EntityMatcher
}

func (m *MatchEither) Desc() string {
	return fmt.Sprintf("%s or %s", m.A.Desc(), m.B.Desc())
}

func (m *MatchEither) Match(e *Entity) bool {
	return m.A.Match(e) || m.B.Match(e)
}

type MonsterByName struct {
	MonsName string
}

func (m *MonsterByName) Desc() string { return m.MonsName }

func (m *MonsterByName) Match(e *Entity) bool {
	return e.Name == m.MonsName
}

// AnyMonster makes a matcher for any of the given names
func AnyMonster(mons ...string) EntityMatcher {
	var ms EntityMatcher = &MonsterByName{MonsName: mons[0]}
	for _, m := range mons[1:] {
		ms = &MatchEither{A: ms, B: &MonsterByName{MonsName: m}}
	}
	return ms
}

// WatchDeath triggers when a matched monster dies
type WatchDeath struct {
	Mons EntityMatcher
}

func (w *WatchDeath) Desc() string {
	return fmt.Sprintf("A %s dies", w.Mons.Desc())
}

func (w *WatchDeath) EventTypes() EventType {
	return EventMonsDeath
}

func (w *WatchDeath) CheckEvent(e Event) bool {
	d, ok := e.(*EventWithEntity)
	return ok && d.EvType == EventMonsDeath && w.Mons.Match(d.EvEntity)
}

// WatchMonsStatus triggers when a monster begins a specific status
type WatchMonsStatus struct {
	EvStatus Status
	Mons     EntityMatcher
}

func (w *WatchMonsStatus) Desc() string {
	return fmt.Sprintf("A %s becomes %s", w.Mons.Desc(), w.EvStatus)
}

func (w *WatchMonsStatus) EventTypes() EventType {
	return EventBeginStatus
}

func (w *WatchMonsStatus) CheckEvent(e Event) bool {
	d, ok := e.(*EventWithStatus)
	return ok && d.EvType == EventBeginStatus && d.EvStatus == w.EvStatus && w.Mons.Match(d.EvEntity)
}

// EventWithStatus marks status changes with the entity and status
type EventWithStatus struct {
	EvType   EventType
	EvEntity *Entity
	EvStatus Status
}

func (ev *EventWithStatus) Type() EventType {
	return ev.EvType
}

// WatchPlayerStatus triggers when the player begins a specific status
type WatchPlayerStatus struct {
	EvStatus Status
}

func (w *WatchPlayerStatus) Desc() string {
	return fmt.Sprintf("you become %s", w.EvStatus)
}

func (w *WatchPlayerStatus) EventTypes() EventType {
	return EventBeginStatus
}

func (w *WatchPlayerStatus) CheckEvent(e Event) bool {
	d, ok := e.(*EventWithStatus)
	return ok && d.EvType == EventBeginStatus && d.EvStatus == w.EvStatus && d.EvEntity.IsActor() && d.EvEntity.Actor().Is(Player)
}

// WatchPlayerDamage triggers when fire or poison hurts the player
type WatchPlayerDamage struct {
	EvStatus Status
}

func (w *WatchPlayerDamage) Desc() string {
	return fmt.Sprintf("you are hurt by %s", w.EvStatus.Name())
}

func (w *WatchPlayerDamage) EventTypes() EventType {
	return EventDamage
}

func (w *WatchPlayerDamage) CheckEvent(e Event) bool {
	d, ok := e.(*EventWithStatus)
	return ok && d.EvType == EventDamage && d.EvStatus == w.EvStatus
}

var watchTypes = []EventType{
	EventMonsDeath,
	EventPickup,
	EventInvoke,
	EventMonsRune,
	EventTeleport,
	EventComestible,
	EventWallBreak,
	EventHPGain,
	EventMenhir,
	EventTotem,
	EventNewLevel,
	EventPlayRune,
}
var untilTypes = []EventType{
	EventWallBreak,
	EventHPGain,
	EventTeleport,
	EventFakePortal,
}

// SpiritCond represents a condition imposed by a spirit
type SpiritCond struct {
	Effect  CondEffect
	Until   EventWatcher
}
type CondEffect interface {
	EventTypes()                EventType
	ApplyEvent(*Spirit, Event)  Consequence
	Enabled()                   SpiritPart
	Desc()                      string
}

func (c *SpiritCond) EventTypes() EventType {
	if c == nil {
		return EventType(0)
	}
	return c.Effect.EventTypes() | c.Until.EventTypes()
}
func (c *SpiritCond) Desc() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "@RCurse:@N %s.", c.Effect.Desc())
	sb.WriteByte('\n')
	fmt.Fprintf(&sb, "@YUntil:@N %s.", c.Until.Desc())
	return sb.String()
}
func (c *SpiritCond) Enabled() SpiritPart {
	if c == nil {
		return HasAll
	}
	return c.Effect.Enabled()
}
func (c *SpiritCond) CanInvoke() bool {
	return c.Enabled().Any(HasInvoke)
}
func (c *SpiritCond) ApplyEvent(sp *Spirit, e Event) Consequence {
	if c == nil {
		return CqNone
	}
	if c.Until.CheckEvent(e) {
		return CqEndCondition
	}
	return c.Effect.ApplyEvent(sp, e)
}

// SpiritPart allows independently disabling part of a spirit
type SpiritPart int

func (t SpiritPart) Any(of SpiritPart) bool {
	return t&of != 0
}

const (
	HasInvoke SpiritPart = 1 << iota
	HasStats
	HasTraits
	HasPassives = HasStats | HasTraits
	HasAll = HasInvoke | HasPassives
)

// Consequence represents a result of a condition being met
type Consequence int

const (
	CqNone Consequence = iota
	CqEndCondition
	CqLeave
	CqDischargeAll
	CqDischarge
)

const nCq = 1 + int(CqDischarge)

func descCq(c Consequence) string {
	switch c {
	case CqLeave:
		return "leaves"
	case CqDischargeAll:
		return "loses all ability charges"
	case CqDischarge:
		return "loses an ability charge"
	default:
		panic("unexpected condition consequence")
	}
}

func TotemEvents(g *Game, ev Event) {
	ty := ev.Type()
	if !g.CheckEvent(ty) {
		return
	}
	for i, e := range g.Entities[:NSpirits] {
		if sp, ok := e.Role.(*Spirit); ok {
			switch sp.Condition.ApplyEvent(sp, ev) {
			case CqEndCondition:
				g.Logf("Your %s spirit has been fulfilled!", e.Name)
				g.StoryLogf("Fulfilled the %s spirit", e.Name)
				g.curseLiftedFog()
				sp.Condition = nil
				g.ComputePlayerStats()
			case CqLeave:
				g.Logf("The %s spirit has left you!", e.Name)
				g.StoryLogf("Lost the %s spirit", e.Name)
				g.spiritLeftDaze()
				g.Entities[i] = emptySlot()
				g.ComputePlayerStats()
			case CqDischargeAll:
				if sp.Charges > 0 {
					g.Logf("Your %s spirit loses its charges!", e.Name)
					sp.Charges = 0
				}
			case CqDischarge:
				sp.Charges = max(0, sp.Charges - 1)
			}
		}
	}
}

// curseLiftedFog creates fog in an X shape when a condition is fulfilled
func (g *Game) curseLiftedFog() {
	pp := g.PP()
	for i := -1; i <= 1; i += 2 {
		for j := -1; j <= 1; j += 2 {
			q := pp.Shift(i, j)
			if g.Map.Passable(q) {
				g.NormalCloudAt(q, 3)
			}
		}
	}
}

// spiritLeftDaze dazes player and monsters in view when a spirit leaves
func (g *Game) spiritLeftDaze() {
	g.PutStatus(PlayerID, g.PlayerActor(), StatusDaze, 2)
	for i, ai := range g.Monsters() {
		ei := g.Entity(i)
		if !g.InFOV(ei.P) {
			continue
		}
		g.PutStatus(i, ai, StatusDaze, 6)
	}
	mp := &MapPath{passable: inMap}
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{g.PP()}, MaxFOVRange)
	g.md.FloodAnimation(nodes, ColorViolet)
}

// CondDisabled prevents invoking the spirit while the condition is active
type CondDisabled struct {
	Part SpiritPart
}

func (cond *CondDisabled) EventTypes() EventType {
	return EventType(0)
}

func (cond *CondDisabled) Desc() string {
	switch cond.Part {
	case HasInvoke:
		return "This spirit cannot be invoked"
	case HasStats:
		return "This spirit doesn't affect your Attack, Defense, or max HP"
	case HasTraits:
		return "This spirit's special traits don't apply"
	case HasAll:
		return "This spirit has no effect"
	case HasPassives:
		return "This spirit can be invoked but has no other effects"
	default:
		panic("unexpected disabled spirit part combination")
	}
}

func (cond *CondDisabled) Enabled() SpiritPart {
	return HasAll & ^cond.Part
}

func (cond *CondDisabled) ApplyEvent(sp *Spirit, e Event) Consequence {
	return CqNone
}

// CondPunish applies a consequence whenever an event triggers
type CondPunish struct {
	When EventWatcher
	Do   Consequence
}

func (c *CondPunish) EventTypes() EventType {
	return c.When.EventTypes()
}

func (c *CondPunish) Desc() string {
	return fmt.Sprintf("This spirit %s if %s", descCq(c.Do), c.When.Desc())
}

func (c *CondPunish) Enabled() SpiritPart {
	return HasAll
}

func (c *CondPunish) ApplyEvent(sp *Spirit, e Event) Consequence {
	if c.When.CheckEvent(e) {
		return c.Do
	}
	return CqNone
}

// OnEvent consequence based on event type
func OnEvent(cq Consequence, ty EventType) CondEffect {
	w := &WatchEvent{Type: ty}
	return &CondPunish{When: w, Do: cq}
}

// OnStatus consequence when the player gets a particular status
func OnStatus(cq Consequence, st Status) CondEffect {
	w := &WatchPlayerStatus{EvStatus: st}
	return &CondPunish{When: w, Do: cq}
}

// OnDamage consequence when the player takes damage from a status
func OnDamage(cq Consequence, st Status) CondEffect {
	w := &WatchPlayerDamage{EvStatus: st}
	return &CondPunish{When: w, Do: cq}
}

func Cond(ef CondEffect, u EventWatcher) *SpiritCond {
	return &SpiritCond{Effect: ef, Until: u}
}

func Dis(p SpiritPart, u EventWatcher) *SpiritCond {
	return Cond(&CondDisabled{Part: p}, u)
}

func Dies(mons ...string) EventWatcher {
	return &WatchDeath{Mons: AnyMonster(mons...)}
}

func Gets(st Status, mons ...string) EventWatcher {
	return &WatchMonsStatus{EvStatus: st, Mons: AnyMonster(mons...)}
}

// GenCondition generates a random condition to attach to a spirit
func (g *Game) GenCondition() *SpiritCond {
	switch g.IntN(25) {
	case  0: return Cond(OnStatus(CqLeave, StatusLignification),    Gets(StatusFire, "walking tree"))
	case  1: return Cond(OnStatus(CqLeave, StatusBerserk),          Dies("barking hound"))
	case  2: return Cond(OnStatus(CqLeave, StatusFoggySkin),        Dies("fire llama"))
	case  3: return Cond(OnEvent (CqLeave, EventTeleport),          Dies("warping wraith"))
	case  4: return Cond(OnStatus(CqDischargeAll, StatusDaze),      Gets(StatusDaze, "temporal cat"))
	case  5: return Cond(OnStatus(CqDischargeAll, StatusImbalance), Dies("lashing frog"))
	case  6: return Cond(OnStatus(CqDischargeAll, StatusFear),      Gets(StatusConfusion, "barking hound"))
	case  7: return Cond(OnEvent (CqDischargeAll, EventHPGain),     Gets(StatusBerserk, "four-headed hydra"))
	case  8: return Cond(OnEvent (CqDischargeAll, EventWallBreak),  Gets(StatusPoison, "rampaging boar", "earth dragon"))
	case  9: return Cond(OnEvent (CqDischarge, EventPickup),        Gets(StatusPoison, "hungry rat"))
	case 10: return Cond(OnDamage(CqDischarge, StatusFire),         Gets(StatusImbalance, "burning phoenix"))
	case 11: return Cond(OnDamage(CqDischarge, StatusPoison),       Dies("mad octopode", "totem wasp"))
	case 12: return Cond(OnDamage(CqDischarge, StatusPoison),       Gets(StatusLignification, "venomous viper"))
	case 13: return Cond(OnStatus(CqDischarge, StatusConfusion),    Gets(StatusFear, "confusing eye"))
	case 14: return Dis(HasInvoke,   Gets(StatusLignification, "thunder porcupine", "berserking spider"))
	case 15: return Dis(HasStats,    Gets(StatusImbalance, "lashing frog", "rampaging boar"))
	case 16: return Dis(HasInvoke,   Gets(StatusFire,    "fire llama", "exploding nadre"))
	case 17: return Dis(HasPassives, Gets(StatusBerserk, "berserking spider", "blinking butterfly"))
	case 18: return Dis(HasInvoke,   Gets(StatusBerserk, "noisy imp", "crazy druid", "walking mushroom"))
	default:
		nu := len(untilTypes)
		u := g.rand.IntN(nu)
		nc := nCq - int(CqLeave)
		c := g.rand.IntN(nc)
		w := c + u + g.rand.IntN(len(watchTypes) - nu - nc)
		e := &CondPunish {
			When: &WatchEvent{Type: watchTypes[w]},
			Do:   CqLeave+Consequence(c),
		}
		return &SpiritCond{ Effect: e, Until: &WatchEvent{Type: untilTypes[u]} }
	}
}
