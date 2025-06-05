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
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emcost"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emweight"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
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
	_ Node[*EquipmentModifier]       = &EquipmentModifier{}
	_ GeneralModifier                = &EquipmentModifier{}
	_ EditorData[*EquipmentModifier] = &EquipmentModifierEditData{}
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
	VTTNotes     string            `json:"vtt_notes,omitempty"`
	Replacements map[string]string `json:"replacements,omitempty"` // Not actually used any longer, but kept so that we can migrate old data
	EquipmentModifierEditDataNonContainerOnly
}

// EquipmentModifierEditDataNonContainerOnly holds the EquipmentModifier data that is only applicable to
// EquipmentModifiers that aren't containers.
type EquipmentModifierEditDataNonContainerOnly struct {
	EquipmentModifierNonContainerSyncData
	Disabled bool `json:"disabled,omitempty"`
}

// EquipmentModifierSyncData holds the EquipmentModifier sync data that is common to both containers and non-containers.
type EquipmentModifierSyncData struct {
	Name             string   `json:"name,omitempty"`
	PageRef          string   `json:"reference,omitempty"`
	PageRefHighlight string   `json:"reference_highlight,omitempty"`
	LocalNotes       string   `json:"local_notes,omitempty"`
	Tags             []string `json:"tags,omitempty"`
}

// EquipmentModifierNonContainerSyncData holds the EquipmentModifier sync data that is only applicable to Equipment
// Modifiers that aren't containers.
type EquipmentModifierNonContainerSyncData struct {
	CostType          emcost.Type   `json:"cost_type,omitempty"`
	CostIsPerLevel    bool          `json:"cost_is_per_level,omitempty"`
	WeightType        emweight.Type `json:"weight_type,omitempty"`
	WeightIsPerLevel  bool          `json:"weight_is_per_level,omitempty"`
	ShowNotesOnWeapon bool          `json:"show_notes_on_weapon,omitempty"`
	TechLevel         string        `json:"tech_level,omitempty"`
	CostAmount        string        `json:"cost,omitempty"`
	WeightAmount      string        `json:"weight,omitempty"`
	Features          Features      `json:"features,omitempty"`
}

type equipmentModifierListData struct {
	Version int                  `json:"version"`
	Rows    []*EquipmentModifier `json:"rows"`
}

// NewEquipmentModifiersFromFile loads an EquipmentModifier list from a file.
func NewEquipmentModifiersFromFile(fileSystem fs.FS, filePath string) ([]*EquipmentModifier, error) {
	var data equipmentModifierListData
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &data); err != nil {
		return nil, errs.NewWithCause(InvalidFileData(), err)
	}
	if err := jio.CheckVersion(data.Version); err != nil {
		return nil, err
	}
	return data.Rows, nil
}

// SaveEquipmentModifiers writes the EquipmentModifier list to the file as JSON.
func SaveEquipmentModifiers(modifiers []*EquipmentModifier, filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, &equipmentModifierListData{
		Version: jio.CurrentDataVersion,
		Rows:    modifiers,
	})
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
func (e *EquipmentModifier) Clone(from LibraryFile, owner DataOwner, parent *EquipmentModifier, preserveID bool) *EquipmentModifier {
	other := NewEquipmentModifier(owner, parent, e.Container())
	other.AdjustSource(from, e.SourcedID, preserveID)
	other.SetOpen(e.IsOpen())
	other.ThirdParty = e.ThirdParty
	other.CopyFrom(e)
	if e.HasChildren() {
		other.Children = make([]*EquipmentModifier, 0, len(e.Children))
		for _, child := range e.Children {
			other.Children = append(other.Children, child.Clone(from, owner, other, preserveID))
		}
	}
	return other
}

