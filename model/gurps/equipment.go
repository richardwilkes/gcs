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
	"fmt"
	"io/fs"
	"slices"
	"strconv"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/tid"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison/enums/align"
)

var (
	_ WeaponOwner                   = &Equipment{}
	_ Node[*Equipment]              = &Equipment{}
	_ TechLevelProvider[*Equipment] = &Equipment{}
	_ EditorData[*Equipment]        = &Equipment{}
)

// Columns that can be used with the equipment method .CellData()
const (
	EquipmentEquippedColumn = iota
	EquipmentQuantityColumn
	EquipmentDescriptionColumn
	EquipmentUsesColumn
	EquipmentTLColumn
	EquipmentLCColumn
	EquipmentCostColumn
	EquipmentExtendedCostColumn
	EquipmentWeightColumn
	EquipmentExtendedWeightColumn
	EquipmentTagsColumn
	EquipmentReferenceColumn
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

// EquipmentData holds the Equipment data that is written to disk.
type EquipmentData struct {
	ContainerBase[*Equipment]
	EquipmentEditData
}

// EquipmentEditData holds the Equipment data that can be edited by the UI detail editor.
type EquipmentEditData struct {
	Name                   string               `json:"description,omitempty"`
	PageRef                string               `json:"reference,omitempty"`
	PageRefHighlight       string               `json:"reference_highlight,omitempty"`
	LocalNotes             string               `json:"notes,omitempty"`
	VTTNotes               string               `json:"vtt_notes,omitempty"`
	TechLevel              string               `json:"tech_level,omitempty"`
	LegalityClass          string               `json:"legality_class,omitempty"`
	Tags                   []string             `json:"tags,omitempty"`
	Modifiers              []*EquipmentModifier `json:"modifiers,omitempty"`
	RatedST                fxp.Int              `json:"rated_strength,omitempty"`
	Quantity               fxp.Int              `json:"quantity,omitempty"`
	Value                  fxp.Int              `json:"value,omitempty"`
	Weight                 fxp.Weight           `json:"weight,omitempty"`
	MaxUses                int                  `json:"max_uses,omitempty"`
	Uses                   int                  `json:"uses,omitempty"`
	Prereq                 *PrereqList          `json:"prereqs,omitempty"`
	Weapons                []*Weapon            `json:"weapons,omitempty"`
	Features               Features             `json:"features,omitempty"`
	Equipped               bool                 `json:"equipped,omitempty"`
	WeightIgnoredForSkills bool                 `json:"ignore_weight_for_skills,omitempty"`
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
		return nil, errs.NewWithCause(InvalidFileDataMsg(), err)
	}
	if data.Type != equipmentListTypeKey {
		return nil, errs.New(UnexpectedFileDataMsg())
	}
	if err := CheckVersion(data.Version); err != nil {
		return nil, err
	}
	return data.Rows, nil
}

// SaveEquipment writes the Equipment list to the file as JSON.
func SaveEquipment(equipment []*Equipment, filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, &equipmentListData{
		Type:    equipmentListTypeKey,
		Version: CurrentDataVersion,
		Rows:    equipment,
	})
}

// NewEquipment creates a new Equipment.
func NewEquipment(entity *Entity, parent *Equipment, container bool) *Equipment {
	e := Equipment{
		EquipmentData: EquipmentData{
			ContainerBase: newContainerBase(parent, KindEquipment, KindEquipmentContainer, container),
			EquipmentEditData: EquipmentEditData{
				LegalityClass: "4",
				Quantity:      fxp.One,
				Equipped:      true,
			},
		},
		Entity: entity,
	}
	e.Name = e.Kind()
	return &e
}

// Clone implements Node.
func (e *Equipment) Clone(entity *Entity, parent *Equipment, preserveID bool) *Equipment {
	other := NewEquipment(entity, parent, e.Container())
	other.CopyFrom(e)
	if preserveID {
		other.LocalID = e.LocalID
	}
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
		ExtendedValue           fxp.Int     `json:"extended_value"`
		ExtendedWeight          fxp.Weight  `json:"extended_weight"`
		ExtendedWeightForSkills *fxp.Weight `json:"extended_weight_for_skills,omitempty"`
		ResolvedNotes           string      `json:"resolved_notes,omitempty"`
		UnsatisfiedReason       string      `json:"unsatisfied_reason,omitempty"`
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
			UnsatisfiedReason:       e.UnsatisfiedReason,
		},
	}
	notes := e.resolveLocalNotes()
	if notes != e.LocalNotes {
		data.Calc.ResolvedNotes = notes
	}
	if e.WeightIgnoredForSkills && e.Equipped {
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
		Type       string   `json:"type"`
		Categories []string `json:"categories"`
	}
	if err := json.Unmarshal(data, &localData); err != nil {
		return err
	}
	localData.itemKind = KindEquipment
	localData.containerKind = KindEquipmentContainer
	if !tid.IsKindAndValid(localData.LocalID, KindEquipmentContainer) && !tid.IsKindAndValid(localData.LocalID, KindEquipment) {
		switch localData.Type {
		case "equipment":
			localData.LocalID = tid.MustNewTID(KindEquipment)
		case "equipment_container":
			localData.LocalID = tid.MustNewTID(KindEquipmentContainer)
		default:
			return errs.New("invalid data type")
		}
	}
	e.EquipmentData = localData.EquipmentData
	e.ClearUnusedFieldsForType()
	e.Tags = ConvertOldCategoriesToTags(e.Tags, localData.Categories)
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

