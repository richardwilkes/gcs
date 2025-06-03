// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	"hash"
	"io/fs"
	"maps"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/affects"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/study"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/tmcost"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/tid"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
	"github.com/richardwilkes/unison/enums/align"
)

var (
	_ WeaponOwner            = &Trait{}
	_ Node[*Trait]           = &Trait{}
	_ TemplatePickerProvider = &Trait{}
	_ LeveledOwner           = &Trait{}
	_ EditorData[*Trait]     = &TraitEditData{}
)

// Columns that can be used with the trait method .CellData()
const (
	TraitDescriptionColumn = iota
	TraitPointsColumn
	TraitTagsColumn
	TraitReferenceColumn
	TraitLibSrcColumn
)

// Trait holds an advantage, disadvantage, quirk, or perk.
type Trait struct {
	TraitData
	owner             DataOwner
	UnsatisfiedReason string
}

// TraitData holds the Trait data that is written to disk.
type TraitData struct {
	SourcedID
	TraitEditData
	ThirdParty map[string]any `json:"third_party,omitempty"`
	Children   []*Trait       `json:"children,omitempty"` // Only for containers
	parent     *Trait
}

// TraitEditData holds the Trait data that can be edited by the UI detail editor.
type TraitEditData struct {
	TraitSyncData
	VTTNotes     string            `json:"vtt_notes,omitempty"`
	UserDesc     string            `json:"userdesc,omitempty"`
	Replacements map[string]string `json:"replacements,omitempty"`
	Modifiers    []*TraitModifier  `json:"modifiers,omitempty"`
	CR           selfctrl.Roll     `json:"cr,omitempty"`
	Disabled     bool              `json:"disabled,omitempty"`
	TraitNonContainerOnlyEditData
	TraitContainerSyncData
}

// TraitNonContainerOnlyEditData holds the Trait data that is only applicable to traits that aren't containers.
type TraitNonContainerOnlyEditData struct {
	TraitNonContainerSyncData
	Levels           fxp.Int     `json:"levels,omitempty"`
	Study            []*Study    `json:"study,omitempty"`
	StudyHoursNeeded study.Level `json:"study_hours_needed,omitempty"`
}

// TraitSyncData holds the Trait sync data that is common to both containers and non-containers.
type TraitSyncData struct {
	Name             string              `json:"name,omitempty"`
	PageRef          string              `json:"reference,omitempty"`
	PageRefHighlight string              `json:"reference_highlight,omitempty"`
	LocalNotes       string              `json:"local_notes,omitempty"`
	Tags             []string            `json:"tags,omitempty"`
	Prereq           *PrereqList         `json:"prereqs,omitempty"`
	CRAdj            selfctrl.Adjustment `json:"cr_adj,omitempty"`
}

// TraitNonContainerSyncData holds the Trait sync data that is only applicable to traits that aren't containers.
type TraitNonContainerSyncData struct {
	BasePoints     fxp.Int   `json:"base_points,omitempty"`
	PointsPerLevel fxp.Int   `json:"points_per_level,omitempty"`
	Weapons        []*Weapon `json:"weapons,omitempty"`
	Features       Features  `json:"features,omitempty"`
	RoundCostDown  bool      `json:"round_down,omitempty"`
	CanLevel       bool      `json:"can_level,omitempty"`
}

// TraitContainerSyncData holds the Trait sync data that is only applicable to traits that are containers.
type TraitContainerSyncData struct {
	Ancestry       string          `json:"ancestry,omitempty"`
	TemplatePicker *TemplatePicker `json:"template_picker,omitempty"`
	ContainerType  container.Type  `json:"container_type,omitempty"`
}

type traitListData struct {
	Version int      `json:"version"`
	Rows    []*Trait `json:"rows"`
}

// NewTraitsFromFile loads an Trait list from a file.
func NewTraitsFromFile(fileSystem fs.FS, filePath string) ([]*Trait, error) {
	var data traitListData
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &data); err != nil {
		return nil, errs.NewWithCause(InvalidFileData(), err)
	}
	if err := jio.CheckVersion(data.Version); err != nil {
		return nil, err
	}
	return data.Rows, nil
}

// SaveTraits writes the Trait list to the file as JSON.
func SaveTraits(traits []*Trait, filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, &traitListData{
		Version: jio.CurrentDataVersion,
		Rows:    traits,
	})
}

