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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/study"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/gcs/v5/model/txtutil"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/zeebo/xxh3"
)

var (
	_ = assertNode[*Skill]
	_ = assertTemplatePickerNode[*Skill]
	_ = assertEditorData[*SkillEditData]

	_ TechLevelProvider         = &Skill{}
	_ SkillAdjustmentProvider   = &Skill{}
	_ TemplatePickerProvider    = &Skill{}
	_ templateChoiceConvertible = &Skill{}
	_ FeatureSwitcher           = &Skill{}
	_ LeveledOwner              = &Skill{}
	_ SkillLevelStepper         = &Skill{}

	_ TemplatePickerProvider = &SkillData{}
	_ TemplatePickerProvider = &SkillEditData{}
	_ TemplatePickerProvider = &SkillContainerOnlySyncData{}
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
	SkillSwitchColumn
)

// SkillAdjustmentProvider interface for objects that can have their skill level adjusted.
type SkillAdjustmentProvider interface {
	RawPointsAdjuster
	IncrementSkillLevel()
	DecrementSkillLevel()
}

// SkillLevelStepper is what the shared bodies of IncrementSkillLevel and DecrementSkillLevel need from a skill or
// spell, so that the point-stepping search lives once rather than in each type.
type SkillLevelStepper interface {
	Enabled() bool
	RawPoints() fxp.Int
	SetRawPoints(points fxp.Int) bool
	WildcardDifficulty() bool
	CurrentCalculatedLevel() fxp.Int
	// CalculatedLevelBeforeGrants is CurrentCalculatedLevel without what the grants made of the level (see
	// grantState.applyLevels), which goes on rising with the points while a Keep Highest grant holds the level shown.
	CalculatedLevelBeforeGrants() fxp.Int
}

// Skill holds the data for a skill.
type Skill struct {
	SkillData
	owner             DataOwner
	LevelData         Level
	UnsatisfiedReason string
	// takesEquipmentPenalty records that the last recalculation pass found this skill's prerequisites unmet on account
	// of equipment they require to be equipped, so that the next pass applies the missing-equipment penalty to its
	// level (see Entity.applyEquipmentPenalties). It is never saved.
	takesEquipmentPenalty bool
	grant                 grantState
}

// SkillData holds the Skill data that is written to disk.
type SkillData struct {
	SourcedID
	Granted *GrantedBy `json:"granted_by,omitempty"`
	SkillEditData
	ThirdParty map[string]any `json:"third_party,omitempty"`
	Children   []*Skill       `json:"children,omitempty"` // Only for containers
	parent     *Skill
}

// SkillEditData holds the Skill data that can be edited by the UI detail editor.
type SkillEditData struct {
	SkillSyncData
	VTTNotes     string            `json:"vtt_notes,omitzero"`
	Replacements map[string]string `json:"replacements,omitempty"`
	Disabled     bool              `json:"disabled,omitzero"`
	ItemSwitch
	preconfigurable
	choiceGrouping
	SkillNonContainerOnlyEditData
	SkillContainerOnlySyncData
	// copiedDefaultsHash is defaultsHash(Defaults) as of CopyFrom, which lets ApplyTo tell whether the defaults have
	// been edited since. Only meaningful on edit data that CopyFrom produced.
	copiedDefaultsHash uint64
}

// SkillNonContainerOnlyEditData holds the Skill data that is only applicable to skills that aren't containers.
type SkillNonContainerOnlyEditData struct {
	SkillNonContainerOnlySyncData
	OptionalSpecialization string        `json:"optional_specialization,omitzero"`
	TechLevel              *string       `json:"tech_level,omitzero"`
	Points                 fxp.Int       `json:"points,omitzero"`
	DefaultedFrom          *SkillDefault `json:"defaulted_from,omitzero"`
	Study                  []*Study      `json:"study,omitempty"`
	StudyHoursNeeded       study.Level   `json:"study_hours_needed,omitzero"`
}

// SkillSyncData holds the skill sync data that is common to both containers and non-containers.
type SkillSyncData = NodeSyncData

// SkillNonContainerOnlySyncData holds the Skill sync data that is only applicable to skills that aren't containers.
type SkillNonContainerOnlySyncData struct {
	Specialization               string              `json:"specialization,omitzero"`
	Difficulty                   AttributeDifficulty `json:"difficulty,omitzero"`
	EncumbrancePenaltyMultiplier fxp.Int             `json:"encumbrance_penalty_multiplier,omitzero"`
	Defaults                     []*SkillDefault     `json:"defaults,omitempty"`
	TechniqueDefault             *SkillDefault       `json:"default,omitzero"`
	TechniqueLimitModifier       *fxp.Int            `json:"limit,omitzero"`
	Prereq                       *PrereqList         `json:"prereqs,omitzero"`
	Weapons                      []*Weapon           `json:"weapons,omitempty"`
	Features                     Features            `json:"features,omitempty"`
}

// SkillContainerOnlySyncData holds the skill sync data that is only applicable to skills that are containers.
type SkillContainerOnlySyncData struct {
	TemplatePicker TemplatePicker `json:"template_picker,omitzero"`
}

// NewSkillsFromFile loads a Skill list from a file.
func NewSkillsFromFile(fileSystem fs.FS, filePath string) ([]*Skill, error) {
	rows, err := loadRows[*Skill](fileSystem, filePath)
	if err != nil {
		return nil, err
	}
	// Fix up some bad data in standalone skill lists where Hard techniques incorrectly had 1 point assigned to them
	// instead of 2.
	Traverse(func(skill *Skill) bool {
		if skill.IsTechnique() && skill.Difficulty.Difficulty == difficulty.Hard && skill.Points == fxp.One {
			skill.Points = fxp.Two
		}
		return false
	}, false, true, rows...)
	return rows, nil
}

// SaveSkills writes the Skill list to the file as JSON.
func SaveSkills(skills []*Skill, filePath string) error {
	return saveRows(filePath, skills)
}

// NewSkillChoiceContainer creates a new template choice container for skills.
func NewSkillChoiceContainer(owner DataOwner, parent *Skill) *Skill {
	s := NewSkill(owner, parent, true)
	s.TemplatePicker = newTemplateChoicePicker()
	s.Name = s.Kind()
	return s
}

