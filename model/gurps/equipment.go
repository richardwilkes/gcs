/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
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
	"strconv"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/gid"
	"github.com/richardwilkes/gcs/v5/model/gurps/measure"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/settings/display"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"golang.org/x/exp/slices"
)

var (
	_ WeaponOwner                   = &Equipment{}
	_ Node[*Equipment]              = &Equipment{}
	_ TechLevelProvider[*Equipment] = &Equipment{}
)

// Columns that can be used with the equipment method .CellData()
const (
	EquipmentEquippedColumn = iota
	EquipmentQuantityColumn
	EquipmentDescriptionColumn
	EquipmentUsesColumn
	EquipmentMaxUsesColumn
	EquipmentTLColumn
	EquipmentLCColumn
	EquipmentCostColumn
	EquipmentExtendedCostColumn
	EquipmentWeightColumn
	EquipmentExtendedWeightColumn
	EquipmentTagsColumn
	EquipmentReferenceColumn
)

var (
	// TechLevelInfo holds the general TL age list
	TechLevelInfo = i18n.Text(`TL0: Stone Age (Prehistory)
TL1: Bronze Age (3500 B.C.+)
TL2: Iron Age (1200 B.C.+)
TL3: Medieval (600 A.D.+)
TL4: Age of Sail (1450+)
TL5: Industrial Revolution (1730+)
TL6: Mechanized Age (1880+)
TL7: Nuclear Age (1940+)
TL8: Digital Age (1980+)
TL9: Microtech Age (2025+?)
TL10: Robotic Age (2070+?)
TL11: Age of Exotic Matter
TL12: Anything Goes`)

	// LegalityClassInfo holds the LC list
	LegalityClassInfo = i18n.Text(`LC0: Banned
LC1: Military
LC2: Restricted
LC3: Licensed
LC4: Open`)
)

const (
	equipmentListTypeKey = "equipment_list"
	equipmentTypeKey     = "equipment"
)

// Equipment holds a piece of equipment.
type Equipment struct {
	EquipmentData
	Entity            *Entity
	UnsatisfiedReason string
}

type equipmentListData struct {
	Type    string       `json:"type"`
	Version int          `json:"version"`
	Rows    []*Equipment `json:"rows"`
}

// NewEquipmentFromFile loads an Equipment list from a file.
func NewEquipmentFromFile(fileSystem fs.FS, filePath string) ([]*Equipment, error) {
	var data equipmentListData
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &data); err != nil {
		return nil, errs.NewWithCause(gid.InvalidFileDataMsg, err)
	}
	if data.Type != equipmentListTypeKey {
		return nil, errs.New(gid.UnexpectedFileDataMsg)
	}
	if err := gid.CheckVersion(data.Version); err != nil {
		return nil, err
	}
	return data.Rows, nil
}

// SaveEquipment writes the Equipment list to the file as JSON.
func SaveEquipment(equipment []*Equipment, filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, &equipmentListData{
		Type:    equipmentListTypeKey,
		Version: gid.CurrentDataVersion,
		Rows:    equipment,
	})
}

// NewEquipment creates a new Equipment.
func NewEquipment(entity *Entity, parent *Equipment, container bool) *Equipment {
	e := Equipment{
		EquipmentData: EquipmentData{
			ContainerBase: newContainerBase[*Equipment](equipmentTypeKey, container),
			EquipmentEditData: EquipmentEditData{
				LegalityClass: "4",
				Quantity:      fxp.One,
				Equipped:      true,
			},
		},
		Entity: entity,
	}
	e.Name = e.Kind()
	e.parent = parent
	return &e
}

// Clone implements Node.
func (e *Equipment) Clone(entity *Entity, parent *Equipment, preserveID bool) *Equipment {
	other := NewEquipment(entity, parent, e.Container())
	if preserveID {
		other.ID = e.ID
	}
	other.IsOpen = e.IsOpen
	other.EquipmentEditData.CopyFrom(e)
	if e.HasChildren() {
		other.Children = make([]*Equipment, 0, len(e.Children))
		for _, child := range e.Children {
			other.Children = append(other.Children, child.Clone(entity, other, preserveID))
		}
	}
	return other
}

