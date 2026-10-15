package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"runtime/debug"

	"codeberg.org/anaseto/gruid/ui"
)

// Stats gathers various game statistics from a run.
type Stats struct {
	ActivatedMenhirs    int            // number of activated menhirs
	Comestibles         map[string]int // number of eaten comestibles (per type)
	Cornered            int            // number of times you felt cornered
	CurseHistory        []string       // record of curse history
	Damage              int            // total damage you got
	Deaths              map[string]int // monster deaths by name
	DeathsPast          map[string]int // monster deaths by name (past levels)
	Digs                int            // number of digs (面墙壁被摧毁s)
	DigsPast            int            // number of digs (past levels)
	EatenComestibles    int            // number of eaten comestibles
	FireClouds          int            // number of spawn fire clouds
	FireCloudsPast      int            // number of spawn fire clouds (past levels)
	Hits                int            // number of times you hit with a bump-attack
	Hurt                int            // number of times you got hurt
	Lucky               int            // number of times you were lucky (could die to bad roll but got lucky roll)
	Misses              int            // number of times you missed a bump-attack
	MonsterTriggers     int            // number of triggered traps by the monsters
	MonsterTriggersPast int            // number of triggered traps by the monsters (past levels)
	NDeaths             int            // number of monster deaths
	NDeathsPast         int            // number of monster deaths (past levels)
	PlayerTriggers      int            // number of triggered traps by the player
	PoisonClouds        int            // number of spawn poison clouds
	PoisonCloudsPast    int            // number of spawn poison clouds (past levels)
	SpiritUses          int            // number of ability uses
	Statuses            []int          // times you got each status
	Waits               int            // number of times you waited
	WallJumps           int            // number of jumps off walls (Gawalt Monkey spirit)
	WallThroughs        int            // number of translucent walls passed (Gawalt Monkey spirit)

	MapTurns            [MapLevels]int // number of turns per level
	MapExplorePerc      [MapLevels]int // exploration percent per level
	MapDeathPerc        [MapLevels]int // monster death percent per level
	MapDamage           [MapLevels]int // endured damage per level
	MapEatenComestibles [MapLevels]int // eaten comestibles per level
	MapSpiritUses       [MapLevels]int // spirit uses per level
	MapActivatedMenhirs [MapLevels]int // activated menhirs per level
	MapTriggeredTraps   [MapLevels]int // number of triggered traps per level
}

// newStats returns newly initialized structure for statistics.
func newStats() *Stats {
	return &Stats{
		Deaths:      map[string]int{},
		DeathsPast:  map[string]int{},
		Comestibles: map[string]int{},
		Statuses:    make([]int, NStatuses),
	}
}

// Death registers the death of a monster with given name.
func (gs *Stats) Death(name string) {
	gs.NDeaths++
	gs.Deaths[name]++
}

// LevelStats collects stats about current level (just before leaving).
func (g *Game) LevelStats() {
	gs := g.Stats
	// Explored percentage.
	n, total := 0, 0
	for p, t := range g.Map.Terrain.All() {
		if !Passable(t) && !g.Map.Connected(p) {
			continue
		}
		total++
		if IsKnown(g.Map.KnownTerrain.At(p)) {
			n++
		}
	}
	gs.MapExplorePerc[g.Map.Level-1] = int(100 * float64(n) / float64(total))
	// Death percentage.
	n, total = 0, 0
	for _, e := range g.NPMapEntities() {
		if a, ok := e.Role.(*Actor); ok {
			total++
			if a.IsDead() {
				n++
			}
		}
	}
	if total == 0 {
		// In no monsters, show 100% (0% would make sense, too,
		// but be less recognizable).
		gs.MapDeathPerc[g.Map.Level-1] = 100
	} else {
		gs.MapDeathPerc[g.Map.Level-1] = int(100 * float64(n) / float64(total))
	}
	// Save information that we show only up to last level for in-progress
	// non-debug runs.
	gs.DeathsPast = maps.Clone(gs.Deaths)
	gs.DigsPast = gs.Digs
	gs.FireCloudsPast = gs.FireClouds
	gs.MonsterTriggersPast = gs.MonsterTriggers
	gs.NDeathsPast = gs.NDeaths
	gs.PoisonCloudsPast = gs.PoisonClouds
}

func (g *Game) getStatDeaths() map[string]int {
	if g.showAllStats() {
		return g.Stats.Deaths
	}
	return g.Stats.DeathsPast
}

func (g *Game) getStatFireClouds() int {
	if g.showAllStats() {
		return g.Stats.FireClouds
	}
	return g.Stats.FireCloudsPast
}

