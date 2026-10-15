package main

import (
	"fmt"
	"log"
	"strings"

	"codeberg.org/anaseto/gruid"
)

// Number of inventory items by type.
const (
	NSpirits     = 3 // number of primary and secondary spirits
	NComestibles = 5 // number of comestibles
)

// Effect represents any kind of applicable effect.
type Effect interface {
	Desc(*Game) string
	Apply(*Game) bool
}

// Ability represents a spirit ability effect.
type Ability interface {
	Effect
	Name() string // name of the ability as displayed in the inventory
}

// Item holds an item that can be used and equipped.
type Item interface {
	// Desc returns a description of the item.
	Desc(*Game) string
	// Use takes a game state and the item's id and reports whether a
	// turn-taking action was performed.
	Use(*Game, ID) bool
}

// IsComestible reports whether the item is a comestible.
func IsComestible(it Item) bool {
	_, ok := it.(*Comestible)
	return ok
}

// Spirit represent an equipable spirit.
type Spirit struct {
	Level        int        // spirit's level (in range 0-1)
	Charges      int        // number of active use charges
	MaxCharges   [2]int     // maximum number of charges per level
	Ability      [2]Ability // active ability effect when used
	BonusAttack  [2]int     // attack bonuses per level (may be zero)
	BonusDefense [2]int     // defense bonuses per level (may be zero)
	BonusTraits  [2]Traits  // bonus traits per level
	BonusHP      [2]int     // bonus hp per level (may be zero)
	Uses         int        // number of times it was used (for statistics)
	Advanced     bool       // advanced secondary spirit
	Condition    *SpiritCond
}

func (sp *Spirit) ExamineDesc(g *Game) string {
	return sp.eitherDesc(g, true)
}

func (sp *Spirit) Desc(g *Game) string {
	return sp.eitherDesc(g, false)
}

func (sp *Spirit) eitherDesc(g *Game, descUpgrade bool) string {
	var sb strings.Builder
	if descUpgrade {
		sb.WriteString("@C图腾灵体。@N\n")
	}
	if cond := sp.Condition; cond != nil {
		if descUpgrade {
			udesc := cond.Upgrade.Desc()
			if udesc != "" {
				sb.WriteString(udesc)
				sb.WriteByte('\n')
			}
		}
		sb.WriteString(cond.Desc())
		sb.WriteByte('\n')
	}
	bonuses := []string{}
	if attack := sp.BonusAttack[sp.Level]; attack > 0 {
		bonuses = append(bonuses, fmt.Sprintf("攻击 %+d", attack))
	}
	if defense := sp.BonusDefense[sp.Level]; defense > 0 {
		bonuses = append(bonuses, fmt.Sprintf("防御 %+d", defense))
	}
	if hp := sp.BonusHP[sp.Level]; hp > 0 {
		bonuses = append(bonuses, fmt.Sprintf("生命 %+d", hp))
	}
	if len(bonuses) > 0 {
		sb.WriteString("@C属性：@N")
		for i, bonus := range bonuses {
			sb.WriteByte(' ')
			sb.WriteString(bonus)
			if i < len(bonuses)-1 {
				sb.WriteString(", ")
			} else {
				sb.WriteByte('.')
			}
		}
		sb.WriteByte('\n')
	}
	fmt.Fprintf(&sb, "@C能力 @S（%s，充能 %d/%d）@C：@N %s",
		sp.GetAbility().Name(), sp.Charges, sp.GetMaxCharges(), sp.GetAbility().Desc(g))
	sb.WriteByte('\n')
	fmt.Fprintf(&sb, "@C特性：@N %s。", g.TraitDesc(Player, sp.BonusTraits[sp.Level]))
	return sb.String()
}

