package main

import (
	"fmt"
	"strings"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
)

// Event represents an occurrence that might trigger a spirit condition
type Event interface {
	Type() EventType
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
	g.WarnTypes = 0
	for _, e := range g.Entities[:NSpirits] {
		if sp, ok := e.Role.(*Spirit); ok {
			g.EventFilter |= sp.Condition.EventTypes()
			g.WarnTypes |= WarnTypes(sp.Condition.leaveWatcher())
		}
	}
}

// WarnLeave checks whether a warning is needed to interact with a comestible or menhir
func (g *Game) WarnLeave(ev EventType) bool {
	if g.confirm { // warning already issued
		g.confirm = false
		return false
	}
	if !g.CheckEvent(ev) {
		return false
	}
	for _, e := range g.Entities[:NSpirits] {
		if sp, ok := e.Role.(*Spirit); ok {
			w := sp.Condition.leaveWatcher()
			if w != nil && ev.Any(w.EventTypes()) {
				g.LogfStyled("Are you sure? Your %s spirit will leave! [y/N]", logConfirm, e.Name)
				return true
			}
		}
	}
	return false
}

const (
	EventMenhir     EventType = 1 << iota // menhir activated
	EventPortal                           // portal worked (used to record end-of-level properties)
	EventNewLevel                         // new level starts (when condition actually triggers)
	EventFakePortal                       // portal malfunctioned

	EventPickup     // picked up a comestible
	EventHeal       // player gained permanent HP
	EventTeleport   // player teleported
	EventDamage     // player takes status-based damage
	EventPlayerTurn // player's turn starts

	EventStatus    // actor status started
	EventMonsDeath // monster died
	EventMonsSeen  // you see a monster
	EventHits      // actor hits another

	EventWallBreak // wall broken
)

// WarnType classifies reasons to give comestibles a warning color in the menu
type WarnType int

func (t WarnType) Has(w WarnType) bool {
	return t&w != 0
}

const (
	WarnHeal WarnType = 1 << iota
	WarnBerserk
	WarnClarity
	WarnFoggySkin
	WarnLignification
	WarnTeleport
)

func WarnTypes(ew EventWatcher) WarnType {
	switch w := ew.(type) {
	case WatchEither:
		return WarnTypes(w.A) | WarnTypes(w.B)
	case WatchEvent:
		switch w.Type {
		case EventTeleport:
			return WarnTeleport
		case EventHeal:
			return WarnHeal
		}
	case WatchPlayerStatus:
		switch w.EvStatus {
		case StatusBerserk:
			return WarnBerserk
		case StatusClarity:
			return WarnClarity
		case StatusFoggySkin:
			return WarnFoggySkin
		case StatusLignification:
			return WarnLignification
		}
	}
	return WarnType(0)
}

// WarnComestible checks whether a specific comestible needs a warning
func (g *Game) WarnComestible(c *Comestible) bool {
	w := g.WarnTypes
	if w == 0 {
		return false
	}
	if w.Has(WarnHeal) && g.willHeal(c) {
		return true
	}
	switch c.Effect.(type) {
	case EffectBerserkingFlower:
		return w.Has(WarnBerserk)
	case EffectClarityLeaves:
		return w.Has(WarnClarity)
	case EffectFoggySkinOnion:
		return w.Has(WarnFoggySkin)
	case EffectLignificationFruit:
		return w.Has(WarnLignification)
	case EffectTeleportMushroom:
		return w.Has(WarnTeleport) && !g.PlayerActor().ResistsMove()
	}
	return false
}

func (g *Game) willHeal(c *Comestible) bool {
	pa := g.PlayerActor()
	if pa.HP == pa.GetMaxHP() {
		return false
	}
	if g.Mod(ModHealingCombat) {
		_, ok := c.Effect.(EffectAmbrosiaBerries)
		return ok
	}
	switch c.Effect.(type) {
	case EffectAmbrosiaBerries, EffectClarityLeaves, EffectFirebreathPepper, EffectFoggySkinOnion:
		return true
	case EffectTeleportMushroom:
		return pa.ResistsMove()
	default:
		return false
	}
}

// EventWatcher
type EventWatcher interface {
	EventTypes() EventType
	CheckEvent(*Game, *Spirit, Event) bool
	Desc() string
}

// EventHappened is used for an EventType with no extra information
type EventHappened struct {
	EvType EventType
}

func (ev *EventHappened) Type() EventType {
	return ev.EvType
}

type WatchEvent struct {
	Type EventType
}

