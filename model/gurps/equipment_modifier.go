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
	"hash"
	"io/fs"
	"maps"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emcost"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emweight"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/unison/enums/align"
)

var (
	_ = assertNode[*EquipmentModifier]
	_ = assertModifierNode[*EquipmentModifier]
	_ = assertEditorData[*EquipmentModifierEditData]

	_ GeneralModifier = &EquipmentModifier{}
)

// Columns that can be used with the equipment modifier method .CellData()
const (
	EquipmentModifierEnabledColumn = iota
	EquipmentModifierDescriptionColumn
	EquipmentModifierTechLevelColumn
	EquipmentModifierCostColumn
	EquipmentModifierWeightColumn
	EquipmentModifierTagsColumn
	EquipmentModifierReferenceColumn
	EquipmentModifierLibSrcColumn
)

// EquipmentModifier holds a modifier to a piece of Equipment.
type EquipmentModifier struct {
	EquipmentModifierData
	owner     DataOwner
	equipment *Equipment
}

// EquipmentModifierData holds the EquipmentModifier data that is written to disk.
type EquipmentModifierData struct {
	SourcedID
	EquipmentModifierEditData
	ThirdParty map[string]any       `json:"third_party,omitempty"`
	Children   []*EquipmentModifier `json:"children,omitempty"` // Only for containers
	parent     *EquipmentModifier
}

// EquipmentModifierEditData holds the EquipmentModifier data that can be edited by the UI detail editor.
type EquipmentModifierEditData struct {
	EquipmentModifierSyncData
	VTTNotes     string            `json:"vtt_notes,omitzero"`
	Replacements map[string]string `json:"replacements,omitempty"` // Legacy; kept only to migrate old data
	EquipmentModifierEditDataNonContainerOnly
	ModifierContainerSyncData
}

// EquipmentModifierEditDataNonContainerOnly holds the EquipmentModifier data that is only applicable to
// EquipmentModifiers that aren't containers.
type EquipmentModifierEditDataNonContainerOnly struct {
	EquipmentModifierNonContainerSyncData
	Disabled bool `json:"disabled,omitzero"`
}

// EquipmentModifierSyncData holds the EquipmentModifier sync data that is common to both containers and non-containers.
type EquipmentModifierSyncData = NodeSyncData

// EquipmentModifierNonContainerSyncData holds the EquipmentModifier sync data that is only applicable to Equipment
// Modifiers that aren't containers.
type EquipmentModifierNonContainerSyncData struct {
	CostType          emcost.Type   `json:"cost_type,omitzero"`
	CostIsPerLevel    bool          `json:"cost_is_per_level,omitzero"`
	CostIsPerPound    bool          `json:"cost_is_per_pound,omitzero"`
	WeightType        emweight.Type `json:"weight_type,omitzero"`
	WeightIsPerLevel  bool          `json:"weight_is_per_level,omitzero"`
	ShowNotesOnWeapon bool          `json:"show_notes_on_weapon,omitzero"`
	// ShortName, when set, names the modifier in its owner's title and notes in place of its name.
	ShortName string `json:"short_name,omitzero"`
	// HideNotes keeps the modifier's notes out of its owner's notes.
	HideNotes bool `json:"hide_notes,omitzero"`
	// ShowInTitle moves the modifier from its owner's notes to the title notes after its owner's name. A modifier with
	// notes of its own still shows them in its owner's notes, unless HideNotes is set.
	ShowInTitle  bool     `json:"show_in_title,omitzero"`
	TechLevel    string   `json:"tech_level,omitzero"`
	CostAmount   string   `json:"cost,omitzero"`
	WeightAmount string   `json:"weight,omitzero"`
	Features     Features `json:"features,omitempty"`
}

// NewEquipmentModifiersFromFile loads an EquipmentModifier list from a file.
func NewEquipmentModifiersFromFile(fileSystem fs.FS, filePath string) ([]*EquipmentModifier, error) {
	return loadRows[*EquipmentModifier](fileSystem, filePath)
}

// SaveEquipmentModifiers writes the EquipmentModifier list to the file as JSON.
func SaveEquipmentModifiers(modifiers []*EquipmentModifier, filePath string) error {
	return saveRows(filePath, modifiers)
}