// MarshalJSON implements json.Marshaler.
func (e *EquipmentModifier) MarshalJSON() ([]byte, error) {
	type calc struct {
		ResolvedNotes string `json:"resolved_notes,omitempty"`
	}
	e.ClearUnusedFieldsForType()
	data := struct {
		EquipmentModifierData
		Calc *calc `json:"calc,omitempty"`
	}{
		EquipmentModifierData: e.EquipmentModifierData,
	}
	notes := e.ResolveLocalNotes()
	if notes != e.LocalNotes {
		data.Calc = &calc{ResolvedNotes: notes}
	}
	return json.Marshal(&data)
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *EquipmentModifier) UnmarshalJSON(data []byte) error {
	var localData struct {
		EquipmentModifierData
		// Old data fields
		Type       string   `json:"type"`
		ExprNotes  string   `json:"notes"`
		Categories []string `json:"categories"`
		IsOpen     bool     `json:"open"`
	}
	if err := json.Unmarshal(data, &localData); err != nil {
		return err
	}
	setOpen := false
	if !tid.IsValid(localData.TID) {
		// Fixup old data that used UUIDs instead of TIDs
		localData.TID = tid.MustNewTID(equipmentModifierKind(strings.HasSuffix(localData.Type, containerKeyPostfix)))
		setOpen = localData.IsOpen
	}
	e.EquipmentModifierData = localData.EquipmentModifierData
	if e.LocalNotes == "" && localData.ExprNotes != "" {
		e.LocalNotes = EmbeddedExprToScript(localData.ExprNotes)
	}
	e.ClearUnusedFieldsForType()
	e.Tags = convertOldCategoriesToTags(e.Tags, localData.Categories)
	slices.Sort(e.Tags)
	if e.Container() {
		for _, one := range e.Children {
			one.parent = e
		}
	}
	if setOpen {
		SetNodeOpen(e, true)
	}
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
		data.Title = HeaderCheckmark
		data.TitleIsImageKey = true
		data.Detail = ModifierEnabledTooltip()
	case EquipmentModifierDescriptionColumn:
		data.Title = i18n.Text("Equipment Modifier")
		data.Primary = true
	case EquipmentModifierTechLevelColumn:
		data.Title = i18n.Text("TL")
		data.Detail = i18n.Text("Tech Level")
	case EquipmentModifierCostColumn:
		data.Title = i18n.Text("Cost Adjustment")
	case EquipmentModifierWeightColumn:
		data.Title = i18n.Text("Weight Adjustment")
	case EquipmentModifierTagsColumn:
		data.Title = i18n.Text("Tags")
	case EquipmentModifierReferenceColumn:
		data.Title = HeaderBookmark
		data.TitleIsImageKey = true
		data.Detail = PageRefTooltip()
	case EquipmentModifierLibSrcColumn:
		data.Title = HeaderDatabase
		data.TitleIsImageKey = true
		data.Detail = LibSrcTooltip()
	}
	return data
}