// MarshalJSON implements json.Marshaler.
func (e *Equipment) MarshalJSON() ([]byte, error) {
	type calc struct {
		ExtendedValue           fxp.Int         `json:"extended_value"`
		ExtendedWeight          measure.Weight  `json:"extended_weight"`
		ExtendedWeightForSkills *measure.Weight `json:"extended_weight_for_skills,omitempty"`
	}
	e.ClearUnusedFieldsForType()
	defUnits := SheetSettingsFor(e.Entity).DefaultWeightUnits
	data := struct {
		EquipmentData
		Calc calc `json:"calc"`
	}{
		EquipmentData: e.EquipmentData,
		Calc: calc{
			ExtendedValue:           e.ExtendedValue(),
			ExtendedWeight:          e.ExtendedWeight(false, defUnits),
			ExtendedWeightForSkills: nil,
		},
	}
	if e.WeightIgnoredForSkills {
		w := e.ExtendedWeight(true, defUnits)
		data.Calc.ExtendedWeightForSkills = &w
	}
	return json.Marshal(&data)
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *Equipment) UnmarshalJSON(data []byte) error {
	var localData struct {
		EquipmentData
		// Old data fields
		Categories []string `json:"categories"`
	}
	if err := json.Unmarshal(data, &localData); err != nil {
		return err
	}
	localData.ClearUnusedFieldsForType()
	e.EquipmentData = localData.EquipmentData
	e.Tags = convertOldCategoriesToTags(e.Tags, localData.Categories)
	slices.Sort(e.Tags)
	if e.Container() {
		if e.Quantity == 0 {
			// Old formats omitted the quantity for containers. Try to see if it was omitted or if it was explicitly
			// set to zero.
			m := make(map[string]any)
			if err := json.Unmarshal(data, &m); err == nil {
				if _, exists := m["quantity"]; !exists {
					e.Quantity = fxp.One
				}
			}
		}
		for _, one := range e.Children {
			one.parent = e
		}
	}
	return nil
}

// CellData returns the cell data information for the given column.
func (e *Equipment) CellData(column int, data *CellData) {
	switch column {
	case EquipmentEquippedColumn:
		data.Type = Toggle
		data.Checked = e.Equipped
		data.Alignment = unison.MiddleAlignment
	case EquipmentQuantityColumn:
		data.Type = Text
		data.Primary = e.Quantity.String()
		data.Alignment = unison.EndAlignment
	case EquipmentDescriptionColumn:
		data.Type = Text
		data.Primary = e.Description()
		data.Secondary = e.SecondaryText(func(option display.Option) bool { return option.Inline() })
		data.UnsatisfiedReason = e.UnsatisfiedReason
		data.Tooltip = e.SecondaryText(func(option display.Option) bool { return option.Tooltip() })
	case EquipmentUsesColumn:
		if e.MaxUses > 0 {
			data.Type = Text
			data.Primary = strconv.Itoa(e.Uses)
			data.Alignment = unison.EndAlignment
		}
	case EquipmentMaxUsesColumn:
		if e.MaxUses > 0 {
			data.Type = Text
			data.Primary = strconv.Itoa(e.MaxUses)
			data.Alignment = unison.EndAlignment
		}
	case EquipmentTLColumn:
		data.Type = Text
		data.Primary = e.TechLevel
		data.Alignment = unison.EndAlignment
	case EquipmentLCColumn:
		data.Type = Text
		data.Primary = e.LegalityClass
		data.Alignment = unison.EndAlignment
	case EquipmentCostColumn:
		data.Type = Text
		data.Primary = e.AdjustedValue().String()
		data.Alignment = unison.EndAlignment
	case EquipmentExtendedCostColumn:
		data.Type = Text
		data.Primary = e.ExtendedValue().String()
		data.Alignment = unison.EndAlignment
	case EquipmentWeightColumn:
		data.Type = Text
		units := SheetSettingsFor(e.Entity).DefaultWeightUnits
		data.Primary = units.Format(e.AdjustedWeight(false, units))
		data.Alignment = unison.EndAlignment
	case EquipmentExtendedWeightColumn:
		data.Type = Text
		units := SheetSettingsFor(e.Entity).DefaultWeightUnits
		data.Primary = units.Format(e.ExtendedWeight(false, units))
		data.Alignment = unison.EndAlignment
	case EquipmentTagsColumn:
		data.Type = Tags
		data.Primary = CombineTags(e.Tags)
	case EquipmentReferenceColumn, PageRefCellAlias:
		data.Type = PageRef
		data.Primary = e.PageRef
		data.Secondary = e.Name
	}
}