// NewEquipmentModifier creates an EquipmentModifier.
func NewEquipmentModifier(owner DataOwner, parent *EquipmentModifier, container bool) *EquipmentModifier {
	var e EquipmentModifier
	e.TID = tid.MustNewTID(equipmentModifierKind(container))
	e.Name = e.Kind()
	e.parent = parent
	e.owner = owner
	e.SetOpen(container)
	return &e
}

// NewEquipmentModifierChoice creates a new equipment modifier choice, which asks for exactly one of its options.
func NewEquipmentModifierChoice(owner DataOwner, parent *EquipmentModifier) *EquipmentModifier {
	e := NewEquipmentModifier(owner, parent, true)
	e.SetMandatoryChoice(true)
	e.Name = e.Kind()
	return e
}

func equipmentModifierKind(container bool) byte {
	if container {
		return kinds.EquipmentModifierContainer
	}
	return kinds.EquipmentModifier
}

// ID returns the local ID of this data.
func (e *EquipmentModifier) ID() tid.TID {
	return e.TID
}

// Container returns true if this is a container.
func (e *EquipmentModifier) Container() bool {
	return tid.IsKind(e.TID, kinds.EquipmentModifierContainer)
}

// HasChildren returns true if this node has children.
func (e *EquipmentModifier) HasChildren() bool {
	return e.Container() && len(e.Children) > 0
}

// NodeChildren returns the children of this node, if any.
func (e *EquipmentModifier) NodeChildren() []*EquipmentModifier {
	return e.Children
}

// SetChildren sets the children of this node.
func (e *EquipmentModifier) SetChildren(children []*EquipmentModifier) {
	e.Children = children
}

// Parent returns the parent.
func (e *EquipmentModifier) Parent() *EquipmentModifier {
	return e.parent
}

// SetParent sets the parent.
func (e *EquipmentModifier) SetParent(parent *EquipmentModifier) {
	e.parent = parent
}

// IsOpen returns true if this node is currently open.
func (e *EquipmentModifier) IsOpen() bool {
	return IsNodeOpen(e)
}

// SetOpen sets the current open state for this node.
func (e *EquipmentModifier) SetOpen(open bool) {
	SetNodeOpen(e, open)
}

// Clone implements Node.
func (e *EquipmentModifier) Clone(from LibraryFile, owner DataOwner, parent *EquipmentModifier, mode CloneMode) *EquipmentModifier {
	other := NewEquipmentModifier(owner, parent, e.Container())
	other.AdjustSource(from, e.SourcedID, mode)
	other.SetOpen(e.IsOpen())
	other.ThirdParty = e.ThirdParty
	other.CopyFrom(e)
	PropagateNodeNoteClosedState(e, other)
	if e.HasChildren() {
		other.Children = make([]*EquipmentModifier, 0, len(e.Children))
		for _, child := range e.Children {
			other.Children = append(other.Children, child.Clone(from, owner, other, mode))
		}
	}
	return other
}

// MarshalJSONTo implements json.MarshalerTo.
func (e *EquipmentModifier) MarshalJSONTo(enc *jsontext.Encoder) error {
	e.ClearUnusedFieldsForType()
	return marshalNodeData(enc, &e.EquipmentModifierData, func() *notesCalc {
		return newNotesCalc(e.ResolveLocalNotes(), e.LocalNotes)
	})
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (e *EquipmentModifier) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var localData struct {
		EquipmentModifierData
		// Old data fields
		Type       string   `json:"type"`
		ExprNotes  string   `json:"notes"`
		Categories []string `json:"categories"`
		IsOpen     bool     `json:"open"`
	}
	if err := json.UnmarshalDecode(dec, &localData); err != nil {
		return err
	}
	open := fixupLegacyTID(&localData.TID, localData.Type, equipmentModifierKind) && localData.IsOpen
	e.EquipmentModifierData = localData.EquipmentModifierData
	e.Replacements = nameable.Normalize(e.Replacements)
	migrateLegacyText(&e.LocalNotes, localData.ExprNotes)
	e.ClearUnusedFieldsForType()
	finishNodeUnmarshal(e, &e.Tags, localData.Categories, open)
	SettleModifierChoices(nil, e)
	return nil
}