// CellData returns the cell data information for the given column.
func (e *EquipmentModifier) CellData(columnID int, data *CellData) {
	switch columnID {
	case EquipmentModifierEnabledColumn:
		if !e.Container() {
			data.Type = cell.Toggle
			data.Checked = e.Enabled()
			data.Alignment = align.Middle
		}
	case EquipmentModifierDescriptionColumn:
		data.Type = cell.Text
		data.Primary = e.NameWithReplacements()
		data.Secondary = e.SecondaryText(func(option display.Option) bool { return option.Inline() })
		data.Tooltip = e.SecondaryText(func(option display.Option) bool { return option.Tooltip() })
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
		data.Type = cell.Tags
		data.Primary = CombineTags(e.Tags)
	case EquipmentModifierReferenceColumn, PageRefCellAlias:
		data.Type = cell.PageRef
		data.Primary = e.PageRef
		if e.PageRefHighlight != "" {
			data.Secondary = e.PageRefHighlight
		} else {
			data.Secondary = e.NameWithReplacements()
		}
	case EquipmentModifierLibSrcColumn:
		data.Type = cell.Text
		data.Alignment = align.Middle
		if !toolbox.IsNil(e.owner) {
			state, _ := e.owner.SourceMatcher().Match(e)
			data.Primary = state.AltString()
			data.Tooltip = state.String()
			if state != srcstate.Custom {
				data.Tooltip += "\n" + e.Source.String()
			}
		}
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

// OwningEquipment returns the owning equipment.
func (e *EquipmentModifier) OwningEquipment() *Equipment {
	return e.equipment
}

// DataOwner returns the data owner.
func (e *EquipmentModifier) DataOwner() DataOwner {
	return e.owner
}

func (e *EquipmentModifier) setEquipment(equipment *Equipment) {
	e.equipment = equipment
	if equipment != nil && len(e.Replacements) != 0 {
		if e.equipment.Replacements == nil {
			e.equipment.Replacements = e.Replacements
		} else {
			for k, v := range e.Replacements {
				if _, exists := e.equipment.Replacements[k]; !exists {
					e.equipment.Replacements[k] = v
				}
			}
		}
		e.Replacements = nil
	}
	if e.Container() {
		for _, child := range e.Children {
			child.setEquipment(equipment)
		}
	}
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

// ResolveLocalNotes resolves the local notes, running any embedded scripts to get the final result.
func (e *EquipmentModifier) ResolveLocalNotes() string {
	return ResolveText(EntityFromNode(e), DeferredNewScriptEquipment(e.equipment), e.LocalNotesWithReplacements())
}

// SecondaryText returns the "secondary" text: the text display below an Trait.
func (e *EquipmentModifier) SecondaryText(optionChecker func(display.Option) bool) string {
	if !optionChecker(SheetSettingsFor(EntityFromNode(e)).NotesDisplay) {
		return ""
	}
	return e.ResolveLocalNotes()
}

// FullDescription returns a full description.
func (e *EquipmentModifier) FullDescription() string {
	var buffer strings.Builder
	buffer.WriteString(e.String())
	if localNotes := e.ResolveLocalNotes(); localNotes != "" {
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
	return e.CostType.Format(e.CostAmount) + " " + e.CostType.String()
}

// WeightDescription returns the formatted weight.
func (e *EquipmentModifier) WeightDescription() string {
	if e.Container() || (e.WeightType == emweight.Original && (e.WeightAmount == "" || strings.HasPrefix(e.WeightAmount, "+0 "))) {
		return ""
	}
	return e.WeightType.Format(e.WeightAmount, SheetSettingsFor(EntityFromNode(e)).DefaultWeightUnits) + " " +
		e.WeightType.String()
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
	if e.equipment == nil {
		return e.Name
	}
	return nameable.Apply(e.Name, e.equipment.Replacements)
}

// LocalNotesWithReplacements returns the local notes with any replacements applied.
func (e *EquipmentModifier) LocalNotesWithReplacements() string {
	if e.equipment == nil {
		return e.LocalNotes
	}
	return nameable.Apply(e.LocalNotes, e.equipment.Replacements)
}

// FillWithNameableKeys adds any nameable keys found in this EquipmentModifier to the provided map.
func (e *EquipmentModifier) FillWithNameableKeys(m, existing map[string]string) {
	if e.Enabled() {
		if existing == nil && e.equipment != nil {
			existing = e.equipment.Replacements
		}
		nameable.Extract(e.Name, m, existing)
		nameable.Extract(e.LocalNotes, m, existing)
		for _, one := range e.Features {
			one.FillWithNameableKeys(m, existing)
		}
	}
}

// ApplyNameableKeys passes this up to the owning equipment to handle.
func (e *EquipmentModifier) ApplyNameableKeys(m map[string]string) {
	if len(m) != 0 && e.equipment != nil {
		e.equipment.ApplyNameableKeys(m)
	}
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

// CostMultiplier returns the amount to multiply the cost by.
func (e *EquipmentModifier) CostMultiplier() fxp.Int {
	return MultiplierForEquipmentModifier(e.equipment, e.CostIsPerLevel)
}

// WeightMultiplier returns the amount to multiply the weight by.
func (e *EquipmentModifier) WeightMultiplier() fxp.Int {
	return MultiplierForEquipmentModifier(e.equipment, e.WeightIsPerLevel)
}

// MultiplierForEquipmentModifier returns the amount to multiply the cost or weight by.
func MultiplierForEquipmentModifier(equipment *Equipment, isPerLevel bool) fxp.Int {
	var multiplier fxp.Int
	if isPerLevel && equipment != nil && equipment.IsLeveled() {
		multiplier = equipment.CurrentLevel()
	}
	if multiplier <= 0 {
		multiplier = fxp.One
	}
	return multiplier
}

// ValueAdjustedForModifiers returns the value after adjusting it for a set of modifiers.
func ValueAdjustedForModifiers(equipment *Equipment, value fxp.Int, modifiers []*EquipmentModifier) fxp.Int {
	// Apply all equipment.OriginalCost
	cost := processNonCFStep(equipment, emcost.Original, value, modifiers)

	// Apply all equipment.BaseCost
	var cf fxp.Int
	Traverse(func(mod *EquipmentModifier) bool {
		mod.equipment = equipment
		if mod.CostType == emcost.Base {
			t := emcost.Base.FromString(mod.CostAmount)
			cf += t.ExtractValue(mod.CostAmount).Mul(mod.CostMultiplier())
			if t == emcost.Multiplier {
				cf -= fxp.One
			}
		}
		return false
	}, true, true, modifiers...)
	if cf != 0 {
		cf = cf.Max(fxp.NegPointEight)
		cost = cost.Mul(cf.Max(fxp.NegPointEight) + fxp.One)
	}

	// Apply all equipment.FinalBaseCost
	cost = processNonCFStep(equipment, emcost.FinalBase, cost, modifiers)

	// Apply all equipment.FinalCost
	cost = processNonCFStep(equipment, emcost.Final, cost, modifiers)

	return cost.Max(0)
}

func processNonCFStep(equipment *Equipment, costType emcost.Type, value fxp.Int, modifiers []*EquipmentModifier) fxp.Int {
	var percentages, additions fxp.Int
	cost := value
	Traverse(func(mod *EquipmentModifier) bool {
		mod.equipment = equipment
		if mod.CostType == costType {
			t := costType.FromString(mod.CostAmount)
			amt := t.ExtractValue(mod.CostAmount).Mul(mod.CostMultiplier())
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

	// Apply all equipment.OriginalWeight
	Traverse(func(mod *EquipmentModifier) bool {
		mod.equipment = equipment
		if mod.WeightType == emweight.Original {
			t := emweight.Original.DetermineModifierWeightValueTypeFromString(mod.WeightAmount)
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

	// Apply all equipment.BaseWeight
	w = processMultiplyAddWeightStep(equipment, emweight.Base, w, defUnits, modifiers)

	// Apply all equipment.FinalBaseWeight
	w = processMultiplyAddWeightStep(equipment, emweight.FinalBase, w, defUnits, modifiers)

	// Apply all equipment.FinalWeight
	w = processMultiplyAddWeightStep(equipment, emweight.Final, w, defUnits, modifiers)

	return fxp.Weight(w.Max(0))
}

func processMultiplyAddWeightStep(equipment *Equipment, weightType emweight.Type, weight fxp.Int, defUnits fxp.WeightUnit, modifiers []*EquipmentModifier) fxp.Int {
	var sum fxp.Int
	Traverse(func(mod *EquipmentModifier) bool {
		mod.equipment = equipment
		if mod.WeightType == weightType {
			t := weightType.DetermineModifierWeightValueTypeFromString(mod.WeightAmount)
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
		return i18n.Text("Equipment Modifier Container")
	}
	return i18n.Text("Equipment Modifier")
}

// ClearUnusedFieldsForType zeroes out the fields that are not applicable to this type (container vs not-container).
func (e *EquipmentModifier) ClearUnusedFieldsForType() {
	if e.Container() {
		e.EquipmentModifierEditDataNonContainerOnly = EquipmentModifierEditDataNonContainerOnly{}
	} else {
		e.Children = nil
	}
}

// GetSource returns the source of this data.
func (e *EquipmentModifier) GetSource() Source {
	return e.Source
}

// ClearSource clears the source of this data.
func (e *EquipmentModifier) ClearSource() {
	e.Source = Source{}
}

// SyncWithSource synchronizes this data with the source.
func (e *EquipmentModifier) SyncWithSource() {
	if !toolbox.IsNil(e.owner) {
		if state, data := e.owner.SourceMatcher().Match(e); state == srcstate.Mismatched {
			if other, ok := data.(*EquipmentModifier); ok {
				e.EquipmentModifierSyncData = other.EquipmentModifierSyncData
				e.Tags = slices.Clone(other.Tags)
				if !e.Container() {
					e.EquipmentModifierNonContainerSyncData = other.EquipmentModifierNonContainerSyncData
					e.Features = other.Features.Clone()
				}
			}
		}
	}
}

// Hash writes this object's contents into the hasher. Note that this only hashes the data that is considered to be
// "source" data, i.e. not expected to be modified by the user after copying from a library.
func (e *EquipmentModifier) Hash(h hash.Hash) {
	e.hash(h)
	if !e.Container() {
		e.EquipmentModifierNonContainerSyncData.hash(h)
	}
}

func (e *EquipmentModifierSyncData) hash(h hash.Hash) {
	hashhelper.String(h, e.Name)
	hashhelper.String(h, e.PageRef)
	hashhelper.String(h, e.PageRefHighlight)
	hashhelper.String(h, e.LocalNotes)
	hashhelper.Num64(h, len(e.Tags))
	for _, tag := range e.Tags {
		hashhelper.String(h, tag)
	}
}

func (e *EquipmentModifierNonContainerSyncData) hash(h hash.Hash) {
	hashhelper.Num8(h, e.CostType)
	hashhelper.Num8(h, e.WeightType)
	hashhelper.Bool(h, e.ShowNotesOnWeapon)
	hashhelper.String(h, e.TechLevel)
	hashhelper.String(h, e.CostAmount)
	hashhelper.String(h, e.WeightAmount)
	hashhelper.Num64(h, len(e.Features))
	for _, feature := range e.Features {
		feature.Hash(h)
	}
}

// CopyFrom implements node.EditorData.
func (e *EquipmentModifierEditData) CopyFrom(other *EquipmentModifier) {
	e.copyFrom(&other.EquipmentModifierEditData)
}

// ApplyTo implements node.EditorData.
func (e *EquipmentModifierEditData) ApplyTo(other *EquipmentModifier) {
	other.copyFrom(e)
}

func (e *EquipmentModifierEditData) copyFrom(other *EquipmentModifierEditData) {
	*e = *other
	e.Tags = txt.CloneStringSlice(other.Tags)
	e.Features = other.Features.Clone()
}
