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
	"fmt"
	"hash"
	"io/fs"
	"maps"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/equipmentsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/maxusesmod"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/skillsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xhash"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison/enums/align"
)

var (
	_ = assertNode[*Equipment]
	_ = assertEditorData[*EquipmentEditData]

	_ WeaponOwner       = &Equipment{}
	_ LeveledOwner      = &Equipment{}
	_ TechLevelProvider = &Equipment{}
	_ FeatureSwitcher   = &Equipment{}

	_ TemplatePickerProvider = &EquipmentData{}
	_ TemplatePickerProvider = &EquipmentEditData{}
	_ TemplatePickerProvider = &EquipmentContainerOnlySyncData{}
)

// Columns that can be used with the equipment method .CellData()
const (
	EquipmentEquippedColumn = iota
	EquipmentQuantityColumn
	EquipmentDescriptionColumn
	_ // Was EquipmentUsesColumn
	EquipmentTLColumn
	EquipmentLCColumn
	EquipmentCostColumn
	EquipmentExtendedCostColumn
	EquipmentWeightColumn
	EquipmentExtendedWeightColumn
	EquipmentTagsColumn
	EquipmentReferenceColumn
	EquipmentLibSrcColumn
	EquipmentSwitchColumn
)

// MaxEquipmentMaxUses is the largest permitted resolved value for an Equipment's maximum uses.
const MaxEquipmentMaxUses = 9999999

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
	EquipmentContainerOnlySyncData
	VTTNotes     string               `json:"vtt_notes,omitzero"`
	Replacements map[string]string    `json:"replacements,omitempty"`
	Modifiers    []*EquipmentModifier `json:"modifiers,omitempty"`
	RatedST      fxp.Int              `json:"rated_strength,omitzero"`
	Quantity     fxp.Int              `json:"quantity"`
	Level        fxp.Int              `json:"level,omitzero"`
	Uses         int                  `json:"uses,omitzero"`
	Equipped     bool                 `json:"equipped,omitzero"`
	ItemSwitch
	preconfigurable
}

// EquipmentSyncData holds the equipment sync data that is common to both containers and non-containers.
type EquipmentSyncData struct {
	Name                   string      `json:"description,omitzero"`
	PageRef                string      `json:"reference,omitzero"`
	PageRefHighlight       string      `json:"reference_highlight,omitzero"`
	LocalNotes             string      `json:"local_notes,omitzero"`
	TechLevel              string      `json:"tech_level,omitzero"`
	LegalityClass          string      `json:"legality_class,omitzero"`
	Tags                   []string    `json:"tags,omitempty"`
	BaseValue              string      `json:"base_value,omitzero"`
	BaseWeight             string      `json:"base_weight,omitzero"`
	MaxUses                int         `json:"max_uses,omitzero"`
	Prereq                 *PrereqList `json:"prereqs,omitzero"`
	Weapons                []*Weapon   `json:"weapons,omitempty"`
	Features               Features    `json:"features,omitempty"`
	WeightIgnoredForSkills bool        `json:"ignore_weight_for_skills,omitzero"`
}

// EquipmentContainerOnlySyncData holds the skill sync data that is only applicable to equipment that are containers.
type EquipmentContainerOnlySyncData struct {
	TemplatePicker *TemplatePicker `json:"template_picker,omitzero"`
}

// TemplatePickerData returns the TemplatePicker data, if any.
func (e *EquipmentContainerOnlySyncData) TemplatePickerData() *TemplatePicker {
	return e.TemplatePicker
}

// SetTemplatePickerData sets the TemplatePicker data.
func (e *EquipmentContainerOnlySyncData) SetTemplatePickerData(tp *TemplatePicker) {
	e.TemplatePicker = tp
}

func (e *EquipmentContainerOnlySyncData) hash(h hash.Hash) {
	e.TemplatePicker.Hash(h)
}

type equipmentListData struct {
	Version int          `json:"version"`
	Rows    []*Equipment `json:"rows"`
}

