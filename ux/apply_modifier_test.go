// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/side"
)

// newApplyModifierTestSheet returns an unmodified sheet whose entity holds the traits Alpha and Beta, the carried
// equipment Rope and Backpack (holding Pouch, holding Knife) and the other equipment Coin.
func newApplyModifierTestSheet(t *testing.T) *Sheet {
	t.Helper()
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	entity.Traits = nil
	for _, name := range []string{"Alpha", "Beta"} {
		trait := gurps.NewTrait(entity, nil, false)
		trait.Name = name
		entity.Traits = append(entity.Traits, trait)
	}
	rope := gurps.NewEquipment(entity, nil, false)
	rope.Name = "Rope"
	backpack := gurps.NewEquipment(entity, nil, true)
	backpack.Name = "Backpack"
	pouch := gurps.NewEquipment(entity, backpack, true)
	pouch.Name = "Pouch"
	knife := gurps.NewEquipment(entity, pouch, false)
	knife.Name = "Knife"
	pouch.Children = []*gurps.Equipment{knife}
	backpack.Children = []*gurps.Equipment{pouch}
	entity.CarriedEquipment = []*gurps.Equipment{rope, backpack}
	coin := gurps.NewEquipment(entity, nil, false)
	coin.Name = "Coin"
	entity.OtherEquipment = []*gurps.Equipment{coin}
	entity.Recalculate()
	sheet.Rebuild(true)
	// NewSheet took the hash the sheet is compared against before any of the above was added.
	sheet.markUnmodified()
	return sheet
}

// equipmentNamed returns the piece of equipment with the given name from the lists, looking inside containers, or nil
// if there is none. The lists are searched afresh each time, since an undo replaces their contents.
func equipmentNamed(name string, lists ...[]*gurps.Equipment) *gurps.Equipment {
	var found *gurps.Equipment
	for _, list := range lists {
		gurps.Traverse(func(e *gurps.Equipment) bool {
			if e.Name == name {
				found = e
				return true
			}
			return false
		}, false, false, list...)
		if found != nil {
			return found
		}
	}
	return nil
}

// appliedModifierNames returns the names of the modifiers, for comparing what a target ended up with.
func appliedModifierNames[M gurps.GeneralModifier](modifiers []M) []string {
	names := make([]string, 0, len(modifiers))
	for _, one := range modifiers {
		names = append(names, one.NameWithReplacements())
	}
	return names
}

// TestAttachModifierClonesGivesEachTargetItsOwnCopies verifies that each target gets an enabled clone of each
// modifier, never the original or a clone shared with another target, and that without an entity no prompt is shown
// and a container's contents are left as they came.
func TestAttachModifierClonesGivesEachTargetItsOwnCopies(t *testing.T) {
	c := check.New(t)
	forbidModifierPrompts(t)
	first := gurps.NewTrait(nil, nil, false)
	first.Name = "First"
	second := gurps.NewTrait(nil, nil, false)
	second.Name = "Second"
	table := newLibraryStyleTraitsTable(first, second)
	tables := []*unison.Table[*Node[*gurps.Trait]]{table}
	provider, ok := table.ClientData()[TableProviderClientKey].(gurps.DataOwnerProvider)
	c.True(ok, "the table must carry its provider")
	owner := provider.DataOwner()
	mod := gurps.NewTraitModifier(nil, nil, false)
	mod.Name = "Ranged"
	mod.Disabled = true
	container := gurps.NewTraitModifier(nil, nil, true)
	container.Name = "Group"
	inner := gurps.NewTraitModifier(nil, container, false)
	inner.Name = "Inner"
	inner.Disabled = true
	container.Children = []*gurps.TraitModifier{inner}

	c.True(attachModifierClones(tables, owner, []*gurps.Trait{first, second}, []*gurps.TraitModifier{mod, container},
		gurps.LibraryFile{}))

	c.Equal([]string{"Ranged", "Group"}, appliedModifierNames(first.Modifiers))
	c.Equal([]string{"Ranged", "Group"}, appliedModifierNames(second.Modifiers))
	c.NotEqual(mod, first.Modifiers[0], "the target must get a clone, not the original")
	c.NotEqual(first.Modifiers[0], second.Modifiers[0], "each target must get a clone of its own")
	c.Equal(first, first.Modifiers[0].Target(), "the clone must know the trait it was attached to")
	c.Equal(second, second.Modifiers[0].Target(), "the clone must know the trait it was attached to")
	c.False(first.Modifiers[0].Disabled, "the clone is switched on")
	c.True(first.Modifiers[1].Children[0].Disabled,
		"the modifiers within a container are left as the library has them, since nothing asked about them")
	c.Nil(mod.Target(), "the original must be left untouched")
	c.True(mod.Disabled, "the original must be left untouched")
	c.True(inner.Disabled, "the original must be left untouched")
}

