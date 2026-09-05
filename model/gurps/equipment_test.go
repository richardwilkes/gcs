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
	"os"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestEquipmentRatedStrength verifies that equipment reports its user-settable rated ST. Unlike traits, skills and
// spells -- whose RatedStrength always returns 0 -- equipment genuinely carries one, and weapon strength calculations
// depend on getting it back.
func TestEquipmentRatedStrength(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	equipment := NewEquipment(e, nil, false)
	c.Equal(fxp.Int(0), equipment.RatedStrength(), "equipment with no rated ST reports 0")
	equipment.RatedST = fxp.FromInteger(12)
	c.Equal(fxp.FromInteger(12), equipment.RatedStrength(), "equipment reports the rated ST it was given")
}

// TestNewEquipmentFromFileAttachesContainerData verifies that loading an equipment list attaches the owner to
// container rows as well as to leaf rows, so that a container's own weapons get their sub-version fixups applied and
// its modifiers can resolve nameable placeholders.
func TestNewEquipmentFromFileAttachesContainerData(t *testing.T) {
	c := check.New(t)
	// container_with_own_data.eqp holds a leveled container that carries a weapon written before weapon
	// sub-versioning existed (no "sv" key) and a modifier with a nameable placeholder, plus a child item with the
	// same content.
	rows, err := NewEquipmentFromFile(os.DirFS("testdata"), "container_with_own_data.eqp")
	c.NoError(err)
	c.Equal(1, len(rows), "the list has a single top-level row")
	if len(rows) != 1 {
		return
	}
	container := rows[0]
	c.Equal(true, container.Container(), "the top-level row is a container")
	c.Equal(1, len(container.Children), "the container has a child")
	if len(container.Weapons) != 1 || len(container.Modifiers) != 1 || len(container.Children) != 1 {
		return
	}
	child := container.Children[0]

	// The container's own weapon must be attached and migrated exactly like the child's.
	for _, one := range []*Equipment{container, child} {
		w := one.Weapons[0]
		c.Equal(WeaponOwner(one), w.Owner, one.Name+": the weapon's owner is set")
		c.Equal(currentWeaponSubVersion, w.SubVersion, one.Name+": the weapon is stamped with the current sub-version")
		c.Equal("", w.Damage.Base, one.Name+": the leveled-damage migration cleared the flat base damage")
		c.Equal("1d", w.Damage.BaseLeveled, one.Name+": the leveled-damage migration moved the base damage")
		c.Equal(w, w.Damage.Owner, one.Name+": the weapon's damage points back at the weapon")
	}

	// The container's own modifiers must be attached so nameable placeholders resolve.
	c.Equal("Oak plating", container.Modifiers[0].NameWithReplacements(),
		"the container's modifier resolves the container's replacements")
	c.Equal("Iron plating", child.Modifiers[0].NameWithReplacements(),
		"the child's modifier resolves the child's replacements")
}

// TestEquipmentEditDataResolvesModifierNameables verifies that the modifier copies an equipment editor works with are
// pointed at their equipment, so their nameable placeholders resolve. Without that, the accessors fall back to the raw
// text and the editor's modifier rows read "@Material@" instead of the replacement the user chose.
func TestEquipmentEditDataResolvesModifierNameables(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	eqp := NewEquipment(entity, nil, false)
	eqp.Name = "Armor"
	eqp.Replacements = map[string]string{"Material": "Steel"}
	mod := NewEquipmentModifier(entity, nil, false)
	mod.Name = "@Material@ plating"
	mod.LocalNotes = "Forged from @Material@"
	eqp.Modifiers = []*EquipmentModifier{mod}

	// What the editor is handed when it opens.
	var edit EquipmentEditData
	edit.CopyFrom(eqp)
	c.Equal(1, len(edit.Modifiers))
	c.True(eqp == edit.Modifiers[0].Target(), "the copy points at the equipment being edited")
	c.Equal("Steel plating", edit.Modifiers[0].NameWithReplacements())
	c.Equal("Forged from Steel", edit.Modifiers[0].LocalNotesWithReplacements())

	// The copy must be a copy: editing it leaves the original alone.
	edit.Modifiers[0].Name = "@Material@ mesh"
	c.Equal("@Material@ plating", mod.Name, "the original modifier is untouched")

	// What applying the editor's data back produces.
	target := NewEquipment(entity, nil, false)
	edit.ApplyTo(target)
	c.Equal(1, len(target.Modifiers))
	c.True(target == target.Modifiers[0].Target(), "the applied copy points at the equipment it was applied to")
	c.Equal("Steel mesh", target.Modifiers[0].NameWithReplacements())
}