func (w WatchEvent) Desc() string {
	switch w.Type {
	case EventMenhir:
		return "you activate a menhir"
	case EventNewLevel:
		return "you reach the next level"
	case EventPickup:
		return "you pick up a comestible"
	case EventFakePortal:
		return "a portal you use malfunctions"
	case EventHeal:
		return "a comestible heals your HP"
	case EventTeleport:
		return "you teleport"
	case EventDamage:
		return "you are hurt by fire or poison"
	case EventMonsDeath:
		return "a monster dies"
	case EventWallBreak:
		return "a wall breaks"
	default:
		return "(unknown)" // includes some generic events that shouldn't be used
	}
}

func (w WatchEvent) EventTypes() EventType {
	return w.Type
}

func (w WatchEvent) CheckEvent(g *Game, sp *Spirit, e Event) bool {
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

func (m MatchEither) Desc() string {
	return fmt.Sprintf("%s or %s", m.A.Desc(), m.B.Desc())
}

func (m MatchEither) Match(e *Entity) bool {
	return m.A.Match(e) || m.B.Match(e)
}

type MonsterByKind struct {
	Kind ActorKind
}

func (m MonsterByKind) Desc() string { return MonsData[m.Kind].Name }

func (m MonsterByKind) Match(e *Entity) bool {
	return e.IsActor() && e.Actor().Is(m.Kind)
}

// AnyMonster makes a matcher for any of the given names
func AnyMonster(mons ...ActorKind) EntityMatcher {
	var ms EntityMatcher = MonsterByKind{Kind: mons[0]}
	for _, m := range mons[1:] {
		ms = MatchEither{A: ms, B: MonsterByKind{Kind: m}}
	}
	return ms
}

// WatchMonsEvent triggers based on a specified event and monster type
type WatchMonsEvent struct {
	Mons   EntityMatcher
	EvType EventType
}

func (w WatchMonsEvent) Desc() string {
	switch w.EvType {
	case EventMonsDeath:
		return fmt.Sprintf("a %s dies", w.Mons.Desc())
	case EventMonsSeen:
		return fmt.Sprintf("you see a %s", w.Mons.Desc())
	default:
		panic("unexpected complete-level condition")
	}
}

func (w WatchMonsEvent) EventTypes() EventType {
	return w.EvType
}

func (w WatchMonsEvent) CheckEvent(g *Game, sp *Spirit, e Event) bool {
	d, ok := e.(*EventWithEntity)
	return ok && w.EvType == d.EvType && w.Mons.Match(d.EvEntity)
}

// WatchMonsStatus triggers when a monster begins a specific status
type WatchMonsStatus struct {
	EvStatus Status
	Mons     EntityMatcher
}

func (w WatchMonsStatus) Desc() string {
	d := w.Mons.Desc()
	if w.EvStatus == StatusFire {
		return fmt.Sprintf("a %s catches fire", d)
	}
	return fmt.Sprintf("a %s becomes %s", d, w.EvStatus)
}

func (w WatchMonsStatus) EventTypes() EventType {
	return EventStatus
}

func (w WatchMonsStatus) CheckEvent(g *Game, sp *Spirit, e Event) bool {
	d, ok := e.(*EventWithStatus)
	return ok && d.EvType == EventStatus && d.EvStatus == w.EvStatus && w.Mons.Match(d.EvEntity)
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

func (ev *EventWithStatus) onPlayer() bool {
	e := ev.EvEntity
	return e.IsActor() && e.Actor().Is(Player)
}

// WatchPlayerStatus triggers when the player begins a specific status
type WatchPlayerStatus struct {
	MatchAny bool // ignore EvStatus; allow any
	EvStatus Status
}

func (w WatchPlayerStatus) Desc() string {
	if w.MatchAny {
		return "you get any status"
	}
	if w.EvStatus == StatusFire {
		return "you catch fire"
	}
	return fmt.Sprintf("you become %s", w.EvStatus)
}

func (w WatchPlayerStatus) EventTypes() EventType {
	return EventStatus
}

func (w WatchPlayerStatus) CheckEvent(g *Game, sp *Spirit, e Event) bool {
	d, ok := e.(*EventWithStatus)
	return ok && d.EvType == EventStatus && (w.MatchAny || d.EvStatus == w.EvStatus) && d.onPlayer()
}

// WatchHits triggers when some actor hits a monster
type WatchHits struct {
	PlayerActs bool // triggers on the player, versus a monster
}

func (w WatchHits) Desc() string {
	if w.PlayerActs {
		return "you hit a monster"
	}
	return "a monster hits another"
}

func (w WatchHits) EventTypes() EventType {
	return EventHits
}

func (w WatchHits) CheckEvent(g *Game, sp *Spirit, e Event) bool {
	d, ok := e.(*EventAffects)
	return ok && !d.EvTarget.Is(Player) && w.PlayerActs == d.EvActor.Is(Player)
}

// EventAffects indicates one actor doing something to another
type EventAffects struct {
	EvType   EventType
	EvActor  *Actor
	EvTarget *Actor
}

func (ev *EventAffects) Type() EventType {
	return ev.EvType
}

// WatchEither triggers when either component does
type WatchEither struct {
	A EventWatcher
	B EventWatcher
}

func (w WatchEither) Desc() string {
	if aw, ok := w.A.(WatchPlayerStatus); ok && aw.EvStatus != StatusFire {
		if bw, ok := w.B.(WatchPlayerStatus); ok {
			return fmt.Sprintf("you become %s or %s", aw.EvStatus, bw.EvStatus)
		}
	}
	return fmt.Sprintf("%s or %s", w.A.Desc(), w.B.Desc())
}

func (w WatchEither) EventTypes() EventType {
	return w.A.EventTypes() | w.B.EventTypes()
}

func (w WatchEither) CheckEvent(g *Game, sp *Spirit, e Event) bool {
	return w.A.CheckEvent(g, sp, e) || w.B.CheckEvent(g, sp, e)
}

// WatchTwoStatus triggers when the player gets a combination of statuses
type WatchTwoStatus struct {
	A Status
	B Status
}

func (w WatchTwoStatus) Desc() string {
	return fmt.Sprintf("you become %s and %s", w.A, w.B)
}

func (w WatchTwoStatus) EventTypes() EventType {
	return EventStatus
}

func (w WatchTwoStatus) CheckEvent(g *Game, sp *Spirit, e Event) bool {
	d, ok := e.(*EventWithStatus)
	if !(ok && d.EvType == EventStatus && d.onPlayer()) {
		return false
	}
	s := g.PlayerActor().Statuses
	return (d.EvStatus == w.A && s.Has(w.B)) || (d.EvStatus == w.B && s.Has(w.A))
}

// WatchPlayerDamage triggers when fire or poison hurts the player
type WatchPlayerDamage struct {
	EvStatus Status
}

func (w WatchPlayerDamage) Desc() string {
	return fmt.Sprintf("you are hurt by %s", w.EvStatus.Name())
}

func (w WatchPlayerDamage) EventTypes() EventType {
	return EventDamage
}

func (w WatchPlayerDamage) CheckEvent(g *Game, sp *Spirit, e Event) bool {
	d, ok := e.(*EventWithStatus)
	return ok && d.EvType == EventDamage && d.EvStatus == w.EvStatus
}

// WatchMoveFar triggers when the player moves more than 1 tile between turn starts
type WatchMoveFar struct {
	PrevPos gruid.Point
}

func (w *WatchMoveFar) Desc() string {
	return "you move by more than one tile in a turn"
}

func (w *WatchMoveFar) EventTypes() EventType {
	return EventPlayerTurn | EventNewLevel
}

func (w *WatchMoveFar) CheckEvent(g *Game, sp *Spirit, e Event) bool {
	if e.Type() == EventNewLevel {
		w.PrevPos = InvalidPos
		return false
	}
	p := w.PrevPos
	n := g.Entity(PlayerID).P
	w.PrevPos = n
	return p != InvalidPos && paths.DistanceManhattan(p, n) > 1
}

// WatchReachLevel triggers when you reach a numbered level
type WatchReachLevel struct {
	Level int
}

func (w WatchReachLevel) Desc() string {
	return fmt.Sprintf("you reach level %d", w.Level)
}

func (w WatchReachLevel) EventTypes() EventType {
	return EventNewLevel
}

func (w WatchReachLevel) CheckEvent(g *Game, sp *Spirit, e Event) bool {
	return g.Map.Level == w.Level
}

// WatchCompleteLevel triggers when a matched monster dies
type WatchCompleteLevel struct {
	Test        CompleteType
	CanComplete bool
}

type CompleteType int

const (
	Charges        CompleteType = iota // still have ability charges
	EmptyInventory                     // no comestibles left
)

func (w *WatchCompleteLevel) Desc() string {
	switch w.Test {
	case Charges:
		return "you complete a full level with some charges remaining"
	case EmptyInventory:
		return "you enter a portal with no comestibles in your inventory"
	default:
		panic("unexpected complete-level condition")
	}
}

func (w *WatchCompleteLevel) EventTypes() EventType {
	return EventPortal | EventNewLevel
}

func (w *WatchCompleteLevel) CheckEvent(g *Game, sp *Spirit, e Event) bool {
	switch e.Type() {
	case EventPortal:
		switch w.Test {
		case Charges:
			if w.CanComplete {
				w.CanComplete = sp.Charges > 0
			}
		case EmptyInventory:
			w.CanComplete = len(g.ComestibleIDs()) == 0
		}
	case EventNewLevel:
		if w.CanComplete {
			return true
		}
		w.CanComplete = true
	}
	return false
}

// UpgradePolicy describes how the condition applies when used as an upgrade
type UpgradePolicy int

const (
	UpgradeIgnores   UpgradePolicy = iota // never used on upgrade
	UpgradeContinues                      // replaces an unfulfilled condition
	UpgradeReplaces                       // replaces any condition
)

func (p UpgradePolicy) Applies(existing bool) bool {
	return p >= UpgradeContinues && (existing || p >= UpgradeReplaces)
}

func (p UpgradePolicy) Desc() string {
	switch p {
	case UpgradeContinues:
		return "(Including when upgrading an unfulfilled spirit:)"
	case UpgradeReplaces:
		return "(Including on any upgrade:)"
	default:
		return ""
	}
}

// SpiritCond represents a condition imposed by a spirit
type SpiritCond struct {
	Effect  CondEffect
	Until   EventWatcher
	Upgrade UpgradePolicy
}
type CondEffect interface {
	EventTypes() EventType
	ApplyEvent(*Game, *Spirit, Event) Consequence
	Enabled() SpiritPart
	Desc() string
}

func (c *SpiritCond) EventTypes() EventType {
	if c == nil {
		return EventType(0)
	}
	return c.Effect.EventTypes() | c.Until.EventTypes()
}
func (c *SpiritCond) Desc() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "@RCurse:@N %s.\n", UpperFirst(c.Effect.Desc()))
	fmt.Fprintf(&sb, "@YUntil:@N %s.", UpperFirst(c.Until.Desc()))
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
func (c *SpiritCond) ApplyEvent(g *Game, sp *Spirit, e Event) Consequence {
	if c == nil {
		return CqNone
	}
	if c.Until.CheckEvent(g, sp, e) {
		return CqEndCondition
	}
	return c.Effect.ApplyEvent(g, sp, e)
}

// leaveWatcher returns the EventWatcher that will make the spirit leave, or nil if none
func (c *SpiritCond) leaveWatcher() EventWatcher {
	if c != nil {
		if p, ok := c.Effect.(CondPunish); ok && p.Do == CqLeave {
			return p.When
		}
	}
	return nil
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
	HasInvStats = HasInvoke | HasStats
	HasAll      = HasInvoke | HasPassives
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
			switch sp.Condition.ApplyEvent(g, sp, ev) {
			case CqEndCondition:
				g.Logf("Your %s spirit has been fulfilled!", e.Name)
				g.StoryLogf("Fulfilled the %s spirit", e.Name)
				g.curseLiftedFog()
				sp.Condition = nil
				g.ComputePlayerStats()
			case CqLeave:
				g.Logf("The %s spirit has left you!", e.Name)
				g.StoryLogf("Lost the %s spirit", e.Name)
				g.spiritLeftShock()
				g.Entities[i] = emptySlot()
				g.ComputePlayerStats()
			case CqDischargeAll:
				if sp.Charges > 0 {
					g.Logf("Your %s spirit loses its charges!", e.Name)
					sp.Charges = 0
				}
			case CqDischarge:
				sp.Charges = max(0, sp.Charges-1)
			}
		}
	}
}