// TestAttachModifierClonesAsksAboutContainersOnASheet verifies that a container among the modifiers applied to a
// sheet's rows brings up the modifier prompt once per target, with that target's clones and the ones pointed at
// directly already enabled, and that the answers are kept. Canceling the prompt for the first target stops there,
// skips the nameables prompt, and leaves every target as it was with nothing shown or reported.
func TestAttachModifierClonesAsksAboutContainersOnASheet(t *testing.T) {
	newModifiers := func() []*gurps.TraitModifier {
		ranged := gurps.NewTraitModifier(nil, nil, false)
		ranged.Name = "Ranged"
		ranged.Disabled = true
		group := gurps.NewTraitModifier(nil, nil, true)
		group.Name = "Group"
		for _, name := range []string{"First", "Second", "Third"} {
			child := gurps.NewTraitModifier(nil, group, false)
			child.Name = name
			child.Disabled = name != "Second"
			group.Children = append(group.Children, child)
		}
		return []*gurps.TraitModifier{ranged, group}
	}
	type prompt struct {
		title   string
		names   []string
		enabled []bool
	}
	// recordPrompts substitutes a prompt that records what it was shown and then lets respond act on it.
	recordPrompts := func(t *testing.T, respond func(title string, modifiers []*gurps.TraitModifier) (changed, canceled bool)) *[]prompt {
		t.Helper()
		var prompts []prompt
		swapForTest(t, &promptForTraitModifiers, func(title string, modifiers []*gurps.TraitModifier) (changed, canceled bool) {
			p := prompt{title: title}
			gurps.Traverse(func(m *gurps.TraitModifier) bool {
				p.names = append(p.names, m.Name)
				p.enabled = append(p.enabled, m.Enabled())
				return false
			}, false, false, modifiers...)
			prompts = append(prompts, p)
			return respond(title, modifiers)
		})
		return &prompts
	}
	wantPrompt := func(title string) prompt {
		return prompt{
			title:   title,
			names:   []string{"Ranged", "Group", "First", "Second", "Third"},
			enabled: []bool{true, true, false, true, false},
		}
	}

	t.Run("answered", func(t *testing.T) {
		c := check.New(t)
		sheet := newApplyModifierTestSheet(t)
		entity := sheet.Entity()
		alpha, beta := entity.Traits[0], entity.Traits[1]
		prompts := recordPrompts(t, func(title string, modifiers []*gurps.TraitModifier) (changed, canceled bool) {
			if title == "Alpha" {
				// Alpha takes the first alternative instead of the second.
				modifiers[1].Children[0].Disabled = false
				modifiers[1].Children[1].Disabled = true
				return true, false
			}
			return false, false
		})

		c.True(attachModifierClones([]*unison.Table[*Node[*gurps.Trait]]{sheet.Traits.Table}, entity,
			[]*gurps.Trait{alpha, beta}, newModifiers(), gurps.LibraryFile{}))
		c.Equal([]prompt{wantPrompt("Alpha"), wantPrompt("Beta")}, *prompts,
			"each target must be asked about its own copies, the ones pointed at directly already switched on")
		c.Equal([]string{"Ranged", "Group"}, appliedModifierNames(alpha.Modifiers))
		c.False(alpha.Modifiers[1].Children[0].Disabled, "Alpha's answer must be kept")
		c.True(alpha.Modifiers[1].Children[1].Disabled, "Alpha's answer must be kept")
		c.True(alpha.Modifiers[1].Children[2].Disabled, "Alpha's answer must be kept")
		c.Equal([]string{"Ranged", "Group"}, appliedModifierNames(beta.Modifiers))
		c.True(beta.Modifiers[1].Children[0].Disabled, "Beta keeps the library's choice")
		c.False(beta.Modifiers[1].Children[1].Disabled, "Beta keeps the library's choice")
		c.True(beta.Modifiers[1].Children[2].Disabled, "Beta keeps the library's choice")
		c.True(sheet.Modified(), "the sheet has been changed")
	})

	t.Run("canceled", func(t *testing.T) {
		c := check.New(t)
		sheet := newApplyModifierTestSheet(t)
		entity := sheet.Entity()
		alpha, beta := entity.Traits[0], entity.Traits[1]
		existing := gurps.NewTraitModifier(nil, nil, false)
		existing.Name = "Existing"
		alpha.AddModifiers(existing)
		sheet.markUnmodified()
		stale := sheet.Traits.Table
		counter := installSyncCounter(sheet)
		prompts := recordPrompts(t, func(title string, _ []*gurps.TraitModifier) (changed, canceled bool) {
			return false, title == "Alpha"
		})
		nameablesShown := 0
		swapForTest(t, &promptForNameables, func(_ []string, _ []map[string]string, _ [][]string) bool {
			nameablesShown++
			return true
		})
		modifiers := append(newModifiers(), newSwitchableTraitModifier("@Material@ Coating"))

		c.False(attachModifierClones([]*unison.Table[*Node[*gurps.Trait]]{stale}, entity,
			[]*gurps.Trait{alpha, beta}, modifiers, gurps.LibraryFile{}), "a canceled prompt must be reported")
		c.Equal(1, len(*prompts), "the prompts must stop at the one that was canceled")
		c.Equal(0, nameablesShown, "the nameables prompt must not follow a canceled modifier prompt")
		c.Equal([]string{"Existing"}, appliedModifierNames(alpha.Modifiers),
			"the copies must be taken back off the target, leaving it the modifiers it had")
		c.Equal(0, len(beta.Modifiers), "a target that was never asked about must be left alone")
		c.False(sheet.Modified(), "the sheet must be left as it was")
		c.Equal(0, counter.count, "nothing must have been shown")
		c.Equal(stale, sheet.Traits.Table, "nothing must have been rebuilt")
	})
}

// TestAttachModifierClonesOnASheetRebuildsWithoutPrompting verifies that attaching enabled clones to an entity's rows
// rebuilds the sheet as modified without prompting, which is what reports the change and replaces the traits list
// once a clone brings a switchable feature.
func TestAttachModifierClonesOnASheetRebuildsWithoutPrompting(t *testing.T) {
	c := check.New(t)
	forbidModifierPrompts(t)
	sheet := newApplyModifierTestSheet(t)
	entity := sheet.Entity()
	c.False(sheet.Modified(), "the sheet must start out unmodified for the check to mean anything")
	entity.ModifiedOn = jio.Time{}
	stale := sheet.Traits.Table
	c.Equal(-1, switchColumnIndex(stale.Columns, gurps.TraitSwitchColumn), "the traits list starts without switches")
	mod := newSwitchableTraitModifier("Ranged")

	c.True(attachModifierClones([]*unison.Table[*Node[*gurps.Trait]]{stale}, entity,
		[]*gurps.Trait{entity.Traits[0], entity.Traits[1]}, []*gurps.TraitModifier{mod}, gurps.LibraryFile{}))
	c.False(entity.Traits[0].Modifiers[0].Disabled, "the clone is switched on")
	c.False(entity.Traits[1].Modifiers[0].Disabled, "the clone is switched on")
	c.True(sheet.Modified(), "attaching modifiers to a sheet's traits must mark the sheet as modified")
	c.NotEqual(jio.Time{}, entity.ModifiedOn, "the rebuild must report the change by bumping the timestamp")
	live := sheet.Traits.Table
	c.NotEqual(stale, live, "gaining the switch column must replace the traits table")
	c.NotEqual(-1, switchColumnIndex(live.Columns, gurps.TraitSwitchColumn),
		"the switched-on clone's switchable feature must bring the switch column into view")
}