// TestEquipmentCloneAttachesModifiersToTheClone verifies that a clone's modifier copies are pointed at the clone rather
// than at the equipment they were cloned from, so that they resolve their nameable placeholders with the clone's
// replacements. Duplicating a nested row in a standalone equipment list never re-attaches the data owner, so a
// mispointed modifier would keep rendering from the original's replacements for the rest of the session.
func TestEquipmentCloneAttachesModifiersToTheClone(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	eqp := NewEquipment(entity, nil, false)
	eqp.Name = "Armor"
	eqp.Replacements = map[string]string{"Material": "Steel"}
	mod := NewEquipmentModifier(entity, nil, true)
	mod.Name = "@Material@ plating"
	child := NewEquipmentModifier(entity, mod, false)
	child.Name = "@Material@ rivets"
	mod.Children = []*EquipmentModifier{child}
	eqp.Modifiers = []*EquipmentModifier{mod}
	eqp.SetDataOwner(entity)

	clone := eqp.Clone(LibraryFile{}, entity, nil, Reference)
	c.Equal(1, len(clone.Modifiers))
	if len(clone.Modifiers) != 1 || len(clone.Modifiers[0].Children) != 1 {
		return
	}
	c.True(clone == clone.Modifiers[0].Target(), "the clone's modifier points at the clone")
	c.True(clone == clone.Modifiers[0].Children[0].Target(), "the clone's child modifier points at the clone")

	// Diverging the clone's replacements must move the clone's modifiers only.
	clone.Replacements["Material"] = "Bronze"
	c.Equal("Bronze plating", clone.Modifiers[0].NameWithReplacements())
	c.Equal("Bronze rivets", clone.Modifiers[0].Children[0].NameWithReplacements())
	c.Equal("Steel plating", eqp.Modifiers[0].NameWithReplacements(), "the original's modifier is unaffected")
	c.Equal("Steel rivets", eqp.Modifiers[0].Children[0].NameWithReplacements(),
		"the original's child modifier is unaffected")
}

// TestEquipmentCloneMigratesLegacyModifierReplacements verifies that cloning equipment whose modifiers still carry the
// legacy per-modifier replacements migrates them into the clone rather than into the equipment being cloned.
func TestEquipmentCloneMigratesLegacyModifierReplacements(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	eqp := NewEquipment(entity, nil, false)
	eqp.Name = "Armor"
	mod := NewEquipmentModifier(entity, nil, false)
	mod.Name = "@Material@ plating"
	mod.Replacements = map[string]string{"Material": "Steel"}
	eqp.Modifiers = []*EquipmentModifier{mod}

	clone := eqp.Clone(LibraryFile{}, entity, nil, Reference)
	c.Equal(0, len(eqp.Replacements), "the equipment being cloned isn't mutated")
	c.Equal(1, len(clone.Modifiers))
	if len(clone.Modifiers) != 1 {
		return
	}
	c.Equal("Steel", clone.Replacements["Material"], "the legacy replacements were migrated into the clone")
	c.Equal("Steel plating", clone.Modifiers[0].NameWithReplacements())

	// The original's modifier still carries its own replacements, so it resolves once it is attached normally.
	eqp.SetDataOwner(entity)
	c.Equal("Steel plating", eqp.Modifiers[0].NameWithReplacements())
}

// TestEquipmentEditDataCapturesMigratedReplacements verifies that populating an editor from equipment whose modifiers
// still carry the legacy per-modifier replacements captures the migrated replacements in the editor's snapshot.
// Without that, applying the editor's data back writes the pre-migration snapshot and silently drops them.
func TestEquipmentEditDataCapturesMigratedReplacements(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	eqp := NewEquipment(entity, nil, false)
	eqp.Name = "Armor"
	mod := NewEquipmentModifier(entity, nil, false)
	mod.Name = "@Material@ plating"
	mod.Replacements = map[string]string{"Material": "Steel"}
	eqp.Modifiers = []*EquipmentModifier{mod}

	var edit EquipmentEditData
	edit.CopyFrom(eqp)
	c.Equal("Steel", edit.Replacements["Material"], "the editor's snapshot has the migrated replacements")

	edit.ApplyTo(eqp)
	c.Equal("Steel", eqp.Replacements["Material"], "applying the editor's data preserves the migrated replacements")
	c.Equal(1, len(eqp.Modifiers))
	if len(eqp.Modifiers) != 1 {
		return
	}
	c.Equal("Steel plating", eqp.Modifiers[0].NameWithReplacements())
}

