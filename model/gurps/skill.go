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
	_ Node[*Skill]                    = &Skill{}
	_ TechLevelProvider[*Skill]       = &Skill{}
	_ SkillAdjustmentProvider[*Skill] = &Skill{}
	_ TemplatePickerProvider          = &Skill{}
)

// Columns that can be used with the skill method .CellData()
const (
	SkillDescriptionColumn = iota
	SkillDifficultyColumn
	SkillTagsColumn
	SkillReferenceColumn
	SkillLevelColumn
	SkillRelativeLevelColumn
	SkillPointsColumn
)

const skillListTypeKey = "skill_list"

// Skill holds the data for a skill.
type Skill struct {
	SkillData
	Entity            *Entity
	LevelData         Level
	UnsatisfiedReason string
}

type skillListData struct {
	Type    string   `json:"type"`
	Version int      `json:"version"`
	Rows    []*Skill `json:"rows"`
}

// NewSkillsFromFile loads an Skill list from a file.
func NewSkillsFromFile(fileSystem fs.FS, filePath string) ([]*Skill, error) {
	var data skillListData
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &data); err != nil {
		return nil, errs.NewWithCause(invalidFileDataMsg(), err)
	}
	if data.Type != skillListTypeKey {
		return nil, errs.New(unexpectedFileDataMsg())
	}
	if err := CheckVersion(data.Version); err != nil {
		return nil, err
	}

	// Fix up some bad data in standalone skill lists where Hard techniques incorrectly had 1 point assigned to them
	// instead of 2.
	Traverse(func(skill *Skill) bool {
		if strings.HasPrefix(skill.Type, TechniqueID) &&
			skill.Difficulty.Difficulty == difficulty.Hard &&
			skill.Points == fxp.One {
			skill.Points = fxp.Two
		}
		return false
	}, false, true, data.Rows...)

	return data.Rows, nil
}

// SaveSkills writes the Skill list to the file as JSON.
func SaveSkills(skills []*Skill, filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, &skillListData{
		Type:    skillListTypeKey,
		Version: CurrentDataVersion,
		Rows:    skills,
	})
}

// NewSkill creates a new Skill.
func NewSkill(e *Entity, parent *Skill, container bool) *Skill {
	return newSkill(e, parent, SkillID, container)
}

// NewTechnique creates a new technique (i.e. a specialized use of a Skill). All parameters may be nil or empty.
func NewTechnique(e *Entity, parent *Skill, skillName string) *Skill {
	if skillName == "" {
		skillName = i18n.Text("Skill")
	}
	s := newSkill(e, parent, TechniqueID, false)
	s.TechniqueDefault = &SkillDefault{
		DefaultType: SkillID,
		Name:        skillName,
	}
	return s
}

func newSkill(e *Entity, parent *Skill, typeKey string, container bool) *Skill {
	s := Skill{
		SkillData: SkillData{
			ContainerBase: newContainerBase[*Skill](typeKey, container),
		},
		Entity: e,
	}
	s.parent = parent
	if container {
		s.TemplatePicker = &TemplatePicker{}
	} else {
		if typeKey != TechniqueID {
			s.Difficulty.Attribute = AttributeIDFor(e, DexterityID)
		}
		s.Difficulty.Difficulty = difficulty.Average
		s.Points = fxp.One
	}
	s.Name = s.Kind()
	return &s
}

// Clone implements Node.
func (s *Skill) Clone(e *Entity, parent *Skill, preserveID bool) *Skill {
	var other *Skill
	if s.Type == TechniqueID {
		other = NewTechnique(e, parent, s.TechniqueDefault.Name)
	} else {
		other = NewSkill(e, parent, s.Container())
		other.IsOpen = s.IsOpen
	}
	if preserveID {
		other.ID = s.ID
	}
	other.ThirdParty = s.ThirdParty
	other.SkillEditData.CopyFrom(s)
	if s.HasChildren() {
		other.Children = make([]*Skill, 0, len(s.Children))
		for _, child := range s.Children {
			other.Children = append(other.Children, child.Clone(e, other, preserveID))
		}
	}
	return other
}

