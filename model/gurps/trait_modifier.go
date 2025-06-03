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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/affects"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/tmcost"
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
	trait *Trait
}

// TraitModifierData holds the TraitModifier data that is written to disk.
type TraitModifierData struct {
	SourcedID
	TraitModifierEditData
	ThirdParty map[string]any   `json:"third_party,omitempty"`
	Children   []*TraitModifier `json:"children,omitempty"` // Only for containers
	parent     *TraitModifier
}

// TraitModifierEditData holds the TraitModifier data that can be edited by the UI detail editor.
type TraitModifierEditData struct {
	TraitModifierSyncData
	VTTNotes     string            `json:"vtt_notes,omitempty"`
	Replacements map[string]string `json:"replacements,omitempty"` // Not actually used any longer, but kept so that we can migrate old data
	TraitModifierEditDataNonContainerOnly
}

// TraitModifierEditDataNonContainerOnly holds the TraitModifier data that is only applicable to
// TraitModifiers that aren't containers.
type TraitModifierEditDataNonContainerOnly struct {
	TraitModifierNonContainerSyncData
	Levels   fxp.Int `json:"levels,omitempty"`
	Disabled bool    `json:"disabled,omitempty"`
}

// TraitModifierSyncData holds the TraitModifier sync data that is common to both containers and non-containers.
type TraitModifierSyncData struct {
	Name             string   `json:"name,omitempty"`
	PageRef          string   `json:"reference,omitempty"`
	PageRefHighlight string   `json:"reference_highlight,omitempty"`
	LocalNotes       string   `json:"local_notes,omitempty"`
	Tags             []string `json:"tags,omitempty"`
}

// TraitModifierNonContainerSyncData holds the TraitModifier sync data that is only applicable to TraitModifiers that
// aren't containers.
type TraitModifierNonContainerSyncData struct {
	Cost              fxp.Int        `json:"cost,omitempty"`
	CostType          tmcost.Type    `json:"cost_type,omitempty"`
	UseLevelFromTrait bool           `json:"use_level_from_trait,omitempty"`
	ShowNotesOnWeapon bool           `json:"show_notes_on_weapon,omitempty"`
	Affects           affects.Option `json:"affects,omitempty"`
	Features          Features       `json:"features,omitempty"`
}

type traitModifierListData struct {
	Version int              `json:"version"`
	Rows    []*TraitModifier `json:"rows"`
}

// NewTraitModifiersFromFile loads a TraitModifier list from a file.
func NewTraitModifiersFromFile(fileSystem fs.FS, filePath string) ([]*TraitModifier, error) {
	var data traitModifierListData
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &data); err != nil {
		return nil, errs.NewWithCause(InvalidFileData(), err)
	}
	if err := jio.CheckVersion(data.Version); err != nil {
		return nil, err
	}
	return data.Rows, nil
}

// SaveTraitModifiers writes the TraitModifier list to the file as JSON.
func SaveTraitModifiers(modifiers []*TraitModifier, filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, &traitModifierListData{
		Version: jio.CurrentDataVersion,
		Rows:    modifiers,
	})
}

