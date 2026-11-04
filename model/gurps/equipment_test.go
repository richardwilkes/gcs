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
	"encoding/json/v2"
	"os"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/eqcontainer"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/model/nameable"
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
	// Assign directly rather than via SetModifiers(), since attaching a modifier migrates its legacy replacements
	// into the equipment. This is the un-migrated state that exists between loading the data and attaching it.
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
	// Assign directly rather than via SetModifiers(), since attaching a modifier migrates its legacy replacements
	// into the equipment. This is the un-migrated state that exists between loading the data and attaching it.
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

// TestEquipmentPageRefCellLooksForTheName verifies that, without a highlight, the page reference cell of a piece of
// equipment looks for its name alone, as those of the other kinds do, since the level and title notes its description
// adds are not in the book.
func TestEquipmentPageRefCellLooksForTheName(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	sword := addCarriedEquipmentWithFeatures(e, "Broadsword", newTitleNoteFor("Own"))
	sword.PageRef = "B271"
	sword.Level = fxp.Two
	fine := NewEquipmentModifier(e, nil, false)
	fine.Name = "Fine"
	fine.ShowInTitle = true
	sword.SetModifiers([]*EquipmentModifier{fine})
	c.Equal("Broadsword 2 (Own; Fine)", sword.String(), "precondition: the description adds the level and title notes")
	var data CellData
	sword.CellData(EquipmentReferenceColumn, &data)
	c.Equal("B271", data.Primary)
	c.Equal("Broadsword", data.Secondary, "the name is looked for without the level or title notes")
	sword.PageRefHighlight = "Broadswords"
	data = CellData{}
	sword.CellData(EquipmentReferenceColumn, &data)
	c.Equal("Broadswords", data.Secondary, "a highlight is looked for as given")
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

// TestEquipmentApplyNameableKeysKeepsAnswersForDisabledModifiers verifies that applying the replacements for the keys
// a piece of equipment has in use keeps those held for its disabled modifiers, which are never offered, while dropping
// any that nothing uses any more.
func TestEquipmentApplyNameableKeysKeepsAnswersForDisabledModifiers(t *testing.T) {
	c := check.New(t)
	equipment := NewEquipment(nil, nil, false)
	equipment.Name = "@Foo@ Sword"
	finish := NewEquipmentModifier(nil, nil, false)
	finish.Name = "@Finish@ Coating"
	finish.SetDisabled(true)
	color := NewEquipmentModifier(nil, nil, false)
	color.Name = "@Color@ Paint"
	equipment.AddModifiers(finish, color)
	equipment.Replacements = map[string]string{"Foo": "Bar", "Finish": "Gilded", "Stale": "Unused"}

	m := make(map[string]string)
	equipment.FillWithNameableKeys(m, nil)
	c.Equal(map[string]string{"Foo": "Bar", "Color": nameable.Unset}, m,
		"only the keys in use are offered, the switched-off modifier's among the ones left out")
	m["Color"] = "Red"
	equipment.ApplyNameableKeys(m)
	c.Equal(map[string]string{"Foo": "Bar", "Color": "Red", "Finish": "Gilded"}, equipment.Replacements,
		"the answers are applied, the switched-off modifier's answer is kept and the unused one is dropped")
	finish.SetDisabled(false)
	c.Equal("Gilded Coating", finish.NameWithReplacements(), "switching the modifier back on must find its answer")

	// With nothing in use and nothing held for a disabled modifier, nothing is held at all.
	finish.SetDisabled(true)
	equipment.Name = "Sword"
	equipment.ApplyNameableKeys(map[string]string{})
	c.Equal(map[string]string{"Finish": "Gilded"}, equipment.Replacements, "the switched-off modifier's answer survives")
	equipment.SetModifiers(nil)
	equipment.ApplyNameableKeys(map[string]string{})
	c.Nil(equipment.Replacements, "nothing left to hold must be held as nothing")
}

func TestDisabledContainerDisablesItsContents(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	box := NewEquipment(e, nil, true)
	box.BaseWeight = "1 lb"
	box.BaseValue = "5"
	item := NewEquipment(e, box, false)
	item.BaseWeight = "3 lb"
	item.BaseValue = "10"
	other := NewEquipment(e, box, false)
	other.BaseWeight = "2 lb"
	other.BaseValue = "20"
	box.Children = []*Equipment{item, other}
	e.CarriedEquipment = append(e.CarriedEquipment, box)
	e.Recalculate()
	c.Equal(fxp.WeightFromInteger(6, fxp.Pound), e.WeightCarried(false))
	c.Equal(fxp.FromInteger(35), e.WealthCarried())

	item.SetDisabled(true)
	e.DiscardCaches() // The carried weight is cached until the caches are discarded, as a recalculation does.
	c.Equal(fxp.WeightFromInteger(3, fxp.Pound), e.WeightCarried(false), "a disabled item adds nothing to its container")
	c.Equal(fxp.WeightFromInteger(3, fxp.Pound), e.WeightCarried(true))
	c.Equal(fxp.FromInteger(25), e.WealthCarried())

	item.SetDisabled(false)
	box.SetDisabled(true)
	e.DiscardCaches()
	c.False(item.Enabled())
	c.False(item.ReallyEquipped())
	c.Equal(fxp.Weight(0), e.WeightCarried(false))
}

func TestDisabledItemsAreLeftOutOfTheListTotals(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	a := addCarriedEquipmentWithFeatures(e, "A")
	a.BaseWeight = "3 lb"
	a.BaseValue = "10"
	b := addCarriedEquipmentWithFeatures(e, "B")
	b.BaseWeight = "2 lb"
	b.BaseValue = "20"
	e.Recalculate()
	title, _ := equipmentTotalsTitle("Carried Equipment", e.CarriedEquipment, e.SheetSettings)
	c.Equal("Carried Equipment (5 lb; $30)", title)

	b.SetDisabled(true)
	e.Recalculate()
	title, _ = equipmentTotalsTitle("Carried Equipment", e.CarriedEquipment, e.SheetSettings)
	c.Equal("Carried Equipment (3 lb; $10)", title)
}

// TestEquipmentContainerTypeDefaultsToPhysical verifies that a container written without a container type, as every
// container was before groups existed, loads as a physical container keeping everything it had, and that only a group
// writes its type out.
func TestEquipmentContainerTypeDefaultsToPhysical(t *testing.T) {
	c := check.New(t)
	var legacy Equipment
	c.NoError(json.Unmarshal([]byte(`{"id":"`+string(NewEquipment(nil, nil, true).TID)+
		`","description":"Backpack","quantity":2,"base_value":"60","base_weight":"3 lb","tech_level":"3"}`),
		&legacy))
	c.True(legacy.IsPhysicalContainer(), "a container with no type must be a physical container")
	c.Equal(fxp.FromInteger(2), legacy.Quantity)
	c.Equal("60", legacy.BaseValue)
	c.Equal("3 lb", legacy.BaseWeight)
	c.Equal("3", legacy.TechLevel)

	data, err := json.Marshal(&legacy)
	c.NoError(err)
	c.False(strings.Contains(string(data), `"container_type"`), "a physical container must not write its type")

	group := NewEquipmentGroup(nil, nil)
	data, err = json.Marshal(group)
	c.NoError(err)
	c.True(strings.Contains(string(data), `"container_type":"group"`), "a group must write its type")
	var loaded Equipment
	c.NoError(json.Unmarshal(data, &loaded))
	c.True(loaded.IsGroup(), "a group must load as one")
}

// TestEquipmentKinds verifies that each kind of equipment is named after what it is.
func TestEquipmentKinds(t *testing.T) {
	c := check.New(t)
	c.Equal("Equipment", NewEquipment(nil, nil, false).Kind())
	c.Equal("Equipment Container", NewEquipment(nil, nil, true).Kind())
	group := NewEquipmentGroup(nil, nil)
	c.Equal("Equipment Group", group.Kind())
	c.Equal("Equipment Group", group.Name)
	choice := NewEquipmentChoiceContainer(nil, nil)
	c.Equal("Equipment Choice", choice.Kind())
	c.Equal("Equipment Choice", choice.Name)
}

// TestEquipmentGroupHasNothingOfItsOwn verifies that a group holds none of the data that would make it a piece of
// equipment in its own right, and so is worth and weighs exactly what its contents do.
func TestEquipmentGroupHasNothingOfItsOwn(t *testing.T) {
	c := check.New(t)
	group := NewEquipmentGroup(nil, nil)
	c.Equal("", group.LegalityClass, "a new group must not start out with a legality class")
	group.Quantity = fxp.FromInteger(3)
	group.BaseValue = "100"
	group.BaseWeight = "5 lb"
	group.TechLevel = "4"
	group.LegalityClass = "2"
	group.MaxUses = 3
	group.Uses = 2
	group.RatedST = fxp.FromInteger(10)
	group.Level = fxp.One
	group.WeightIgnoredForSkills = true
	group.SwitchedOn = true
	group.Preconfigured = true
	group.Prereq = NewPrereqList()
	group.Features = Features{NewContainedWeightReduction()}
	group.Modifiers = []*EquipmentModifier{NewEquipmentModifier(nil, nil, false)}
	group.Weapons = []*Weapon{NewWeapon(group, true)}
	for _, one := range []*Equipment{newEquipmentItem("Rope", "10", "2 lb"), newEquipmentItem("Torch", "3", "1 lb")} {
		one.SetParent(group)
		group.Children = append(group.Children, one)
	}
	group.ClearUnusedFieldsForType()

	c.Equal(fxp.One, group.Quantity, "a group's quantity must always be one")
	c.Equal("", group.BaseValue)
	c.Equal("", group.BaseWeight)
	c.Equal("", group.TechLevel)
	c.Equal("", group.LegalityClass)
	c.Equal(0, group.MaxUses)
	c.Equal(0, group.Uses)
	c.Equal(fxp.Int(0), group.RatedST)
	c.Equal(fxp.Int(0), group.Level)
	c.False(group.WeightIgnoredForSkills)
	c.False(group.SwitchedOn)
	c.False(group.Preconfigured)
	c.True(group.Prereq.IsZero())
	c.Equal(0, len(group.Features))
	c.Equal(0, len(group.Modifiers))
	c.Equal(0, len(group.Weapons))
	c.Equal(2, len(group.Children), "a group must keep its contents")

	c.Equal(fxp.FromInteger(13), group.ExtendedValue(), "a group must be worth what its contents are")
	c.Equal(fxp.Weight(fxp.FromInteger(3)), group.ExtendedWeight(false, fxp.Pound),
		"a group must weigh what its contents do")
	c.False(group.RequiresTL(), "a group has no tech level of its own")
	c.False(group.CanPreconfigureContainer(), "a group has no modifiers of its own to preconfigure")
	c.True(NewEquipment(nil, nil, true).CanPreconfigureContainer(), "a physical container may still be preconfigured")
}

// TestEquipmentGroupConversion verifies that a physical container and a group can be converted into each other, that
// converting to a group reports and removes what a group can't hold, and that a choice container can't skip straight
// to either a physical container or a piece of equipment that isn't a container.
func TestEquipmentGroupConversion(t *testing.T) {
	c := check.New(t)
	backpack := NewEquipment(nil, nil, true)
	backpack.Name = "Backpack"
	c.True(backpack.CanConvertToGroup())
	c.False(backpack.CanConvertToPhysicalContainer(), "a physical container already is one")
	c.Equal(0, len(backpack.GroupConversionLosses()), "a new container has nothing worth a warning to lose")
	backpack.LegalityClass = "2"
	c.Equal([]string{"legality class"}, backpack.GroupConversionLosses(), "a changed legality class is worth one")
	backpack.BaseValue = "60"
	backpack.Modifiers = []*EquipmentModifier{NewEquipmentModifier(nil, nil, false)}
	c.Equal([]string{"value", "legality class", "modifiers"}, backpack.GroupConversionLosses())

	backpack.ConvertToGroup()
	c.True(backpack.IsGroup())
	c.Equal("", backpack.BaseValue)
	c.Equal(0, len(backpack.Modifiers))
	c.False(backpack.CanConvertToGroup(), "a group already is one")

	c.True(backpack.CanConvertToPhysicalContainer())
	backpack.ConvertToPhysicalContainer()
	c.True(backpack.IsPhysicalContainer())
	c.Equal("4", backpack.LegalityClass, "a container must get back the legality class new equipment starts with")

	c.False(NewEquipment(nil, nil, false).CanConvertToGroup(), "only a container may become a group")

	choice := NewEquipmentChoiceContainer(nil, nil)
	c.False(choice.CanConvertToPhysicalContainer(), "a choice container must become a group first")
	c.False(choice.CanConvertToGroup(), "a choice container becomes a group by losing its choices")
	c.False(choice.CanConvertToFromContainer(), "a choice container must not become a plain piece of equipment")
	c.True(NewEquipmentGroup(nil, nil).CanConvertToFromContainer(), "an empty group may become a plain item")
}

// TestEquipmentSyncToGroupClearsWhatAGroupCantHold verifies that a physical container synced to a library entry that
// has since become a group drops what a group can't hold, rather than keeping it hidden until the next save.
func TestEquipmentSyncToGroupClearsWhatAGroupCantHold(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	libFile := LibraryFile{Library: "Test Library", Path: "Test" + EquipmentExt}
	source := NewEquipmentGroup(nil, nil)
	source.Name = "Kit"
	stubLibrarySources(t, e.SourceMatcher(), libFile, source)
	local := NewEquipment(e, nil, true)
	local.Name = "Kit"
	local.Quantity = fxp.FromInteger(2)
	local.RatedST = fxp.FromInteger(10)
	local.Modifiers = []*EquipmentModifier{NewEquipmentModifier(e, nil, false)}
	local.Source = Source{LibraryFile: libFile, TID: source.TID}
	state, _ := e.SourceMatcher().Match(local)
	c.Equal(srcstate.Mismatched, state, "precondition: the local copy differs from its source")

	local.SyncWithSource()
	c.True(local.IsGroup(), "the container must have become a group, as its source is")
	c.Equal(fxp.One, local.Quantity, "a group's quantity must be one")
	c.Equal(fxp.Int(0), local.RatedST)
	c.Equal(0, len(local.Modifiers))
}

// TestEquipmentGroupRoundTripsThroughAnItem verifies that a group converted to a plain item and back comes back as a
// physical container, since the item it became is a piece of equipment in its own right, and that the item keeps no
// trace of having been a group but gets the legality class new equipment starts out with, while a physical container
// converted to an item keeps whatever legality class it had.
func TestEquipmentGroupRoundTripsThroughAnItem(t *testing.T) {
	c := check.New(t)
	group := NewEquipmentGroup(nil, nil)
	group.ConvertToNonContainer()
	c.False(group.Container(), "the group must have become an item")
	c.Equal(eqcontainer.Container, group.ContainerType, "the item must not keep the group's container type")
	c.Equal("4", group.LegalityClass, "the item must get the legality class new equipment starts out with")
	group.BaseValue = "10"
	group.ConvertToContainer()
	c.True(group.IsPhysicalContainer(), "an item must become a physical container")
	c.Equal("10", group.BaseValue, "the value the item was given must be kept")
	c.Equal(fxp.FromInteger(10), group.ExtendedValue())

	state := group.ContainerConversionState()
	group.ConvertToGroup()
	group.RestoreContainerConversionState(state)
	c.True(group.IsPhysicalContainer(), "restoring the state must put the container type back")

	backpack := NewEquipment(nil, nil, true)
	backpack.LegalityClass = ""
	backpack.ConvertToNonContainer()
	c.Equal("", backpack.LegalityClass, "a legality class the user cleared must be left alone")
}

// TestEditorDataCantChangeTheKindOfContainer verifies that data an editor took from a container before the container
// was converted to another kind can't put the old kind back when applied, since the kind isn't among the data an editor
// edits.
func TestEditorDataCantChangeTheKindOfContainer(t *testing.T) {
	c := check.New(t)
	backpack := NewEquipment(nil, nil, true)
	var data EquipmentEditData
	data.CopyFrom(backpack)
	backpack.ConvertToGroup()
	data.ApplyTo(backpack)
	c.True(backpack.IsGroup(), "applying the stale data must leave the group a group")

	clone := backpack.Clone(LibraryFile{}, nil, nil, Copy)
	c.True(clone.IsGroup(), "a clone must keep the kind of container")
}