// TestEquipmentCellDataAppliesDisplayFormatsOnPagesOnly verifies that the sheet's equipment weight and value display
// formats are applied only when the cell is being displayed on a page, with the exact value offered as a tooltip, and
// that the same equipment renders exactly everywhere else (editors, library lists, and sort requests). The value format
// pads with zeros and the values are chosen to need that padding, so that dropping the padding from the value path
// would be caught.
func TestEquipmentCellDataAppliesDisplayFormatsOnPagesOnly(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	entity.SheetSettings.EquipmentWeightFormat = fxp.NumberFormat{Places: fxp.TwoPlaces}
	entity.SheetSettings.EquipmentValueFormat = fxp.NumberFormat{Places: fxp.TwoPlaces, PadWithZeros: true}
	equipment := NewEquipment(entity, nil, false)
	equipment.BaseWeight = "7.5127 lb"
	equipment.BaseValue = "1234.5"
	equipment.Quantity = fxp.Two
	entity.CarriedEquipment = append(entity.CarriedEquipment, equipment)

	for _, d := range []struct {
		column      int
		exact       string
		forPageText string
		forPageTip  string
		description string
	}{
		{EquipmentWeightColumn, "7.5127 lb", "7.51 lb", "7.5127 lb", "weight"},
		{EquipmentExtendedWeightColumn, "15.0254 lb", "15.03 lb", "15.0254 lb", "extended weight"},
		{EquipmentCostColumn, "1,234.5", "1,234.50", "1,234.5", "value"},
		{EquipmentExtendedCostColumn, "2,469", "2,469.00", "2,469", "extended value"},
	} {
		var data CellData
		equipment.CellData(d.column, &data)
		c.Equal(d.exact, data.Primary, "%s off the page", d.description)
		c.Equal("", data.Tooltip, "%s off the page has no tooltip", d.description)
		data = CellData{ForPage: true}
		equipment.CellData(d.column, &data)
		c.Equal(d.forPageText, data.Primary, "%s on the page", d.description)
		c.Equal(d.forPageTip, data.Tooltip, "%s on the page offers the exact value", d.description)
		c.True(data.ForPage, "%s leaves the input flag alone", d.description)
	}

	// A value that rounds to itself needs no tooltip.
	equipment.BaseWeight = "7.5 lb"
	data := CellData{ForPage: true}
	equipment.CellData(EquipmentWeightColumn, &data)
	c.Equal("7.5 lb", data.Primary)
	c.Equal("", data.Tooltip)
	entity.SheetSettings.EquipmentWeightFormat.PadWithZeros = true
	data = CellData{ForPage: true}
	equipment.CellData(EquipmentWeightColumn, &data)
	c.Equal("7.50 lb", data.Primary)
	c.Equal("7.5 lb", data.Tooltip)
}

