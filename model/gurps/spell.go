/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package gurps

import (
	"context"
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/entity"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/unison/enums/align"
)

var (
	_ Node[*Spell]                    = &Spell{}
	_ TechLevelProvider[*Spell]       = &Spell{}
	_ SkillAdjustmentProvider[*Spell] = &Spell{}
	_ TemplatePickerProvider          = &Spell{}
)

// Columns that can be used with the spell method .CellData()
const (
	SpellDescriptionColumn = iota
	SpellResistColumn
	SpellClassColumn
	SpellCollegeColumn
	SpellCastCostColumn
	SpellMaintainCostColumn
	SpellCastTimeColumn
	SpellDurationColumn
	SpellDifficultyColumn
	SpellTagsColumn
	SpellReferenceColumn
	SpellLevelColumn
	SpellRelativeLevelColumn
	SpellPointsColumn
	SpellDescriptionForPageColumn
)

const spellListTypeKey = "spell_list"

// Spell holds the data for a spell.
type Spell struct {
	SpellData
	Entity            *Entity
	LevelData         Level
	UnsatisfiedReason string
}

type spellListData struct {
	Type    string   `json:"type"`
	Version int      `json:"version"`
	Rows    []*Spell `json:"rows"`
}

// NewSpellsFromFile loads an Spell list from a file.
func NewSpellsFromFile(fileSystem fs.FS, filePath string) ([]*Spell, error) {
	var data spellListData
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &data); err != nil {
		return nil, errs.NewWithCause(invalidFileDataMsg(), err)
	}
	if data.Type != spellListTypeKey {
		return nil, errs.New(unexpectedFileDataMsg())
	}
	if err := CheckVersion(data.Version); err != nil {
		return nil, err
	}
	return data.Rows, nil
}

// SaveSpells writes the Spell list to the file as JSON.
func SaveSpells(spells []*Spell, filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, &spellListData{
		Type:    spellListTypeKey,
		Version: CurrentDataVersion,
		Rows:    spells,
	})
}

// NewSpell creates a new Spell.
func NewSpell(e *Entity, parent *Spell, container bool) *Spell {
	s := newSpell(e, parent, SpellID, container)
	s.UpdateLevel()
	return s
}

// NewRitualMagicSpell creates a new Ritual Magic Spell.
func NewRitualMagicSpell(e *Entity, parent *Spell, _ bool) *Spell {
	s := newSpell(e, parent, RitualMagicSpellID, false)
	s.RitualSkillName = "Ritual Magic"
	s.SetRawPoints(0)
	return s
}

func newSpell(e *Entity, parent *Spell, typeKey string, container bool) *Spell {
	s := Spell{
		SpellData: SpellData{
			ContainerBase: newContainerBase[*Spell](typeKey, container),
		},
		Entity: e,
	}
	s.parent = parent
	if container {
		s.TemplatePicker = &TemplatePicker{}
	} else {
		s.Difficulty.Attribute = AttributeIDFor(e, "iq")
		s.Difficulty.Difficulty = difficulty.Hard
		s.PowerSource = i18n.Text("Arcane")
		s.Class = i18n.Text("Regular")
		s.CastingCost = "1"
		s.CastingTime = "1 sec"
		s.Duration = "Instant"
		s.Points = fxp.One
	}
	s.Name = s.Kind()
	return &s
}

// Clone implements Node.
func (s *Spell) Clone(e *Entity, parent *Spell, preserveID bool) *Spell {
	var other *Spell
	if s.Type == RitualMagicSpellID {
		other = NewRitualMagicSpell(e, parent, false)
	} else {
		other = NewSpell(e, parent, s.Container())
		other.IsOpen = s.IsOpen
	}
	if preserveID {
		other.ID = s.ID
	}
	other.ThirdParty = s.ThirdParty
	other.SpellEditData.CopyFrom(s)
	if s.HasChildren() {
		other.Children = make([]*Spell, 0, len(s.Children))
		for _, child := range s.Children {
			other.Children = append(other.Children, child.Clone(e, other, preserveID))
		}
	}
	return other
}