// curseClouds makes clouds in an X shape around the player
func (g *Game) curseClouds(k CloudKind, dur int) {
	pp := g.PP()
	for i := -1; i <= 1; i += 2 {
		for j := -1; j <= 1; j += 2 {
			q := pp.Shift(i, j)
			if g.Map.Passable(q) {
				g.AddCloud(Cloud{Kind: k, P: q, Duration: dur})
			}
		}
	}
}

// curseLiftedFog creates fog in an X shape when a condition is fulfilled
func (g *Game) curseLiftedFog() {
	g.curseClouds(CloudNormal, 3)
}

// spiritLeftShock makes poison in an X, dazes the player, and stops time
func (g *Game) spiritLeftShock() {
	g.PutStatus(PlayerID, g.PlayerActor(), StatusDaze, 2)
	g.PutStatus(PlayerID, g.PlayerActor(), StatusTimeStop, 6)
	g.curseClouds(CloudPoison, 1)
	mp := &MapPath{passable: inMap}
	nodes := g.PR.BreadthFirstMap(mp, []gruid.Point{g.PP()}, MaxFOVRange)
	g.md.FloodAnimation(nodes, ColorViolet)
}

// CondDisabled prevents invoking the spirit while the condition is active
type CondDisabled struct {
	Part SpiritPart
}

