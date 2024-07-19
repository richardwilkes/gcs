// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"context"
	"encoding/binary"
	"fmt"
	"hash"
	"io/fs"
	"maps"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/study"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/tid"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/unison/enums/align"
)

var (
	_ Node[*Spell]                    = &Spell{}
	_ TechLevelProvider[*Spell]       = &Spell{}
	_ SkillAdjustmentProvider[*Spell] = &Spell{}
	_ TemplatePickerProvider          = &Spell{}
	_ EditorData[*Spell]              = &SpellEditData{}
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
	SpellLibSrcColumn
)

// Spell holds the data for a spell.
type Spell struct {
	SpellData
	owner             DataOwner
	LevelData         Level
	UnsatisfiedReason string
}

// SpellData holds the Spell data that is written to disk.
type SpellData struct {
	SourcedID
	SpellEditData
	ThirdParty map[string]any `json:"third_party,omitempty"`
	Children   []*Spell       `json:"children,omitempty"` // Only for containers
	parent     *Spell
}

// SpellEditData holds the Spell data that can be edited by the UI detail editor.
type SpellEditData struct {
	Name             string            `json:"name,omitempty"`
	PageRef          string            `json:"reference,omitempty"`
	PageRefHighlight string            `json:"reference_highlight,omitempty"`
	LocalNotes       string            `json:"notes,omitempty"`
	VTTNotes         string            `json:"vtt_notes,omitempty"`
	Tags             []string          `json:"tags,omitempty"`
	Replacements     map[string]string `json:"replacements,omitempty"`
	SpellNonContainerOnlyEditData
	SkillContainerOnlyEditData
}

// SpellNonContainerOnlyEditData holds the Spell data that is only applicable to spells that aren't containers.
type SpellNonContainerOnlyEditData struct {
	TechLevel         *string             `json:"tech_level,omitempty"`
	Difficulty        AttributeDifficulty `json:"difficulty,omitempty"`
	College           CollegeList         `json:"college,omitempty"`
	PowerSource       string              `json:"power_source,omitempty"`
	Class             string              `json:"spell_class,omitempty"`
	Resist            string              `json:"resist,omitempty"`
	CastingCost       string              `json:"casting_cost,omitempty"`
	MaintenanceCost   string              `json:"maintenance_cost,omitempty"`
	CastingTime       string              `json:"casting_time,omitempty"`
	Duration          string              `json:"duration,omitempty"`
	RitualSkillName   string              `json:"base_skill,omitempty"`
	RitualPrereqCount int                 `json:"prereq_count,omitempty"`
	Points            fxp.Int             `json:"points,omitempty"`
	Prereq            *PrereqList         `json:"prereqs,omitempty"`
	Weapons           []*Weapon           `json:"weapons,omitempty"`
	Study             []*Study            `json:"study,omitempty"`
	StudyHoursNeeded  study.Level         `json:"study_hours_needed,omitempty"`
}

type spellListData struct {
	Version int      `json:"version"`
	Rows    []*Spell `json:"rows"`
}

// NewSpellsFromFile loads an Spell list from a file.
func NewSpellsFromFile(fileSystem fs.FS, filePath string) ([]*Spell, error) {
	var data spellListData
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &data); err != nil {
		return nil, errs.NewWithCause(InvalidFileData(), err)
	}
	if err := jio.CheckVersion(data.Version); err != nil {
		return nil, err
	}
	return data.Rows, nil
}

// SaveSpells writes the Spell list to the file as JSON.
func SaveSpells(spells []*Spell, filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, &spellListData{
		Version: jio.CurrentDataVersion,
		Rows:    spells,
	})
}