// MarshalJSON implements json.Marshaler.
func (s *Spell) MarshalJSON() ([]byte, error) {
	s.ClearUnusedFieldsForType()
	type calcLeast struct {
		ResolvedNotes string `json:"resolved_notes,omitempty"`
	}
	var cl calcLeast
	notes := s.resolveLocalNotes()
	if notes != s.LocalNotes {
		cl.ResolvedNotes = notes
	}
	if s.Container() || s.LevelData.Level <= 0 {
		value := &struct {
			SpellData
			Calc *calcLeast `json:"calc,omitempty"`
		}{
			SpellData: s.SpellData,
		}
		if cl != (calcLeast{}) {
			value.Calc = &cl
		}
		return json.Marshal(value)
	}
	type calc struct {
		Level              fxp.Int `json:"level"`
		RelativeSkillLevel string  `json:"rsl"`
		UnsatisfiedReason  string  `json:"unsatisfied_reason,omitempty"`
		calcLeast
	}
	data := struct {
		SpellData
		Calc calc `json:"calc"`
	}{
		SpellData: s.SpellData,
		Calc: calc{
			Level:              s.LevelData.Level,
			RelativeSkillLevel: s.RelativeLevel(),
			UnsatisfiedReason:  s.UnsatisfiedReason,
			calcLeast:          cl,
		},
	}
	return json.Marshal(&data)
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *Spell) UnmarshalJSON(data []byte) error {
	var localData struct {
		SpellData
		// Old data fields
		Categories []string `json:"categories"`
	}
	if err := json.Unmarshal(data, &localData); err != nil {
		return err
	}
	localData.ClearUnusedFieldsForType()
	s.SpellData = localData.SpellData
	s.Tags = convertOldCategoriesToTags(s.Tags, localData.Categories)
	slices.Sort(s.Tags)
	if s.Container() {
		for _, one := range s.Children {
			one.parent = s
		}
	}
	return nil
}

// TemplatePickerData returns the TemplatePicker data, if any.
func (s *Spell) TemplatePickerData() *TemplatePicker {
	return s.TemplatePicker
}

// SpellHeaderData returns the header data information for the given spell column.
func SpellHeaderData(columnID int) HeaderData {
	var data HeaderData
	switch columnID {
	case SpellDescriptionColumn, SpellDescriptionForPageColumn:
		data.Title = i18n.Text("Spell")
		data.Primary = true
	case SpellResistColumn:
		data.Title = i18n.Text("Resist")
		data.Detail = i18n.Text("Resistance")
	case SpellClassColumn:
		data.Title = i18n.Text("Class")
	case SpellCollegeColumn:
		data.Title = i18n.Text("College")
	case SpellCastCostColumn:
		data.Title = i18n.Text("Cast")
		data.Detail = i18n.Text("The mana cost to cast the spell")
	case SpellMaintainCostColumn:
		data.Title = i18n.Text("Maintain")
		data.Detail = i18n.Text("The mana cost to maintain the spell")
	case SpellCastTimeColumn:
		data.Title = i18n.Text("Time")
		data.Detail = i18n.Text("The time required to cast the spell")
	case SpellDurationColumn:
		data.Title = i18n.Text("Duration")
	case SpellDifficultyColumn:
		data.Title = i18n.Text("Diff")
		data.Detail = i18n.Text("Difficulty")
	case SpellTagsColumn:
		data.Title = i18n.Text("Tags")
	case SpellReferenceColumn:
		data.Title = HeaderBookmark
		data.TitleIsImageKey = true
		data.Detail = PageRefTooltipText()
	case SpellLevelColumn:
		data.Title = i18n.Text("SL")
		data.Detail = i18n.Text("Skill Level")
	case SpellRelativeLevelColumn:
		data.Title = i18n.Text("RSL")
		data.Detail = i18n.Text("Relative Skill Level")
	case SpellPointsColumn:
		data.Title = i18n.Text("Pts")
		data.Detail = i18n.Text("Points")
	}
	return data
}