func (cond CondDisabled) EventTypes() EventType {
	return EventType(0)
}

func (cond CondDisabled) Desc() string {
	switch cond.Part {
	case HasInvoke:
		return "This spirit cannot be invoked"
	case HasStats:
		return "This spirit doesn’t affect your Stats"
	case HasInvStats:
		return "This spirit can’t be invoked and doesn’t affect your Stats"
	case HasTraits:
		return "This spirit’s Traits don’t apply" // unused
	case HasAll:
		return "This spirit has no effect" // unused
	case HasPassives:
		return "This spirit can be invoked but has no other effects"
	default:
		panic("unexpected disabled spirit part combination")
	}
}

func (cond CondDisabled) Enabled() SpiritPart {
	return HasAll & ^cond.Part
}

func (cond CondDisabled) ApplyEvent(g *Game, sp *Spirit, e Event) Consequence {
	return CqNone
}

// CondPunish applies a consequence whenever an event triggers
type CondPunish struct {
	When EventWatcher
	Do   Consequence
}

func (c CondPunish) EventTypes() EventType {
	return c.When.EventTypes()
}

func (c CondPunish) Desc() string {
	return fmt.Sprintf("This spirit %s if %s", descCq(c.Do), c.When.Desc())
}

func (c CondPunish) Enabled() SpiritPart {
	return HasAll
}