// MarshalJSON implements json.Marshaler.
func (s *Skill) MarshalJSON() ([]byte, error) {
	s.ClearUnusedFieldsForType()
	type calcNoLevel struct {
		ResolvedNotes     string `json:"resolved_notes,omitempty"`
		UnsatisfiedReason string `json:"unsatisfied_reason,omitempty"`
	}
	cnl := calcNoLevel{UnsatisfiedReason: s.UnsatisfiedReason}
	notes := s.resolveLocalNotes()
	if notes != s.LocalNotes {
		cnl.ResolvedNotes = notes
	}
	if s.Container() || s.LevelData.Level <= 0 {
		value := &struct {
			SkillData
			Calc *calcNoLevel `json:"calc,omitempty"`
		}{
			SkillData: s.SkillData,
		}
		if cnl != (calcNoLevel{}) {
			value.Calc = &cnl
		}
		return json.Marshal(value)
	}
	type calc struct {
		Level              fxp.Int `json:"level"`
		RelativeSkillLevel string  `json:"rsl"`
		calcNoLevel
	}
	return json.Marshal(&struct {
		SkillData
		Calc calc `json:"calc"`
	}{
		SkillData: s.SkillData,
		Calc: calc{
			Level:              s.LevelData.Level,
			RelativeSkillLevel: s.RelativeLevel(),
			calcNoLevel:        cnl,
		},
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *Skill) UnmarshalJSON(data []byte) error {
	var localData struct {
		SkillData
		// Old data fields
		Categories []string `json:"categories"`
	}
	if err := json.Unmarshal(data, &localData); err != nil {
		return err
	}
	localData.ClearUnusedFieldsForType()
	s.SkillData = localData.SkillData
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
func (s *Skill) TemplatePickerData() *TemplatePicker {
	return s.TemplatePicker
}

// SkillsHeaderData returns the header data information for the given skill column.
func SkillsHeaderData(columnID int) HeaderData {
	var data HeaderData
	switch columnID {
	case SkillDescriptionColumn:
		data.Title = i18n.Text("Skill / Technique")
		data.Primary = true
	case SkillDifficultyColumn:
		data.Title = i18n.Text("Diff")
		data.Detail = i18n.Text("Difficulty")
	case SkillTagsColumn:
		data.Title = i18n.Text("Tags")
	case SkillReferenceColumn:
		data.Title = HeaderBookmark
		data.TitleIsImageKey = true
		data.Detail = PageRefTooltipText()
	case SkillLevelColumn:
		data.Title = i18n.Text("SL")
		data.Detail = i18n.Text("Skill Level")
	case SkillRelativeLevelColumn:
		data.Title = i18n.Text("RSL")
		data.Detail = i18n.Text("Relative Skill Level")
	case SkillPointsColumn:
		data.Title = i18n.Text("Pts")
		data.Detail = i18n.Text("Points")
	}
	return data
}

// CellData returns the cell data information for the given column.
func (s *Skill) CellData(columnID int, data *CellData) {
	switch columnID {
	case SkillDescriptionColumn:
		data.Type = cell.Text
		data.Primary = s.Description()
		data.Secondary = s.SecondaryText(func(option display.Option) bool { return option.Inline() })
		data.UnsatisfiedReason = s.UnsatisfiedReason
		data.Tooltip = s.SecondaryText(func(option display.Option) bool { return option.Tooltip() })
		data.TemplateInfo = s.TemplatePicker.Description()
	case SkillDifficultyColumn:
		if !s.Container() {
			data.Type = cell.Text
			data.Primary = s.Difficulty.Description(s.Entity)
		}
	case SkillTagsColumn:
		data.Type = cell.Tags
		data.Primary = CombineTags(s.Tags)
	case SkillReferenceColumn, PageRefCellAlias:
		data.Type = cell.PageRef
		data.Primary = s.PageRef
		if s.PageRefHighlight != "" {
			data.Secondary = s.PageRefHighlight
		} else {
			data.Secondary = s.Name
		}
	case SkillLevelColumn:
		if !s.Container() {
			data.Type = cell.Text
			level := s.CalculateLevel(nil)
			data.Primary = level.LevelAsString(s.Container())
			if level.Tooltip != "" {
				data.Tooltip = includesModifiersFrom() + ":" + level.Tooltip
			}
			data.Alignment = align.End
		}
	case SkillRelativeLevelColumn:
		if !s.Container() {
			data.Type = cell.Text
			data.Primary = FormatRelativeSkill(s.Entity, s.Type, s.Difficulty, s.AdjustedRelativeLevel())
			if tooltip := s.CalculateLevel(nil).Tooltip; tooltip != "" {
				data.Tooltip = includesModifiersFrom() + ":" + tooltip
			}
		}
	case SkillPointsColumn:
		data.Type = cell.Text
		var tooltip xio.ByteBuffer
		data.Primary = s.AdjustedPoints(&tooltip).String()
		data.Alignment = align.End
		if tooltip.Len() != 0 {
			data.Tooltip = includesModifiersFrom() + ":" + tooltip.String()
		}
	}
}

// FormatRelativeSkill formats the relative skill for display.
func FormatRelativeSkill(e *Entity, typ string, difficulty AttributeDifficulty, rsl fxp.Int) string {
	switch {
	case rsl == fxp.Min:
		return "-"
	case strings.HasPrefix(typ, SkillID) || strings.HasPrefix(typ, SpellID):
		s := ResolveAttributeName(e, difficulty.Attribute)
		rsl = rsl.Trunc()
		if rsl != 0 {
			s += rsl.StringWithSign()
		}
		return s
	default:
		return rsl.Trunc().StringWithSign()
	}
}

// Depth returns the number of parents this node has.
func (s *Skill) Depth() int {
	count := 0
	p := s.parent
	for p != nil {
		count++
		p = p.parent
	}
	return count
}

// OwningEntity returns the owning Entity.
func (s *Skill) OwningEntity() *Entity {
	return s.Entity
}

// SetOwningEntity sets the owning entity and configures any sub-components as needed.
func (s *Skill) SetOwningEntity(e *Entity) {
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

// DefaultSkill returns the skill currently defaulted to, or nil.
func (s *Skill) DefaultSkill() *Skill {
	if s.Entity == nil {
		return nil
	}
	if strings.HasPrefix(s.Type, TechniqueID) {
		return s.Entity.BaseSkill(s.TechniqueDefault, true)
	}
	return s.Entity.BaseSkill(s.DefaultedFrom, true)
}

// HasDefaultTo returns true if the set of possible defaults includes the other skill.
func (s *Skill) HasDefaultTo(other *Skill) bool {
	for _, def := range s.resolveToSpecificDefaults() {
		if def.SkillBased() && def.Name == other.Name && (def.Specialization == "" || def.Specialization == other.Specialization) {
			return true
		}
	}
	return false
}

// Notes implements WeaponOwner.
func (s *Skill) Notes() string {
	return s.resolveLocalNotes()
}

// ModifierNotes returns the notes due to modifiers.
func (s *Skill) ModifierNotes() string {
	if strings.HasPrefix(s.Type, TechniqueID) {
		return i18n.Text("Default: ") + s.TechniqueDefault.FullName(s.Entity) + s.TechniqueDefault.ModifierAsString()
	}
	if s.Difficulty.Difficulty != difficulty.Wildcard {
		defSkill := s.DefaultSkill()
		if defSkill != nil && s.DefaultedFrom != nil {
			return i18n.Text("Default: ") + defSkill.String() + s.DefaultedFrom.ModifierAsString()
		}
	}
	return ""
}

// FeatureList returns the list of Features.
func (s *Skill) FeatureList() Features {
	return s.Features
}

// TagList returns the list of tags.
func (s *Skill) TagList() []string {
	return s.Tags
}

// RatedStrength always return 0 for skills.
func (s *Skill) RatedStrength() fxp.Int {
	return 0
}

// Description implements WeaponOwner.
func (s *Skill) Description() string {
	return s.String()
}

// SecondaryText returns the less important information that should be displayed with the description.
func (s *Skill) SecondaryText(optionChecker func(display.Option) bool) string {
	var buffer strings.Builder
	prefs := SheetSettingsFor(s.Entity)
	if optionChecker(prefs.ModifiersDisplay) {
		text := s.ModifierNotes()
		if strings.TrimSpace(text) != "" {
			buffer.WriteString(text)
		}
	}
	if optionChecker(prefs.NotesDisplay) {
		AppendStringOntoNewLine(&buffer, strings.TrimSpace(s.Notes()))
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

func (s *Skill) String() string {
	var buffer strings.Builder
	buffer.WriteString(s.Name)
	if !s.Container() {
		if s.TechLevel != nil {
			buffer.WriteString("/TL")
			buffer.WriteString(*s.TechLevel)
		}
		if s.Specialization != "" {
			buffer.WriteString(" (")
			buffer.WriteString(s.Specialization)
			buffer.WriteByte(')')
		}
	}
	return buffer.String()
}

func (s *Skill) resolveLocalNotes() string {
	return EvalEmbeddedRegex.ReplaceAllStringFunc(s.LocalNotes, s.Entity.EmbeddedEval)
}

// RelativeLevel returns the adjusted relative level as a string.
func (s *Skill) RelativeLevel() string {
	if s.Container() || s.LevelData.Level <= 0 {
		return ""
	}
	rsl := s.AdjustedRelativeLevel()
	switch {
	case rsl == fxp.Min:
		return "-"
	case strings.HasPrefix(s.Type, SkillID):
		return ResolveAttributeName(s.Entity, s.Difficulty.Attribute) + rsl.StringWithSign()
	default:
		return rsl.StringWithSign()
	}
}

// AdjustedRelativeLevel returns the relative skill level.
func (s *Skill) AdjustedRelativeLevel() fxp.Int {
	if s.Container() {
		return fxp.Min
	}
	if s.Entity != nil && s.LevelData.Level > 0 {
		if strings.HasPrefix(s.Type, TechniqueID) {
			return s.LevelData.RelativeLevel + s.TechniqueDefault.Modifier
		}
		return s.LevelData.RelativeLevel
	}
	return fxp.Min
}

// RawPoints returns the unadjusted points.
func (s *Skill) RawPoints() fxp.Int {
	return s.Points
}

// SetRawPoints sets the unadjusted points and updates the level. Returns true if the level changed.
func (s *Skill) SetRawPoints(points fxp.Int) bool {
	s.Points = points
	return s.UpdateLevel()
}

// AdjustedPoints returns the points, adjusted for any bonuses.
func (s *Skill) AdjustedPoints(tooltip *xio.ByteBuffer) fxp.Int {
	if s.Container() {
		var total fxp.Int
		for _, one := range s.Children {
			total += one.AdjustedPoints(tooltip)
		}
		return total
	}
	return AdjustedPointsForNonContainerSkillOrTechnique(s.Entity, s.Points, s.Name, s.Specialization, s.Tags, tooltip)
}

// AdjustedPointsForNonContainerSkillOrTechnique returns the points, adjusted for any bonuses.
func AdjustedPointsForNonContainerSkillOrTechnique(e *Entity, points fxp.Int, name, specialization string, tags []string, tooltip *xio.ByteBuffer) fxp.Int {
	if e != nil && e.Type == entity.PC {
		points += e.SkillPointBonusFor(name, specialization, tags, tooltip)
		points = points.Max(0)
	}
	return points
}

// IncrementSkillLevel adds enough points to increment the skill level to the next level.
func (s *Skill) IncrementSkillLevel() {
	if !s.Container() {
		basePoints := s.Points.Trunc() + fxp.One
		maxPoints := basePoints
		if s.Difficulty.Difficulty == difficulty.Wildcard {
			maxPoints += fxp.Twelve
		} else {
			maxPoints += fxp.Four
		}
		oldLevel := s.CalculateLevel(nil).Level
		for points := basePoints; points < maxPoints; points += fxp.One {
			s.SetRawPoints(points)
			if s.CalculateLevel(nil).Level > oldLevel {
				break
			}
		}
	}
}

// DecrementSkillLevel removes enough points to decrement the skill level to the previous level.
func (s *Skill) DecrementSkillLevel() {
	if !s.Container() && s.Points > 0 {
		basePoints := s.Points.Trunc()
		minPoints := basePoints
		if s.Difficulty.Difficulty == difficulty.Wildcard {
			minPoints -= fxp.Twelve
		} else {
			minPoints -= fxp.Four
		}
		minPoints = minPoints.Max(0)
		oldLevel := s.CalculateLevel(nil).Level
		for points := basePoints; points >= minPoints; points -= fxp.One {
			s.SetRawPoints(points)
			if s.CalculateLevel(nil).Level < oldLevel {
				break
			}
		}
		if s.Points > 0 {
			oldLevel = s.CalculateLevel(nil).Level
			for s.Points > 0 {
				s.SetRawPoints((s.Points - fxp.One).Max(0))
				if s.CalculateLevel(nil).Level != oldLevel {
					s.Points += fxp.One
					break
				}
			}
		}
	}
}

// CalculateLevel returns the computed level without updating it.
func (s *Skill) CalculateLevel(excludes map[string]bool) Level {
	points := s.AdjustedPoints(nil)
	if strings.HasPrefix(s.Type, SkillID) {
		return CalculateSkillLevel(s.Entity, s.Name, s.Specialization, s.Tags, s.DefaultedFrom, s.Difficulty, points,
			s.EncumbrancePenaltyMultiplier)
	}
	return CalculateTechniqueLevel(s.Entity, s.Name, s.Specialization, s.Tags, s.TechniqueDefault,
		s.Difficulty.Difficulty, points, true, s.TechniqueLimitModifier, excludes)
}

// CalculateSkillLevel returns the calculated level for a skill.
func CalculateSkillLevel(e *Entity, name, specialization string, tags []string, def *SkillDefault, attrDiff AttributeDifficulty, points, encumbrancePenaltyMultiplier fxp.Int) Level {
	var tooltip xio.ByteBuffer
	relativeLevel := attrDiff.Difficulty.BaseRelativeLevel()
	level := e.ResolveAttributeCurrent(attrDiff.Attribute)
	if level != fxp.Min {
		if e.SheetSettings.UseHalfStatDefaults {
			level = level.Div(fxp.Two).Trunc() + fxp.Five
		}
		if attrDiff.Difficulty == difficulty.Wildcard {
			points = points.Div(fxp.Three)
		} else if def != nil && def.Points > 0 {
			points += def.Points
		}
		points = points.Trunc()
		switch {
		case points == fxp.One:
			// relativeLevel is preset to this point value
		case points > fxp.One && points < fxp.Four:
			relativeLevel += fxp.One
		case points >= fxp.Four:
			relativeLevel += fxp.One + points.Div(fxp.Four).Trunc()
		case attrDiff.Difficulty != difficulty.Wildcard && def != nil && def.Points < 0:
			relativeLevel = def.AdjLevel - level
		default:
			level = fxp.Min
			relativeLevel = 0
		}
		if level != fxp.Min {
			level += relativeLevel
			if attrDiff.Difficulty != difficulty.Wildcard && def != nil && level < def.AdjLevel {
				level = def.AdjLevel
			}
			if e != nil {
				bonus := e.SkillBonusFor(name, specialization, tags, &tooltip)
				level += bonus
				relativeLevel += bonus
				bonus = e.EncumbranceLevel(true).Penalty().Mul(encumbrancePenaltyMultiplier)
				level += bonus
				if bonus != 0 {
					fmt.Fprintf(&tooltip, i18n.Text("\nEncumbrance [%s]"), bonus.StringWithSign())
				}
			}
		}
	}
	return Level{
		Level:         level,
		RelativeLevel: relativeLevel,
		Tooltip:       tooltip.String(),
	}
}

// CalculateTechniqueLevel returns the calculated level for a technique.
func CalculateTechniqueLevel(e *Entity, name, specialization string, tags []string, def *SkillDefault, diffLevel difficulty.Level, points fxp.Int, requirePoints bool, limitModifier *fxp.Int, excludes map[string]bool) Level {
	var tooltip xio.ByteBuffer
	var relativeLevel fxp.Int
	level := fxp.Min
	if e != nil {
		if def.DefaultType == SkillID {
			if list := e.SkillNamed(def.Name, def.Specialization, requirePoints, excludes); len(list) > 0 {
				sk := list[0]
				var buf strings.Builder
				buf.WriteString(def.Name)
				if def.Specialization != "" {
					buf.WriteString(" (")
					buf.WriteString(def.Specialization)
					buf.WriteByte(')')
				}
				if excludes == nil {
					excludes = make(map[string]bool)
				}
				excludes[buf.String()] = true
				switch sk.Type {
				case SkillID:
					if sk.DefaultedFrom == nil ||
						(sk.DefaultedFrom.Name != name || sk.DefaultedFrom.Specialization != specialization) {
						level = sk.CalculateLevel(excludes).Level
					}
				case TechniqueID:
					if sk.TechniqueDefault != nil &&
						(sk.TechniqueDefault.Name != name || sk.TechniqueDefault.Specialization != specialization) {
						level = sk.CalculateLevel(excludes).Level
					}
				default:
				}
			}
		} else {
			// Take the modifier back out, as we wanted the base, not the final value.
			level = def.SkillLevelFast(e, true, nil, false) - def.Modifier
		}
		if level != fxp.Min {
			baseLevel := level
			level += def.Modifier
			if diffLevel == difficulty.Hard {
				points -= fxp.One
			}
			if points > 0 {
				relativeLevel = points
			}
			if level != fxp.Min {
				relativeLevel += e.SkillBonusFor(name, specialization, tags, &tooltip)
				level += relativeLevel
			}
			if limitModifier != nil {
				if maximum := baseLevel + *limitModifier; level > maximum {
					relativeLevel -= level - maximum
					level = maximum
				}
			}
		}
	}
	return Level{
		Level:         level,
		RelativeLevel: relativeLevel,
		Tooltip:       tooltip.String(),
	}
}

// UpdateLevel updates the level of the skill, returning true if it has changed.
func (s *Skill) UpdateLevel() bool {
	saved := s.LevelData
	s.DefaultedFrom = s.bestDefaultWithPoints(nil)
	s.LevelData = s.CalculateLevel(nil)
	return saved != s.LevelData
}

func (s *Skill) bestDefaultWithPoints(excluded *SkillDefault) *SkillDefault {
	if strings.HasPrefix(s.Type, TechniqueID) {
		return nil
	}
	best := s.bestDefault(excluded)
	if best != nil {
		baseLine := (s.Entity.ResolveAttributeCurrent(s.Difficulty.Attribute) + s.Difficulty.Difficulty.BaseRelativeLevel()).Trunc()
		level := best.Level.Trunc()
		best.AdjLevel = level
		switch {
		case level == baseLine:
			best.Points = fxp.One
		case level == baseLine+fxp.One:
			best.Points = fxp.Two
		case level > baseLine+fxp.One:
			best.Points = fxp.Four.Mul(level - (baseLine + fxp.One))
		default:
			best.Points = -level.Max(0)
		}
	}
	return best
}

func (s *Skill) bestDefault(excluded *SkillDefault) *SkillDefault {
	if s.Entity == nil || len(s.Defaults) == 0 {
		return nil
	}
	excludes := make(map[string]bool)
	excludes[s.String()] = true
	var bestDef *SkillDefault
	best := fxp.Min
	for _, def := range s.resolveToSpecificDefaults() {
		// For skill-based defaults, prune out any that already use a default that we are involved with
		if def.Equivalent(excluded) || s.inDefaultChain(def, make(map[*Skill]bool)) {
			continue
		}
		if level := s.calcSkillDefaultLevel(def, excludes); best < level {
			best = level
			bestDef = def.CloneWithoutLevelOrPoints()
			bestDef.Level = level
		}
	}
	return bestDef
}

func (s *Skill) calcSkillDefaultLevel(def *SkillDefault, excludes map[string]bool) fxp.Int {
	level := def.SkillLevel(s.Entity, true, excludes, strings.HasPrefix(s.Type, SkillID))
	if def.SkillBased() {
		if other := s.Entity.BestSkillNamed(def.Name, def.Specialization, true, excludes); other != nil {
			level -= s.Entity.SkillBonusFor(def.Name, def.Specialization, s.Tags, nil)
		}
	}
	return level
}

func (s *Skill) inDefaultChain(def *SkillDefault, lookedAt map[*Skill]bool) bool {
	if s.Entity == nil || def == nil || !def.SkillBased() {
		return false
	}
	for _, one := range s.Entity.SkillNamed(def.Name, def.Specialization, true, nil) {
		if one == s {
			return true
		}
		if !lookedAt[one] {
			lookedAt[one] = true
			if s.inDefaultChain(one.DefaultedFrom, lookedAt) {
				return true
			}
		}
	}
	return false
}

func (s *Skill) resolveToSpecificDefaults() []*SkillDefault {
	result := make([]*SkillDefault, 0, len(s.Defaults))
	for _, def := range s.Defaults {
		if s.Entity == nil || def == nil || !def.SkillBased() {
			result = append(result, def)
		} else {
			for _, one := range s.Entity.SkillNamed(def.Name, def.Specialization, true,
				map[string]bool{s.String(): true}) {
				local := *def
				local.Specialization = one.Specialization
				result = append(result, &local)
			}
		}
	}
	return result
}

// TechniqueSatisfied returns true if the Technique is satisfied.
func (s *Skill) TechniqueSatisfied(tooltip *xio.ByteBuffer, prefix string) bool {
	if strings.HasPrefix(s.Type, SkillID) || !s.TechniqueDefault.SkillBased() {
		return true
	}
	sk := s.Entity.BestSkillNamed(s.TechniqueDefault.Name, s.TechniqueDefault.Specialization, false, nil)
	satisfied := sk != nil && (strings.HasPrefix(sk.Type, TechniqueID) || sk.Points > 0)
	if !satisfied && tooltip != nil {
		tooltip.WriteString(prefix)
		if sk == nil {
			tooltip.WriteString(i18n.Text("Requires a skill named "))
		} else {
			tooltip.WriteString(i18n.Text("Requires at least 1 point in the skill named "))
		}
		tooltip.WriteString(s.TechniqueDefault.FullName(s.Entity))
	}
	return satisfied
}

// TL implements TechLevelProvider.
func (s *Skill) TL() string {
	if s.TechLevel != nil {
		return *s.TechLevel
	}
	return ""
}

// RequiresTL implements TechLevelProvider.
func (s *Skill) RequiresTL() bool {
	return s.TechLevel != nil
}

// SetTL implements TechLevelProvider.
func (s *Skill) SetTL(tl string) {
	if s.TechLevel != nil {
		*s.TechLevel = tl
	}
}

// Enabled returns true if this node is enabled.
func (s *Skill) Enabled() bool {
	return true
}

// FillWithNameableKeys adds any nameable keys found to the provided map.
func (s *Skill) FillWithNameableKeys(m map[string]string) {
	Extract(s.Name, m)
	Extract(s.LocalNotes, m)
	Extract(s.Specialization, m)
	if s.Prereq != nil {
		s.Prereq.FillWithNameableKeys(m)
	}
	if s.TechniqueDefault != nil {
		s.TechniqueDefault.FillWithNameableKeys(m)
	}
	for _, one := range s.Defaults {
		one.FillWithNameableKeys(m)
	}
	for _, one := range s.Features {
		one.FillWithNameableKeys(m)
	}
	for _, one := range s.Weapons {
		one.FillWithNameableKeys(m)
	}
}

// ApplyNameableKeys replaces any nameable keys found with the corresponding values in the provided map.
func (s *Skill) ApplyNameableKeys(m map[string]string) {
	s.Name = Apply(s.Name, m)
	s.LocalNotes = Apply(s.LocalNotes, m)
	s.Specialization = Apply(s.Specialization, m)
	if s.Prereq != nil {
		s.Prereq.ApplyNameableKeys(m)
	}
	if s.TechniqueDefault != nil {
		s.TechniqueDefault.ApplyNameableKeys(m)
	}
	for _, one := range s.Defaults {
		one.ApplyNameableKeys(m)
	}
	for _, one := range s.Features {
		one.ApplyNameableKeys(m)
	}
	for _, one := range s.Weapons {
		one.ApplyNameableKeys(m)
	}
}

// CanSwapDefaults returns true if this skill's default can be swapped.
func (s *Skill) CanSwapDefaults() bool {
	return s.Type != TechniqueID && !s.Container() && s.AdjustedPoints(nil) > 0
}

// CanSwapDefaultsWith returns true if this skill's default can be swapped with the other skill.
func (s *Skill) CanSwapDefaultsWith(other *Skill) bool {
	return other != nil && s.CanSwapDefaults() && other.HasDefaultTo(s)
}

// BestSwappableSkill returns the best skill to swap with.
func (s *Skill) BestSwappableSkill() *Skill {
	if s.Entity == nil {
		return nil
	}
	var best *Skill
	Traverse(func(other *Skill) bool {
		if s == other.DefaultSkill() && other.CanSwapDefaultsWith(s) {
			if best == nil || best.CalculateLevel(nil).Level < other.CalculateLevel(nil).Level {
				best = other
			}
		}
		return false
	}, true, true, s.Entity.Skills...)
	return best
}

// SwapDefaults causes this skill's default to be swapped.
func (s *Skill) SwapDefaults() {
	def := s.DefaultedFrom
	s.DefaultedFrom = nil
	if baseSkill := s.Entity.BaseSkill(s.bestDefault(nil), true); baseSkill != nil {
		s.DefaultedFrom = s.bestDefaultWithPoints(def)
		baseSkill.UpdateLevel()
		s.UpdateLevel()
	}
}