// NewTrait creates a new Trait.
func NewTrait(owner DataOwner, parent *Trait, isContainer bool) *Trait {
	var t Trait
	t.TID = tid.MustNewTID(traitKind(isContainer))
	t.parent = parent
	t.owner = owner
	t.Name = t.Kind()
	if t.Container() {
		t.TemplatePicker = &TemplatePicker{}
	}
	t.SetOpen(isContainer)
	return &t
}

func traitKind(isContainer bool) byte {
	if isContainer {
		return kinds.TraitContainer
	}
	return kinds.Trait
}

// ID returns the local ID of this data.
func (t *Trait) ID() tid.TID {
	return t.TID
}

// Container returns true if this is a container.
func (t *Trait) Container() bool {
	return tid.IsKind(t.TID, kinds.TraitContainer)
}

// HasChildren returns true if this node has children.
func (t *Trait) HasChildren() bool {
	return t.Container() && len(t.Children) > 0
}

// NodeChildren returns the children of this node, if any.
func (t *Trait) NodeChildren() []*Trait {
	return t.Children
}

// SetChildren sets the children of this node.
func (t *Trait) SetChildren(children []*Trait) {
	t.Children = children
}

// Parent returns the parent.
func (t *Trait) Parent() *Trait {
	return t.parent
}

// SetParent sets the parent.
func (t *Trait) SetParent(parent *Trait) {
	t.parent = parent
}

// IsOpen returns true if this node is currently open.
func (t *Trait) IsOpen() bool {
	return IsNodeOpen(t)
}

// SetOpen sets the current open state for this node.
func (t *Trait) SetOpen(open bool) {
	SetNodeOpen(t, open)
}

// Clone implements Node.
func (t *Trait) Clone(from LibraryFile, owner DataOwner, parent *Trait, preserveID bool) *Trait {
	other := NewTrait(owner, parent, t.Container())
	other.AdjustSource(from, t.SourcedID, preserveID)
	other.SetOpen(t.IsOpen())
	other.ThirdParty = t.ThirdParty
	other.CopyFrom(t)
	if t.HasChildren() {
		other.Children = make([]*Trait, 0, len(t.Children))
		for _, child := range t.Children {
			other.Children = append(other.Children, child.Clone(from, owner, other, preserveID))
		}
	}
	return other
}

// MarshalJSON implements json.Marshaler.
func (t *Trait) MarshalJSON() ([]byte, error) {
	type calc struct {
		Points            fxp.Int `json:"points"`
		UnsatisfiedReason string  `json:"unsatisfied_reason,omitempty"`
		ResolvedNotes     string  `json:"resolved_notes,omitempty"`
	}
	t.ClearUnusedFieldsForType()
	data := struct {
		TraitData
		Calc calc `json:"calc"`
	}{
		TraitData: t.TraitData,
		Calc: calc{
			Points:            t.AdjustedPoints(),
			UnsatisfiedReason: t.UnsatisfiedReason,
		},
	}
	notes := t.ResolveLocalNotes()
	if notes != t.LocalNotes {
		data.Calc.ResolvedNotes = notes
	}
	return json.Marshal(&data)
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Trait) UnmarshalJSON(data []byte) error {
	var localData struct {
		TraitData
		// Old data fields
		Type         string   `json:"type"`
		ExprNotes    string   `json:"notes"`
		Categories   []string `json:"categories"`
		Mental       bool     `json:"mental"`
		Physical     bool     `json:"physical"`
		Social       bool     `json:"social"`
		Exotic       bool     `json:"exotic"`
		Supernatural bool     `json:"supernatural"`
		IsOpen       bool     `json:"open"`
	}
	if err := json.Unmarshal(data, &localData); err != nil {
		return err
	}
	setOpen := false
	if !tid.IsValid(localData.TID) {
		// Fixup old data that used UUIDs instead of TIDs
		localData.TID = tid.MustNewTID(traitKind(strings.HasSuffix(localData.Type, containerKeyPostfix)))
		setOpen = localData.IsOpen
	}
	t.TraitData = localData.TraitData
	if t.LocalNotes == "" && localData.ExprNotes != "" {
		t.LocalNotes = EmbeddedExprToScript(localData.ExprNotes)
	}
	// Force the CanLevel flag, if needed
	if !t.Container() && (t.Levels != 0 || t.PointsPerLevel != 0) {
		t.CanLevel = true
	}
	t.ClearUnusedFieldsForType()
	t.transferOldTypeFlagToTags(i18n.Text("Mental"), localData.Mental)
	t.transferOldTypeFlagToTags(i18n.Text("Physical"), localData.Physical)
	t.transferOldTypeFlagToTags(i18n.Text("Social"), localData.Social)
	t.transferOldTypeFlagToTags(i18n.Text("Exotic"), localData.Exotic)
	t.transferOldTypeFlagToTags(i18n.Text("Supernatural"), localData.Supernatural)
	t.Tags = convertOldCategoriesToTags(t.Tags, localData.Categories)
	slices.Sort(t.Tags)
	if t.Container() {
		for _, one := range t.Children {
			one.parent = t
		}
	}
	if setOpen {
		SetNodeOpen(t, true)
	}
	return nil
}

