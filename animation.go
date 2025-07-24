package main

import (
	"time"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
)

// These constants represent various common durations for animation frames.
const (
	AnimDurShort       = 40 * time.Millisecond
	AnimDurShortMedium = 60 * time.Millisecond
	AnimDurMedium      = 80 * time.Millisecond
	AnimDurMediumLong  = 100 * time.Millisecond
	AnimDurLong        = 150 * time.Millisecond
)

// Animations provides support for animations.
type Animations struct {
	frames []AnimFrame // frame queue
	grid   gruid.Grid  // drawing grid
	pgrid  gruid.Grid  // previous drawing grid
	idx    int         // animation sequence counter
	draw   bool        // whether to draw an animation frame or not
}

// AnimFrame represents an animation frame.
type AnimFrame struct {
	Cells    []gruid.FrameCell
	Duration time.Duration
}

// Finish clears the frame queue.
func (a *Animations) Finish() {
	a.idx++
	a.frames = nil
}

// Done reports whether there are no more animations in the queue.
func (a *Animations) Done() bool {
	return len(a.frames) == 0
}

// Draw sets a given rune and foreground color at a given position.
func (a *Animations) Draw(p gruid.Point, r rune, fg gruid.Color) {
	c := a.grid.At(p)
	c.Rune = r
	c.Style.Fg = fg
	a.grid.Set(p, c)
}

// DrawWith sets a given rune and foreground and background color and
// attributes at a given position.
func (a *Animations) DrawWith(p gruid.Point, r rune, fg, bg gruid.Color, attrs gruid.AttrMask) {
	c := a.grid.At(p)
	c.Rune = r
	c.Style.Fg = fg
	c.Style.Bg = bg
	c.Style.Attrs = attrs
	a.grid.Set(p, c)
}

// DrawRune sets a given rune at a given position.
func (a *Animations) DrawRune(p gruid.Point, r rune) {
	c := a.grid.At(p)
	c.Rune = r
	a.grid.Set(p, c)
}

// DrawColor sets a given color at a given position.
func (a *Animations) DrawColor(p gruid.Point, fg gruid.Color) {
	c := a.grid.At(p)
	c.Style.Fg = fg
	a.grid.Set(p, c)
}

// DrawReverse sets a given rune and foreground color with reverse attribute.
func (a *Animations) DrawReverse(p gruid.Point, r rune, fg gruid.Color) {
	c := a.grid.At(p)
	c.Rune = r
	c.Style.Fg = fg
	c.Style.Attrs |= AttrReverse
	a.grid.Set(p, c)
}

// Frame emits a new animation frame with given duration.
func (a *Animations) Frame(d time.Duration) {
	frame := AnimFrame{}
	frame.Duration = d
	it := a.grid.Iterator()
	itp := a.pgrid.Iterator()
	for it.Next() && itp.Next() {
		if it.Cell() == itp.Cell() {
			continue
		}
		frame.Cells = append(frame.Cells, gruid.FrameCell{P: it.P(), Cell: it.Cell()})
	}
	a.frames = append(a.frames, frame)
	a.pgrid.Copy(a.grid)
}

type msgAnim int

// InitAnimations initializes drawing grids for animations.
func (md *model) InitAnimations() {
	gd := md.gd.Slice(md.gd.Range().Shift(0, 2, 0, -1))
	max := gd.Size()
	md.anims.grid = gruid.NewGrid(max.X, max.Y)
	md.anims.pgrid = gruid.NewGrid(max.X, max.Y)
	md.anims.grid.Copy(gd)
	md.anims.pgrid.Copy(gd)
}

// AnimNext returns a ticker command appropriately timed for the next frame.
func (md *model) AnimNext() gruid.Cmd {
	d := md.anims.frames[0].Duration
	idx := md.anims.idx
	return func() gruid.Msg {
		t := time.NewTimer(d)
		<-t.C
		return msgAnim(idx)
	}
}

// AnimCmd returns a command with current animation sequence counter, which is
// used when updating to check whether the current animation is still relevant
// or not.
func (md *model) AnimCmd() gruid.Cmd {
	if len(md.anims.frames) == 0 {
		return nil
	}
	idx := md.anims.idx
	return func() gruid.Msg {
		return msgAnim(idx)
	}
}

// startAnimSeq is a helper function for starting a new animation sequence from
// a freshly drawn map representing the current state.
func (md *model) startAnimSeq() {
	if md.anims.Done() {
		md.resetAnimations()
	}
	md.drawMap(md.anims.grid)
}

func (md *model) resetAnimations() {
	gd := md.gd.Slice(md.gd.Range().Shift(0, 2, 0, -1))
	md.anims.grid.Copy(gd)
	md.anims.pgrid.Copy(gd)
}

// AnimationFrame updates the map and waits briefly.
func (md *model) AnimationFrame() {
	if DisableAnimations {
		return
	}
	md.startAnimSeq()
	md.anims.Frame(AnimDurShort)
}

// HitAnimation animates a player's non-killing hit at a given position.
func (md *model) HitAnimation(p gruid.Point) {
	if DisableAnimations || !md.g.InFOV(p) {
		return
	}
	md.startAnimSeq()
	md.anims.Draw(p, '√', ColorMagenta)
	md.anims.Frame(AnimDurMedium)
}

