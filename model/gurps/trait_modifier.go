// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"hash"
	"io/fs"
	"maps"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/affects"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emweight"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison/enums/align"
)

var (
	_ = assertNode[*TraitModifier]
	_ = assertModifierNode[*TraitModifier]
	_ = assertEditorData[*TraitModifierEditData]

	_ GeneralModifier = &TraitModifier{}
	_ LeveledOwner    = &TraitModifier{}
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
	VTTNotes     string            `json:"vtt_notes,omitzero"`
	Replacements map[string]string `json:"replacements,omitempty"` // Not actually used any longer, but kept so that we can migrate old data
	TraitModifierEditDataNonContainerOnly
}

// TraitModifierEditDataNonContainerOnly holds the TraitModifier data that is only applicable to
// TraitModifiers that aren't containers.
type TraitModifierEditDataNonContainerOnly struct {
	TraitModifierNonContainerSyncData
	Levels   fxp.Int `json:"levels,omitzero"`
	Disabled bool    `json:"disabled,omitzero"`
}

// TraitModifierSyncData holds the TraitModifier sync data that is common to both containers and non-containers.
type TraitModifierSyncData struct {
	Name             string   `json:"name,omitzero"`
	PageRef          string   `json:"reference,omitzero"`
	PageRefHighlight string   `json:"reference_highlight,omitzero"`
	LocalNotes       string   `json:"local_notes,omitzero"`
	Tags             []string `json:"tags,omitempty"`
}

// TraitModifierNonContainerSyncData holds the TraitModifier sync data that is only applicable to TraitModifiers that
// aren't containers.
type TraitModifierNonContainerSyncData struct {
	CostAdj           string         `json:"cost_adj,omitzero"`
	UseLevelFromTrait bool           `json:"use_level_from_trait,omitzero"`
	CostIgnoresLevel  bool           `json:"cost_ignores_level,omitzero"`
	ShowNotesOnWeapon bool           `json:"show_notes_on_weapon,omitzero"`
	Affects           affects.Option `json:"affects,omitzero"`
	Features          Features       `json:"features,omitempty"`
}

type traitModifierListData struct {
	Version int              `json:"version"`
	Rows    []*TraitModifier `json:"rows"`
}

// NewTraitModifiersFromFile loads a TraitModifier list from a file.
func NewTraitModifiersFromFile(fileSystem fs.FS, filePath string) ([]*TraitModifier, error) {
	var data traitModifierListData
	if err := jio.LoadFromFile(fileSystem, filePath, &data); err != nil {
		return nil, errs.NewWithCause(InvalidFileData(), err)
	}
	if err := jio.CheckVersion(data.Version); err != nil {
		return nil, err
	}
	return data.Rows, nil
}