func (t *Trait) transferOldTypeFlagToTags(name string, flag bool) {
	if flag && !slices.Contains(t.Tags, name) {
		t.Tags = append(t.Tags, name)
	}
}

// EffectivelyDisabled returns true if this node or a parent is disabled.
func (t *Trait) EffectivelyDisabled() bool {
	if t.Disabled {
		return true
	}
	p := t.Parent()
	for p != nil {
		if p.Disabled {
			return true
		}
		p = p.Parent()
	}
	return false
}

// TemplatePickerData returns the TemplatePicker data, if any.
func (t *Trait) TemplatePickerData() *TemplatePicker {
	return t.TemplatePicker
}

// TraitsHeaderData returns the header data information for the given trait column.
func TraitsHeaderData(columnID int) HeaderData {
	var data HeaderData
	switch columnID {
	case TraitDescriptionColumn:
		data.Title = i18n.Text("Trait")
		data.Primary = true
	case TraitPointsColumn:
		data.Title = i18n.Text("Pts")
		data.Detail = i18n.Text("Points")
		data.Less = fxp.IntLessFromString
	case TraitTagsColumn:
		data.Title = i18n.Text("Tags")
	case TraitReferenceColumn:
		data.Title = HeaderBookmark
		data.TitleIsImageKey = true
		data.Detail = PageRefTooltip()
	case TraitLibSrcColumn:
		data.Title = HeaderDatabase
		data.TitleIsImageKey = true
		data.Detail = LibSrcTooltip()
	}
	return data
}

// CellData returns the cell data information for the given column.
func (t *Trait) CellData(columnID int, data *CellData) {
	data.Dim = !t.Enabled()
	switch columnID {
	case TraitDescriptionColumn:
		data.Type = cell.Text
		data.Primary = t.String()
		data.Secondary = t.SecondaryText(func(option display.Option) bool { return option.Inline() })
		data.Disabled = t.EffectivelyDisabled()
		data.UnsatisfiedReason = t.UnsatisfiedReason
		data.Tooltip = t.SecondaryText(func(option display.Option) bool { return option.Tooltip() })
		data.TemplateInfo = t.TemplatePicker.Description()
		if t.Container() {
			switch t.ContainerType {
			case container.AlternativeAbilities:
				data.InlineTag = i18n.Text("Alternate")
			case container.Ancestry:
				data.InlineTag = i18n.Text("Ancestry")
			case container.Attributes:
				data.InlineTag = i18n.Text("Attribute")
			case container.MetaTrait:
				data.InlineTag = i18n.Text("Meta")
			default:
			}
		}
	case TraitPointsColumn:
		data.Type = cell.Text
		data.Primary = t.AdjustedPoints().String()
		data.Alignment = align.End
	case TraitTagsColumn:
		data.Type = cell.Tags
		data.Primary = CombineTags(t.Tags)
	case TraitReferenceColumn, PageRefCellAlias:
		data.Type = cell.PageRef
		data.Primary = t.PageRef
		if t.PageRefHighlight != "" {
			data.Secondary = t.PageRefHighlight
		} else {
			data.Secondary = t.NameWithReplacements()
		}
	case TraitLibSrcColumn:
		data.Type = cell.Text
		data.Alignment = align.Middle
		if !toolbox.IsNil(t.owner) {
			state, _ := t.owner.SourceMatcher().Match(t)
			data.Primary = state.AltString()
			data.Tooltip = state.String()
			if state != srcstate.Custom {
				data.Tooltip += "\n" + t.Source.String()
			}
		}
	}
}