// NewSpell creates a new Spell.
func NewSpell(owner DataOwner, parent *Spell, container bool) *Spell {
	s := Spell{
		SpellData: SpellData{
			SourcedID: SourcedID{
				TID: tid.MustNewTID(spellKind(container)),
			},
			parent: parent,
		},
		owner: owner,
	}
	if container {
		s.TemplatePicker = &TemplatePicker{}
	} else {
		s.Difficulty.Attribute = AttributeIDFor(EntityFromNode(&s), "iq")
		s.Difficulty.Difficulty = difficulty.Hard
		s.PowerSource = i18n.Text("Arcane")
		s.Class = i18n.Text("Regular")
		s.CastingCost = "1"
		s.CastingTime = "1 sec"
		s.Duration = "Instant"
		s.Points = fxp.One
	}
	s.Name = s.Kind()
	s.UpdateLevel()
	s.SetOpen(container)
	return &s
}

// NewRitualMagicSpell creates a new Ritual Magic Spell.
func NewRitualMagicSpell(owner DataOwner, parent *Spell, _ bool) *Spell {
	s := Spell{
		SpellData: SpellData{
			SourcedID: SourcedID{
				TID: tid.MustNewTID(kinds.RitualMagicSpell),
			},
			parent: parent,
		},
		owner: owner,
	}
	s.Difficulty.Attribute = AttributeIDFor(EntityFromNode(&s), "iq")
	s.Difficulty.Difficulty = difficulty.Hard
	s.PowerSource = i18n.Text("Arcane")
	s.Class = i18n.Text("Regular")
	s.CastingCost = "1"
	s.CastingTime = "1 sec"
	s.Duration = "Instant"
	s.Points = fxp.One
	s.RitualSkillName = "Ritual Magic"
	s.Name = s.Kind()
	s.SetRawPoints(0)
	return &s
}

func spellKind(container bool) byte {
	if container {
		return kinds.SpellContainer
	}
	return kinds.Spell
}

// ID returns the local ID of this data.
func (s *Spell) ID() tid.TID {
	return s.TID
}

// Container returns true if this is a container.
func (s *Spell) Container() bool {
	return tid.IsKind(s.TID, kinds.SpellContainer)
}

// HasChildren returns true if this node has children.
func (s *Spell) HasChildren() bool {
	return s.Container() && len(s.Children) > 0
}

// NodeChildren returns the children of this node, if any.
func (s *Spell) NodeChildren() []*Spell {
	return s.Children
}

// SetChildren sets the children of this node.
func (s *Spell) SetChildren(children []*Spell) {
	s.Children = children
}

// Parent returns the parent.
func (s *Spell) Parent() *Spell {
	return s.parent
}

// SetParent sets the parent.
func (s *Spell) SetParent(parent *Spell) {
	s.parent = parent
}

// IsOpen returns true if this node is currently open.
func (s *Spell) IsOpen() bool {
	return IsNodeOpen(s)
}

// SetOpen sets the current open state for this node.
func (s *Spell) SetOpen(open bool) {
	SetNodeOpen(s, open)
}

// IsRitualMagic returns true if this is a Ritual Magic Spell.
func (s *Spell) IsRitualMagic() bool {
	return tid.IsKind(s.TID, kinds.RitualMagicSpell)
}