// SaveTraitModifiers writes the TraitModifier list to the file as JSON.
func SaveTraitModifiers(modifiers []*TraitModifier, filePath string) error {
	return jio.SaveToFile(filePath, &traitModifierListData{
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
func (t *TraitModifier) Clone(from LibraryFile, owner DataOwner, parent *TraitModifier, mode CloneMode) *TraitModifier {
	other := NewTraitModifier(owner, parent, t.Container())
	other.AdjustSource(from, t.SourcedID, mode)
	other.SetOpen(t.IsOpen())
	other.ThirdParty = t.ThirdParty
	other.CopyFrom(t)
	PropagateNodeNoteClosedState(t, other)
	if t.HasChildren() {
		other.Children = make([]*TraitModifier, 0, len(t.Children))
		for _, child := range t.Children {
			other.Children = append(other.Children, child.Clone(from, owner, other, mode))
		}
	}
	return other
}

// MarshalJSONTo implements json.MarshalerTo.
func (t *TraitModifier) MarshalJSONTo(enc *jsontext.Encoder) error {
	type calc struct {
		ResolvedNotes string `json:"resolved_notes,omitzero"`
	}
	t.ClearUnusedFieldsForType()
	if omitCalc(enc) {
		return json.MarshalEncode(enc, &t.TraitModifierData)
	}
	data := struct {
		TraitModifierData
		Calc *calc `json:"calc,omitzero"`
	}{
		TraitModifierData: t.TraitModifierData,
	}
	notes := t.ResolveLocalNotes()
	if notes != t.LocalNotes {
		data.Calc = &calc{ResolvedNotes: notes}
	}
	return json.MarshalEncode(enc, &data)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (t *TraitModifier) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var localData struct {
		TraitModifierData
		// Old data fields
		Cost       fxp.Int  `json:"cost"`
		CostType   string   `json:"cost_type"`
		Type       string   `json:"type"`
		ExprNotes  string   `json:"notes"`
		Categories []string `json:"categories"`
		IsOpen     bool     `json:"open"`
	}
	if err := json.UnmarshalDecode(dec, &localData); err != nil {
		return err
	}
	setOpen := false
	if !tid.IsValid(localData.TID) {
		// Fixup old data that used UUIDs instead of TIDs
		localData.TID = tid.MustNewTID(traitModifierKind(strings.HasSuffix(localData.Type, containerKeyPostfix)))
		setOpen = localData.IsOpen
	}
	if localData.CostAdj == "" && localData.Cost != 0 {
		switch localData.CostType {
		case "points":
			localData.CostAdj = localData.Cost.String()
		case "multiplier":
			localData.CostAdj = "x" + localData.Cost.String()
		default:
			localData.CostAdj = localData.Cost.String() + "%"
		}
	}
	t.TraitModifierData = localData.TraitModifierData
	t.Replacements = nameable.Normalize(t.Replacements)
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
	data.Self = t
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
		if !xreflect.IsNil(t.owner) {
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

// Target returns the trait being targeted for modification
func (t *TraitModifier) Target() *Trait {
	return t.trait
}

// DataOwner returns the data owner.
func (t *TraitModifier) DataOwner() DataOwner {
	return t.owner
}

// SetTarget sets the trait being targeted for modification and configures any sub-components as needed.
func (t *TraitModifier) SetTarget(target *Trait) *TraitModifier {
	// Set the target node for this modifier
	t.trait = target

	// COMPAT: Promote replacements from this node up to the target node
	if target != nil && len(t.Replacements) != 0 {
		target.Replacements = mergeReplacements(target.Replacements, t.Replacements)
		t.Replacements = nil
	}

	// Cascade the operation
	if t.Container() {
		for _, child := range t.Children {
			child.SetTarget(target)
		}
	}

	return t
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

// CostModifierType returns the type of cost modifier.
func (t *TraitModifier) CostModifierType() emweight.Value {
	return emweight.ValueFromString(t.CostAdj)
}

// CostModifier returns the total cost modifier.
func (t *TraitModifier) CostModifier() fxp.Fraction {
	return t.CostModifierForTrait(t.trait)
}

// CostModifierForTrait returns the total cost modifier as applied to the given trait, which is used in place of the
// modifier's own owning trait when resolving a "use level from trait" modifier. This allows a modifier inherited from a
// parent container (or one held only by an editor's working copy) to be costed against the trait it is being applied
// to without re-pointing the modifier at that trait. Containers have no cost modifier and always yield zero. When
// CostIgnoresLevel is set, the adjustment is not multiplied by the level here, which lets a modifier drive per-level
// features from the trait's level without its own price being scaled by that level as well. A point adder whose
// Affects is "levels only" is still charged for each of the trait's levels either way, since AdjustedPoints folds it
// into the trait's cost per level.
func (t *TraitModifier) CostModifierForTrait(trait *Trait) fxp.Fraction {
	if t.Container() {
		return fxp.Fraction{Denominator: fxp.One}
	}
	f := t.CostModifierType().ExtractFraction(t.CostAdj)
	if !t.CostIgnoresLevel && t.isLeveledForTrait(trait) {
		f.Numerator = f.Numerator.Mul(t.costLevelsForTrait(trait))
	}
	f.Normalize()
	return f
}

// costLevelsForTrait returns the number of levels the cost adjustment is multiplied by when applied to the given
// trait. A "use level from owner" modifier is costed against the trait's purchased levels rather than its current
// level: bonus-granted levels are free, so they must not attract enhancement or limitation percentages. This matches
// Trait.AdjustedPoints, which derives the leveled base cost from those same purchased levels.
func (t *TraitModifier) costLevelsForTrait(trait *Trait) fxp.Int {
	var levels fxp.Int
	if t.UseLevelFromTrait {
		if trait != nil && trait.IsLeveled() {
			levels = trait.Levels
		}
	} else {
		levels = t.Levels
	}
	if levels <= 0 {
		levels = fxp.One
	}
	return levels
}

// IsLeveled returns true if this TraitModifier is leveled.
func (t *TraitModifier) IsLeveled() bool {
	return t.isLeveledForTrait(t.trait)
}

// isLeveledForTrait returns true if this TraitModifier is leveled when applied to the given trait, which stands in for
// the modifier's own owning trait when resolving a "use level from trait" modifier.
func (t *TraitModifier) isLeveledForTrait(trait *Trait) bool {
	if t.Container() {
		return false
	}
	if t.UseLevelFromTrait {
		return trait != nil && trait.IsLeveled()
	}
	return t.Levels > 0
}

// RawCurrentLevel returns the current level of the modifier or zero if it is not leveled.
func (t *TraitModifier) RawCurrentLevel() fxp.Int {
	var level fxp.Int
	if t.UseLevelFromTrait {
		if t.trait != nil && t.trait.IsLeveled() {
			level = t.trait.CurrentLevel()
		}
	} else {
		level = t.Levels
	}
	return level
}

// CurrentLevel returns the current level of the modifier or zero if it is not leveled. Minimum of 1 will be returned
// if it has levels at all. Unlike the levels the cost is computed from, a "use level from owner" modifier reports the
// trait's current level, bonus-granted levels included, since this drives the per-level features the modifier carries
// and how it displays, neither of which is a matter of what was paid for.
func (t *TraitModifier) CurrentLevel() fxp.Int {
	if t.Enabled() && t.IsLeveled() {
		return t.RawCurrentLevel().Max(fxp.One)
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
	return ResolveText(EntityFromNode(t), deferredNewScriptTraitModifier(t), t.LocalNotesWithReplacements())
}

// SecondaryText returns the "secondary" text: the text displayed below the modifier.
func (t *TraitModifier) SecondaryText(optionChecker func(display.Option) bool) string {
	return modifierSecondaryText(t, optionChecker)
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
	base := t.CostModifierType().Format(t.CostModifier().Simplify())
	if desc := t.Affects.AltString(); desc != "" {
		base += " " + desc
	}
	return base
}

// NameWithReplacements returns the name with any replacements applied.
func (t *TraitModifier) NameWithReplacements() string {
	return applyOwnerReplacements(t.Name, t.trait)
}

// LocalNotesWithReplacements returns the local notes with any replacements applied.
func (t *TraitModifier) LocalNotesWithReplacements() string {
	return applyOwnerReplacements(t.LocalNotes, t.trait)
}

// NameableReplacements returns the replacements to be used with Nameables.
func (t *TraitModifier) NameableReplacements() map[string]string {
	if t == nil || t.trait == nil {
		return nil
	}
	return t.trait.Replacements
}

// FillWithNameableKeys adds any nameable keys found in this TraitModifier to the provided map. Containers take part
// too, since their name and notes are displayed with replacements applied just as a leaf modifier's are.
func (t *TraitModifier) FillWithNameableKeys(m, existing map[string]string) {
	if t.Enabled() {
		if existing == nil {
			existing = t.NameableReplacements()
		}
		nameable.Extract(
			m, existing,
			t.Name,
			t.LocalNotes,
		)
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
	syncFromSource(t, func(other *TraitModifier) {
		t.TraitModifierSyncData = other.TraitModifierSyncData
		t.Tags = slices.Clone(other.Tags)
		if !t.Container() {
			t.TraitModifierNonContainerSyncData = other.TraitModifierNonContainerSyncData
			t.Features = other.Features.Clone()
		}
	})
}

// Hash writes this object's contents into the hasher. Note that this only hashes the data that is considered to be
// "source" data, i.e. not expected to be modified by the user after copying from a library.
func (t *TraitModifier) Hash(h hash.Hash) {
	t.hash(h)
	if t.Container() {
		// Containers carry no further sync data, so mark them the same way a nil value is marked elsewhere
		xhash.Num8(h, uint8(255))
	} else {
		t.TraitModifierNonContainerSyncData.hash(h)
	}
}

func (t *TraitModifierSyncData) hash(h hash.Hash) {
	xhash.StringWithLen(h, t.Name)
	xhash.StringWithLen(h, t.PageRef)
	xhash.StringWithLen(h, t.PageRefHighlight)
	xhash.StringWithLen(h, t.LocalNotes)
	xhash.Num64(h, len(t.Tags))
	for _, tag := range t.Tags {
		xhash.StringWithLen(h, tag)
	}
}

func (t *TraitModifierNonContainerSyncData) hash(h hash.Hash) {
	xhash.StringWithLen(h, t.CostAdj)
	xhash.Bool(h, t.UseLevelFromTrait)
	xhash.Bool(h, t.CostIgnoresLevel)
	xhash.Bool(h, t.ShowNotesOnWeapon)
	xhash.Num8(h, t.Affects)
	xhash.Num64(h, len(t.Features))
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
	t.Tags = slices.Clone(other.Tags)
	t.Replacements = maps.Clone(other.Replacements)
	t.Features = other.Features.Clone()
}
