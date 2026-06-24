// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"cmp"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"hash"
	"io/fs"
	"maps"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/study"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison/enums/align"
)

var (
	_ Node[*Skill]                    = &Skill{}
	_ TechLevelProvider[*Skill]       = &Skill{}
	_ SkillAdjustmentProvider[*Skill] = &Skill{}
	_ TemplatePickerProvider          = &Skill{}
	_ EditorData[*Skill]              = &SkillEditData{}
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
	SkillLibSrcColumn
)

// Skill holds the data for a skill.
type Skill struct {
	SkillData
	owner             DataOwner
	LevelData         Level
	UnsatisfiedReason string
}

// SkillData holds the Skill data that is written to disk.
type SkillData struct {
	SourcedID
	SkillEditData
	ThirdParty map[string]any `json:"third_party,omitzero"`
	Children   []*Skill       `json:"children,omitzero"` // Only for containers
	parent     *Skill
}

// SkillEditData holds the Skill data that can be edited by the UI detail editor.
type SkillEditData struct {
	SkillSyncData
	VTTNotes     string            `json:"vtt_notes,omitzero"`
	Replacements map[string]string `json:"replacements,omitzero"`
	SkillNonContainerOnlyEditData
	SkillContainerOnlySyncData
}

// SkillNonContainerOnlyEditData holds the Skill data that is only applicable to skills that aren't containers.
type SkillNonContainerOnlyEditData struct {
	SkillNonContainerOnlySyncData
	TechLevel        *string       `json:"tech_level,omitzero"`
	Points           fxp.Int       `json:"points,omitzero"`
	DefaultedFrom    *SkillDefault `json:"defaulted_from,omitzero"`
	Study            []*Study      `json:"study,omitzero"`
	StudyHoursNeeded study.Level   `json:"study_hours_needed,omitzero"`
}

// SkillSyncData holds the skill sync data that is common to both containers and non-containers.
type SkillSyncData struct {
	Name             string   `json:"name,omitzero"`
	PageRef          string   `json:"reference,omitzero"`
	PageRefHighlight string   `json:"reference_highlight,omitzero"`
	LocalNotes       string   `json:"local_notes,omitzero"`
	Tags             []string `json:"tags,omitzero"`
}

// SkillNonContainerOnlySyncData holds the Skill sync data that is only applicable to skills that aren't containers.
type SkillNonContainerOnlySyncData struct {
	Specialization               string              `json:"specialization,omitzero"`
	OptionalSpecialization       string              `json:"optional_specialization,omitzero"`
	Difficulty                   AttributeDifficulty `json:"difficulty,omitzero"`
	EncumbrancePenaltyMultiplier fxp.Int             `json:"encumbrance_penalty_multiplier,omitzero"`
	Defaults                     []*SkillDefault     `json:"defaults,omitzero"`
	TechniqueDefault             *SkillDefault       `json:"default,omitzero"`
	TechniqueLimitModifier       *fxp.Int            `json:"limit,omitzero"`
	Prereq                       *PrereqList         `json:"prereqs,omitzero"`
	Weapons                      []*Weapon           `json:"weapons,omitzero"`
	Features                     Features            `json:"features,omitzero"`
}

// SkillContainerOnlySyncData holds the skill sync data that is only applicable to skills that are containers.
type SkillContainerOnlySyncData struct {
	TemplatePicker *TemplatePicker `json:"template_picker,omitzero"`
}

type skillListData struct {
	Version int      `json:"version"`
	Rows    []*Skill `json:"rows"`
}

// NewSkillsFromFile loads an Skill list from a file.
func NewSkillsFromFile(fileSystem fs.FS, filePath string) ([]*Skill, error) {
	var data skillListData
	if err := jio.Load(fileSystem, filePath, &data); err != nil {
		return nil, errs.NewWithCause(InvalidFileData(), err)
	}
	if err := jio.CheckVersion(data.Version); err != nil {
		return nil, err
	}
	Traverse(func(skill *Skill) bool {
		// Fix up some bad data in standalone skill lists where Hard techniques incorrectly had 1 point assigned to them
		// instead of 2.
		if skill.IsTechnique() &&
			skill.Difficulty.Difficulty == difficulty.Hard &&
			skill.Points == fxp.One {
			skill.Points = fxp.Two
		}

		skill.SetDataOwner(nil)
		return false
	}, false, true, data.Rows...)
	return data.Rows, nil
}

// SaveSkills writes the Skill list to the file as JSON.
func SaveSkills(skills []*Skill, filePath string) error {
	return jio.SaveToFile(filePath, &skillListData{
		Version: jio.CurrentDataVersion,
		Rows:    skills,
	})
}

// NewSkill creates a new Skill.
func NewSkill(owner DataOwner, parent *Skill, container bool) *Skill {
	var s Skill
	s.TID = tid.MustNewTID(skillKind(container))
	s.parent = parent
	s.owner = owner
	if container {
		s.TemplatePicker = &TemplatePicker{}
	} else {
		s.Difficulty.Attribute = AttributeIDFor(EntityFromNode(&s), DexterityID)
		s.Difficulty.Difficulty = difficulty.Average
		s.Points = fxp.One
	}
	s.Name = s.Kind()
	s.SetOpen(container)
	return &s
}

func skillKind(container bool) byte {
	if container {
		return kinds.SkillContainer
	}
	return kinds.Skill
}