// Depth returns the number of parents this node has.
func (e *Equipment) Depth() int {
	count := 0
	p := e.parent
	for p != nil {
		count++
		p = p.parent
	}
	return count
}

// OwningEntity returns the owning Entity.
func (e *Equipment) OwningEntity() *Entity {
	return e.Entity
}

// SetOwningEntity sets the owning entity and configures any sub-components as needed.
func (e *Equipment) SetOwningEntity(entity *Entity) {
	e.Entity = entity
	for _, w := range e.Weapons {
		w.SetOwner(e)
	}
	if e.Container() {
		for _, child := range e.Children {
			child.SetOwningEntity(entity)
		}
	}
	for _, m := range e.Modifiers {
		m.SetOwningEntity(entity)
	}
}

// Description returns a description.
func (e *Equipment) Description() string {
	return e.Name
}

// SecondaryText returns the "secondary" text: the text display below the description.
func (e *Equipment) SecondaryText(optionChecker func(display.Option) bool) string {
	var buffer strings.Builder
	settings := SheetSettingsFor(e.Entity)
	if optionChecker(settings.ModifiersDisplay) {
		if notes := e.ModifierNotes(); notes != "" {
			if buffer.Len() != 0 {
				buffer.WriteByte('\n')
			}
			buffer.WriteString(notes)
		}
	}
	if e.LocalNotes != "" && optionChecker(settings.NotesDisplay) {
		if buffer.Len() != 0 {
			buffer.WriteByte('\n')
		}
		buffer.WriteString(e.LocalNotes)
	}
	return buffer.String()
}

// String implements fmt.Stringer.
func (e *Equipment) String() string {
	return e.Name
}

// Notes returns the local notes.
func (e *Equipment) Notes() string {
	return e.LocalNotes
}

// FeatureList returns the list of Features.
func (e *Equipment) FeatureList() Features {
	return e.Features
}

// TagList returns the list of tags.
func (e *Equipment) TagList() []string {
	return e.Tags
}

// AdjustedValue returns the value after adjustments for any modifiers. Does not include the value of children.
func (e *Equipment) AdjustedValue() fxp.Int {
	return ValueAdjustedForModifiers(e.Value, e.Modifiers)
}

// ExtendedValue returns the extended value.
func (e *Equipment) ExtendedValue() fxp.Int {
	if e.Quantity <= 0 {
		return 0
	}
	value := e.AdjustedValue()
	if e.Container() {
		for _, one := range e.Children {
			value += one.ExtendedValue()
		}
	}
	return value.Mul(e.Quantity)
}

// AdjustedWeight returns the weight after adjustments for any modifiers. Does not include the weight of children.
func (e *Equipment) AdjustedWeight(forSkills bool, defUnits measure.WeightUnits) measure.Weight {
	if forSkills && e.WeightIgnoredForSkills {
		return 0
	}
	return WeightAdjustedForModifiers(e.Weight, e.Modifiers, defUnits)
}

// ExtendedWeight returns the extended weight.
func (e *Equipment) ExtendedWeight(forSkills bool, defUnits measure.WeightUnits) measure.Weight {
	return ExtendedWeightAdjustedForModifiers(defUnits, e.Quantity, e.Weight, e.Modifiers, e.Features, e.Children, forSkills, e.WeightIgnoredForSkills)
}

