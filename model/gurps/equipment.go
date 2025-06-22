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
	"fmt"
	"hash"
	"io/fs"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
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
	_ WeaponOwner                   = &Equipment{}
	_ LeveledOwner                  = &Equipment{}
	_ Node[*Equipment]              = &Equipment{}
	_ TechLevelProvider[*Equipment] = &Equipment{}
	_ EditorData[*Equipment]        = &EquipmentEditData{}
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
	EquipmentLibSrcColumn
)

// Equipment holds a piece of equipment.
type Equipment struct {
	EquipmentData
	owner             DataOwner
	UnsatisfiedReason string
}

// EquipmentData holds the Equipment data that is written to disk.
type EquipmentData struct {
	SourcedID
	EquipmentEditData
	ThirdParty map[string]any `json:"third_party,omitempty"`
	Children   []*Equipment   `json:"children,omitempty"` // Only for containers
	parent     *Equipment
}

// EquipmentEditData holds the Equipment data that can be edited by the UI detail editor.
type EquipmentEditData struct {
	EquipmentSyncData
	VTTNotes     string               `json:"vtt_notes,omitempty"`
	Replacements map[string]string    `json:"replacements,omitempty"`
	Modifiers    []*EquipmentModifier `json:"modifiers,omitempty"`
	RatedST      fxp.Int              `json:"rated_strength,omitempty"`
	Quantity     fxp.Int              `json:"quantity,omitempty"`
	Level        fxp.Int              `json:"level,omitempty"`
	Uses         int                  `json:"uses,omitempty"`
	Equipped     bool                 `json:"equipped,omitempty"`
}

// EquipmentSyncData holds the equipment sync data that is common to both containers and non-containers.
type EquipmentSyncData struct {
	Name                   string      `json:"description,omitempty"`
	PageRef                string      `json:"reference,omitempty"`
	PageRefHighlight       string      `json:"reference_highlight,omitempty"`
	LocalNotes             string      `json:"local_notes,omitempty"`
	TechLevel              string      `json:"tech_level,omitempty"`
	LegalityClass          string      `json:"legality_class,omitempty"`
	Tags                   []string    `json:"tags,omitempty"`
	BaseValue              string      `json:"base_value,omitempty"`
	BaseWeight             string      `json:"base_weight,omitempty"`
	MaxUses                int         `json:"max_uses,omitempty"`
	Prereq                 *PrereqList `json:"prereqs,omitempty"`
	Weapons                []*Weapon   `json:"weapons,omitempty"`
	Features               Features    `json:"features,omitempty"`
	WeightIgnoredForSkills bool        `json:"ignore_weight_for_skills,omitempty"`
}

type equipmentListData struct {
	Version int          `json:"version"`
	Rows    []*Equipment `json:"rows"`
}

// NewEquipmentFromFile loads an Equipment list from a file.
func NewEquipmentFromFile(fileSystem fs.FS, filePath string) ([]*Equipment, error) {
	var data equipmentListData
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &data); err != nil {
		return nil, errs.NewWithCause(InvalidFileData(), err)
	}
	if err := jio.CheckVersion(data.Version); err != nil {
		return nil, err
	}
	return data.Rows, nil
}

// SaveEquipment writes the Equipment list to the file as JSON.
func SaveEquipment(equipment []*Equipment, filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, &equipmentListData{
		Version: jio.CurrentDataVersion,
		Rows:    equipment,
	})
}

// NewEquipment creates a new Equipment.
func NewEquipment(owner DataOwner, parent *Equipment, container bool) *Equipment {
	var e Equipment
	e.TID = tid.MustNewTID(equipmentKind(container))
	e.Name = e.Kind()
	e.LegalityClass = "4"
	e.Quantity = fxp.One
	e.Equipped = true
	e.parent = parent
	e.owner = owner
	e.SetOpen(container)
	return &e
}

func equipmentKind(container bool) byte {
	if container {
		return kinds.EquipmentContainer
	}
	return kinds.Equipment
}

// ID returns the local ID of this data.
func (e *Equipment) ID() tid.TID {
	return e.TID
}