func (c CondPunish) ApplyEvent(g *Game, sp *Spirit, e Event) Consequence {
	if c.When.CheckEvent(g, sp, e) {
		return c.Do
	}
	return CqNone
}

// OnEvent consequence based on event type
func OnEvent(cq Consequence, ty EventType) CondEffect {
	w := WatchEvent{Type: ty}
	return CondPunish{When: w, Do: cq}
}

// OnStatus consequence when the player gets a particular status
func OnStatus(cq Consequence, st Status) CondEffect {
	w := WatchPlayerStatus{EvStatus: st}
	return CondPunish{When: w, Do: cq}
}

// OnDamage consequence when the player takes damage from a status
func OnDamage(cq Consequence, st Status) CondEffect {
	w := WatchPlayerDamage{EvStatus: st}
	return CondPunish{When: w, Do: cq}
}

func Cond(ef CondEffect, u EventWatcher) *SpiritCond {
	return &SpiritCond{Effect: ef, Until: u}
}

func CondU(ef CondEffect, u EventWatcher) *SpiritCond {
	return &SpiritCond{Effect: ef, Until: u, Upgrade: UpgradeReplaces}
}

func CondC(ef CondEffect, u EventWatcher) *SpiritCond {
	return &SpiritCond{Effect: ef, Until: u, Upgrade: UpgradeContinues}
}

func Compl(t CompleteType, c Consequence, w EventWatcher) *SpiritCond {
	return CondC(CondPunish{When: w, Do: c}, &WatchCompleteLevel{Test: t})
}

func Dis(p SpiritPart, u EventWatcher) *SpiritCond {
	return Cond(CondDisabled{Part: p}, u)
}

func DisU(p SpiritPart, u EventWatcher) *SpiritCond {
	return CondU(CondDisabled{Part: p}, u)
}

func Dies(mons ...ActorKind) EventWatcher {
	return WatchMonsEvent{EvType: EventMonsDeath, Mons: AnyMonster(mons...)}
}

func Seen(mons ...ActorKind) EventWatcher {
	return WatchMonsEvent{EvType: EventMonsSeen, Mons: AnyMonster(mons...)}
}

func Gets(st Status, mons ...ActorKind) EventWatcher {
	return WatchMonsStatus{EvStatus: st, Mons: AnyMonster(mons...)}
}