// TestEquipmentHeaderDataAppliesDisplayFormats verifies that the totals in the carried and other equipment headers on
// a page use the sheet's equipment weight and value display formats, and that whenever those formats change how a
// total reads, the exact totals are offered as the header's tooltip. The totals are chosen so that the padding, not
// just the rounding, is what changes them.
func TestEquipmentHeaderDataAppliesDisplayFormats(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	carried := NewEquipment(entity, nil, false)
	carried.BaseWeight = "8 lb"
	carried.BaseValue = "1234.5678"
	entity.CarriedEquipment = append(entity.CarriedEquipment, carried)
	other := NewEquipment(entity, nil, false)
	other.BaseWeight = "0.5 lb"
	other.BaseValue = "0.5"
	entity.OtherEquipment = append(entity.OtherEquipment, other)

	header := EquipmentHeaderData(EquipmentDescriptionColumn, entity, true, true)
	c.Equal("Carried Equipment (8 lb; $1,234.5678)", header.Title)
	c.Equal("", header.Detail, "with no display preference, the title already holds the exact totals")
	header = EquipmentHeaderData(EquipmentDescriptionColumn, entity, false, true)
	c.Equal("Other Equipment (0.5 lb; $0.5)", header.Title)
	c.Equal("", header.Detail)

	entity.SheetSettings.DefaultWeightUnits = fxp.Kilogram
	entity.SheetSettings.EquipmentWeightFormat = fxp.NumberFormat{Places: fxp.OnePlace, PadWithZeros: true}
	entity.SheetSettings.EquipmentValueFormat = fxp.NumberFormat{Places: fxp.TwoPlaces, PadWithZeros: true}
	header = EquipmentHeaderData(EquipmentDescriptionColumn, entity, true, true)
	c.Equal("Carried Equipment (4.0 kg; $1,234.57)", header.Title)
	c.Equal("4 kg; $1,234.5678", header.Detail, "the exact totals are offered when the formats change them")
	header = EquipmentHeaderData(EquipmentDescriptionColumn, entity, false, true)
	c.Equal("Other Equipment (0.3 kg; $0.50)", header.Title)
	c.Equal("0.25 kg; $0.5", header.Detail)

	// Totals the formats leave as they are need no tooltip, even with formats in effect.
	entity.SheetSettings.DefaultWeightUnits = fxp.Pound
	carried.BaseWeight = "4.5 lb"
	carried.BaseValue = "1234.56"
	header = EquipmentHeaderData(EquipmentDescriptionColumn, entity, true, true)
	c.Equal("Carried Equipment (4.5 lb; $1,234.56)", header.Title)
	c.Equal("", header.Detail)

	c.Equal("Equipment", EquipmentHeaderData(EquipmentDescriptionColumn, entity, true, false).Title,
		"off the page, the header carries no totals")
	c.Equal("", EquipmentHeaderData(EquipmentDescriptionColumn, entity, true, false).Detail)
}

// TestEquipmentDisplayFormatsWithoutEntityUseDefaultSheetSettings verifies that equipment which belongs to no entity
// -- the rows of a template or a loot sheet -- takes its display formats from the default sheet settings, both for its
// cells and for the page header totals, since that is where those pages' weight units come from as well.
func TestEquipmentDisplayFormatsWithoutEntityUseDefaultSheetSettings(t *testing.T) {
	c := check.New(t)
	global := GlobalSettings().Sheet
	savedUnits := global.DefaultWeightUnits
	savedWeightFormat := global.EquipmentWeightFormat
	savedValueFormat := global.EquipmentValueFormat
	t.Cleanup(func() {
		global.DefaultWeightUnits = savedUnits
		global.EquipmentWeightFormat = savedWeightFormat
		global.EquipmentValueFormat = savedValueFormat
	})
	global.DefaultWeightUnits = fxp.Pound
	global.EquipmentWeightFormat = fxp.NumberFormat{Places: fxp.OnePlace, PadWithZeros: true}
	global.EquipmentValueFormat = fxp.NumberFormat{Places: fxp.ZeroPlaces}

	equipment := NewEquipment(nil, nil, false)
	equipment.BaseWeight = "8 lb"
	equipment.BaseValue = "1234.5678"
	data := CellData{ForPage: true}
	equipment.CellData(EquipmentWeightColumn, &data)
	c.Equal("8.0 lb", data.Primary)
	c.Equal("8 lb", data.Tooltip)
	data = CellData{ForPage: true}
	equipment.CellData(EquipmentCostColumn, &data)
	c.Equal("1,235", data.Primary)
	c.Equal("1,234.5678", data.Tooltip)
	data = CellData{}
	equipment.CellData(EquipmentCostColumn, &data)
	c.Equal("1,234.5678", data.Primary, "off the page, the default formats do not apply either")

	template := NewTemplate()
	template.Equipment = []*Equipment{equipment}
	header := EquipmentHeaderData(EquipmentDescriptionColumn, template, true, true)
	c.Equal("Equipment (8.0 lb; $1,235)", header.Title)
	c.Equal("8 lb; $1,234.5678", header.Detail)

	loot := NewLoot()
	loot.Equipment = []*Equipment{equipment}
	header = EquipmentHeaderData(EquipmentDescriptionColumn, loot, false, true)
	c.Equal("Equipment (8.0 lb; $1,235)", header.Title)
	c.Equal("8 lb; $1,234.5678", header.Detail)
}
