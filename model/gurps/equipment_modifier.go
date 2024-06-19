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
	"io/fs"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emcost"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emweight"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
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
)

const (
	equipmentModifierListTypeKey = "eqp_modifier_list"
	equipmentModifierTypeKey     = "eqp_modifier"
)

// EquipmentModifier holds a modifier to a piece of Equipment.
type EquipmentModifier struct {
	EquipmentModifierData
	Entity *Entity
}

// EquipmentModifierData holds the EquipmentModifier data that is written to disk.
type EquipmentModifierData struct {
	ContainerBase[*EquipmentModifier]
	EquipmentModifierEditData
}

// EquipmentModifierEditData holds the EquipmentModifier data that can be edited by the UI detail editor.
type EquipmentModifierEditData struct {
	Name             string   `json:"name,omitempty"`
	PageRef          string   `json:"reference,omitempty"`
	PageRefHighlight string   `json:"reference_highlight,omitempty"`
	LocalNotes       string   `json:"notes,omitempty"`
	VTTNotes         string   `json:"vtt_notes,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	EquipmentModifierEditDataNonContainerOnly
}

// EquipmentModifierEditDataNonContainerOnly holds the EquipmentModifier data that is only applicable to
// EquipmentModifiers that aren't containers.
type EquipmentModifierEditDataNonContainerOnly struct {
	CostType     emcost.Type   `json:"cost_type,omitempty"`
	WeightType   emweight.Type `json:"weight_type,omitempty"`
	Disabled     bool          `json:"disabled,omitempty"`
	TechLevel    string        `json:"tech_level,omitempty"`
	CostAmount   string        `json:"cost,omitempty"`
	WeightAmount string        `json:"weight,omitempty"`
	Features     Features      `json:"features,omitempty"`
}

type equipmentModifierListData struct {
	Type    string               `json:"type"`
	Version int                  `json:"version"`
	Rows    []*EquipmentModifier `json:"rows"`
}

// NewEquipmentModifiersFromFile loads an EquipmentModifier list from a file.
func NewEquipmentModifiersFromFile(fileSystem fs.FS, filePath string) ([]*EquipmentModifier, error) {
	var data equipmentModifierListData
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &data); err != nil {
		return nil, errs.NewWithCause(InvalidFileDataMsg(), err)
	}
	if data.Type != equipmentModifierListTypeKey {
		return nil, errs.New(UnexpectedFileDataMsg())
	}
	if err := CheckVersion(data.Version); err != nil {
		return nil, err
	}
	return data.Rows, nil
}

// SaveEquipmentModifiers writes the EquipmentModifier list to the file as JSON.
func SaveEquipmentModifiers(modifiers []*EquipmentModifier, filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, &equipmentModifierListData{
		Type:    equipmentModifierListTypeKey,
		Version: CurrentDataVersion,
		Rows:    modifiers,
	})
}

// NewEquipmentModifier creates an EquipmentModifier.
func NewEquipmentModifier(entity *Entity, parent *EquipmentModifier, container bool) *EquipmentModifier {
	a := &EquipmentModifier{
		EquipmentModifierData: EquipmentModifierData{
			ContainerBase: newContainerBase[*EquipmentModifier](equipmentModifierTypeKey, container),
		},
		Entity: entity,
	}
	a.Name = a.Kind()
	a.parent = parent
	return a
}

// Clone implements Node.
func (m *EquipmentModifier) Clone(entity *Entity, parent *EquipmentModifier, preserveID bool) *EquipmentModifier {
	other := NewEquipmentModifier(entity, parent, m.Container())
	if preserveID {
		other.ID = m.ID
	}
	other.IsOpen = m.IsOpen
	other.ThirdParty = m.ThirdParty
	other.EquipmentModifierEditData.CopyFrom(m)
	if m.HasChildren() {
		other.Children = make([]*EquipmentModifier, 0, len(m.Children))
		for _, child := range m.Children {
			other.Children = append(other.Children, child.Clone(entity, other, preserveID))
		}
	}
	return other
}

// MarshalJSON implements json.Marshaler.
func (m *EquipmentModifier) MarshalJSON() ([]byte, error) {
	type calc struct {
		ResolvedNotes string `json:"resolved_notes,omitempty"`
	}
	m.ClearUnusedFieldsForType()
	data := struct {
		EquipmentModifierData
		Calc *calc `json:"calc,omitempty"`
	}{
		EquipmentModifierData: m.EquipmentModifierData,
	}
	notes := m.resolveLocalNotes()
	if notes != m.LocalNotes {
		data.Calc = &calc{ResolvedNotes: notes}
	}
	return json.Marshal(&data)
}

// UnmarshalJSON implements json.Unmarshaler.
func (m *EquipmentModifier) UnmarshalJSON(data []byte) error {
	var localData struct {
		EquipmentModifierData
		// Old data fields
		Categories []string `json:"categories"`
	}
	if err := json.Unmarshal(data, &localData); err != nil {
		return err
	}
	localData.ClearUnusedFieldsForType()
	m.EquipmentModifierData = localData.EquipmentModifierData
	m.Tags = ConvertOldCategoriesToTags(m.Tags, localData.Categories)
	slices.Sort(m.Tags)
	if m.Container() {
		for _, one := range m.Children {
			one.parent = m
		}
	}
	return nil
}

// TagList returns the list of tags.
func (m *EquipmentModifier) TagList() []string {
	return m.Tags
}

// EquipmentModifierHeaderData returns the header data information for the given equipment modifier column.
func EquipmentModifierHeaderData(columnID int) HeaderData {
	var data HeaderData
	switch columnID {
	case EquipmentModifierEnabledColumn:
		data.Title = HeaderCheckmark
		data.TitleIsImageKey = true
		data.Detail = i18n.Text("Whether this item is enabled. Items that are not enabled do not apply any features they may normally contribute to the character.")
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
		data.Detail = PageRefTooltipText()
	}
	return data
}

// CellData returns the cell data information for the given column.
func (m *EquipmentModifier) CellData(columnID int, data *CellData) {
	switch columnID {
	case EquipmentModifierEnabledColumn:
		if !m.Container() {
			data.Type = cell.Toggle
			data.Checked = m.Enabled()
			data.Alignment = align.Middle
		}
	case EquipmentModifierDescriptionColumn:
		data.Type = cell.Text
		data.Primary = m.Name
		data.Secondary = m.SecondaryText(func(option display.Option) bool { return option.Inline() })
		data.Tooltip = m.SecondaryText(func(option display.Option) bool { return option.Tooltip() })
	case EquipmentModifierTechLevelColumn:
		if !m.Container() {
			data.Type = cell.Text
			data.Primary = m.TechLevel
		}
	case EquipmentModifierCostColumn:
		if !m.Container() {
			data.Type = cell.Text
			data.Primary = m.CostDescription()
		}
	case EquipmentModifierWeightColumn:
		if !m.Container() {
			data.Type = cell.Text
			data.Primary = m.WeightDescription()
		}
	case EquipmentModifierTagsColumn:
		data.Type = cell.Tags
		data.Primary = CombineTags(m.Tags)
	case EquipmentModifierReferenceColumn, PageRefCellAlias:
		data.Type = cell.PageRef
		data.Primary = m.PageRef
		if m.PageRefHighlight != "" {
			data.Secondary = m.PageRefHighlight
		} else {
			data.Secondary = m.Name
		}
	}
}

// Depth returns the number of parents this node has.
func (m *EquipmentModifier) Depth() int {
	count := 0
	p := m.parent
	for p != nil {
		count++
		p = p.parent
	}
	return count
}

// OwningEntity returns the owning Entity.
func (m *EquipmentModifier) OwningEntity() *Entity {
	return m.Entity
}

// SetOwningEntity sets the owning entity and configures any sub-components as needed.
func (m *EquipmentModifier) SetOwningEntity(entity *Entity) {
	m.Entity = entity
	if m.Container() {
		for _, child := range m.Children {
			child.SetOwningEntity(entity)
		}
	}
}

func (m *EquipmentModifier) String() string {
	return m.Name
}

func (m *EquipmentModifier) resolveLocalNotes() string {
	return EvalEmbeddedRegex.ReplaceAllStringFunc(m.LocalNotes, m.Entity.EmbeddedEval)
}

// SecondaryText returns the "secondary" text: the text display below an Trait.
func (m *EquipmentModifier) SecondaryText(optionChecker func(display.Option) bool) string {
	if !optionChecker(SheetSettingsFor(m.Entity).NotesDisplay) {
		return ""
	}
	return m.resolveLocalNotes()
}

// FullDescription returns a full description.
func (m *EquipmentModifier) FullDescription() string {
	var buffer strings.Builder
	buffer.WriteString(m.String())
	if localNotes := m.resolveLocalNotes(); localNotes != "" {
		buffer.WriteString(" (")
		buffer.WriteString(m.LocalNotes)
		buffer.WriteByte(')')
	}
	if SheetSettingsFor(m.Entity).ShowEquipmentModifierAdj {
		costDesc := m.CostDescription()
		weightDesc := m.WeightDescription()
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
func (m *EquipmentModifier) FullCostDescription() string {
	cost := m.CostDescription()
	weight := m.WeightDescription()
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
func (m *EquipmentModifier) CostDescription() string {
	if m.Container() || (m.CostType == emcost.Original && (m.CostAmount == "" || m.CostAmount == "+0")) {
		return ""
	}
	return m.CostType.Format(m.CostAmount) + " " + m.CostType.String()
}

// WeightDescription returns the formatted weight.
func (m *EquipmentModifier) WeightDescription() string {
	if m.Container() || (m.WeightType == emweight.Original && (m.WeightAmount == "" || strings.HasPrefix(m.WeightAmount, "+0 "))) {
		return ""
	}
	return m.WeightType.Format(m.WeightAmount, SheetSettingsFor(m.Entity).DefaultWeightUnits) + " " + m.WeightType.String()
}

// FillWithNameableKeys adds any nameable keys found in this EquipmentModifier to the provided map.
func (m *EquipmentModifier) FillWithNameableKeys(keyMap map[string]string) {
	if m.Enabled() {
		Extract(m.Name, keyMap)
		Extract(m.LocalNotes, keyMap)
		for _, one := range m.Features {
			one.FillWithNameableKeys(keyMap)
		}
	}
}

// ApplyNameableKeys replaces any nameable keys found in this EquipmentModifier with the corresponding values in the provided map.
func (m *EquipmentModifier) ApplyNameableKeys(keyMap map[string]string) {
	if m.Enabled() {
		m.Name = Apply(m.Name, keyMap)
		m.LocalNotes = Apply(m.LocalNotes, keyMap)
		for _, one := range m.Features {
			one.ApplyNameableKeys(keyMap)
		}
	}
}

// Enabled returns true if this node is enabled.
func (m *EquipmentModifier) Enabled() bool {
	return !m.Disabled || m.Container()
}

// SetEnabled makes the node enabled, if possible.
func (m *EquipmentModifier) SetEnabled(enabled bool) {
	if !m.Container() {
		m.Disabled = !enabled
	}
}

// ValueAdjustedForModifiers returns the value after adjusting it for a set of modifiers.
func ValueAdjustedForModifiers(value fxp.Int, modifiers []*EquipmentModifier) fxp.Int {
	// Apply all equipment.OriginalCost
	cost := processNonCFStep(emcost.Original, value, modifiers)

	// Apply all equipment.BaseCost
	var cf fxp.Int
	Traverse(func(mod *EquipmentModifier) bool {
		if mod.CostType == emcost.Base {
			t := emcost.Base.FromString(mod.CostAmount)
			cf += t.ExtractValue(mod.CostAmount)
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
	cost = processNonCFStep(emcost.FinalBase, cost, modifiers)

	// Apply all equipment.FinalCost
	cost = processNonCFStep(emcost.Final, cost, modifiers)

	return cost.Max(0)
}

func processNonCFStep(costType emcost.Type, value fxp.Int, modifiers []*EquipmentModifier) fxp.Int {
	var percentages, additions fxp.Int
	cost := value
	Traverse(func(mod *EquipmentModifier) bool {
		if mod.CostType == costType {
			t := costType.FromString(mod.CostAmount)
			amt := t.ExtractValue(mod.CostAmount)
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
func WeightAdjustedForModifiers(weight fxp.Weight, modifiers []*EquipmentModifier, defUnits fxp.WeightUnit) fxp.Weight {
	var percentages fxp.Int
	w := fxp.Int(weight)

	// Apply all equipment.OriginalWeight
	Traverse(func(mod *EquipmentModifier) bool {
		if mod.WeightType == emweight.Original {
			t := emweight.Original.DetermineModifierWeightValueTypeFromString(mod.WeightAmount)
			amt := t.ExtractFraction(mod.WeightAmount).Value()
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
	w = processMultiplyAddWeightStep(emweight.Base, w, defUnits, modifiers)

	// Apply all equipment.FinalBaseWeight
	w = processMultiplyAddWeightStep(emweight.FinalBase, w, defUnits, modifiers)

	// Apply all equipment.FinalWeight
	w = processMultiplyAddWeightStep(emweight.Final, w, defUnits, modifiers)

	return fxp.Weight(w.Max(0))
}

func processMultiplyAddWeightStep(weightType emweight.Type, weight fxp.Int, defUnits fxp.WeightUnit, modifiers []*EquipmentModifier) fxp.Int {
	var sum fxp.Int
	Traverse(func(mod *EquipmentModifier) bool {
		if mod.WeightType == weightType {
			t := weightType.DetermineModifierWeightValueTypeFromString(mod.WeightAmount)
			f := t.ExtractFraction(mod.WeightAmount)
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
func (d *EquipmentModifierData) Kind() string {
	return d.kind(i18n.Text("Equipment Modifier"))
}

// ClearUnusedFieldsForType zeroes out the fields that are not applicable to this type (container vs not-container).
func (d *EquipmentModifierData) ClearUnusedFieldsForType() {
	d.clearUnusedFields()
	if d.Container() {
		d.CostType = 0
		d.WeightType = 0
		d.Disabled = false
		d.TechLevel = ""
		d.CostAmount = ""
		d.WeightAmount = ""
		d.Features = nil
	}
}

// CopyFrom implements node.EditorData.
func (d *EquipmentModifierEditData) CopyFrom(mod *EquipmentModifier) {
	d.copyFrom(&mod.EquipmentModifierEditData)
}

// ApplyTo implements node.EditorData.
func (d *EquipmentModifierEditData) ApplyTo(mod *EquipmentModifier) {
	mod.EquipmentModifierEditData.copyFrom(d)
}

func (d *EquipmentModifierEditData) copyFrom(other *EquipmentModifierEditData) {
	*d = *other
	d.Tags = txt.CloneStringSlice(d.Tags)
	d.Features = other.Features.Clone()
}
