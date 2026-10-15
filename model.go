// This files defines the model structure, as well as initialization functions.

package main

import (
	"fmt"
	"log"
	"maps"
	"math/rand/v2"
	"runtime"
	"slices"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
	"codeberg.org/anaseto/gruid/ui"
)

const (
	UIWidth  = 80 // UI width
	UIHeight = 24 // UI height
)

var (
	LogGame   = false       // write game logs to file
	ColorMode = ColorMode16 // default 16-color palette
)

// colorMode represents various color compatibility modes.
type colorMode int

const (
	ColorMode16    colorMode = iota
	ColorMode8               // use 8-color compatibility mode (default for windows)
	ColorMode256             // use solarized 256-color approximation
	ColorMode24bit           // use true color selenized palette
)

// GameConfig contains the current game config.
var GameConfig = Config{
	DarkColors:    true, // default to dark theme
	Tiles:         true, // default to tiles (when available)
	VersionNumber: ConfigVersionNumber,
}

// mode represents the main model mode
type mode int

const (
	modeLoadGame           mode = iota // game load screen (load game)
	modeNewGame                        // primary spirit selection
	modeNewGameMods                    // mod selection
	modeNormal                         // map game mode
	modeCritical                       // hp critical warning pause
	modePager                          // pager (logs, help, lore)
	modeMenu                           // menu (game menu, inventory, settings)
	modeWizardConfirmation             // waiting for wizard mode confirmation
	modeEnd                            // game end: win or death
	modeQuitConfirmation               // waiting for no-save quit confirmation
	modeQuitting                       // wait until end message
	modeUseConfirmation                // menhir or comestible pickup warning
	modeIntro
)

// model describes the gruid.Model of the game.
type model struct {
	action             Action     // action to handle
	anims              Animations // animations
	auto               *auto      // auto-travel mode info
	desc               *ui.Label  // description label (for monsters, terrain)
	equipPager         *ui.Pager  // description pager for equipping item
	gameEnded          bool       // whether the game ended
	textDescription    *ui.Pager
	textDescriptionKey string
	drawGroundDesc     bool       // whether to draw extra equipping desc
	g                  *Game      // game state
	gd                 gruid.Grid // drawing grid
	keysNormal         map[gruid.Key]Action
	keysTarget         map[gruid.Key]Action
	log                *ui.Label // game's last log messages
	menu               *menu     // menus (Menu, Settings, Inventory, Keymaps)
	menuActions        []Action  // invokable actions in last game/help/config menu
	mode               mode      // main mode
	pager              *pager    // pager (logs and the like)
	status             *gameStatus
	targ               *targeting
}

func (md *model) init() gruid.Effect {
	md.mode = modeLoadGame
	md.initStructures()
	md.initWidgets()
	md.initKeys()
	md.applyKeyConfig()

	g := md.g
	load, err := g.Load()
	md.g.md = md // handy cycle
	g.rand = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	if !load {
		// Start a new game: go to spirit selection menu.
		// Initialize mods before starting the game, because we need
		// that information during the new game settings.
		if len(GameConfig.Mods) == NMods && !ResetMods {
			g.Mods = slices.Clone(GameConfig.Mods)
		} else {
			g.Mods = make([]bool, NMods)
		}
		md.openSpiritSelectionMenu(modeNewGame)
		if LargeTextPanels {
			md.mode = modeIntro
		}
	}
	if err != nil {
		log.Printf("错误：%v", err)
	}
	md.targ.CancelExamine()
	md.InitAnimations()
	if runtime.GOOS == "js" {
		if load {
			md.updateStatus()
			md.mode = modeNormal
		}
		return nil
	}
	return gruid.Sub(subSig)
}

func (md *model) initStructures() {
	md.auto = &auto{}
	md.auto.PRauto = paths.NewPathRange(gruid.NewRange(0, 0, MapWidth, MapHeight))
}