// TagList returns the list of tags.
func (e *EquipmentModifier) TagList() []string {
	return e.Tags
}

// EquipmentModifierHeaderData returns the header data information for the given equipment modifier column.
func EquipmentModifierHeaderData(columnID int) HeaderData {
	var data HeaderData
	switch columnID {
	case EquipmentModifierEnabledColumn:
		data = enabledHeaderData()
	case EquipmentModifierDescriptionColumn:
		data.Title = i18n.Text("Equipment Modifier")
		data.Primary = true
	case EquipmentModifierTechLevelColumn:
		data = abbreviatedHeaderData(i18n.Text("TL"), i18n.Text("Tech Level"))
	case EquipmentModifierCostColumn:
		data.Title = i18n.Text("Cost Adjustment")
	case EquipmentModifierWeightColumn:
		data.Title = i18n.Text("Weight Adjustment")
	case EquipmentModifierTagsColumn:
		data = tagsHeaderData()
	case EquipmentModifierReferenceColumn:
		data = pageRefHeaderData()
	case EquipmentModifierLibSrcColumn:
		data = libSrcHeaderData()
	}
	return data
}

// CellData returns the cell data information for the given column.
func (e *EquipmentModifier) CellData(columnID int, data *CellData) {
	data.Self = e
	switch columnID {
	case EquipmentModifierEnabledColumn:
		if !e.Container() {
			data.Type = cell.Toggle
			data.Name = i18n.Text("Enabled")
			data.Checked = e.Enabled()
			data.Alignment = align.Middle
		}
	case EquipmentModifierDescriptionColumn:
		data.Type = cell.Text
		data.Primary = e.NameWithReplacements()
		data.Secondary = e.SecondaryText(func(option display.Option) bool { return option.Inline() })
		data.Tooltip = e.SecondaryText(func(option display.Option) bool { return option.Tooltip() })
		fillModifierChoiceCell(e, data)
	case EquipmentModifierTechLevelColumn:
		if !e.Container() {
			data.Type = cell.Text
			data.Primary = e.TechLevel
		}
	case EquipmentModifierCostColumn:
		if !e.Container() {
			data.Type = cell.Text
			data.Primary = e.CostDescription()
		}
	case EquipmentModifierWeightColumn:
		if !e.Container() {
			data.Type = cell.Text
			data.Primary = e.WeightDescription()
		}
	case EquipmentModifierTagsColumn:
		fillTagsCell(data, e.Tags)
	case EquipmentModifierReferenceColumn, PageRefCellAlias:
		fillPageRefCell(data, e.PageRef, e.PageRefHighlight, e.NameWithReplacements)
	case EquipmentModifierLibSrcColumn:
		fillLibSrcCell(data, e.owner, e)
	}
}

// Depth returns the number of parents this node has.
func (e *EquipmentModifier) Depth() int {
	count := 0
	p := e.parent
	for p != nil {
		count++
		p = p.parent
	}
	return count
}

// Target returns the equipment being targeted for modification
func (e *EquipmentModifier) Target() *Equipment {
	return e.equipment
}

// DataOwner returns the data owner.
func (e *EquipmentModifier) DataOwner() DataOwner {
	return e.owner
}

// SetTarget sets the equipment being targeted for modification and configures any sub-components as needed
func (e *EquipmentModifier) SetTarget(target *Equipment) *EquipmentModifier {
	e.equipment = target

	// COMPAT: Promote replacements from this node up to the target node
	if target != nil && len(e.Replacements) != 0 {
		target.Replacements = mergeReplacements(target.Replacements, e.Replacements)
		e.Replacements = nil
	}

	if e.Container() {
		for _, child := range e.Children {
			child.SetTarget(target)
		}
	}

	return e
}

// SetDataOwner sets the data owner and configures any sub-components as needed.
func (e *EquipmentModifier) SetDataOwner(owner DataOwner) {
	e.owner = owner
	if e.Container() {
		for _, child := range e.Children {
			child.SetDataOwner(owner)
		}
	}
}