func (s *Skill) canPickSeparately() bool {
	return s.Container() && s.TemplatePicker.IsZero()
}

func (s *Skill) canBecomeTemplateChoiceContainer() bool {
	return true
}

func (s *Skill) templateChoiceContainerExclusions() []string {
	var list []string
	if s.VTTNotes != "" {
		list = append(list, i18n.Text("VTT notes"))
	}
	if len(s.Tags) != 0 {
		list = append(list, i18n.Text("tags"))
	}
	return list
}

func (s *Skill) clearTemplateChoiceContainerExclusions() {
	s.VTTNotes = ""
	s.Tags = nil
}

func (s *Skill) clearStudy() {
	s.Study = nil
	s.StudyHoursNeeded = study.Standard
}

// NewSkill creates a new Skill.
func NewSkill(owner DataOwner, parent *Skill, container bool) *Skill {
	var s Skill
	s.TID = tid.MustNewTID(skillKind(container))
	s.parent = parent
	s.owner = owner
	if !container {
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
func (s *Skill) Clone(from LibraryFile, owner DataOwner, parent *Skill, mode CloneMode) *Skill {
	var other *Skill
	if s.IsTechnique() {
		var skillName string
		if s.TechniqueDefault != nil {
			skillName = s.TechniqueDefault.Name.Qualifier
		}
		other = NewTechnique(owner, parent, skillName)
	} else {
		other = NewSkill(owner, parent, s.Container())
		other.SetOpen(s.IsOpen())
	}
	other.AdjustSource(from, s.SourcedID, mode)
	other.ThirdParty = s.ThirdParty
	other.copyFrom(other, &s.SkillEditData, s.Container(), false, mode, s.IsTechnique())
	other.Granted = cloneGrantedBy(s.Granted, s, other)
	if s.Granted != nil && other.Granted == nil {
		// Where no grant claims the copy, it is bought with the points of the tier the grant stood for.
		other.Points = EntityFromNode(s).grantTierPointsFor(s)
	}
	if EntityFromNode(other) == nil {
		// Disabling only means something on a sheet.
		other.SetDisabled(false)
	}
	PropagateNodeNoteClosedState(s, other)
	if s.HasChildren() {
		other.Children = make([]*Skill, 0, len(s.Children))
		for _, child := range s.Children {
			other.Children = append(other.Children, child.Clone(from, owner, other, mode))
		}
	}
	return other
}

// MarshalJSONTo implements json.MarshalerTo.
func (s *Skill) MarshalJSONTo(enc *jsontext.Encoder) error {
	s.ClearUnusedFieldsForType()
	slices.Sort(s.Tags)
	return marshalNodeData(enc, &s.SkillData, func() *leveledCalc {
		return newLeveledCalc(s.Container(), s.LevelData.Level, s.RelativeLevel(), s.UnsatisfiedReason,
			s.ResolveLocalNotes(), s.LocalNotes)
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
	open := fixupLegacyTID(&localData.TID, localData.Type, func(container bool) byte {
		if localData.Type == "technique" {
			return kinds.Technique
		}
		return skillKind(container)
	}) && localData.IsOpen
	s.SkillData = localData.SkillData
	s.Weapons = dropNilNodes(s.Weapons)
	s.Replacements = nameable.Normalize(s.Replacements)
	s.Defaults = slices.DeleteFunc(s.Defaults, func(one *SkillDefault) bool { return one == nil })
	if s.TechniqueDefault != nil && s.TechniqueDefault.SkillBased() {
		s.TechniqueDefault.Name.Compare = criteria.IsText
	}
	migrateLegacyText(&s.LocalNotes, localData.ExprNotes)
	s.ClearUnusedFieldsForType()
	finishNodeUnmarshal(s, &s.Tags, localData.Categories, open)
	return nil
}

// TemplatePickerData implements TemplatePickerProvider.
func (s *SkillContainerOnlySyncData) TemplatePickerData() ([]picker.Type, *TemplatePicker) {
	return picker.TypesForSkills, &s.TemplatePicker
}

// SkillsHeaderData returns the header data information for the given skill column.
func SkillsHeaderData(columnID int) HeaderData {
	var data HeaderData
	switch columnID {
	case SkillDescriptionColumn:
		data.Title = i18n.Text("Skill / Technique")
		data.Primary = true
	case SkillDifficultyColumn:
		data = abbreviatedHeaderData(i18n.Text("Diff"), i18n.Text("Difficulty"))
	case SkillTagsColumn:
		data = tagsHeaderData()
	case SkillReferenceColumn:
		data = pageRefHeaderData()
	case SkillLevelColumn:
		data = abbreviatedHeaderData(i18n.Text("SL"), i18n.Text("Skill Level"))
	case SkillRelativeLevelColumn:
		data = abbreviatedHeaderData(i18n.Text("RSL"), i18n.Text("Relative Skill Level"))
	case SkillPointsColumn:
		data = abbreviatedHeaderData(i18n.Text("Pts"), i18n.Text("Points"))
		data.Less = PointsLessFromString
	case SkillLibSrcColumn:
		data = libSrcHeaderData()
	case SkillSwitchColumn:
		data = switchHeaderData()
	}
	return data
}

// CellData returns the cell data information for the given column.
func (s *Skill) CellData(columnID int, data *CellData) {
	data.Self = s
	data.Dim = !s.Enabled()
	switch columnID {
	case SkillDescriptionColumn:
		data.Type = cell.Text
		data.Primary = s.String()
		// A granted one whose grant is not in effect is dimmed but not struck through.
		data.Disabled = data.Dim && !s.grantNotInEffect()
		data.GrantedBy = grantedByText(s.Granted, &s.grant)
		data.Grants = EntityFromNode(s).grantSummary(s)
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
		fillTagsCell(data, s.Tags)
	case SkillReferenceColumn, PageRefCellAlias:
		fillPageRefCell(data, s.PageRef, s.PageRefHighlight, s.NameWithReplacements)
	case SkillLevelColumn:
		if !s.Container() {
			data.Type = cell.Text
			// The level is the one the last recalculation recorded, which is what every other reader of a skill's
			// level uses; a disabled one has no level, which LevelAsString shows as "-".
			data.Primary = s.LevelData.LevelAsString(s.Container())
			if s.LevelData.Tooltip != "" {
				data.Tooltip = IncludesModifiersFrom() + ":" + s.LevelData.Tooltip
			}
			data.Alignment = align.End
		}
	case SkillRelativeLevelColumn:
		if !s.Container() {
			data.Type = cell.Text
			data.Primary = FormatRelativeSkill(EntityFromNode(s), s.IsTechnique(), s.AdjustedDifficulty(),
				s.AdjustedRelativeLevel())
			if s.LevelData.Tooltip != "" {
				data.Tooltip = IncludesModifiersFrom() + ":" + s.LevelData.Tooltip
			}
		}
	case SkillPointsColumn:
		// An organizing group is never picked as a whole, so its total means nothing.
		if IsOrganizingGroup(s) {
			break
		}
		var tooltip xbytes.InsertBuffer
		r := s.PointsRange(&tooltip)
		if tooltip.Len() != 0 {
			data.Tooltip = IncludesModifiersFrom() + ":" + tooltip.String()
		}
		fillPointsCell(data, r)
	case SkillLibSrcColumn:
		fillLibSrcCell(data, s.owner, s)
	case SkillSwitchColumn:
		if s.HasSwitchableFeatures() {
			data.Type = cell.Switch
			data.Name = SwitchCellName()
			data.Checked = s.SwitchedOn
			data.Alignment = align.Middle
			// A container never gets here (see HasSwitchableFeatures), so there are never contents to cascade to.
			data.Tooltip = SwitchCellTooltip(false)
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
	return def.bestMatchingSkill(e, s.Replacements, requirePoints, nil)
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

// ActiveFeatures returns the features of this skill that currently take effect, i.e. all of them except any switchable
// ones while the skill's switch is off.
func (s *Skill) ActiveFeatures() Features {
	return s.Features.Active(s.SwitchedOn)
}

// HasSwitchableFeatures implements FeatureSwitcher.
func (s *Skill) HasSwitchableFeatures() bool {
	return !s.Container() && s.Features.AnySwitchable()
}

// TagList returns the list of tags.
func (s *Skill) TagList() []string {
	return s.Tags
}

// RatedStrength always returns 0 for skills.
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
		txtutil.AppendStringOntoNewLine(&buffer, strings.TrimSpace(s.Notes()))
		txtutil.AppendStringOntoNewLine(&buffer, StudyHoursProgressText(ResolveStudyHours(s.Study), s.StudyHoursNeeded, false))
	}
	addTooltipForSkillLevelAdj(optionChecker, prefs, s.LevelData, &buffer)
	return buffer.String()
}

func addTooltipForSkillLevelAdj(optionChecker func(display.Option) bool, prefs *SheetSettings, level Level, to txtutil.LineBuilder) {
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
			txtutil.AppendStringOntoNewLine(to, levelTooltip)
		}
	}
}

func (s *Skill) String() string {
	var buffer strings.Builder
	buffer.WriteString(s.NameWithReplacements())
	parts := appendTitleNotes(nil, s.Replacements, true, s.Features)
	if !s.Container() {
		if s.TechLevel != nil {
			buffer.WriteString("/TL")
			buffer.WriteString(*s.TechLevel)
		}
		specialization := s.SpecializationWithReplacements()
		optionalSpecialization := s.OptionalSpecializationWithReplacements()
		if specialization != "" {
			parts = append(parts, specialization)
		}
		if optionalSpecialization != "" {
			parts = append(parts, optionalSpecialization)
		}
	}
	buffer.WriteString(parenthetical(appendTitleNotes(parts, s.Replacements, false, s.Features)))
	return buffer.String()
}

// TitleNotes returns the text of the skill's title note features, those placed before the built-in parts first.
func (s *Skill) TitleNotes() []string {
	return appendTitleNotes(appendTitleNotes(nil, s.Replacements, true, s.Features), s.Replacements, false, s.Features)
}

// IsLeveled implements LeveledOwner.
func (s *Skill) IsLeveled() bool {
	return !s.Container()
}

// CurrentLevel implements LeveledOwner.
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
		if s.IsTechnique() && s.TechniqueDefault != nil {
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

// SetRawPoints sets the unadjusted points and updates the level. Returns true if the level, or the default the skill
// settled on, changed.
func (s *Skill) SetRawPoints(points fxp.Int) bool {
	s.Points = points
	return s.UpdateLevel()
}

// AdjustedPoints returns the points, adjusted for any bonuses, or nothing when disabled. A container whose choice
// outcomes all cost the same reports that cost; see PointsRange for one whose outcomes differ.
func (s *Skill) AdjustedPoints(tooltip *xbytes.InsertBuffer) fxp.Int {
	return s.adjustedPointsExcept(nil, tooltip)
}

// PointsCost returns what the skill costs the character: its adjusted points, or nothing for a granted skill, whose
// adjusted points stand for its tier rather than for points spent. It is what the sheet's points column shows, so the
// exports and scripts report it.
func (s *Skill) PointsCost() fxp.Int {
	if s.Granted != nil {
		return 0
	}
	return s.AdjustedPoints(nil)
}

// PointsWithGrants returns the points with those grants gave the skill at the last recalculation added, as
// AdjustedPoints adds them, for a level computed from points other than the skill's own.
func (s *Skill) PointsWithGrants(points fxp.Int) fxp.Int {
	return s.grant.applyPoints(nil, points, nil)
}

// LevelWithGrants returns the level with the levels grants merged into the skill at the last recalculation added, as
// CalculateLevel adds them, for a level computed from data other than the skill's own. levelAt returns the level that
// data gives with the given points in place of the skill's own, which the level a Keep Highest grant keeps is read
// with.
func (s *Skill) LevelWithGrants(level Level, levelAt func(points fxp.Int) fxp.Int) Level {
	return s.grant.applyLevels(nil, level, levelAt)
}

// pointsForPrereqsOf returns the skill's points as the prerequisites of exclude see them: as AdjustedPoints does, but
// without the points grants carried by exclude gave it, since an item never meets its own prerequisites with what it
// grants.
func (s *Skill) pointsForPrereqsOf(exclude any) fxp.Int {
	return s.adjustedPointsExcept(exclude, nil)
}

// pointsWithoutGrants returns the given points, the skill's own or those standing for them, adjusted for the bonuses,
// as AdjustedPoints adjusts them, but without what grants gave it.
func (s *Skill) pointsWithoutGrants(points fxp.Int) fxp.Int {
	return AdjustedPointsForNonContainerSkillOrTechnique(EntityFromNode(s), points, s.NameWithReplacements(),
		s.SpecializationWithReplacements(), s.OptionalSpecializationWithReplacements(), s.Tags, nil)
}

// levelForPrereqsOf returns the skill's level as the prerequisites of exclude see it: the level the last pass recorded,
// unless exclude carries a grant, in which case it is computed afresh without what the grants exclude carries gave the
// skill, nor what the items they put on the sheet give it through their bonuses or a default resting on them (see
// Entity.JudgingPrereqsOf), since an item never meets its own prerequisites with what it grants. The grants need not be
// in effect this pass, since the recorded level may hold what they gave in the last.
func (s *Skill) levelForPrereqsOf(exclude any) fxp.Int {
	if !EntityFromNode(s).carriesGrant(exclude) {
		return s.LevelData.Level
	}
	return s.calculateLevelWith(exclude, s.bestDefaultWithPoints(nil), nil).Level
}

// adjustedPointsExcept is AdjustedPoints without the points grants carried by exclude gave the skill, or by the item
// whose prerequisites are being judged when exclude is nil; see grantState.pointsExcept and Entity.judgingOr.
func (s *Skill) adjustedPointsExcept(exclude any, tooltip *xbytes.InsertBuffer) fxp.Int {
	if !s.Enabled() {
		return 0
	}
	exclude = EntityFromNode(s).judgingOr(exclude)
	if s.Container() {
		// The tooltip goes no further: a container never puts anything into it; see PointsRange for why.
		if !s.TemplatePicker.IsZero() {
			return pickerContainerPoints(s.TemplatePicker, TemplateChoiceOptions(s))
		}
		var total fxp.Int
		for _, one := range s.Children {
			// A granted skill costs nothing, whatever points its tier stands for.
			if one.Granted == nil {
				total += one.AdjustedPoints(nil)
			}
		}
		return total
	}
	return s.grant.applyPoints(exclude, AdjustedPointsForNonContainerSkillOrTechnique(EntityFromNode(s), s.Points,
		s.NameWithReplacements(), s.SpecializationWithReplacements(), s.OptionalSpecializationWithReplacements(), s.Tags,
		tooltip), tooltip)
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

// PointsRange returns the span of point costs this skill may end up being worth once every choice in or under it has
// been made; with no choice left, it holds the PointsCost value. The tooltip may be nil, and only a non-container
// fills it: the notes name each bonus source without saying which row it landed on, so rolling up a container's
// children would give an unattributed, repetitive pile.
func (s *Skill) PointsRange(tooltip *xbytes.InsertBuffer) NumericRange {
	if !s.Enabled() || s.Granted != nil {
		// A granted skill costs nothing, whatever points its tier stands for.
		return NumericRangeOf(0)
	}
	if !s.Container() {
		return NumericRangeOf(s.AdjustedPoints(tooltip))
	}
	if value, settled := settledPickerCost(s.TemplatePicker); settled {
		return NumericRangeOf(value)
	}
	// A picker with nothing to pick from, and a container carrying no picker at all, both come back as the total of
	// the children, which is what everything inside a container being taken costs.
	return pointsRangeForPicker(s.TemplatePicker, childPointsRanges(TemplateChoiceOptions(s)))
}

// RawPointsRange returns the same span as PointsRange, but with every skill in it counted by its raw points rather than
// its adjusted ones. This is what a template picker counts, since it measures what is being bought rather than what
// the sheet's bonuses make of it.
func (s *Skill) RawPointsRange() NumericRange {
	if !s.Enabled() {
		return NumericRangeOf(0)
	}
	if !s.Container() {
		return NumericRangeOf(s.RawPoints())
	}
	return containerRawPointsRange(s.TemplatePicker, TemplateChoiceOptions(s))
}

// AdjustedPointsForNonContainerSkillOrTechnique returns the points, adjusted for any bonuses.
func AdjustedPointsForNonContainerSkillOrTechnique(e *Entity, points fxp.Int, name, specialization, optionalSpecialization string, tags []string, tooltip *xbytes.InsertBuffer) fxp.Int {
	if e != nil {
		points += e.SkillPointBonusFor(name, specialization, optionalSpecialization, tags, tooltip)
		points = points.Max(0)
	}
	return points
}

// IncrementSkillLevel adds enough points to increment the skill level to the next level. Containers and disabled
// skills are left alone.
func (s *Skill) IncrementSkillLevel() {
	if !s.Container() {
		incrementSkillLevel(s)
	}
}

// DecrementSkillLevel removes enough points to decrement the skill level to the previous level. Containers and
// disabled skills are left alone.
func (s *Skill) DecrementSkillLevel() {
	if !s.Container() {
		decrementSkillLevel(s)
	}
}

// WildcardDifficulty implements SkillLevelStepper.
func (s *Skill) WildcardDifficulty() bool {
	return s.Difficulty.Difficulty == difficulty.Wildcard
}

// CurrentCalculatedLevel implements SkillLevelStepper.
func (s *Skill) CurrentCalculatedLevel() fxp.Int {
	return s.CalculateLevel(nil).Level
}

// CalculatedLevelBeforeGrants implements SkillLevelStepper.
func (s *Skill) CalculatedLevelBeforeGrants() fxp.Int {
	return s.levelBeforeGrants(nil, s.DefaultedFrom, nil).Level
}

// skillLevelStride returns the widest span of points that needs to be searched to find the next or previous level.
// Wildcard skills cost three times as much per level as other skills, so their span is three times as wide.
func skillLevelStride(p SkillLevelStepper) fxp.Int {
	if p.WildcardDifficulty() {
		return fxp.Twelve
	}
	return fxp.Four
}

// incrementSkillLevel is the shared body of Skill.IncrementSkillLevel and Spell.IncrementSkillLevel. A disabled item
// has no level to step, since AdjustedPoints reports nothing for it, so it is left alone, as is one with no higher
// level to reach, such as a technique at its limit or a skill whose attribute is missing. The search goes on a stride
// at a time while the level the points buy is still rising, since a Keep Highest grant holds the level shown until the
// level bought passes the one kept, which may be several strides off, and stops once a stride passes without the level
// bought rising, as no further points would raise it.
func incrementSkillLevel(p SkillLevelStepper) {
	if !p.Enabled() {
		return
	}
	original := p.RawPoints()
	stride := skillLevelStride(p)
	oldLevel := p.CurrentCalculatedLevel()
	bought := p.CalculatedLevelBeforeGrants()
	for basePoints := original.Floor() + fxp.One; ; basePoints += stride {
		for points := basePoints; points < basePoints+stride; points += fxp.One {
			p.SetRawPoints(points)
			if p.CurrentCalculatedLevel() > oldLevel {
				return
			}
		}
		next := p.CalculatedLevelBeforeGrants()
		if next <= bought {
			break
		}
		bought = next
	}
	p.SetRawPoints(original)
}

// decrementSkillLevel is the shared body of Skill.DecrementSkillLevel and Spell.DecrementSkillLevel. A disabled item
// is left alone for the same reason as in incrementSkillLevel.
func decrementSkillLevel(p SkillLevelStepper) {
	if !p.Enabled() || p.RawPoints() <= 0 {
		return
	}
	basePoints := p.RawPoints().Floor()
	minPoints := (basePoints - skillLevelStride(p)).Max(0)
	oldLevel := p.CurrentCalculatedLevel()
	for points := basePoints; points >= minPoints; points -= fxp.One {
		p.SetRawPoints(points)
		if p.CurrentCalculatedLevel() < oldLevel {
			break
		}
	}
	// Having found a lower level, back off to the fewest points that still produce it.
	oldLevel = p.CurrentCalculatedLevel()
	for p.RawPoints() > 0 {
		p.SetRawPoints((p.RawPoints() - fxp.One).Max(0))
		if p.CurrentCalculatedLevel() != oldLevel {
			p.SetRawPoints(p.RawPoints() + fxp.One)
			break
		}
	}
}

// CalculateLevel returns the computed level without updating it.
func (s *Skill) CalculateLevel(excludes map[string]bool) Level {
	return s.calculateLevelExcept(nil, excludes)
}

// calculateLevelExcept is CalculateLevel without the levels and points grants carried by exclude gave the skill (see
// grantState.applyLevelNotes), and with its default picked afresh while the prerequisites of an item carrying a grant
// are judged; see defaultExcept.
func (s *Skill) calculateLevelExcept(exclude any, excludes map[string]bool) Level {
	return s.calculateLevelWith(exclude, s.defaultExcept(exclude, excludes), excludes)
}

// defaultExcept returns the default the skill recorded or, when exclude (nil, or the item whose prerequisites are being
// judged; see Entity.judgingOr) carries a grant and the skill recorded one, a default picked afresh as those
// prerequisites see the sheet, since the recorded one was picked, and its level taken, with what the grants gave the
// skills it rests on. It is picked as bestDefault picks, but with the levels taken with the skill added to the excludes
// given rather than with it alone excluded, so that a chain of defaults never comes back to a skill it passes through:
// the recorded default, if the skill keeps it (see keepsDefault), it resolves and it does not lead back to the skill,
// and otherwise the best of the skill's defaults. Each pick is kept while those prerequisites are judged and reused
// only for the same skill with the same excludes, so that the answers are those of picking afresh every time, while a
// pick beneath a skill weighing several defaults is not made again for each of them, and so on down every chain. The
// picks are still as many as the distinct excludes a skill is reached with, which double with each member of a set of
// skills with points that default to one another, so judging stays fast only while such a set stays small.
func (s *Skill) defaultExcept(exclude any, excludes map[string]bool) *SkillDefault {
	e := EntityFromNode(s)
	if s.DefaultedFrom == nil || !e.carriesGrant(e.judgingOr(exclude)) {
		return s.DefaultedFrom
	}
	excludes = maps.Clone(excludes)
	if excludes == nil {
		excludes = make(map[string]bool)
	}
	excludes[s.String()] = true
	key := newJudgedDefaultKey(s, excludes)
	if def, ok := e.judgedDefaults[key]; ok {
		return def
	}
	var def, excluded *SkillDefault
	if s.keepsDefault() && !s.inDefaultChain(s.DefaultedFrom, make(map[*Skill]bool)) {
		if level := s.calcSkillDefaultLevel(s.DefaultedFrom, excludes); level != fxp.Min {
			def = s.DefaultedFrom.CloneWithoutLevelOrPoints()
			def.Level = level
		} else {
			excluded = s.DefaultedFrom
		}
	}
	if def == nil {
		def = s.bestDefaultWith(excluded, excludes)
	}
	if def != nil {
		s.setDefaultPoints(def)
	}
	if e.judgedDefaults == nil {
		e.judgedDefaults = make(map[judgedDefaultKey]*SkillDefault)
	}
	e.judgedDefaults[key] = def
	return def
}

// judgedDefaultKey identifies the default Skill.defaultExcept picks for a skill with the levels taken with a set of
// excludes, which it holds as their names sorted and joined.
type judgedDefaultKey struct {
	skill    *Skill
	excludes string
}

// newJudgedDefaultKey returns the key of the default the skill picks with the levels taken with the given excludes.
func newJudgedDefaultKey(s *Skill, excludes map[string]bool) judgedDefaultKey {
	names := make([]string, 0, len(excludes))
	for name, excluded := range excludes {
		if excluded {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return judgedDefaultKey{skill: s, excludes: strings.Join(names, "\x00")}
}

// calculateLevelWith is calculateLevelExcept with the given default in place of the one defaultExcept returns. With
// exclude nil, what the grants of the item whose prerequisites are being judged gave is left out; see Entity.judgingOr.
func (s *Skill) calculateLevelWith(exclude any, def *SkillDefault, excludes map[string]bool) Level {
	e := EntityFromNode(s)
	exclude = e.judgingOr(exclude)
	level := s.levelBeforeGrants(exclude, def, excludes)
	if s.IsTechnique() {
		return level
	}
	return s.grant.applyLevels(exclude, level, func(tierPoints fxp.Int) fxp.Int {
		return s.levelAt(e, def, tierPoints).Level
	})
}

// levelBeforeGrants is calculateLevelWith without what the grants made of the level, though with the points they gave;
// see grantState.applyLevels.
func (s *Skill) levelBeforeGrants(exclude any, def *SkillDefault, excludes map[string]bool) Level {
	points := s.adjustedPointsExcept(exclude, nil)
	if s.IsTechnique() {
		return CalculateTechniqueLevel(EntityFromNode(s), s.Replacements, s.NameWithReplacements(),
			s.SpecializationWithReplacements(), s.Tags, s.TechniqueDefault, s.AdjustedDifficulty().Difficulty, points, true,
			s.TechniqueLimitModifier, excludes)
	}
	return CalculateSkillLevel(EntityFromNode(s), s.NameWithReplacements(), s.SpecializationWithReplacements(),
		s.OptionalSpecializationWithReplacements(), s.Tags, def, s.AdjustedDifficulty(), points,
		s.EncumbrancePenaltyMultiplier)
}

// CalculateSkillLevel returns the calculated level for a skill.
func CalculateSkillLevel(e *Entity, name, specialization, optionalSpecialization string, tags []string, def *SkillDefault, attrDiff AttributeDifficulty, points, encumbrancePenaltyMultiplier fxp.Int) Level {
	var tooltip xbytes.InsertBuffer
	relativeLevel := attrDiff.Difficulty.BaseRelativeLevel()
	level := e.resolveAttributeCurrentForPrereqs(attrDiff.Attribute)
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
		if def.Type() == SkillID {
			// Match against the full name and specialization criteria (not just the qualifier strings) so that, e.g.,
			// a "whose specialization is anything" default resolves to any matching specialization rather than staying
			// locked on the qualifier text that may linger from a prior "is" selection. See issue #1061.
			if list := def.matchingSkills(e, replacements, requirePoints, excludes); len(list) > 0 {
				if excludes == nil {
					excludes = make(map[string]bool)
				}
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
						// Exclude this skill while computing its own level so it cannot circularly default back to
						// itself, then restore the prior state so it remains available to the other candidates.
						skStr := sk.String()
						had := excludes[skStr]
						excludes[skStr] = true
						if candidate := sk.CalculateLevel(excludes).Level; level < candidate {
							level = candidate
						}
						if !had {
							delete(excludes, skStr)
						}
					}
				}
			}
		} else if t := def.Type(); (t == ParryID || t == BlockID) && e.carriesGrant(e.judging) {
			// While the prerequisites of an item carrying a grant are judged, the skill a parry or block default rests
			// on is taken at the level they see, as it is for a skill default, rather than the one it recorded, which
			// holds what the grant gave; see Entity.JudgingPrereqsOf. Each skill is excluded while its own level is
			// computed, as above, so that one leading back to this technique cannot recurse without end.
			if excludes == nil {
				excludes = make(map[string]bool)
			}
			for _, sk := range def.matchingSkills(e, replacements, true, excludes) {
				if !def.isTLPermitted(e, sk.TL()) {
					continue
				}
				skStr := sk.String()
				had := excludes[skStr]
				excludes[skStr] = true
				if candidate := sk.CalculateLevel(excludes).Level; level < candidate {
					level = candidate
				}
				if !had {
					delete(excludes, skStr)
				}
			}
			if level != fxp.Min {
				level = defenseLevelFromSkill(level, e.defenseBonus(t))
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

// UpdateLevel updates the level of the skill, returning true if it, or the default the skill settled on, has changed.
func (s *Skill) UpdateLevel() bool {
	savedLevel := s.LevelData
	if !s.Enabled() {
		s.LevelData = Level{}
		return savedLevel != s.LevelData
	}
	savedDefaultedFrom := s.DefaultedFrom
	var level Level
	if anyScriptAbandonedDuring(EntityFromNode(s), func() {
		// Set the default before computing the level, since CalculateLevel reads it from the skill. Computing against
		// the previous pass's default would leave the level a pass behind, and since Entity.Recalculate stops once the
		// levels settle, that pass would never come.
		s.DefaultedFrom = s.bestDefaultWithPoints(nil)
		level = s.CalculateLevel(nil)
	}) {
		// A script was stopped before it could answer, so both values came from a stand-in. Keep the old ones, which
		// are merely out of date rather than wrong -- DefaultedFrom is written to disk, so it would be wrong there too.
		// The next recalculation that gets through its scripts puts both right.
		s.DefaultedFrom = savedDefaultedFrom
		return false
	}
	s.LevelData = level
	// A change to the default is reported too, since other skills read it: a skill judging whether it may default to
	// this one, or a technique judging whether this one is usable as its base.
	return savedLevel != s.LevelData || !sameSkillDefault(savedDefaultedFrom, s.DefaultedFrom)
}

// sameSkillDefault returns true if the two defaults, either of which may be nil, hold the same values.
func sameSkillDefault(a, b *SkillDefault) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func (s *Skill) bestDefaultWithPoints(excluded *SkillDefault) *SkillDefault {
	if s.IsTechnique() {
		return nil
	}
	best := s.bestDefault(excluded)
	if best != nil {
		s.setDefaultPoints(best)
	}
	return best
}

// setDefaultPoints sets the default's adjusted level to its level rounded down, and its points to those that level
// stands for in the skill.
func (s *Skill) setDefaultPoints(def *SkillDefault) {
	baseLine := (EntityFromNode(s).resolveAttributeCurrentForPrereqs(s.AdjustedDifficulty().Attribute) +
		s.AdjustedDifficulty().Difficulty.BaseRelativeLevel()).Floor()
	level := def.Level.Floor()
	def.AdjLevel = level
	switch {
	case level == baseLine:
		def.Points = fxp.One
	case level == baseLine+fxp.One:
		def.Points = fxp.Two
	case level > baseLine+fxp.One:
		def.Points = fxp.Four.Mul(level - (baseLine + fxp.One))
	default:
		def.Points = -level.Max(0)
	}
}

func (s *Skill) bestDefault(excluded *SkillDefault) *SkillDefault {
	return s.bestDefaultWith(excluded, map[string]bool{s.String(): true})
}

// keepsDefault returns true if the skill keeps the default it recorded for as long as that resolves. A skill with
// points that already records which default it uses keeps it, rather than silently jumping to whichever default
// currently yields the highest level; the user changes it explicitly via "Swap Defaults". Skills with no points, and
// those whose defaults are purely synthetic (the automatic -2 default to a sibling's optional specialization), always
// track the dynamically-best default.
func (s *Skill) keepsDefault() bool {
	return s.DefaultedFrom != nil && s.AdjustedPoints(nil) > 0 && len(s.Defaults) > 0
}

// bestDefaultWith is bestDefault with the levels of the defaults taken with the given excludes, which hold the skill.
func (s *Skill) bestDefaultWith(excluded *SkillDefault, excludes map[string]bool) *SkillDefault {
	if EntityFromNode(s) == nil {
		return nil
	}
	keepCurrent := excluded == nil && s.keepsDefault()
	var bestDef, preferredDef *SkillDefault
	best := fxp.Min
	for _, def := range s.resolveToSpecificDefaults() {
		// Skip the excluded default and any that circle back through this skill.
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
// resulting level to the lowest. Candidates that currently default back to this skill are included: the user may choose
// one, and SwapToNextDefault breaks the resulting cycle so the chosen skill re-resolves to its own best default.
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
// choose among them via "Swap Defaults". Only skills that declare their own defaults are swappable; purely synthetic
// (optional-specialization) defaults track the best default automatically.
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
	// this explicit choice wins; the recalculation below resolves it to its own best remaining default, keeping the two
	// from defaulting to each other.
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
		if other := def.bestMatchingSkill(e, s.Replacements, true, excludes); other != nil {
			// The bonus being undone was added inside other.CalculateLevel(), which queries it with the other skill's
			// own qualifiers and tags, so all of them must come from 'other' here, not from this skill.
			level -= e.SkillBonusFor(other.NameWithReplacements(), other.SpecializationWithReplacements(),
				other.OptionalSpecializationWithReplacements(), other.Tags, nil)
		}
	}
	return level
}

func (s *Skill) inDefaultChain(def *SkillDefault, lookedAt map[*Skill]bool) bool {
	e := EntityFromNode(s)
	if e == nil || def == nil || !def.SkillBased() {
		return false
	}
	for _, one := range def.matchingSkills(e, s.Replacements, true, nil) {
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
		if def == nil {
			continue
		}
		if e == nil || !def.SkillBased() {
			result = append(result, def)
		} else {
			for _, one := range def.matchingSkills(e, s.Replacements, true, map[string]bool{s.String(): true}) {
				local := *def
				// The copy is pinned to the matched skill by name and specialization, so the tag criteria has already
				// done its job. Leaving it on would carry it into the persisted DefaultedFrom and make an expansion of
				// a tag default look different from a plain default to the same skill, so that neither the swap list
				// nor the "keep the current default" check could pair them.
				local.Tags = criteria.Text{}
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
			// a fully-unspecialized skill would pick up defaults to differently-specialized siblings.
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
	if !s.IsTechnique() || s.TechniqueDefault == nil || !s.TechniqueDefault.SkillBased() {
		return true
	}
	e := EntityFromNode(s)
	// The skill is one the default selects that has points, as the technique's level is computed from those (see
	// Entity.SkillMatching), so that a granted skill, whose points are those of its tier, counts, and a skill without
	// points, which may be the highest of them by defaulting from another, does not stand in the way of one that has.
	if s.TechniqueDefault.bestMatchingSkill(e, s.Replacements, true, nil) != nil {
		return true
	}
	if tooltip != nil {
		tooltip.WriteString(prefix)
		// A default that names its skill outright reads as "a skill named X"; one that selects it by other criteria is
		// described by FullName, and the sentence is shaped around that description instead.
		named := s.TechniqueDefault.namesSkill(s.Replacements)
		sk := s.TechniqueDefault.bestMatchingSkill(e, s.Replacements, false, nil)
		switch {
		case sk == nil && named:
			tooltip.WriteString(i18n.Text("Requires a skill named "))
		case sk == nil:
			tooltip.WriteString(i18n.Text("Requires "))
		case named:
			tooltip.WriteString(i18n.Text("Requires at least 1 point in the skill named "))
		default:
			tooltip.WriteString(i18n.Text("Requires at least 1 point in "))
		}
		tooltip.WriteString(s.TechniqueDefault.FullName(e, s.Replacements))
	}
	return false
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

// Enabled returns true if this Skill and all of its parents are enabled, and, for a granted skill, its grant is in
// effect.
func (s *Skill) Enabled() bool {
	for p := s; p != nil; p = p.parent {
		if p.IsDisabled() || p.grantNotInEffect() {
			return false
		}
	}
	return true
}

// IsDisabled implements Disableable.
func (s *SkillEditData) IsDisabled() bool {
	return s.Disabled
}

// SetDisabled implements Disableable.
func (s *SkillEditData) SetDisabled(disabled bool) {
	s.Disabled = disabled
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

// Notes returns the resolved local notes.
func (s *Skill) Notes() string {
	return s.ResolveLocalNotes()
}

// ModifierNotes returns the notes due to modifiers.
func (s *Skill) ModifierNotes() string {
	if s.IsTechnique() && s.TechniqueDefault != nil {
		def := s.TechniqueDefault
		text := def.FullName(EntityFromNode(s), s.Replacements)
		// Show the skill that actually matched rather than the default's own qualifier text, which may be stale (e.g.
		// when the specialization comparison is "anything" but a specialization was previously entered). See issue
		// #1061.
		if def.SkillBased() {
			if defSkill := s.DefaultSkill(); defSkill != nil {
				var buffer strings.Builder
				buffer.WriteString(defSkill.String())
				switch def.Type() {
				case ParryID:
					buffer.WriteString(i18n.Text(" Parry"))
				case BlockID:
					buffer.WriteString(i18n.Text(" Block"))
				}
				text = buffer.String()
			}
		}
		return i18n.Text("Default: ") + text + def.ModifierAsString()
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
	nameable.Extract(
		m, existing,
		s.Name,
		s.LocalNotes,
		s.Specialization,
		s.OptionalSpecialization,
	)
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
	// A granted skill changes only with its grant, which would remake it with its default as it was.
	return !s.IsTechnique() && !s.Container() && s.Granted == nil && s.AdjustedPoints(nil) > 0
}

// CanSwapDefaultsWith returns true if this skill's default can be swapped with the other skill, which a granted skill
// never is, for the reason CanSwapDefaults gives.
func (s *Skill) CanSwapDefaultsWith(other *Skill) bool {
	return other != nil && other.Granted == nil && s.CanSwapDefaults() && other.HasDefaultTo(s)
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
		if !s.TemplatePicker.IsZero() {
			return i18n.Text("Skill Choice")
		}
		return i18n.Text("Skill Container")
	}
	return i18n.Text("Skill")
}

// ClearUnusedFieldsForType zeroes out the fields that are not applicable to this type (container vs not-container).
func (s *Skill) ClearUnusedFieldsForType() {
	if !s.canPickSeparately() {
		s.PickSeparately = false
	}
	if s.Container() {
		s.SkillNonContainerOnlyEditData = SkillNonContainerOnlyEditData{}
		// A container's features never apply, so it can never have anything to switch (see HasSwitchableFeatures).
		// Clearing the switch keeps a stale on-state from being carried around with no way for the user to reach it.
		s.ItemSwitch = ItemSwitch{}
		s.Difficulty = AttributeDifficulty{omit: true}
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

// SyncWithSource synchronizes this data with the source.
func (s *Skill) SyncWithSource() {
	syncFromSource(s, func(other *Skill) {
		s.SkillSyncData = other.SkillSyncData
		s.Tags = slices.Clone(other.Tags)
		s.Nameables = nameable.CloneDetails(other.Nameables)
		if s.Container() {
			s.SkillContainerOnlySyncData = other.SkillContainerOnlySyncData
			return
		}
		// As in SkillEditData.ApplyTo: a recorded default is only kept while the declared defaults it was chosen from
		// stay the same.
		if defaultsHash(s.Defaults) != defaultsHash(other.Defaults) {
			s.DefaultedFrom = nil
		}
		s.SkillNonContainerOnlySyncData = other.SkillNonContainerOnlySyncData
		s.Defaults = cloneSkillDefaults(other.Defaults)
		s.TechniqueDefault = cloneSkillDefault(other.TechniqueDefault)
		s.TechniqueLimitModifier = clonePtr(other.TechniqueLimitModifier)
		s.Prereq = other.Prereq.CloneResolvingEmpty(false, true)
		s.Weapons = CloneWeapons(other.Weapons, s, Reference)
		s.Features = other.Features.Clone()
	})
}

// Hash writes this object's contents into the hasher. Note that this only hashes the data that is considered to be
// "source" data, i.e. not expected to be modified by the user after copying from a library.
func (s *Skill) Hash(h hash.Hash) {
	xhash.Num8(h, syncKind(s))
	s.SkillSyncData.hash(h)
	if s.Container() {
		s.SkillContainerOnlySyncData.hash(h)
	} else {
		s.SkillNonContainerOnlySyncData.hash(h)
	}
}

func (s *SkillContainerOnlySyncData) hash(h hash.Hash) {
	s.TemplatePicker.Hash(h)
}

func (s *SkillNonContainerOnlySyncData) hash(h hash.Hash) {
	xhash.StringWithLen(h, s.Specialization)
	s.Difficulty.Hash(h)
	xhash.Num64(h, s.EncumbrancePenaltyMultiplier)
	hashList(h, s.Defaults)
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
	hashList(h, s.Weapons)
	hashList(h, s.Features)
}

// CopyFrom implements EditorData.
func (s *SkillEditData) CopyFrom(other *Skill) {
	s.copyFrom(other, &other.SkillEditData, other.Container(), false, Copy, other.IsTechnique())
	s.copiedDefaultsHash = defaultsHash(s.Defaults)
}

// NameableReplacements returns the replacements to be used with Nameables.
func (s *SkillEditData) NameableReplacements() map[string]string {
	return s.Replacements
}

// SetNameableReplacements sets the replacements to be used with Nameables.
func (s *SkillEditData) SetNameableReplacements(replacements map[string]string) {
	s.Replacements = replacements
}

// ApplyTo implements EditorData.
func (s *SkillEditData) ApplyTo(other *Skill) {
	// A recorded default persists across recalculations (see bestDefault) only while the declared defaults it was
	// chosen from stay the same; once they are edited, it is chosen afresh. Compare against the defaults as of
	// CopyFrom, not the target's current ones: an undo applies a snapshot whose recorded default was chosen from that
	// snapshot's own defaults.
	rechoose := defaultsHash(s.Defaults) != s.copiedDefaultsHash
	other.copyFrom(other, s, other.Container(), true, Copy, other.IsTechnique())
	slices.Sort(other.Tags)
	if rechoose {
		other.DefaultedFrom = nil
	}
}

// defaultsHash summarizes a list of declared defaults, in order, by their source data alone: the level and points a
// resolved default carries are ignored.
func defaultsHash(defaults []*SkillDefault) uint64 {
	h := xxh3.New()
	xhash.Num64(h, len(defaults))
	for _, def := range defaults {
		if def == nil {
			xhash.Num8(h, uint8(255))
		} else {
			def.Hash(h)
		}
	}
	return h.Sum64()
}

// copyFrom copies other into s. isApply distinguishes staging the editor's working copy from committing it back, and
// only affects how an empty Prereq list is resolved. mode controls how the weapons are cloned -- CopyFrom/ApplyTo above
// always use Copy, since that's staging or committing the same skill's own data; Clone passes its own mode through.
func (s *SkillEditData) copyFrom(skill *Skill, other *SkillEditData, isContainer, isApply bool, mode CloneMode, isTechnique bool) {
	*s = *other
	s.Tags = slices.Clone(other.Tags)
	s.Replacements = maps.Clone(other.Replacements)
	s.Nameables = nameable.CloneDetails(other.Nameables)
	s.TechLevel = clonePtr(other.TechLevel)
	s.DefaultedFrom = cloneSkillDefault(other.DefaultedFrom)
	s.Defaults = nil
	s.TechniqueDefault = nil
	s.TechniqueLimitModifier = nil
	if isTechnique {
		s.TechniqueDefault = cloneSkillDefault(other.TechniqueDefault)
		s.TechniqueLimitModifier = clonePtr(other.TechniqueLimitModifier)
	} else {
		s.Defaults = cloneSkillDefaults(other.Defaults)
	}
	s.Prereq = s.Prereq.CloneResolvingEmpty(isContainer, isApply)
	s.Weapons = CloneWeapons(other.Weapons, skill, mode)
	s.Features = other.Features.Clone()
	s.Study = cloneStudyList(other.Study)
}