func (md *model) initWidgets() {
	md.log = ui.NewLabel(ui.StyledText{}.WithMarkups(Markups))
	md.desc = ui.NewLabel(ui.StyledText{}.WithMarkups(Markups))
	md.desc.AdjustWidth = false
	md.equipPager = ui.NewPager(ui.PagerConfig{
		Grid: gruid.NewGrid(menuPanelWidth(), menuPanelHeight()),
		Box:  &ui.Box{},
	})
	md.pager = &pager{}
	md.pager.pg = ui.NewPager(ui.PagerConfig{
		Grid: gruid.NewGrid(TextPanelWidth, textPanelHeight()),
		Box:  &ui.Box{},
		Keys: ui.PagerKeys{Quit: []gruid.Key{gruid.KeySpace, "x", "X", gruid.KeyEscape}},
	})
	md.pager.markup = ui.StyledText{}.WithMarkups(Markups)
	style := ui.MenuStyle{
		Active: gruid.Style{Fg: ColorYellow},
	}
	md.status = &gameStatus{}
	md.status.desc = ui.NewLabel(ui.StyledText{}.WithMarkups(Markups))
	md.status.menu = ui.NewMenu(ui.MenuConfig{
		Grid:  gruid.NewGrid(UIWidth-2, 1),
		Style: ui.MenuStyle{Layout: gruid.Point{0, 1}, Active: style.Active},
	})
	md.menu = &menu{}
	md.menu.main = ui.NewMenu(ui.MenuConfig{
		Grid:  gruid.NewGrid(menuPanelWidth(), menuPanelHeight()),
		Box:   &ui.Box{},
		Style: style,
		Keys:  ui.MenuKeys{Quit: []gruid.Key{gruid.KeySpace, "x", "X", gruid.KeyEscape}},
	})
	md.menu.keys = ui.NewMenu(ui.MenuConfig{
		Grid:  gruid.NewGrid(TextPanelWidth, textPanelHeight()),
		Box:   &ui.Box{},
		Style: style,
		Keys:  ui.MenuKeys{Quit: []gruid.Key{gruid.KeySpace, "x", "X", gruid.KeyEscape}},
	})
}

// DefaultKeysNormal represents the default keybindings in normal mode (map).
var DefaultKeysNormal = map[gruid.Key]Action{
	gruid.KeyEscape:     ActionNone{},
	gruid.KeyArrowLeft:  ActionBump{Delta: gruid.Point{-1, 0}},
	gruid.KeyArrowDown:  ActionBump{Delta: gruid.Point{0, 1}},
	gruid.KeyArrowUp:    ActionBump{Delta: gruid.Point{0, -1}},
	gruid.KeyArrowRight: ActionBump{Delta: gruid.Point{1, 0}},
	"h":                 ActionBump{Delta: gruid.Point{-1, 0}},
	"j":                 ActionBump{Delta: gruid.Point{0, 1}},
	"k":                 ActionBump{Delta: gruid.Point{0, -1}},
	"l":                 ActionBump{Delta: gruid.Point{1, 0}},
	"H":                 ActionRun{Delta: gruid.Point{-1, 0}},
	"J":                 ActionRun{Delta: gruid.Point{0, 1}},
	"K":                 ActionRun{Delta: gruid.Point{0, -1}},
	"L":                 ActionRun{Delta: gruid.Point{1, 0}},
	".":                 ActionWait{},
	gruid.KeyEnter:      ActionWait{},
	"o":                 ActionAutoExplore{},
	"+":                 ActionNextMonster{},
	"-":                 ActionPreviousMonster{},
	"%":                 ActionNextItem{itemComestible},
	"!":                 ActionNextItem{itemTotem},
	"&":                 ActionNextItem{itemMenhir},
	">":                 ActionNextItem{itemPortal},
	"=":                 ActionNextItem{itemRune},
	"x":                 ActionExamineModeToggle{},
	"e":                 ActionInteract{},
	"i":                 ActionInventory{},
	gruid.KeySpace:      ActionMenu{},
	"?":                 ActionHelp{},
	"#":                 ActionDump{},
	"S":                 ActionSaveQuit{},
	"C":                 ActionConfig{},
	gruid.KeyTab:        ActionConfig{},
	":":                 ActionSetKeys{},
	"m":                 ActionViewMessages{},
	"Q":                 ActionQuit{},
	"W":                 ActionWizard{},
	gruid.KeyPageDown:   ActionScroll{Delta: gruid.Point{0, -1}},
	"d":                 ActionScroll{Delta: gruid.Point{0, -1}},
	gruid.KeyPageUp:     ActionScroll{Delta: gruid.Point{0, 1}},
	"u":                 ActionScroll{Delta: gruid.Point{0, 1}},
	"»":                 ActionWizardNextLevel{},
}