// CellData returns the cell data information for the given column.
func (s *Spell) CellData(columnID int, data *CellData) {
	switch columnID {
	case SpellDescriptionColumn:
		data.Type = cell.Text
		data.Primary = s.Description()
		data.Secondary = s.SecondaryText(func(option display.Option) bool { return option.Inline() })
		data.UnsatisfiedReason = s.UnsatisfiedReason
		data.Tooltip = s.SecondaryText(func(option display.Option) bool { return option.Tooltip() })
		data.TemplateInfo = s.TemplatePicker.Description()
	case SpellResistColumn:
		if !s.Container() {
			data.Type = cell.Text
			data.Primary = s.Resist
		}
	case SpellClassColumn:
		if !s.Container() {
			data.Type = cell.Text
			data.Primary = s.Class
		}
	case SpellCollegeColumn:
		if !s.Container() {
			data.Type = cell.Text
			data.Primary = strings.Join(s.College, ", ")
		}
	case SpellCastCostColumn:
		if !s.Container() {
			data.Type = cell.Text
			data.Primary = s.CastingCost
		}
	case SpellMaintainCostColumn:
		if !s.Container() {
			data.Type = cell.Text
			data.Primary = s.MaintenanceCost
		}
	case SpellCastTimeColumn:
		if !s.Container() {
			data.Type = cell.Text
			data.Primary = s.CastingTime
		}
	case SpellDurationColumn:
		if !s.Container() {
			data.Type = cell.Text
			data.Primary = s.Duration
		}
	case SpellDifficultyColumn:
		if !s.Container() {
			data.Type = cell.Text
			data.Primary = s.Difficulty.Description(s.Entity)
		}
	case SpellTagsColumn:
		data.Type = cell.Tags
		data.Primary = CombineTags(s.Tags)
	case SpellReferenceColumn, PageRefCellAlias:
		data.Type = cell.PageRef
		data.Primary = s.PageRef
		if s.PageRefHighlight != "" {
			data.Secondary = s.PageRefHighlight
		} else {
			data.Secondary = s.Name
		}
	case SpellLevelColumn:
		if !s.Container() {
			data.Type = cell.Text
			level := s.CalculateLevel()
			data.Primary = level.LevelAsString(s.Container())
			if level.Tooltip != "" {
				data.Tooltip = includesModifiersFrom() + ":" + level.Tooltip
			}
			data.Alignment = align.End
		}
	case SpellRelativeLevelColumn:
		if !s.Container() {
			data.Type = cell.Text
			rsl := s.AdjustedRelativeLevel()
			if rsl == fxp.Min {
				data.Primary = "-"
			} else {
				data.Primary = ResolveAttributeName(s.Entity, s.Difficulty.Attribute)
				if rsl != 0 {
					data.Primary += rsl.StringWithSign()
				}
			}
			if tooltip := s.CalculateLevel().Tooltip; tooltip != "" {
				data.Tooltip = includesModifiersFrom() + ":" + tooltip
			}
		}
	case SpellPointsColumn:
		data.Type = cell.Text
		var tooltip xio.ByteBuffer
		data.Primary = s.AdjustedPoints(&tooltip).String()
		data.Alignment = align.End
		if tooltip.Len() != 0 {
			data.Tooltip = includesModifiersFrom() + ":" + tooltip.String()
		}
	case SpellDescriptionForPageColumn:
		s.CellData(SpellDescriptionColumn, data)
		if !s.Container() {
			var buffer strings.Builder
			addPartToBuffer(&buffer, i18n.Text("Resistance"), s.Resist)
			addPartToBuffer(&buffer, i18n.Text("Class"), s.Class)
			addPartToBuffer(&buffer, i18n.Text("Cast"), s.CastingCost)
			addPartToBuffer(&buffer, i18n.Text("Maintain"), s.MaintenanceCost)
			addPartToBuffer(&buffer, i18n.Text("Time"), s.CastingTime)
			addPartToBuffer(&buffer, i18n.Text("Duration"), s.Duration)
			addPartToBuffer(&buffer, i18n.Text("College"), strings.Join(s.College, ", "))
			if buffer.Len() != 0 {
				if data.Secondary == "" {
					data.Secondary = buffer.String()
				} else {
					data.Secondary += "\n" + buffer.String()
				}
			}
		}
	}
}

func addPartToBuffer(buffer *strings.Builder, label, content string) {
	if content != "" && content != "-" {
		if buffer.Len() != 0 {
			buffer.WriteString("; ")
		}
		buffer.WriteString(label)
		buffer.WriteString(": ")
		buffer.WriteString(content)
	}
}