// TestAttachModifierClonesCanceledPromptPutsTheTargetBack verifies that canceling the nameables prompt takes the
// clones back off the target and puts back its replacements, which attaching a modifier saved by an older version adds
// to.
func TestAttachModifierClonesCanceledPromptPutsTheTargetBack(t *testing.T) {
	c := check.New(t)
	forbidModifierPrompts(t)
	sheet := newApplyModifierTestSheet(t)
	entity := sheet.Entity()
	trait := entity.Traits[0]
	trait.Name = "@Foo@ Alpha"
	trait.Replacements = map[string]string{"Foo": "Bar"}
	existing := gurps.NewTraitModifier(nil, nil, false)
	existing.Name = "Existing"
	trait.AddModifiers(existing)
	sheet.markUnmodified()
	named := gurps.NewTraitModifier(nil, nil, false)
	named.Name = "@Material@ Coating"
	named.Replacements = map[string]string{"Material": "Steel"} // As an older version saved it.
	shown := 0
	swapForTest(t, &promptForNameables, func(_ []string, _ []map[string]string, _ [][]string) bool {
		shown++
		c.Equal(map[string]string{"Foo": "Bar", "Material": "Steel"}, trait.Replacements,
			"the modifier's replacements must have been moved onto the trait by the time the prompt is up")
		return false
	})

	c.False(attachModifierClones([]*unison.Table[*Node[*gurps.Trait]]{sheet.Traits.Table}, entity,
		[]*gurps.Trait{trait}, []*gurps.TraitModifier{named}, gurps.LibraryFile{}),
		"a canceled prompt must be reported")
	c.Equal(1, shown, "the nameables prompt must have been shown")
	c.Equal([]string{"Existing"}, appliedModifierNames(trait.Modifiers),
		"the clone must be taken back off, leaving the trait the modifier it had")
	c.Equal(map[string]string{"Foo": "Bar"}, trait.Replacements, "the trait's replacements must be put back")
	c.False(sheet.Modified(), "the sheet must be left as it was")
}

// TestModifierDestinations verifies which open dockables the Apply Modifier command offers: those showing a non-empty
// list of the kind's rows, a sheet's two equipment lists counting as one destination. A filtered list, an empty one,
// or one the sheet's layout leaves off the page is not offered.
func TestModifierDestinations(t *testing.T) {
	c := check.New(t)
	sheet := newApplyModifierTestSheet(t)
	templateData := gurps.NewTemplate()
	templateTrait := gurps.NewTrait(templateData, nil, false)
	templateTrait.Name = "Template Trait"
	templateData.Traits = []*gurps.Trait{templateTrait}
	templateItem := gurps.NewEquipment(templateData, nil, false)
	templateItem.Name = "Template Item"
	templateData.Equipment = []*gurps.Equipment{templateItem}
	template := newTestTemplateDockable("Template", templateData)
	emptyTemplate := newTestTemplateDockable("Empty", gurps.NewTemplate())
	loot := newTestLootSheet(t)
	lootItem := gurps.NewEquipment(loot.loot, nil, false)
	lootItem.Name = "Gem"
	loot.loot.Equipment = []*gurps.Equipment{lootItem}
	loot.Rebuild(true)
	libraryTrait := gurps.NewTrait(nil, nil, false)
	libraryTrait.Name = "Library Trait"
	traits := NewTraitTableDockable("traits"+gurps.TraitsExt, []*gurps.Trait{libraryTrait})
	libraryItem := gurps.NewEquipment(nil, nil, false)
	libraryItem.Name = "Library Item"
	equipment := NewEquipmentTableDockable("equipment"+gurps.EquipmentExt, []*gurps.Equipment{libraryItem})
	all := []unison.Dockable{sheet, template, emptyTemplate, loot, traits, equipment}

	titles := func(destinations []FileBackedDockable) []string {
		result := make([]string, 0, len(destinations))
		for _, d := range destinations {
			result = append(result, d.Title())
		}
		return result
	}
	traitKind := traitModifierTargetKind()
	equipmentKind := equipmentModifierTargetKind()
	c.Equal([]string{sheet.Title(), template.Title(), traits.Title()},
		titles(modifierDestinations(traitKind, all)),
		"trait modifiers go to whatever shows traits: the sheet, the template with a trait and the trait library")
	c.Equal([]string{sheet.Title(), template.Title(), loot.Title(), equipment.Title()},
		titles(modifierDestinations(equipmentKind, all)),
		"equipment modifiers go to whatever shows equipment: the sheet, the template with an item, the loot sheet and the equipment library")
	c.Equal(1, len(equipmentKind.lists(template)), "a template offers its equipment list")
	c.Equal(template.Equipment.Table, equipmentKind.lists(template)[0].table)

	lists := equipmentKind.lists(sheet)
	c.Equal(2, len(lists), "a sheet offers both of its equipment lists")
	c.Equal("Carried Equipment", lists[0].name)
	c.Equal("Other Equipment", lists[1].name)
	c.Equal(1, len(traitKind.lists(sheet)), "a sheet offers its traits list")
	c.Equal("", traitKind.lists(sheet)[0].name, "a lone list needs no name")

	// A filtered list is left out, as a drop onto it would be refused.
	traits.table.ApplyHierarchicalFilter(func(row *Node[*gurps.Trait]) bool { return row.Data().Name != "Library Trait" })
	c.True(traits.table.IsFiltered(), "the library must be filtered for the check to mean anything")
	c.Equal([]string{sheet.Title(), template.Title()}, titles(modifierDestinations(traitKind, all)),
		"a filtered library must not be offered")

	// A list the sheet's layout leaves off the page is left out.
	entity := sheet.Entity()
	entity.SheetSettings.Layout = entity.SheetSettings.Layout.Filtered(func(key string) bool {
		return key != gurps.BlockTraitsKey
	})
	sheet.Rebuild(true)
	c.False(entity.SheetSettings.Layout.Contains(gurps.BlockTraitsKey), "the traits block must be off the page")
	c.Equal([]string{template.Title()}, titles(modifierDestinations(traitKind, all)),
		"a sheet whose layout hides its traits must not be offered them")
	c.Equal(2, len(equipmentKind.lists(sheet)), "the sheet's equipment lists are unaffected")
}

