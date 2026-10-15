package ui

import (
	"fmt"
	"strings"

	"codeberg.org/anaseto/gruid"
)

// StyledText is a simple text formatter and styler. The zero value can be
// used, but you may prefer using Text and Textf.
type StyledText struct {
	markups map[rune]gruid.Style
	text    string
	style   gruid.Style
}

// Text is a shorthand for StyledText{}.WithText and creates a new styled text
// with the given text and the default style.
func Text(text string) StyledText {
	return StyledText{text: text}
}

// Textf returns a new styled text with the given formatted text, and default style.
func Textf(format string, a ...any) StyledText {
	return StyledText{text: fmt.Sprintf(format, a...)}
}

// NewStyledText returns a new styled text with the given text and style. It is
// a shorthand for StyledText{}.With.
func NewStyledText(text string, st gruid.Style) StyledText {
	return StyledText{text: text, style: st}
}

// Text returns the current styled text as a string.
func (stt StyledText) Text() string {
	return stt.text
}

// WithText returns a derived styled text with updated text.
func (stt StyledText) WithText(text string) StyledText {
	stt.text = text
	return stt
}

// WithTextf returns a derived styled text with updated formatted text.
func (stt StyledText) WithTextf(format string, a ...any) StyledText {
	stt.text = fmt.Sprintf(format, a...)
	return stt
}

// Style returns the text default style.
func (stt StyledText) Style() gruid.Style {
	return stt.style
}

// WithStyle returns a derived styled text with a updated default style.
func (stt StyledText) WithStyle(style gruid.Style) StyledText {
	stt.style = style
	return stt
}

// With returns a derived styled text with new next and style.
func (stt StyledText) With(text string, style gruid.Style) StyledText {
	stt.text = text
	stt.style = style
	return stt
}

// WithMarkup returns a derived styled text with a new markup @r available for
// a given style.  Markup starts by a @ sign, and is followed then by a rune
// indicating the particular style. Default style is marked by @N.
//
// This simple markup is inspired from github/gdamore/tcell, but with somewhat
// different defaults: unless at least one non-default markup is registered,
// markup commands processing is not activated, and @ is treated as any other
// character.
func (stt StyledText) WithMarkup(r rune, style gruid.Style) StyledText {
	if r == ' ' || r == '\n' || r == '\r' {
		// avoid strange cases that can conflict with format
		return stt
	}
	if len(stt.markups) == 0 {
		stt.markups = map[rune]gruid.Style{}
	} else {
		omarks := stt.markups
		stt.markups = make(map[rune]gruid.Style, len(omarks))
		for k, v := range omarks {
			stt.markups[k] = v
		}
	}
	if r == 'N' {
		// N has a built-in meaning
		stt.style = style
		return stt
	}
	stt.markups[r] = style
	return stt
}

// WithMarkups is the same as WithMarkup but passing a whole map of rune-style
// associations.
func (stt StyledText) WithMarkups(markups map[rune]gruid.Style) StyledText {
	stt.markups = markups
	return stt
}

// Markups returns a copy of the markups currently defined for the styled text.
func (stt StyledText) Markups() map[rune]gruid.Style {
	markups := make(map[rune]gruid.Style, len(stt.markups))
	for k, v := range stt.markups {
		markups[k] = v
	}
	return markups
}

// Iter iterates a function for all couples positions and cells representing
// the styled text, and returns the minimum (w, h) size in cells which can fit
// the text.
func (stt StyledText) Iter(fn func(gruid.Point, gruid.Cell)) gruid.Point {
	x, y := 0, 0
	xmax := 0
	c := gruid.Cell{Style: stt.style}
	markup := stt.markups != nil // whether markup is activated
	procm := false               // processing markup
	for _, r := range stt.text {
		if r == '\r' {
			// We ignore carriage return.
			continue
		}
		if markup {
			if procMarkup(procm, r) {
				if procm {
					c.Style = stt.markupStyle(r)
				}
				procm = !procm
				continue
			}
			procm = false
		}
		if r == '\n' {
			if x > xmax {
				xmax = x
			}
			x = 0
			y++
			continue
		}
		c.Rune = r
		fn(gruid.Point{X: x, Y: y}, c)
		x++
	}
	if x > xmax {
		xmax = x
	}
	if xmax > 0 || y > 0 {
		y++ // at least one line
	}
	return gruid.Point{X: xmax, Y: y}
}

func procMarkup(procm bool, r rune) bool {
	if procm {
		return r != '@'
	}
	return r == '@'
}

// Size returns the minimum (w, h) size in cells which can fit the text.
func (stt StyledText) Size() gruid.Point {
	x, y := 0, 0
	xmax := 0
	markup := stt.markups != nil // whether markup is activated
	procm := false               // processing markup
	for _, r := range stt.text {
		if r == '\r' {
			// We ignore carriage return.
			continue
		}
		if markup {
			if procMarkup(procm, r) {
				procm = !procm
				continue
			}
			procm = false
		}
		if r == '\n' {
			if x > xmax {
				xmax = x
			}
			x = 0
			y++
			continue
		}
		x++
	}
	if x > xmax {
		xmax = x
	}
	if xmax > 0 || y > 0 {
		y++ // at least one line
	}
	return gruid.Point{X: xmax, Y: y}
}