// Depth returns the number of parents this node has.
func (s *Spell) Depth() int {
	count := 0
	p := s.parent
	for p != nil {
		count++
		p = p.parent
	}
	return count
}

// RelativeLevel returns the adjusted relative level as a string.
func (s *Spell) RelativeLevel() string {
	if s.Container() || s.LevelData.Level <= 0 {
		return ""
	}
	rsl := s.AdjustedRelativeLevel()
	switch {
	case rsl == fxp.Min:
		return "-"
	case s.Type != RitualMagicSpellID:
		return ResolveAttributeName(s.Entity, s.Difficulty.Attribute) + rsl.StringWithSign()
	default:
		return rsl.StringWithSign()
	}
}

// AdjustedRelativeLevel returns the relative skill level.
func (s *Spell) AdjustedRelativeLevel() fxp.Int {
	if s.Container() {
		return fxp.Min
	}
	if s.Entity != nil && s.CalculateLevel().Level > 0 {
		return s.LevelData.RelativeLevel
	}
	return fxp.Min
}

// UpdateLevel updates the level of the spell, returning true if it has changed.
func (s *Spell) UpdateLevel() bool {
	saved := s.LevelData
	if strings.HasPrefix(s.Type, SpellID) {
		s.LevelData = CalculateSpellLevel(s.Entity, s.Name, s.PowerSource, s.College, s.Tags, s.Difficulty,
			s.AdjustedPoints(nil))
	} else {
		s.LevelData = CalculateRitualMagicSpellLevel(s.Entity, s.Name, s.PowerSource, s.RitualSkillName,
			s.RitualPrereqCount, s.College, s.Tags, s.Difficulty, s.AdjustedPoints(nil))
	}
	return saved != s.LevelData
}

// CalculateLevel returns the computed level without updating it.
func (s *Spell) CalculateLevel() Level {
	if strings.HasPrefix(s.Type, SpellID) {
		return CalculateSpellLevel(s.Entity, s.Name, s.PowerSource, s.College, s.Tags, s.Difficulty,
			s.AdjustedPoints(nil))
	}
	return CalculateRitualMagicSpellLevel(s.Entity, s.Name, s.PowerSource, s.RitualSkillName, s.RitualPrereqCount,
		s.College, s.Tags, s.Difficulty, s.AdjustedPoints(nil))
}

// IncrementSkillLevel adds enough points to increment the skill level to the next level.
func (s *Spell) IncrementSkillLevel() {
	if !s.Container() {
		basePoints := s.Points.Trunc() + fxp.One
		maxPoints := basePoints
		if s.Difficulty.Difficulty == difficulty.Wildcard {
			maxPoints += fxp.Twelve
		} else {
			maxPoints += fxp.Four
		}
		oldLevel := s.CalculateLevel().Level
		for points := basePoints; points < maxPoints; points += fxp.One {
			s.SetRawPoints(points)
			if s.CalculateLevel().Level > oldLevel {
				break
			}
		}
	}
}

// DecrementSkillLevel removes enough points to decrement the skill level to the previous level.
func (s *Spell) DecrementSkillLevel() {
	if !s.Container() && s.Points > 0 {
		basePoints := s.Points.Trunc()
		minPoints := basePoints
		if s.Difficulty.Difficulty == difficulty.Wildcard {
			minPoints -= fxp.Twelve
		} else {
			minPoints -= fxp.Four
		}
		minPoints = minPoints.Max(0)
		oldLevel := s.CalculateLevel().Level
		for points := basePoints; points >= minPoints; points -= fxp.One {
			s.SetRawPoints(points)
			if s.CalculateLevel().Level < oldLevel {
				break
			}
		}
		if s.Points > 0 {
			oldLevel = s.CalculateLevel().Level
			for s.Points > 0 {
				s.SetRawPoints((s.Points - fxp.One).Max(0))
				if s.CalculateLevel().Level != oldLevel {
					s.Points += fxp.One
					break
				}
			}
		}
	}
}

