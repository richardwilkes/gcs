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
	"encoding/binary"
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
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/tid"
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
	EquipmentModifierLibSrcColumn
)

// EquipmentModifier holds a modifier to a piece of Equipment.
type EquipmentModifier struct {
	EquipmentModifierData
	owner DataOwner
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
	e := EquipmentModifier{
		EquipmentModifierData: EquipmentModifierData{
			SourcedID: SourcedID{
				TID: tid.MustNewTID(equipmentModifierKind(container)),
			},
			parent: parent,
		},
		owner: owner,
	}
	e.Name = e.Kind()
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
	other.EquipmentModifierEditData.CopyFrom(e)
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
	notes := e.resolveLocalNotes()
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
		data.Primary = e.Name
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
			data.Secondary = e.Name
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

// DataOwner returns the data owner.
func (e *EquipmentModifier) DataOwner() DataOwner {
	return e.owner
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
	return e.Name
}

func (e *EquipmentModifier) resolveLocalNotes() string {
	return EvalEmbeddedRegex.ReplaceAllStringFunc(e.LocalNotes, EntityFromNode(e).EmbeddedEval)
}

// SecondaryText returns the "secondary" text: the text display below an Trait.
func (e *EquipmentModifier) SecondaryText(optionChecker func(display.Option) bool) string {
	if !optionChecker(SheetSettingsFor(EntityFromNode(e)).NotesDisplay) {
		return ""
	}
	return e.resolveLocalNotes()
}

// FullDescription returns a full description.
func (e *EquipmentModifier) FullDescription() string {
	var buffer strings.Builder
	buffer.WriteString(e.String())
	if localNotes := e.resolveLocalNotes(); localNotes != "" {
		buffer.WriteString(" (")
		buffer.WriteString(e.LocalNotes)
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

// FillWithNameableKeys adds any nameable keys found in this EquipmentModifier to the provided map.
func (e *EquipmentModifier) FillWithNameableKeys(keyMap map[string]string) {
	if e.Enabled() {
		Extract(e.Name, keyMap)
		Extract(e.LocalNotes, keyMap)
		for _, one := range e.Features {
			one.FillWithNameableKeys(keyMap)
		}
	}
}

// ApplyNameableKeys replaces any nameable keys found in this EquipmentModifier with the corresponding values in the provided map.
func (e *EquipmentModifier) ApplyNameableKeys(keyMap map[string]string) {
	if e.Enabled() {
		e.Name = Apply(e.Name, keyMap)
		e.LocalNotes = Apply(e.LocalNotes, keyMap)
		for _, one := range e.Features {
			one.ApplyNameableKeys(keyMap)
		}
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
func (e *EquipmentModifier) Kind() string {
	if e.Container() {
		return i18n.Text("Equipment Modifier Container")
	}
	return i18n.Text("Equipment Modifier")
}

// ClearUnusedFieldsForType zeroes out the fields that are not applicable to this type (container vs not-container).
func (e *EquipmentModifier) ClearUnusedFieldsForType() {
	if e.Container() {
		e.CostType = 0
		e.WeightType = 0
		e.Disabled = false
		e.TechLevel = ""
		e.CostAmount = ""
		e.WeightAmount = ""
		e.Features = nil
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
				e.Name = other.Name
				e.PageRef = other.PageRef
				e.PageRefHighlight = other.PageRefHighlight
				e.LocalNotes = other.LocalNotes
				e.Tags = slices.Clone(other.Tags)
				if !e.Container() {
					e.CostType = other.CostType
					e.WeightType = other.WeightType
					e.TechLevel = other.TechLevel
					e.CostAmount = other.CostAmount
					e.WeightAmount = other.WeightAmount
					e.Features = other.Features.Clone()
				}
			}
		}
	}
}

// Hash writes this object's contents into the hasher. Note that this only hashes the data that is considered to be
// "source" data, i.e. not expected to be modified by the user after copying from a library.
func (e *EquipmentModifier) Hash(h hash.Hash) {
	_, _ = h.Write([]byte(e.Name))
	_, _ = h.Write([]byte(e.PageRef))
	_, _ = h.Write([]byte(e.PageRefHighlight))
	_, _ = h.Write([]byte(e.LocalNotes))
	for _, tag := range e.Tags {
		_, _ = h.Write([]byte(tag))
	}
	if !e.Container() {
		_ = binary.Write(h, binary.LittleEndian, e.CostType)
		_ = binary.Write(h, binary.LittleEndian, e.WeightType)
		_, _ = h.Write([]byte(e.TechLevel))
		_, _ = h.Write([]byte(e.CostAmount))
		_, _ = h.Write([]byte(e.WeightAmount))
		for _, feature := range e.Features {
			feature.Hash(h)
		}
	}
}

// CopyFrom implements node.EditorData.
func (e *EquipmentModifierEditData) CopyFrom(other *EquipmentModifier) {
	e.copyFrom(&other.EquipmentModifierEditData)
}

// ApplyTo implements node.EditorData.
func (e *EquipmentModifierEditData) ApplyTo(other *EquipmentModifier) {
	other.EquipmentModifierEditData.copyFrom(e)
}

func (e *EquipmentModifierEditData) copyFrom(other *EquipmentModifierEditData) {
	*e = *other
	e.Tags = txt.CloneStringSlice(e.Tags)
	e.Features = other.Features.Clone()
}
