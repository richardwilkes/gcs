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

package model

import (
	"context"
	"io/fs"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	gid2 "github.com/richardwilkes/gcs/v5/model/gid"
	"github.com/richardwilkes/gcs/v5/model/jio"
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
	FullCostDescription() string
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
		return nil, errs.NewWithCause(gid2.InvalidFileDataMsg, err)
	}
	if data.Type != traitModifierListTypeKey {
		return nil, errs.New(gid2.UnexpectedFileDataMsg)
	}
	if err := gid2.CheckVersion(data.Version); err != nil {
		return nil, err
	}
	return data.Rows, nil
}

// SaveTraitModifiers writes the TraitModifier list to the file as JSON.
func SaveTraitModifiers(modifiers []*TraitModifier, filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, &traitModifierListData{
		Type:    traitModifierListTypeKey,
		Version: gid2.CurrentDataVersion,
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
func (m *TraitModifier) Clone(entity *Entity, parent *TraitModifier, preserveID bool) *TraitModifier {
	other := NewTraitModifier(entity, parent, m.Container())
	if preserveID {
		other.ID = m.ID
	}
	other.IsOpen = m.IsOpen
	other.TraitModifierEditData.CopyFrom(m)
	if m.HasChildren() {
		other.Children = make([]*TraitModifier, 0, len(m.Children))
		for _, child := range m.Children {
			other.Children = append(other.Children, child.Clone(entity, other, preserveID))
		}
	}
	return other
}

// MarshalJSON implements json.Marshaler.
func (m *TraitModifier) MarshalJSON() ([]byte, error) {
	m.ClearUnusedFieldsForType()
	return json.Marshal(&m.TraitModifierData)
}

// UnmarshalJSON implements json.Unmarshaler.
func (m *TraitModifier) UnmarshalJSON(data []byte) error {
	var localData struct {
		TraitModifierData
		// Old data fields
		Categories []string `json:"categories"`
	}
	if err := json.Unmarshal(data, &localData); err != nil {
		return err
	}
	localData.ClearUnusedFieldsForType()
	m.TraitModifierData = localData.TraitModifierData
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
func (m *TraitModifier) TagList() []string {
	return m.Tags
}

// CellData returns the cell data information for the given column.
func (m *TraitModifier) CellData(column int, data *CellData) {
	switch column {
	case TraitModifierEnabledColumn:
		if !m.Container() {
			data.Type = ToggleCellType
			data.Checked = m.Enabled()
			data.Alignment = unison.MiddleAlignment
		}
	case TraitModifierDescriptionColumn:
		data.Type = TextCellType
		data.Primary = m.Name
		data.Secondary = m.SecondaryText(func(option DisplayOption) bool { return option.Inline() })
		data.Tooltip = m.SecondaryText(func(option DisplayOption) bool { return option.Tooltip() })
	case TraitModifierCostColumn:
		if !m.Container() {
			data.Type = TextCellType
			data.Primary = m.CostDescription()
		}
	case TraitModifierTagsColumn:
		data.Type = TagsCellType
		data.Primary = CombineTags(m.Tags)
	case TraitModifierReferenceColumn, PageRefCellAlias:
		data.Type = PageRefCellType
		data.Primary = m.PageRef
		data.Secondary = m.Name
	}
}

// Depth returns the number of parents this node has.
func (m *TraitModifier) Depth() int {
	count := 0
	p := m.parent
	for p != nil {
		count++
		p = p.parent
	}
	return count
}

// OwningEntity returns the owning Entity.
func (m *TraitModifier) OwningEntity() *Entity {
	return m.Entity
}

// SetOwningEntity sets the owning entity and configures any sub-components as needed.
func (m *TraitModifier) SetOwningEntity(entity *Entity) {
	m.Entity = entity
	if m.Container() {
		for _, child := range m.Children {
			child.SetOwningEntity(entity)
		}
	}
}

// CostModifier returns the total cost modifier.
func (m *TraitModifier) CostModifier() fxp.Int {
	if m.Levels > 0 {
		return m.Cost.Mul(m.Levels)
	}
	return m.Cost
}

// HasLevels returns true if this TraitModifier has levels.
func (m *TraitModifier) HasLevels() bool {
	return !m.Container() && m.CostType == PercentageTraitModifierCostType && m.Levels > 0
}

func (m *TraitModifier) String() string {
	var buffer strings.Builder
	buffer.WriteString(m.Name)
	if m.HasLevels() {
		buffer.WriteByte(' ')
		buffer.WriteString(m.Levels.String())
	}
	return buffer.String()
}

// SecondaryText returns the "secondary" text: the text display below an Trait.
func (m *TraitModifier) SecondaryText(optionChecker func(DisplayOption) bool) string {
	var buffer strings.Builder
	settings := SheetSettingsFor(m.Entity)
	if m.LocalNotes != "" && optionChecker(settings.NotesDisplay) {
		if buffer.Len() != 0 {
			buffer.WriteByte('\n')
		}
		buffer.WriteString(m.LocalNotes)
	}
	return buffer.String()
}

// FullDescription returns a full description.
func (m *TraitModifier) FullDescription() string {
	var buffer strings.Builder
	buffer.WriteString(m.String())
	if m.LocalNotes != "" {
		buffer.WriteString(" (")
		buffer.WriteString(m.LocalNotes)
		buffer.WriteByte(')')
	}
	if SheetSettingsFor(m.Entity).ShowTraitModifierAdj {
		buffer.WriteString(" [")
		buffer.WriteString(m.CostDescription())
		buffer.WriteByte(']')
	}
	return buffer.String()
}

// FullCostDescription is the same as CostDescription().
func (m *TraitModifier) FullCostDescription() string {
	return m.CostDescription()
}

// CostDescription returns the formatted cost.
func (m *TraitModifier) CostDescription() string {
	if m.Container() {
		return ""
	}
	var base string
	switch m.CostType {
	case PercentageTraitModifierCostType:
		if m.HasLevels() {
			base = m.Cost.Mul(m.Levels).StringWithSign()
		} else {
			base = m.Cost.StringWithSign()
		}
		base += PercentageTraitModifierCostType.String()
	case PointsTraitModifierCostType:
		base = m.Cost.StringWithSign()
	case MultiplierTraitModifierCostType:
		return m.CostType.String() + m.Cost.String()
	default:
		jot.Errorf("unhandled cost type: %d", m.CostType)
		base = m.Cost.StringWithSign() + PercentageTraitModifierCostType.String()
	}
	if desc := m.Affects.AltString(); desc != "" {
		base += " " + desc
	}
	return base
}

// FillWithNameableKeys adds any nameable keys found in this TraitModifier to the provided map.
func (m *TraitModifier) FillWithNameableKeys(keyMap map[string]string) {
	if !m.Container() && m.Enabled() {
		Extract(m.Name, keyMap)
		Extract(m.LocalNotes, keyMap)
		for _, one := range m.Features {
			one.FillWithNameableKeys(keyMap)
		}
	}
}

// ApplyNameableKeys replaces any nameable keys found in this TraitModifier with the corresponding values in the provided map.
func (m *TraitModifier) ApplyNameableKeys(keyMap map[string]string) {
	if !m.Container() && m.Enabled() {
		m.Name = Apply(m.Name, keyMap)
		m.LocalNotes = Apply(m.LocalNotes, keyMap)
		for _, one := range m.Features {
			one.ApplyNameableKeys(keyMap)
		}
	}
}

// Enabled returns true if this node is enabled.
func (m *TraitModifier) Enabled() bool {
	return !m.Disabled || m.Container()
}

// SetEnabled makes the node enabled, if possible.
func (m *TraitModifier) SetEnabled(enabled bool) {
	if !m.Container() {
		m.Disabled = !enabled
	}
}