// ExtendedWeightAdjustedForModifiers calculates the extended weight.
func ExtendedWeightAdjustedForModifiers(defUnits measure.WeightUnits, qty fxp.Int, baseWeight measure.Weight, modifiers []*EquipmentModifier, features Features, children []*Equipment, forSkills, weightIgnoredForSkills bool) measure.Weight {
	if qty <= 0 {
		return 0
	}
	var base fxp.Int
	if !forSkills || !weightIgnoredForSkills {
		base = fxp.Int(WeightAdjustedForModifiers(baseWeight, modifiers, defUnits))
	}
	if len(children) != 0 {
		var contained fxp.Int
		for _, one := range children {
			contained += fxp.Int(one.ExtendedWeight(forSkills, defUnits))
		}
		var percentage, reduction fxp.Int
		for _, one := range features {
			if cwr, ok := one.(*ContainedWeightReduction); ok {
				if cwr.IsPercentageReduction() {
					percentage += cwr.PercentageReduction()
				} else {
					reduction += fxp.Int(cwr.FixedReduction(defUnits))
				}
			}
		}
		Traverse(func(mod *EquipmentModifier) bool {
			for _, f := range mod.Features {
				if cwr, ok := f.(*ContainedWeightReduction); ok {
					if cwr.IsPercentageReduction() {
						percentage += cwr.PercentageReduction()
					} else {
						reduction += fxp.Int(cwr.FixedReduction(defUnits))
					}
				}
			}
			return false
		}, true, true, modifiers...)
		if percentage >= fxp.Hundred {
			contained = 0
		} else if percentage > 0 {
			contained -= contained.Mul(percentage).Div(fxp.Hundred)
		}
		base += (contained - reduction).Max(0)
	}
	return measure.Weight(base.Mul(qty))
}

// FillWithNameableKeys adds any nameable keys found to the provided map.
func (e *Equipment) FillWithNameableKeys(m map[string]string) {
	Extract(e.Name, m)
	Extract(e.LocalNotes, m)
	if e.Prereq != nil {
		e.Prereq.FillWithNameableKeys(m)
	}
	for _, one := range e.Features {
		one.FillWithNameableKeys(m)
	}
	for _, one := range e.Weapons {
		one.FillWithNameableKeys(m)
	}
	Traverse(func(mod *EquipmentModifier) bool {
		mod.FillWithNameableKeys(m)
		return false
	}, true, true, e.Modifiers...)
}

// ApplyNameableKeys replaces any nameable keys found with the corresponding values in the provided map.
func (e *Equipment) ApplyNameableKeys(m map[string]string) {
	e.Name = Apply(e.Name, m)
	e.LocalNotes = Apply(e.LocalNotes, m)
	if e.Prereq != nil {
		e.Prereq.ApplyNameableKeys(m)
	}
	for _, one := range e.Features {
		one.ApplyNameableKeys(m)
	}
	for _, one := range e.Weapons {
		one.ApplyNameableKeys(m)
	}
	Traverse(func(mod *EquipmentModifier) bool {
		mod.ApplyNameableKeys(m)
		return false
	}, true, true, e.Modifiers...)
}

// DisplayLegalityClass returns a display version of the LegalityClass.
func (e *Equipment) DisplayLegalityClass() string {
	lc := strings.TrimSpace(e.LegalityClass)
	switch lc {
	case "0":
		return i18n.Text("LC0: Banned")
	case "1":
		return i18n.Text("LC1: Military")
	case "2":
		return i18n.Text("LC2: Restricted")
	case "3":
		return i18n.Text("LC3: Licensed")
	case "4":
		return i18n.Text("LC4: Open")
	default:
		return lc
	}
}

// ActiveModifierFor returns the first modifier that matches the name (case-insensitive).
func (e *Equipment) ActiveModifierFor(name string) *EquipmentModifier {
	var found *EquipmentModifier
	Traverse(func(mod *EquipmentModifier) bool {
		if strings.EqualFold(mod.Name, name) {
			found = mod
			return true
		}
		return false
	}, true, true, e.Modifiers...)
	return found
}

// ModifierNotes returns the notes due to modifiers.
func (e *Equipment) ModifierNotes() string {
	var buffer strings.Builder
	Traverse(func(mod *EquipmentModifier) bool {
		if buffer.Len() != 0 {
			buffer.WriteString("; ")
		}
		buffer.WriteString(mod.FullDescription())
		return false
	}, true, true, e.Modifiers...)
	return buffer.String()
}

// TL implements TechLevelProvider.
func (e *Equipment) TL() string {
	return e.TechLevel
}

// RequiresTL implements TechLevelProvider.
func (e *Equipment) RequiresTL() bool {
	return true
}

// SetTL implements TechLevelProvider.
func (e *Equipment) SetTL(tl string) {
	e.TechLevel = tl
}

// Enabled returns true if this node is enabled.
func (e *Equipment) Enabled() bool {
	return true
}