// CalculateSpellLevel returns the calculated spell level.
func CalculateSpellLevel(e *Entity, name, powerSource string, colleges, tags []string, attrDiff AttributeDifficulty, pts fxp.Int) Level {
	var tooltip xio.ByteBuffer
	relativeLevel := attrDiff.Difficulty.BaseRelativeLevel()
	level := fxp.Min
	if e != nil {
		pts = pts.Trunc()
		level = e.ResolveAttributeCurrent(attrDiff.Attribute)
		if attrDiff.Difficulty == difficulty.Wildcard {
			pts = pts.Div(fxp.Three).Trunc()
		}
		switch {
		case pts < fxp.One:
			level = fxp.Min
			relativeLevel = 0
		case pts == fxp.One:
		// relativeLevel is preset to this point value
		case pts < fxp.Four:
			relativeLevel += fxp.One
		default:
			relativeLevel += fxp.One + pts.Div(fxp.Four).Trunc()
		}
		if level != fxp.Min {
			relativeLevel += e.SpellBonusFor(name, powerSource, colleges, tags, &tooltip)
			relativeLevel = relativeLevel.Trunc()
			level += relativeLevel
		}
	}
	return Level{
		Level:         level,
		RelativeLevel: relativeLevel,
		Tooltip:       tooltip.String(),
	}
}

// CalculateRitualMagicSpellLevel returns the calculated spell level.
func CalculateRitualMagicSpellLevel(e *Entity, name, powerSource, ritualSkillName string, ritualPrereqCount int, colleges, tags []string, difficulty AttributeDifficulty, points fxp.Int) Level {
	var skillLevel Level
	if len(colleges) == 0 {
		skillLevel = determineRitualMagicSkillLevelForCollege(e, name, "", ritualSkillName, ritualPrereqCount,
			tags, difficulty, points)
	} else {
		for _, college := range colleges {
			possible := determineRitualMagicSkillLevelForCollege(e, name, college, ritualSkillName,
				ritualPrereqCount, tags, difficulty, points)
			if skillLevel.Level < possible.Level {
				skillLevel = possible
			}
		}
	}
	if e != nil {
		tooltip := &xio.ByteBuffer{}
		tooltip.WriteString(skillLevel.Tooltip)
		levels := e.SpellBonusFor(name, powerSource, colleges, tags, tooltip).Trunc()
		skillLevel.Level += levels
		skillLevel.RelativeLevel += levels
		skillLevel.Tooltip = tooltip.String()
	}
	return skillLevel
}

func determineRitualMagicSkillLevelForCollege(e *Entity, name, college, ritualSkillName string, ritualPrereqCount int, tags []string, difficulty AttributeDifficulty, points fxp.Int) Level {
	def := &SkillDefault{
		DefaultType:    SkillID,
		Name:           ritualSkillName,
		Specialization: college,
		Modifier:       fxp.From(-ritualPrereqCount),
	}
	if college == "" {
		def.Name = ""
	}
	var limit fxp.Int
	skillLevel := CalculateTechniqueLevel(e, name, college, tags, def, difficulty.Difficulty, points, false, &limit, nil)
	// CalculateTechniqueLevel() does not add the default skill modifier to the relative level, only to the final level
	skillLevel.RelativeLevel += def.Modifier
	def.Specialization = ""
	def.Modifier -= fxp.Six
	fallback := CalculateTechniqueLevel(e, name, college, tags, def, difficulty.Difficulty, points, false, &limit, nil)
	fallback.RelativeLevel += def.Modifier
	if skillLevel.Level >= fallback.Level {
		return skillLevel
	}
	return fallback
}

// RitualMagicSatisfied returns true if the Ritual Magic Spell is satisfied.
func (s *Spell) RitualMagicSatisfied(tooltip *xio.ByteBuffer, prefix string) bool {
	if s.Type != RitualMagicSpellID {
		return true
	}
	if len(s.College) == 0 {
		if tooltip != nil {
			tooltip.WriteString(prefix)
			tooltip.WriteString(i18n.Text("Must be assigned to a college"))
		}
		return false
	}
	for _, college := range s.College {
		if s.Entity.BestSkillNamed(s.RitualSkillName, college, false, nil) != nil {
			return true
		}
	}
	if s.Entity.BestSkillNamed(s.RitualSkillName, "", false, nil) != nil {
		return true
	}
	if tooltip != nil {
		tooltip.WriteString(prefix)
		tooltip.WriteString(i18n.Text("Requires a skill named "))
		tooltip.WriteString(s.RitualSkillName)
		tooltip.WriteString(" (")
		tooltip.WriteString(s.College[0])
		tooltip.WriteByte(')')
		for _, college := range s.College[1:] {
			tooltip.WriteString(i18n.Text(" or "))
			tooltip.WriteString(s.RitualSkillName)
			tooltip.WriteString(" (")
			tooltip.WriteString(college)
			tooltip.WriteByte(')')
		}
	}
	return false
}