func (g *Game) getStatMonsterTriggers() int {
	if g.showAllStats() {
		return g.Stats.MonsterTriggers
	}
	return g.Stats.MonsterTriggersPast
}

func (g *Game) getStatNDeaths() int {
	if g.showAllStats() {
		return g.Stats.NDeaths
	}
	return g.Stats.NDeathsPast
}

func (g *Game) getStatPoisonClouds() int {
	if g.showAllStats() {
		return g.Stats.PoisonClouds
	}
	return g.Stats.PoisonCloudsPast
}

func (g *Game) getStatExcludeStr(markup bool) string {
	if g.showAllStats() {
		return ""
	}
	if markup {
		return " @S（不含当前层）@N"
	}
	return "（不含当前层）"
}

// showAllStats reports whether game statistics should show everything, including
// informations of current level that could be unknown yet to the player.
func (g *Game) showAllStats() bool { return g.Wizard.Extra || g.win || g.PlayerActor().IsDead() }

// DumpEndSummary produces the game statistics short summary displayed at the end
// of the game. When markup is enabled, in-game viewing is assumed.
func (g *Game) DumpEndSummary(markup bool) string {
	var sb strings.Builder
	var version string
	info, ok := debug.ReadBuildInfo()
	if ok {
		version = info.Main.Version
	} else {
		version = Version
	}
	fmt.Fprintf(&sb, " ♣ Shamogu 游戏总结 %s ♣\n\n", version)
	if g.Wizard.Mode != WizardNone {
		fmt.Fprintf(&sb, "已启用向导模式（复活 %s",
			CountNoun("次", g.Wizard.Resurrections))
		if g.Wizard.Extra {
			sb.WriteString("，额外作弊选项")
		}
		sb.WriteString("）。\n")
	}
	if !markup {
		pa := g.PlayerActor()
		if g.win {
			fmt.Fprintf(&sb, "你摧毁了腐化之球！\n")
		} else if pa.IsDead() {
			fmt.Fprintf(&sb, "你在探索地城第 %d 层时死亡。\n", g.Map.Level)
		} else {
			fmt.Fprintf(&sb, "你正在探索地城第 %d 层。\n", g.Map.Level)
		}
		fmt.Fprintf(&sb, "你在地城中行动了 %d 回合。\n", g.Turn)
	}
	fmt.Fprintf(&sb, "此番冒险中共有 %s 怪物死亡%s。\n", CountNoun("只", g.getStatNDeaths()), g.getStatExcludeStr(markup))
	g.dumpMods(&sb, markup)
	g.dumpPlayer(&sb, markup)
	return sb.String()
}

func (g *Game) modCount() int {
	var n int
	for _, b := range g.Mods {
		if b {
			n++
		}
	}
	return n
}

// Dump produces the game statistics full summary.
func (g *Game) Dump() string {
	var sb strings.Builder
	sb.WriteString(g.DumpEndSummary(false))
	g.dumpComestibles(&sb, false)
	g.dumpLastMessages(&sb, false)
	g.dumpMap(&sb, false)
	g.dumpMonsterDeaths(&sb, false)
	g.dumpEatenComestibles(&sb, false)
	g.dumpStatistics(&sb, false)
	g.dumpCurseHistory(&sb, false)
	g.dumpTimeline(&sb, false)
	return sb.String()
}

// DumpInGame produces the statistics summary for in-game pager viewing (with
// markup enabled).
func (g *Game) DumpInGame() string {
	var sb strings.Builder
	sb.WriteString(g.DumpEndSummary(true))
	g.dumpMonsterDeaths(&sb, true)
	g.dumpEatenComestibles(&sb, true)
	g.dumpStatistics(&sb, true)
	g.dumpCurseHistory(&sb, true)
	g.dumpTimeline(&sb, true)
	return sb.String()
}

func (g *Game) dumpMods(sb *strings.Builder, markup bool) {
	if g.modCount() == 0 {
		return
	}
	sb.WriteString(section("模组", markup))
	for i, b := range g.Mods {
		if !b {
			continue
		}
		fmt.Fprintf(sb, "- %s\n", Mod(i).String())
	}
}

func (g *Game) dumpPlayer(sb *strings.Builder, markup bool) {
	pa := g.PlayerActor()
	if !markup {
		fmt.Fprintf(sb, "\n生命:%d/%d 攻:%d 防:%d",
			pa.HP, pa.GetMaxHP(), pa.GetAttack(), pa.GetDefense())
		for i, turns := range pa.Statuses {
			if turns <= 0 {
				continue
			}
			sb.WriteByte(' ')
			st := Status(i)
			fmt.Fprintf(sb, "%s(%d)", g.StatusAbbr(pa, st), turns)
		}
		sb.WriteByte('\n')
	}
	sb.WriteString(section("灵体", markup))
	for i, sp := range g.PlayerSpirits() {
		ei := g.Entity(i)
		fmt.Fprintf(sb, "- %s（%s，充能 %d/%d，使用 %s）\n",
			ei.Name, sp.GetAbility().Name(),
			sp.Charges, sp.GetMaxCharges(),
			times(sp.Uses))
	}
}