func (sp *Spirit) UpgradeDesc(g *Game, enew *Entity) string {
	desc := sp.Desc(g)
	var sb strings.Builder
	sb.WriteString("@M可使用灵体精华升级此灵体。@N\n")
	hasCond := sp.Condition != nil
	if cond := enew.Role.(*Spirit).Condition; cond != nil && cond.Upgrade.Applies(hasCond) {
		if hasCond {
			sb.WriteString("@R替换@N ")
		}
		sb.WriteString(cond.Desc())
		sb.WriteByte('\n')
	}
	bonuses := []string{}
	if attack := sp.BonusAttack[1] - sp.BonusAttack[0]; attack > 0 {
		bonuses = append(bonuses, fmt.Sprintf("攻击 %+d", attack))
	}
	if defense := sp.BonusDefense[1] - sp.BonusDefense[0]; defense > 0 {
		bonuses = append(bonuses, fmt.Sprintf("防御 %+d", defense))
	}
	if hp := sp.BonusHP[1] - sp.BonusHP[0]; hp > 0 {
		bonuses = append(bonuses, fmt.Sprintf("生命 %+d", hp))
	}
	if len(bonuses) > 0 {
		sb.WriteString("@B属性：@N")
		for i, bonus := range bonuses {
			sb.WriteByte(' ')
			sb.WriteString(bonus)
			if i < len(bonuses)-1 {
				sb.WriteString(", ")
			} else {
				sb.WriteByte('.')
			}
		}
		sb.WriteByte('\n')
	}
	fmt.Fprintf(&sb, "@B充能：@N %+d", sp.MaxCharges[1]-sp.MaxCharges[0])
	if sp.BonusTraits[1] != sp.BonusTraits[0] {
		sb.WriteByte('\n')
		fmt.Fprintf(&sb, "@B新特性：@N %s。", g.TraitDesc(Player, sp.BonusTraits[1]&^sp.BonusTraits[0]))
	}
	return desc + "\n\n" + sb.String()
}

func (sp *Spirit) Use(g *Game, i ID) bool {
	if !sp.Condition.CanInvoke() {
		g.Log("暂时无法召唤此灵体！")
		return false
	}
	if sp.Charges == 0 {
		g.Log("没有剩余充能。")
		return false
	}
	pa := g.PlayerActor()
	if pa.Has(StatusDaze) && !pa.DoesAny(ResistanceDaze) {
		g.Log("眩晕时无法召唤灵体。")
		return false
	}
	confused := pa.Has(StatusConfusion)
	if confused && pa.HP <= 2 {
		g.Log("混乱且濒死（生命值 < 3）时无法召唤灵体。")
		return false
	}
	ohp := pa.HP
	done := sp.GetAbility().Apply(g)
	if !done {
		return false
	}
	sp.Charges = max(0, sp.Charges-1)
	sp.Uses++
	g.Stats.SpiritUses++
	g.Stats.MapSpiritUses[g.Map.Level-1]++
	name := sp.GetAbility().Name()
	if confused {
		g.LogfStyled("混乱时召唤灵体会使你受到 %d 点伤害。", logHurtPlayer, ConfusionDamage)
		g.InflictDamage(PlayerID, pa, ConfusionDamage, AttackOther)
		g.StoryLogf("在迷向时使用“%s”能力（生命值：%d/%d，充能：%d/%d）",
			name, pa.HP, pa.GetMaxHP(), sp.Charges, sp.GetMaxCharges())
	} else if hp := pa.HP; hp == ohp {
		g.StoryLogf("使用“%s”能力（充能：%d/%d）", name, sp.Charges, sp.GetMaxCharges())
	} else {
		g.StoryLogf("使用“%s”能力（生命值：%d/%d，充能：%d/%d）", name, hp, pa.GetMaxHP(), sp.Charges, sp.GetMaxCharges())
	}
	return true
}

// GetMaxCharges returns the maximum number of charges given the current spirit
// upgrade level.
func (sp *Spirit) GetMaxCharges() int {
	return sp.MaxCharges[sp.Level]
}

// GetAbility returns the maximum number of charges given the current spirit
// upgrade level.
func (sp *Spirit) GetAbility() Ability {
	return sp.Ability[sp.Level]
}

const ConfusionDamage = 1

// EmptyTotem represents a totem without spirit.
type EmptyTotem struct{}

func (sp *EmptyTotem) Desc(_ *Game) string {
	return "图腾中没有灵体，或灵体已经离开。"
}

func (sp *EmptyTotem) Use(g *Game, i ID) bool {
	g.Log("无法从空图腾中装备灵体。")
	return false
}

// Comestible represents a collectable comestible item.
type Comestible struct {
	Effect Effect // effect on use
}

func (c *Comestible) Desc(g *Game) string {
	return c.Effect.Desc(g)
}

func (c *Comestible) ExamineDesc(g *Game) string {
	return fmt.Sprintf("@C食物。@N\n%s", c.Effect.Desc(g))
}