// NewTechnique creates a new technique (i.e. a specialized use of a Skill). All parameters may be nil or empty.
func NewTechnique(owner DataOwner, parent *Skill, skillName string) *Skill {
	var s Skill
	s.TID = tid.MustNewTID(kinds.Technique)
	s.parent = parent
	s.owner = owner
	s.Difficulty.Difficulty = difficulty.Average
	s.Points = fxp.One
	if skillName == "" {
		skillName = i18n.Text("Skill")
	}
	s.TechniqueDefault = &SkillDefault{
		DefaultType: SkillID,
		Name: criteria.Text{
			TextData: criteria.TextData{
				Qualifier: skillName,
				Compare:   criteria.IsText,
			},
		},
	}
	s.Name = s.Kind()
	return &s
}

// ID returns the local ID of this data.
func (s *Skill) ID() tid.TID {
	return s.TID
}

// Container returns true if this is a container.
func (s *Skill) Container() bool {
	return tid.IsKind(s.TID, kinds.SkillContainer)
}

// HasChildren returns true if this node has children.
func (s *Skill) HasChildren() bool {
	return s.Container() && len(s.Children) > 0
}

// NodeChildren returns the children of this node, if any.
func (s *Skill) NodeChildren() []*Skill {
	return s.Children
}

// SetChildren sets the children of this node.
func (s *Skill) SetChildren(children []*Skill) {
	s.Children = children
}

// Parent returns the parent.
func (s *Skill) Parent() *Skill {
	return s.parent
}

// SetParent sets the parent.
func (s *Skill) SetParent(parent *Skill) {
	s.parent = parent
}

// IsOpen returns true if this node is currently open.
func (s *Skill) IsOpen() bool {
	return IsNodeOpen(s)
}

// SetOpen sets the current open state for this node.
func (s *Skill) SetOpen(open bool) {
	SetNodeOpen(s, open)
}

// IsTechnique returns true if this is a technique.
func (s *Skill) IsTechnique() bool {
	return tid.IsKind(s.TID, kinds.Technique)
}

// Clone implements Node.
func (s *Skill) Clone(from LibraryFile, owner DataOwner, parent *Skill, preserveID bool) *Skill {
	var other *Skill
	if s.IsTechnique() {
		other = NewTechnique(owner, parent, s.TechniqueDefault.Name.Qualifier)
	} else {
		other = NewSkill(owner, parent, s.Container())
		other.SetOpen(s.IsOpen())
	}
	other.AdjustSource(from, s.SourcedID, preserveID)
	other.ThirdParty = s.ThirdParty
	other.CopyFrom(s)
	PropagateNodeNoteClosedState(s, other)
	if s.HasChildren() {
		other.Children = make([]*Skill, 0, len(s.Children))
		for _, child := range s.Children {
			other.Children = append(other.Children, child.Clone(from, owner, other, preserveID))
		}
	}
	return other
}