// TestApplySelectedModifiersAppliesASelectedContainerOnceAndRecordsItsSource verifies that with a container and one of
// its own modifiers both selected, the container alone is applied, and that the copies record the library file they
// came from, so that "Sync With Sources" can find them.
func TestApplySelectedModifiersAppliesASelectedContainerOnceAndRecordsItsSource(t *testing.T) {
	c := check.New(t)
	forbidModifierPrompts(t)
	sheet := newApplyModifierTestSheet(t)
	entity := sheet.Entity()
	alpha := entity.Traits[0]
	_, user := useTestLibraries(t, c)
	group := gurps.NewTraitModifier(nil, nil, true)
	group.Name = "Group"
	inner := gurps.NewTraitModifier(nil, group, false)
	inner.Name = "Inner"
	group.Children = []*gurps.TraitModifier{inner}
	library := NewTraitModifierTableDockable(filepath.Join(user.Path(), "Sub", "mods"+gurps.TraitModifiersExt),
		[]*gurps.TraitModifier{group})
	library.table.SetSelectionMap(map[tid.TID]bool{group.ID(): true, inner.ID(): true})
	c.Equal(2, len(library.table.SelectedRows(false)),
		"both the container and the modifier within it must be selected for the check to mean anything")
	swapForTest(t, &promptForModifierDestination, func(_ []FileBackedDockable) (FileBackedDockable, bool) {
		return sheet, true
	})
	swapForTest(t, &showModifierTargetsPrompt, func(_ string, list unison.Paneler, _ ...*unison.Label) bool {
		targets, ok := list.(*unison.List[modifierTargetChoice[*gurps.Trait]])
		if !ok {
			t.Errorf("the target prompt must hold a list of trait choices, not a %T", list)
			return false
		}
		targets.Select(false, 0)
		return true
	})
	// The container prompt is answered by leaving its contents as they are.
	containerPrompts := stubTraitModifierPrompt(t, func(_ []*gurps.TraitModifier) bool { return false })

	applySelectedModifiers(library.table, traitModifierTargetKind())

	c.Equal(1, *containerPrompts, "the target must have been asked about the container's contents")
	c.Equal([]string{"Group"}, appliedModifierNames(alpha.Modifiers),
		"the container must be applied once, and the modifier within it not applied on its own as well")
	if len(alpha.Modifiers) != 1 {
		return
	}
	c.Equal([]string{"Inner"}, appliedModifierNames(alpha.Modifiers[0].Children),
		"the modifier within the container must arrive as part of it")
	want := gurps.LibraryFile{Library: user.Key(), Path: "Sub/mods" + gurps.TraitModifiersExt}
	c.Equal(gurps.Source{LibraryFile: want, TID: group.TID}, alpha.Modifiers[0].Source,
		"the copy must record the library file it was taken from")
	c.Equal(gurps.Source{LibraryFile: want, TID: inner.TID}, alpha.Modifiers[0].Children[0].Source,
		"the copy of the modifier within the container must record the library file too")
}

// TestCanReceiveModifiersNeedsAProviderAndAnUndoManager verifies that a list without its provider or an undo manager
// can't receive modifiers.
func TestCanReceiveModifiersNeedsAProviderAndAnUndoManager(t *testing.T) {
	c := check.New(t)
	sheet := newApplyModifierTestSheet(t)
	table := sheet.Traits.Table
	c.True(canReceiveModifiers(table), "a sheet's list with traits in it can receive modifiers")

	provider := table.ClientData()[TableProviderClientKey]
	delete(table.ClientData(), TableProviderClientKey)
	c.False(canReceiveModifiers(table), "a list without its provider cannot receive modifiers")
	table.ClientData()[TableProviderClientKey] = provider
	c.True(canReceiveModifiers(table), "the list must be back to receiving modifiers once its provider is back")

	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Loose"
	loose := newLibraryStyleTraitsTable(trait)
	c.Nil(unison.UndoManagerFor(loose), "the loose table must have no undo manager for the check to mean anything")
	c.False(canReceiveModifiers(loose), "a list with no undo manager above it cannot receive modifiers")
}

// TestModifierTargetChoicesLabelRowsWithWhereTheySit verifies the rows the target prompt offers: every row of every
// list in table order, containers included, labeled with its name, its containers from the nearest outward and, with
// more than one list, the list it is in.
func TestModifierTargetChoicesLabelRowsWithWhereTheySit(t *testing.T) {
	c := check.New(t)
	sheet := newApplyModifierTestSheet(t)
	lists := equipmentModifierTargetKind().lists(sheet)
	c.Equal(2, len(lists))

	choices := modifierTargetChoices(lists)
	labels := make([]string, 0, len(choices))
	depths := make([]int, 0, len(choices))
	names := make([]string, 0, len(choices))
	tables := make([]*unison.Table[*Node[*gurps.Equipment]], 0, len(choices))
	for _, choice := range choices {
		labels = append(labels, choice.String())
		depths = append(depths, choice.depth)
		names = append(names, choice.target.Name)
		tables = append(tables, choice.table)
	}
	c.Equal([]string{
		"Rope (in Carried Equipment)",
		"Backpack (in Carried Equipment)",
		"Pouch (in Backpack, Carried Equipment)",
		"Knife (in Pouch, Backpack, Carried Equipment)",
		"Coin (in Other Equipment)",
	}, labels)
	c.Equal([]int{0, 0, 1, 2, 0}, depths)
	c.Equal([]string{"Rope", "Backpack", "Pouch", "Knife", "Coin"}, names)
	carried, other := sheet.CarriedEquipment.Table, sheet.OtherEquipment.Table
	c.Equal([]*unison.Table[*Node[*gurps.Equipment]]{carried, carried, carried, carried, other}, tables,
		"each choice must know the table its target sits in")

	// With one list, only the containers are named.
	choices = modifierTargetChoices(lists[:1])
	labels = labels[:0]
	for _, choice := range choices {
		labels = append(labels, choice.String())
	}
	c.Equal([]string{"Rope", "Backpack", "Pouch (in Backpack)", "Knife (in Pouch, Backpack)"}, labels)
}

