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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/affects"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emweight"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/frequency"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/maxusesmod"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selector"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/study"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/traitsel"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/unison/enums/align"
)

var (
	_ = assertNode[*Trait]
	_ = assertEditorData[*TraitEditData]

	_ WeaponOwner            = &Trait{}
	_ TemplatePickerProvider = &Trait{}
	_ LeveledOwner           = &Trait{}
	_ FeatureSwitcher        = &Trait{}

	_ TemplatePickerProvider = &TraitData{}
	_ TemplatePickerProvider = &TraitEditData{}
	_ TemplatePickerProvider = &TraitContainerSyncData{}
)

// Columns that can be used with the trait method .CellData()
const (
	TraitDescriptionColumn = iota
	TraitPointsColumn
	TraitTagsColumn
	TraitReferenceColumn
	TraitLibSrcColumn
	TraitSwitchColumn
)

// Trait holds an advantage, disadvantage, quirk, or perk.
type Trait struct {
	TraitData
	owner             DataOwner
	UnsatisfiedReason string
	resolvingLevel    bool
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
	VTTNotes     string            `json:"vtt_notes,omitzero"`
	UserDesc     string            `json:"userdesc,omitzero"`
	Replacements map[string]string `json:"replacements,omitempty"`
	Modifiers    []*TraitModifier  `json:"modifiers,omitempty"`
	SelfControl  selfctrl.Roll     `json:"cr,omitzero"`
	Frequency    frequency.Roll    `json:"frequency,omitzero"`
	Disabled     bool              `json:"disabled,omitzero"`
	ItemSwitch
	preconfigurable
	TraitNonContainerOnlyEditData
	TraitContainerSyncData
}

// TraitNonContainerOnlyEditData holds the Trait data that is only applicable to traits that aren't containers.
type TraitNonContainerOnlyEditData struct {
	TraitNonContainerSyncData
	Levels           fxp.Int     `json:"levels,omitzero"`
	Study            []*Study    `json:"study,omitempty"`
	StudyHoursNeeded study.Level `json:"study_hours_needed,omitzero"`
}

// TraitSyncData holds the Trait sync data that is common to both containers and non-containers.
type TraitSyncData struct {
	NodeSyncData
	Prereq         *PrereqList         `json:"prereqs,omitzero"`
	SelfControlAdj selfctrl.Adjustment `json:"cr_adj,omitzero"`
}

// TraitNonContainerSyncData holds the Trait sync data that is only applicable to traits that aren't containers.
type TraitNonContainerSyncData struct {
	BasePoints     fxp.Int   `json:"base_points,omitzero"`
	PointsPerLevel fxp.Int   `json:"points_per_level,omitzero"`
	MaxLevels      string    `json:"max_levels,omitzero"`
	Weapons        []*Weapon `json:"weapons,omitempty"`
	Features       Features  `json:"features,omitempty"`
	RoundCostDown  bool      `json:"round_down,omitzero"`
	CanLevel       bool      `json:"can_level,omitzero"`
}

// TraitContainerSyncData holds the Trait sync data that is only applicable to traits that are containers.
type TraitContainerSyncData struct {
	Ancestry         string         `json:"ancestry,omitzero"`
	TemplatePicker   TemplatePicker `json:"template_picker,omitzero"`
	ContainerType    container.Type `json:"container_type,omitzero"`
	AlternativeSlots int            `json:"alternative_slots,omitzero"`
}

// NewTraitsFromFile loads a Trait list from a file.
func NewTraitsFromFile(fileSystem fs.FS, filePath string) ([]*Trait, error) {
	return loadRows[*Trait](fileSystem, filePath)
}

// SaveTraits writes the Trait list to the file as JSON.
func SaveTraits(traits []*Trait, filePath string) error {
	return saveRows(filePath, traits)
}