// Depth returns the number of parents this node has.
func (t *Trait) Depth() int {
	count := 0
	p := t.parent
	for p != nil {
		count++
		p = p.parent
	}
	return count
}

// DataOwner returns the data owner.
func (t *Trait) DataOwner() DataOwner {
	return t.owner
}

// SetDataOwner sets the data owner and configures any sub-components as needed.
func (t *Trait) SetDataOwner(owner DataOwner) {
	t.owner = owner
	if t.Container() {
		for _, child := range t.Children {
			child.SetDataOwner(owner)
		}
	} else {
		for _, w := range t.Weapons {
			w.SetOwner(t)
		}
	}
	for _, m := range t.Modifiers {
		m.setTrait(t)
		m.SetDataOwner(owner)
	}
}

// IsLeveled returns true if the Trait is capable of having levels.
func (t *Trait) IsLeveled() bool {
	return t.CanLevel && !t.Container()
}

// CurrentLevel returns the current level of the trait or zero if it is not leveled.
func (t *Trait) CurrentLevel() fxp.Int {
	if t.Enabled() && t.IsLeveled() {
		return t.Levels
	}
	return 0
}

// AdjustedPoints returns the total points, taking levels and modifiers into account.
func (t *Trait) AdjustedPoints() fxp.Int {
	if t.EffectivelyDisabled() {
		return 0
	}
	if !t.Container() {
		return AdjustedPoints(EntityFromNode(t), t, t.CanLevel, t.BasePoints, t.Levels, t.PointsPerLevel, t.CR,
			t.AllModifiers(), t.RoundCostDown)
	}
	var points fxp.Int
	if t.ContainerType == container.AlternativeAbilities {
		values := make([]fxp.Int, len(t.Children))
		for i, one := range t.Children {
			values[i] = one.AdjustedPoints()
			if values[i] > points {
				points = values[i]
			}
		}
		maximum := points
		found := false
		for _, v := range values {
			if !found && maximum == v {
				found = true
			} else {
				points += fxp.ApplyRounding(calculateModifierPoints(v, fxp.Twenty), t.RoundCostDown)
			}
		}
	} else {
		for _, one := range t.Children {
			points += one.AdjustedPoints()
		}
	}
	return points
}

// AllModifiers returns the modifiers plus any inherited from parents.
func (t *Trait) AllModifiers() []*TraitModifier {
	all := make([]*TraitModifier, len(t.Modifiers))
	copy(all, t.Modifiers)
	p := t.parent
	for p != nil {
		all = append(all, p.Modifiers...)
		p = p.parent
	}
	return all
}

// Enabled returns true if this Trait and all of its parents are enabled.
func (t *Trait) Enabled() bool {
	if t.Disabled {
		return false
	}
	p := t.parent
	for p != nil {
		if p.Disabled {
			return false
		}
		p = p.parent
	}
	return true
}

// NameWithReplacements returns the name with any replacements applied.
func (t *Trait) NameWithReplacements() string {
	return nameable.Apply(t.Name, t.Replacements)
}

// LocalNotesWithReplacements returns the local notes with any replacements applied.
func (t *Trait) LocalNotesWithReplacements() string {
	return nameable.Apply(t.LocalNotes, t.Replacements)
}

// UserDescWithReplacements returns the user description with any replacements applied.
func (t *Trait) UserDescWithReplacements() string {
	return nameable.Apply(t.UserDesc, t.Replacements)
}

// Description returns a description, which doesn't include any levels.
func (t *Trait) Description() string {
	return t.NameWithReplacements()
}

// String implements fmt.Stringer.
func (t *Trait) String() string {
	var buffer strings.Builder
	buffer.WriteString(t.Description())
	if t.IsLeveled() {
		buffer.WriteByte(' ')
		buffer.WriteString(t.Levels.String())
	}
	return buffer.String()
}

// Notes returns the local notes.
func (t *Trait) Notes() string {
	return t.ResolveLocalNotes()
}