// DefaultKeysNormal represents the default keybindings changes in
// targeting/examine mode with respect to normal mode.
var DefaultKeysTarget = map[gruid.Key]Action{
	gruid.KeyArrowLeft:  ActionCursorBump{Delta: gruid.Point{-1, 0}},
	gruid.KeyArrowDown:  ActionCursorBump{Delta: gruid.Point{0, 1}},
	gruid.KeyArrowUp:    ActionCursorBump{Delta: gruid.Point{0, -1}},
	gruid.KeyArrowRight: ActionCursorBump{Delta: gruid.Point{1, 0}},
	"h":                 ActionCursorBump{Delta: gruid.Point{-1, 0}},
	"j":                 ActionCursorBump{Delta: gruid.Point{0, 1}},
	"k":                 ActionCursorBump{Delta: gruid.Point{0, -1}},
	"l":                 ActionCursorBump{Delta: gruid.Point{1, 0}},
	"H":                 ActionCursorRun{Delta: gruid.Point{-1, 0}},
	"J":                 ActionCursorRun{Delta: gruid.Point{0, 1}},
	"K":                 ActionCursorRun{Delta: gruid.Point{0, -1}},
	"L":                 ActionCursorRun{Delta: gruid.Point{1, 0}},
	gruid.KeyEnter:      ActionTravel{},
	".":                 ActionTravel{},
	gruid.KeyEscape:     ActionExamineModeToggle{},
}

func (md *model) initKeys() {
	md.keysNormal = maps.Clone(DefaultKeysNormal)
	md.keysTarget = maps.Clone(DefaultKeysTarget)
}

func (md *model) applyKeyConfig() {
	if GameConfig.NormalModeKeys != nil {
		md.keysNormal = GameConfig.NormalModeKeys
		// For ensuring menu access and esc functionality.
		md.keysNormal[gruid.KeySpace] = ActionMenu{}
		md.keysNormal[gruid.KeyEscape] = ActionNone{}
		// Ensure some new non-configurable actions are set on old
		// but compatible configs.
		if _, ok := md.keysNormal["»"]; !ok {
			md.keysNormal["»"] = ActionWizardNextLevel{}
		}
	}
	if GameConfig.ExamineModeKeys != nil {
		md.keysTarget = GameConfig.ExamineModeKeys
		// For ensuring back to normal mode.
		md.keysTarget[gruid.KeyEscape] = ActionExamineModeToggle{}
	}
}

// InitConfig loads saved config, if any, and initializes GameConfig.
func InitConfig() error {
	_, err := LoadConfig()
	if err != nil {
		err = fmt.Errorf("读取配置时出错：%v", err)
		saverr := SaveConfig()
		if saverr != nil {
			log.Printf("重置损坏配置时出错：%v", err)
		}
		return err
	}
	return nil
}

var spiritKeys = []rune{'H', 'B', 'F', 'f', 'c', 'C', 'b'}

func (md *model) openSpiritSelectionMenu(m mode) {
	hstyle := gruid.Style{Fg: ColorCyan}
	entries := []ui.MenuEntry{}
	title := "新游戏"
	for i, si := range primarySpirits {
		r := spiritKeys[i]
		switch r {
		case 'H':
			entries = append(entries, ui.MenuEntry{
				Text:     ui.Text("主灵体选择").WithStyle(hstyle),
				Disabled: true,
			})
		}
		advanced := ""
		if i >= 5 {
			advanced = " @S（高级）@N"
		}
		name := si.Name
		_, isVampiric := si.Ability[0].(EffectVampirism)
		if md.g.Mod(ModHealingCombat) && !VampiricHC && isVampiric {
			name = "@R吸血@N蝙蝠"
		}
		entries = append(entries, ui.MenuEntry{
			Text: ui.Textf("%c - %s%s", r, name, advanced).WithMarkups(Markups),
			Keys: []gruid.Key{gruid.Key(r)},
		})
	}
	altBgEntries(entries)
	md.menu.main.SetBox(&ui.Box{Title: ui.Text(title).WithStyle(gruid.Style{Fg: ColorYellow})})
	md.menu.main.SetEntries(entries)
	md.menu.main.SetActiveInvokable(0)
	md.updateItemDesc(spiritEntity(md.g.Mods, primarySpirits[0]))
	md.mode = m
	md.menu.mode = modeSelection // not really needed
}