func (e *EquipmentModifier) String() string {
	return e.NameWithReplacements()
}

// CompactName returns how the modifier is named in its owner's title and notes: its short name, or its name when it has
// none.
func (e *EquipmentModifier) CompactName() string {
	return cmp.Or(e.ShortNameWithReplacements(), e.NameWithReplacements())
}

// ShortNameWithReplacements returns the short name with any replacements applied.
func (e *EquipmentModifier) ShortNameWithReplacements() string {
	return applyOwnerReplacements(e.ShortName, e.equipment)
}

// ShowsInTitle returns true if the modifier is shown in its owner's title notes.
func (e *EquipmentModifier) ShowsInTitle() bool {
	return !e.Container() && e.ShowInTitle
}

// ResolveLocalNotes resolves the local notes, running any embedded scripts to get the final result.
func (e *EquipmentModifier) ResolveLocalNotes() string {
	return ResolveText(EntityFromNode(e), deferredNewScriptEquipmentModifier(e), e.LocalNotesWithReplacements())
}

// ShowsNotesOnWeapon returns true if the modifier's notes are to be shown on the weapons of the equipment it belongs
// to.
func (e *EquipmentModifier) ShowsNotesOnWeapon() bool {
	return e.ShowNotesOnWeapon
}

// SecondaryText returns the "secondary" text: the text displayed below the modifier.
func (e *EquipmentModifier) SecondaryText(optionChecker func(display.Option) bool) string {
	return modifierSecondaryText(e, optionChecker)
}

// FullDescription returns a full description.
func (e *EquipmentModifier) FullDescription() string {
	return e.describe(e.String(), true)
}

// NotesDescription returns the description shown in its owner's notes, which names it by its CompactName and leaves out
// its notes when HideNotes is set.
func (e *EquipmentModifier) NotesDescription() string {
	return e.describe(e.CompactName(), !e.HideNotes)
}

// ShowsInNotes returns true if the modifier appears in its owner's notes: always, unless it is shown in the title, in
// which case only when it has notes to show.
func (e *EquipmentModifier) ShowsInNotes() bool {
	return !e.ShowsInTitle() || (!e.HideNotes && e.ResolveLocalNotes() != "")
}

func (e *EquipmentModifier) describe(name string, withNotes bool) string {
	var buffer strings.Builder
	buffer.WriteString(name)
	if localNotes := e.ResolveLocalNotes(); withNotes && localNotes != "" {
		buffer.WriteString(" (")
		buffer.WriteString(localNotes)
		buffer.WriteByte(')')
	}
	if SheetSettingsFor(EntityFromNode(e)).ShowEquipmentModifierAdj {
		costDesc := e.CostDescription()
		weightDesc := e.WeightDescription()
		if costDesc != "" || weightDesc != "" {
			buffer.WriteString(" [")
			buffer.WriteString(costDesc)
			if weightDesc != "" {
				if costDesc != "" {
					buffer.WriteString("; ")
				}
				buffer.WriteString(weightDesc)
			}
			buffer.WriteByte(']')
		}
	}
	return buffer.String()
}

// FullCostDescription returns a combination of the cost and weight descriptions.
func (e *EquipmentModifier) FullCostDescription() string {
	cost := e.CostDescription()
	weight := e.WeightDescription()
	switch {
	case cost == "" && weight == "":
		return ""
	case cost == "":
		return weight
	case weight == "":
		return cost
	default:
		return cost + "; " + weight
	}
}

// CostDescription returns the formatted cost.
func (e *EquipmentModifier) CostDescription() string {
	if e.Container() || (e.CostType == emcost.Original && (e.CostAmount == "" || e.CostAmount == "+0")) {
		return ""
	}
	var buffer strings.Builder
	buffer.WriteString(e.CostType.Format(e.CostAmount))
	if e.CostIsPerLevel {
		buffer.WriteString(i18n.Text(" per level"))
	}
	if e.CostIsPerPound {
		buffer.WriteString(i18n.Text(" per pound"))
	}
	buffer.WriteByte(' ')
	buffer.WriteString(e.CostType.String())
	return buffer.String()
}