// GenCondition randomly generates the sequence of conditions for a run
// 8 are generated but any that correspond to empty totems will be ignored
func (g *Game) GenConditions() []*SpiritCond {
	sp := g.ProcInfo.Spirits
	cs := make([]*SpiritCond, 8)
	noRecharges := g.Mod(ModNoRecharges)
	rqDischargeAll := CqDischargeAll
	if noRecharges {
		rqDischargeAll = CqDischarge
	}
	for i, j := range g.rand.Perm(10)[0:8] {
		aqLeave := CqLeave
		if sp[i].Advanced && !noRecharges {
			aqLeave = CqDischargeAll
		}
		switch j {
		case 0:
			cs[i] = Cond(OnStatus(aqLeave, StatusLignification), WatchEither{A: Gets(StatusFire, WalkingTree), B: Seen(WalkingMushroom)})
		case 1:
			cs[i] = Cond(OnStatus(aqLeave, StatusFoggySkin), Gets(StatusFire, FireLlama, ExplodingNadre))
		case 2:
			cs[i] = Cond(OnStatus(aqLeave, StatusClarity), Gets(StatusConfusion, BlinkButterfly))
		case 3:
			cs[i] = Cond(OnEvent(aqLeave, EventTeleport), Seen(WarpingWraith))
		case 4:
			cs[i] = Cond(OnDamage(aqLeave, StatusPoison), Dies(MadOctopode, TotemWasp))
		case 5:
			cs[i] = Cond(OnDamage(CqDischarge, StatusFire), Gets(StatusImbalance, BurningPhoenix))
		case 6:
			cs[i] = Cond(OnEvent(rqDischargeAll, EventHeal), Gets(StatusBerserk, FourHeadedHydra))
		case 7:
			cs[i] = Dis(HasStats, WatchHits{PlayerActs: false})
		case 8:
			cs[i] = Dis(HasInvoke, Seen(TotemWasp, BlazingGolem, DraggingAlligator))
		default:
			cs[i] = Dis(HasPassives, Gets(StatusBerserk, NoisyImp, CrazyDruid, WalkingMushroom))
		}
	}
	if g.rand.IntN(4) != 0 {
		var w EventWatcher
		switch g.rand.IntN(6) {
		case 0:
			w = WatchEither{A: WatchPlayerStatus{EvStatus: StatusDaze}, B: WatchPlayerStatus{EvStatus: StatusImbalance}}
		case 1:
			w = WatchEither{A: WatchPlayerStatus{EvStatus: StatusConfusion}, B: WatchPlayerStatus{EvStatus: StatusBerserk}}
		case 2:
			w = WatchEither{A: WatchPlayerStatus{EvStatus: StatusFire}, B: WatchPlayerStatus{EvStatus: StatusPoison}}
		case 3:
			w = WatchEither{A: WatchEvent{Type: EventMenhir}, B: WatchEvent{Type: EventHeal}}
		case 4:
			w = WatchEvent{Type: EventPickup}
		default:
			w = Dies(ThunderPorcupine)
		}
		c := CondPunish{When: w, Do: CqLeave}
		cs[g.rand.IntN(5)/3] = Cond(c, WatchReachLevel{Level: 4})
	}
	for _, j := range g.rand.Perm(5)[0:2] {
		i := 2 + g.IntN(4) // levels 3 to 6
		if (sp[i].Advanced && j > 2) || (noRecharges && j != 0) {
			continue // too hard
		}
		switch j {
		case 0:
			cs[i] = Compl(EmptyInventory, CqLeave, WatchEvent{Type: EventPickup})
		case 1:
			cs[i] = Compl(Charges, CqDischargeAll, &WatchMoveFar{PrevPos: InvalidPos})
		case 2:
			cs[i] = Compl(Charges, CqDischargeAll, WatchEvent{Type: EventMonsDeath})
		case 3:
			cs[i] = Compl(Charges, CqDischarge, WatchPlayerStatus{MatchAny: true})
		case 4:
			cs[i] = Compl(Charges, CqDischarge, WatchHits{PlayerActs: true})
		}
	}
	nu := 1 + g.IntN(2)
	for _, j := range g.rand.Perm(7)[0:nu] {
		i := 3 + g.IntN(5)        // levels 4 to 8
		if noRecharges && j < 2 { // single-discharge
			continue
		}
		switch j {
		case 0:
			cs[i] = CondU(OnEvent(CqDischarge, EventPickup), Gets(StatusPoison, HungryRat))
		case 1:
			cs[i] = CondU(OnEvent(CqDischarge, EventWallBreak), Gets(StatusPoison, RampagingBoar, EarthDragon))
		case 2:
			cs[i] = CondU(OnStatus(rqDischargeAll, StatusDaze), Dies(TemporalCat))
		case 3:
			cs[i] = CondU(OnStatus(rqDischargeAll, StatusFear), Gets(StatusConfusion, BarkingHound))
		case 4:
			cs[i] = CondU(OnStatus(rqDischargeAll, StatusImbalance), Gets(StatusImbalance, LashingFrog, RampagingBoar))
		case 5:
			cs[i] = DisU(HasInvStats, WatchTwoStatus{A: StatusConfusion, B: StatusImbalance})
		case 6:
			cs[i] = DisU(HasInvStats, WatchTwoStatus{A: StatusLignification, B: StatusFear})
		}
	}
	if g.IntN(3) == 0 {
		cs[7] = DisU(HasInvoke, Seen(FearsomeLich, UndeadKnight))
	}
	return cs
}