// NewTraitModifier creates a TraitModifier.
func NewTraitModifier(owner DataOwner, parent *TraitModifier, container bool) *TraitModifier {
	var t TraitModifier
	t.TID = tid.MustNewTID(traitModifierKind(container))
	t.parent = parent
	t.owner = owner
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
	other.AdjustSource(from, t.SourcedID, preserveID)
	other.SetOpen(t.IsOpen())
	other.ThirdParty = t.ThirdParty
	other.CopyFrom(t)
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
	type calc struct {
		ResolvedNotes string `json:"resolved_notes,omitempty"`
	}
	t.ClearUnusedFieldsForType()
	data := struct {
		TraitModifierData
		Calc *calc `json:"calc,omitempty"`
	}{
		TraitModifierData: t.TraitModifierData,
	}
	notes := t.ResolveLocalNotes()
	if notes != t.LocalNotes {
		data.Calc = &calc{ResolvedNotes: notes}
	}
	return json.Marshal(&data)
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *TraitModifier) UnmarshalJSON(data []byte) error {
	var localData struct {
		TraitModifierData
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
		localData.TID = tid.MustNewTID(traitModifierKind(strings.HasSuffix(localData.Type, containerKeyPostfix)))
		setOpen = localData.IsOpen
	}
	t.TraitModifierData = localData.TraitModifierData
	if t.LocalNotes == "" && localData.ExprNotes != "" {
		t.LocalNotes = EmbeddedExprToScript(localData.ExprNotes)
	}
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
		data.Detail = ModifierEnabledTooltip()
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
		data.Detail = PageRefTooltip()
	case TraitModifierLibSrcColumn:
		data.Title = HeaderDatabase
		data.TitleIsImageKey = true
		data.Detail = LibSrcTooltip()
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
		data.Primary = t.NameWithReplacements()
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
			data.Secondary = t.NameWithReplacements()
		}
	case TraitModifierLibSrcColumn:
		data.Type = cell.Text
		data.Alignment = align.Middle
		if !toolbox.IsNil(t.owner) {
			state, _ := t.owner.SourceMatcher().Match(t)
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

// OwningTrait returns the owning trait.
func (t *TraitModifier) OwningTrait() *Trait {
	return t.trait
}

// DataOwner returns the data owner.
func (t *TraitModifier) DataOwner() DataOwner {
	return t.owner
}

func (t *TraitModifier) setTrait(trait *Trait) {
	t.trait = trait
	if trait != nil && len(t.Replacements) != 0 {
		if t.trait.Replacements == nil {
			t.trait.Replacements = t.Replacements
		} else {
			for k, v := range t.Replacements {
				if _, exists := t.trait.Replacements[k]; !exists {
					t.trait.Replacements[k] = v
				}
			}
		}
		t.Replacements = nil
	}
	if t.Container() {
		for _, child := range t.Children {
			child.setTrait(trait)
		}
	}
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
	return t.Cost.Mul(t.CostMultiplier())
}

// IsLeveled returns true if this TraitModifier is leveled.
func (t *TraitModifier) IsLeveled() bool {
	if t.Container() {
		return false
	}
	if t.UseLevelFromTrait {
		if t.trait == nil {
			return false
		}
		return t.trait.IsLeveled()
	}
	return t.Levels > 0
}

// CostMultiplier returns the amount to multiply the cost by.
func (t *TraitModifier) CostMultiplier() fxp.Int {
	if !t.IsLeveled() {
		return fxp.One
	}
	return CostMultiplierForTraitModifier(t.Levels, t.trait, t.UseLevelFromTrait)
}

// CostMultiplierForTraitModifier returns the amount to multiply the cost by.
func CostMultiplierForTraitModifier(baseLevels fxp.Int, trait *Trait, useLevelFromTrait bool) fxp.Int {
	var multiplier fxp.Int
	if useLevelFromTrait {
		if trait != nil && trait.IsLeveled() {
			multiplier = trait.CurrentLevel()
		}
	} else {
		multiplier = baseLevels
	}
	if multiplier <= 0 {
		multiplier = fxp.One
	}
	return multiplier
}

// CurrentLevel returns the current level of the modifier or zero if it is not leveled.
func (t *TraitModifier) CurrentLevel() fxp.Int {
	if t.Enabled() && t.IsLeveled() {
		return t.CostMultiplier()
	}
	return 0
}

func (t *TraitModifier) String() string {
	var buffer strings.Builder
	buffer.WriteString(t.NameWithReplacements())
	if t.IsLeveled() {
		buffer.WriteByte(' ')
		buffer.WriteString(t.CurrentLevel().String())
	}
	return buffer.String()
}

// ResolveLocalNotes resolves the local notes, running any embedded scripts to get the final result.
func (t *TraitModifier) ResolveLocalNotes() string {
	return ResolveText(EntityFromNode(t), deferredNewScriptTrait(t.trait), t.LocalNotesWithReplacements())
}

// SecondaryText returns the "secondary" text: the text display below an Trait.
func (t *TraitModifier) SecondaryText(optionChecker func(display.Option) bool) string {
	if !optionChecker(SheetSettingsFor(EntityFromNode(t)).NotesDisplay) {
		return ""
	}
	return t.ResolveLocalNotes()
}

// FullDescription returns a full description.
func (t *TraitModifier) FullDescription() string {
	var buffer strings.Builder
	buffer.WriteString(t.String())
	if localNotes := t.ResolveLocalNotes(); localNotes != "" {
		buffer.WriteString(" (")
		buffer.WriteString(localNotes)
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
		base = t.CostModifier().StringWithSign() + tmcost.Percentage.String()
	case tmcost.Points:
		base = t.CostModifier().StringWithSign()
	case tmcost.Multiplier:
		return t.CostType.String() + t.CostModifier().String()
	default:
		errs.Log(errs.New("unknown cost type"), "type", int(t.CostType))
		base = t.CostModifier().StringWithSign() + tmcost.Percentage.String()
	}
	if desc := t.Affects.AltString(); desc != "" {
		base += " " + desc
	}
	return base
}

// NameWithReplacements returns the name with any replacements applied.
func (t *TraitModifier) NameWithReplacements() string {
	if t.trait == nil {
		return t.Name
	}
	return nameable.Apply(t.Name, t.trait.Replacements)
}

// LocalNotesWithReplacements returns the local notes with any replacements applied.
func (t *TraitModifier) LocalNotesWithReplacements() string {
	if t.trait == nil {
		return t.LocalNotes
	}
	return nameable.Apply(t.LocalNotes, t.trait.Replacements)
}

// NameableReplacements returns the replacements to be used with Nameables.
func (t *TraitModifier) NameableReplacements() map[string]string {
	if t == nil || t.trait == nil {
		return nil
	}
	return t.trait.Replacements
}

// FillWithNameableKeys adds any nameable keys found in this TraitModifier to the provided map.
func (t *TraitModifier) FillWithNameableKeys(m, existing map[string]string) {
	if !t.Container() && t.Enabled() {
		if existing == nil && t.trait != nil {
			existing = t.trait.Replacements
		}
		nameable.Extract(t.Name, m, existing)
		nameable.Extract(t.LocalNotes, m, existing)
		for _, one := range t.Features {
			one.FillWithNameableKeys(m, existing)
		}
	}
}

// ApplyNameableKeys passes this up to the owning trait to handle.
func (t *TraitModifier) ApplyNameableKeys(m map[string]string) {
	if len(m) != 0 && t.trait != nil {
		t.trait.ApplyNameableKeys(m)
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
		t.TraitModifierEditDataNonContainerOnly = TraitModifierEditDataNonContainerOnly{}
	} else {
		t.Children = nil
	}
}

// GetSource returns the source of this data.
func (t *TraitModifier) GetSource() Source {
	return t.Source
}

// ClearSource clears the source of this data.
func (t *TraitModifier) ClearSource() {
	t.Source = Source{}
}

// SyncWithSource synchronizes this data with the source.
func (t *TraitModifier) SyncWithSource() {
	if !toolbox.IsNil(t.owner) {
		if state, data := t.owner.SourceMatcher().Match(t); state == srcstate.Mismatched {
			if other, ok := data.(*TraitModifier); ok {
				t.TraitModifierSyncData = other.TraitModifierSyncData
				t.Tags = slices.Clone(other.Tags)
				if !t.Container() {
					t.TraitModifierNonContainerSyncData = other.TraitModifierNonContainerSyncData
					t.Features = other.Features.Clone()
				}
			}
		}
	}
}

// Hash writes this object's contents into the hasher. Note that this only hashes the data that is considered to be
// "source" data, i.e. not expected to be modified by the user after copying from a library.
func (t *TraitModifier) Hash(h hash.Hash) {
	t.hash(h)
	if t.Container() {
		hashhelper.Num8(h, uint8(255))
	} else {
		t.TraitModifierNonContainerSyncData.hash(h)
	}
}

func (t *TraitModifierSyncData) hash(h hash.Hash) {
	hashhelper.String(h, t.Name)
	hashhelper.String(h, t.PageRef)
	hashhelper.String(h, t.PageRefHighlight)
	hashhelper.String(h, t.LocalNotes)
	hashhelper.Num64(h, len(t.Tags))
	for _, tag := range t.Tags {
		hashhelper.String(h, tag)
	}
}

func (t *TraitModifierNonContainerSyncData) hash(h hash.Hash) {
	hashhelper.Num64(h, t.Cost)
	hashhelper.Num8(h, t.CostType)
	hashhelper.Bool(h, t.UseLevelFromTrait)
	hashhelper.Bool(h, t.ShowNotesOnWeapon)
	hashhelper.Num8(h, t.Affects)
	hashhelper.Num64(h, len(t.Features))
	for _, feature := range t.Features {
		feature.Hash(h)
	}
}

// CopyFrom implements node.EditorData.
func (t *TraitModifierEditData) CopyFrom(other *TraitModifier) {
	t.copyFrom(&other.TraitModifierEditData)
}

// ApplyTo implements node.EditorData.
func (t *TraitModifierEditData) ApplyTo(other *TraitModifier) {
	other.copyFrom(t)
}

func (t *TraitModifierEditData) copyFrom(other *TraitModifierEditData) {
	*t = *other
	t.Tags = txt.CloneStringSlice(other.Tags)
	t.Features = other.Features.Clone()
}