// EquipmentHeaderData returns the header data information for the given equipment column.
func EquipmentHeaderData(columnID int, entity *Entity, carried, forPage bool) HeaderData {
	var data HeaderData
	switch columnID {
	case EquipmentEquippedColumn:
		data.Title = HeaderCheckmark
		data.TitleIsImageKey = true
		data.Detail = i18n.Text("Whether this piece of equipment is equipped or just carried. Items that are not equipped do not apply any features they may normally contribute to the character.")
	case EquipmentQuantityColumn:
		data.Title = i18n.Text("#")
		data.Detail = i18n.Text("Quantity")
	case EquipmentDescriptionColumn:
		data.Title = i18n.Text("Equipment")
		if forPage && entity != nil {
			if carried {
				data.Title = fmt.Sprintf(i18n.Text("Carried Equipment (%s; $%s)"),
					entity.SheetSettings.DefaultWeightUnits.Format(entity.WeightCarried(false)),
					entity.WealthCarried().Comma())
			} else {
				data.Title = fmt.Sprintf(i18n.Text("Other Equipment ($%s)"), entity.WealthNotCarried().Comma())
			}
		}
		data.Primary = true
	case EquipmentUsesColumn:
		data.Title = i18n.Text("Uses")
		data.Detail = i18n.Text("The number of uses remaining")
	case EquipmentTLColumn:
		data.Title = i18n.Text("TL")
		data.Detail = i18n.Text("Tech Level")
	case EquipmentLCColumn:
		data.Title = i18n.Text("LC")
		data.Detail = i18n.Text("Legality Class")
	case EquipmentCostColumn:
		data.Title = HeaderCoins
		data.TitleIsImageKey = true
		data.Detail = i18n.Text("The value of one of these pieces of equipment")
	case EquipmentExtendedCostColumn:
		data.Title = HeaderStackedCoins
		data.TitleIsImageKey = true
		data.Detail = i18n.Text("The value of all of these pieces of equipment, plus the value of any contained equipment")
	case EquipmentWeightColumn:
		data.Title = HeaderWeight
		data.TitleIsImageKey = true
		data.Detail = i18n.Text("The weight of one of these pieces of equipment")
	case EquipmentExtendedWeightColumn:
		data.Title = HeaderStackedWeight
		data.TitleIsImageKey = true
		data.Detail = i18n.Text("The weight of all of these pieces of equipment, plus the weight of any contained equipment")
	case EquipmentTagsColumn:
		data.Title = i18n.Text("Tags")
	case EquipmentReferenceColumn:
		data.Title = HeaderBookmark
		data.TitleIsImageKey = true
		data.Detail = PageRefTooltipText()
	}
	return data
}