// WeightDescription returns the formatted weight.
func (e *EquipmentModifier) WeightDescription() string {
	if e.Container() || (e.WeightType == emweight.Original &&
		(e.WeightAmount == "" || strings.HasPrefix(e.WeightAmount, "+0 "))) {
		return ""
	}
	var buffer strings.Builder
	buffer.WriteString(e.WeightType.Format(e.WeightAmount, SheetSettingsFor(EntityFromNode(e)).DefaultWeightUnits))
	if e.WeightIsPerLevel {
		buffer.WriteString(i18n.Text(" per level"))
	}
	buffer.WriteByte(' ')
	buffer.WriteString(e.WeightType.String())
	return buffer.String()
}

// NameableReplacements returns the replacements to be used with Nameables.
func (e *EquipmentModifier) NameableReplacements() map[string]string {
	if e == nil || e.equipment == nil {
		return nil
	}
	return e.equipment.Replacements
}

// NameWithReplacements returns the name with any replacements applied.
func (e *EquipmentModifier) NameWithReplacements() string {
	return applyOwnerReplacements(e.Name, e.equipment)
}

// LocalNotesWithReplacements returns the local notes with any replacements applied.
func (e *EquipmentModifier) LocalNotesWithReplacements() string {
	return applyOwnerReplacements(e.LocalNotes, e.equipment)
}

// FillWithNameableKeys adds any nameable keys found in this EquipmentModifier to the provided map, provided it is
// enabled. Containers take part too, since their name and notes are displayed with replacements applied just as a leaf
// modifier's are.
func (e *EquipmentModifier) FillWithNameableKeys(m, existing map[string]string) {
	if e.Enabled() {
		e.fillWithNameableKeysEvenIfDisabled(m, existing)
	}
}

// fillWithNameableKeysEvenIfDisabled implements Modifier.
func (e *EquipmentModifier) fillWithNameableKeysEvenIfDisabled(m, existing map[string]string) {
	if existing == nil {
		existing = e.NameableReplacements()
	}
	nameable.Extract(
		m, existing,
		e.Name,
		e.ShortName,
		e.LocalNotes,
	)
	for _, one := range e.Features {
		one.FillWithNameableKeys(m, existing)
	}
}

// ApplyNameableKeys merges the values for this modifier's keys into the owning equipment's replacements, which is
// where they are kept (see modifierNameableReplacements).
func (e *EquipmentModifier) ApplyNameableKeys(m map[string]string) {
	if e.equipment != nil {
		e.equipment.Replacements = modifierNameableReplacements(e.equipment.Replacements, e, m)
	}
}

// enabledVariant implements Modifier.
func (e *EquipmentModifier) enabledVariant() *EquipmentModifier { //nolint:unused // Only called through the Modifier constraint
	if e.Enabled() {
		return e
	}
	variant := *e
	variant.Disabled = false
	return &variant
}

// Enabled returns true if this node is enabled.
func (e *EquipmentModifier) Enabled() bool {
	return !e.Disabled || e.Container()
}

// SetEnabled makes the node enabled, if possible.
func (e *EquipmentModifier) SetEnabled(enabled bool) {
	if !e.Container() {
		e.Disabled = !enabled
	}
}

// CostMultiplier returns the amount to multiply the cost by. A per-pound cost is multiplied by weight, what the
// equipment weighs with the modifiers being costed, or by its base weight if that is greater, rounded up to a whole
// pound and never less than one.
func (e *EquipmentModifier) CostMultiplier(weight fxp.Int) fxp.Int {
	multiplier := multiplierForEquipmentModifier(e.equipment, e.CostIsPerLevel)
	if e.CostIsPerPound {
		baseWeight := fxp.Int(e.equipment.ResolvedBaseWeight())
		multiplier = multiplier.Mul(max(weight, baseWeight).Ceil().Max(fxp.One))
	}
	return multiplier
}

// WeightMultiplier returns the amount to multiply the weight by.
func (e *EquipmentModifier) WeightMultiplier() fxp.Int {
	return multiplierForEquipmentModifier(e.equipment, e.WeightIsPerLevel)
}

