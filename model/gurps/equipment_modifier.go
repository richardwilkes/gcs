/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package gurps

import (
	"context"
	"io/fs"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/unison/enums/align"
)

var (
	_ Node[*EquipmentModifier] = &EquipmentModifier{}
	_ GeneralModifier          = &EquipmentModifier{}
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

type equipmentModifierListData struct {
	Type    string               `json:"type"`
	Version int                  `json:"version"`
	Rows    []*EquipmentModifier `json:"rows"`
}

// NewEquipmentModifiersFromFile loads an EquipmentModifier list from a file.
func NewEquipmentModifiersFromFile(fileSystem fs.FS, filePath string) ([]*EquipmentModifier, error) {
	var data equipmentModifierListData
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &data); err != nil {
		return nil, errs.NewWithCause(invalidFileDataMsg(), err)
	}
	if data.Type != equipmentModifierListTypeKey {
		return nil, errs.New(unexpectedFileDataMsg())
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
		Calc calc `json:"calc"`
	}{
		EquipmentModifierData: m.EquipmentModifierData,
	}
	notes := m.resolveLocalNotes()
	if notes != m.LocalNotes {
		data.Calc.ResolvedNotes = notes
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
	m.Tags = convertOldCategoriesToTags(m.Tags, localData.Categories)
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

// CellData returns the cell data information for the given column.
func (m *EquipmentModifier) CellData(columnID int, data *CellData) {
	switch columnID {
	case EquipmentModifierEnabledColumn:
		if !m.Container() {
			data.Type = ToggleCellType
			data.Checked = m.Enabled()
			data.Alignment = align.Middle
		}
	case EquipmentModifierDescriptionColumn:
		data.Type = TextCellType
		data.Primary = m.Name
		data.Secondary = m.SecondaryText(func(option DisplayOption) bool { return option.Inline() })
		data.Tooltip = m.SecondaryText(func(option DisplayOption) bool { return option.Tooltip() })
	case EquipmentModifierTechLevelColumn:
		if !m.Container() {
			data.Type = TextCellType
			data.Primary = m.TechLevel
		}
	case EquipmentModifierCostColumn:
		if !m.Container() {
			data.Type = TextCellType
			data.Primary = m.CostDescription()
		}
	case EquipmentModifierWeightColumn:
		if !m.Container() {
			data.Type = TextCellType
			data.Primary = m.WeightDescription()
		}
	case EquipmentModifierTagsColumn:
		data.Type = TagsCellType
		data.Primary = CombineTags(m.Tags)
	case EquipmentModifierReferenceColumn, PageRefCellAlias:
		data.Type = PageRefCellType
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
func (m *EquipmentModifier) SecondaryText(optionChecker func(DisplayOption) bool) string {
	var buffer strings.Builder
	settings := SheetSettingsFor(m.Entity)
	if localNotes := m.resolveLocalNotes(); localNotes != "" && optionChecker(settings.NotesDisplay) {
		if buffer.Len() != 0 {
			buffer.WriteByte('\n')
		}
		buffer.WriteString(localNotes)
	}
	return buffer.String()
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
	if m.Container() || (m.CostType == OriginalEquipmentModifierCostType && (m.CostAmount == "" || m.CostAmount == "+0")) {
		return ""
	}
	return m.CostType.Format(m.CostAmount) + " " + m.CostType.String()
}

// WeightDescription returns the formatted weight.
func (m *EquipmentModifier) WeightDescription() string {
	if m.Container() || (m.WeightType == OriginalEquipmentModifierWeightType && (m.WeightAmount == "" || strings.HasPrefix(m.WeightAmount, "+0 "))) {
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
	cost := processNonCFStep(OriginalEquipmentModifierCostType, value, modifiers)

	// Apply all equipment.BaseCost
	var cf fxp.Int
	Traverse(func(mod *EquipmentModifier) bool {
		if mod.CostType == BaseEquipmentModifierCostType {
			t := BaseEquipmentModifierCostType.FromString(mod.CostAmount)
			cf += t.ExtractValue(mod.CostAmount)
			if t == MultiplierEquipmentModifierCostValueType {
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
	cost = processNonCFStep(FinalBaseEquipmentModifierCostType, cost, modifiers)

	// Apply all equipment.FinalCost
	cost = processNonCFStep(FinalEquipmentModifierCostType, cost, modifiers)

	return cost.Max(0)
}

func processNonCFStep(costType EquipmentModifierCostType, value fxp.Int, modifiers []*EquipmentModifier) fxp.Int {
	var percentages, additions fxp.Int
	cost := value
	Traverse(func(mod *EquipmentModifier) bool {
		if mod.CostType == costType {
			t := costType.FromString(mod.CostAmount)
			amt := t.ExtractValue(mod.CostAmount)
			switch t {
			case AdditionEquipmentModifierCostValueType:
				additions += amt
			case PercentageEquipmentModifierCostValueType:
				percentages += amt
			case MultiplierEquipmentModifierCostValueType:
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
func WeightAdjustedForModifiers(weight fxp.Weight, modifiers []*EquipmentModifier, defUnits fxp.WeightUnits) fxp.Weight {
	var percentages fxp.Int
	w := fxp.Int(weight)

	// Apply all equipment.OriginalWeight
	Traverse(func(mod *EquipmentModifier) bool {
		if mod.WeightType == OriginalEquipmentModifierWeightType {
			t := OriginalEquipmentModifierWeightType.DetermineModifierWeightValueTypeFromString(mod.WeightAmount)
			amt := t.ExtractFraction(mod.WeightAmount).Value()
			if t == AdditionEquipmentModifierWeightValueType {
				w += fxp.TrailingWeightUnitsFromString(mod.WeightAmount, defUnits).ToPounds(amt)
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
	w = processMultiplyAddWeightStep(BaseEquipmentModifierWeightType, w, defUnits, modifiers)

	// Apply all equipment.FinalBaseWeight
	w = processMultiplyAddWeightStep(FinalBaseEquipmentModifierWeightType, w, defUnits, modifiers)

	// Apply all equipment.FinalWeight
	w = processMultiplyAddWeightStep(FinalEquipmentModifierWeightType, w, defUnits, modifiers)

	return fxp.Weight(w.Max(0))
}

func processMultiplyAddWeightStep(weightType EquipmentModifierWeightType, weight fxp.Int, defUnits fxp.WeightUnits, modifiers []*EquipmentModifier) fxp.Int {
	var sum fxp.Int
	Traverse(func(mod *EquipmentModifier) bool {
		if mod.WeightType == weightType {
			t := weightType.DetermineModifierWeightValueTypeFromString(mod.WeightAmount)
			f := t.ExtractFraction(mod.WeightAmount)
			switch t {
			case AdditionEquipmentModifierWeightValueType:
				sum += fxp.TrailingWeightUnitsFromString(mod.WeightAmount, defUnits).ToPounds(f.Value())
			case PercentageMultiplierEquipmentModifierWeightValueType:
				weight = weight.Mul(f.Numerator).Div(f.Denominator.Mul(fxp.Hundred))
			case MultiplierEquipmentModifierWeightValueType:
				weight = weight.Mul(f.Numerator).Div(f.Denominator)
			}
		}
		return false
	}, true, true, modifiers...)
	return weight + sum
}