// OwningEntity returns the owning Entity.
func (s *Spell) OwningEntity() *Entity {
	return s.Entity
}

// SetOwningEntity sets the owning entity and configures any sub-components as needed.
func (s *Spell) SetOwningEntity(e *Entity) {
	s.Entity = e
	if s.Container() {
		for _, child := range s.Children {
			child.SetOwningEntity(e)
		}
	} else {
		for _, w := range s.Weapons {
			w.SetOwner(s)
		}
	}
}

// Notes implements WeaponOwner.
func (s *Spell) Notes() string {
	return s.resolveLocalNotes()
}

// Rituals returns the rituals required to cast the spell.
func (s *Spell) Rituals() string {
	if s.Container() || !(s.Entity != nil && s.Entity.Type == entity.PC && s.Entity.SheetSettings.ShowSpellAdj) {
		return ""
	}
	level := s.CalculateLevel().Level
	switch {
	case level < fxp.Ten:
		return i18n.Text("Ritual: need both hands and feet free and must speak; Time: 2x")
	case level < fxp.Fifteen:
		return i18n.Text("Ritual: speak quietly and make a gesture")
	case level < fxp.Twenty:
		ritual := i18n.Text("Ritual: speak a word or two OR make a small gesture")
		if strings.Contains(strings.ToLower(s.Class), "blocking") {
			return ritual
		}
		return ritual + i18n.Text("; Cost: -1")
	default:
		adj := fxp.As[int]((level - fxp.Fifteen).Div(fxp.Five))
		class := strings.ToLower(s.Class)
		time := ""
		if !strings.Contains(class, "missile") {
			time = fmt.Sprintf(i18n.Text("; Time: x1/%d, rounded up, min 1 sec"), 1<<adj)
		}
		cost := ""
		if !strings.Contains(class, "blocking") {
			cost = fmt.Sprintf(i18n.Text("; Cost: -%d"), adj+1)
		}
		return i18n.Text("Ritual: none") + time + cost
	}
}

// FeatureList returns the list of Features.
func (s *Spell) FeatureList() Features {
	return nil
}

// TagList returns the list of tags.
func (s *Spell) TagList() []string {
	return s.Tags
}

// RatedStrength always return 0 for spells.
func (s *Spell) RatedStrength() fxp.Int {
	return 0
}

// Description implements WeaponOwner.
func (s *Spell) Description() string {
	return s.String()
}

// SecondaryText returns the less important information that should be displayed with the description.
func (s *Spell) SecondaryText(optionChecker func(display.Option) bool) string {
	var buffer strings.Builder
	prefs := SheetSettingsFor(s.Entity)
	if optionChecker(prefs.NotesDisplay) {
		AppendStringOntoNewLine(&buffer, strings.TrimSpace(s.Notes()))
		AppendStringOntoNewLine(&buffer, s.Rituals())
		AppendStringOntoNewLine(&buffer, StudyHoursProgressText(ResolveStudyHours(s.Study), s.StudyHoursNeeded, false))
	}
	if optionChecker(prefs.SkillLevelAdjDisplay) {
		if s.LevelData.Tooltip != "" && s.LevelData.Tooltip != noAdditionalModifiers() {
			levelTooltip := strings.ReplaceAll(strings.TrimSpace(s.LevelData.Tooltip), "\n", ", ")
			msg := includesModifiersFrom()
			if strings.HasPrefix(levelTooltip, msg+",") {
				levelTooltip = msg + ":" + levelTooltip[len(msg)+1:]
			}
			AppendStringOntoNewLine(&buffer, levelTooltip)
		}
	}
	return buffer.String()
}