// DeathAnimation animates a monster's death at a given position.
func (md *model) DeathAnimation(p gruid.Point) {
	if DisableAnimations || !md.g.InFOV(p) {
		return
	}
	md.startAnimSeq()
	md.anims.Draw(p, '∞', ColorMagenta)
	md.anims.Frame(AnimDurMedium)
}

// MoveAnimation animates an actor's straight movement between two positions
// (possibly adjacent).
func (md *model) MoveAnimation(from, to gruid.Point) {
	if DisableAnimations {
		return
	}
	if from == to {
		// should not happen
		return
	}
	done := false
	delta := to.Sub(from)
	delta = delta.Div(max(delta.X, -delta.X, delta.Y, -delta.Y))
	for p := from; p != to; p = p.Add(delta) {
		if md.g.InFOV(p) {
			if !done {
				done = true
				md.startAnimSeq()
			}
			md.anims.Draw(p, '×', ColorForeground)
		}
	}
	if done {
		md.anims.Frame(AnimDurMedium)
	}
}

// RangedAttackAnimation animates a bump-ranged attack.
func (md *model) RangedAttackAnimation(a *Actor, from, to gruid.Point) {
	if DisableAnimations {
		return
	}
	if from == to {
		// should not happen
		return
	}
	fg := ColorForeground
	switch {
	case a.DoesAny(MonsSpitFire):
		fg = ColorOrange
	case a.DoesAny(MonsFear):
		fg = ColorMagenta
	case a.DoesAny(MonsConfusion):
		fg = ColorGreen
	}
	done := false
	delta := to.Sub(from)
	delta = delta.Div(max(delta.X, -delta.X, delta.Y, -delta.Y))
	for p := from; p != to; p = p.Add(delta) {
		if md.g.InFOV(p) {
			if !done {
				done = true
				md.startAnimSeq()
			}
			md.anims.Draw(p, '*', fg)
		}
	}
	if done {
		md.anims.Frame(AnimDurMedium)
	}
}

// WoundedAnimations alerts the player when they're hurt.
func (md *model) WoundedAnimation() {
	if DisableAnimations {
		return
	}
	g := md.g
	md.startAnimSeq()
	md.anims.DrawColor(g.PP(), ColorOrange)
	md.anims.Frame(AnimDurMedium)
}

// TeleportAnimation animates an actor teleport between two positions.
func (md *model) TeleportAnimation(from, to gruid.Point, showto bool) {
	if DisableAnimations {
		return
	}
	md.startAnimSeq()
	md.anims.Draw(from, 'Φ', ColorCyan)
	if showto {
		md.anims.Draw(to, 'Φ', ColorCyan)
	}
	md.anims.Frame(AnimDurMediumLong)
}

// FirebreathAnimation animates a firebreath attack from a given point and
// direction.
func (md *model) FirebreathAnimation(from, dir gruid.Point) {
	if DisableAnimations {
		return
	}
	md.startAnimSeq()
	g := md.g
	for k := 0; k < 4; k++ {
		p := from
		for i := 0; i < MaxFOVRange; i++ {
			t := g.Map.Terrain.At(p)
			if !Passable(t) {
				break
			}
			if !g.InFOV(p) {
				continue
			}
			r := '*'
			switch g.IntN(3) {
			case 0:
				r = '×'
			case 1:
				r = '+'
			}
			md.anims.DrawReverse(p, r, ColorOrange)
			p = p.Add(dir)
		}
		md.anims.Frame(AnimDurShortMedium)
	}
}

// LightningAnimation animates lighting affecting the given nodes.
func (md *model) LightningAnimation(ps []paths.Node) {
	if DisableAnimations {
		return
	}
	md.startAnimSeq()
	g := md.g
	for k := 0; k < 4; k++ {
		for _, n := range ps {
			p := n.P
			if !g.InFOV(p) {
				continue
			}
			r := '*'
			switch g.IntN(3) {
			case 0:
				r = '×'
			case 1:
				r = '+'
			}
			md.anims.Draw(p, r, ColorYellow)
		}
		md.anims.Frame(AnimDurShortMedium)
	}
}

// ExplosionAnimations animates an explosion affecting the given points.
func (md *model) ExplosionAnimation(ps []gruid.Point) {
	if DisableAnimations {
		return
	}
	md.startAnimSeq()
	g := md.g
	for k := 0; k < 4; k++ {
		for _, p := range ps {
			if !g.InFOV(p) {
				continue
			}
			r := '*'
			switch g.IntN(5) {
			case 0:
				r = '×'
			case 1:
				r = '+'
			case 2:
				r = '}'
			case 3:
				r = ';'
			}
			md.anims.DrawReverse(p, r, ColorOrange)
		}
		md.anims.Frame(AnimDurShortMedium)
	}
}

// OrbDestructionAnimation animates the destruction of the orb.
func (md *model) OrbDestructionAnimation() {
	if DisableAnimations {
		return
	}
	md.startAnimSeq()
	g := md.g
	c := gruid.Cell{
		Rune:  '♣',
		Style: gruid.Style{Fg: ColorGreen, Bg: ColorBackground, Attrs: AttrInMap}}
	for k := 0; k < 7; k++ {
		for p := range g.Map.Terrain.Points() {
			switch g.rand.IntN(4) {
			case 0:
				md.anims.grid.Set(p, c)
			case 1:
				if k >= 4 {
					md.anims.grid.Set(p, c)
				}
			}
		}
		md.anims.Frame(AnimDurLong)
	}
}
