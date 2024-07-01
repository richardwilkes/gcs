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
	"github.com/richardwilkes/gcs/v5/model/message"
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

const skillListTypeKey = "skill_list"

// Skill holds the data for a skill.
type Skill struct {
	SkillData
	owner             DataOwner
	LevelData         Level
	UnsatisfiedReason string
}

// SkillData holds the Skill data that is written to disk.
type SkillData struct {
	TID    tid.TID `json:"id"`
	Source Source  `json:"source,omitempty"`
	SkillEditData
	ThirdParty map[string]any `json:"third_party,omitempty"`
	Children   []*Skill       `json:"children,omitempty"` // Only for containers
	parent     *Skill
}

// SkillEditData holds the Skill data that can be edited by the UI detail editor.
type SkillEditData struct {
	Name             string   `json:"name,omitempty"`
	PageRef          string   `json:"reference,omitempty"`
	PageRefHighlight string   `json:"reference_highlight,omitempty"`
	LocalNotes       string   `json:"notes,omitempty"`
	VTTNotes         string   `json:"vtt_notes,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	SkillNonContainerOnlyEditData
	SkillContainerOnlyEditData
}

// SkillNonContainerOnlyEditData holds the Skill data that is only applicable to skills that aren't containers.
type SkillNonContainerOnlyEditData struct {
	Specialization               string              `json:"specialization,omitempty"`
	TechLevel                    *string             `json:"tech_level,omitempty"`
	Difficulty                   AttributeDifficulty `json:"difficulty,omitempty"`
	Points                       fxp.Int             `json:"points,omitempty"`
	EncumbrancePenaltyMultiplier fxp.Int             `json:"encumbrance_penalty_multiplier,omitempty"`
	DefaultedFrom                *SkillDefault       `json:"defaulted_from,omitempty"`
	Defaults                     []*SkillDefault     `json:"defaults,omitempty"`
	TechniqueDefault             *SkillDefault       `json:"default,omitempty"`
	TechniqueLimitModifier       *fxp.Int            `json:"limit,omitempty"`
	Prereq                       *PrereqList         `json:"prereqs,omitempty"`
	Weapons                      []*Weapon           `json:"weapons,omitempty"`
	Features                     Features            `json:"features,omitempty"`
	Study                        []*Study            `json:"study,omitempty"`
	StudyHoursNeeded             study.Level         `json:"study_hours_needed,omitempty"`
}

// SkillContainerOnlyEditData holds the Skill data that is only applicable to skills that are containers.
type SkillContainerOnlyEditData struct {
	TemplatePicker *TemplatePicker `json:"template_picker,omitempty"`
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
		return nil, errs.NewWithCause(message.InvalidFileData(), err)
	}
	if data.Type != skillListTypeKey {
		return nil, errs.New(message.UnexpectedFileData())
	}
	if err := jio.CheckVersion(data.Version); err != nil {
		return nil, err
	}

	// Fix up some bad data in standalone skill lists where Hard techniques incorrectly had 1 point assigned to them
	// instead of 2.
	Traverse(func(skill *Skill) bool {
		if skill.IsTechnique() &&
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
		Version: jio.CurrentDataVersion,
		Rows:    skills,
	})
}

// NewSkill creates a new Skill.
func NewSkill(owner DataOwner, parent *Skill, container bool) *Skill {
	s := Skill{
		SkillData: SkillData{
			TID:    tid.MustNewTID(skillKind(container)),
			parent: parent,
		},
		owner: owner,
	}
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
	s := Skill{
		SkillData: SkillData{
			TID:    tid.MustNewTID(kinds.Technique),
			parent: parent,
		},
		owner: owner,
	}
	s.Difficulty.Difficulty = difficulty.Average
	s.Points = fxp.One
	if skillName == "" {
		skillName = i18n.Text("Skill")
	}
	s.TechniqueDefault = &SkillDefault{
		DefaultType: SkillID,
		Name:        skillName,
	}
	s.Name = s.Kind()
	return &s
}

// GetSource returns the source of this data.
func (s *Skill) GetSource() Source {
	return s.Source
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
		other = NewTechnique(owner, parent, s.TechniqueDefault.Name)
	} else {
		other = NewSkill(owner, parent, s.Container())
		other.SetOpen(s.IsOpen())
	}
	other.Source.LibraryFile = from
	other.Source.TID = s.TID
	if preserveID {
		other.TID = s.TID
	}
	other.ThirdParty = s.ThirdParty
	other.SkillEditData.CopyFrom(s)
	if s.HasChildren() {
		other.Children = make([]*Skill, 0, len(s.Children))
		for _, child := range s.Children {
			other.Children = append(other.Children, child.Clone(from, owner, other, preserveID))
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
		if localData.Type == "technique" {
			kind = kinds.Technique
		} else {
			kind = skillKind(strings.HasSuffix(localData.Type, ContainerKeyPostfix))
		}
		localData.TID = tid.MustNewTID(kind)
		setOpen = localData.IsOpen
	}
	s.SkillData = localData.SkillData
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
		data.Detail = message.PageRefTooltip()
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
		data.Detail = message.LibSrcTooltip()
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
			data.Primary = s.Difficulty.Description(EntityFromNode(s))
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
				data.Tooltip = message.IncludesModifiersFrom() + ":" + level.Tooltip
			}
			data.Alignment = align.End
		}
	case SkillRelativeLevelColumn:
		if !s.Container() {
			data.Type = cell.Text
			data.Primary = FormatRelativeSkill(EntityFromNode(s), s.IsTechnique(), s.Difficulty,
				s.AdjustedRelativeLevel())
			if tooltip := s.CalculateLevel(nil).Tooltip; tooltip != "" {
				data.Tooltip = message.IncludesModifiersFrom() + ":" + tooltip
			}
		}
	case SkillPointsColumn:
		data.Type = cell.Text
		var tooltip xio.ByteBuffer
		data.Primary = s.AdjustedPoints(&tooltip).String()
		data.Alignment = align.End
		if tooltip.Len() != 0 {
			data.Tooltip = message.IncludesModifiersFrom() + ":" + tooltip.String()
		}
	case SkillLibSrcColumn:
		data.Type = cell.Text
		data.Alignment = align.Middle
		if !toolbox.IsNil(s.owner) {
			state := s.owner.SourceMatcher().Match(s)
			data.Primary = state.AltString()
			data.Tooltip = state.String()
			if state != srcstate.Custom {
				data.Tooltip += "\n" + s.Source.String()
			}
		}
	}
}

// FormatRelativeSkill formats the relative skill for display.
func FormatRelativeSkill(e *Entity, numOnly bool, difficulty AttributeDifficulty, rsl fxp.Int) string {
	switch {
	case rsl == fxp.Min:
		return "-"
	case numOnly:
		return rsl.Trunc().StringWithSign()
	default:
		s := ResolveAttributeName(e, difficulty.Attribute)
		rsl = rsl.Trunc()
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

// DefaultSkill returns the skill currently defaulted to, or nil.
func (s *Skill) DefaultSkill() *Skill {
	e := EntityFromNode(s)
	if e == nil {
		return nil
	}
	if s.IsTechnique() {
		return e.BaseSkill(s.TechniqueDefault, true)
	}
	return e.BaseSkill(s.DefaultedFrom, true)
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
	if s.IsTechnique() {
		return i18n.Text("Default: ") + s.TechniqueDefault.FullName(EntityFromNode(s)) +
			s.TechniqueDefault.ModifierAsString()
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
		if level.Tooltip != "" && level.Tooltip != message.NoAdditionalModifiers() {
			levelTooltip := level.Tooltip
			msg := message.IncludesModifiersFrom()
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
	return EvalEmbeddedRegex.ReplaceAllStringFunc(s.LocalNotes, EntityFromNode(s).EmbeddedEval)
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
		return ResolveAttributeName(EntityFromNode(s), s.Difficulty.Attribute) + rsl.StringWithSign()
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
func (s *Skill) AdjustedPoints(tooltip *xio.ByteBuffer) fxp.Int {
	if s.Container() {
		var total fxp.Int
		for _, one := range s.Children {
			total += one.AdjustedPoints(tooltip)
		}
		return total
	}
	return AdjustedPointsForNonContainerSkillOrTechnique(EntityFromNode(s), s.Points, s.Name, s.Specialization,
		s.Tags, tooltip)
}

// AdjustedPointsForNonContainerSkillOrTechnique returns the points, adjusted for any bonuses.
func AdjustedPointsForNonContainerSkillOrTechnique(e *Entity, points fxp.Int, name, specialization string, tags []string, tooltip *xio.ByteBuffer) fxp.Int {
	if e != nil {
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
	if s.IsTechnique() {
		return CalculateTechniqueLevel(EntityFromNode(s), s.Name, s.Specialization, s.Tags, s.TechniqueDefault,
			s.Difficulty.Difficulty, points, true, s.TechniqueLimitModifier, excludes)
	}
	return CalculateSkillLevel(EntityFromNode(s), s.Name, s.Specialization, s.Tags, s.DefaultedFrom, s.Difficulty,
		points, s.EncumbrancePenaltyMultiplier)
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
				if sk.IsTechnique() {
					if sk.TechniqueDefault != nil &&
						(sk.TechniqueDefault.Name != name || sk.TechniqueDefault.Specialization != specialization) {
						level = sk.CalculateLevel(excludes).Level
					}
				} else {
					if sk.DefaultedFrom == nil ||
						(sk.DefaultedFrom.Name != name || sk.DefaultedFrom.Specialization != specialization) {
						level = sk.CalculateLevel(excludes).Level
					}
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
	if s.IsTechnique() {
		return nil
	}
	best := s.bestDefault(excluded)
	if best != nil {
		baseLine := (EntityFromNode(s).ResolveAttributeCurrent(s.Difficulty.Attribute) +
			s.Difficulty.Difficulty.BaseRelativeLevel()).Trunc()
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
	if EntityFromNode(s) == nil || len(s.Defaults) == 0 {
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
	e := EntityFromNode(s)
	level := def.SkillLevel(e, true, excludes, !s.IsTechnique())
	if def.SkillBased() {
		if other := e.BestSkillNamed(def.Name, def.Specialization, true, excludes); other != nil {
			level -= e.SkillBonusFor(def.Name, def.Specialization, s.Tags, nil)
		}
	}
	return level
}

func (s *Skill) inDefaultChain(def *SkillDefault, lookedAt map[*Skill]bool) bool {
	e := EntityFromNode(s)
	if e == nil || def == nil || !def.SkillBased() {
		return false
	}
	for _, one := range e.SkillNamed(def.Name, def.Specialization, true, nil) {
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
			for _, one := range e.SkillNamed(def.Name, def.Specialization, true,
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
	if !s.IsTechnique() || !s.TechniqueDefault.SkillBased() {
		return true
	}
	e := EntityFromNode(s)
	sk := e.BestSkillNamed(s.TechniqueDefault.Name, s.TechniqueDefault.Specialization, false, nil)
	satisfied := sk != nil && (sk.IsTechnique() || sk.Points > 0)
	if !satisfied && tooltip != nil {
		tooltip.WriteString(prefix)
		if sk == nil {
			tooltip.WriteString(i18n.Text("Requires a skill named "))
		} else {
			tooltip.WriteString(i18n.Text("Requires at least 1 point in the skill named "))
		}
		tooltip.WriteString(s.TechniqueDefault.FullName(e))
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
	if baseSkill := EntityFromNode(s).BaseSkill(s.bestDefault(nil), true); baseSkill != nil {
		s.DefaultedFrom = s.bestDefaultWithPoints(def)
		baseSkill.UpdateLevel()
		s.UpdateLevel()
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
		s.Specialization = ""
		s.TechLevel = nil
		s.Difficulty = AttributeDifficulty{omit: true}
		s.Points = 0
		s.EncumbrancePenaltyMultiplier = 0
		s.DefaultedFrom = nil
		s.Defaults = nil
		s.TechniqueDefault = nil
		s.TechniqueLimitModifier = nil
		s.Prereq = nil
		s.Weapons = nil
		s.Features = nil
		s.StudyHoursNeeded = study.Standard
		if s.TemplatePicker == nil {
			s.TemplatePicker = &TemplatePicker{}
		}
	} else {
		s.Children = nil
		s.Difficulty.omit = false
		s.TemplatePicker = nil
	}
}

// Hash writes this object's contents into the hasher. Note that this only hashes the data that is considered to be
// "source" data, i.e. not expected to be modified by the user after copying from a library.
func (s *Skill) Hash(h hash.Hash) {
	_, _ = h.Write([]byte(s.Name))
	_, _ = h.Write([]byte(s.PageRef))
	_, _ = h.Write([]byte(s.PageRefHighlight))
	_, _ = h.Write([]byte(s.LocalNotes))
	_, _ = h.Write([]byte(s.VTTNotes))
	for _, tag := range s.Tags {
		_, _ = h.Write([]byte(tag))
	}
	if s.Container() {
		s.TemplatePicker.Hash(h)
	} else {
		_, _ = h.Write([]byte(s.Specialization))
		if s.TechLevel != nil {
			_, _ = h.Write([]byte(*s.TechLevel))
		}
		s.Difficulty.Hash(h)
		_ = binary.Write(h, binary.LittleEndian, s.EncumbrancePenaltyMultiplier)
		for _, one := range s.Defaults {
			one.Hash(h)
		}
		if s.TechniqueDefault != nil {
			s.TechniqueDefault.Hash(h)
		}
		if s.TechniqueLimitModifier != nil {
			_ = binary.Write(h, binary.LittleEndian, s.TechniqueLimitModifier)
		}
		s.Prereq.Hash(h)
		for _, weapon := range s.Weapons {
			weapon.Hash(h)
		}
		for _, feature := range s.Features {
			feature.Hash(h)
		}
	}
}

// CopyFrom implements node.EditorData.
func (s *SkillEditData) CopyFrom(other *Skill) {
	s.copyFrom(other.owner, &other.SkillEditData, other.Container(), false)
}

// ApplyTo implements node.EditorData.
func (s *SkillEditData) ApplyTo(other *Skill) {
	other.SkillEditData.copyFrom(other.owner, s, other.Container(), true)
}

func (s *SkillEditData) copyFrom(owner DataOwner, other *SkillEditData, isContainer, isApply bool) {
	*s = *other
	s.Tags = txt.CloneStringSlice(s.Tags)
	if other.TechLevel != nil {
		tl := *other.TechLevel
		s.TechLevel = &tl
	}
	if other.DefaultedFrom != nil {
		def := *other.DefaultedFrom
		s.DefaultedFrom = &def
	}
	s.Defaults = nil
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
			s.TechniqueDefault.Name = ""
			s.TechniqueDefault.Specialization = ""
		}
	}
	if other.TechniqueLimitModifier != nil {
		mod := *other.TechniqueLimitModifier
		s.TechniqueLimitModifier = &mod
	}
	s.Prereq = s.Prereq.CloneResolvingEmpty(isContainer, isApply)
	s.Weapons = nil
	if len(other.Weapons) != 0 {
		s.Weapons = make([]*Weapon, len(other.Weapons))
		for i, w := range other.Weapons {
			s.Weapons[i] = w.Clone(LibraryFile{}, owner, nil, true)
		}
	}
	s.Features = other.Features.Clone()
	if len(other.Study) != 0 {
		s.Study = make([]*Study, len(other.Study))
		for i := range other.Study {
			s.Study[i] = other.Study[i].Clone()
		}
	}
	s.TemplatePicker = other.TemplatePicker.Clone()
}
