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
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"hash"
	"io/fs"
	"maps"
	"slices"
	"strconv"
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
)

var (
	_ = assertNode[*Spell]
	_ = assertTemplatePickerNode[*Spell]
	_ = assertEditorData[*SpellEditData]

	_ TechLevelProvider         = &Spell{}
	_ SkillAdjustmentProvider   = &Spell{}
	_ TemplatePickerProvider    = &Spell{}
	_ templateChoiceConvertible = &Spell{}
	_ FeatureSwitcher           = &Spell{}
	_ LeveledOwner              = &Spell{}
	_ SkillLevelStepper         = &Spell{}

	_ TemplatePickerProvider = &SpellData{}
	_ TemplatePickerProvider = &SpellEditData{}
	_ TemplatePickerProvider = &SpellContainerOnlySyncData{}
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
	SpellPrereqCountColumn
	SpellTagsColumn
	SpellReferenceColumn
	SpellLevelColumn
	SpellRelativeLevelColumn
	SpellPointsColumn
	SpellDescriptionForPageColumn
	SpellLibSrcColumn
	SpellSwitchColumn
)

// Spell holds the data for a spell.
type Spell struct {
	SpellData
	owner             DataOwner
	LevelData         Level
	UnsatisfiedReason string
	// takesEquipmentPenalty records that the last recalculation pass found this spell's prerequisites unmet on account
	// of equipment they require to be equipped, so that the next pass applies the missing-equipment penalty to its
	// level (see Entity.applyEquipmentPenalties). It is never saved.
	takesEquipmentPenalty bool
	grant                 grantState
}

// SpellData holds the Spell data that is written to disk.
type SpellData struct {
	SourcedID
	Granted *GrantedBy `json:"granted_by,omitempty"`
	SpellEditData
	ThirdParty map[string]any `json:"third_party,omitempty"`
	Children   []*Spell       `json:"children,omitempty"` // Only for containers
	parent     *Spell
}

// SpellEditData holds the Spell data that can be edited by the UI detail editor.
type SpellEditData struct {
	SpellSyncData
	VTTNotes     string            `json:"vtt_notes,omitzero"`
	Replacements map[string]string `json:"replacements,omitempty"`
	Disabled     bool              `json:"disabled,omitzero"`
	ItemSwitch
	preconfigurable
	choiceGrouping
	SpellNonContainerOnlyEditData
	SpellContainerOnlySyncData
}

// SpellNonContainerOnlyEditData holds the Spell data that is only applicable to spells that aren't containers.
type SpellNonContainerOnlyEditData struct {
	SpellNonContainerOnlySyncData
	TechLevel        *string     `json:"tech_level,omitzero"`
	Points           fxp.Int     `json:"points,omitzero"`
	Study            []*Study    `json:"study,omitempty"`
	StudyHoursNeeded study.Level `json:"study_hours_needed,omitzero"`
}

// SpellContainerOnlySyncData holds the spell sync data that is only applicable to spells that are containers.
type SpellContainerOnlySyncData struct {
	TemplatePicker TemplatePicker `json:"template_picker,omitzero"`
}

// SpellSyncData holds the spell sync data that is common to both containers and non-containers.
type SpellSyncData = NodeSyncData

// SpellNonContainerOnlySyncData holds the spell sync data that is only applicable to spells that aren't containers.
type SpellNonContainerOnlySyncData struct {
	Difficulty      AttributeDifficulty `json:"difficulty,omitzero"`
	College         CollegeList         `json:"college,omitempty"`
	PowerSource     string              `json:"power_source,omitzero"`
	Class           string              `json:"spell_class,omitzero"`
	Resist          string              `json:"resist,omitzero"`
	CastingCost     string              `json:"casting_cost,omitzero"`
	MaintenanceCost string              `json:"maintenance_cost,omitzero"`
	CastingTime     string              `json:"casting_time,omitzero"`
	Duration        string              `json:"duration,omitzero"`
	Item            string              `json:"item,omitzero"`
	RitualSkillName string              `json:"base_skill,omitzero"`
	PrereqCount     int                 `json:"prereq_count,omitzero"`
	Prereq          *PrereqList         `json:"prereqs,omitzero"`
	Weapons         []*Weapon           `json:"weapons,omitempty"`
	Features        Features            `json:"features,omitempty"`
}

// NewSpellsFromFile loads a Spell list from a file.
func NewSpellsFromFile(fileSystem fs.FS, filePath string) ([]*Spell, error) {
	return loadRows[*Spell](fileSystem, filePath)
}

// SaveSpells writes the Spell list to the file as JSON.
func SaveSpells(spells []*Spell, filePath string) error {
	return saveRows(filePath, spells)
}

// NewSpellChoiceContainer creates a new template choice container for spells.
func NewSpellChoiceContainer(owner DataOwner, parent *Spell) *Spell {
	s := NewSpell(owner, parent, true)
	s.TemplatePicker = newTemplateChoicePicker()
	s.Name = s.Kind()
	return s
}

func (s *Spell) canPickSeparately() bool {
	return s.Container() && s.TemplatePicker.IsZero()
}

func (s *Spell) canBecomeTemplateChoiceContainer() bool {
	return true
}