// Conditional comestibles, only usable if conditions are met
type ConditionalComestible interface {
	CanApply(*Game, ID) bool
}

type EffectSlipperyPea struct{}

func (eff EffectSlipperyPea) CanApply(g *Game, i ID) bool {
	count := 0
	for _, turns := range g.PlayerActor().Statuses {
		if turns > 0 {
			count++
		}
	}
	return count >= 2
}

func (eff EffectSlipperyPea) Apply(g *Game) bool {
	pa := g.PlayerActor()
	g.eatMsg("slippery pea", "")
	for i := range pa.Statuses {
		pa.Statuses[i] = 0
	}
	pa.HP = min(pa.HP, pa.GetMaxHP())
	g.PutStatus1(PlayerID, pa, StatusClarity, DurationSprint)
	g.PutStatus1(PlayerID, pa, StatusSprint, DurationSprint)
	return true
}

func (eff EffectSlipperyPea) Desc(g *Game) string {
	return fmt.Sprintf("Eat only if you currently have at least 2 statuses. Removes all your statuses without the usual effects, then makes you @BClear-Minded@N and @BSprinting@N and for %d turns.",
		DurationSprint)
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
	pa := g.PlayerActor()
	g.eatMsg("fortress oyster", "")
	g.PutStatus1(PlayerID, pa, StatusLignification, 5)
	g.PutStatus1(PlayerID, pa, StatusBerserk, 5)
	g.PutStatus1(PlayerID, pa, StatusVampirism, 5)
	return true
}