func multiplierForEquipmentModifier(equipment *Equipment, isPerLevel bool) fxp.Int {
	var multiplier fxp.Int
	if isPerLevel && equipment != nil && equipment.IsLeveled() {
		multiplier = equipment.CurrentLevel()
	}
	if multiplier <= 0 {
		multiplier = fxp.One
	}
	return multiplier
}

// ValueAdjustedForModifiers returns the value after adjusting it for a set of modifiers, weight being what the equipment
// weighs with them (see CostMultiplier).
func ValueAdjustedForModifiers(equipment *Equipment, value, weight fxp.Int, modifiers []*EquipmentModifier) fxp.Int {
	cost := processNonCFStep(equipment, emcost.Original, value, weight, modifiers)

	var cf fxp.Int
	Traverse(func(mod *EquipmentModifier) bool {
		mod.equipment = equipment
		if mod.CostType == emcost.Base {
			t := emcost.Base.FromString(mod.CostAmount)
			cf += t.ExtractValue(mod.CostAmount).Mul(mod.CostMultiplier(weight))
			if t == emcost.Multiplier {
				cf -= fxp.One
			}
		}
		return false
	}, true, true, modifiers...)
	if cf != 0 {
		cost = cost.Mul(cf.Max(fxp.NegPointEight) + fxp.One)
	}

	cost = processNonCFStep(equipment, emcost.FinalBase, cost, weight, modifiers)

	cost = processNonCFStep(equipment, emcost.Final, cost, weight, modifiers)

	return cost.Max(0)
}

func processNonCFStep(equipment *Equipment, costType emcost.Type, value, weight fxp.Int, modifiers []*EquipmentModifier) fxp.Int {
	var percentages, additions fxp.Int
	cost := value
	Traverse(func(mod *EquipmentModifier) bool {
		mod.equipment = equipment
		if mod.CostType == costType {
			t := costType.FromString(mod.CostAmount)
			amt := t.ExtractValue(mod.CostAmount).Mul(mod.CostMultiplier(weight))
			switch t {
			case emcost.Addition:
				additions += amt
			case emcost.Percentage:
				percentages += amt
			case emcost.Multiplier:
				cost = cost.Mul(amt)
			}
		}
		return false
	}, true, true, modifiers...)
	cost += additions
	if percentages != 0 {
		cost += value.Mul(percentages.Div(fxp.Hundred))
	}
	return cost
}

// WeightAdjustedForModifiers returns the weight after adjusting it for a set of modifiers.
func WeightAdjustedForModifiers(equipment *Equipment, weight fxp.Weight, modifiers []*EquipmentModifier, defUnits fxp.WeightUnit) fxp.Weight {
	var percentages fxp.Int
	w := fxp.Int(weight)

	Traverse(func(mod *EquipmentModifier) bool {
		mod.equipment = equipment
		if mod.WeightType == emweight.Original {
			t := emweight.Original.FromString(mod.WeightAmount)
			f := t.ExtractFraction(mod.WeightAmount)
			f.Normalize()
			f.Numerator = f.Numerator.Mul(mod.WeightMultiplier())
			amt := f.Value()
			if t == emweight.Addition {
				w += fxp.TrailingWeightUnitFromString(mod.WeightAmount, defUnits).ToPounds(amt)
			} else {
				percentages += amt
			}
		}
		return false
	}, true, true, modifiers...)
	if percentages != 0 {
		w += fxp.Int(weight).Mul(percentages.Div(fxp.Hundred))
	}

	w = processMultiplyAddWeightStep(equipment, emweight.Base, w, defUnits, modifiers)

	w = processMultiplyAddWeightStep(equipment, emweight.FinalBase, w, defUnits, modifiers)

	w = processMultiplyAddWeightStep(equipment, emweight.Final, w, defUnits, modifiers)

	return fxp.Weight(w.Max(0))
}