// ResolveLocalNotes resolves the local notes, running any embedded scripts to get the final result.
func (t *Trait) ResolveLocalNotes() string {
	return ResolveText(EntityFromNode(t), deferredNewScriptTrait(t), t.LocalNotesWithReplacements())
}

// FeatureList returns the list of Features.
func (t *Trait) FeatureList() Features {
	return t.Features
}

// TagList returns the list of tags.
func (t *Trait) TagList() []string {
	return t.Tags
}

// RatedStrength always return 0 for traits.
func (t *Trait) RatedStrength() fxp.Int {
	return 0
}

// NameableReplacements returns the replacements to be used with Nameables.
func (t *Trait) NameableReplacements() map[string]string {
	if t == nil {
		return nil
	}
	return t.Replacements
}

// FillWithNameableKeys adds any nameable keys found to the provided map.
func (t *Trait) FillWithNameableKeys(m, existing map[string]string) {
	if existing == nil {
		existing = t.Replacements
	}
	nameable.Extract(t.Name, m, existing)
	nameable.Extract(t.LocalNotes, m, existing)
	nameable.Extract(t.UserDesc, m, existing)
	if t.Prereq != nil {
		t.Prereq.FillWithNameableKeys(m, existing)
	}
	for _, one := range t.Features {
		one.FillWithNameableKeys(m, existing)
	}
	for _, one := range t.Weapons {
		one.FillWithNameableKeys(m, existing)
	}
	Traverse(func(mod *TraitModifier) bool {
		mod.FillWithNameableKeys(m, existing)
		return false
	}, true, true, t.Modifiers...)
}

// ApplyNameableKeys replaces any nameable keys found with the corresponding values in the provided map.
func (t *Trait) ApplyNameableKeys(m map[string]string) {
	needed := make(map[string]string)
	t.FillWithNameableKeys(needed, nil)
	t.Replacements = nameable.Reduce(needed, m)
}

// ActiveModifierFor returns the first modifier that matches the name (case-insensitive).
func (t *Trait) ActiveModifierFor(name string) *TraitModifier {
	var found *TraitModifier
	Traverse(func(mod *TraitModifier) bool {
		if strings.EqualFold(mod.NameWithReplacements(), name) {
			found = mod
			return true
		}
		return false
	}, true, true, t.Modifiers...)
	return found
}

// ModifierNotes returns the notes due to modifiers.
func (t *Trait) ModifierNotes() string {
	var buffer strings.Builder
	if t.CR != selfctrl.NoCR {
		buffer.WriteString(t.CR.String())
		if t.CRAdj != selfctrl.NoCRAdj {
			buffer.WriteString(", ")
			buffer.WriteString(t.CRAdj.Description(t.CR))
		}
	}
	Traverse(func(mod *TraitModifier) bool {
		if buffer.Len() != 0 {
			buffer.WriteString("; ")
		}
		buffer.WriteString(mod.FullDescription())
		return false
	}, true, true, t.Modifiers...)
	return buffer.String()
}

// SecondaryText returns the "secondary" text: the text display below an Trait.
func (t *Trait) SecondaryText(optionChecker func(display.Option) bool) string {
	var buffer strings.Builder
	settings := SheetSettingsFor(EntityFromNode(t))
	userDesc := t.UserDescWithReplacements()
	if userDesc != "" && optionChecker(settings.UserDescriptionDisplay) {
		buffer.WriteString(userDesc)
	}
	if optionChecker(settings.ModifiersDisplay) {
		AppendStringOntoNewLine(&buffer, t.ModifierNotes())
	}
	if optionChecker(settings.NotesDisplay) {
		AppendStringOntoNewLine(&buffer, strings.TrimSpace(t.Notes()))
		AppendStringOntoNewLine(&buffer, StudyHoursProgressText(ResolveStudyHours(t.Study), t.StudyHoursNeeded, false))
	}
	return buffer.String()
}

// HasTag returns true if 'tag' is present in 'tags'. This check both ignores case and can check for subsets that are
// colon-separated.
func HasTag(tag string, tags []string) bool {
	tag = strings.TrimSpace(tag)
	for _, one := range tags {
		for _, part := range strings.Split(one, ":") {
			if strings.EqualFold(tag, strings.TrimSpace(part)) {
				return true
			}
		}
	}
	return false
}

