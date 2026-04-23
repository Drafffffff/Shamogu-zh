package main

import (
	"fmt"
	"strings"
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
		if sp, ok := e.Role.(*Spirit); ok && sp.Condition != nil {
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
	EventBeginStatus                  // player status started

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
	EvDesc  string
}

func (w *WatchEvent) Desc() string {
	return w.EvDesc
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

type WatchDeath struct {
	MonsName string
}

func (w *WatchDeath) Desc() string {
	return fmt.Sprintf("A %s dies", w.MonsName)
}

func (w *WatchDeath) EventTypes() EventType {
	return EventMonsDeath
}

func (w *WatchDeath) CheckEvent(e Event) bool {
	d, ok := e.(*EventWithEntity)
	return ok && d.EvType == EventMonsDeath && d.EvEntity.Name == w.MonsName
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
	return ok && d.EvType == EventBeginStatus && d.EvStatus == w.EvStatus
}

var eventWatchers = []WatchEvent{
	{Type: EventMonsDeath,   EvDesc:"a monster dies"},
	{Type: EventPickup,      EvDesc:"you pick up a comestible"},
	{Type: EventInvoke,      EvDesc:"you invoke a spirit"},
	{Type: EventMonsRune,    EvDesc:"a monster triggers a rune"},
	{Type: EventTeleport,    EvDesc:"you teleport"},
	{Type: EventComestible,  EvDesc:"you eat a comestible"},
	{Type: EventWallBreak,   EvDesc:"a wall breaks"},
	{Type: EventHPGain,      EvDesc:"you gain HP"},
	{Type: EventMenhir,      EvDesc:"you activate a Menhir"},
	{Type: EventTotem,       EvDesc:"you use a totem"},
	{Type: EventNewLevel,    EvDesc:"you reach the next level"},
	{Type: EventPlayRune,    EvDesc:"you trigger a rune"},
}
var untilEvents = []WatchEvent{
	{Type: EventWallBreak,   EvDesc:"a wall breaks"},
	{Type: EventHPGain,      EvDesc:"you gain HP"},
	{Type: EventTeleport,    EvDesc:"you teleport"},
	{Type: EventFakePortal,  EvDesc:"a portal you use malfunctions"},
}

// SpiritCond represents a condition imposed by a spirit
type SpiritCond struct {
	Effect  CondEffect
	Until   EventWatcher
}
type CondEffect interface {
	EventTypes()  EventType
	ApplyEvent(*Spirit, Event) Consequence
	CanInvoke()   bool
	Desc()  string
}

func (c *SpiritCond) EventTypes() EventType {
	return c.Effect.EventTypes() | c.Until.EventTypes()
}
func (c *SpiritCond) Desc() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "@RCurse:@N %s.", c.Effect.Desc())
	sb.WriteByte('\n')
	fmt.Fprintf(&sb, "@YUntil:@N %s.", c.Until.Desc())
	return sb.String()
}
func (c *SpiritCond) CanInvoke() bool {
	return c.Effect.CanInvoke()
}
func (c *SpiritCond) ApplyEvent(sp *Spirit, e Event) Consequence {
	if c.Until.CheckEvent(e) {
		return CqEndCondition
	}
	return c.Effect.ApplyEvent(sp, e)
}

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
		if sp, ok := e.Role.(*Spirit); ok && sp.Condition != nil {
			switch sp.Condition.ApplyEvent(sp, ev) {
			case CqEndCondition:
				g.Logf("Your %s spirit has been fulfilled!", e.Name)
				sp.Condition = nil
				g.UpdateEventFilter()
			case CqLeave:
				g.Logf("The %s spirit has left you!", e.Name)
				g.StoryLogf("Lost the %s spirit", e.Name)
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

// EffectCantInvoke prevents invoking the spirit while the condition is active
type EffectCantInvoke struct {}

func (cond *EffectCantInvoke) EventTypes() EventType {
	return EventType(0)
}

func (cond *EffectCantInvoke) Desc() string {
	return "This spirit cannot be invoked"
}

func (cond *EffectCantInvoke) CanInvoke() bool {
	return false
}

func (cond *EffectCantInvoke) ApplyEvent(sp *Spirit, e Event) Consequence {
	return CqNone
}

// EffectPunish applies a consequence whenever an event triggers
type EffectPunish struct {
	When EventWatcher
	Do   Consequence
}

func (c *EffectPunish) EventTypes() EventType {
	return c.When.EventTypes()
}

func (c *EffectPunish) Desc() string {
	return fmt.Sprintf("This spirit %s if %s", descCq(c.Do), c.When.Desc())
}

func (c *EffectPunish) CanInvoke() bool {
	return true
}

func (c *EffectPunish) ApplyEvent(sp *Spirit, e Event) Consequence {
	if c.When.CheckEvent(e) {
		return c.Do
	}
	return CqNone
}

// OnStatus consequence when the player gets a particular status
func OnStatus(cq Consequence, st Status) CondEffect {
	w := &WatchPlayerStatus{EvStatus: st}
	return &EffectPunish{When: w, Do: cq}
}

// UntilDeath condition that ends when a particular monster type dies
func UntilDeath(mons string, ef CondEffect) *SpiritCond {
	u := &WatchDeath{MonsName: mons}
	return &SpiritCond{Effect: ef, Until: u}
}

// GenCondition generates a random condition to attach to a spirit
func (g *Game) GenCondition() *SpiritCond {
	var e CondEffect
	nu := len(untilEvents)
	u := g.rand.IntN(nu)
	switch g.IntN(8) {
	case 0:
		e = &EffectCantInvoke{}
	case 1, 2:
		nc := nCq - int(CqLeave)
		c := g.rand.IntN(nc)
		w := c + u + g.rand.IntN(len(eventWatchers) - nu - nc)
		e = &EffectPunish {
			When: &eventWatchers[w],
			Do:   CqLeave+Consequence(c),
		}
	default:
		switch g.IntN(9) {
		case 0: return UntilDeath("walking tree",    OnStatus(CqLeave, StatusLignification))
		case 1: return UntilDeath("barking hound",   OnStatus(CqLeave, StatusBerserk))
		case 2: return UntilDeath("fire llama",      OnStatus(CqLeave, StatusFoggySkin))
		case 3: return UntilDeath("burning phoenix", OnStatus(CqDischarge, StatusFire)) // or on HP loss from burning?
		case 4: return UntilDeath("temporal cat",    OnStatus(CqDischargeAll, StatusDaze))
		case 5: return UntilDeath("lashing frog",    OnStatus(CqDischargeAll, StatusImbalance))
		case 6: return UntilDeath("barking hound",   OnStatus(CqDischargeAll, StatusFear))
		case 7: return UntilDeath("mad octopode",    OnStatus(CqDischargeAll, StatusPoison)) // totem wasp?
		case 8: return UntilDeath("warping wraith", &EffectPunish{When: &WatchEvent{Type: EventTeleport, EvDesc:"you teleport"}, Do: CqLeave})
		}
	}
	return &SpiritCond{ Effect: e, Until: &untilEvents[u] }
}