func (s *Spell) templateChoiceContainerExclusions() []string {
	var list []string
	if s.VTTNotes != "" {
		list = append(list, i18n.Text("VTT notes"))
	}
	if len(s.Tags) != 0 {
		list = append(list, i18n.Text("tags"))
	}
	return list
}

func (s *Spell) clearTemplateChoiceContainerExclusions() {
	s.VTTNotes = ""
	s.Tags = nil
}

func (s *Spell) clearStudy() {
	s.Study = nil
	s.StudyHoursNeeded = study.Standard
}

// NewSpell creates a new Spell.
func NewSpell(owner DataOwner, parent *Spell, container bool) *Spell {
	var s Spell
	s.TID = tid.MustNewTID(spellKind(container))
	s.parent = parent
	s.owner = owner
	if !container {
		s.Difficulty.Attribute = AttributeIDFor(EntityFromNode(&s), IntelligenceID)
		s.Difficulty.Difficulty = difficulty.Hard
		s.PowerSource = "Arcane"
		s.Class = "Regular"
		s.CastingCost = "1"
		s.CastingTime = "1 second"
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
	var s Spell
	s.TID = tid.MustNewTID(kinds.RitualMagicSpell)
	s.parent = parent
	s.owner = owner
	s.Difficulty.Attribute = AttributeIDFor(EntityFromNode(&s), IntelligenceID)
	s.Difficulty.Difficulty = difficulty.Hard
	s.PowerSource = "Arcane"
	s.Class = "Regular"
	s.CastingCost = "1"
	s.CastingTime = "1 second"
	s.Duration = "Instant"
	s.RitualSkillName = "Ritual Magic"
	s.Name = s.Kind()
	// Unlike a normal spell, a new ritual magic spell starts with no points, since it can still be used at its default
	// level from the ritual magic skill. This also updates the level.
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
func (s *Spell) Clone(from LibraryFile, owner DataOwner, parent *Spell, mode CloneMode) *Spell {
	var other *Spell
	if s.IsRitualMagic() {
		other = NewRitualMagicSpell(owner, parent, false)
	} else {
		other = NewSpell(owner, parent, s.Container())
		other.SetOpen(s.IsOpen())
	}
	other.AdjustSource(from, s.SourcedID, mode)
	other.ThirdParty = s.ThirdParty
	other.copyFrom(other, &s.SpellEditData, s.Container(), false, mode)
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
		other.Children = make([]*Spell, 0, len(s.Children))
		for _, child := range s.Children {
			other.Children = append(other.Children, child.Clone(from, owner, other, mode))
		}
	}
	return other
}

// MarshalJSONTo implements json.MarshalerTo.
func (s *Spell) MarshalJSONTo(enc *jsontext.Encoder) error {
	s.ClearUnusedFieldsForType()
	slices.Sort(s.Tags)
	slices.Sort(s.College)
	data := &s.SpellData
	if s.Granted != nil && omitCalc(enc) {
		data = withoutGrantDetails(data)
	}
	return marshalNodeData(enc, data, func() *leveledCalc {
		return newLeveledCalc(s.Container(), s.LevelData.Level, s.RelativeLevel(), s.UnsatisfiedReason,
			s.ResolveLocalNotes(), s.LocalNotes)
	})
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (s *Spell) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var localData struct {
		SpellData
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
		if localData.Type == "ritual_magic_spell" {
			return kinds.RitualMagicSpell
		}
		return spellKind(container)
	}) && localData.IsOpen
	s.SpellData = localData.SpellData
	s.Weapons = dropNilNodes(s.Weapons)
	s.Replacements = nameable.Normalize(s.Replacements)
	migrateLegacyText(&s.LocalNotes, localData.ExprNotes)
	s.ClearUnusedFieldsForType()
	finishNodeUnmarshal(s, &s.Tags, localData.Categories, open)
	return nil
}

// TemplatePickerData implements TemplatePickerProvider.
func (s *SpellContainerOnlySyncData) TemplatePickerData() ([]picker.Type, *TemplatePicker) {
	return picker.TypesForSpells, &s.TemplatePicker
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
		data = abbreviatedHeaderData(i18n.Text("Diff"), i18n.Text("Difficulty"))
	case SpellPrereqCountColumn:
		data = abbreviatedHeaderData(i18n.Text("P#"), i18n.Text("Prerequisite Count"))
	case SpellTagsColumn:
		data = tagsHeaderData()
	case SpellReferenceColumn:
		data = pageRefHeaderData()
	case SpellLevelColumn:
		data = abbreviatedHeaderData(i18n.Text("SL"), i18n.Text("Skill Level"))
	case SpellRelativeLevelColumn:
		data = abbreviatedHeaderData(i18n.Text("RSL"), i18n.Text("Relative Skill Level"))
	case SpellPointsColumn:
		data = abbreviatedHeaderData(i18n.Text("Pts"), i18n.Text("Points"))
		data.Less = PointsLessFromString
	case SpellLibSrcColumn:
		data = libSrcHeaderData()
	case SpellSwitchColumn:
		data = switchHeaderData()
	}
	return data
}