// CombineTags combines multiple tags into a single string.
func CombineTags(tags []string) string {
	return strings.Join(tags, ", ")
}

// ExtractTags from a combined tags string.
func ExtractTags(tags string) []string {
	var list []string
	for _, one := range strings.Split(tags, ",") {
		if one = strings.TrimSpace(one); one != "" {
			list = append(list, one)
		}
	}
	return list
}

// AdjustedPoints returns the total points, taking levels and modifiers into account. 'entity' and 'dataOwner' may be
// nil.
func AdjustedPoints(entity *Entity, trait *Trait, canLevel bool, basePoints, levels, pointsPerLevel fxp.Int, cr selfctrl.Roll, modifiers []*TraitModifier, roundCostDown bool) fxp.Int {
	if !canLevel {
		levels = 0
		pointsPerLevel = 0
	}
	var baseEnh, levelEnh, baseLim, levelLim fxp.Int
	multiplier := cr.Multiplier()
	Traverse(func(mod *TraitModifier) bool {
		mod.setTrait(trait)
		modifier := mod.CostModifier()
		switch mod.CostType {
		case tmcost.Percentage:
			switch mod.Affects {
			case affects.Total:
				if modifier < 0 {
					baseLim += modifier
					levelLim += modifier
				} else {
					baseEnh += modifier
					levelEnh += modifier
				}
			case affects.BaseOnly:
				if modifier < 0 {
					baseLim += modifier
				} else {
					baseEnh += modifier
				}
			case affects.LevelsOnly:
				if modifier < 0 {
					levelLim += modifier
				} else {
					levelEnh += modifier
				}
			}
		case tmcost.Points:
			if mod.Affects == affects.LevelsOnly {
				if canLevel {
					pointsPerLevel += modifier
				}
			} else {
				basePoints += modifier
			}
		case tmcost.Multiplier:
			multiplier = multiplier.Mul(modifier)
		}
		return false
	}, true, true, modifiers...)
	modifiedBasePoints := basePoints
	leveledPoints := pointsPerLevel.Mul(levels)
	if baseEnh != 0 || baseLim != 0 || levelEnh != 0 || levelLim != 0 {
		if SheetSettingsFor(entity).UseMultiplicativeModifiers {
			if baseEnh == levelEnh && baseLim == levelLim {
				modifiedBasePoints = modifyPoints(modifyPoints(modifiedBasePoints+leveledPoints, baseEnh), (-fxp.Eighty).Max(baseLim))
			} else {
				modifiedBasePoints = modifyPoints(modifyPoints(modifiedBasePoints, baseEnh), (-fxp.Eighty).Max(baseLim)) +
					modifyPoints(modifyPoints(leveledPoints, levelEnh), (-fxp.Eighty).Max(levelLim))
			}
		} else {
			baseMod := (-fxp.Eighty).Max(baseEnh + baseLim)
			levelMod := (-fxp.Eighty).Max(levelEnh + levelLim)
			if baseMod == levelMod {
				modifiedBasePoints = modifyPoints(modifiedBasePoints+leveledPoints, baseMod)
			} else {
				modifiedBasePoints = modifyPoints(modifiedBasePoints, baseMod) + modifyPoints(leveledPoints, levelMod)
			}
		}
	} else {
		modifiedBasePoints += leveledPoints
	}
	return fxp.ApplyRounding(modifiedBasePoints.Mul(multiplier), roundCostDown)
}

func modifyPoints(points, modifier fxp.Int) fxp.Int {
	return points + calculateModifierPoints(points, modifier)
}

func calculateModifierPoints(points, modifier fxp.Int) fxp.Int {
	return points.Mul(modifier).Div(fxp.Hundred)
}

// Kind returns the kind of data.
func (t *Trait) Kind() string {
	if t.Container() {
		return i18n.Text("Trait Container")
	}
	return i18n.Text("Trait")
}

// ClearUnusedFieldsForType zeroes out the fields that are not applicable to this type (container vs not-container).
func (t *Trait) ClearUnusedFieldsForType() {
	if t.Container() {
		t.TraitNonContainerOnlyEditData = TraitNonContainerOnlyEditData{}
		if t.TemplatePicker == nil {
			t.TemplatePicker = &TemplatePicker{}
		}
	} else {
		t.TraitContainerSyncData = TraitContainerSyncData{}
		t.Children = nil
		if !t.CanLevel {
			t.Levels = 0
			t.PointsPerLevel = 0
		}
	}
}