// Container returns true if this is a container.
func (e *Equipment) Container() bool {
	return tid.IsKind(e.TID, kinds.EquipmentContainer)
}

// HasChildren returns true if this node has children.
func (e *Equipment) HasChildren() bool {
	return e.Container() && len(e.Children) > 0
}

// NodeChildren returns the children of this node, if any.
func (e *Equipment) NodeChildren() []*Equipment {
	return e.Children
}

// SetChildren sets the children of this node.
func (e *Equipment) SetChildren(children []*Equipment) {
	e.Children = children
}

// Parent returns the parent.
func (e *Equipment) Parent() *Equipment {
	return e.parent
}

// SetParent sets the parent.
func (e *Equipment) SetParent(parent *Equipment) {
	e.parent = parent
}

// IsOpen returns true if this node is currently open.
func (e *Equipment) IsOpen() bool {
	return IsNodeOpen(e)
}

// SetOpen sets the current open state for this node.
func (e *Equipment) SetOpen(open bool) {
	SetNodeOpen(e, open)
}

// Clone implements Node.
func (e *Equipment) Clone(from LibraryFile, owner DataOwner, parent *Equipment, preserveID bool) *Equipment {
	other := NewEquipment(owner, parent, e.Container())
	other.AdjustSource(from, e.SourcedID, preserveID)
	other.SetOpen(e.IsOpen())
	other.ThirdParty = e.ThirdParty
	other.CopyFrom(e)
	if e.HasChildren() {
		other.Children = make([]*Equipment, 0, len(e.Children))
		for _, child := range e.Children {
			other.Children = append(other.Children, child.Clone(from, owner, other, preserveID))
		}
	}
	return other
}

// MarshalJSON implements json.Marshaler.
func (e *Equipment) MarshalJSON() ([]byte, error) {
	type calc struct {
		Value                   fxp.Int     `json:"value"`
		ExtendedValue           fxp.Int     `json:"extended_value"`
		Weight                  fxp.Weight  `json:"weight"`
		ExtendedWeight          fxp.Weight  `json:"extended_weight"`
		ExtendedWeightForSkills *fxp.Weight `json:"extended_weight_for_skills,omitempty"`
		ResolvedNotes           string      `json:"resolved_notes,omitempty"`
		UnsatisfiedReason       string      `json:"unsatisfied_reason,omitempty"`
	}
	e.ClearUnusedFieldsForType()
	defUnits := SheetSettingsFor(EntityFromNode(e)).DefaultWeightUnits
	data := struct {
		EquipmentData
		Calc calc `json:"calc"`
	}{
		EquipmentData: e.EquipmentData,
		Calc: calc{
			Value:                   e.AdjustedValue(),
			ExtendedValue:           e.ExtendedValue(),
			Weight:                  e.AdjustedWeight(false, defUnits),
			ExtendedWeight:          e.ExtendedWeight(false, defUnits),
			ExtendedWeightForSkills: nil,
			UnsatisfiedReason:       e.UnsatisfiedReason,
		},
	}
	notes := e.ResolveLocalNotes()
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
		Type       string     `json:"type"`
		ExprNotes  string     `json:"notes"`
		Categories []string   `json:"categories"`
		Value      fxp.Int    `json:"value"`
		Weight     fxp.Weight `json:"weight"`
		IsOpen     bool       `json:"open"`
	}
	if err := json.Unmarshal(data, &localData); err != nil {
		return err
	}
	setOpen := false
	if !tid.IsValid(localData.TID) {
		// Fixup old data that used UUIDs instead of TIDs
		localData.TID = tid.MustNewTID(equipmentKind(strings.HasSuffix(localData.Type, containerKeyPostfix)))
		setOpen = localData.IsOpen
	}
	e.EquipmentData = localData.EquipmentData
	if e.BaseValue == "" && localData.Value != 0 {
		e.BaseValue = localData.Value.String()
	}
	if e.BaseWeight == "" && localData.Weight != 0 {
		e.BaseWeight = fxp.Pound.Format(localData.Weight)
	}
	if e.LocalNotes == "" && localData.ExprNotes != "" {
		e.LocalNotes = EmbeddedExprToScript(localData.ExprNotes)
	}
	e.ClearUnusedFieldsForType()
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
	if setOpen {
		SetNodeOpen(e, true)
	}
	return nil
}