// CellData returns the cell data information for the given column.
func (e *Equipment) CellData(columnID int, data *CellData) {
	data.Dim = e.Quantity == 0
	e1 := e
	for !data.Dim && e1.Parent() != nil {
		e1 = e1.Parent()
		data.Dim = e1.Quantity == 0
	}
	switch columnID {
	case EquipmentEquippedColumn:
		data.Type = cell.Toggle
		data.Checked = e.Equipped
		data.Alignment = align.Middle
	case EquipmentQuantityColumn:
		data.Type = cell.Text
		data.Primary = e.Quantity.Comma()
		data.Alignment = align.End
	case EquipmentDescriptionColumn:
		data.Type = cell.Text
		data.Primary = e.Description()
		data.Secondary = e.SecondaryText(func(option display.Option) bool { return option.Inline() })
		data.UnsatisfiedReason = e.UnsatisfiedReason
		data.Tooltip = e.SecondaryText(func(option display.Option) bool { return option.Tooltip() })
	case EquipmentUsesColumn:
		if e.MaxUses > 0 {
			data.Type = cell.Text
			data.Primary = strconv.Itoa(e.Uses)
			data.Alignment = align.End
			data.Tooltip = fmt.Sprintf(i18n.Text("Maximum Uses: %d"), e.MaxUses)
		}
	case EquipmentTLColumn:
		data.Type = cell.Text
		data.Primary = e.TechLevel
		data.Alignment = align.End
	case EquipmentLCColumn:
		data.Type = cell.Text
		data.Primary = e.LegalityClass
		data.Alignment = align.End
	case EquipmentCostColumn:
		data.Type = cell.Text
		data.Primary = e.AdjustedValue().Comma()
		data.Alignment = align.End
	case EquipmentExtendedCostColumn:
		data.Type = cell.Text
		data.Primary = e.ExtendedValue().Comma()
		data.Alignment = align.End
	case EquipmentWeightColumn:
		data.Type = cell.Text
		units := SheetSettingsFor(e.Entity).DefaultWeightUnits
		data.Primary = units.Format(e.AdjustedWeight(false, units))
		data.Alignment = align.End
	case EquipmentExtendedWeightColumn:
		data.Type = cell.Text
		units := SheetSettingsFor(e.Entity).DefaultWeightUnits
		data.Primary = units.Format(e.ExtendedWeight(false, units))
		data.Alignment = align.End
	case EquipmentTagsColumn:
		data.Type = cell.Tags
		data.Primary = CombineTags(e.Tags)
	case EquipmentReferenceColumn, PageRefCellAlias:
		data.Type = cell.PageRef
		data.Primary = e.PageRef
		if e.PageRefHighlight != "" {
			data.Secondary = e.PageRefHighlight
		} else {
			data.Secondary = e.Name
		}
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
		AppendStringOntoNewLine(&buffer, e.ModifierNotes())
	}
	if optionChecker(settings.NotesDisplay) {
		var localBuffer strings.Builder
		if e.RatedST != 0 {
			localBuffer.WriteString(i18n.Text("Rated ST "))
			localBuffer.WriteString(e.RatedST.String())
		}
		if localNotes := e.resolveLocalNotes(); localNotes != "" {
			if localBuffer.Len() != 0 {
				localBuffer.WriteString("; ")
			}
			localBuffer.WriteString(localNotes)
		}
		AppendBufferOntoNewLine(&buffer, &localBuffer)
	}
	return buffer.String()
}

// String implements fmt.Stringer.
func (e *Equipment) String() string {
	return e.Name
}

func (e *Equipment) resolveLocalNotes() string {
	return EvalEmbeddedRegex.ReplaceAllStringFunc(e.LocalNotes, e.Entity.EmbeddedEval)
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

// RatedStrength always return 0 for traits.
func (e *Equipment) RatedStrength() fxp.Int {
	return e.RatedST
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
func (e *Equipment) AdjustedWeight(forSkills bool, defUnits fxp.WeightUnit) fxp.Weight {
	if forSkills && e.WeightIgnoredForSkills && e.Equipped {
		return 0
	}
	return WeightAdjustedForModifiers(e.Weight, e.Modifiers, defUnits)
}

// ExtendedWeight returns the extended weight.
func (e *Equipment) ExtendedWeight(forSkills bool, defUnits fxp.WeightUnit) fxp.Weight {
	return ExtendedWeightAdjustedForModifiers(defUnits, e.Quantity, e.Weight, e.Modifiers, e.Features, e.Children, forSkills, e.WeightIgnoredForSkills && e.Equipped)
}

// ExtendedWeightAdjustedForModifiers calculates the extended weight.
func ExtendedWeightAdjustedForModifiers(defUnits fxp.WeightUnit, qty fxp.Int, baseWeight fxp.Weight, modifiers []*EquipmentModifier, features Features, children []*Equipment, forSkills, weightIgnoredForSkills bool) fxp.Weight {
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
	return fxp.Weight(base.Mul(qty))
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

// Kind returns the kind of data.
func (e *Equipment) Kind() string {
	return e.kind(i18n.Text("Equipment"))
}

// ClearUnusedFieldsForType zeroes out the fields that are not applicable to this type (container vs not-container).
func (e *Equipment) ClearUnusedFieldsForType() {
	e.clearUnusedFields()
}

// CopyFrom implements node.EditorData.
func (e *Equipment) CopyFrom(other *Equipment) {
	e.copyFrom(other.Entity, other, false)
	e.LocalID = tid.MustNewTID(e.LocalID[0])
}

// ApplyTo implements node.EditorData.
func (e *Equipment) ApplyTo(other *Equipment) {
	id := other.LocalID
	other.copyFrom(other.Entity, e, true)
	other.LocalID = id
}

func (e *Equipment) copyFrom(entity *Entity, other *Equipment, isApply bool) {
	e.EquipmentData = other.EquipmentData
	e.Tags = txt.CloneStringSlice(e.Tags)
	e.Modifiers = nil
	if len(other.Modifiers) != 0 {
		e.Modifiers = make([]*EquipmentModifier, 0, len(other.Modifiers))
		for _, one := range other.Modifiers {
			e.Modifiers = append(e.Modifiers, one.Clone(entity, nil, true))
		}
	}
	e.Prereq = e.Prereq.CloneResolvingEmpty(false, isApply)
	e.Weapons = nil
	if len(other.Weapons) != 0 {
		e.Weapons = make([]*Weapon, 0, len(other.Weapons))
		for _, one := range other.Weapons {
			e.Weapons = append(e.Weapons, one.Clone(entity, nil, true))
		}
	}
	e.Features = other.Features.Clone()
}