// GetSource returns the source of this data.
func (t *Trait) GetSource() Source {
	return t.Source
}

// ClearSource clears the source of this data.
func (t *Trait) ClearSource() {
	t.Source = Source{}
}

// SyncWithSource synchronizes this data with the source.
func (t *Trait) SyncWithSource() {
	if !toolbox.IsNil(t.owner) {
		if state, data := t.owner.SourceMatcher().Match(t); state == srcstate.Mismatched {
			if other, ok := data.(*Trait); ok {
				t.TraitSyncData = other.TraitSyncData
				t.Tags = slices.Clone(other.Tags)
				t.Prereq = other.Prereq.CloneResolvingEmpty(false, true)
				if t.Container() {
					t.TraitContainerSyncData = other.TraitContainerSyncData
					t.TemplatePicker = other.TemplatePicker.Clone()
				} else {
					t.TraitNonContainerSyncData = other.TraitNonContainerSyncData
					t.Weapons = CloneWeapons(other.Weapons, false)
					t.Features = other.Features.Clone()
				}
			}
		}
	}
}

// Hash writes this object's contents into the hasher. Note that this only hashes the data that is considered to be
// "source" data, i.e. not expected to be modified by the user after copying from a library.
func (t *Trait) Hash(h hash.Hash) {
	t.TraitSyncData.hash(h)
	if t.Container() {
		t.TraitContainerSyncData.hash(h)
	} else {
		t.TraitNonContainerSyncData.hash(h)
	}
}

func (t *TraitSyncData) hash(h hash.Hash) {
	hashhelper.String(h, t.Name)
	hashhelper.String(h, t.PageRef)
	hashhelper.String(h, t.PageRefHighlight)
	hashhelper.String(h, t.LocalNotes)
	hashhelper.Num64(h, len(t.Tags))
	for _, tag := range t.Tags {
		hashhelper.String(h, tag)
	}
	hashhelper.Num8(h, t.CRAdj)
	t.Prereq.Hash(h)
}

func (t *TraitNonContainerSyncData) hash(h hash.Hash) {
	hashhelper.Num64(h, t.BasePoints)
	hashhelper.Num64(h, t.PointsPerLevel)
	hashhelper.Num64(h, len(t.Weapons))
	for _, one := range t.Weapons {
		one.Hash(h)
	}
	hashhelper.Num64(h, len(t.Features))
	for _, one := range t.Features {
		one.Hash(h)
	}
	hashhelper.Bool(h, t.RoundCostDown)
	hashhelper.Bool(h, t.CanLevel)
}

func (t *TraitContainerSyncData) hash(h hash.Hash) {
	hashhelper.String(h, t.Ancestry)
	t.TemplatePicker.Hash(h)
	hashhelper.Num8(h, t.ContainerType)
}

// CopyFrom implements node.EditorData.
func (t *TraitEditData) CopyFrom(other *Trait) {
	t.copyFrom(other.owner, &other.TraitEditData, false)
}

// ApplyTo implements node.EditorData.
func (t *TraitEditData) ApplyTo(other *Trait) {
	other.copyFrom(other.owner, t, true)
}

func (t *TraitEditData) copyFrom(owner DataOwner, other *TraitEditData, isApply bool) {
	*t = *other
	t.Tags = txt.CloneStringSlice(other.Tags)
	t.Replacements = maps.Clone(other.Replacements)
	t.Modifiers = nil
	if len(other.Modifiers) != 0 {
		t.Modifiers = make([]*TraitModifier, 0, len(other.Modifiers))
		for _, one := range other.Modifiers {
			t.Modifiers = append(t.Modifiers, one.Clone(one.Source.LibraryFile, owner, nil, isApply))
		}
	}
	t.Prereq = t.Prereq.CloneResolvingEmpty(false, isApply)
	t.Weapons = CloneWeapons(other.Weapons, isApply)
	t.Features = other.Features.Clone()
	if len(other.Study) != 0 {
		t.Study = make([]*Study, len(other.Study))
		for i := range other.Study {
			t.Study[i] = other.Study[i].Clone()
		}
	}
	t.TemplatePicker = t.TemplatePicker.Clone()
}