func (g *Game) dumpComestibles(sb *strings.Builder, markup bool) {
	sb.WriteString(section("食物", markup))
	for i := range g.PlayerComestibles() {
		ei := g.Entity(i)
		fmt.Fprintf(sb, "- %s\n", ei.Name)
	}
}

func (g *Game) dumpLastMessages(sb *strings.Builder, markup bool) {
	sb.WriteString(section("最近消息", markup))
	for _, e := range g.Logs.Entries[max(0, len(g.Logs.Entries)-20):] {
		fmt.Fprintf(sb, "%s\n", e.dumpString())
	}
}

func (g *Game) dumpMap(sb *strings.Builder, markup bool) {
	sb.WriteString(section("地图", markup))
	// Entities.
	var ids CacheGrid[ID]
	ids = ids.New()
	sortedIDs := g.sortedIDsByRenderOrder()
	for _, i := range sortedIDs {
		ei := g.Entity(i)
		if a, ok := ei.Role.(*Actor); ok && i != PlayerID && a.IsDead() {
			continue
		}
		ids.Set(ei.P, i)
	}
	// Monster noise.
	var noise CacheGrid[ID]
	noise = noise.New()
	for _, i := range sortedIDs {
		ei := g.Entity(i)
		if ai, ok := ei.Role.(*Actor); ok && ei.Noise && ai.IsAlive() {
			noise.Set(ei.P, i)
			continue
		}
	}
	// Draw map.
	for p := range g.Map.KnownTerrain.Points() {
		if p.X%MapWidth == 0 {
			if p.Y == 0 {
				sb.WriteRune('|')
			} else {
				sb.WriteString("|\n|")
			}
		}
		r := g.Map.KnownTerrainRuneAt(p)
		if r == '♫' {
			r = ' '
		}
		if !g.InFOV(p) {
			if i := noise.At(p); i > 0 {
				r = '♫'
			} else if _, ok := g.Map.Noise[p]; ok {
				r = '♪'
			}
			sb.WriteRune(r)
			continue
		}
		if g.Map.Clouds.At(p).Kind != NoCloud && (r == '.' || r == '"' || r == '^') {
			r = '≡'
		}
		if i := ids.At(p); i > 0 {
			r = g.Entity(i).Rune
		}
		sb.WriteRune(r)
	}
	sb.WriteString("|\n")
}

func (g *Game) dumpMonsterDeaths(sb *strings.Builder, markup bool) {
	if g.getStatNDeaths() == 0 {
		return
	}
	var t string
	if markup {
		t = section(fmt.Sprintf("怪物死亡%s@C", g.getStatExcludeStr(markup)), markup)
	} else {
		t = section(fmt.Sprintf("怪物死亡%s", g.getStatExcludeStr(markup)), markup)
	}
	sb.WriteString(t)
	deaths := g.getStatDeaths()
	monsters := slices.Sorted(maps.Keys(deaths))
	for _, mons := range monsters {
		fmt.Fprintf(sb, "- %s: %d\n", mons, deaths[mons])
	}
}

func (g *Game) dumpEatenComestibles(sb *strings.Builder, markup bool) {
	if g.Stats.EatenComestibles == 0 {
		return
	}
	sb.WriteString(section("食用食物", markup))
	comestibles := slices.Sorted(maps.Keys(g.Stats.Comestibles))
	for _, co := range comestibles {
		fmt.Fprintf(sb, "- %s: %d\n", co, g.Stats.Comestibles[co])
	}
}