// Format formats the text so that lines longer than a certain width get
// wrapped at word boundaries, if possible. It preserves spaces at the
// beginning of a line.
func (stt StyledText) Format(width int) StyledText {
	if width < 1 {
		return stt
	}
	// The game uses one rune per cell, including Chinese glyphs. Hard wrap by
	// displayed cells so unspaced Chinese paragraphs cannot disappear offscreen.
	var out strings.Builder
	col := 0
	markup := false
	runes := []rune(stt.text)
	for i, r := range runes {
		if r == '\r' {
			continue
		}
		if stt.markups != nil && r == '@' && !markup {
			out.WriteRune(r)
			markup = true
			continue
		}
		if markup {
			markup = false
			if r != '@' {
				out.WriteRune(r)
				continue
			}
		}
		if r == '\n' {
			out.WriteRune(r)
			col = 0
			continue
		}
		// Keep closing punctuation with its preceding character. Otherwise
		// a larger Chinese font can leave a lone full stop on the next row.
		closing := 0
		for j := i + 1; j < len(runes); j++ {
			if stt.markups != nil && runes[j] == '@' && j+1 < len(runes) && runes[j+1] != '@' {
				j++
				continue
			}
			if !strings.ContainsRune("，。！？；：、）》」』】〕〉〗〙〛…,.!?;:", runes[j]) {
				break
			}
			closing++
		}
		if col > 0 && closing > 0 && closing < width && col+1+closing > width {
			out.WriteByte('\n')
			col = 0
		}
		if col >= width {
			out.WriteByte('\n')
			col = 0
			if r == ' ' {
				continue
			}
		}
		out.WriteRune(r)
		col++
	}
	stt.text = strings.TrimRight(out.String(), " \n")
	return stt
}

func addSpace(s *strings.Builder, b bool) {
	if b {
		s.WriteRune('\n')
	} else {
		s.WriteRune(' ')
	}
}

func (stt StyledText) markupStyle(r rune) gruid.Style {
	if r == 'N' {
		return stt.style
	}
	st, ok := stt.markups[r]
	if ok {
		return st
	}
	return stt.style
}

// Lines splits the styled text into lines of styled texts, preserving markup.
func (stt StyledText) Lines() []StyledText {
	if stt.markups == nil {
		lines := strings.Split(stt.Text(), "\n")
		stts := make([]StyledText, len(lines))
		for i, line := range lines {
			stts[i] = stt.WithText(strings.TrimRight(line, "\r"))
		}
		return stts
	}
	markupRuneStart := 'N' // current markup rune (at line start)
	markupRune := 'N'      // current markup rune
	procm := false         // processing markup
	stts := []StyledText{}
	from := 0
	doNewline := func(i int) {
		procm = false
		line := strings.TrimRight(stt.text[from:i], "\r")
		if markupRuneStart != 'N' {
			line = "@" + string(markupRuneStart) + line
		}
		markupRuneStart = markupRune
		stts = append(stts, stt.WithText(line))
		from = i + 1
	}
	for i, r := range stt.text {
		if r == '\n' {
			doNewline(i)
			continue
		}
		switch {
		case procm:
			procm = false
			if r != '@' && r != ' ' && r != '\r' {
				markupRune = r
			}
		case r == '@':
			procm = true
		}
	}
	if from != len(stt.text) {
		doNewline(len(stt.text))
	}
	return stts
}

// Draw displays the styled text in a given grid. It returns the smallest grid
// slice containing the drawn part. Note that the grid is not cleared with
// spaces beforehand by this function, not even the returned one, you should
// use the styled text with a label for this.
func (stt StyledText) Draw(gd gruid.Grid) gruid.Grid {
	it := gd.Iterator()
	if !it.Next() {
		return gd
	}
	x, y := 0, 0
	xmax := 0
	c := gruid.Cell{Style: stt.style}
	markup := stt.markups != nil // whether markup is activated
	procm := false               // processing markup
	for _, r := range stt.text {
		if r == '\r' {
			// We ignore carriage return.
			continue
		}
		if markup {
			if procMarkup(procm, r) {
				if procm {
					c.Style = stt.markupStyle(r)
				}
				procm = !procm
				continue
			}
			procm = false
		}
		p := it.P()
		if r == '\n' {
			if x > xmax {
				xmax = x
			}
			x = 0
			y++
			if p.Y < y {
				it.SetP(gruid.Point{x, y})
				if it.P().Y != y {
					break
				}
			}
			continue
		}
		x++
		if p.Y > y {
			continue
		}
		c.Rune = r
		it.SetCell(c)
		if !it.Next() {
			break
		}
	}
	if x > xmax {
		xmax = x
	}
	if xmax > 0 || y > 0 {
		y++ // at least one line
	}
	return gd.Slice(gruid.NewRange(0, 0, xmax, y))
}