func (c *Comestible) Use(g *Game, i ID) bool {
	ef := c.Effect
	if g.Mod(ModTotemConditions) {
		co, cond := ef.(ConditionalComestible)
		if cond && !g.snack && !g.PlayerActor().DoesAny(Gluttony) && !co.CanApply(g, i) {
			// Unless the player is the Bear, forbid eating unless
			// condition is met.
			// NOTE: We have to check both snack and Gluttony above
			// because some curses may disable the latter, but we
			// want snack to still ignore requirements even then.
			g.Log(CantEatMsg(ef))
			return false
		}
	}
	pp := g.PP()
	done := ef.Apply(g)
	if !done {
		// Should never happen with comestibles.
		log.Printf("错误：未能吃下%s。", g.Entity(i).Name)
		return false
	}
	g.Stats.Comestibles[g.Entity(i).Name]++
	g.Entities[i] = emptySlot()
	g.Stats.EatenComestibles++
	g.Stats.MapEatenComestibles[g.Map.Level-1]++
	g.handleGluttonyOnEating()
	for i, ai := range g.Monsters() {
		if ai.Is(HungryRat) && g.InFOV(g.Entity(i).P) {
			g.PutStatus(i, ai, StatusBerserk, DurationBerserk)
			// Make monster hunt in case it was still wandering
			// (with no extra log message).
			ai.Behavior.State = Hunting
			ai.Behavior.Target = pp
		}
	}
	return true
}

func (g *Game) handleGluttonyOnEating() {
	if GluttonyRework {
		return
	}
	switch pa := g.PlayerActor(); {
	case pa.Has(StatusGluttony):
		pa.Statuses[StatusGluttony] = 0
		g.LogStyled("你感到饱了。", logStatusEnd)
	case pa.DoesAny(Gluttony) && !g.snack:
		g.PutStatus1(PlayerID, pa, StatusGluttony, DurationGluttony)
	}
}

// EquipItemAt equips the item currently in the ground at index i.
func (g *Game) EquipItemAt(i ID) bool {
	ei := g.Entity(i)
	j, it := g.ItemAt(g.PP())
	if j <= 0 {
		// should not happen
		g.LogStyled("无法装备此物品（错误）。", logError)
		return false
	}
	ej := g.Entity(j)
	switch it := it.(type) {
	case *Spirit:
		pa := g.PlayerActor()
		ohpmax := pa.MaxHP
		itc := it.Condition // nil if ModTotemConditions is not enabled
		if spi, ok := ei.Role.(*Spirit); ok {
			if spi.Level >= 1 {
				// Should not happen.
				g.Log("灵体已无法继续升级。")
				return false
			}
			spi.Level++
			spi.Charges += spi.MaxCharges[1] - spi.MaxCharges[0]
			if itc != nil && itc.Upgrade.Applies(spi.Condition != nil) {
				g.recordCurse(ei.Name, itc)
				spi.Condition = itc
			}
			g.Logf("你升级了%s灵体。", ei.Name)
			g.StoryLogf("已升级%s灵体", ei.Name)
		} else {
			if itc != nil {
				g.recordCurse(ej.Name, itc)
			}
			g.Entities[i], g.Entities[j] = g.Entities[j], g.Entities[i]
			g.Logf("你选择了%s灵体。", ej.Name)
			g.StoryLogf("已选择%s灵体", ej.Name)
		}
		g.ComputePlayerStats()
		if hpmaxDelta := pa.MaxHP - ohpmax; hpmaxDelta > 0 {
			g.AdjustHP(PlayerID, pa, hpmaxDelta)
		}
		ej.P = InvalidPos
		et := g.Entity(g.genEmptyTotemAt(g.PP()))
		et.Seen = true
		g.totemPickupBerserk()
		return true
	case *Comestible:
		if ei.IsItem() {
			g.Logf("你拾起%s（替换%s）。", ej.Name, ei.Name)
			g.StoryLogf("已拾起%s（替换%s）", One(ej.Name), ei.Name)
			ei.P = ej.P
		} else {
			g.Logf("你拾起%s。", ej.Name)
			g.StoryLogf("已拾起%s（食物：%d）", One(ej.Name), g.ComestibleCount()+1)
		}
		ej.P = InvalidPos
		g.Entities[i], g.Entities[j] = g.Entities[j], g.Entities[i]
		TotemEvents(g, &EventHappened{EvType: EventPickup})
		return true
	default:
		panic("意外的物品类型")
	}
}

// totemPickupBerserk makes monsters in view berserk (after you pick a totemic
// spirit).
func (g *Game) totemPickupBerserk() {
	berserk := false
	for i, ai := range g.Monsters() {
		ei := g.Entity(i)
		if !g.InFOV(ei.P) {
			continue
		}
		if g.PutStatus(i, ai, StatusBerserk, DurationBerserkTotem) {
			berserk = true
		}
	}
	if berserk {
		g.LogStyled("拾取灵体会使视野中的怪物陷入狂暴！", logSpecial)
	}
}

