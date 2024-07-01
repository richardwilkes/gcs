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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/affects"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/tmcost"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/gcs/v5/model/message"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/tid"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison/enums/align"
)

var (
	_ Node[*TraitModifier]       = &TraitModifier{}
	_ GeneralModifier            = &TraitModifier{}
	_ LeveledOwner               = &TraitModifier{}
	_ EditorData[*TraitModifier] = &TraitModifierEditData{}
)

// Columns that can be used with the trait modifier method .CellData()
const (
	TraitModifierEnabledColumn = iota
	TraitModifierDescriptionColumn
	TraitModifierCostColumn
	TraitModifierTagsColumn
	TraitModifierReferenceColumn
	TraitModifierLibSrcColumn
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
	owner DataOwner
}

// TraitModifierData holds the TraitModifier data that is written to disk.
type TraitModifierData struct {
	TID    tid.TID `json:"id"`
	Source Source  `json:"source,omitempty"`
	TraitModifierEditData
	ThirdParty map[string]any   `json:"third_party,omitempty"`
	Children   []*TraitModifier `json:"children,omitempty"` // Only for containers
	parent     *TraitModifier
}

// TraitModifierEditData holds the TraitModifier data that can be edited by the UI detail editor.
type TraitModifierEditData struct {
	Name             string   `json:"name,omitempty"`
	PageRef          string   `json:"reference,omitempty"`
	PageRefHighlight string   `json:"reference_highlight,omitempty"`
	LocalNotes       string   `json:"notes,omitempty"`
	VTTNotes         string   `json:"vtt_notes,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	TraitModifierEditDataNonContainerOnly
}

// TraitModifierEditDataNonContainerOnly holds the TraitModifier data that is only applicable to
// TraitModifiers that aren't containers.
type TraitModifierEditDataNonContainerOnly struct {
	Cost     fxp.Int        `json:"cost,omitempty"`
	Levels   fxp.Int        `json:"levels,omitempty"`
	Affects  affects.Option `json:"affects,omitempty"`
	CostType tmcost.Type    `json:"cost_type,omitempty"`
	Disabled bool           `json:"disabled,omitempty"`
	Features Features       `json:"features,omitempty"`
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
		return nil, errs.NewWithCause(message.InvalidFileData(), err)
	}
	if data.Type != traitModifierListTypeKey {
		return nil, errs.New(message.UnexpectedFileData())
	}
	if err := jio.CheckVersion(data.Version); err != nil {
		return nil, err
	}
	return data.Rows, nil
}

// SaveTraitModifiers writes the TraitModifier list to the file as JSON.
func SaveTraitModifiers(modifiers []*TraitModifier, filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, &traitModifierListData{
		Type:    traitModifierListTypeKey,
		Version: jio.CurrentDataVersion,
		Rows:    modifiers,
	})
}

// NewTraitModifier creates a TraitModifier.
func NewTraitModifier(owner DataOwner, parent *TraitModifier, container bool) *TraitModifier {
	t := TraitModifier{
		TraitModifierData: TraitModifierData{
			TID:    tid.MustNewTID(traitModifierKind(container)),
			parent: parent,
		},
		owner: owner,
	}
	t.Name = t.Kind()
	t.SetOpen(container)
	return &t
}

func traitModifierKind(container bool) byte {
	if container {
		return kinds.TraitModifierContainer
	}
	return kinds.TraitModifier
}

// GetSource returns the source of this data.
func (t *TraitModifier) GetSource() Source {
	return t.Source
}

// ID returns the local ID of this data.
func (t *TraitModifier) ID() tid.TID {
	return t.TID
}

// Container returns true if this is a container.
func (t *TraitModifier) Container() bool {
	return tid.IsKind(t.TID, kinds.TraitModifierContainer)
}

// HasChildren returns true if this node has children.
func (t *TraitModifier) HasChildren() bool {
	return t.Container() && len(t.Children) > 0
}

// NodeChildren returns the children of this node, if any.
func (t *TraitModifier) NodeChildren() []*TraitModifier {
	return t.Children
}

// SetChildren sets the children of this node.
func (t *TraitModifier) SetChildren(children []*TraitModifier) {
	t.Children = children
}

// Parent returns the parent.
func (t *TraitModifier) Parent() *TraitModifier {
	return t.parent
}

// SetParent sets the parent.
func (t *TraitModifier) SetParent(parent *TraitModifier) {
	t.parent = parent
}

// IsOpen returns true if this node is currently open.
func (t *TraitModifier) IsOpen() bool {
	return IsNodeOpen(t)
}

// SetOpen sets the current open state for this node.
func (t *TraitModifier) SetOpen(open bool) {
	SetNodeOpen(t, open)
}

// Clone implements Node.
func (t *TraitModifier) Clone(from LibraryFile, owner DataOwner, parent *TraitModifier, preserveID bool) *TraitModifier {
	other := NewTraitModifier(owner, parent, t.Container())
	other.Source.LibraryFile = from
	other.Source.TID = t.TID
	if preserveID {
		other.TID = t.TID
	}
	other.SetOpen(t.IsOpen())
	other.ThirdParty = t.ThirdParty
	other.TraitModifierEditData.CopyFrom(t)
	if t.HasChildren() {
		other.Children = make([]*TraitModifier, 0, len(t.Children))
		for _, child := range t.Children {
			other.Children = append(other.Children, child.Clone(from, owner, other, preserveID))
		}
	}
	return other
}

// MarshalJSON implements json.Marshaler.
func (t *TraitModifier) MarshalJSON() ([]byte, error) {
	t.ClearUnusedFieldsForType()
	return json.Marshal(&t.TraitModifierData)
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *TraitModifier) UnmarshalJSON(data []byte) error {
	var localData struct {
		TraitModifierData
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
		localData.TID = tid.MustNewTID(traitModifierKind(strings.HasSuffix(localData.Type, ContainerKeyPostfix)))
		setOpen = localData.IsOpen
	}
	t.TraitModifierData = localData.TraitModifierData
	t.ClearUnusedFieldsForType()
	t.Tags = convertOldCategoriesToTags(t.Tags, localData.Categories)
	slices.Sort(t.Tags)
	if t.Container() {
		for _, one := range t.Children {
			one.parent = t
		}
	}
	if setOpen {
		SetNodeOpen(t, true)
	}
	return nil
}

// TagList returns the list of tags.
func (t *TraitModifier) TagList() []string {
	return t.Tags
}

// TraitModifierHeaderData returns the header data information for the given trait modifier column.
func TraitModifierHeaderData(columnID int) HeaderData {
	var data HeaderData
	switch columnID {
	case TraitModifierEnabledColumn:
		data.Title = HeaderCheckmark
		data.TitleIsImageKey = true
		data.Detail = message.ModifierEnabledTooltip()
	case TraitModifierDescriptionColumn:
		data.Title = i18n.Text("Trait Modifier")
		data.Primary = true
	case TraitModifierCostColumn:
		data.Title = i18n.Text("Cost Adjustment")
	case TraitModifierTagsColumn:
		data.Title = i18n.Text("Tags")
	case TraitModifierReferenceColumn:
		data.Title = HeaderBookmark
		data.TitleIsImageKey = true
		data.Detail = message.PageRefTooltip()
	case TraitModifierLibSrcColumn:
		data.Title = HeaderDatabase
		data.TitleIsImageKey = true
		data.Detail = message.LibSrcTooltip()
	}
	return data
}

// CellData returns the cell data information for the given column.
func (t *TraitModifier) CellData(columnID int, data *CellData) {
	switch columnID {
	case TraitModifierEnabledColumn:
		if !t.Container() {
			data.Type = cell.Toggle
			data.Checked = t.Enabled()
			data.Alignment = align.Middle
		}
	case TraitModifierDescriptionColumn:
		data.Type = cell.Text
		data.Primary = t.Name
		data.Secondary = t.SecondaryText(func(option display.Option) bool { return option.Inline() })
		data.Tooltip = t.SecondaryText(func(option display.Option) bool { return option.Tooltip() })
	case TraitModifierCostColumn:
		if !t.Container() {
			data.Type = cell.Text
			data.Primary = t.CostDescription()
		}
	case TraitModifierTagsColumn:
		data.Type = cell.Tags
		data.Primary = CombineTags(t.Tags)
	case TraitModifierReferenceColumn, PageRefCellAlias:
		data.Type = cell.PageRef
		data.Primary = t.PageRef
		if t.PageRefHighlight != "" {
			data.Secondary = t.PageRefHighlight
		} else {
			data.Secondary = t.Name
		}
	case TraitModifierLibSrcColumn:
		data.Type = cell.Text
		data.Alignment = align.Middle
		if !toolbox.IsNil(t.owner) {
			state := t.owner.SourceMatcher().Match(t)
			data.Primary = state.AltString()
			data.Tooltip = state.String()
			if state != srcstate.Custom {
				data.Tooltip += "\n" + t.Source.String()
			}
		}
	}
}

// Depth returns the number of parents this node has.
func (t *TraitModifier) Depth() int {
	count := 0
	p := t.parent
	for p != nil {
		count++
		p = p.parent
	}
	return count
}

// DataOwner returns the data owner.
func (t *TraitModifier) DataOwner() DataOwner {
	return t.owner
}

// SetDataOwner sets the data owner and configures any sub-components as needed.
func (t *TraitModifier) SetDataOwner(owner DataOwner) {
	t.owner = owner
	if t.Container() {
		for _, child := range t.Children {
			child.SetDataOwner(owner)
		}
	}
}

// CostModifier returns the total cost modifier.
func (t *TraitModifier) CostModifier() fxp.Int {
	if t.Levels > 0 {
		return t.Cost.Mul(t.Levels)
	}
	return t.Cost
}

// IsLeveled returns true if this TraitModifier is leveled.
func (t *TraitModifier) IsLeveled() bool {
	return !t.Container() && t.CostType == tmcost.Percentage && t.Levels > 0
}

// CurrentLevel returns the current level of the modifier or zero if it is not leveled.
func (t *TraitModifier) CurrentLevel() fxp.Int {
	if t.Enabled() && t.IsLeveled() {
		return t.Levels
	}
	return 0
}

func (t *TraitModifier) String() string {
	var buffer strings.Builder
	buffer.WriteString(t.Name)
	if t.IsLeveled() {
		buffer.WriteByte(' ')
		buffer.WriteString(t.Levels.String())
	}
	return buffer.String()
}

// SecondaryText returns the "secondary" text: the text display below an Trait.
func (t *TraitModifier) SecondaryText(optionChecker func(display.Option) bool) string {
	if optionChecker(SheetSettingsFor(EntityFromNode(t)).NotesDisplay) {
		return t.LocalNotes
	}
	return ""
}

// FullDescription returns a full description.
func (t *TraitModifier) FullDescription() string {
	var buffer strings.Builder
	buffer.WriteString(t.String())
	if t.LocalNotes != "" {
		buffer.WriteString(" (")
		buffer.WriteString(t.LocalNotes)
		buffer.WriteByte(')')
	}
	if SheetSettingsFor(EntityFromNode(t)).ShowTraitModifierAdj {
		buffer.WriteString(" [")
		buffer.WriteString(t.CostDescription())
		buffer.WriteByte(']')
	}
	return buffer.String()
}

// FullCostDescription is the same as CostDescription().
func (t *TraitModifier) FullCostDescription() string {
	return t.CostDescription()
}

// CostDescription returns the formatted cost.
func (t *TraitModifier) CostDescription() string {
	if t.Container() {
		return ""
	}
	var base string
	switch t.CostType {
	case tmcost.Percentage:
		if t.IsLeveled() {
			base = t.Cost.Mul(t.Levels).StringWithSign()
		} else {
			base = t.Cost.StringWithSign()
		}
		base += tmcost.Percentage.String()
	case tmcost.Points:
		base = t.Cost.StringWithSign()
	case tmcost.Multiplier:
		return t.CostType.String() + t.Cost.String()
	default:
		errs.Log(errs.New("unknown cost type"), "type", int(t.CostType))
		base = t.Cost.StringWithSign() + tmcost.Percentage.String()
	}
	if desc := t.Affects.AltString(); desc != "" {
		base += " " + desc
	}
	return base
}

// FillWithNameableKeys adds any nameable keys found in this TraitModifier to the provided map.
func (t *TraitModifier) FillWithNameableKeys(keyMap map[string]string) {
	if !t.Container() && t.Enabled() {
		Extract(t.Name, keyMap)
		Extract(t.LocalNotes, keyMap)
		for _, one := range t.Features {
			one.FillWithNameableKeys(keyMap)
		}
	}
}

// ApplyNameableKeys replaces any nameable keys found in this TraitModifier with the corresponding values in the
// provided map.
func (t *TraitModifier) ApplyNameableKeys(keyMap map[string]string) {
	if !t.Container() && t.Enabled() {
		t.Name = Apply(t.Name, keyMap)
		t.LocalNotes = Apply(t.LocalNotes, keyMap)
		for _, one := range t.Features {
			one.ApplyNameableKeys(keyMap)
		}
	}
}

// Enabled returns true if this node is enabled.
func (t *TraitModifier) Enabled() bool {
	return !t.Disabled || t.Container()
}

// SetEnabled makes the node enabled, if possible.
func (t *TraitModifier) SetEnabled(enabled bool) {
	if !t.Container() {
		t.Disabled = !enabled
	}
}

// Kind returns the kind of data.
func (t *TraitModifier) Kind() string {
	if t.Container() {
		return i18n.Text("Trait Modifier Container")
	}
	return i18n.Text("Trait Modifier")
}

// ClearUnusedFieldsForType zeroes out the fields that are not applicable to this type (container vs not-container).
func (t *TraitModifier) ClearUnusedFieldsForType() {
	if t.Container() {
		t.CostType = 0
		t.Disabled = false
		t.Cost = 0
		t.Levels = 0
		t.Affects = 0
		t.Features = nil
	} else {
		t.Children = nil
	}
}

// Hash writes this object's contents into the hasher. Note that this only hashes the data that is considered to be
// "source" data, i.e. not expected to be modified by the user after copying from a library.
func (t *TraitModifier) Hash(h hash.Hash) {
	_, _ = h.Write([]byte(t.Name))
	_, _ = h.Write([]byte(t.PageRef))
	_, _ = h.Write([]byte(t.PageRefHighlight))
	_, _ = h.Write([]byte(t.LocalNotes))
	_, _ = h.Write([]byte(t.VTTNotes))
	for _, tag := range t.Tags {
		_, _ = h.Write([]byte(tag))
	}
	if !t.Container() {
		_ = binary.Write(h, binary.LittleEndian, t.Cost)
		_ = binary.Write(h, binary.LittleEndian, t.Levels)
		_ = binary.Write(h, binary.LittleEndian, t.Affects)
		_ = binary.Write(h, binary.LittleEndian, t.CostType)
		for _, feature := range t.Features {
			feature.Hash(h)
		}
	}
}

// CopyFrom implements node.EditorData.
func (t *TraitModifierEditData) CopyFrom(other *TraitModifier) {
	t.copyFrom(&other.TraitModifierEditData)
}

// ApplyTo implements node.EditorData.
func (t *TraitModifierEditData) ApplyTo(other *TraitModifier) {
	other.TraitModifierEditData.copyFrom(t)
}

func (t *TraitModifierEditData) copyFrom(other *TraitModifierEditData) {
	*t = *other
	t.Tags = txt.CloneStringSlice(t.Tags)
	t.Features = other.Features.Clone()
}
