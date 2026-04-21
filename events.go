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

// GenCondition generates a random condition to attach to a spirit
func (g *Game) GenCondition() *SpiritCond {
	var e CondEffect
	nu := len(untilEvents)
	u := g.rand.IntN(nu)
	switch g.IntN(3) {
	case 0:
		e = &EffectCantInvoke{}
	default:
		nc := nCq - int(CqLeave)
		c := g.rand.IntN(nc)
		w := c + u + g.rand.IntN(len(eventWatchers) - nu - nc)
		e = &EffectPunish {
			When: &eventWatchers[w],
			Do:   CqLeave+Consequence(c),
		}
	}
	return &SpiritCond{ Effect: e, Until: &untilEvents[u] }
}