func (eff EffectFortressOyster) Desc(g *Game) string {
	return "Eat only if you are next to at least 3 monsters. Makes you @BLignified@N, @BBerserk@N, and @BVampiric@N for 5 turns."
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
	g.eatMsg("runethorn rose", "")
	pp := g.PP()
	var t MagicRune // type of rune
	if r, ok := g.anyRuneAt(pp); ok {
		t = r.Rune
	} else {
		t = MagicRune(g.IntN(NRunes))
	}
	for i := -1; i <= 1; i += 2 {
		for j := -1; j <= 1; j += 2 {
			q := pp
			cont := true
			for range 2 {
				q = q.Shift(i, j)
				cont = cont && g.validRunePlace(q)
				if !cont {
					size := g.Map.Terrain.Size()
					for !g.validRunePlace(q) {
						q = gruid.Point{g.rand.IntN(size.X), g.rand.IntN(size.Y)}
					}
				}
				i, _ := g.genRunicTrapAt(t, q)
				g.SenseEntity(i, "sense")
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
	return "Eat only if you standing on a runic trap. Create 8 runes of its type: up to two at each diagonal, or randomly around the map."
}

type EffectVanishingSnail struct{}

func (eff EffectVanishingSnail) CanApply(g *Game, i ID) bool {
	count := 0
	for _, ei := range g.NPMapEntities() {
		ea, ok := ei.Role.(*Actor)
		if ok && !ea.IsDead() && ea.Behavior.State == Hunting {
			count++
		}
	}
	return count >= 4
}

func (eff EffectVanishingSnail) Apply(g *Game) bool {
	g.eatMsg("vanishing snail", "")
	for _, ei := range g.NPMapEntities() {
		a, ok := ei.Role.(*Actor)
		if ok && a.Behavior.State == Hunting {
			a.Behavior.State = Wandering
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
	g.PutStatus1(PlayerID, pa, StatusShadow, 10)
	return true
}

func (eff EffectVanishingSnail) Desc(g *Game) string {
	return "Eat only if at least 4 monsters are hunting you. All monsters stop hunting and any that are next to you blink away. Become @BShadowy@N for 10 turns."
}

type EffectTunnellingAcorn struct{}

func (eff EffectTunnellingAcorn) CanApply(g *Game, i ID) bool {
	pp := g.PP()
	n := pp.Add(g.Dir)
	return inMap(n) && !g.Map.Passable(n) && g.Map.AdjacentNonPassableCount(pp) >= 3
}

func (eff EffectTunnellingAcorn) Apply(g *Game) bool {
	g.eatMsg("tunnelling acorn", "")
	p := g.PP().Add(g.Dir)
	for inMap(p) && !g.Map.Passable(p) {
		g.BreakWall(p)
		p = p.Add(g.Dir)
	}
	return true
}

func (eff EffectTunnellingAcorn) Desc(g *Game) string {
	return "Eat only if you are next to 3 walls, and facing one of them. Silently break all walls in front of you, until reaching a non-wall."
}

type EffectWarpingBean struct{}

func (g *Game) nearestBeanTarget(allowNone bool) gruid.Point {
	p := g.PP().Add(g.Dir)
	for inMap(p) && g.InFOV(p) {
		if i, it := g.ItemAt(p); i >= 0 {
			switch it.(type) {
			case *Spirit, *EmptyTotem, *Menhir, *Portal, *CorruptionOrb:
				if i, ai := g.ActorAt(p); i < 0 || !ai.ResistsMove() {
					return p
				}
			}
		}
		p = p.Add(g.Dir)
	}
	if allowNone {
		return p.Sub(g.Dir)
	}
	return InvalidPos
}

func (eff EffectWarpingBean) CanApply(g *Game, i ID) bool {
	if g.PlayerActor().ResistsMove() {
		return false
	}
	return g.nearestBeanTarget(false) != InvalidPos
}

func (eff EffectWarpingBean) Apply(g *Game) bool {
	g.eatMsg("warping bean", "")
	to := g.nearestBeanTarget(true)
	for p := g.PP().Add(g.Dir); p != to; p = p.Add(g.Dir) {
		if i, ai := g.ActorAt(p); i >= 0 {
			g.TeleportActor(i, ai, 0)
		}
	}
	g.MoveActor(PlayerID, g.PlayerActor(), to, MovTeleport)
	return true
}

func (eff EffectWarpingBean) Desc(g *Game) string {
	return "Eat only if you see a totem, menhir, portal, or orb directly in front of you and not blocked by a woody monster. Warp to the nearest one, teleporting away monsters on the way to it."
}

type EffectSeeingPotato struct{}

func (eff EffectSeeingPotato) CanApply(g *Game, i ID) bool {
	j, _ := g.ItemAt(g.PP())
	return i >= 0 && i == j
}

func (eff EffectSeeingPotato) Apply(g *Game) bool {
	g.eatMsg("seeing potato’s stalk", "")
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
			g.SenseEntity(i, "sense")
		}
	}
	return true
}

func (eff EffectSeeingPotato) Desc(g *Game) string {
	return "Eat only from the ground and not your inventory. Sense all floor tiles within 16 spaces, and anything on those tiles."
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
	g.eatMsg("tricky chestnut", "")
	pp := g.PP()
	var tot *Spirit
	if i, it := g.ItemAt(pp); i >= 0 {
		switch o := it.(type) {
		case *Portal:
			o.Fake = false
			o.Used = false
			g.LogStyled("The portal is fixed!", logSpecial)
			return true
		case *Spirit:
			tot = o
		}
	}
	// Recharge spirits by default
	for _, e := range g.Entities[:NSpirits] {
		if sp, ok := e.Role.(*Spirit); ok {
			sp.Charges = sp.MaxCharges[sp.Level]
		}
	}
	g.LogStyled("Your spirits are recharged!", logSpecial)
	if tot != nil { // turn totem to empty if present
		g.Entity(g.TotemID()).P = InvalidPos
		et := g.Entity(g.genEmptyTotemAt(pp))
		et.Seen = true
	}
	return true
}

func (eff EffectTrickyChestnut) Desc(g *Game) string {
	return "Eat if you are standing on a broken portal, to make it work again.\nEat if you are standing on an unused totem, to empty the totem, and restore all your spirits to full charges."
}

func CantEatMsg(ef Effect) string {
	switch ef.(type) {
	case EffectSlipperyPea, EffectVanishingSnail, EffectWarpingBean:
		return "You can’t grab it!"
	case EffectFortressOyster, EffectTunnellingAcorn, EffectTrickyChestnut:
		return "It won’t open!"
	case EffectRunethornRose:
		return "You can’t get past its thorns!"
	case EffectSeeingPotato:
		return "It has to be planted in the ground!"
	default:
		panic("unexpected inedible comestible")
	}
}

var CondComs = []itemInfo{
	{"slippery pea", EffectSlipperyPea{}},
	{"fortress oyster", EffectFortressOyster{}},
	{"runethorn rose", EffectRunethornRose{}},
	{"vanishing snail", EffectVanishingSnail{}},
	{"tunnelling acorn", EffectTunnellingAcorn{}},
	{"warping bean", EffectWarpingBean{}},
	{"seeing potato", EffectSeeingPotato{}},
	{"tricky chestnut", EffectTrickyChestnut{}},
}

func (g *Game) GetConditionalComestible(level int) itemInfo {
	n := len(CondComs)
	if level == 9 {
		n-- // chestnuts are useless on the last level
	}
	return CondComs[g.IntN(n)]
}