// EquipmentHeaderData returns the header data information for the given equipment column.
func EquipmentHeaderData(columnID int, provider EquipmentListProvider, carried, forPage bool) HeaderData {
	var data HeaderData
	switch columnID {
	case EquipmentEquippedColumn:
		data.Title = HeaderCheckmark
		data.TitleIsImageKey = true
		data.Detail = i18n.Text("Whether this piece of equipment is equipped or just carried. Items that are not equipped do not apply any features they may normally contribute to the character.")
	case EquipmentQuantityColumn:
		data.Title = i18n.Text("#")
		data.Detail = i18n.Text("Quantity")
		data.Less = fxp.IntLessFromString
	case EquipmentDescriptionColumn:
		data.Title = i18n.Text("Equipment")
		if forPage {
			var weight fxp.Weight
			var value fxp.Int
			units := provider.DataOwner().WeightUnit()
			if carried {
				title := i18n.Text("Carried Equipment")
				if _, ok := provider.(*Template); ok {
					title = i18n.Text("Equipment")
				}
				for _, one := range provider.CarriedEquipmentList() {
					weight += one.ExtendedWeight(false, units)
					value += one.ExtendedValue()
				}
				data.Title = fmt.Sprintf(i18n.Text("%s (%s; $%s)"), title, units.Format(weight), value.Comma())
			} else {
				title := i18n.Text("Other Equipment")
				if _, ok := provider.(*Loot); ok {
					title = i18n.Text("Equipment")
				}
				for _, one := range provider.OtherEquipmentList() {
					weight += one.ExtendedWeight(false, units)
					value += one.ExtendedValue()
				}
				data.Title = fmt.Sprintf(i18n.Text("%s (%s; $%s)"), title, units.Format(weight), value.Comma())
			}
		}
		data.Primary = true
	case EquipmentUsesColumn:
		data.Title = i18n.Text("Uses")
		data.Detail = i18n.Text("The number of uses remaining")
		data.Less = fxp.IntLessFromString
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
		data.Less = fxp.IntLessFromString
	case EquipmentExtendedCostColumn:
		data.Title = HeaderStackedCoins
		data.TitleIsImageKey = true
		data.Detail = i18n.Text("The value of all of these pieces of equipment, plus the value of any contained equipment")
		data.Less = fxp.IntLessFromString
	case EquipmentWeightColumn:
		data.Title = HeaderWeight
		data.TitleIsImageKey = true
		data.Detail = i18n.Text("The weight of one of these pieces of equipment")
		data.Less = fxp.WeightLessFromStringFunc(provider.DataOwner().WeightUnit())
	case EquipmentExtendedWeightColumn:
		data.Title = HeaderStackedWeight
		data.TitleIsImageKey = true
		data.Detail = i18n.Text("The weight of all of these pieces of equipment, plus the weight of any contained equipment")
		data.Less = fxp.WeightLessFromStringFunc(provider.DataOwner().WeightUnit())
	case EquipmentTagsColumn:
		data.Title = i18n.Text("Tags")
	case EquipmentReferenceColumn:
		data.Title = HeaderBookmark
		data.TitleIsImageKey = true
		data.Detail = PageRefTooltip()
	case EquipmentLibSrcColumn:
		data.Title = HeaderDatabase
		data.TitleIsImageKey = true
		data.Detail = LibSrcTooltip()
	}
	return data
}