func processMultiplyAddWeightStep(equipment *Equipment, weightType emweight.Type, weight fxp.Int, defUnits fxp.WeightUnit, modifiers []*EquipmentModifier) fxp.Int {
	var sum fxp.Int
	Traverse(func(mod *EquipmentModifier) bool {
		mod.equipment = equipment
		if mod.WeightType == weightType {
			t := weightType.FromString(mod.WeightAmount)
			f := t.ExtractFraction(mod.WeightAmount)
			f.Normalize()
			f.Numerator = f.Numerator.Mul(mod.WeightMultiplier())
			switch t {
			case emweight.Addition:
				sum += fxp.TrailingWeightUnitFromString(mod.WeightAmount, defUnits).ToPounds(f.Value())
			case emweight.PercentageMultiplier:
				weight = weight.Mul(f.Numerator).Div(f.Denominator.Mul(fxp.Hundred))
			case emweight.Multiplier:
				weight = weight.Mul(f.Numerator).Div(f.Denominator)
			}
		}
		return false
	}, true, true, modifiers...)
	return weight + sum
}

// Kind returns the kind of data.
func (e *EquipmentModifier) Kind() string {
	if e.Container() {
		if e.IsChoice() {
			return i18n.Text("Equipment Modifier Choice")
		}
		return i18n.Text("Equipment Modifier Group")
	}
	return i18n.Text("Equipment Modifier")
}

// ClearUnusedFieldsForType zeroes out the fields that are not applicable to this type (container vs not-container).
func (e *EquipmentModifier) ClearUnusedFieldsForType() {
	if e.Container() {
		e.EquipmentModifierEditDataNonContainerOnly = EquipmentModifierEditDataNonContainerOnly{}
		e.VTTNotes = ""
		e.normalizeModifierChoice()
	} else {
		e.ModifierContainerSyncData = ModifierContainerSyncData{}
		e.Children = nil
	}
}

// SyncWithSource synchronizes this data with the source.
func (e *EquipmentModifier) SyncWithSource() {
	syncFromSource(e, func(other *EquipmentModifier) {
		e.EquipmentModifierSyncData = other.EquipmentModifierSyncData
		e.Tags = slices.Clone(other.Tags)
		if e.Container() {
			e.ModifierContainerSyncData = other.ModifierContainerSyncData
			SettleModifierChoicesAround(e)
		} else {
			e.EquipmentModifierNonContainerSyncData = other.EquipmentModifierNonContainerSyncData
			e.Features = other.Features.Clone()
		}
	})
}

// Hash writes this object's contents into the hasher. Note that this only hashes the data that is considered to be
// "source" data, i.e. not expected to be modified by the user after copying from a library.
func (e *EquipmentModifier) Hash(h hash.Hash) {
	e.EquipmentModifierSyncData.hash(h)
	if e.Container() {
		e.ModifierContainerSyncData.hash(h)
	} else {
		e.EquipmentModifierNonContainerSyncData.hash(h)
	}
}

func (e *EquipmentModifierNonContainerSyncData) hash(h hash.Hash) {
	xhash.Num8(h, e.CostType)
	xhash.Bool(h, e.CostIsPerLevel)
	xhash.Bool(h, e.CostIsPerPound)
	xhash.Num8(h, e.WeightType)
	xhash.Bool(h, e.WeightIsPerLevel)
	xhash.Bool(h, e.ShowNotesOnWeapon)
	xhash.StringWithLen(h, e.ShortName)
	xhash.Bool(h, e.ShowInTitle)
	xhash.Bool(h, e.HideNotes)
	xhash.StringWithLen(h, e.TechLevel)
	xhash.StringWithLen(h, e.CostAmount)
	xhash.StringWithLen(h, e.WeightAmount)
	hashList(h, e.Features)
}

// CopyFrom implements EditorData.
func (e *EquipmentModifierEditData) CopyFrom(other *EquipmentModifier) {
	e.copyFrom(&other.EquipmentModifierEditData)
}

// ApplyTo implements EditorData.
func (e *EquipmentModifierEditData) ApplyTo(other *EquipmentModifier) {
	other.copyFrom(e)
}

func (e *EquipmentModifierEditData) copyFrom(other *EquipmentModifierEditData) {
	*e = *other
	e.Tags = slices.Clone(other.Tags)
	e.Replacements = maps.Clone(other.Replacements)
	e.Features = other.Features.Clone()
}