func (g *Game) dumpStatistics(sb *strings.Builder, markup bool) {
	sb.WriteString(section("统计", markup))
	// 注意：游戏进行中查看统计时，可能泄露少量信息（例如被摧毁的墙壁数，
	// 或包括视野外陷阱在内的触发次数）。为避免这些泄漏而增加复杂度并不值得。
	timesPer100 := func(n int) string {
		return fmt.Sprintf("%s（每 100 回合 %.1f 次）",
			times(n), float64(n)*100/float64(max(1, g.Turn)))
	}
	fmt.Fprintf(sb, "你使用灵体 %s，食用食物 %s，激活石碑 %s。\n",
		times(g.Stats.SpiritUses),
		CountNoun("份", g.Stats.EatenComestibles),
		CountNoun("次", g.Stats.ActivatedMenhirs))
	fmt.Fprintf(sb, "你触发陷阱 %s。\n", CountNoun("次", g.Stats.PlayerTriggers))
	fmt.Fprintf(sb, "你命中敌人 %s。\n", timesPer100(g.Stats.Hits))
	fmt.Fprintf(sb, "你未命中敌人 %s。\n", timesPer100(g.Stats.Misses))
	fmt.Fprintf(sb, "你等待了 %s。\n", timesPer100(g.Stats.Waits))
	fmt.Fprintf(sb, "你受伤 %s，共承受 %d 点伤害。\n", times(g.Stats.Hurt), g.Stats.Damage)
	fmt.Fprintf(sb, "你走运 %s。\n", times(g.Stats.Lucky))
	if g.Stats.Cornered > 0 {
		fmt.Fprintf(sb, "你陷入绝境 %s。\n", times(g.Stats.Cornered))
	}
	if g.Stats.WallJumps > 0 {
		fmt.Fprintf(sb, "你借墙反弹跳跃 %s。\n", times(g.Stats.WallJumps))
	}
	if g.Stats.WallThroughs > 0 {
		fmt.Fprintf(sb, "你穿过半透明墙 %s。\n", times(g.Stats.WallThroughs))
	}
	for i, n := range g.Stats.Statuses {
		st := Status(i)
		if n > 0 {
			fmt.Fprintf(sb, "你处于%s状态 %s。\n", st, times(n))
		}
	}
	ses := g.getStatExcludeStr(markup)
	fmt.Fprintf(sb, "你遭遇%s与%s%s。\n",
		CountNoun("团火焰云雾", g.getStatFireClouds()),
		CountNoun("团毒雾", g.getStatPoisonClouds()),
		ses)
	fmt.Fprintf(sb, "被摧毁的墙壁：%s%s。\n", CountNoun("面", g.Stats.Digs), ses)
	fmt.Fprintf(sb, "怪物触发陷阱：%s%s。\n", CountNoun("次", g.getStatMonsterTriggers()), ses)
	// Statistics per map level.
	levels := g.Map.Level - 1
	if g.win || g.PlayerActor().IsDead() {
		levels++
	}
	if levels == 0 {
		return
	}
	sb.WriteByte('\n')
	hfmt := "%-20s"
	fmt.Fprintf(sb, hfmt, "项目/层数")
	for i := range levels {
		fmt.Fprintf(sb, " %4d", i+1)
	}
	sb.WriteByte('\n')
	perLevel := func(title string, x []int) {
		fmt.Fprintf(sb, hfmt, title)
		for _, n := range x {
			fmt.Fprintf(sb, " %4d", n)
		}
		sb.WriteByte('\n')
	}
	perLevel("回合数", g.Stats.MapTurns[:levels])
	perLevel("探索度（%）", g.Stats.MapExplorePerc[:levels])
	perLevel("死亡怪物（%）", g.Stats.MapDeathPerc[:levels])
	perLevel("承受伤害", g.Stats.MapDamage[:levels])
	perLevel("食用食物", g.Stats.MapEatenComestibles[:levels])
	perLevel("使用灵体", g.Stats.MapSpiritUses[:levels])
	perLevel("激活石碑", g.Stats.MapActivatedMenhirs[:levels])
	perLevel("触发陷阱", g.Stats.MapTriggeredTraps[:levels])
}

func (g *Game) dumpCurseHistory(sb *strings.Builder, markup bool) {
	if !g.Mod(ModTotemConditions) {
		return
	}
	sb.WriteString(section("诅咒", markup))
	for _, s := range g.Stats.CurseHistory {
		stt := ui.Text(s).Format(76)
		sb.WriteString(strings.ReplaceAll(stt.Text(), "\n", "\n  "))
		sb.WriteByte('\n')
	}
}

func (g *Game) recordCurse(spname string, c *SpiritCond) {
	s := fmt.Sprintf("- %s在第 %d 层受到诅咒：%s。\n持续条件：%s。",
		spname, g.Map.Level, UpperFirst(c.Effect.Desc()), UpperFirst(c.Until.Desc()))
	g.Stats.CurseHistory = append(g.Stats.CurseHistory, s)
}

func (g *Game) dumpTimeline(sb *strings.Builder, markup bool) {
	sb.WriteString(section("时间线", markup))
	for _, s := range g.Logs.Story {
		sb.WriteString(s)
		sb.WriteByte('\n')
	}
}

func section(s string, markup bool) string {
	if markup {
		return "\n@C" + s + ":@N\n"
	}
	return "\n" + s + ":\n"

}

func times(n int) string {
	return CountNoun("次", n)
}