// NewEquipmentFromFile loads an Equipment list from a file.
func NewEquipmentFromFile(fileSystem fs.FS, filePath string) ([]*Equipment, error) {
	var data equipmentListData
	if err := jio.LoadFromFile(fileSystem, filePath, &data); err != nil {
		return nil, errs.NewWithCause(InvalidFileData(), err)
	}
	if err := jio.CheckVersion(data.Version); err != nil {
		return nil, err
	}
	// SetDataOwner recurses into children on its own, so only the top-level rows need to be visited. Containers must
	// not be skipped: they carry their own weapons and modifiers, which would otherwise never be attached.
	for _, item := range data.Rows {
		item.SetDataOwner(nil)
	}
	return data.Rows, nil
}

// SaveEquipment writes the Equipment list to the file as JSON.
func SaveEquipment(equipment []*Equipment, filePath string) error {
	AdjustEquipmentUsesForSave(equipment)
	return jio.SaveToFile(filePath, &equipmentListData{
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
	if container {
		e.TemplatePicker = &TemplatePicker{}
	}
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
func (e *Equipment) Clone(from LibraryFile, owner DataOwner, parent *Equipment, mode CloneMode) *Equipment {
	other := NewEquipment(owner, parent, e.Container())
	other.AdjustSource(from, e.SourcedID, mode)
	other.SetOpen(e.IsOpen())
	other.ThirdParty = e.ThirdParty
	other.copyFrom(other, &e.EquipmentEditData, false, mode)
	PropagateNodeNoteClosedState(e, other)
	if e.HasChildren() {
		other.Children = make([]*Equipment, 0, len(e.Children))
		for _, child := range e.Children {
			other.Children = append(other.Children, child.Clone(from, owner, other, mode))
		}
	}
	return other
}

// MarshalJSONTo implements json.MarshalerTo.
func (e *Equipment) MarshalJSONTo(enc *jsontext.Encoder) error {
	type calc struct {
		Value                   fxp.Int     `json:"value"`
		ExtendedValue           fxp.Int     `json:"extended_value"`
		Weight                  fxp.Weight  `json:"weight"`
		ExtendedWeight          fxp.Weight  `json:"extended_weight"`
		ExtendedWeightForSkills *fxp.Weight `json:"extended_weight_for_skills,omitzero"`
		ResolvedNotes           string      `json:"resolved_notes,omitzero"`
		UnsatisfiedReason       string      `json:"unsatisfied_reason,omitzero"`
	}
	e.ClearUnusedFieldsForType()
	if omitCalc(enc) {
		return json.MarshalEncode(enc, &e.EquipmentData)
	}
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
	if e.WeightIgnoredForSkills && e.ReallyEquipped() {
		w := e.ExtendedWeight(true, defUnits)
		data.Calc.ExtendedWeightForSkills = &w
	}
	return json.MarshalEncode(enc, &data)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (e *Equipment) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
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
	if err := json.UnmarshalDecode(dec, &localData); err != nil {
		return err
	}
	setOpen := false
	if !tid.IsValid(localData.TID) {
		// Fixup old data that used UUIDs instead of TIDs
		localData.TID = tid.MustNewTID(equipmentKind(strings.HasSuffix(localData.Type, containerKeyPostfix)))
		setOpen = localData.IsOpen
	}
	e.EquipmentData = localData.EquipmentData
	e.Replacements = nameable.Normalize(e.Replacements)
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
	case EquipmentSwitchColumn:
		data.Title = HeaderSwitch
		data.TitleIsImageKey = true
		data.Detail = SwitchHeaderTooltip()
	}
	return data
}

// CellData returns the cell data information for the given column.
func (e *Equipment) CellData(columnID int, data *CellData) {
	data.Self = e
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
		data.Tooltip = i18n.Text("Click to toggle whether this piece of equipment is equipped or just carried. Items that are not equipped do not apply any features they may normally contribute to the character. Note that if a parent container is not equipped, none of its contents are considered to be equipped either and any checkmark here will be dimmed to reflect this.")
		if !e.ReallyEquipped() {
			data.Dim = true
		}
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
		if !xreflect.IsNil(e.owner) {
			state, _ := e.owner.SourceMatcher().Match(e)
			data.Primary = state.AltString()
			data.Tooltip = state.String()
			if state != srcstate.Custom {
				data.Tooltip += "\n" + e.Source.String()
			}
		}
	case EquipmentSwitchColumn:
		// Only items that actually have something to switch get a cell; the rest are left blank.
		if e.HasSwitchableFeatures() {
			data.Type = cell.Switch
			data.Checked = e.SwitchedOn
			data.Alignment = align.Middle
			data.Tooltip = SwitchCellTooltip(e.Container())
			// Dim (but leave usable) a switch that would change nothing if thrown right now. The character only
			// collects features from carried equipment that is really equipped, and this column is present for the
			// other equipment list as well, where the equipped flag is meaningless -- nothing clears it when an item
			// is created in or moved to that list. Neither state is the whole answer, though: the features the
			// equipment resolves for itself take effect no matter which list it lives in.
			//
			// The tests are ordered by cost, since this runs for every column of every row on each sort and each
			// keystroke of a search: ReallyEquipped only walks the parent chain, IsCarried additionally scans the
			// entity's root equipment lists, and switchMattersWhileUnequipped traverses every modifier, so each is
			// only reached when the cheaper ones ahead of it left the answer open.
			if (!e.ReallyEquipped() || !e.IsCarried()) && !e.switchMattersWhileUnequipped() {
				data.Dim = true
			}
		}
	}
}

// ReallyEquipped returns true if this equipment is equipped and has a quantity > 0 and all of its parents do too.
func (e *Equipment) ReallyEquipped() bool {
	if !e.Equipped || e.Quantity <= 0 {
		return false
	}
	p := e.parent
	for p != nil {
		if !p.Equipped || p.Quantity <= 0 {
			return false
		}
		p = p.parent
	}
	return true
}

// IsCarried returns true if this equipment is not rooted in the other equipment list of the entity that owns it. The
// entity only collects features from carried equipment (see Entity.processFeatures), so an item in the other equipment
// list contributes nothing to the character no matter what its equipped state says -- and that state can easily say
// "equipped", since new equipment starts out equipped and nothing clears the flag when an item is created in or moved
// to the other list. Equipment with no owning entity -- a library list, a template or a loot sheet -- has no second
// list to be told apart from, so it is considered carried.
//
// A row on a sheet is always rooted in one of the two lists, and the other equipment list is checked first, since it is
// normally the far shorter of the two. Only a root found in neither list needs any further work, and there are two ways
// to get one. An editor's working clone is made with the row's own parent, which is nil for a top-level row, leaving
// the clone rooted outside both lists while keeping the ID of the row it stands for; that ID is what settles the
// question, so the clone's preview of Extended Value and Extended Weight agrees with the sheet no matter which list the
// row came from. A row in flight between the lists has no such counterpart to be found and is treated as carried, which
// mirrors the no-entity case.
func (e *Equipment) IsCarried() bool {
	entity := EntityFromNode(e)
	if entity == nil {
		return true
	}
	root := e
	for root.parent != nil {
		root = root.parent
	}
	if slices.Contains(entity.OtherEquipment, root) {
		return false
	}
	if slices.Contains(entity.CarriedEquipment, root) {
		return true
	}
	// An orphan root: look for the row it stands for in the other equipment list, at any depth. No matching lookup in
	// the carried equipment list is needed, since "carried" is the answer whenever nothing says otherwise.
	standsForOther := false
	Traverse(func(other *Equipment) bool {
		standsForOther = other.TID == root.TID
		return standsForOther
	}, false, false, entity.OtherEquipment...)
	return !standsForOther
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
			fmt.Fprintf(&localBuffer, i18n.Text("Rated ST %s"), e.RatedST.Comma())
		}
		if maxUses := e.ResolvedMaxUses(); maxUses > 0 {
			if localBuffer.Len() != 0 {
				localBuffer.WriteString("; ")
			}
			fmt.Fprintf(&localBuffer, i18n.Text("%s of %s uses left"), xstrings.CommaInt(min(e.Uses, maxUses)),
				xstrings.CommaInt(maxUses))
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
	buffer.WriteString(e.NameWithReplacements())
	if e.IsLeveled() {
		buffer.WriteByte(' ')
		buffer.WriteString(e.Level.String())
	}
	return buffer.String()
}

// ResolveLocalNotes resolves the local notes, running any embedded scripts to get the final result.
func (e *Equipment) ResolveLocalNotes() string {
	return ResolveText(EntityFromNode(e), deferredNewScriptEquipment(e), e.LocalNotesWithReplacements())
}

// Notes returns the local notes.
func (e *Equipment) Notes() string {
	return e.LocalNotesWithReplacements()
}

// ActiveFeatures returns the features of this equipment that currently take effect, i.e. all of them except any
// switchable ones while the equipment's switch is off. Features of the equipment's modifiers are not included.
func (e *Equipment) ActiveFeatures() Features {
	return e.Features.Active(e.SwitchedOn)
}

// HasSwitchableFeatures implements FeatureSwitcher.
func (e *Equipment) HasSwitchableFeatures() bool {
	if e.Features.AnySwitchable() {
		return true
	}
	return anyModifierSwitchable(e.Modifiers, func(mod *EquipmentModifier) Features { return mod.Features })
}

// switchMattersWhileUnequipped returns true if any of the switchable features this equipment currently contributes --
// its own or those of its enabled modifiers -- would still take effect while the entity isn't collecting from the
// equipment, i.e. while it isn't really equipped or isn't carried at all. The owning entity only collects features
// from carried equipment that is really equipped (see Entity.processFeatures), but a handful of features are resolved
// by the equipment itself, no matter which list it lives in or whether it is equipped, so the switch controlling one
// of those is never inert.
func (e *Equipment) switchMattersWhileUnequipped() bool {
	if e.anySwitchableMattersWhileUnequipped(e.Features) {
		return true
	}
	return anyEnabledNonContainerModifier(e.Modifiers, func(mod *EquipmentModifier) bool {
		return e.anySwitchableMattersWhileUnequipped(mod.Features)
	})
}

// anySwitchableMattersWhileUnequipped returns true if any of the given features is switchable and is one this
// equipment resolves for itself rather than one that only reaches the character through the owning entity.
func (e *Equipment) anySwitchableMattersWhileUnequipped(features Features) bool {
	for _, one := range features {
		if !one.IsSwitchable() {
			continue
		}
		switch f := one.(type) {
		case *ContainedWeightReduction:
			// Applied by ContainedWeightAdjustedForModifiers, which runs for every container in both the carried and
			// the other equipment lists -- but there is nothing to reduce without contents.
			if len(e.Children) != 0 {
				return true
			}
		case *EquipmentMaxUsesBonus:
			// "To this equipment" bonuses are applied by ResolvedMaxUses, which the equipment resolves for itself.
			if f.SelectionType == equipmentsel.ThisEquipment {
				return true
			}
		case *WeaponBonus:
			// "To this weapon" bonuses are resolved by the weapon itself, and the weapons of unequipped equipment are
			// still displayed when the sheet is set to show them all. "To a named weapon" bonuses count, too: the
			// entity's own collection only reaches really-equipped items, but the weapon additionally reads its
			// owner's active features directly, no matter where that owner lives, so such a bonus still lands on this
			// equipment's own weapon whenever the name, usage and tag criteria match it.
			if (f.SelectionType == wsel.ThisWeapon || f.SelectionType == wsel.WithName) && len(e.Weapons) != 0 {
				return true
			}
		case *SkillBonus:
			// Likewise for a skill bonus aimed at this equipment's own weapons.
			if f.SelectionType == skillsel.ThisWeapon && len(e.Weapons) != 0 {
				return true
			}
		}
	}
	return false
}

// TagList returns the list of tags.
func (e *Equipment) TagList() []string {
	return e.Tags
}

// RatedStrength returns the rated ST for this equipment, or 0 if it has none.
func (e *Equipment) RatedStrength() fxp.Int {
	return e.RatedST
}

// BaseValueWithReplacements returns the base value with any replacements applied.
func (e *Equipment) BaseValueWithReplacements() string {
	return nameable.Apply(e.BaseValue, e.Replacements)
}

// ResolvedBaseValue resolves the base value, running any embedded scripts to get the final result.
func (e *Equipment) ResolvedBaseValue() fxp.Int {
	return ResolveToNumber(EntityFromNode(e), deferredNewScriptEquipment(e), e.BaseValueWithReplacements()).
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
	return ResolveToWeight(entity, deferredNewScriptEquipment(e), e.BaseWeightWithReplacements(),
		SheetSettingsFor(entity).DefaultWeightUnits)
}

// AdjustedWeight returns the weight after adjustments for any modifiers. Does not include the weight of children.
func (e *Equipment) AdjustedWeight(forSkills bool, defUnits fxp.WeightUnit) fxp.Weight {
	if forSkills && e.WeightIgnoredForSkills && e.ReallyEquipped() {
		return 0
	}
	return WeightAdjustedForModifiers(e, e.ResolvedBaseWeight(), e.Modifiers, defUnits)
}

// ExtendedWeight returns the extended weight.
func (e *Equipment) ExtendedWeight(forSkills bool, defUnits fxp.WeightUnit) fxp.Weight {
	return ExtendedWeightAdjustedForModifiers(e, defUnits, e.Quantity, e.ResolvedBaseWeight(), e.Modifiers, e.Features, e.Children, forSkills, e.WeightIgnoredForSkills && e.ReallyEquipped())
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
	base += fxp.Int(ContainedWeightAdjustedForModifiers(equipment, defUnits, modifiers, features, children, forSkills))
	return fxp.Weight(base.Mul(qty))
}

// ContainedWeight returns the weight of the contents of this equipment, after any contained weight reductions have been
// applied. This is the weight held by a single instance of this equipment, so it is unaffected by the quantity, and
// does not include the weight of the equipment itself.
func (e *Equipment) ContainedWeight(forSkills bool, defUnits fxp.WeightUnit) fxp.Weight {
	return ContainedWeightAdjustedForModifiers(e, defUnits, e.Modifiers, e.Features, e.Children, forSkills)
}

// ContainedWeightAdjustedForModifiers calculates the weight of the contents of a container, after applying any
// contained weight reductions supplied by the features and modifiers. This is the weight held by a single instance of
// the container and does not include the weight of the container itself.
func ContainedWeightAdjustedForModifiers(equipment *Equipment, defUnits fxp.WeightUnit, modifiers []*EquipmentModifier, features Features, children []*Equipment, forSkills bool) fxp.Weight {
	if len(children) == 0 {
		return 0
	}
	var contained fxp.Int
	for _, one := range children {
		contained += fxp.Int(one.ExtendedWeight(forSkills, defUnits))
	}
	// Switchable reductions, whether on the equipment or its modifiers, only apply while the equipment's switch is on.
	switchedOn := equipment != nil && equipment.SwitchedOn
	var percentage, reduction fxp.Int
	for _, one := range features.Active(switchedOn) {
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
		for _, f := range mod.Features.Active(switchedOn) {
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
	return fxp.Weight((contained - reduction).Max(0))
}

// ResolvedMaxUses returns the MaxUses adjusted by any applicable EquipmentMaxUsesBonus features, clamped to the range
// [0, MaxEquipmentMaxUses]. "This equipment" bonuses attached to this item or its enabled modifiers are always applied;
// "equipment whose name" bonuses are gathered from the owning entity, if there is one.
func (e *Equipment) ResolvedMaxUses() int {
	addition := fxp.Int(0)
	percentage := fxp.Int(0)
	multiplier := fxp.One
	have := false
	apply := func(bonus *EquipmentMaxUsesBonus) {
		have = true
		amount := bonus.AdjustedAmount()
		switch bonus.Operation() {
		case maxusesmod.Percentage:
			percentage += amount
		case maxusesmod.Multiplier:
			if amount <= 0 {
				amount = fxp.One
			}
			multiplier = multiplier.Mul(amount)
		default: // maxusesmod.Addition
			addition += amount
		}
	}
	applyThisEquipment := func(features Features) {
		for _, f := range features {
			if bonus, ok := f.(*EquipmentMaxUsesBonus); ok && bonus.SelectionType == equipmentsel.ThisEquipment {
				// The level driving a per-level bonus comes from the item the bonus is attached to, matching how
				// Entity.processFeatures assigns the leveled owner for equipment features.
				bonus.SetLeveledOwner(e)
				apply(bonus)
			}
		}
	}
	applyThisEquipment(e.ActiveFeatures())
	Traverse(func(mod *EquipmentModifier) bool {
		applyThisEquipment(mod.Features.Active(e.SwitchedOn))
		return false
	}, true, true, e.Modifiers...)
	if entity := EntityFromNode(e); entity != nil {
		for _, bonus := range entity.EquipmentMaxUsesBonusesFor(e.NameWithReplacements(), e.TagList(), nil) {
			apply(bonus)
		}
	}
	if !have {
		return e.MaxUses
	}
	result := fxp.FromInteger(e.MaxUses) + addition
	result += result.Mul(percentage).Div(fxp.Hundred)
	result = result.Mul(multiplier)
	return result.Max(0).Min(fxp.FromInteger(MaxEquipmentMaxUses)).AsInteger[int]()
}

// ResolvedUses returns the current Uses capped at ResolvedMaxUses. This is the value that should be displayed; the
// stored Uses value itself is intentionally left unchanged (a feature that lowers the maximum below the stored Uses
// only affects what is shown) until the user edits it or the data is saved.
func (e *Equipment) ResolvedUses() int {
	if maxUses := e.ResolvedMaxUses(); e.Uses > maxUses {
		return maxUses
	}
	return e.Uses
}

// AdjustUsesToResolvedMax caps the stored Uses at the ResolvedMaxUses value, bringing a Uses value that a feature has
// pushed above the maximum back into range. This mutates the stored value and is intended to be called just before
// saving.
func (e *Equipment) AdjustUsesToResolvedMax() {
	if maxUses := e.ResolvedMaxUses(); e.Uses > maxUses {
		e.Uses = maxUses
	}
}

// AdjustEquipmentUsesForSave caps the stored Uses of every piece of equipment in the list (including its children) at
// its ResolvedMaxUses. Call this on a data object's equipment just before saving it to disk.
func AdjustEquipmentUsesForSave(list []*Equipment) {
	Traverse(func(e *Equipment) bool {
		e.AdjustUsesToResolvedMax()
		return false
	}, false, false, list...)
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
	nameable.Extract(
		m, existing,
		e.Name,
		e.LocalNotes,
		e.BaseValue,
		e.BaseWeight,
	)
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
	e.ClearUnusedFieldsForType()
}

// ConvertToNonContainer converts this node to a non-container.
func (e *Equipment) ConvertToNonContainer() {
	e.TID = tid.TID(kinds.Equipment) + e.TID[1:]
	e.ClearUnusedFieldsForType()
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
	if e.Container() {
		if e.TemplatePicker == nil {
			e.TemplatePicker = &TemplatePicker{}
		}
	} else {
		e.Children = nil
		e.EquipmentContainerOnlySyncData = EquipmentContainerOnlySyncData{}
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
	if !xreflect.IsNil(e.owner) {
		if state, data := e.owner.SourceMatcher().Match(e); state == srcstate.Mismatched {
			if other, ok := data.(*Equipment); ok {
				e.EquipmentSyncData = other.EquipmentSyncData
				e.Tags = slices.Clone(other.Tags)
				if e.Container() {
					e.TemplatePicker = other.TemplatePicker.Clone()
				}
				e.Prereq = other.Prereq.CloneResolvingEmpty(false, true)
				e.Weapons = CloneWeapons(other.Weapons, e, Reference)
				e.Features = other.Features.Clone()
			}
		}
	}
}

// Hash writes this object's contents into the hasher. Note that this only hashes the data that is considered to be
// "source" data, i.e. not expected to be modified by the user after copying from a library.
func (e *Equipment) Hash(h hash.Hash) {
	e.EquipmentSyncData.hash(h)
	if e.Container() {
		e.EquipmentContainerOnlySyncData.hash(h)
	}
}

func (e *EquipmentSyncData) hash(h hash.Hash) {
	if e == nil {
		return
	}
	xhash.StringWithLen(h, e.Name)
	xhash.StringWithLen(h, e.PageRef)
	xhash.StringWithLen(h, e.PageRefHighlight)
	xhash.StringWithLen(h, e.LocalNotes)
	xhash.StringWithLen(h, e.TechLevel)
	xhash.StringWithLen(h, e.LegalityClass)
	xhash.Num64(h, len(e.Tags))
	for _, tag := range e.Tags {
		xhash.StringWithLen(h, tag)
	}
	xhash.StringWithLen(h, e.BaseValue)
	xhash.StringWithLen(h, e.BaseWeight)
	xhash.Num64(h, e.MaxUses)
	e.Prereq.Hash(h)
	xhash.Num64(h, len(e.Weapons))
	for _, weapon := range e.Weapons {
		weapon.Hash(h)
	}
	xhash.Num64(h, len(e.Features))
	for _, feature := range e.Features {
		feature.Hash(h)
	}
	xhash.Bool(h, e.WeightIgnoredForSkills)
}

// CopyFrom implements node.EditorData.
func (e *EquipmentEditData) CopyFrom(other *Equipment) {
	e.copyFrom(other, &other.EquipmentEditData, false, Copy)
}

// SetNameableReplacements sets the replacements to be used with Nameables.
func (e *EquipmentEditData) SetNameableReplacements(replacements map[string]string) {
	e.Replacements = replacements
}

// ApplyTo implements node.EditorData.
func (e *EquipmentEditData) ApplyTo(other *Equipment) {
	other.copyFrom(other, e, true, Copy)
}

// copyFrom copies other into e. isApply distinguishes staging the editor's working copy from committing it
// back, and only affects how an empty Prereq list is resolved. mode controls how nested modifiers and
// weapons are cloned -- CopyFrom/ApplyTo above always use Copy, since that's staging or committing the same
// equipment's own data, not producing a new one; Clone passes its own mode through.
func (e *EquipmentEditData) copyFrom(equipment *Equipment, other *EquipmentEditData, isApply bool, mode CloneMode) {
	*e = *other
	e.Tags = slices.Clone(other.Tags)
	e.Replacements = maps.Clone(other.Replacements)
	e.Modifiers = nil
	if len(other.Modifiers) != 0 {
		e.Modifiers = make([]*EquipmentModifier, 0, len(other.Modifiers))
		for _, one := range other.Modifiers {
			// The LibraryFile for this clone must come from the parent equipment rather than
			// from `one`. This covers the case where the source data *is* the authoritative
			// source and therefore carries no source information of its own. `one.Source.LibraryFile`
			// is empty and `AdjustSource` won't set source data on the copy. `equipment.Source.LibraryFile`
			// holds the already-adjusted source for the equipment copy, so it always has the
			// correct library path.
			//
			// Background: when GCS clones an item from one library into another location (as
			// opposed to duplicating in place), it passes the *source* library as the first
			// argument to `Clone`. That path, combined with the IDs from the source nodes, is
			// what builds the `source` values for the clone.
			cloned := one.Clone(equipment.Source.LibraryFile, equipment.owner, nil, mode)
			// Point the copy at the equipment it belongs to, so that its nameable placeholders can be resolved with
			// that equipment's replacements. Without this, the copies held in an editor show their raw placeholders
			// (e.g. "@Material@"), since the accessors fall back to the unsubstituted text when there is no equipment.
			cloned.setEquipment(equipment)
			e.Modifiers = append(e.Modifiers, cloned)
		}
		// setEquipment() migrates a modifier's legacy replacements into the equipment it was pointed at, which isn't
		// the holder of this data when an editor is being populated, so pick up anything it added. This is a no-op
		// when this data is the equipment's own, since both maps are then the same one.
		for k, v := range equipment.Replacements {
			if _, exists := e.Replacements[k]; !exists {
				if e.Replacements == nil {
					e.Replacements = make(map[string]string)
				}
				e.Replacements[k] = v
			}
		}
	}
	e.Prereq = e.Prereq.CloneResolvingEmpty(false, isApply)
	e.Weapons = CloneWeapons(other.Weapons, equipment, mode)
	e.Features = other.Features.Clone()
	e.TemplatePicker = other.TemplatePicker.Clone()
}

// CanPreconfigureContainer implements Preconfigurable.
func (e *EquipmentEditData) CanPreconfigureContainer() bool {
	return true
}