// TestApplyModifiersToBothEquipmentListsIsOneEdit verifies that applying to targets in both of a sheet's equipment
// lists attaches the modifier to each of them as a single undoable edit, so that one Undo takes the modifier off both
// and one Redo puts it back on both.
func TestApplyModifiersToBothEquipmentListsIsOneEdit(t *testing.T) {
	c := check.New(t)
	sheet := newApplyModifierTestSheet(t)
	entity := sheet.Entity()
	entity.ModifiedOn = jio.Time{}
	mgr := unison.UndoManagerFor(sheet.CarriedEquipment.Table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	c.False(mgr.CanUndo())
	mod := gurps.NewEquipmentModifier(nil, nil, false)
	mod.Name = "Fine"
	mod.Disabled = true
	forbidModifierPrompts(t)
	rope := equipmentNamed("Rope", entity.CarriedEquipment)
	coin := equipmentNamed("Coin", entity.OtherEquipment)

	c.True(applyModifiersTo([]*unison.Table[*Node[*gurps.Equipment]]{sheet.CarriedEquipment.Table, sheet.OtherEquipment.Table},
		[]*gurps.Equipment{rope, coin}, []*gurps.EquipmentModifier{mod}, gurps.LibraryFile{}))
	c.Equal([]string{"Fine"}, appliedModifierNames(equipmentNamed("Rope", entity.CarriedEquipment).Modifiers))
	c.Equal([]string{"Fine"}, appliedModifierNames(equipmentNamed("Coin", entity.OtherEquipment).Modifiers))
	c.False(equipmentNamed("Rope", entity.CarriedEquipment).Modifiers[0].Disabled, "the applied copy is switched on")
	c.False(equipmentNamed("Coin", entity.OtherEquipment).Modifiers[0].Disabled, "the applied copy is switched on")
	c.Equal(0, len(equipmentNamed("Knife", entity.CarriedEquipment).Modifiers), "only the chosen targets are touched")
	c.True(sheet.Modified(), "the sheet has been changed")
	c.NotEqual(jio.Time{}, entity.ModifiedOn, "the change must be reported")
	c.True(mgr.CanUndo(), "the change must be undoable")
	c.Equal("Undo "+applyModifierAction.Title, mgr.UndoTitle(), "the edit is named for the command")

	mgr.Undo()
	c.False(mgr.CanUndo(), "one command must make exactly one edit")
	c.Equal(0, len(equipmentNamed("Rope", entity.CarriedEquipment).Modifiers), "undo must take the modifier back off")
	c.Equal(0, len(equipmentNamed("Coin", entity.OtherEquipment).Modifiers), "undo must take the modifier back off")
	c.False(sheet.Modified(), "undo must leave the sheet as it was")

	mgr.Redo()
	c.Equal([]string{"Fine"}, appliedModifierNames(equipmentNamed("Rope", entity.CarriedEquipment).Modifiers))
	c.Equal([]string{"Fine"}, appliedModifierNames(equipmentNamed("Coin", entity.OtherEquipment).Modifiers))
}

// TestApplyModifiersToCanceledPromptChangesNothing verifies that canceling the nameables prompt leaves the sheet
// untouched: the targets keep the modifiers they had, the sheet is neither rebuilt nor re-synced, its timestamp is not
// bumped, and nothing is left to undo.
func TestApplyModifiersToCanceledPromptChangesNothing(t *testing.T) {
	c := check.New(t)
	sheet := newApplyModifierTestSheet(t)
	entity := sheet.Entity()
	sturdy := gurps.NewEquipmentModifier(nil, nil, false)
	sturdy.Name = "Sturdy"
	equipmentNamed("Rope", entity.CarriedEquipment).AddModifiers(sturdy)
	sheet.markUnmodified()
	stamp := jio.Time(time.Date(2020, time.January, 2, 3, 4, 5, 0, time.UTC))
	entity.ModifiedOn = stamp
	mgr := unison.UndoManagerFor(sheet.CarriedEquipment.Table)
	carried := sheet.CarriedEquipment.Table
	other := sheet.OtherEquipment.Table
	counter := installSyncCounter(sheet)
	mod := newSwitchableEquipmentModifier("Fine @Quality@")
	forbidModifierPrompts(t)
	shown := 0
	swapForTest(t, &promptForNameables, func(titles []string, _ []map[string]string, _ [][]string) bool {
		shown++
		c.Equal(2, len(titles), "one prompt covers the copy each target got")
		c.Equal(0, counter.count, "nothing may be shown before the prompt is answered")
		return false
	})
	rope := equipmentNamed("Rope", entity.CarriedEquipment)
	coin := equipmentNamed("Coin", entity.OtherEquipment)

	c.False(applyModifiersTo([]*unison.Table[*Node[*gurps.Equipment]]{carried, other},
		[]*gurps.Equipment{rope, coin}, []*gurps.EquipmentModifier{mod}, gurps.LibraryFile{}))
	c.Equal(1, shown, "the nameables prompt must have been shown once")
	c.Equal([]string{"Sturdy"}, appliedModifierNames(equipmentNamed("Rope", entity.CarriedEquipment).Modifiers),
		"the first target must be left with the modifier it had")
	c.Equal(0, len(equipmentNamed("Coin", entity.OtherEquipment).Modifiers), "the second target must be left alone")
	c.False(mgr.CanUndo(), "a canceled command must leave nothing to undo")
	c.False(sheet.Modified(), "a canceled command must leave the sheet unmodified")
	c.Equal(stamp, entity.ModifiedOn, "a canceled command must not bump the modification timestamp")
	c.Equal(0, counter.count, "a canceled command must not re-sync the sheet")
	c.Equal(carried, sheet.CarriedEquipment.Table, "a canceled command must not rebuild the sheet")
	c.Equal(other, sheet.OtherEquipment.Table, "a canceled command must not rebuild the sheet")
}

// TestApplyModifiersToKeepsEveryNameableAnswer verifies that applying several modifiers with nameable keys to one trait
// keeps every answer along with the trait's existing replacements.
func TestApplyModifiersToKeepsEveryNameableAnswer(t *testing.T) {
	c := check.New(t)
	forbidModifierPrompts(t)
	sheet := newApplyModifierTestSheet(t)
	trait := sheet.Entity().Traits[0]
	trait.Name = "@Foo@ Alpha"
	trait.Replacements = map[string]string{"Foo": "Bar"}
	material := gurps.NewTraitModifier(nil, nil, false)
	material.Name = "@Material@ Coating"
	color := gurps.NewTraitModifier(nil, nil, false)
	color.Name = "@Color@ Paint"
	answers := map[string]string{"Material": "Steel", "Color": "Red"}
	shown := 0
	swapForTest(t, &promptForNameables, func(titles []string, nameables []map[string]string, _ [][]string) bool {
		shown++
		c.Equal(2, len(titles), "one prompt covers both modifiers")
		for _, one := range nameables {
			for k := range one {
				one[k] = answers[k]
			}
		}
		return true
	})

	c.True(applyModifiersTo([]*unison.Table[*Node[*gurps.Trait]]{sheet.Traits.Table}, []*gurps.Trait{trait},
		[]*gurps.TraitModifier{material, color}, gurps.LibraryFile{}))
	c.Equal(1, shown, "the nameables prompt must have been shown once")
	c.Equal(map[string]string{"Foo": "Bar", "Material": "Steel", "Color": "Red"}, trait.Replacements)
	c.Equal("Bar Alpha", trait.NameWithReplacements(), "the trait's own substitution must survive")
	c.Equal([]string{"Steel Coating", "Red Paint"}, appliedModifierNames(trait.Modifiers),
		"every modifier's answer must survive")
}

// TestApplyModifiersToDestinationsWithoutAnEntity verifies the command on a template, a loot sheet and a library list,
// none of which has an entity: the clones are attached enabled with no prompt of any kind, the change is reported to
// the dockable, and it is recorded as one undoable edit.
func TestApplyModifiersToDestinationsWithoutAnEntity(t *testing.T) {
	forbidPrompts := func(t *testing.T) {
		t.Helper()
		forbidModifierPrompts(t)
		swapForTest(t, &promptForNameables, func(titles []string, _ []map[string]string, _ [][]string) bool {
			t.Errorf("the nameables prompt must not be shown, but was shown for %v", titles)
			return false
		})
	}
	// checkUndo verifies that the change was recorded as a single undoable edit whose undo takes the modifier back off
	// and leaves the destination unmodified, and whose redo puts it back. modifierCount has to look the target up
	// afresh, since an undo puts copies of the rows back.
	checkUndo := func(c check.Checker, dest FileBackedDockable, mgr *unison.UndoManager, modifierCount func() int) {
		c.NotNil(mgr, "the destination must have an undo manager")
		c.True(mgr.CanUndo(), "the change must be undoable")
		c.Equal("Undo "+applyModifierAction.Title, mgr.UndoTitle(), "the edit is named for the command")
		mgr.Undo()
		c.False(mgr.CanUndo(), "one command must make exactly one edit")
		c.Equal(0, modifierCount(), "undo must take the modifier back off")
		c.False(dest.Modified(), "undo must leave the destination as it was")
		c.True(mgr.CanRedo(), "the change must be redoable")
		mgr.Redo()
		c.Equal(1, modifierCount(), "redo must put the modifier back")
		c.True(dest.Modified(), "redo must leave the destination changed again")
		c.False(mgr.CanRedo(), "one redo must exhaust the edit")
	}

	t.Run("template", func(t *testing.T) {
		c := check.New(t)
		forbidPrompts(t)
		data := gurps.NewTemplate()
		trait := gurps.NewTrait(data, nil, false)
		trait.Name = "Template Trait"
		data.Traits = []*gurps.Trait{trait}
		template := newTestTemplateDockable("Template", data)
		template.markUnmodified()
		counter := &syncCounter{}
		counter.Self = counter
		template.AsPanel().AddChild(counter)
		mod := gurps.NewTraitModifier(nil, nil, false)
		mod.Name = "@Material@ Coating"
		mod.Disabled = true

		c.True(applyModifiersTo([]*unison.Table[*Node[*gurps.Trait]]{template.Traits.Table},
			[]*gurps.Trait{trait}, []*gurps.TraitModifier{mod}, gurps.LibraryFile{}))
		c.Equal([]string{"@Material@ Coating"}, appliedModifierNames(trait.Modifiers))
		c.False(trait.Modifiers[0].Disabled, "the clone is switched on")
		c.True(template.Modified(), "the template has been changed")
		c.True(counter.count > 0, "the change must be reported to the template, which brings its page up to date")
		checkUndo(c, template, unison.UndoManagerFor(template.Traits.Table),
			func() int { return len(template.template.Traits[0].Modifiers) })
	})

	t.Run("loot sheet", func(t *testing.T) {
		c := check.New(t)
		forbidPrompts(t)
		loot := newTestLootSheet(t)
		item := gurps.NewEquipment(loot.loot, nil, false)
		item.Name = "Gem"
		loot.loot.Equipment = []*gurps.Equipment{item}
		loot.Rebuild(true)
		loot.markUnmodified()
		loot.loot.ModifiedOn = jio.Time{}
		mod := gurps.NewEquipmentModifier(nil, nil, false)
		mod.Name = "@Material@ Coating"
		mod.Disabled = true

		c.True(applyModifiersTo([]*unison.Table[*Node[*gurps.Equipment]]{loot.Equipment.Table},
			[]*gurps.Equipment{item}, []*gurps.EquipmentModifier{mod}, gurps.LibraryFile{}))
		c.Equal([]string{"@Material@ Coating"}, appliedModifierNames(item.Modifiers))
		c.False(item.Modifiers[0].Disabled, "the clone is switched on")
		c.True(loot.Modified(), "the loot sheet has been changed")
		c.NotEqual(jio.Time{}, loot.loot.ModifiedOn, "the change must be reported to the loot sheet")
		checkUndo(c, loot, unison.UndoManagerFor(loot.Equipment.Table),
			func() int { return len(loot.loot.Equipment[0].Modifiers) })
	})

	t.Run("library", func(t *testing.T) {
		c := check.New(t)
		forbidPrompts(t)
		registerKeyBindingsOnce.Do(registerActions)
		RegisterKnownFileTypes() // Docking asks for the library's icon, which comes from the file type registry.
		swapForTest(t, &Workspace.DocumentDock, NewDocumentDock())
		trait := gurps.NewTrait(nil, nil, false)
		trait.Name = "Library Trait"
		library := NewTraitTableDockable("traits"+gurps.TraitsExt, []*gurps.Trait{trait})
		library.markUnmodified()
		// The library is docked, so that there is a tab whose title marks it as modified or not.
		Workspace.DocumentDock.DockTo(library, nil, side.Left)
		modifiedTitle := "*" + library.Title()
		tabShowsModified := func() bool {
			dc := unison.Ancestor[*unison.DockContainer](library)
			return dc != nil && dc.AsPanel().HasInSelfOrDescendants(func(p *unison.Panel) bool {
				s, ok := p.Self.(fmt.Stringer)
				return ok && s.String() == modifiedTitle
			})
		}
		c.False(tabShowsModified(), "the library's tab starts out showing it as unmodified")
		mod := gurps.NewTraitModifier(nil, nil, false)
		mod.Name = "@Material@ Coating"
		mod.Disabled = true

		c.True(applyModifiersTo([]*unison.Table[*Node[*gurps.Trait]]{library.table}, []*gurps.Trait{trait},
			[]*gurps.TraitModifier{mod}, gurps.LibraryFile{}))
		c.Equal([]string{"@Material@ Coating"}, appliedModifierNames(trait.Modifiers))
		c.False(trait.Modifiers[0].Disabled, "the clone is switched on")
		c.True(library.Modified(), "the library has been changed")
		c.True(tabShowsModified(), "the change must be reported to the library, which marks its tab as modified")
		checkUndo(c, library, unison.UndoManagerFor(library.table),
			func() int { return len(library.table.RootRows()[0].Data().Modifiers) })
	})
}

// TestApplyModifiersToOpensTheTargetsContainers verifies that the containers holding a target are opened before it is
// selected, that undo closes the containers the command opened, and that redo opens them again before putting the
// selection back.
func TestApplyModifiersToOpensTheTargetsContainers(t *testing.T) {
	c := check.New(t)
	forbidModifierPrompts(t)
	sheet := newApplyModifierTestSheet(t)
	entity := sheet.Entity()
	backpack := equipmentNamed("Backpack", entity.CarriedEquipment)
	pouch := equipmentNamed("Pouch", entity.CarriedEquipment)
	knife := equipmentNamed("Knife", entity.CarriedEquipment)
	backpack.SetOpen(false)
	pouch.SetOpen(false)
	table := sheet.CarriedEquipment.Table
	table.SyncToModel()
	c.Equal(1, table.LastRowIndex(), "only Rope and the closed Backpack are showing")
	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	mod := gurps.NewEquipmentModifier(nil, nil, false)
	mod.Name = "Fine"
	tables := []*unison.Table[*Node[*gurps.Equipment]]{table}
	selectedName := func() string {
		live := sheet.CarriedEquipment.Table
		if row := live.FirstSelectedRowIndex(); row != -1 && len(live.SelectedRows(false)) == 1 {
			return live.RowFromIndex(row).Data().Name
		}
		return ""
	}

	c.True(applyModifiersTo(tables, []*gurps.Equipment{knife}, []*gurps.EquipmentModifier{mod}, gurps.LibraryFile{}))
	c.True(backpack.IsOpen(), "the containers holding the target must be opened")
	c.True(pouch.IsOpen(), "the containers holding the target must be opened")
	c.Equal(3, sheet.CarriedEquipment.Table.LastRowIndex(), "the list must show the opened containers' contents")
	c.Equal("Knife", selectedName(), "the target must be the one row selected")
	revealModifierTargets(sheet, tables)
	c.Equal("Knife", selectedName(), "revealing the target must leave it the one row selected")

	mgr.Undo()
	c.Equal(0, len(equipmentNamed("Knife", entity.CarriedEquipment).Modifiers), "undo must take the modifier back off")
	c.False(equipmentNamed("Backpack", entity.CarriedEquipment).IsOpen(), "undo must close the containers the command opened")
	c.False(equipmentNamed("Pouch", entity.CarriedEquipment).IsOpen(), "undo must close the containers the command opened")
	c.Equal(1, sheet.CarriedEquipment.Table.LastRowIndex(), "the list must be back to showing Rope and the closed Backpack")
	c.Equal("", selectedName(), "undo must put back the empty selection the command started from")
	mgr.Redo()
	c.Equal([]string{"Fine"}, appliedModifierNames(equipmentNamed("Knife", entity.CarriedEquipment).Modifiers),
		"redo must put the modifier back")
	c.True(equipmentNamed("Backpack", entity.CarriedEquipment).IsOpen(), "redo must open the target's containers again")
	c.True(equipmentNamed("Pouch", entity.CarriedEquipment).IsOpen(), "redo must open the target's containers again")
	c.Equal(3, sheet.CarriedEquipment.Table.LastRowIndex(), "the list must show the reopened containers' contents")
	c.Equal("Knife", selectedName(), "redo must leave the target selected, as the command did")
}

// TestModifierTargetCellFactoryIndentsNestedTargets verifies that the target prompt's rows are indented one level per
// container, a top-level target being drawn as the default factory draws it.
func TestModifierTargetCellFactoryIndentsNestedTargets(t *testing.T) {
	c := check.New(t)
	list := unison.NewList[modifierTargetChoice[*gurps.Trait]]()
	factory := &modifierTargetCellFactory[*gurps.Trait]{}
	leftInset := func(depth int) float32 {
		choice := modifierTargetChoice[*gurps.Trait]{label: "Row", depth: depth}
		cell := factory.CreateCell(list, choice, depth, unison.ThemeOnSurface, unison.ThemeSurface, false, false)
		border, ok := cell.AsPanel().Border().(*unison.EmptyBorder)
		c.True(ok, "a cell must keep the default factory's empty border")
		if !ok {
			return 0
		}
		return border.Insets().Left
	}
	plain := unison.DefaultCellFactory{}
	defaultCell := plain.CreateCell(list, modifierTargetChoice[*gurps.Trait]{label: "Row"}, 0, unison.ThemeOnSurface,
		unison.ThemeSurface, false, false)
	defaultBorder, ok := defaultCell.AsPanel().Border().(*unison.EmptyBorder)
	c.True(ok, "the default factory must give its cells an empty border for the comparison to mean anything")
	if !ok {
		return
	}
	base := defaultBorder.Insets().Left
	level := unison.FieldFont.LineHeight()
	c.Equal(base, leftInset(0), "a top-level target is not indented")
	c.Equal(base+level, leftInset(1), "a target inside a container is indented one level")
	c.Equal(base+2*level, leftInset(2), "a target two containers deep is indented two levels")
}

// TestPromptForSingleDestinationSkipsTheDialogForOneChoice verifies that promptForSingleDestination and
// PromptForDestination hand a lone choice back without a dialog and report no choice at all as nothing chosen. Neither
// can put up a dialog here, since there is no window.
func TestPromptForSingleDestinationSkipsTheDialogForOneChoice(t *testing.T) {
	c := check.New(t)
	sheet := newApplyModifierTestSheet(t)
	dest, ok := promptForSingleDestination([]FileBackedDockable{sheet})
	c.True(ok, "a lone choice must be taken without asking")
	c.Equal(any(sheet), any(dest), "the lone choice must be the one handed back")
	_, ok = promptForSingleDestination[FileBackedDockable](nil)
	c.False(ok, "no choice at all must be reported as nothing chosen")
	c.Equal([]FileBackedDockable{sheet}, PromptForDestination([]FileBackedDockable{sheet}),
		"a lone choice must come back as it is")
	c.Equal(0, len(PromptForDestination[FileBackedDockable](nil)), "no choice at all must come back empty")
}

// TestApplyModifierCommandNeedsASelectionAndADestination verifies that the command needs both a selection and an open
// destination, stays available from the library's filter field and on a filtered library, and is never offered by a
// modifier list inside an editor (see NewTableDockable).
func TestApplyModifierCommandNeedsASelectionAndADestination(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(registerActions)
	RegisterKnownFileTypes() // Docking asks for the sheet's icon, which comes from the file type registry.
	sheet := newApplyModifierTestSheet(t)
	swapForTest(t, &Workspace.DocumentDock, NewDocumentDock())
	mod := gurps.NewTraitModifier(nil, nil, false)
	mod.Name = "Ranged"
	melee := gurps.NewTraitModifier(nil, nil, false)
	melee.Name = "Melee"
	library := NewTraitModifierTableDockable("mods"+gurps.TraitModifiersExt, []*gurps.TraitModifier{mod, melee})
	canApply := func(table *unison.Table[*Node[*gurps.TraitModifier]]) bool {
		return table.CanPerformCmd(table, ApplyModifierItemID)
	}
	c.False(canApply(library.table), "nothing to apply and nowhere to apply it to")
	library.table.SelectAll()
	c.False(canApply(library.table), "nowhere to apply to without a destination")

	library.table.ClearSelection()
	Workspace.DocumentDock.DockTo(sheet, nil, side.Left)
	c.Equal(1, len(modifierDestinations(traitModifierTargetKind(), AllDockables())),
		"the docked sheet must be a destination for the check to mean anything")
	c.False(canApply(library.table), "nothing to apply without a selection")
	library.table.SelectAll()
	c.True(canApply(library.table), "with a selection and a destination, the command must be available")

	c.True(library.filterField.CanPerformCmd(library.filterField, ApplyModifierItemID),
		"the command must stay available while the focus is in the library's filter field")

	// The library's rows are only read, so a filtered library is still a source, unlike a filtered destination.
	library.table.ApplyHierarchicalFilter(func(row *Node[*gurps.TraitModifier]) bool { return row.Data().Name != "Ranged" })
	c.True(library.table.IsFiltered(), "the library must be filtered for the check to mean anything")
	c.True(library.table.HasSelection(), "the modifier left in view must still be selected for the check to mean anything")
	c.True(canApply(library.table), "a filtered modifier library must still offer the command")
	library.table.ApplyHierarchicalFilter(nil)

	equipmentMod := gurps.NewEquipmentModifier(nil, nil, false)
	equipmentMod.Name = "Fine"
	equipmentLibrary := NewEquipmentModifierTableDockable("mods"+gurps.EquipmentModifiersExt,
		[]*gurps.EquipmentModifier{equipmentMod})
	canApplyEquipment := func() bool {
		return equipmentLibrary.table.CanPerformCmd(equipmentLibrary.table, ApplyModifierItemID)
	}
	c.False(canApplyEquipment(), "nothing to apply without a selection")
	equipmentLibrary.table.SelectAll()
	c.True(canApplyEquipment(), "an equipment modifier library offers the command just as a trait modifier library does")

	root := &stubRebuildable{}
	root.Self = root
	editorModifiers := []*gurps.TraitModifier{mod.Clone(gurps.LibraryFile{}, sheet.Entity(), nil, gurps.Copy)}
	editor := newTraitModifiersPanel(root, sheet.Entity(), &editorModifiers)
	editor.table.SelectAll()
	c.True(editor.table.HasSelection(), "the editor's modifier must be selected for the check to mean anything")
	c.False(canApply(editor.table), "a modifier list inside an editor must not offer the command")
}

// TestApplySelectedModifiersAcrossBothEquipmentLists runs the whole command for equipment modifiers with the prompts
// answered non-interactively: the docked sheet is offered and picked, the target prompt is headed for the sheet's
// equipment, and choosing Rope and Coin attaches the modifier to both as one undoable edit, leaving each selected in
// its own list.
func TestApplySelectedModifiersAcrossBothEquipmentLists(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(registerActions)
	RegisterKnownFileTypes() // Docking asks for the sheet's icon, which comes from the file type registry.
	forbidModifierPrompts(t)
	sheet := newApplyModifierTestSheet(t)
	entity := sheet.Entity()
	swapForTest(t, &Workspace.DocumentDock, NewDocumentDock())
	Workspace.DocumentDock.DockTo(sheet, nil, side.Left)
	mod := gurps.NewEquipmentModifier(nil, nil, false)
	mod.Name = "Fine"
	mod.Disabled = true
	library := NewEquipmentModifierTableDockable("mods"+gurps.EquipmentModifiersExt, []*gurps.EquipmentModifier{mod})
	library.table.SelectAll()
	var offered []string
	swapForTest(t, &promptForModifierDestination, func(choices []FileBackedDockable) (FileBackedDockable, bool) {
		for _, choice := range choices {
			offered = append(offered, choice.Title())
		}
		return sheet, true
	})
	var header string
	var labels []string
	swapForTest(t, &showModifierTargetsPrompt, func(h string, list unison.Paneler, _ ...*unison.Label) bool {
		header = h
		targets, ok := list.(*unison.List[modifierTargetChoice[*gurps.Equipment]])
		if !ok {
			t.Errorf("the target prompt must hold a list of equipment choices, not a %T", list)
			return false
		}
		for i := range targets.Count() {
			choice := targets.DataAtIndex(i)
			labels = append(labels, choice.String())
			if name := choice.target.Name; name == "Rope" || name == "Coin" {
				targets.Select(true, i)
			}
		}
		return true
	})

	applySelectedModifiers(library.table, equipmentModifierTargetKind())

	c.Equal([]string{sheet.Title()}, offered, "the docked sheet must be offered as the destination")
	c.Equal("Choose the equipment in "+sheet.Title()+" that should receive the selected modifiers:", header)
	c.Equal([]string{
		"Rope (in Carried Equipment)",
		"Backpack (in Carried Equipment)",
		"Pouch (in Backpack, Carried Equipment)",
		"Knife (in Pouch, Backpack, Carried Equipment)",
		"Coin (in Other Equipment)",
	}, labels, "the target prompt must offer the equipment of both lists")
	c.Equal([]string{"Fine"}, appliedModifierNames(equipmentNamed("Rope", entity.CarriedEquipment).Modifiers))
	c.Equal([]string{"Fine"}, appliedModifierNames(equipmentNamed("Coin", entity.OtherEquipment).Modifiers))
	c.False(equipmentNamed("Rope", entity.CarriedEquipment).Modifiers[0].Disabled, "the applied copy is switched on")
	c.Equal(0, len(equipmentNamed("Backpack", entity.CarriedEquipment).Modifiers), "only the chosen targets are touched")
	c.Equal(0, len(equipmentNamed("Knife", entity.CarriedEquipment).Modifiers), "only the chosen targets are touched")
	selectedNames := func(table *unison.Table[*Node[*gurps.Equipment]]) []string {
		rows := table.SelectedRows(false)
		names := make([]string, 0, len(rows))
		for _, row := range rows {
			names = append(names, row.Data().Name)
		}
		return names
	}
	c.Equal([]string{"Rope"}, selectedNames(sheet.CarriedEquipment.Table), "the carried target must be left selected")
	c.Equal([]string{"Coin"}, selectedNames(sheet.OtherEquipment.Table), "the other target must be left selected")
	mgr := unison.UndoManagerFor(sheet.CarriedEquipment.Table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	c.True(sheet.Modified(), "the sheet has been changed")
	c.Equal("Undo "+applyModifierAction.Title, mgr.UndoTitle(), "the edit is named for the command")
	mgr.Undo()
	c.False(mgr.CanUndo(), "the command must make exactly one edit, though it touched two lists")
	c.Equal(0, len(equipmentNamed("Rope", entity.CarriedEquipment).Modifiers), "undo must take the modifier back off")
	c.Equal(0, len(equipmentNamed("Coin", entity.OtherEquipment).Modifiers), "undo must take the modifier back off")
}