// NewTrait creates a new Trait.
func NewTrait(owner DataOwner, parent *Trait, isContainer bool) *Trait {
	var t Trait
	t.TID = tid.MustNewTID(traitKind(isContainer))
	t.parent = parent
	t.owner = owner
	t.Name = t.Kind()
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
func (t *Trait) Clone(from LibraryFile, owner DataOwner, parent *Trait, mode CloneMode) *Trait {
	other := NewTrait(owner, parent, t.Container())
	other.AdjustSource(from, t.SourcedID, mode)
	other.SetOpen(t.IsOpen())
	other.ThirdParty = t.ThirdParty
	other.copyFrom(other, &t.TraitEditData, false, mode)
	PropagateNodeNoteClosedState(t, other)
	if t.HasChildren() {
		other.Children = make([]*Trait, 0, len(t.Children))
		for _, child := range t.Children {
			other.Children = append(other.Children, child.Clone(from, owner, other, mode))
		}
	}
	return other
}

// MarshalJSONTo implements json.MarshalerTo.
func (t *Trait) MarshalJSONTo(enc *jsontext.Encoder) error {
	type calc struct {
		Points            fxp.Int  `json:"points"`
		UnsatisfiedReason string   `json:"unsatisfied_reason,omitzero"`
		ResolvedNotes     string   `json:"resolved_notes,omitzero"`
		CurrentLevel      *fxp.Int `json:"current_level,omitzero"`
	}
	t.ClearUnusedFieldsForType()
	if omitCalc(enc) {
		return json.MarshalEncode(enc, &t.TraitData)
	}
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
	if t.IsLeveled() {
		level := t.CurrentLevel()
		data.Calc.CurrentLevel = &level
	}
	return json.MarshalEncode(enc, &data)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (t *Trait) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
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
	if err := json.UnmarshalDecode(dec, &localData); err != nil {
		return err
	}
	open := fixupLegacyTID(&localData.TID, localData.Type, traitKind) && localData.IsOpen
	t.TraitData = localData.TraitData
	t.Replacements = nameable.Normalize(t.Replacements)
	migrateLegacyText(&t.LocalNotes, localData.ExprNotes)
	// Force the CanLevel flag, if needed
	if !t.Container() {
		if t.Levels < 0 {
			t.Levels = 0
		}
		if t.Levels != 0 || t.PointsPerLevel != 0 {
			t.CanLevel = true
		}
	}
	t.ClearUnusedFieldsForType()
	t.transferOldTypeFlagToTags(i18n.Text("Mental"), localData.Mental)
	t.transferOldTypeFlagToTags(i18n.Text("Physical"), localData.Physical)
	t.transferOldTypeFlagToTags(i18n.Text("Social"), localData.Social)
	t.transferOldTypeFlagToTags(i18n.Text("Exotic"), localData.Exotic)
	t.transferOldTypeFlagToTags(i18n.Text("Supernatural"), localData.Supernatural)
	finishNodeUnmarshal(t, &t.Tags, localData.Categories, open)
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

// TemplatePickerData implements TemplatePickerProvider.
func (t *TraitContainerSyncData) TemplatePickerData() ([]picker.Type, *TemplatePicker) {
	return picker.TypesForTraits, &t.TemplatePicker
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
		data = tagsHeaderData()
	case TraitReferenceColumn:
		data = pageRefHeaderData()
	case TraitLibSrcColumn:
		data = libSrcHeaderData()
	case TraitSwitchColumn:
		data = switchHeaderData()
	}
	return data
}

// CellData returns the cell data information for the given column.
func (t *Trait) CellData(columnID int, data *CellData) {
	data.Self = t
	data.Dim = !t.Enabled()
	switch columnID {
	case TraitDescriptionColumn:
		data.Type = cell.Text
		var tooltip, overrideTooltip xbytes.InsertBuffer
		data.Primary = t.NameAndLevel(&tooltip)
		var buffer strings.Builder
		if resolvedSelfControl := t.ResolvedSelfControl(&overrideTooltip); resolvedSelfControl > selfctrl.None {
			buffer.WriteString(resolvedSelfControl.ShortString())
		}
		if resolvedFrequency := t.ResolvedFrequency(&overrideTooltip); resolvedFrequency > frequency.None {
			if buffer.Len() > 0 {
				buffer.WriteString(", ")
			}
			buffer.WriteString(resolvedFrequency.ShortString())
		}
		if buffer.Len() > 0 {
			data.Primary += " (" + buffer.String() + ")"
		}
		data.Secondary = t.SecondaryText(func(option display.Option) bool { return option.Inline() })
		data.Disabled = t.EffectivelyDisabled()
		data.UnsatisfiedReason = t.UnsatisfiedReason
		data.Tooltip = t.SecondaryText(func(option display.Option) bool { return option.Tooltip() })
		if tooltip.Len() != 0 {
			t := i18n.Text("Trait level adjustments:\n") + strings.ReplaceAll(tooltip.String(), "\n", "\n- ")
			if data.Tooltip == "" {
				data.Tooltip = t
			} else {
				data.Tooltip = t + "\n---\n" + data.Tooltip
			}
		}
		if overrideTooltip.Len() != 0 {
			t := i18n.Text("Overrides:\n") + strings.ReplaceAll(overrideTooltip.String(), "\n", "\n- ")
			if data.Tooltip == "" {
				data.Tooltip = t
			} else {
				data.Tooltip = t + "\n---\n" + data.Tooltip
			}
		}
		data.TemplateInfo = t.TemplatePicker.String()
		if t.Container() {
			switch t.ContainerType {
			case container.AlternativeAbilities:
				if slots := t.ResolvedAlternativeSlots(); slots > 1 {
					data.InlineTag = fmt.Sprintf(i18n.Text("Alternate x%d"), slots)
				} else {
					data.InlineTag = i18n.Text("Alternate")
				}
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
		fillTagsCell(data, t.Tags)
	case TraitReferenceColumn, PageRefCellAlias:
		fillPageRefCell(data, t.PageRef, t.PageRefHighlight, t.NameWithReplacements)
	case TraitLibSrcColumn:
		fillLibSrcCell(data, t.owner, t)
	case TraitSwitchColumn:
		// Only items that actually have something to switch get a cell; the rest are left blank.
		if t.HasSwitchableFeatures() {
			data.Type = cell.Switch
			data.Checked = t.SwitchedOn
			data.Alignment = align.Middle
			data.Tooltip = SwitchCellTooltip(t.Container())
			// A disabled trait -- or one inside a disabled container -- has already been dimmed above, which is
			// exactly what the switch cell wants: throwing the switch of a disabled trait changes nothing the user can
			// see, since every collection pass (features, weapons, reactions and conditional modifiers) skips traits
			// that aren't enabled, and the one thing the trait resolves for itself from its own switchable features,
			// ResolvedMaxLevels, is only consulted for enabled traits. The cell stays a live switch either way; the
			// dimming is purely visual.
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
	if t.Enabled() {
		return t.internalCurrentLevel(nil)
	}
	return 0
}

func (t *Trait) internalCurrentLevel(tooltip *xbytes.InsertBuffer) fxp.Int {
	if !t.IsLeveled() {
		return 0
	}
	if t.resolvingLevel {
		// A per-level trait bonus whose leveled owner is this trait -- either because the bonus it carries matches its
		// own name, or because two traits adjust each other's level -- would re-enter here while resolving the
		// adjustment. Fall back to the unadjusted level to break the cycle rather than recursing until the stack
		// overflows.
		return t.Levels.Max(0)
	}
	t.resolvingLevel = true
	defer func() { t.resolvingLevel = false }()
	var levelAdjustment fxp.Int
	if entity := EntityFromNode(t); entity != nil {
		levelAdjustment = entity.TraitBonusFor(t.NameWithReplacements(), t.Tags, tooltip)
	}
	return (t.Levels + levelAdjustment).Max(0)
}

// ResolvedMaxLevels returns the maximum level for this trait, resolving the MaxLevels expression (a plain number or an
// embedded script) and applying any matching TraitMaxLevelBonus features. A return value of zero means the trait has no
// maximum level. "This trait" bonuses attached to this trait or its enabled modifiers are always applied; "traits whose
// name" bonuses are gathered from the owning entity, if there is one. Bonuses only adjust a maximum the trait already
// declares -- a trait with no maximum stays unlimited no matter what matches it.
func (t *Trait) ResolvedMaxLevels() fxp.Int {
	if !t.IsLeveled() {
		return 0
	}
	base := fxp.Int(0)
	if strings.TrimSpace(t.MaxLevels) != "" {
		base = ResolveToNumber(EntityFromNode(t), deferredNewScriptTrait(t), t.MaxLevels)
	}
	addition := fxp.Int(0)
	percentage := fxp.Int(0)
	multiplier := fxp.One
	have := false
	apply := func(bonus *TraitMaxLevelBonus) {
		have = true
		amount := bonus.AdjustedAmount()
		switch bonus.Operation() {
		case maxusesmod.Percentage:
			percentage += amount
		case maxusesmod.Multiplier:
			if amount <= 0 {
				amount = fxp.One
			}
			multiplier = multiplier.Mul(amount)
		default: // maxusesmod.Addition
			addition += amount
		}
	}
	applyThisTrait := func(features Features, leveledOwner LeveledOwner) {
		for _, f := range features {
			if bonus, ok := f.(*TraitMaxLevelBonus); ok && bonus.SelectionType == traitsel.ThisTrait {
				// The level driving a per-level bonus comes from the node the bonus is attached to, matching how
				// Entity.processFeatures assigns the leveled owner for trait and trait modifier features.
				bonus.SetLeveledOwner(leveledOwner)
				apply(bonus)
			}
		}
	}
	applyThisTrait(t.ActiveFeatures(), t)
	Traverse(func(mod *TraitModifier) bool {
		applyThisTrait(mod.Features.Active(t.SwitchedOn), mod)
		return false
	}, true, true, t.Modifiers...)
	if entity := EntityFromNode(t); entity != nil {
		for _, bonus := range entity.TraitMaxLevelBonusesFor(t.NameWithReplacements(), t.Tags, nil) {
			apply(bonus)
		}
	}
	if !have || base <= 0 {
		// A trait with no declared maximum is unlimited. Bonuses adjust an existing cap, so one must never be allowed
		// to manufacture a cap from a base of zero -- that would turn a bonus meant to raise a limit into one that
		// imposes it.
		return base.Max(0)
	}
	result := base + addition
	result += result.Mul(percentage).Div(fxp.Hundred)
	result = result.Mul(multiplier)
	return result.Max(0)
}

// AdjustedPoints returns the total points, taking levels and modifiers into account.
func (t *Trait) AdjustedPoints() fxp.Int {
	if t.EffectivelyDisabled() {
		return 0
	}
	if !t.Container() {
		return AdjustedPoints(EntityFromNode(t), t, t.CanLevel, t.BasePoints, t.Levels, t.PointsPerLevel,
			t.SelfControl, t.Frequency, t.AllModifiers(), t.RoundCostDown)
	}
	var points fxp.Int
	if t.ContainerType == container.AlternativeAbilities {
		values := make([]fxp.Int, len(t.Children))
		for i, one := range t.Children {
			values[i] = one.AdjustedPoints()
		}
		slices.SortFunc(values, func(a, b fxp.Int) int { return cmp.Compare(b, a) })
		slots := min(t.ResolvedAlternativeSlots(), len(values))
		for i, v := range values {
			if i < slots {
				points += v
			} else {
				points += fxp.ApplyRounding(v.Mul(fxp.Twenty).Div(fxp.Hundred), t.RoundCostDown)
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

// ResolveSelector resolves a trait-scoped multi-state field, applying any SelectorOverride features that target the
// given field and match this trait. base is the field's intrinsic value. When more than one override matches, the
// winner is chosen by the override-resolution ladder (priority, then specificity), and, if tooltip is non-nil, the full
// contest is reported.
func (t *Trait) ResolveSelector(field selector.Field, base string, tooltip *xbytes.InsertBuffer) string {
	entity := EntityFromNode(t)
	if entity == nil {
		return base
	}
	var candidates []OverrideCandidate[string]
	for _, override := range entity.features.selectorOverrides {
		if override.Field == field && override.MatchesTrait(t) {
			candidates = append(candidates, OverrideCandidate[string]{Value: override.Value, Override: override})
		}
	}
	return ResolveOverride(base, candidates, func(s string) string { return s }, tooltip)
}

// ResolvedFrequency returns the trait's frequency of appearance after applying any matching selector override.
func (t *Trait) ResolvedFrequency(tooltip *xbytes.InsertBuffer) frequency.Roll {
	return frequencyFromStateKey(t.ResolveSelector(selector.TraitFrequency, frequencyStateKey(t.Frequency), tooltip))
}

// ResolvedSelfControl returns the trait's self-control roll after applying any matching selector override.
func (t *Trait) ResolvedSelfControl(tooltip *xbytes.InsertBuffer) selfctrl.Roll {
	return selfControlRollFromStateKey(t.ResolveSelector(selector.TraitSelfControlRoll,
		selfControlRollStateKey(t.SelfControl), tooltip))
}

// ResolvedSelfControlAdjustment returns the trait's self-control adjustment after applying any matching selector
// override.
func (t *Trait) ResolvedSelfControlAdjustment(tooltip *xbytes.InsertBuffer) selfctrl.Adjustment {
	return selfctrl.ExtractAdjustment(t.ResolveSelector(selector.TraitSelfControlAdjustment,
		t.SelfControlAdj.Key(), tooltip))
}

// LocalNotesWithReplacements returns the local notes with any replacements applied.
func (t *Trait) LocalNotesWithReplacements() string {
	return nameable.Apply(t.LocalNotes, t.Replacements)
}

// UserDescWithReplacements returns the user description with any replacements applied.
func (t *Trait) UserDescWithReplacements() string {
	return nameable.Apply(t.UserDesc, t.Replacements)
}

// String implements fmt.Stringer.
func (t *Trait) String() string {
	return t.NameAndLevel(nil)
}

// NameAndLevel returns the name and level of the trait.
func (t *Trait) NameAndLevel(tooltip *xbytes.InsertBuffer) string {
	var buffer strings.Builder
	buffer.WriteString(t.NameWithReplacements())
	if t.IsLeveled() {
		buffer.WriteByte(' ')
		buffer.WriteString(t.internalCurrentLevel(tooltip).String())
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

// ActiveFeatures returns the features of this trait that currently take effect, i.e. all of them except any switchable
// ones while the trait's switch is off. Features of the trait's modifiers are not included.
func (t *Trait) ActiveFeatures() Features {
	return t.Features.Active(t.SwitchedOn)
}

// HasSwitchableFeatures implements FeatureSwitcher.
func (t *Trait) HasSwitchableFeatures() bool {
	if !t.Container() && t.Features.AnySwitchable() {
		return true
	}
	return anyModifierSwitchable(t.Modifiers, func(mod *TraitModifier) Features { return mod.Features })
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
	nameable.Extract(
		m, existing,
		t.Name,
		t.LocalNotes,
		t.UserDesc,
	)
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
	}, true, false, t.Modifiers...)
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

// ModifierNotes returns the notes due to modifiers, including the self-control and frequency rolls, if any.
func (t *Trait) ModifierNotes() string {
	return t.modifierNotes(true, true)
}

// modifierNotes returns the notes due to modifiers. The self-control roll and frequency roll lines may be individually
// suppressed; this is intended for export templates that emit those rolls separately via Trait.SelfControl and
// Trait.Frequency.
func (t *Trait) modifierNotes(includeSelfControl, includeFrequency bool) string {
	var lines []string
	if resolvedSelfControl := t.ResolvedSelfControl(nil); includeSelfControl && resolvedSelfControl != selfctrl.None {
		resolvedAdjustment := t.ResolvedSelfControlAdjustment(nil)
		var buffer strings.Builder
		buffer.WriteString(i18n.Text("Self-Control Roll (CR): "))
		buffer.WriteString(resolvedSelfControl.String())
		if resolvedAdjustment != selfctrl.NoAdjustment {
			buffer.WriteString(", ")
			buffer.WriteString(resolvedAdjustment.Description(resolvedSelfControl))
		}
		lines = append(lines, buffer.String())
	}
	if resolvedFrequency := t.ResolvedFrequency(nil); includeFrequency && resolvedFrequency != frequency.None {
		lines = append(lines, fmt.Sprintf(i18n.Text("Frequency Roll (FR): %s"), resolvedFrequency))
	}
	var buffer strings.Builder
	Traverse(func(mod *TraitModifier) bool {
		if buffer.Len() != 0 {
			buffer.WriteString("; ")
		}
		buffer.WriteString(mod.FullDescription())
		return false
	}, true, true, t.Modifiers...)
	if buffer.Len() != 0 {
		lines = append(lines, buffer.String())
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n")
}

// SecondaryText returns the "secondary" text: the text display below an Trait.
func (t *Trait) SecondaryText(optionChecker func(display.Option) bool) string {
	var buffer strings.Builder
	settings := SheetSettingsFor(EntityFromNode(t))
	if optionChecker(settings.UserDescriptionDisplay) {
		if userDesc := t.UserDescWithReplacements(); userDesc != "" {
			buffer.WriteString(userDesc)
		}
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

// HasTag returns true if 'tag' is present in 'tags'. This check ignores case and matches either a whole tag or one of
// the colon-separated subsets within a tag.
func HasTag(tag string, tags []string) bool {
	tag = strings.TrimSpace(tag)
	for _, one := range tags {
		if strings.EqualFold(tag, strings.TrimSpace(one)) {
			return true
		}
		for part := range strings.SplitSeq(one, ":") {
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
	for one := range strings.SplitSeq(tags, ",") {
		if one = strings.TrimSpace(one); one != "" {
			list = append(list, one)
		}
	}
	return list
}

// AdjustedPoints returns the total points, taking levels and modifiers into account. 'entity' and 'trait' may be nil.
// 'trait' is only used to resolve the level of "use level from trait" modifiers; the modifiers themselves are left
// untouched, since the list may contain modifiers inherited from parent containers, which belong to those parents.
func AdjustedPoints(entity *Entity, trait *Trait, canLevel bool, basePoints, levels, pointsPerLevel fxp.Int, cr selfctrl.Roll, fr frequency.Roll, modifiers []*TraitModifier, roundCostDown bool) fxp.Int {
	if !canLevel {
		levels = 0
		pointsPerLevel = 0
	}
	baseLim := fxp.Fraction{Denominator: fxp.One}
	levelLim := fxp.Fraction{Denominator: fxp.One}
	baseEnh := fxp.Fraction{Denominator: fxp.One}
	levelEnh := fxp.Fraction{Denominator: fxp.One}
	multiplier := fxp.Fraction{
		Numerator:   cr.Multiplier().Mul(fr.Multiplier()),
		Denominator: fxp.One,
	}
	Traverse(func(mod *TraitModifier) bool {
		modifier := mod.CostModifierForTrait(trait)
		switch mod.CostModifierType() {
		case emweight.Addition:
			if mod.Affects == affects.LevelsOnly {
				if canLevel {
					pointsPerLevel += modifier.Value()
				}
			} else {
				basePoints += modifier.Value()
			}
		case emweight.PercentageAdder:
			switch mod.Affects {
			case affects.Total:
				if modifier.Numerator < 0 {
					baseLim = baseLim.Add(modifier)
					levelLim = levelLim.Add(modifier)
				} else {
					baseEnh = baseEnh.Add(modifier)
					levelEnh = levelEnh.Add(modifier)
				}
			case affects.BaseOnly:
				if modifier.Numerator < 0 {
					baseLim = baseLim.Add(modifier)
				} else {
					baseEnh = baseEnh.Add(modifier)
				}
			case affects.LevelsOnly:
				if modifier.Numerator < 0 {
					levelLim = levelLim.Add(modifier)
				} else {
					levelEnh = levelEnh.Add(modifier)
				}
			}
		case emweight.PercentageMultiplier:
			multiplier = multiplier.Mul(modifier).Div(fxp.Fraction{Numerator: fxp.Hundred, Denominator: fxp.One})
		case emweight.Multiplier:
			multiplier = multiplier.Mul(modifier)
		}
		return false
	}, true, true, modifiers...)
	modifiedBasePoints := fxp.Fraction{Numerator: basePoints, Denominator: fxp.One}
	leveledPoints := fxp.Fraction{Numerator: pointsPerLevel.Mul(levels), Denominator: fxp.One}
	if baseEnh.Numerator != 0 || baseLim.Numerator != 0 || levelEnh.Numerator != 0 || levelLim.Numerator != 0 {
		if SheetSettingsFor(entity).UseMultiplicativeModifiers {
			if baseEnh == levelEnh && baseLim == levelLim {
				if baseLim.Value() < -fxp.Eighty {
					baseLim.Numerator = -fxp.Eighty
					baseLim.Denominator = fxp.One
				}
				modifiedBasePoints = modifyPoints(modifyPoints(modifiedBasePoints.Add(leveledPoints), baseEnh), baseLim)
			} else {
				if baseLim.Value() < -fxp.Eighty {
					baseLim.Numerator = -fxp.Eighty
					baseLim.Denominator = fxp.One
				}
				modifiedBasePoints = modifyPoints(modifyPoints(modifiedBasePoints, baseEnh), baseLim)
				if levelLim.Value() < -fxp.Eighty {
					levelLim.Numerator = -fxp.Eighty
					levelLim.Denominator = fxp.One
				}
				leveledPts := modifyPoints(modifyPoints(leveledPoints, levelEnh), levelLim)
				modifiedBasePoints = modifiedBasePoints.Add(leveledPts)
			}
		} else {
			baseMod := baseEnh.Add(baseLim)
			if baseMod.Value() < -fxp.Eighty {
				baseMod.Numerator = -fxp.Eighty
				baseMod.Denominator = fxp.One
			}
			levelMod := levelEnh.Add(levelLim)
			if levelMod.Value() < -fxp.Eighty {
				levelMod.Numerator = -fxp.Eighty
				levelMod.Denominator = fxp.One
			}
			if baseMod == levelMod {
				modifiedBasePoints = modifyPoints(modifiedBasePoints.Add(leveledPoints), baseMod)
			} else {
				modifiedBasePoints = modifyPoints(modifiedBasePoints, baseMod)
				modifiedBasePoints = modifiedBasePoints.Add(modifyPoints(leveledPoints, levelMod))
			}
		}
	} else {
		modifiedBasePoints = modifiedBasePoints.Add(leveledPoints)
	}
	return fxp.ApplyRounding(modifiedBasePoints.Mul(multiplier).Value(), roundCostDown)
}

func modifyPoints(points, modifier fxp.Fraction) fxp.Fraction {
	return points.Add(points.Mul(modifier).Div(fxp.Fraction{Numerator: fxp.Hundred, Denominator: fxp.One}))
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
		if t.ContainerType != container.Ancestry {
			t.Ancestry = ""
		}
		if t.ContainerType != container.AlternativeAbilities {
			t.AlternativeSlots = 0
		} else {
			t.AlternativeSlots = max(t.AlternativeSlots, 1)
		}
	} else {
		t.TraitContainerSyncData = TraitContainerSyncData{}
		t.Children = nil
		if !t.CanLevel {
			t.Levels = 0
			t.PointsPerLevel = 0
			t.MaxLevels = ""
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
	syncFromSource(t, func(other *Trait) {
		t.TraitSyncData = other.TraitSyncData
		t.Tags = slices.Clone(other.Tags)
		t.Prereq = other.Prereq.CloneResolvingEmpty(false, true)
		if t.Container() {
			t.TraitContainerSyncData = other.TraitContainerSyncData
		} else {
			t.TraitNonContainerSyncData = other.TraitNonContainerSyncData
			t.Weapons = CloneWeapons(other.Weapons, t, Reference)
			t.Features = other.Features.Clone()
		}
	})
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
	t.NodeSyncData.hash(h)
	xhash.Num8(h, t.SelfControlAdj)
	t.Prereq.Hash(h)
}

func (t *TraitNonContainerSyncData) hash(h hash.Hash) {
	xhash.Num64(h, t.BasePoints)
	xhash.Num64(h, t.PointsPerLevel)
	xhash.StringWithLen(h, t.MaxLevels)
	hashList(h, t.Weapons)
	hashList(h, t.Features)
	xhash.Bool(h, t.RoundCostDown)
	xhash.Bool(h, t.CanLevel)
}

func (t *TraitContainerSyncData) hash(h hash.Hash) {
	t.TemplatePicker.Hash(h)
	xhash.Num8(h, t.ContainerType)
	switch t.ContainerType {
	case container.Ancestry:
		xhash.StringWithLen(h, t.Ancestry)
	case container.AlternativeAbilities:
		xhash.Num64(h, t.AlternativeSlots)
	}
}

// ResolvedAlternativeSlots returns the number of children that should be billed at full price when this is
// an Alternative Abilities container, i.e. the number of alternative abilities that can be active simultaneously.
func (t *TraitContainerSyncData) ResolvedAlternativeSlots() int {
	return max(t.AlternativeSlots, 1)
}

// CopyFrom implements node.EditorData.
func (t *TraitEditData) CopyFrom(other *Trait) {
	t.copyFrom(other, &other.TraitEditData, false, Copy)
}

// SetNameableReplacements sets the replacements to be used with Nameables.
func (t *TraitEditData) SetNameableReplacements(replacements map[string]string) {
	t.Replacements = replacements
}

// ApplyTo implements node.EditorData.
func (t *TraitEditData) ApplyTo(other *Trait) {
	other.copyFrom(other, t, true, Copy)
}

// copyFrom copies other into t. isApply distinguishes staging the editor's working copy from committing it
// back, and only affects how an empty Prereq list is resolved. mode controls how nested modifiers and
// weapons are cloned -- CopyFrom/ApplyTo above always use Copy, since that's staging or committing the same
// trait's own data, not producing a new one; Clone passes its own mode through.
func (t *TraitEditData) copyFrom(trait *Trait, other *TraitEditData, isApply bool, mode CloneMode) {
	*t = *other
	t.Tags = slices.Clone(other.Tags)
	t.Replacements = maps.Clone(other.Replacements)
	// Each copy is pointed at the trait it belongs to, so that a "use level from owner" modifier can resolve its level.
	// Prior to this, the copies held in an editor only acquired their trait as a side effect of a point cost
	// computation.
	t.Modifiers = cloneModifiers(other.Modifiers, trait, mode, func(m *TraitModifier) { m.setTrait(trait) })
	// setTrait() migrates a modifier's legacy replacements into the trait it was pointed at, which isn't the holder of
	// this data when an editor is being populated, so pick up anything it added. This is a no-op when this data is the
	// trait's own, since both maps are then the same one.
	t.Replacements = mergeReplacements(t.Replacements, trait.Replacements)
	t.Prereq = t.Prereq.CloneResolvingEmpty(false, isApply)
	t.Weapons = CloneWeapons(other.Weapons, trait, mode)
	t.Features = other.Features.Clone()
	if len(other.Study) != 0 {
		t.Study = make([]*Study, len(other.Study))
		for i := range other.Study {
			t.Study[i] = other.Study[i].Clone()
		}
	}
}

// CanPreconfigureContainer implements Preconfigurable.
func (t *TraitEditData) CanPreconfigureContainer() bool {
	return true
}