// MarshalJSONTo implements json.MarshalerTo.
func (s *Skill) MarshalJSONTo(enc *jsontext.Encoder) error {
	s.ClearUnusedFieldsForType()
	type calcNoLevel struct {
		ResolvedNotes     string `json:"resolved_notes,omitzero"`
		UnsatisfiedReason string `json:"unsatisfied_reason,omitzero"`
	}
	cnl := calcNoLevel{UnsatisfiedReason: s.UnsatisfiedReason}
	notes := s.ResolveLocalNotes()
	if notes != s.LocalNotes {
		cnl.ResolvedNotes = notes
	}
	if s.Container() || s.LevelData.Level <= 0 {
		value := &struct {
			SkillData
			Calc *calcNoLevel `json:"calc,omitzero"`
		}{
			SkillData: s.SkillData,
		}
		if cnl != (calcNoLevel{}) {
			value.Calc = &cnl
		}
		return json.MarshalEncode(enc, value)
	}
	type calc struct {
		Level              fxp.Int `json:"level"`
		RelativeSkillLevel string  `json:"rsl"`
		calcNoLevel
	}
	return json.MarshalEncode(enc, &struct {
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

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (s *Skill) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var localData struct {
		SkillData
		// Old data fields
		Type       string   `json:"type"`
		ExprNotes  string   `json:"notes"`
		Categories []string `json:"categories"`
		IsOpen     bool     `json:"open"`
	}
	if err := json.UnmarshalDecode(dec, &localData); err != nil {
		return err
	}
	setOpen := false
	if !tid.IsValid(localData.TID) {
		// Fixup old data that used UUIDs instead of TIDs
		var kind byte
		if localData.Type == "technique" {
			kind = kinds.Technique
		} else {
			kind = skillKind(strings.HasSuffix(localData.Type, containerKeyPostfix))
		}
		localData.TID = tid.MustNewTID(kind)
		setOpen = localData.IsOpen
	}
	s.SkillData = localData.SkillData
	if s.TechniqueDefault != nil {
		s.TechniqueDefault.Name.Compare = criteria.IsText
	}
	if s.LocalNotes == "" && localData.ExprNotes != "" {
		s.LocalNotes = EmbeddedExprToScript(localData.ExprNotes)
	}
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
		data.Detail = PageRefTooltip()
	case SkillLevelColumn:
		data.Title = i18n.Text("SL")
		data.Detail = i18n.Text("Skill Level")
	case SkillRelativeLevelColumn:
		data.Title = i18n.Text("RSL")
		data.Detail = i18n.Text("Relative Skill Level")
	case SkillPointsColumn:
		data.Title = i18n.Text("Pts")
		data.Detail = i18n.Text("Points")
	case SkillLibSrcColumn:
		data.Title = HeaderDatabase
		data.TitleIsImageKey = true
		data.Detail = LibSrcTooltip()
	}
	return data
}

// CellData returns the cell data information for the given column.
func (s *Skill) CellData(columnID int, data *CellData) {
	data.Self = s
	switch columnID {
	case SkillDescriptionColumn:
		data.Type = cell.Text
		data.Primary = s.String()
		data.Secondary = s.SecondaryText(func(option display.Option) bool { return option.Inline() })
		data.UnsatisfiedReason = s.UnsatisfiedReason
		data.Tooltip = s.SecondaryText(func(option display.Option) bool { return option.Tooltip() })
		data.TemplateInfo = s.TemplatePicker.String()
	case SkillDifficultyColumn:
		if !s.Container() {
			data.Type = cell.Text
			diff := s.AdjustedDifficulty()
			data.Primary = diff.Description(EntityFromNode(s))
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
			data.Secondary = s.NameWithReplacements()
		}
	case SkillLevelColumn:
		if !s.Container() {
			data.Type = cell.Text
			level := s.CalculateLevel(nil)
			data.Primary = level.LevelAsString(s.Container())
			if level.Tooltip != "" {
				data.Tooltip = IncludesModifiersFrom() + ":" + level.Tooltip
			}
			data.Alignment = align.End
		}
	case SkillRelativeLevelColumn:
		if !s.Container() {
			data.Type = cell.Text
			data.Primary = FormatRelativeSkill(EntityFromNode(s), s.IsTechnique(), s.AdjustedDifficulty(),
				s.AdjustedRelativeLevel())
			if tooltip := s.CalculateLevel(nil).Tooltip; tooltip != "" {
				data.Tooltip = IncludesModifiersFrom() + ":" + tooltip
			}
		}
	case SkillPointsColumn:
		data.Type = cell.Text
		var tooltip xbytes.InsertBuffer
		data.Primary = s.AdjustedPoints(&tooltip).String()
		data.Alignment = align.End
		if tooltip.Len() != 0 {
			data.Tooltip = IncludesModifiersFrom() + ":" + tooltip.String()
		}
	case SkillLibSrcColumn:
		data.Type = cell.Text
		data.Alignment = align.Middle
		if !xreflect.IsNil(s.owner) {
			state, _ := s.owner.SourceMatcher().Match(s)
			data.Primary = state.AltString()
			data.Tooltip = state.String()
			if state != srcstate.Custom {
				data.Tooltip += "\n" + s.Source.String()
			}
		}
	}
}

// FormatRelativeSkill formats the relative skill for display.
func FormatRelativeSkill(e *Entity, numOnly bool, diff AttributeDifficulty, rsl fxp.Int) string {
	switch {
	case rsl == fxp.Min:
		return "-"
	case numOnly:
		return rsl.Floor().StringWithSign()
	default:
		s := ResolveAttributeName(e, diff.Attribute)
		rsl = rsl.Floor()
		if rsl != 0 {
			s += rsl.StringWithSign()
		}
		return s
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

// DataOwner returns the data owner.
func (s *Skill) DataOwner() DataOwner {
	return s.owner
}

// SetDataOwner sets the data owner and configures any sub-components as needed.
func (s *Skill) SetDataOwner(owner DataOwner) {
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

// BaseSkill returns the best skill for the given default, or nil.
func (s *Skill) BaseSkill(e *Entity, def *SkillDefault, requirePoints bool) *Skill {
	if e == nil || def == nil || !def.SkillBased() {
		return nil
	}
	return e.BestSkillMatching(def.Name, def.Specialization, s.Replacements, requirePoints, nil)
}

// DefaultSkill returns the skill currently defaulted to, or nil.
func (s *Skill) DefaultSkill() *Skill {
	e := EntityFromNode(s)
	if e == nil {
		return nil
	}
	if s.IsTechnique() {
		return s.BaseSkill(e, s.TechniqueDefault, true)
	}
	return s.BaseSkill(e, s.DefaultedFrom, true)
}

// HasDefaultTo returns true if the set of possible defaults includes the other skill.
func (s *Skill) HasDefaultTo(other *Skill) bool {
	for _, def := range s.resolveToSpecificDefaults() {
		if def.SkillBased() && def.NameWithReplacements(s.Replacements) == other.NameWithReplacements() &&
			skillSpecializationMatches(def.Specialization, s.Replacements,
				other.SpecializationWithReplacements(), other.OptionalSpecializationWithReplacements()) {
			return true
		}
	}
	return false
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

// SecondaryText returns the less important information that should be displayed with the description.
func (s *Skill) SecondaryText(optionChecker func(display.Option) bool) string {
	var buffer strings.Builder
	prefs := SheetSettingsFor(EntityFromNode(s))
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
	addTooltipForSkillLevelAdj(optionChecker, prefs, s.LevelData, &buffer)
	return buffer.String()
}

func addTooltipForSkillLevelAdj(optionChecker func(display.Option) bool, prefs *SheetSettings, level Level, to LineBuilder) {
	if optionChecker(prefs.SkillLevelAdjDisplay) {
		if level.Tooltip != "" && level.Tooltip != NoAdditionalModifiers() {
			levelTooltip := level.Tooltip
			msg := IncludesModifiersFrom()
			if !strings.HasPrefix(levelTooltip, msg) {
				levelTooltip = msg + ":" + levelTooltip
			}
			if optionChecker(display.Inline) {
				levelTooltip = strings.ReplaceAll(strings.ReplaceAll(levelTooltip, ":\n", ": "), "\n", ", ")
			}
			AppendStringOntoNewLine(to, levelTooltip)
		}
	}
}

func (s *Skill) String() string {
	var buffer strings.Builder
	buffer.WriteString(s.NameWithReplacements())
	if !s.Container() {
		if s.TechLevel != nil {
			buffer.WriteString("/TL")
			buffer.WriteString(*s.TechLevel)
		}
		if s.Specialization != "" || s.OptionalSpecialization != "" {
			buffer.WriteString(" (")
			if s.Specialization != "" {
				buffer.WriteString(s.SpecializationWithReplacements())

				if s.OptionalSpecialization != "" {
					buffer.WriteString(", ")
				}
			}
			if s.OptionalSpecialization != "" {
				buffer.WriteString(s.OptionalSpecializationWithReplacements())
			}
			buffer.WriteByte(')')
		}
	}
	return buffer.String()
}

// IsLeveled implements LeveledOwner
func (s *Skill) IsLeveled() bool {
	return !s.Container()
}

// CurrentLevel implements LeveledOwner
func (s *Skill) CurrentLevel() fxp.Int {
	return s.LevelData.Level
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
	case s.IsTechnique():
		return rsl.StringWithSign()
	default:
		return ResolveAttributeName(EntityFromNode(s), s.AdjustedDifficulty().Attribute) + rsl.StringWithSign()
	}
}

// AdjustedRelativeLevel returns the relative skill level.
func (s *Skill) AdjustedRelativeLevel() fxp.Int {
	if s.Container() {
		return fxp.Min
	}
	if EntityFromNode(s) != nil && s.LevelData.Level > 0 {
		if s.IsTechnique() {
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
func (s *Skill) AdjustedPoints(tooltip *xbytes.InsertBuffer) fxp.Int {
	if s.Container() {
		var total fxp.Int
		for _, one := range s.Children {
			total += one.AdjustedPoints(tooltip)
		}
		return total
	}
	return AdjustedPointsForNonContainerSkillOrTechnique(EntityFromNode(s), s.Points, s.NameWithReplacements(),
		s.SpecializationWithReplacements(), s.OptionalSpecializationWithReplacements(), s.Tags, tooltip)
}

// AdjustedDifficulty returns the difficulty, adjusted in case of an optional specialization.
func (s *Skill) AdjustedDifficulty() AttributeDifficulty {
	diff := s.Difficulty

	if s.OptionalSpecializationWithReplacements() != "" {
		if diff.Difficulty > difficulty.Easy && diff.Difficulty != difficulty.Wildcard {
			diff.Difficulty--
		}
	}

	return diff
}

// AdjustedPointsForNonContainerSkillOrTechnique returns the points, adjusted for any bonuses.
func AdjustedPointsForNonContainerSkillOrTechnique(e *Entity, points fxp.Int, name, specialization, optionalSpecialization string, tags []string, tooltip *xbytes.InsertBuffer) fxp.Int {
	if e != nil {
		points += e.SkillPointBonusFor(name, specialization, optionalSpecialization, tags, tooltip)
		points = points.Max(0)
	}
	return points
}

// IncrementSkillLevel adds enough points to increment the skill level to the next level.
func (s *Skill) IncrementSkillLevel() {
	if !s.Container() {
		basePoints := s.Points.Floor() + fxp.One
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
		basePoints := s.Points.Floor()
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
	if s.IsTechnique() {
		return CalculateTechniqueLevel(EntityFromNode(s), s.Replacements, s.NameWithReplacements(),
			s.SpecializationWithReplacements(), s.Tags, s.TechniqueDefault, s.AdjustedDifficulty().Difficulty, points, true,
			s.TechniqueLimitModifier, excludes)
	}
	return CalculateSkillLevel(EntityFromNode(s), s.NameWithReplacements(), s.SpecializationWithReplacements(), s.OptionalSpecializationWithReplacements(), s.Tags,
		s.DefaultedFrom, s.AdjustedDifficulty(), points, s.EncumbrancePenaltyMultiplier)
}

// CalculateSkillLevel returns the calculated level for a skill.
func CalculateSkillLevel(e *Entity, name, specialization, optionalSpecialization string, tags []string, def *SkillDefault, attrDiff AttributeDifficulty, points, encumbrancePenaltyMultiplier fxp.Int) Level {
	var tooltip xbytes.InsertBuffer
	relativeLevel := attrDiff.Difficulty.BaseRelativeLevel()
	level := e.ResolveAttributeCurrent(attrDiff.Attribute)
	if level != fxp.Min {
		if e.SheetSettings.UseHalfStatDefaults {
			level = level.Div(fxp.Two).Floor() + fxp.Five
		}
		if attrDiff.Difficulty == difficulty.Wildcard {
			points = points.Div(fxp.Three)
		} else if def != nil && def.Points > 0 {
			points += def.Points
		}
		points = points.Floor()
		switch {
		case points == fxp.One:
			// relativeLevel is preset to this point value
		case points > fxp.One && points < fxp.Four:
			relativeLevel += fxp.One
		case points >= fxp.Four:
			relativeLevel += fxp.One + points.Div(fxp.Four).Floor()
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
				bonus := e.SkillBonusFor(name, specialization, optionalSpecialization, tags, &tooltip)
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
func CalculateTechniqueLevel(e *Entity, replacements map[string]string, name, specialization string, tags []string, def *SkillDefault, diffLevel difficulty.Level, points fxp.Int, requirePoints bool, limitModifier *fxp.Int, excludes map[string]bool) Level {
	var tooltip xbytes.InsertBuffer
	var relativeLevel fxp.Int
	level := fxp.Min
	if e != nil && def != nil {
		if def.DefaultType == SkillID {
			defName := def.NameWithReplacements(replacements)
			defSpec := def.SpecializationWithReplacements(replacements)
			if list := e.SkillNamed(defName, defSpec, requirePoints, excludes); len(list) > 0 {
				var buf strings.Builder
				buf.WriteString(defName)
				if defSpec != "" {
					buf.WriteString(" (")
					buf.WriteString(defSpec)
					buf.WriteByte(')')
				}
				if excludes == nil {
					excludes = make(map[string]bool)
				}
				excludes[buf.String()] = true
				// Use the highest-level matching skill that isn't (circularly) defaulting back to this technique,
				// rather than whichever happens to be encountered first.
				for _, sk := range list {
					var usable bool
					if sk.IsTechnique() {
						usable = sk.TechniqueDefault != nil &&
							(sk.TechniqueDefault.NameWithReplacements(replacements) != name ||
								sk.TechniqueDefault.SpecializationWithReplacements(replacements) != specialization)
					} else {
						usable = sk.DefaultedFrom == nil ||
							sk.DefaultedFrom.NameWithReplacements(replacements) != name ||
							sk.DefaultedFrom.SpecializationWithReplacements(replacements) != specialization
					}
					if usable {
						if candidate := sk.CalculateLevel(excludes).Level; level < candidate {
							level = candidate
						}
					}
				}
			}
		} else {
			// Take the modifier back out, as we wanted the base, not the final value.
			level = def.SkillLevelFast(e, replacements, true, nil, false) - def.Modifier
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
				relativeLevel += e.SkillBonusFor(name, specialization, "", tags, &tooltip)
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
	if s.IsTechnique() {
		return nil
	}
	best := s.bestDefault(excluded)
	if best != nil {
		baseLine := (EntityFromNode(s).ResolveAttributeCurrent(s.AdjustedDifficulty().Attribute) +
			s.AdjustedDifficulty().Difficulty.BaseRelativeLevel()).Floor()
		level := best.Level.Floor()
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
	if EntityFromNode(s) == nil {
		return nil
	}
	excludes := make(map[string]bool)
	excludes[s.String()] = true
	// A skill with points that already records which default it uses should keep using that same default as long as it
	// remains resolvable, rather than silently jumping to whichever default currently yields the highest level. The
	// user changes it explicitly via "Swap Defaults". This only applies to skills that declare their own defaults; a
	// skill whose defaults are purely synthetic (e.g. the automatic -2 default to a sibling's optional specialization)
	// always tracks the dynamically-best default. Skills with no points also show the dynamically-best default.
	keepCurrent := excluded == nil && s.DefaultedFrom != nil && s.AdjustedPoints(nil) > 0 && len(s.Defaults) > 0
	var bestDef, preferredDef *SkillDefault
	best := fxp.Min
	for _, def := range s.resolveToSpecificDefaults() {
		// For skill-based defaults, prune out any that already use a default that we are involved with
		if def.Equivalent(s.Replacements, excluded) || s.inDefaultChain(def, make(map[*Skill]bool)) {
			continue
		}
		level := s.calcSkillDefaultLevel(def, excludes)
		if level == fxp.Min {
			continue
		}
		if best < level {
			best = level
			bestDef = def.CloneWithoutLevelOrPoints()
			bestDef.Level = level
		}
		if keepCurrent && def.Equivalent(s.Replacements, s.DefaultedFrom) {
			preferredDef = def.CloneWithoutLevelOrPoints()
			preferredDef.Level = level
		}
	}
	if preferredDef != nil {
		return preferredDef
	}
	return bestDef
}

// resolvableDefaults returns the distinct, currently-resolvable defaults this skill could use, ordered from the highest
// resulting level to the lowest. Candidates that currently default back to this skill are still included: the user may
// choose one, and SwapToNextDefault breaks the resulting cycle so the chosen skill re-resolves to its own best default.
func (s *Skill) resolvableDefaults() []*SkillDefault {
	if EntityFromNode(s) == nil {
		return nil
	}
	excludes := map[string]bool{s.String(): true}
	var result []*SkillDefault
	for _, def := range s.resolveToSpecificDefaults() {
		level := s.calcSkillDefaultLevel(def, excludes)
		if level == fxp.Min {
			continue
		}
		dup := false
		for _, existing := range result {
			if existing.Equivalent(s.Replacements, def) {
				dup = true
				break
			}
		}
		if dup {
			continue
		}
		clone := def.CloneWithoutLevelOrPoints()
		clone.Level = level
		result = append(result, clone)
	}
	slices.SortStableFunc(result, func(a, b *SkillDefault) int {
		return cmp.Compare(b.Level, a.Level)
	})
	return result
}

// AlternateDefaultsAvailable returns true if this skill has more than one default it could use, allowing the user to
// choose among them via "Swap Defaults". Only skills that declare their own defaults are swappable; a skill whose
// defaults are purely synthetic (the automatic optional-specialization defaults) tracks the best default automatically.
func (s *Skill) AlternateDefaultsAvailable() bool {
	return s.CanSwapDefaults() && len(s.Defaults) > 0 && len(s.resolvableDefaults()) > 1
}

// SwapToNextDefault changes the default this skill uses to the next available one (wrapping around), preserving the
// choice until the user changes it again.
func (s *Skill) SwapToNextDefault() {
	candidates := s.resolvableDefaults()
	if len(candidates) < 2 {
		return
	}
	idx := -1
	if s.DefaultedFrom != nil {
		for i, def := range candidates {
			if def.Equivalent(s.Replacements, s.DefaultedFrom) {
				idx = i
				break
			}
		}
	}
	s.DefaultedFrom = candidates[(idx+1)%len(candidates)]
	// If the newly chosen skill currently defaults back to this one (directly or transitively), clear its default so
	// this explicit choice wins; the recalculation below then resolves it to its own best remaining default. This keeps
	// two skills from ending up defaulting to each other.
	if s.inDefaultChain(s.DefaultedFrom, make(map[*Skill]bool)) {
		if target := s.DefaultSkill(); target != nil && target != s {
			target.DefaultedFrom = nil
		}
	}
	if e := EntityFromNode(s); e != nil {
		e.Recalculate()
	}
}

func (s *Skill) calcSkillDefaultLevel(def *SkillDefault, excludes map[string]bool) fxp.Int {
	e := EntityFromNode(s)
	level := def.SkillLevel(e, s.Replacements, true, excludes, !s.IsTechnique())
	if level == fxp.Min {
		return level
	}
	if def.SkillBased() {
		if other := e.BestSkillMatching(def.Name, def.Specialization, s.Replacements, true, excludes); other != nil {
			level -= e.SkillBonusFor(other.NameWithReplacements(), other.SpecializationWithReplacements(),
				other.OptionalSpecializationWithReplacements(), s.Tags, nil)
		}
	}
	return level
}

func (s *Skill) inDefaultChain(def *SkillDefault, lookedAt map[*Skill]bool) bool {
	e := EntityFromNode(s)
	if e == nil || def == nil || !def.SkillBased() {
		return false
	}
	for _, one := range e.SkillMatching(def.Name, def.Specialization, s.Replacements, true, nil) {
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
	e := EntityFromNode(s)

	result := make([]*SkillDefault, 0, len(s.Defaults))
	for _, def := range s.Defaults {
		if e == nil || def == nil || !def.SkillBased() {
			result = append(result, def)
		} else {
			for _, one := range e.SkillMatching(def.Name, def.Specialization, s.Replacements, true,
				map[string]bool{s.String(): true}) {
				local := *def
				local.Name.Compare = criteria.IsText
				local.Name.Qualifier = one.NameWithReplacements()
				local.Specialization.Compare = criteria.IsText
				if spec := one.SpecializationWithReplacements(); spec != "" {
					local.Specialization.Qualifier = spec
				} else {
					local.Specialization.Qualifier = one.OptionalSpecializationWithReplacements()
				}
				if one.OptionalSpecializationWithReplacements() != "" &&
					def.Specialization.Matches(s.Replacements, one.SpecializationWithReplacements()) {
					local.Modifier -= fxp.Two
				}
				result = append(result, &local)
			}
		}
	}

	if e != nil && s.OptionalSpecializationWithReplacements() == "" {
		excludes := map[string]bool{s.String(): true}
		for _, one := range e.SkillNamed(s.NameWithReplacements(), s.SpecializationWithReplacements(), true, excludes) {
			optSpec := one.OptionalSpecializationWithReplacements()
			if optSpec == "" {
				continue
			}
			// The optional-specialty default only relates skills that share the same required specialization (differing
			// only in the optional one). SkillNamed treats an empty specialization as a wildcard, so without this check
			// a fully-unspecialized skill would pick up defaults to siblings that have a different required
			// specialization.
			if one.SpecializationWithReplacements() != s.SpecializationWithReplacements() {
				continue
			}
			def := &SkillDefault{DefaultType: SkillID, Modifier: -fxp.Two}
			def.Name.Compare = criteria.IsText
			def.Name.Qualifier = one.NameWithReplacements()
			def.Specialization.Compare = criteria.IsText
			if spec := one.SpecializationWithReplacements(); spec != "" {
				def.Specialization.Qualifier = spec
			} else {
				def.Specialization.Qualifier = optSpec
			}
			result = append(result, def)
		}
	}

	return result
}

// TechniqueSatisfied returns true if the Technique is satisfied.
func (s *Skill) TechniqueSatisfied(tooltip *xbytes.InsertBuffer, prefix string) bool {
	if !s.IsTechnique() || !s.TechniqueDefault.SkillBased() {
		return true
	}
	e := EntityFromNode(s)
	sk := e.BestSkillNamed(s.TechniqueDefault.NameWithReplacements(s.Replacements),
		s.TechniqueDefault.SpecializationWithReplacements(s.Replacements), false, nil)
	satisfied := sk != nil && (sk.IsTechnique() || sk.Points > 0)
	if !satisfied && tooltip != nil {
		tooltip.WriteString(prefix)
		if sk == nil {
			tooltip.WriteString(i18n.Text("Requires a skill named "))
		} else {
			tooltip.WriteString(i18n.Text("Requires at least 1 point in the skill named "))
		}
		tooltip.WriteString(s.TechniqueDefault.FullName(e, s.Replacements))
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

// NameWithReplacements returns the name with any replacements applied.
func (s *Skill) NameWithReplacements() string {
	return nameable.Apply(s.Name, s.Replacements)
}

// SpecializationWithReplacements returns the specialization with any replacements applied.
func (s *Skill) SpecializationWithReplacements() string {
	return nameable.Apply(s.Specialization, s.Replacements)
}

// OptionalSpecializationWithReplacements returns the optional specialization with any replacements applied.
func (s *Skill) OptionalSpecializationWithReplacements() string {
	return nameable.Apply(s.OptionalSpecialization, s.Replacements)
}

// LocalNotesWithReplacements returns the local notes with any replacements applied.
func (s *Skill) LocalNotesWithReplacements() string {
	return nameable.Apply(s.LocalNotes, s.Replacements)
}

// Notes implements WeaponOwner.
func (s *Skill) Notes() string {
	return s.ResolveLocalNotes()
}

// ModifierNotes returns the notes due to modifiers.
func (s *Skill) ModifierNotes() string {
	if s.IsTechnique() {
		return i18n.Text("Default: ") + s.TechniqueDefault.FullName(EntityFromNode(s), s.Replacements) +
			s.TechniqueDefault.ModifierAsString()
	}
	if s.AdjustedDifficulty().Difficulty != difficulty.Wildcard {
		defSkill := s.DefaultSkill()
		if defSkill != nil && s.DefaultedFrom != nil {
			return i18n.Text("Default: ") + defSkill.String() + s.DefaultedFrom.ModifierAsString()
		}
	}
	return ""
}

// ResolveLocalNotes resolves the local notes, running any embedded scripts to get the final result.
func (s *Skill) ResolveLocalNotes() string {
	return ResolveText(EntityFromNode(s), deferredNewScriptSkill(s), s.LocalNotesWithReplacements())
}

// NameableReplacements returns the replacements to be used with Nameables.
func (s *Skill) NameableReplacements() map[string]string {
	if s == nil {
		return nil
	}
	return s.Replacements
}

// FillWithNameableKeys adds any nameable keys found to the provided map.
func (s *Skill) FillWithNameableKeys(m, existing map[string]string) {
	if existing == nil {
		existing = s.Replacements
	}
	nameable.Extract(s.Name, m, existing)
	nameable.Extract(s.LocalNotes, m, existing)
	nameable.Extract(s.Specialization, m, existing)
	if s.Prereq != nil {
		s.Prereq.FillWithNameableKeys(m, existing)
	}
	if s.TechniqueDefault != nil {
		s.TechniqueDefault.FillWithNameableKeys(m, existing)
	}
	for _, one := range s.Defaults {
		one.FillWithNameableKeys(m, existing)
	}
	for _, one := range s.Features {
		one.FillWithNameableKeys(m, existing)
	}
	for _, one := range s.Weapons {
		one.FillWithNameableKeys(m, existing)
	}
}

// ApplyNameableKeys replaces any nameable keys found with the corresponding values in the provided map.
func (s *Skill) ApplyNameableKeys(m map[string]string) {
	needed := make(map[string]string)
	s.FillWithNameableKeys(needed, nil)
	s.Replacements = nameable.Reduce(needed, m)
}

// CanSwapDefaults returns true if this skill's default can be swapped.
func (s *Skill) CanSwapDefaults() bool {
	return !s.IsTechnique() && !s.Container() && s.AdjustedPoints(nil) > 0
}

// CanSwapDefaultsWith returns true if this skill's default can be swapped with the other skill.
func (s *Skill) CanSwapDefaultsWith(other *Skill) bool {
	return other != nil && s.CanSwapDefaults() && other.HasDefaultTo(s)
}

// BestSwappableSkill returns the best skill to swap with.
func (s *Skill) BestSwappableSkill() *Skill {
	e := EntityFromNode(s)
	if e == nil {
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
	}, true, true, e.Skills...)
	return best
}

// SwapDefaults causes this skill's default to be swapped.
func (s *Skill) SwapDefaults() {
	def := s.DefaultedFrom
	s.DefaultedFrom = nil
	if e := EntityFromNode(s); e != nil {
		if baseSkill := s.BaseSkill(e, s.bestDefault(nil), true); baseSkill != nil {
			s.DefaultedFrom = s.bestDefaultWithPoints(def)
			baseSkill.UpdateLevel()
			s.UpdateLevel()
		}
	}
}

// Kind returns the kind of data.
func (s *Skill) Kind() string {
	if s.IsTechnique() {
		return i18n.Text("Technique")
	}
	if s.Container() {
		return i18n.Text("Skill Container")
	}
	return i18n.Text("Skill")
}

// ClearUnusedFieldsForType zeroes out the fields that are not applicable to this type (container vs not-container).
func (s *Skill) ClearUnusedFieldsForType() {
	if s.Container() {
		s.SkillNonContainerOnlyEditData = SkillNonContainerOnlyEditData{}
		s.Difficulty = AttributeDifficulty{omit: true}
		if s.TemplatePicker == nil {
			s.TemplatePicker = &TemplatePicker{}
		}
	} else {
		s.SkillContainerOnlySyncData = SkillContainerOnlySyncData{}
		s.Children = nil
		s.Difficulty.omit = false
		if s.IsTechnique() {
			s.Defaults = nil
		} else {
			s.TechniqueDefault = nil
			s.TechniqueLimitModifier = nil
		}
	}
}

// GetSource returns the source of this data.
func (s *Skill) GetSource() Source {
	return s.Source
}

// ClearSource clears the source of this data.
func (s *Skill) ClearSource() {
	s.Source = Source{}
}

// SyncWithSource synchronizes this data with the source.
func (s *Skill) SyncWithSource() {
	if !xreflect.IsNil(s.owner) {
		if state, data := s.owner.SourceMatcher().Match(s); state == srcstate.Mismatched {
			if other, ok := data.(*Skill); ok {
				s.SkillSyncData = other.SkillSyncData
				s.Tags = slices.Clone(other.Tags)
				if s.Container() {
					s.SkillContainerOnlySyncData = other.SkillContainerOnlySyncData
					s.TemplatePicker = other.TemplatePicker.Clone()
				} else {
					s.SkillNonContainerOnlySyncData = other.SkillNonContainerOnlySyncData
					if len(other.Defaults) != 0 {
						s.Defaults = make([]*SkillDefault, len(other.Defaults))
						for i, def := range other.Defaults {
							def2 := *def
							s.Defaults[i] = &def2
						}
					}
					if other.TechniqueDefault != nil {
						def := *other.TechniqueDefault
						s.TechniqueDefault = &def
						if !DefaultTypeIsSkillBased(other.TechniqueDefault.DefaultType) {
							s.TechniqueDefault.Name = criteria.Text{}
							s.TechniqueDefault.Specialization = criteria.Text{}
						}
					}
					if other.TechniqueLimitModifier != nil {
						mod := *other.TechniqueLimitModifier
						s.TechniqueLimitModifier = &mod
					}
					s.Prereq = other.Prereq.CloneResolvingEmpty(false, true)
					s.Weapons = CloneWeapons(other.Weapons, false)
					s.Features = other.Features.Clone()
				}
			}
		}
	}
}

// Hash writes this object's contents into the hasher. Note that this only hashes the data that is considered to be
// "source" data, i.e. not expected to be modified by the user after copying from a library.
func (s *Skill) Hash(h hash.Hash) {
	s.SkillSyncData.hash(h)
	if s.Container() {
		s.SkillContainerOnlySyncData.hash(h)
	} else {
		s.SkillNonContainerOnlySyncData.hash(h)
	}
}

func (s *SkillSyncData) hash(h hash.Hash) {
	xhash.StringWithLen(h, s.Name)
	xhash.StringWithLen(h, s.PageRef)
	xhash.StringWithLen(h, s.PageRefHighlight)
	xhash.StringWithLen(h, s.LocalNotes)
	xhash.Num64(h, len(s.Tags))
	for _, tag := range s.Tags {
		xhash.StringWithLen(h, tag)
	}
}

func (s *SkillContainerOnlySyncData) hash(h hash.Hash) {
	s.TemplatePicker.Hash(h)
}

func (s *SkillNonContainerOnlySyncData) hash(h hash.Hash) {
	xhash.StringWithLen(h, s.Specialization)
	s.Difficulty.Hash(h)
	xhash.Num64(h, s.EncumbrancePenaltyMultiplier)
	xhash.Num64(h, len(s.Defaults))
	for _, one := range s.Defaults {
		one.Hash(h)
	}
	if s.TechniqueDefault != nil {
		s.TechniqueDefault.Hash(h)
	} else {
		xhash.Num8(h, uint8(255))
	}
	if s.TechniqueLimitModifier != nil {
		xhash.Num64(h, *s.TechniqueLimitModifier)
	} else {
		xhash.Num8(h, uint8(255))
	}
	s.Prereq.Hash(h)
	xhash.Num64(h, len(s.Weapons))
	for _, weapon := range s.Weapons {
		weapon.Hash(h)
	}
	xhash.Num64(h, len(s.Features))
	for _, feature := range s.Features {
		feature.Hash(h)
	}
}

// CopyFrom implements node.EditorData.
func (s *SkillEditData) CopyFrom(other *Skill) {
	s.copyFrom(&other.SkillEditData, other.Container(), false, other.IsTechnique())
}

// ApplyTo implements node.EditorData.
func (s *SkillEditData) ApplyTo(other *Skill) {
	other.copyFrom(s, other.Container(), true, other.IsTechnique())
}

func (s *SkillEditData) copyFrom(other *SkillEditData, isContainer, isApply, isTechnique bool) {
	*s = *other
	s.Tags = slices.Clone(other.Tags)
	s.Replacements = maps.Clone(other.Replacements)
	if other.TechLevel != nil {
		tl := *other.TechLevel
		s.TechLevel = &tl
	}
	if other.DefaultedFrom != nil {
		def := *other.DefaultedFrom
		s.DefaultedFrom = &def
	}
	s.Defaults = nil
	s.TechniqueDefault = nil
	s.TechniqueLimitModifier = nil
	if isTechnique {
		if other.TechniqueDefault != nil {
			def := *other.TechniqueDefault
			s.TechniqueDefault = &def
			if !DefaultTypeIsSkillBased(other.TechniqueDefault.DefaultType) {
				s.TechniqueDefault.Name.Qualifier = ""
				s.TechniqueDefault.Specialization.Qualifier = ""
			}
		}
		if other.TechniqueLimitModifier != nil {
			mod := *other.TechniqueLimitModifier
			s.TechniqueLimitModifier = &mod
		}
	} else if len(other.Defaults) != 0 {
		s.Defaults = make([]*SkillDefault, len(other.Defaults))
		for i, def := range other.Defaults {
			def2 := *def
			s.Defaults[i] = &def2
		}
	}
	s.Prereq = s.Prereq.CloneResolvingEmpty(isContainer, isApply)
	s.Weapons = CloneWeapons(other.Weapons, isApply)
	s.Features = other.Features.Clone()
	if len(other.Study) != 0 {
		s.Study = make([]*Study, len(other.Study))
		for i := range other.Study {
			s.Study[i] = other.Study[i].Clone()
		}
	}
	s.TemplatePicker = other.TemplatePicker.Clone()
}