// Clone implements Node.
func (s *Spell) Clone(from LibraryFile, owner DataOwner, parent *Spell, preserveID bool) *Spell {
	var other *Spell
	if s.IsRitualMagic() {
		other = NewRitualMagicSpell(owner, parent, false)
	} else {
		other = NewSpell(owner, parent, s.Container())
		other.SetOpen(s.IsOpen())
	}
	other.AdjustSource(from, s.SourcedID, preserveID)
	other.ThirdParty = s.ThirdParty
	other.SpellEditData.CopyFrom(s)
	if s.HasChildren() {
		other.Children = make([]*Spell, 0, len(s.Children))
		for _, child := range s.Children {
			other.Children = append(other.Children, child.Clone(from, owner, other, preserveID))
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
		Type       string   `json:"type"`
		Categories []string `json:"categories"`
		IsOpen     bool     `json:"open"`
	}
	if err := json.Unmarshal(data, &localData); err != nil {
		return err
	}
	setOpen := false
	if !tid.IsValid(localData.TID) {
		// Fixup old data that used UUIDs instead of TIDs
		var kind byte
		if localData.Type == "ritual_magic_spell" {
			kind = kinds.RitualMagicSpell
		} else {
			kind = spellKind(strings.HasSuffix(localData.Type, containerKeyPostfix))
		}
		localData.TID = tid.MustNewTID(kind)
		setOpen = localData.IsOpen
	}
	s.SpellData = localData.SpellData
	s.ClearUnusedFieldsForType()
	s.Tags = convertOldCategoriesToTags(s.Tags, localData.Categories)
	slices.Sort(s.Tags)
	if s.Container() {
		for _, one := range s.Children {
			one.parent = s
		}
	}
	if setOpen {
		SetNodeOpen(s, true)
	}
	return nil
}

// TemplatePickerData returns the TemplatePicker data, if any.
func (s *Spell) TemplatePickerData() *TemplatePicker {
	return s.TemplatePicker
}

// SpellsHeaderData returns the header data information for the given spell column.
func SpellsHeaderData(columnID int) HeaderData {
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
		data.Detail = PageRefTooltip()
	case SpellLevelColumn:
		data.Title = i18n.Text("SL")
		data.Detail = i18n.Text("Skill Level")
	case SpellRelativeLevelColumn:
		data.Title = i18n.Text("RSL")
		data.Detail = i18n.Text("Relative Skill Level")
	case SpellPointsColumn:
		data.Title = i18n.Text("Pts")
		data.Detail = i18n.Text("Points")
	case SpellLibSrcColumn:
		data.Title = HeaderDatabase
		data.TitleIsImageKey = true
		data.Detail = LibSrcTooltip()
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
			data.Primary = s.ResistWithReplacements()
		}
	case SpellClassColumn:
		if !s.Container() {
			data.Type = cell.Text
			data.Primary = s.ClassWithReplacements()
		}
	case SpellCollegeColumn:
		if !s.Container() {
			data.Type = cell.Text
			data.Primary = strings.Join(s.CollegeWithReplacements(), ", ")
		}
	case SpellCastCostColumn:
		if !s.Container() {
			data.Type = cell.Text
			data.Primary = s.CastingCostWithReplacements()
		}
	case SpellMaintainCostColumn:
		if !s.Container() {
			data.Type = cell.Text
			data.Primary = s.MaintenanceCostWithReplacements()
		}
	case SpellCastTimeColumn:
		if !s.Container() {
			data.Type = cell.Text
			data.Primary = s.CastingTimeWithReplacements()
		}
	case SpellDurationColumn:
		if !s.Container() {
			data.Type = cell.Text
			data.Primary = s.DurationWithReplacements()
		}
	case SpellDifficultyColumn:
		if !s.Container() {
			data.Type = cell.Text
			data.Primary = s.Difficulty.Description(EntityFromNode(s))
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
			data.Secondary = s.NameWithReplacements()
		}
	case SpellLevelColumn:
		if !s.Container() {
			data.Type = cell.Text
			level := s.CalculateLevel()
			data.Primary = level.LevelAsString(s.Container())
			if level.Tooltip != "" {
				data.Tooltip = IncludesModifiersFrom() + ":" + level.Tooltip
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
				data.Primary = ResolveAttributeName(EntityFromNode(s), s.Difficulty.Attribute)
				if rsl != 0 {
					data.Primary += rsl.StringWithSign()
				}
			}
			if tooltip := s.CalculateLevel().Tooltip; tooltip != "" {
				data.Tooltip = IncludesModifiersFrom() + ":" + tooltip
			}
		}
	case SpellPointsColumn:
		data.Type = cell.Text
		var tooltip xio.ByteBuffer
		data.Primary = s.AdjustedPoints(&tooltip).String()
		data.Alignment = align.End
		if tooltip.Len() != 0 {
			data.Tooltip = IncludesModifiersFrom() + ":" + tooltip.String()
		}
	case SpellDescriptionForPageColumn:
		s.CellData(SpellDescriptionColumn, data)
		if !s.Container() {
			var buffer strings.Builder
			addPartToBuffer(&buffer, i18n.Text("Resistance"), s.ResistWithReplacements())
			addPartToBuffer(&buffer, i18n.Text("Class"), s.ClassWithReplacements())
			addPartToBuffer(&buffer, i18n.Text("Cast"), s.CastingCostWithReplacements())
			addPartToBuffer(&buffer, i18n.Text("Maintain"), s.MaintenanceCostWithReplacements())
			addPartToBuffer(&buffer, i18n.Text("Time"), s.CastingTimeWithReplacements())
			addPartToBuffer(&buffer, i18n.Text("Duration"), s.DurationWithReplacements())
			addPartToBuffer(&buffer, i18n.Text("College"), strings.Join(s.CollegeWithReplacements(), ", "))
			if buffer.Len() != 0 {
				if data.Secondary == "" {
					data.Secondary = buffer.String()
				} else {
					data.Secondary += "\n" + buffer.String()
				}
			}
		}
	case SpellLibSrcColumn:
		data.Type = cell.Text
		data.Alignment = align.Middle
		if !toolbox.IsNil(s.owner) {
			state, _ := s.owner.SourceMatcher().Match(s)
			data.Primary = state.AltString()
			data.Tooltip = state.String()
			if state != srcstate.Custom {
				data.Tooltip += "\n" + s.Source.String()
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
	case s.IsRitualMagic():
		return rsl.StringWithSign()
	default:
		return ResolveAttributeName(EntityFromNode(s), s.Difficulty.Attribute) + rsl.StringWithSign()
	}
}

// AdjustedRelativeLevel returns the relative skill level.
func (s *Spell) AdjustedRelativeLevel() fxp.Int {
	if s.Container() {
		return fxp.Min
	}
	if EntityFromNode(s) != nil && s.CalculateLevel().Level > 0 {
		return s.LevelData.RelativeLevel
	}
	return fxp.Min
}

// UpdateLevel updates the level of the spell, returning true if it has changed.
func (s *Spell) UpdateLevel() bool {
	saved := s.LevelData
	colleges := s.CollegeWithReplacements()
	if s.IsRitualMagic() {
		s.LevelData = CalculateRitualMagicSpellLevel(EntityFromNode(s), s.NameWithReplacements(),
			s.PowerSourceWithReplacements(), s.RitualSkillNameWithReplacements(), s.RitualPrereqCount, colleges,
			s.Tags, s.Difficulty, s.AdjustedPoints(nil))
	} else {
		s.LevelData = CalculateSpellLevel(EntityFromNode(s), s.NameWithReplacements(), s.PowerSourceWithReplacements(),
			colleges, s.Tags, s.Difficulty, s.AdjustedPoints(nil))
	}
	return saved != s.LevelData
}

// CalculateLevel returns the computed level without updating it.
func (s *Spell) CalculateLevel() Level {
	if s.IsRitualMagic() {
		return CalculateRitualMagicSpellLevel(EntityFromNode(s), s.NameWithReplacements(),
			s.PowerSourceWithReplacements(), s.RitualSkillNameWithReplacements(), s.RitualPrereqCount,
			s.CollegeWithReplacements(), s.Tags, s.Difficulty, s.AdjustedPoints(nil))
	}
	return CalculateSpellLevel(EntityFromNode(s), s.NameWithReplacements(), s.PowerSourceWithReplacements(),
		s.CollegeWithReplacements(), s.Tags, s.Difficulty, s.AdjustedPoints(nil))
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
	skillLevel := CalculateTechniqueLevel(e, nil, name, college, tags, def, difficulty.Difficulty, points, false,
		&limit, nil)
	// CalculateTechniqueLevel() does not add the default skill modifier to the relative level, only to the final level
	skillLevel.RelativeLevel += def.Modifier
	def.Specialization = ""
	def.Modifier -= fxp.Six
	fallback := CalculateTechniqueLevel(e, nil, name, college, tags, def, difficulty.Difficulty, points, false,
		&limit, nil)
	fallback.RelativeLevel += def.Modifier
	if skillLevel.Level >= fallback.Level {
		return skillLevel
	}
	return fallback
}

// RitualMagicSatisfied returns true if the Ritual Magic Spell is satisfied.
func (s *Spell) RitualMagicSatisfied(tooltip *xio.ByteBuffer, prefix string) bool {
	if !s.IsRitualMagic() {
		return true
	}
	colleges := s.CollegeWithReplacements()
	if len(colleges) == 0 {
		if tooltip != nil {
			tooltip.WriteString(prefix)
			tooltip.WriteString(i18n.Text("Must be assigned to a college"))
		}
		return false
	}
	e := EntityFromNode(s)
	for _, college := range colleges {
		if e.BestSkillNamed(s.RitualSkillNameWithReplacements(), college, false, nil) != nil {
			return true
		}
	}
	if e.BestSkillNamed(s.RitualSkillNameWithReplacements(), "", false, nil) != nil {
		return true
	}
	if tooltip != nil {
		tooltip.WriteString(prefix)
		tooltip.WriteString(i18n.Text("Requires a skill named "))
		ritual := s.RitualSkillNameWithReplacements()
		tooltip.WriteString(ritual)
		tooltip.WriteString(" (")
		tooltip.WriteString(colleges[0])
		tooltip.WriteByte(')')
		for _, college := range colleges[1:] {
			tooltip.WriteString(i18n.Text(" or "))
			tooltip.WriteString(ritual)
			tooltip.WriteString(" (")
			tooltip.WriteString(college)
			tooltip.WriteByte(')')
		}
	}
	return false
}

// DataOwner returns the data owner.
func (s *Spell) DataOwner() DataOwner {
	return s.owner
}

// SetDataOwner sets the data owner and configures any sub-components as needed.
func (s *Spell) SetDataOwner(owner DataOwner) {
	s.owner = owner
	if s.Container() {
		for _, child := range s.Children {
			child.SetDataOwner(owner)
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
	e := EntityFromNode(s)
	if s.Container() || !(e != nil && e.SheetSettings.ShowSpellAdj) {
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
		if strings.Contains(strings.ToLower(s.ClassWithReplacements()), "blocking") {
			return ritual
		}
		return ritual + i18n.Text("; Cost: -1")
	default:
		adj := fxp.As[int]((level - fxp.Fifteen).Div(fxp.Five))
		class := strings.ToLower(s.ClassWithReplacements())
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
	prefs := SheetSettingsFor(EntityFromNode(s))
	if optionChecker(prefs.NotesDisplay) {
		AppendStringOntoNewLine(&buffer, strings.TrimSpace(s.Notes()))
		AppendStringOntoNewLine(&buffer, s.Rituals())
		AppendStringOntoNewLine(&buffer, StudyHoursProgressText(ResolveStudyHours(s.Study), s.StudyHoursNeeded, false))
	}
	addTooltipForSkillLevelAdj(optionChecker, prefs, s.LevelData, &buffer)
	return buffer.String()
}

func (s *Spell) String() string {
	var buffer strings.Builder
	buffer.WriteString(s.NameWithReplacements())
	if !s.Container() {
		if s.TechLevel != nil {
			buffer.WriteString("/TL")
			buffer.WriteString(*s.TechLevel)
		}
	}
	return buffer.String()
}

func (s *Spell) resolveLocalNotes() string {
	return EvalEmbeddedRegex.ReplaceAllStringFunc(s.LocalNotesWithReplacements(), EntityFromNode(s).EmbeddedEval)
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
	return AdjustedPointsForNonContainerSpell(EntityFromNode(s), s.Points, s.NameWithReplacements(),
		s.PowerSourceWithReplacements(), s.CollegeWithReplacements(), s.Tags, tooltip)
}

// AdjustedPointsForNonContainerSpell returns the points, adjusted for any bonuses.
func AdjustedPointsForNonContainerSpell(e *Entity, points fxp.Int, name, powerSource string, colleges, tags []string, tooltip *xio.ByteBuffer) fxp.Int {
	if e != nil {
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

// NameableReplacements returns the replacements to be used with Nameables.
func (s *Spell) NameableReplacements() map[string]string {
	return s.Replacements
}

// NameWithReplacements returns the name with any replacements applied.
func (s *Spell) NameWithReplacements() string {
	return ApplyNameables(s.Name, s.Replacements)
}

// LocalNotesWithReplacements returns the local notes with any replacements applied.
func (s *Spell) LocalNotesWithReplacements() string {
	return ApplyNameables(s.LocalNotes, s.Replacements)
}

// PowerSourceWithReplacements returns the power source with any replacements applied.
func (s *Spell) PowerSourceWithReplacements() string {
	return ApplyNameables(s.PowerSource, s.Replacements)
}

// ClassWithReplacements returns the class with any replacements applied.
func (s *Spell) ClassWithReplacements() string {
	return ApplyNameables(s.Class, s.Replacements)
}

// ResistWithReplacements returns the resist with any replacements applied.
func (s *Spell) ResistWithReplacements() string {
	return ApplyNameables(s.Resist, s.Replacements)
}

// CastingCostWithReplacements returns the casting cost with any replacements applied.
func (s *Spell) CastingCostWithReplacements() string {
	return ApplyNameables(s.CastingCost, s.Replacements)
}

// MaintenanceCostWithReplacements returns the maintenance cost with any replacements applied.
func (s *Spell) MaintenanceCostWithReplacements() string {
	return ApplyNameables(s.MaintenanceCost, s.Replacements)
}

// CastingTimeWithReplacements returns the casting time with any replacements applied.
func (s *Spell) CastingTimeWithReplacements() string {
	return ApplyNameables(s.CastingTime, s.Replacements)
}

// DurationWithReplacements returns the duration with any replacements applied.
func (s *Spell) DurationWithReplacements() string {
	return ApplyNameables(s.Duration, s.Replacements)
}

// RitualSkillNameWithReplacements returns the ritual skill name with any replacements applied.
func (s *Spell) RitualSkillNameWithReplacements() string {
	return ApplyNameables(s.RitualSkillName, s.Replacements)
}

// CollegeWithReplacements returns the college(s) with any replacements applied.
func (s *Spell) CollegeWithReplacements() []string {
	return ApplyNameablesToList(s.College, s.Replacements)
}

// FillWithNameableKeys adds any nameable keys found to the provided map.
func (s *Spell) FillWithNameableKeys(m map[string]string) {
	ExtractNameables(s.Name, m)
	ExtractNameables(s.LocalNotes, m)
	ExtractNameables(s.PowerSource, m)
	ExtractNameables(s.Class, m)
	ExtractNameables(s.Resist, m)
	ExtractNameables(s.CastingCost, m)
	ExtractNameables(s.MaintenanceCost, m)
	ExtractNameables(s.CastingTime, m)
	ExtractNameables(s.Duration, m)
	ExtractNameables(s.RitualSkillName, m)
	for _, one := range s.College {
		ExtractNameables(one, m)
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
	needed := make(map[string]string)
	s.FillWithNameableKeys(needed)
	s.Replacements = RetainNeededReplacements(needed, m)
}

// Kind returns the kind of data.
func (s *Spell) Kind() string {
	if s.IsRitualMagic() {
		return i18n.Text("Ritual Magic Spell")
	}
	if s.Container() {
		return i18n.Text("Spell Container")
	}
	return i18n.Text("Spell")
}

// ClearUnusedFieldsForType zeroes out the fields that are not applicable to this type (container vs not-container).
func (s *Spell) ClearUnusedFieldsForType() {
	if s.Container() {
		s.TechLevel = nil
		s.Difficulty = AttributeDifficulty{omit: true}
		s.College = nil
		s.PowerSource = ""
		s.Class = ""
		s.Resist = ""
		s.CastingCost = ""
		s.MaintenanceCost = ""
		s.CastingTime = ""
		s.Duration = ""
		s.RitualSkillName = ""
		s.RitualPrereqCount = 0
		s.Points = 0
		s.Prereq = nil
		s.Weapons = nil
		s.StudyHoursNeeded = study.Standard
		if s.TemplatePicker == nil {
			s.TemplatePicker = &TemplatePicker{}
		}
	} else {
		s.Children = nil
		s.TemplatePicker = nil
		s.Difficulty.omit = false
	}
}

// GetSource returns the source of this data.
func (s *Spell) GetSource() Source {
	return s.Source
}

// ClearSource clears the source of this data.
func (s *Spell) ClearSource() {
	s.Source = Source{}
}

// SyncWithSource synchronizes this data with the source.
func (s *Spell) SyncWithSource() {
	if !toolbox.IsNil(s.owner) {
		if state, data := s.owner.SourceMatcher().Match(s); state == srcstate.Mismatched {
			if other, ok := data.(*Spell); ok {
				s.Name = other.Name
				s.PageRef = other.PageRef
				s.PageRefHighlight = other.PageRefHighlight
				s.LocalNotes = other.LocalNotes
				s.Tags = slices.Clone(other.Tags)
				if s.Container() {
					s.TemplatePicker = other.TemplatePicker.Clone()
				} else {
					s.Difficulty = other.Difficulty
					s.College = txt.CloneStringSlice(s.College)
					s.PowerSource = other.PowerSource
					s.Class = other.Class
					s.Resist = other.Resist
					s.CastingCost = other.CastingCost
					s.MaintenanceCost = other.MaintenanceCost
					s.CastingTime = other.CastingTime
					s.Duration = other.Duration
					s.RitualSkillName = other.RitualSkillName
					s.RitualPrereqCount = other.RitualPrereqCount
					s.Prereq = other.Prereq.CloneResolvingEmpty(false, true)
					s.Weapons = CloneWeapons(other.Weapons, false)
				}
			}
		}
	}
}

// Hash writes this object's contents into the hasher. Note that this only hashes the data that is considered to be
// "source" data, i.e. not expected to be modified by the user after copying from a library.
func (s *Spell) Hash(h hash.Hash) {
	_, _ = h.Write([]byte(s.Name))
	_, _ = h.Write([]byte(s.PageRef))
	_, _ = h.Write([]byte(s.PageRefHighlight))
	_, _ = h.Write([]byte(s.LocalNotes))
	for _, tag := range s.Tags {
		_, _ = h.Write([]byte(tag))
	}
	if s.Container() {
		s.TemplatePicker.Hash(h)
	} else {
		s.Difficulty.Hash(h)
		for _, college := range s.College {
			_, _ = h.Write([]byte(college))
		}
		_, _ = h.Write([]byte(s.PowerSource))
		_, _ = h.Write([]byte(s.Class))
		_, _ = h.Write([]byte(s.Resist))
		_, _ = h.Write([]byte(s.CastingCost))
		_, _ = h.Write([]byte(s.MaintenanceCost))
		_, _ = h.Write([]byte(s.CastingTime))
		_, _ = h.Write([]byte(s.Duration))
		_, _ = h.Write([]byte(s.RitualSkillName))
		_ = binary.Write(h, binary.LittleEndian, int64(s.RitualPrereqCount))
		s.Prereq.Hash(h)
		for _, weapon := range s.Weapons {
			weapon.Hash(h)
		}
	}
}

// CopyFrom implements node.EditorData.
func (s *SpellEditData) CopyFrom(other *Spell) {
	s.copyFrom(&other.SpellEditData, other.Container(), false)
}

// ApplyTo implements node.EditorData.
func (s *SpellEditData) ApplyTo(other *Spell) {
	other.SpellEditData.copyFrom(s, other.Container(), true)
}

func (s *SpellEditData) copyFrom(other *SpellEditData, isContainer, isApply bool) {
	*s = *other
	s.Tags = txt.CloneStringSlice(other.Tags)
	s.Replacements = maps.Clone(other.Replacements)
	if other.TechLevel != nil {
		tl := *other.TechLevel
		s.TechLevel = &tl
	}
	s.College = txt.CloneStringSlice(other.College)
	s.Prereq = s.Prereq.CloneResolvingEmpty(isContainer, isApply)
	s.Weapons = CloneWeapons(other.Weapons, isApply)
	if len(other.Study) != 0 {
		s.Study = make([]*Study, len(other.Study))
		for i := range other.Study {
			s.Study[i] = other.Study[i].Clone()
		}
	}
	s.TemplatePicker = s.TemplatePicker.Clone()
}