// CellData returns the cell data information for the given column.
func (e *Equipment) CellData(columnID int, data *CellData) {
	data.Dim = e.Quantity == 0
	e1 := e
	for !data.Dim && e1.parent != nil {
		e1 = e1.parent
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
		data.Primary = e.String()
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
		units := SheetSettingsFor(EntityFromNode(e)).DefaultWeightUnits
		data.Primary = units.Format(e.AdjustedWeight(false, units))
		data.Alignment = align.End
	case EquipmentExtendedWeightColumn:
		data.Type = cell.Text
		units := SheetSettingsFor(EntityFromNode(e)).DefaultWeightUnits
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
			data.Secondary = e.String()
		}
	case EquipmentLibSrcColumn:
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
func (e *Equipment) Depth() int {
	count := 0
	p := e.parent
	for p != nil {
		count++
		p = p.parent
	}
	return count
}

// DataOwner returns the data owner.
func (e *Equipment) DataOwner() DataOwner {
	return e.owner
}

// SetDataOwner sets the data owner and configures any sub-components as needed.
func (e *Equipment) SetDataOwner(owner DataOwner) {
	e.owner = owner
	for _, w := range e.Weapons {
		w.SetOwner(e)
	}
	if e.Container() {
		for _, child := range e.Children {
			child.SetDataOwner(owner)
		}
	}
	for _, m := range e.Modifiers {
		m.setEquipment(e)
		m.SetDataOwner(owner)
	}
}

// IsLeveled returns true if the equipment is capable of having levels.
func (e *Equipment) IsLeveled() bool {
	return e.Level > 0
}

// CurrentLevel returns the current level of the equipment or zero if it is not leveled.
func (e *Equipment) CurrentLevel() fxp.Int {
	if e.IsLeveled() {
		return e.Level
	}
	return 0
}

// Description returns a description, which doesn't include any levels.
func (e *Equipment) Description() string {
	return e.NameWithReplacements()
}

// SecondaryText returns the "secondary" text: the text display below the description.
func (e *Equipment) SecondaryText(optionChecker func(display.Option) bool) string {
	var buffer strings.Builder
	settings := SheetSettingsFor(EntityFromNode(e))
	if optionChecker(settings.ModifiersDisplay) {
		AppendStringOntoNewLine(&buffer, e.ModifierNotes())
	}
	if optionChecker(settings.NotesDisplay) {
		var localBuffer strings.Builder
		if e.RatedST != 0 {
			localBuffer.WriteString(i18n.Text("Rated ST "))
			localBuffer.WriteString(e.RatedST.String())
		}
		if localNotes := e.ResolveLocalNotes(); localNotes != "" {
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
	var buffer strings.Builder
	buffer.WriteString(e.Description())
	if e.IsLeveled() {
		buffer.WriteByte(' ')
		buffer.WriteString(e.Level.String())
	}
	return buffer.String()
}

// ResolveLocalNotes resolves the local notes, running any embedded scripts to get the final result.
func (e *Equipment) ResolveLocalNotes() string {
	return ResolveText(EntityFromNode(e), DeferredNewScriptEquipment(e), e.LocalNotesWithReplacements())
}

// Notes returns the local notes.
func (e *Equipment) Notes() string {
	return e.LocalNotesWithReplacements()
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

// BaseValueWithReplacements returns the base value with any replacements applied.
func (e *Equipment) BaseValueWithReplacements() string {
	return nameable.Apply(e.BaseValue, e.Replacements)
}

// ResolvedBaseValue resolves the base value, running any embedded scripts to get the final result.
func (e *Equipment) ResolvedBaseValue() fxp.Int {
	return ResolveToNumber(EntityFromNode(e), DeferredNewScriptEquipment(e), e.BaseValueWithReplacements()).
		Min(fxp.Max - 1).Max(0)
}

// AdjustedValue returns the value after adjustments for any modifiers. Does not include the value of children.
func (e *Equipment) AdjustedValue() fxp.Int {
	return ValueAdjustedForModifiers(e, e.ResolvedBaseValue(), e.Modifiers)
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

// ExtendedValueOfJustOne returns the extended value of just one piece of this equipment, including the value of
// children.
func (e *Equipment) ExtendedValueOfJustOne() fxp.Int {
	if e.Quantity <= 0 {
		return 0
	}
	value := e.AdjustedValue()
	if e.Container() {
		for _, one := range e.Children {
			value += one.ExtendedValue()
		}
	}
	return value
}

// BaseWeightWithReplacements returns the base weight with any replacements applied.
func (e *Equipment) BaseWeightWithReplacements() string {
	return nameable.Apply(e.BaseWeight, e.Replacements)
}

// ResolvedBaseWeight resolves the base weight, running any embedded scripts to get the final result.
func (e *Equipment) ResolvedBaseWeight() fxp.Weight {
	entity := EntityFromNode(e)
	return ResolveToWeight(entity, DeferredNewScriptEquipment(e), e.BaseWeightWithReplacements(),
		SheetSettingsFor(entity).DefaultWeightUnits)
}

// AdjustedWeight returns the weight after adjustments for any modifiers. Does not include the weight of children.
func (e *Equipment) AdjustedWeight(forSkills bool, defUnits fxp.WeightUnit) fxp.Weight {
	if forSkills && e.WeightIgnoredForSkills && e.Equipped {
		return 0
	}
	return WeightAdjustedForModifiers(e, e.ResolvedBaseWeight(), e.Modifiers, defUnits)
}

// ExtendedWeight returns the extended weight.
func (e *Equipment) ExtendedWeight(forSkills bool, defUnits fxp.WeightUnit) fxp.Weight {
	return ExtendedWeightAdjustedForModifiers(e, defUnits, e.Quantity, e.ResolvedBaseWeight(), e.Modifiers, e.Features, e.Children, forSkills, e.WeightIgnoredForSkills && e.Equipped)
}

// ExtendedWeightAdjustedForModifiers calculates the extended weight.
func ExtendedWeightAdjustedForModifiers(equipment *Equipment, defUnits fxp.WeightUnit, qty fxp.Int, baseWeight fxp.Weight, modifiers []*EquipmentModifier, features Features, children []*Equipment, forSkills, weightIgnoredForSkills bool) fxp.Weight {
	if qty <= 0 {
		return 0
	}
	var base fxp.Int
	if !forSkills || !weightIgnoredForSkills {
		base = fxp.Int(WeightAdjustedForModifiers(equipment, baseWeight, modifiers, defUnits))
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
			mod.setEquipment(equipment)
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

// NameableReplacements returns the replacements to be used with Nameables.
func (e *Equipment) NameableReplacements() map[string]string {
	if e == nil {
		return nil
	}
	return e.Replacements
}

// NameWithReplacements returns the name with any replacements applied.
func (e *Equipment) NameWithReplacements() string {
	return nameable.Apply(e.Name, e.Replacements)
}

// LocalNotesWithReplacements returns the local notes with any replacements applied.
func (e *Equipment) LocalNotesWithReplacements() string {
	return nameable.Apply(e.LocalNotes, e.Replacements)
}

// FillWithNameableKeys adds any nameable keys found to the provided map.
func (e *Equipment) FillWithNameableKeys(m, existing map[string]string) {
	if existing == nil {
		existing = e.Replacements
	}
	nameable.Extract(e.Name, m, existing)
	nameable.Extract(e.LocalNotes, m, existing)
	nameable.Extract(e.BaseValue, m, existing)
	nameable.Extract(e.BaseWeight, m, existing)
	if e.Prereq != nil {
		e.Prereq.FillWithNameableKeys(m, existing)
	}
	for _, one := range e.Features {
		one.FillWithNameableKeys(m, existing)
	}
	for _, one := range e.Weapons {
		one.FillWithNameableKeys(m, existing)
	}
	Traverse(func(mod *EquipmentModifier) bool {
		mod.FillWithNameableKeys(m, existing)
		return false
	}, true, true, e.Modifiers...)
}

// ApplyNameableKeys replaces any nameable keys found with the corresponding values in the provided map.
func (e *Equipment) ApplyNameableKeys(m map[string]string) {
	needed := make(map[string]string)
	e.FillWithNameableKeys(needed, nil)
	e.Replacements = nameable.Reduce(needed, m)
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
		if strings.EqualFold(mod.NameWithReplacements(), name) {
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

// CanConvertToFromContainer returns true if this node can be converted to/from a container.
func (e *Equipment) CanConvertToFromContainer() bool {
	return !e.Container() || !e.HasChildren()
}

// ConvertToContainer converts this node to a container.
func (e *Equipment) ConvertToContainer() {
	e.TID = tid.TID(kinds.EquipmentContainer) + e.TID[1:]
}

// ConvertToNonContainer converts this node to a non-container.
func (e *Equipment) ConvertToNonContainer() {
	e.TID = tid.TID(kinds.Equipment) + e.TID[1:]
}

// Kind returns the kind of data.
func (e *Equipment) Kind() string {
	if e.Container() {
		return i18n.Text("Equipment Container")
	}
	return i18n.Text("Equipment")
}

// ClearUnusedFieldsForType zeroes out the fields that are not applicable to this type (container vs not-container).
func (e *Equipment) ClearUnusedFieldsForType() {
	if !e.Container() {
		e.Children = nil
	}
}

// GetSource returns the source of this data.
func (e *Equipment) GetSource() Source {
	return e.Source
}

// ClearSource clears the source of this data.
func (e *Equipment) ClearSource() {
	e.Source = Source{}
}

// SyncWithSource synchronizes this data with the source.
func (e *Equipment) SyncWithSource() {
	if !toolbox.IsNil(e.owner) {
		if state, data := e.owner.SourceMatcher().Match(e); state == srcstate.Mismatched {
			if other, ok := data.(*Equipment); ok {
				e.EquipmentSyncData = other.EquipmentSyncData
				e.Tags = slices.Clone(other.Tags)
				e.Prereq = other.Prereq.CloneResolvingEmpty(false, true)
				e.Weapons = CloneWeapons(other.Weapons, false)
				e.Features = other.Features.Clone()
			}
		}
	}
}

// Hash writes this object's contents into the hasher. Note that this only hashes the data that is considered to be
// "source" data, i.e. not expected to be modified by the user after copying from a library.
func (e *Equipment) Hash(h hash.Hash) {
	e.hash(h)
}

func (e *EquipmentSyncData) hash(h hash.Hash) {
	hashhelper.String(h, e.Name)
	hashhelper.String(h, e.PageRef)
	hashhelper.String(h, e.PageRefHighlight)
	hashhelper.String(h, e.LocalNotes)
	hashhelper.String(h, e.TechLevel)
	hashhelper.String(h, e.LegalityClass)
	hashhelper.Num64(h, len(e.Tags))
	for _, tag := range e.Tags {
		hashhelper.String(h, tag)
	}
	hashhelper.String(h, e.BaseValue)
	hashhelper.String(h, e.BaseWeight)
	hashhelper.Num64(h, e.MaxUses)
	e.Prereq.Hash(h)
	hashhelper.Num64(h, len(e.Weapons))
	for _, weapon := range e.Weapons {
		weapon.Hash(h)
	}
	hashhelper.Num64(h, len(e.Features))
	for _, feature := range e.Features {
		feature.Hash(h)
	}
	hashhelper.Bool(h, e.WeightIgnoredForSkills)
}

// CopyFrom implements node.EditorData.
func (e *EquipmentEditData) CopyFrom(other *Equipment) {
	e.copyFrom(other.owner, &other.EquipmentEditData, false)
}

// ApplyTo implements node.EditorData.
func (e *EquipmentEditData) ApplyTo(other *Equipment) {
	other.copyFrom(other.owner, e, true)
}

func (e *EquipmentEditData) copyFrom(owner DataOwner, other *EquipmentEditData, isApply bool) {
	*e = *other
	e.Tags = txt.CloneStringSlice(other.Tags)
	e.Replacements = maps.Clone(other.Replacements)
	e.Modifiers = nil
	if len(other.Modifiers) != 0 {
		e.Modifiers = make([]*EquipmentModifier, 0, len(other.Modifiers))
		for _, one := range other.Modifiers {
			e.Modifiers = append(e.Modifiers, one.Clone(one.Source.LibraryFile, owner, nil, isApply))
		}
	}
	e.Prereq = e.Prereq.CloneResolvingEmpty(false, isApply)
	e.Weapons = CloneWeapons(other.Weapons, isApply)
	e.Features = other.Features.Clone()
}