// Menhir represent a static magical menhir stone.
type Menhir struct {
	Used   bool   // whether already activated
	Effect Effect // effect on use
}

func (m *Menhir) Desc(g *Game) string {
	if !m.Used {
		return fmt.Sprintf("@C可激活。@N\n%s", m.Effect.Desc(g))
	}
	return m.Effect.Desc(g)
}

func (m *Menhir) Use(g *Game, i ID) bool {
	if m.Used {
		g.Log("石碑已经失去能量。")
		return false
	}
	if g.WarnLeave(EventMenhir) {
		g.md.mode = modeUseConfirmation
		return false
	}
	done := m.Effect.Apply(g)
	if !done {
		return false
	}
	g.StoryLogf("已激活%s", One(g.Entity(i).Name))
	m.Used = true
	TotemEvents(g, &EventHappened{EvType: EventMenhir})
	g.Stats.ActivatedMenhirs++
	g.Stats.MapActivatedMenhirs[g.Map.Level-1]++
	return true
}

// Portal represent a magical portal leading to the next level.
type Portal struct {
	Fake bool // whether the portal is fake
	Used bool // whether fake but already activated
}

func (p *Portal) Desc(_ *Game) string {
	if p.Used {
		return "外表正常的魔法传送门发生了故障，无法通往任何地方。"
	}
	return fmt.Sprintf("@C可激活。@N\n%s", "通往下一层的魔法传送门。可补满灵体充能和生命值。")
}

func (p *Portal) Use(g *Game, i ID) bool {
	if p.Used {
		g.Log("传送门已发生故障，无法再次触发。")
		return false
	}
	if p.Fake {
		p.Used = true
		g.useFakePortal()
		return true
	}
	if err := g.Save(); err != nil {
		log.Printf("进入新一层前保存游戏失败：%v", err)
	}
	g.StoryLog("进入了魔法传送门！")
	g.LogStyled("你进入传送门。", logSpecial)
	g.NextLevel()
	return false
}

func (g *Game) useFakePortal() {
	g.LogStyled("你尝试激活魔法传送门……", logSpecial)
	g.StoryLog("触发了故障传送门")
	const maxdist = MaxFOVRange + 3
	mp := &MapPath{passable: g.Map.Passable}
	g.PR.BreadthFirstMap(mp, []gruid.Point{g.PP()}, maxdist)
	g.MakeNoise(g.PP(), NoiseFakePortal)
	g.LogStyled("传来一阵诡异的声音，但什么也没有发生！", logSpecial)
	TotemEvents(g, &EventHappened{EvType: EventFakePortal})
	for i, ai := range g.Monsters() {
		ei := g.Entity(i)
		if g.PR.BreadthFirstMapAt(ei.P) <= maxdist {
			g.PutStatus(i, ai, StatusFear, DurationFearFakePortal)
		}
	}
}

// NextLevel proceeds to the next level (assuming we're not at the last one).
func (g *Game) NextLevel() {
	TotemEvents(g, &EventHappened{EvType: EventPortal})
	// We return false because after generating a new level we don't want
	// to end the turn. However, we still upgrade the turn counter, as it
	// wouldn't be intuitive otherwise.
	g.IncrTurn()   // call before InitLevel
	g.LevelStats() // call before InitLevel
	clear(g.Map.Noise)
	g.InitLevel()
	g.Logs.NextTick = g.Logs.Index
	g.md.targ.CancelExamine()
	TotemEvents(g, &EventHappened{EvType: EventNewLevel})
}

// CorruptionOrb represents the orb of corruption.
type CorruptionOrb struct {
	Broken bool // whether you broke the orb of corruption or not
}

func (o *CorruptionOrb) Desc(_ *Game) string {
	if o.Broken {
		// Should not happen in practice for now, unless we add extra
		// content after destroying the orb.
		return "你摧毁的野兽腐化源头。"
	}
	return fmt.Sprintf("@C可激活。@N\n%s", "你需要摧毁的野兽腐化源头。注意，传送时魔法球有时会主动把你传送得更远。")
}

func (o *CorruptionOrb) Use(g *Game, i ID) bool {
	o.Broken = true
	if g.Wizard.Mode == WizardNone {
		g.LogStyled("你摧毁了腐化之球……胜利！", logSpecial)
	} else {
		g.LogStyled("你摧毁了腐化之球……**向导模式**！", logSpecial)
	}
	g.StoryLog("摧毁了腐化魔法球！")
	g.md.OrbDestructionAnimation()
	g.LevelStats()
	g.md.logConfirmContinue()
	g.IncrTurn()
	g.win = true
	g.md.mode = modeEnd
	return false
}
