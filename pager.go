package main

import (
	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/ui"
)

// pager gathers the structures used in the various kinds of pagers.
type pager struct {
	big *ui.Pager
	// small  *ui.Pager
	mode   pagerMode
	markup ui.StyledText
	lines  []ui.StyledText
}

// pagerMode represents the available pager modes
type pagerMode int

const (
	modeLogs pagerMode = iota
	modeDump
	modeHelp
	// modeLore // unused
)

func (md *model) updatePager(msg gruid.Msg) gruid.Effect {
	switch md.pager.mode {
	// case modeLore:
	// 	md.pager.small.Update(msg)
	// 	if md.pager.small.Action() == ui.PagerQuit {
	// 		md.mode = modeNormal
	// 	}
	case modeDump:
		md.pager.big.Update(msg)
		if md.pager.big.Action() == ui.PagerQuit {
			md.mode = modeQuitting
			return gruid.End()
		}
	default:
		md.pager.big.Update(msg)
		if md.pager.big.Action() == ui.PagerQuit {
			md.mode = modeNormal
		}
	}
	return nil
}
