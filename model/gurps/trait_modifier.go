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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/gid"
	"github.com/richardwilkes/gcs/v5/model/gurps/nameables"
	"github.com/richardwilkes/gcs/v5/model/gurps/trait"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/settings/display"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/unison"
	"golang.org/x/exp/slices"
)

var (
	_ Node[*TraitModifier] = &TraitModifier{}
	_ GeneralModifier      = &TraitModifier{}
)

// Columns that can be used with the trait modifier method .CellData()
const (
	TraitModifierEnabledColumn = iota
	TraitModifierDescriptionColumn
	TraitModifierCostColumn
	TraitModifierTagsColumn
	TraitModifierReferenceColumn
)

const (
	traitModifierListTypeKey = "modifier_list"
	traitModifierTypeKey     = "modifier"
)

// GeneralModifier is used for common access to modifiers.
type GeneralModifier interface {
	Container() bool
	Depth() int
	FullDescription() string
	Enabled() bool
	SetEnabled(enabled bool)
}

// TraitModifier holds a modifier to an Trait.
type TraitModifier struct {
	TraitModifierData
	Entity *Entity
}

type traitModifierListData struct {
	Type    string           `json:"type"`
	Version int              `json:"version"`
	Rows    []*TraitModifier `json:"rows"`
}

// NewTraitModifiersFromFile loads a TraitModifier list from a file.
func NewTraitModifiersFromFile(fileSystem fs.FS, filePath string) ([]*TraitModifier, error) {
	var data traitModifierListData
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &data); err != nil {
		return nil, errs.NewWithCause(gid.InvalidFileDataMsg, err)
	}
	if data.Type != traitModifierListTypeKey {
		return nil, errs.New(gid.UnexpectedFileDataMsg)
	}
	if err := gid.CheckVersion(data.Version); err != nil {
		return nil, err
	}
	return data.Rows, nil
}

// SaveTraitModifiers writes the TraitModifier list to the file as JSON.
func SaveTraitModifiers(modifiers []*TraitModifier, filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, &traitModifierListData{
		Type:    traitModifierListTypeKey,
		Version: gid.CurrentDataVersion,
		Rows:    modifiers,
	})
}

// NewTraitModifier creates a TraitModifier.
func NewTraitModifier(entity *Entity, parent *TraitModifier, container bool) *TraitModifier {
	a := &TraitModifier{
		TraitModifierData: TraitModifierData{
			ContainerBase: newContainerBase[*TraitModifier](traitModifierTypeKey, container),
		},
		Entity: entity,
	}
	a.Name = a.Kind()
	a.parent = parent
	return a
}

// Clone implements Node.
func (a *TraitModifier) Clone(entity *Entity, parent *TraitModifier, preserveID bool) *TraitModifier {
	other := NewTraitModifier(entity, parent, a.Container())
	if preserveID {
		other.ID = a.ID
	}
	other.IsOpen = a.IsOpen
	other.TraitModifierEditData.CopyFrom(a)
	if a.HasChildren() {
		other.Children = make([]*TraitModifier, 0, len(a.Children))
		for _, child := range a.Children {
			other.Children = append(other.Children, child.Clone(entity, other, preserveID))
		}
	}
	return other
}

// MarshalJSON implements json.Marshaler.
func (a *TraitModifier) MarshalJSON() ([]byte, error) {
	a.ClearUnusedFieldsForType()
	return json.Marshal(&a.TraitModifierData)
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *TraitModifier) UnmarshalJSON(data []byte) error {
	var localData struct {
		TraitModifierData
		// Old data fields
		Categories []string `json:"categories"`
	}
	if err := json.Unmarshal(data, &localData); err != nil {
		return err
	}
	localData.ClearUnusedFieldsForType()
	a.TraitModifierData = localData.TraitModifierData
	a.Tags = convertOldCategoriesToTags(a.Tags, localData.Categories)
	slices.Sort(a.Tags)
	if a.Container() {
		for _, one := range a.Children {
			one.parent = a
		}
	}
	return nil
}

// TagList returns the list of tags.
func (a *TraitModifier) TagList() []string {
	return a.Tags
}

// CellData returns the cell data information for the given column.
func (a *TraitModifier) CellData(column int, data *CellData) {
	switch column {
	case TraitModifierEnabledColumn:
		if !a.Container() {
			data.Type = Toggle
			data.Checked = a.Enabled()
			data.Alignment = unison.MiddleAlignment
		}
	case TraitModifierDescriptionColumn:
		data.Type = Text
		data.Primary = a.Name
		data.Secondary = a.SecondaryText(func(option display.Option) bool { return option.Inline() })
		data.Tooltip = a.SecondaryText(func(option display.Option) bool { return option.Tooltip() })
	case TraitModifierCostColumn:
		if !a.Container() {
			data.Type = Text
			data.Primary = a.CostDescription()
		}
	case TraitModifierTagsColumn:
		data.Type = Tags
		data.Primary = CombineTags(a.Tags)
	case TraitModifierReferenceColumn, PageRefCellAlias:
		data.Type = PageRef
		data.Primary = a.PageRef
		data.Secondary = a.Name
	}
}