// CellData returns the cell data information for the given column.
func (s *Spell) CellData(columnID int, data *CellData) {
	data.Self = s
	data.Dim = !s.Enabled()
	switch columnID {
	case SpellDescriptionColumn:
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
	case SpellPrereqCountColumn:
		if !s.Container() {
			data.Type = cell.Text
			data.Alignment = align.End
			if s.PrereqCount > 0 {
				data.Primary = strconv.Itoa(s.PrereqCount)
			}
		}
	case SpellTagsColumn:
		fillTagsCell(data, s.Tags)
	case SpellReferenceColumn, PageRefCellAlias:
		fillPageRefCell(data, s.PageRef, s.PageRefHighlight, s.NameWithReplacements)
	case SpellLevelColumn:
		if !s.Container() {
			data.Type = cell.Text
			// The level is the one the last recalculation recorded, which is what every other reader of a spell's
			// level uses; a disabled one has no level, which LevelAsString shows as "-".
			data.Primary = s.LevelData.LevelAsString(s.Container())
			if s.LevelData.Tooltip != "" {
				data.Tooltip = IncludesModifiersFrom() + ":" + s.LevelData.Tooltip
			}
			data.Alignment = align.End
		}
	case SpellRelativeLevelColumn:
		if !s.Container() {
			data.Type = cell.Text
			// A ritual magic spell's relative level is measured against its ritual magic skill rather than against an
			// attribute, so it is shown as a bare signed number, just as RelativeLevel() and the spell editor do.
			data.Primary = FormatRelativeSkill(EntityFromNode(s), s.IsRitualMagic(), s.Difficulty,
				s.AdjustedRelativeLevel())
			if s.LevelData.Tooltip != "" {
				data.Tooltip = IncludesModifiersFrom() + ":" + s.LevelData.Tooltip
			}
		}
	case SpellPointsColumn:
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
			addPartToBuffer(&buffer, i18n.Text("Item"), s.ItemWithReplacements())
			if buffer.Len() != 0 {
				text := buffer.String()
				nd := SheetSettingsFor(EntityFromNode(s)).NotesDisplay
				if nd.Inline() {
					if data.Secondary == "" {
						data.Secondary = text
					} else {
						data.Secondary += "\n" + text
					}
				}
				if nd.Tooltip() {
					if data.Tooltip == "" {
						data.Tooltip = text
					} else {
						data.Tooltip += "\n" + text
					}
				}
			}
		}
	case SpellLibSrcColumn:
		fillLibSrcCell(data, s.owner, s)
	case SpellSwitchColumn:
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

func addPartToBuffer(buffer *strings.Builder, label, content string) {
	if content != "" && content != "-" {
		if buffer.Len() != 0 {
			buffer.WriteString("; ")
		}
		buffer.WriteString("**")
		buffer.WriteString(label)
		buffer.WriteString(":** ")
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
	if EntityFromNode(s) != nil && s.LevelData.Level > 0 {
		return s.LevelData.RelativeLevel
	}
	return fxp.Min
}

// UpdateLevel updates the level of the spell, returning true if it has changed.
func (s *Spell) UpdateLevel() bool {
	saved := s.LevelData
	if !s.Enabled() {
		s.LevelData = Level{}
		return saved != s.LevelData
	}
	var level Level
	// A level computed from a script that was stopped before it could produce an answer is not kept, for the reasons
	// given in Skill.UpdateLevel; the level already here is out of date rather than wrong.
	if anyScriptAbandonedDuring(EntityFromNode(s), func() { level = s.CalculateLevel() }) {
		return false
	}
	s.LevelData = level
	return saved != s.LevelData
}

// CalculateLevel returns the computed level without updating it.
func (s *Spell) CalculateLevel() Level {
	return s.calculateLevel(s.AdjustedPoints(nil))
}

// levelForPrereqsOf returns the spell's level as the prerequisites of exclude see it: the level the last pass recorded,
// unless exclude carries a grant, in which case it is computed afresh without what the grants exclude carries gave the
// spell, nor what the items they put on the sheet give it through their bonuses (see Entity.JudgingPrereqsOf), since an
// item never meets its own prerequisites with what it grants. The grants need not be in effect this pass, since the
// recorded level may hold what they gave in the last.
func (s *Spell) levelForPrereqsOf(exclude any) fxp.Int {
	if !EntityFromNode(s).carriesGrant(exclude) {
		return s.LevelData.Level
	}
	return s.calculateLevelExcept(exclude, s.adjustedPointsExcept(exclude, nil)).Level
}

// calculateLevel returns the level the spell has with the given adjusted points.
func (s *Spell) calculateLevel(points fxp.Int) Level {
	return s.calculateLevelExcept(nil, points)
}

// calculateLevelExcept is calculateLevel without the levels grants carried by exclude merged into the spell, or by the
// item whose prerequisites are being judged when exclude is nil; see grantState.applyLevelNotes and Entity.judgingOr.
func (s *Spell) calculateLevelExcept(exclude any, points fxp.Int) Level {
	exclude = EntityFromNode(s).judgingOr(exclude)
	return s.grant.applyLevels(exclude, s.levelFromAdjustedPoints(points), func(tierPoints fxp.Int) fxp.Int {
		return s.levelAt(EntityFromNode(s), tierPoints).Level
	})
}

// levelFromAdjustedPoints returns the level the spell has with the given adjusted points, without what grants made of
// it.
func (s *Spell) levelFromAdjustedPoints(points fxp.Int) Level {
	if s.IsRitualMagic() {
		return CalculateRitualMagicSpellLevel(EntityFromNode(s), s.NameWithReplacements(),
			s.PowerSourceWithReplacements(), s.RitualSkillNameWithReplacements(), s.PrereqCount,
			s.CollegeWithReplacements(), s.Tags, s.Difficulty, points)
	}
	return CalculateSpellLevel(EntityFromNode(s), s.NameWithReplacements(), s.PowerSourceWithReplacements(),
		s.CollegeWithReplacements(), s.Tags, s.Difficulty, points)
}

// IncrementSkillLevel adds enough points to increment the skill level to the next level. Containers and disabled
// spells are left alone.
func (s *Spell) IncrementSkillLevel() {
	if !s.Container() {
		incrementSkillLevel(s)
	}
}

// DecrementSkillLevel removes enough points to decrement the skill level to the previous level. Containers and
// disabled spells are left alone.
func (s *Spell) DecrementSkillLevel() {
	if !s.Container() {
		decrementSkillLevel(s)
	}
}

// WildcardDifficulty implements SkillLevelStepper.
func (s *Spell) WildcardDifficulty() bool {
	return s.Difficulty.Difficulty == difficulty.Wildcard
}

// CurrentCalculatedLevel implements SkillLevelStepper.
func (s *Spell) CurrentCalculatedLevel() fxp.Int {
	return s.CalculateLevel().Level
}

// CalculatedLevelBeforeGrants implements SkillLevelStepper.
func (s *Spell) CalculatedLevelBeforeGrants() fxp.Int {
	return s.levelFromAdjustedPoints(s.AdjustedPoints(nil)).Level
}

// CalculateSpellLevel returns the calculated spell level.
func CalculateSpellLevel(e *Entity, name, powerSource string, colleges, tags []string, attrDiff AttributeDifficulty, pts fxp.Int) Level {
	var tooltip xbytes.InsertBuffer
	relativeLevel := attrDiff.Difficulty.BaseRelativeLevel()
	level := fxp.Min
	if e != nil {
		pts = pts.Floor()
		level = e.resolveAttributeCurrentForPrereqs(attrDiff.Attribute)
		if attrDiff.Difficulty == difficulty.Wildcard {
			pts = pts.Div(fxp.Three).Floor()
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
			relativeLevel += fxp.One + pts.Div(fxp.Four).Floor()
		}
		if level != fxp.Min {
			relativeLevel += e.SpellBonusFor(name, powerSource, colleges, tags, &tooltip)
			relativeLevel = relativeLevel.Floor()
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
func CalculateRitualMagicSpellLevel(e *Entity, name, powerSource, ritualSkillName string, prereqCount int, colleges, tags []string, diff AttributeDifficulty, points fxp.Int) Level {
	// Start at fxp.Min rather than the zero value, since a zero value would be a phantom level 0 that beats every
	// negative or unresolvable (fxp.Min) college level.
	skillLevel := Level{Level: fxp.Min}
	if len(colleges) == 0 {
		skillLevel = determineRitualMagicSkillLevelForCollege(e, name, "", ritualSkillName, prereqCount,
			tags, diff, points)
	} else {
		for _, college := range colleges {
			possible := determineRitualMagicSkillLevelForCollege(e, name, college, ritualSkillName,
				prereqCount, tags, diff, points)
			if skillLevel.Level < possible.Level {
				skillLevel = possible
			}
		}
	}
	// As with CalculateSkillLevel() and CalculateTechniqueLevel(), don't apply bonuses to an unresolvable level, since
	// fxp.Min is a sentinel, not a number that can be adjusted.
	if e != nil && skillLevel.Level != fxp.Min {
		tooltip := &xbytes.InsertBuffer{}
		tooltip.WriteString(skillLevel.Tooltip)
		levels := e.SpellBonusFor(name, powerSource, colleges, tags, tooltip).Floor()
		skillLevel.Level += levels
		skillLevel.RelativeLevel += levels
		skillLevel.Tooltip = tooltip.String()
	}
	return skillLevel
}

func determineRitualMagicSkillLevelForCollege(e *Entity, name, college, ritualSkillName string, prereqCount int, tags []string, diff AttributeDifficulty, points fxp.Int) Level {
	// A blank name is matched as is, as RitualMagicSatisfied matches it, rather than as a wildcard.
	def := &SkillDefault{
		DefaultType:    SkillID,
		Name:           criteria.Text{TextData: criteria.TextData{Compare: criteria.IsText, Qualifier: ritualSkillName}},
		Specialization: criteria.Text{TextData: criteria.TextData{Compare: criteria.IsText, Qualifier: college}},
		Modifier:       fxp.FromInteger(-prereqCount),
	}
	var limit fxp.Int
	skillLevel := CalculateTechniqueLevel(e, nil, name, college, tags, def, diff.Difficulty, points, false,
		&limit, nil)
	// CalculateTechniqueLevel() does not add the default skill modifier to the relative level, only to the final level
	skillLevel.RelativeLevel += def.Modifier
	def.Specialization.Qualifier = ""
	def.Modifier -= fxp.Six
	// The fallback defaults from a generically-named skill that carries no specialization. SkillNamed() treats an empty
	// specialization as a wildcard, so the specialized skills are excluded to avoid matching one from another college.
	fallback := CalculateTechniqueLevel(e, nil, name, college, tags, def, diff.Difficulty, points, false,
		&limit, specializedRitualSkills(e, ritualSkillName))
	fallback.RelativeLevel += def.Modifier
	if skillLevel.Level >= fallback.Level {
		return skillLevel
	}
	return fallback
}

// specializedRitualSkills returns the set of skills (keyed by Skill.String()) named ritualSkillName that carry a
// specialization, for use as an excludes set when looking up a ritual skill that has none.
func specializedRitualSkills(e *Entity, ritualSkillName string) map[string]bool {
	if e == nil || ritualSkillName == "" {
		return nil
	}
	var excludes map[string]bool
	for _, sk := range e.SkillNamed(ritualSkillName, "", false, nil) {
		if sk.SpecializationWithReplacements() != "" {
			if excludes == nil {
				excludes = make(map[string]bool)
			}
			excludes[sk.String()] = true
		}
	}
	return excludes
}

// RitualMagicSatisfied returns true if the Ritual Magic Spell is satisfied.
func (s *Spell) RitualMagicSatisfied(tooltip *xbytes.InsertBuffer, prefix string) bool {
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
	ritual := s.RitualSkillNameWithReplacements()
	for _, college := range colleges {
		if len(e.SkillNamed(ritual, college, false, nil)) != 0 {
			return true
		}
	}
	// Fall back to a generically-named skill that carries no specialization, required or optional, as the level
	// calculation does. SkillNamed treats an empty specialization as a wildcard, so each result has to be checked to
	// ensure the skill truly has none.
	for _, sk := range e.SkillNamed(ritual, "", false, nil) {
		if sk.SpecializationWithReplacements() == "" && sk.OptionalSpecializationWithReplacements() == "" {
			return true
		}
	}
	if tooltip != nil {
		tooltip.WriteString(prefix)
		tooltip.WriteString(i18n.Text("Requires a skill named "))
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

// Notes returns the resolved local notes.
func (s *Spell) Notes() string {
	return s.ResolveLocalNotes()
}

// Rituals returns the rituals required to cast the spell.
func (s *Spell) Rituals() string {
	e := EntityFromNode(s)
	if s.Container() || e == nil || !e.SheetSettings.ShowSpellAdj {
		return ""
	}
	var buffer strings.Builder
	// The level is the one the last recalculation recorded, except for a disabled spell, whose recorded level is
	// cleared and whose adjusted points are zero; its rituals are still those of the level it would have if enabled,
	// which for a granted spell is the level its grant would give.
	level := s.LevelData.Level
	if !s.Enabled() {
		level = s.calculateLevel(AdjustedPointsForNonContainerSpell(e, s.Points+e.grantTierPointsFor(s),
			s.NameWithReplacements(), s.PowerSourceWithReplacements(), s.CollegeWithReplacements(), s.Tags, nil)).Level
	}
	ritualText := i18n.Text("Ritual")
	switch {
	case level < fxp.Ten:
		addPartToBuffer(&buffer, ritualText, i18n.Text("need both hands and feet free and must speak"))
		addPartToBuffer(&buffer, i18n.Text("Time"), i18n.Text("Doubled"))
		return buffer.String()
	case level < fxp.Fifteen:
		addPartToBuffer(&buffer, ritualText, i18n.Text("speak quietly and make a gesture"))
		return buffer.String()
	case level < fxp.Twenty:
		addPartToBuffer(&buffer, ritualText, i18n.Text("speak a word or two OR make a small gesture"))
		if strings.Contains(strings.ToLower(s.ClassWithReplacements()), "blocking") {
			return buffer.String()
		}
		addPartToBuffer(&buffer, i18n.Text("Cost"), "-1")
		return buffer.String()
	default:
		addPartToBuffer(&buffer, ritualText, i18n.Text("none"))
		adj := (level - fxp.Fifteen).Div(fxp.Five).AsInteger[int]()
		class := strings.ToLower(s.ClassWithReplacements())
		if !strings.Contains(class, "missile") {
			addPartToBuffer(&buffer, i18n.Text("Time"), i18n.Text("x1/%d, rounded up (min 1 second)", 1<<adj))
		}
		if !strings.Contains(class, "blocking") {
			addPartToBuffer(&buffer, i18n.Text("Cost"), fmt.Sprintf("-%d", adj+1))
		}
		return buffer.String()
	}
}

// IsLeveled implements LeveledOwner.
func (s *Spell) IsLeveled() bool {
	return !s.Container()
}

// CurrentLevel implements LeveledOwner.
func (s *Spell) CurrentLevel() fxp.Int {
	return s.LevelData.Level
}

// ActiveFeatures returns the features of this spell that currently take effect, i.e. all of them except any switchable
// ones while the spell's switch is off.
func (s *Spell) ActiveFeatures() Features {
	return s.Features.Active(s.SwitchedOn)
}

// HasSwitchableFeatures implements FeatureSwitcher.
func (s *Spell) HasSwitchableFeatures() bool {
	return !s.Container() && s.Features.AnySwitchable()
}

// TagList returns the list of tags.
func (s *Spell) TagList() []string {
	return s.Tags
}

// RatedStrength always returns 0 for spells.
func (s *Spell) RatedStrength() fxp.Int {
	return 0
}

// SecondaryText returns the less important information that should be displayed with the description.
func (s *Spell) SecondaryText(optionChecker func(display.Option) bool) string {
	var buffer strings.Builder
	prefs := SheetSettingsFor(EntityFromNode(s))
	if optionChecker(prefs.NotesDisplay) {
		txtutil.AppendStringOntoNewLine(&buffer, strings.TrimSpace(s.Notes()))
		txtutil.AppendStringOntoNewLine(&buffer, s.Rituals())
		txtutil.AppendStringOntoNewLine(&buffer, StudyHoursProgressText(ResolveStudyHours(s.Study), s.StudyHoursNeeded, false))
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
	buffer.WriteString(parenthetical(s.TitleNotes()))
	return buffer.String()
}

// TitleNotes returns the text of the spell's title note features, those placed before the built-in parts first.
func (s *Spell) TitleNotes() []string {
	return appendTitleNotes(appendTitleNotes(nil, s.Replacements, true, s.Features), s.Replacements, false, s.Features)
}

// ResolveLocalNotes resolves the local notes, running any embedded scripts to get the final result.
func (s *Spell) ResolveLocalNotes() string {
	return ResolveText(EntityFromNode(s), deferredNewScriptSpell(s), s.LocalNotesWithReplacements())
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

// AdjustedPoints returns the points, adjusted for any bonuses, or nothing when disabled. A container whose choice
// outcomes all cost the same reports that cost; see PointsRange for one whose outcomes differ.
func (s *Spell) AdjustedPoints(tooltip *xbytes.InsertBuffer) fxp.Int {
	return s.adjustedPointsExcept(nil, tooltip)
}

// PointsCost returns what the spell costs the character: its adjusted points, or nothing for a granted spell, whose
// adjusted points stand for its tier rather than for points spent. It is what the sheet's points column shows, so the
// exports and scripts report it.
func (s *Spell) PointsCost() fxp.Int {
	if s.Granted != nil {
		return 0
	}
	return s.AdjustedPoints(nil)
}

// PointsWithGrants returns the points with those grants gave the spell at the last recalculation added, as
// AdjustedPoints adds them, for a level computed from points other than the spell's own.
func (s *Spell) PointsWithGrants(points fxp.Int) fxp.Int {
	return s.grant.applyPoints(nil, points, nil)
}

// LevelWithGrants returns the level with the levels grants merged into the spell at the last recalculation added, as
// CalculateLevel adds them, for a level computed from data other than the spell's own. levelAt returns the level that
// data gives with the given points in place of the spell's own, which the level a Keep Highest grant keeps is read
// with.
func (s *Spell) LevelWithGrants(level Level, levelAt func(points fxp.Int) fxp.Int) Level {
	return s.grant.applyLevels(nil, level, levelAt)
}

// pointsForPrereqsOf returns the spell's points as the prerequisites of exclude see them: as AdjustedPoints does, but
// without the points grants carried by exclude gave it, since an item never meets its own prerequisites with what it
// grants.
func (s *Spell) pointsForPrereqsOf(exclude any) fxp.Int {
	return s.adjustedPointsExcept(exclude, nil)
}

// pointsWithoutGrants returns the given points, the spell's own or those standing for them, adjusted for the bonuses,
// as AdjustedPoints adjusts them, but without what grants gave it.
func (s *Spell) pointsWithoutGrants(points fxp.Int) fxp.Int {
	return AdjustedPointsForNonContainerSpell(EntityFromNode(s), points, s.NameWithReplacements(),
		s.PowerSourceWithReplacements(), s.CollegeWithReplacements(), s.Tags, nil)
}

// adjustedPointsExcept is AdjustedPoints without the points grants carried by exclude gave the spell, or by the item
// whose prerequisites are being judged when exclude is nil; see grantState.pointsExcept and Entity.judgingOr.
func (s *Spell) adjustedPointsExcept(exclude any, tooltip *xbytes.InsertBuffer) fxp.Int {
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
			// A granted spell costs nothing, whatever points its tier stands for.
			if one.Granted == nil {
				total += one.AdjustedPoints(nil)
			}
		}
		return total
	}
	return s.grant.applyPoints(exclude, AdjustedPointsForNonContainerSpell(EntityFromNode(s), s.Points,
		s.NameWithReplacements(), s.PowerSourceWithReplacements(), s.CollegeWithReplacements(), s.Tags, tooltip), tooltip)
}

// PointsRange returns the span of point costs this spell may end up being worth once every choice in or under it has
// been made; with no choice left, it holds the PointsCost value. The tooltip may be nil, and only a non-container
// fills it: the notes name each bonus source without saying which row it landed on, so rolling up a container's
// children would give an unattributed, repetitive pile.
func (s *Spell) PointsRange(tooltip *xbytes.InsertBuffer) NumericRange {
	if !s.Enabled() || s.Granted != nil {
		// A granted spell costs nothing, whatever points its tier stands for.
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

// RawPointsRange returns the same span as PointsRange, but with every spell in it counted by its raw points rather than
// its adjusted ones. This is what a template picker counts, since it measures what is being bought rather than what
// the sheet's bonuses make of it.
func (s *Spell) RawPointsRange() NumericRange {
	if !s.Enabled() {
		return NumericRangeOf(0)
	}
	if !s.Container() {
		return NumericRangeOf(s.RawPoints())
	}
	return containerRawPointsRange(s.TemplatePicker, TemplateChoiceOptions(s))
}

// AdjustedPointsForNonContainerSpell returns the points, adjusted for any bonuses.
func AdjustedPointsForNonContainerSpell(e *Entity, points fxp.Int, name, powerSource string, colleges, tags []string, tooltip *xbytes.InsertBuffer) fxp.Int {
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

// Enabled returns true if this Spell and all of its parents are enabled, and, for a granted spell, its grant is in
// effect.
func (s *Spell) Enabled() bool {
	for p := s; p != nil; p = p.parent {
		if p.IsDisabled() || p.grantNotInEffect() {
			return false
		}
	}
	return true
}

// IsDisabled implements Disableable.
func (s *SpellEditData) IsDisabled() bool {
	return s.Disabled
}

// SetDisabled implements Disableable.
func (s *SpellEditData) SetDisabled(disabled bool) {
	s.Disabled = disabled
}

// NameableReplacements returns the replacements to be used with Nameables.
func (s *Spell) NameableReplacements() map[string]string {
	return s.Replacements
}

// NameWithReplacements returns the name with any replacements applied.
func (s *Spell) NameWithReplacements() string {
	return nameable.Apply(s.Name, s.Replacements)
}

// LocalNotesWithReplacements returns the local notes with any replacements applied.
func (s *Spell) LocalNotesWithReplacements() string {
	return nameable.Apply(s.LocalNotes, s.Replacements)
}

// PowerSourceWithReplacements returns the power source with any replacements applied.
func (s *Spell) PowerSourceWithReplacements() string {
	return nameable.Apply(s.PowerSource, s.Replacements)
}

// ClassWithReplacements returns the class with any replacements applied.
func (s *Spell) ClassWithReplacements() string {
	return nameable.Apply(s.Class, s.Replacements)
}

// ResistWithReplacements returns the resist with any replacements applied.
func (s *Spell) ResistWithReplacements() string {
	return nameable.Apply(s.Resist, s.Replacements)
}

// CastingCostWithReplacements returns the casting cost with any replacements applied.
func (s *Spell) CastingCostWithReplacements() string {
	return nameable.Apply(s.CastingCost, s.Replacements)
}

// MaintenanceCostWithReplacements returns the maintenance cost with any replacements applied.
func (s *Spell) MaintenanceCostWithReplacements() string {
	return nameable.Apply(s.MaintenanceCost, s.Replacements)
}

// CastingTimeWithReplacements returns the casting time with any replacements applied.
func (s *Spell) CastingTimeWithReplacements() string {
	return nameable.Apply(s.CastingTime, s.Replacements)
}

// DurationWithReplacements returns the duration with any replacements applied.
func (s *Spell) DurationWithReplacements() string {
	return nameable.Apply(s.Duration, s.Replacements)
}

// ItemWithReplacements returns the item with any replacements applied.
func (s *Spell) ItemWithReplacements() string {
	return nameable.Apply(s.Item, s.Replacements)
}

// RitualSkillNameWithReplacements returns the ritual skill name with any replacements applied.
func (s *Spell) RitualSkillNameWithReplacements() string {
	return nameable.Apply(s.RitualSkillName, s.Replacements)
}

// CollegeWithReplacements returns the college(s) with any replacements applied.
func (s *Spell) CollegeWithReplacements() []string {
	return nameable.ApplyToList(s.College, s.Replacements)
}

// FillWithNameableKeys adds any nameable keys found to the provided map.
func (s *Spell) FillWithNameableKeys(m, existing map[string]string) {
	if existing == nil {
		existing = s.Replacements
	}
	nameable.Extract(
		m, existing,
		s.Name,
		s.LocalNotes,
		s.PowerSource,
		s.Class,
		s.Resist,
		s.CastingCost,
		s.MaintenanceCost,
		s.CastingTime,
		s.Duration,
		s.Item,
		s.RitualSkillName,
	)
	nameable.Extract(m, existing, s.College...)
	if s.Prereq != nil {
		s.Prereq.FillWithNameableKeys(m, existing)
	}
	for _, one := range s.Weapons {
		one.FillWithNameableKeys(m, existing)
	}
	for _, one := range s.Features {
		one.FillWithNameableKeys(m, existing)
	}
}

// ApplyNameableKeys replaces any nameable keys found with the corresponding values in the provided map.
func (s *Spell) ApplyNameableKeys(m map[string]string) {
	needed := make(map[string]string)
	s.FillWithNameableKeys(needed, nil)
	s.Replacements = nameable.Reduce(needed, m)
}

// Kind returns the kind of data.
func (s *Spell) Kind() string {
	if s.IsRitualMagic() {
		return i18n.Text("Ritual Magic Spell")
	}
	if s.Container() {
		if !s.TemplatePicker.IsZero() {
			return i18n.Text("Spell Choice")
		}
		return i18n.Text("Spell Container")
	}
	return i18n.Text("Spell")
}

// ClearUnusedFieldsForType zeroes out the fields that are not applicable to this type (container vs not-container).
func (s *Spell) ClearUnusedFieldsForType() {
	if !s.canPickSeparately() {
		s.PickSeparately = false
	}
	if s.Container() {
		s.SpellNonContainerOnlyEditData = SpellNonContainerOnlyEditData{}
		// A container's features never apply, so it can never have anything to switch (see HasSwitchableFeatures).
		// Clearing the switch keeps a stale on-state from being carried around with no way for the user to reach it.
		s.ItemSwitch = ItemSwitch{}
		s.Difficulty = AttributeDifficulty{omit: true}
	} else {
		s.SpellContainerOnlySyncData = SpellContainerOnlySyncData{}
		s.Children = nil
		s.Difficulty.omit = false
	}
}

// SyncWithSource synchronizes this data with the source.
func (s *Spell) SyncWithSource() {
	syncFromSource(s, func(other *Spell) {
		s.SpellSyncData = other.SpellSyncData
		s.Tags = slices.Clone(other.Tags)
		s.Nameables = nameable.CloneDetails(other.Nameables)
		if s.Container() {
			s.SpellContainerOnlySyncData = other.SpellContainerOnlySyncData
		} else {
			s.SpellNonContainerOnlySyncData = other.SpellNonContainerOnlySyncData
			s.College = slices.Clone(other.College)
			s.Prereq = other.Prereq.CloneResolvingEmpty(false, true)
			s.Weapons = CloneWeapons(other.Weapons, s, Reference)
			s.Features = other.Features.Clone()
		}
	})
}

// Hash writes this object's contents into the hasher. Note that this only hashes the data that is considered to be
// "source" data, i.e. not expected to be modified by the user after copying from a library.
func (s *Spell) Hash(h hash.Hash) {
	xhash.Num8(h, syncKind(s))
	s.SpellSyncData.hash(h)
	if s.Container() {
		s.SpellContainerOnlySyncData.hash(h)
	} else {
		s.SpellNonContainerOnlySyncData.hash(h)
	}
}

func (s *SpellContainerOnlySyncData) hash(h hash.Hash) {
	s.TemplatePicker.Hash(h)
}

func (s *SpellNonContainerOnlySyncData) hash(h hash.Hash) {
	s.Difficulty.Hash(h)
	hashStrings(h, s.College)
	xhash.StringWithLen(h, s.PowerSource)
	xhash.StringWithLen(h, s.Class)
	xhash.StringWithLen(h, s.Resist)
	xhash.StringWithLen(h, s.CastingCost)
	xhash.StringWithLen(h, s.MaintenanceCost)
	xhash.StringWithLen(h, s.CastingTime)
	xhash.StringWithLen(h, s.Duration)
	xhash.StringWithLen(h, s.Item)
	xhash.StringWithLen(h, s.RitualSkillName)
	xhash.Num64(h, s.PrereqCount)
	s.Prereq.Hash(h)
	hashList(h, s.Weapons)
	hashList(h, s.Features)
}

// CopyFrom implements EditorData.
func (s *SpellEditData) CopyFrom(other *Spell) {
	s.copyFrom(other, &other.SpellEditData, other.Container(), false, Copy)
}

// NameableReplacements returns the replacements to be used with Nameables.
func (s *SpellEditData) NameableReplacements() map[string]string {
	return s.Replacements
}

// SetNameableReplacements sets the replacements to be used with Nameables.
func (s *SpellEditData) SetNameableReplacements(replacements map[string]string) {
	s.Replacements = replacements
}

// ApplyTo implements EditorData.
func (s *SpellEditData) ApplyTo(other *Spell) {
	other.copyFrom(other, s, other.Container(), true, Copy)
	slices.Sort(other.Tags)
	slices.Sort(other.College)
}

// copyFrom copies other into s. isApply distinguishes staging the editor's working copy from committing it back, and
// only affects how an empty Prereq list is resolved. mode controls how the weapons are cloned -- CopyFrom/ApplyTo above
// always use Copy, since that's staging or committing the same spell's own data; Clone passes its own mode through.
func (s *SpellEditData) copyFrom(spell *Spell, other *SpellEditData, isContainer, isApply bool, mode CloneMode) {
	*s = *other
	s.Tags = slices.Clone(other.Tags)
	s.Replacements = maps.Clone(other.Replacements)
	s.Nameables = nameable.CloneDetails(other.Nameables)
	s.TechLevel = clonePtr(other.TechLevel)
	s.College = slices.Clone(other.College)
	s.Prereq = s.Prereq.CloneResolvingEmpty(isContainer, isApply)
	s.Weapons = CloneWeapons(other.Weapons, spell, mode)
	s.Features = other.Features.Clone()
	s.Study = cloneStudyList(other.Study)
}