func (s *Spell) String() string {
	var buffer strings.Builder
	buffer.WriteString(s.Name)
	if !s.Container() {
		if s.TechLevel != nil {
			buffer.WriteString("/TL")
			buffer.WriteString(*s.TechLevel)
		}
	}
	return buffer.String()
}

func (s *Spell) resolveLocalNotes() string {
	return EvalEmbeddedRegex.ReplaceAllStringFunc(s.LocalNotes, s.Entity.EmbeddedEval)
}

// RawPoints returns the unadjusted points.
func (s *Spell) RawPoints() fxp.Int {
	return s.Points
}

// SetRawPoints sets the unadjusted points and updates the level. Returns true if the level changed.
func (s *Spell) SetRawPoints(points fxp.Int) bool {
	s.Points = points
	return s.UpdateLevel()
}

// AdjustedPoints returns the points, adjusted for any bonuses.
func (s *Spell) AdjustedPoints(tooltip *xio.ByteBuffer) fxp.Int {
	if s.Container() {
		var total fxp.Int
		for _, one := range s.Children {
			total += one.AdjustedPoints(tooltip)
		}
		return total
	}
	return AdjustedPointsForNonContainerSpell(s.Entity, s.Points, s.Name, s.PowerSource, s.College, s.Tags, tooltip)
}

// AdjustedPointsForNonContainerSpell returns the points, adjusted for any bonuses.
func AdjustedPointsForNonContainerSpell(e *Entity, points fxp.Int, name, powerSource string, colleges, tags []string, tooltip *xio.ByteBuffer) fxp.Int {
	if e != nil && e.Type == entity.PC {
		points += e.SpellPointBonusFor(name, powerSource, colleges, tags, tooltip)
		points = points.Max(0)
	}
	return points
}

// TL implements TechLevelProvider.
func (s *Spell) TL() string {
	if s.TechLevel != nil {
		return *s.TechLevel
	}
	return ""
}

// RequiresTL implements TechLevelProvider.
func (s *Spell) RequiresTL() bool {
	return s.TechLevel != nil
}

// SetTL implements TechLevelProvider.
func (s *Spell) SetTL(tl string) {
	if s.TechLevel != nil {
		*s.TechLevel = tl
	}
}

// Enabled returns true if this node is enabled.
func (s *Spell) Enabled() bool {
	return true
}

// FillWithNameableKeys adds any nameable keys found to the provided map.
func (s *Spell) FillWithNameableKeys(m map[string]string) {
	Extract(s.Name, m)
	Extract(s.LocalNotes, m)
	Extract(s.PowerSource, m)
	Extract(s.Class, m)
	Extract(s.Resist, m)
	Extract(s.CastingCost, m)
	Extract(s.MaintenanceCost, m)
	Extract(s.CastingTime, m)
	Extract(s.Duration, m)
	Extract(s.RitualSkillName, m)
	for _, one := range s.College {
		Extract(one, m)
	}
	if s.Prereq != nil {
		s.Prereq.FillWithNameableKeys(m)
	}
	for _, one := range s.Weapons {
		one.FillWithNameableKeys(m)
	}
}

// ApplyNameableKeys replaces any nameable keys found with the corresponding values in the provided map.
func (s *Spell) ApplyNameableKeys(m map[string]string) {
	s.Name = Apply(s.Name, m)
	s.LocalNotes = Apply(s.LocalNotes, m)
	s.PowerSource = Apply(s.PowerSource, m)
	s.Class = Apply(s.Class, m)
	s.Resist = Apply(s.Resist, m)
	s.CastingCost = Apply(s.CastingCost, m)
	s.MaintenanceCost = Apply(s.MaintenanceCost, m)
	s.CastingTime = Apply(s.CastingTime, m)
	s.Duration = Apply(s.Duration, m)
	s.RitualSkillName = Apply(s.RitualSkillName, m)
	for i, one := range s.College {
		s.College[i] = Apply(one, m)
	}
	if s.Prereq != nil {
		s.Prereq.ApplyNameableKeys(m)
	}
	for _, one := range s.Weapons {
		one.ApplyNameableKeys(m)
	}
}