// Depth returns the number of parents this node has.
func (a *TraitModifier) Depth() int {
	count := 0
	p := a.parent
	for p != nil {
		count++
		p = p.parent
	}
	return count
}

// OwningEntity returns the owning Entity.
func (a *TraitModifier) OwningEntity() *Entity {
	return a.Entity
}

// SetOwningEntity sets the owning entity and configures any sub-components as needed.
func (a *TraitModifier) SetOwningEntity(entity *Entity) {
	a.Entity = entity
	if a.Container() {
		for _, child := range a.Children {
			child.SetOwningEntity(entity)
		}
	}
}

// CostModifier returns the total cost modifier.
func (a *TraitModifier) CostModifier() fxp.Int {
	if a.Levels > 0 {
		return a.Cost.Mul(a.Levels)
	}
	return a.Cost
}

// HasLevels returns true if this TraitModifier has levels.
func (a *TraitModifier) HasLevels() bool {
	return !a.Container() && a.CostType == trait.Percentage && a.Levels > 0
}

func (a *TraitModifier) String() string {
	var buffer strings.Builder
	buffer.WriteString(a.Name)
	if a.HasLevels() {
		buffer.WriteByte(' ')
		buffer.WriteString(a.Levels.String())
	}
	return buffer.String()
}

// SecondaryText returns the "secondary" text: the text display below an Trait.
func (a *TraitModifier) SecondaryText(optionChecker func(display.Option) bool) string {
	var buffer strings.Builder
	settings := SheetSettingsFor(a.Entity)
	if a.LocalNotes != "" && optionChecker(settings.NotesDisplay) {
		if buffer.Len() != 0 {
			buffer.WriteByte('\n')
		}
		buffer.WriteString(a.LocalNotes)
	}
	return buffer.String()
}

// FullDescription returns a full description.
func (a *TraitModifier) FullDescription() string {
	var buffer strings.Builder
	buffer.WriteString(a.String())
	if a.LocalNotes != "" {
		buffer.WriteString(" (")
		buffer.WriteString(a.LocalNotes)
		buffer.WriteByte(')')
	}
	if SheetSettingsFor(a.Entity).ShowTraitModifierAdj {
		buffer.WriteString(" [")
		buffer.WriteString(a.CostDescription())
		buffer.WriteByte(']')
	}
	return buffer.String()
}

// CostDescription returns the formatted cost.
func (a *TraitModifier) CostDescription() string {
	if a.Container() {
		return ""
	}
	var base string
	switch a.CostType {
	case trait.Percentage:
		if a.HasLevels() {
			base = a.Cost.Mul(a.Levels).StringWithSign()
		} else {
			base = a.Cost.StringWithSign()
		}
		base += trait.Percentage.String()
	case trait.Points:
		base = a.Cost.StringWithSign()
	case trait.Multiplier:
		return a.CostType.String() + a.Cost.String()
	default:
		jot.Errorf("unhandled cost type: %d", a.CostType)
		base = a.Cost.StringWithSign() + trait.Percentage.String()
	}
	if desc := a.Affects.AltString(); desc != "" {
		base += " " + desc
	}
	return base
}

// FillWithNameableKeys adds any nameable keys found in this TraitModifier to the provided map.
func (a *TraitModifier) FillWithNameableKeys(m map[string]string) {
	if !a.Container() && a.Enabled() {
		nameables.Extract(a.Name, m)
		nameables.Extract(a.LocalNotes, m)
		nameables.Extract(a.VTTNotes, m)
		for _, one := range a.Features {
			one.FillWithNameableKeys(m)
		}
	}
}

// ApplyNameableKeys replaces any nameable keys found in this TraitModifier with the corresponding values in the provided map.
func (a *TraitModifier) ApplyNameableKeys(m map[string]string) {
	if !a.Container() && a.Enabled() {
		a.Name = nameables.Apply(a.Name, m)
		a.LocalNotes = nameables.Apply(a.LocalNotes, m)
		a.VTTNotes = nameables.Apply(a.VTTNotes, m)
		for _, one := range a.Features {
			one.ApplyNameableKeys(m)
		}
	}
}

// Enabled returns true if this node is enabled.
func (a *TraitModifier) Enabled() bool {
	return !a.Disabled || a.Container()
}

// SetEnabled makes the node enabled, if possible.
func (a *TraitModifier) SetEnabled(enabled bool) {
	if !a.Container() {
		a.Disabled = !enabled
	}
}
