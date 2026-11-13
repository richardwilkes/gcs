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
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/role"
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
	modifier := gurps.NewTraitModifier(nil, nil, false)
	modifier.Name = "Ranged"
	modifier.SetDisabled(true)
	container := gurps.NewTraitModifier(nil, nil, true)
	container.Name = "Group"
	inner := gurps.NewTraitModifier(nil, container, false)
	inner.Name = "Inner"
	inner.SetDisabled(true)
	container.Children = []*gurps.TraitModifier{inner}

	c.True(attachModifierClones("", tables, owner, []*gurps.Trait{first, second}, []*gurps.TraitModifier{modifier, container},
		gurps.LibraryFile{}))

	c.Equal([]string{"Ranged", "Group"}, appliedModifierNames(first.Modifiers))
	c.Equal([]string{"Ranged", "Group"}, appliedModifierNames(second.Modifiers))
	c.NotEqual(modifier, first.Modifiers[0], "the target must get a clone, not the original")
	c.NotEqual(first.Modifiers[0], second.Modifiers[0], "each target must get a clone of its own")
	c.Equal(first, first.Modifiers[0].Target(), "the clone must know the trait it was attached to")
	c.Equal(second, second.Modifiers[0].Target(), "the clone must know the trait it was attached to")
	c.False(first.Modifiers[0].IsDisabled(), "the clone is switched on")
	c.True(first.Modifiers[1].Children[0].IsDisabled(),
		"the modifiers within a container are left as the library has them, since nothing asked about them")
	c.Nil(modifier.Target(), "the original must be left untouched")
	c.True(modifier.IsDisabled(), "the original must be left untouched")
	c.True(inner.IsDisabled(), "the original must be left untouched")
}

// TestAttachModifierClonesAsksAboutContainersOnASheet verifies that a container among the modifiers applied to a
// sheet's rows brings up the modifier prompt once per target, with that target's clones and the ones pointed at
// directly already enabled, and that the answers are kept. Canceling the prompt for the first target stops there,
// skips the nameables prompt, and leaves every target as it was with nothing shown or reported.
func TestAttachModifierClonesAsksAboutContainersOnASheet(t *testing.T) {
	newModifiers := func() []*gurps.TraitModifier {
		ranged := gurps.NewTraitModifier(nil, nil, false)
		ranged.Name = "Ranged"
		ranged.SetDisabled(true)
		group := gurps.NewTraitModifier(nil, nil, true)
		group.Name = "Group"
		for _, name := range []string{"First", "Second", "Third"} {
			child := gurps.NewTraitModifier(nil, group, false)
			child.Name = name
			child.SetDisabled(name != "Second")
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
		uxtest.SwapForTest(t, &promptForTraitModifiers, func(info *modifierPromptInfo, modifiers []*gurps.TraitModifier) (changed, canceled bool) {
			title := info.name
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
				modifiers[1].Children[0].SetDisabled(false)
				modifiers[1].Children[1].SetDisabled(true)
				return true, false
			}
			return false, false
		})

		c.True(attachModifierClones("", []*unison.Table[*Node[*gurps.Trait]]{sheet.Traits.Table}, entity,
			[]*gurps.Trait{alpha, beta}, newModifiers(), gurps.LibraryFile{}))
		c.Equal([]prompt{wantPrompt("Alpha"), wantPrompt("Beta")}, *prompts,
			"each target must be asked about its own copies, the ones pointed at directly already switched on")
		c.Equal([]string{"Ranged", "Group"}, appliedModifierNames(alpha.Modifiers))
		c.False(alpha.Modifiers[1].Children[0].IsDisabled(), "Alpha's answer must be kept")
		c.True(alpha.Modifiers[1].Children[1].IsDisabled(), "Alpha's answer must be kept")
		c.True(alpha.Modifiers[1].Children[2].IsDisabled(), "Alpha's answer must be kept")
		c.Equal([]string{"Ranged", "Group"}, appliedModifierNames(beta.Modifiers))
		c.True(beta.Modifiers[1].Children[0].IsDisabled(), "Beta keeps the library's choice")
		c.False(beta.Modifiers[1].Children[1].IsDisabled(), "Beta keeps the library's choice")
		c.True(beta.Modifiers[1].Children[2].IsDisabled(), "Beta keeps the library's choice")
		c.True(sheet.Modified(), "the sheet has been changed")
	})

	t.Run("the rest picked at random", func(t *testing.T) {
		c := check.New(t)
		sheet := newApplyModifierTestSheet(t)
		entity := sheet.Entity()
		alpha, beta := entity.Traits[0], entity.Traits[1]
		var asked []string
		uxtest.SwapForTest(t, &promptForTraitModifiers, func(info *modifierPromptInfo, modifiers []*gurps.TraitModifier) (changed, canceled bool) {
			asked = append(asked, info.name)
			info.random.rest = true
			modifiers[0].Children[1].SetDisabled(false)
			return true, false
		})
		choice := newTraitModifierChoiceFor(nil, true, []string{"Low", "High"})
		c.True(attachModifierClones("", []*unison.Table[*Node[*gurps.Trait]]{sheet.Traits.Table}, entity,
			[]*gurps.Trait{alpha, beta}, []*gurps.TraitModifier{choice}, gurps.LibraryFile{}))
		c.Equal([]string{"Alpha"}, asked, "once the rest are to be picked at random, Beta isn't asked for its pick")
		c.True(gurps.ModifierChoiceIsResolved(beta.Modifiers[0]), "which is made for it")
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
		uxtest.SwapForTest(t, &promptForNameables, slicedNameablesPrompt(func(_ []string, _ []map[string]string, _ [][]string) bool {
			nameablesShown++
			return true
		}))
		modifiers := append(newModifiers(), newSwitchableTraitModifier("@Material@ Coating"))

		c.False(attachModifierClones("", []*unison.Table[*Node[*gurps.Trait]]{stale}, entity,
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
	modifier := newSwitchableTraitModifier("Ranged")

	c.True(attachModifierClones("", []*unison.Table[*Node[*gurps.Trait]]{stale}, entity,
		[]*gurps.Trait{entity.Traits[0], entity.Traits[1]}, []*gurps.TraitModifier{modifier}, gurps.LibraryFile{}))
	c.False(entity.Traits[0].Modifiers[0].IsDisabled(), "the clone is switched on")
	c.False(entity.Traits[1].Modifiers[0].IsDisabled(), "the clone is switched on")
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
	uxtest.SwapForTest(t, &promptForNameables, slicedNameablesPrompt(func(_ []string, _ []map[string]string, _ [][]string) bool {
		shown++
		c.Equal(map[string]string{"Foo": "Bar", "Material": "Steel"}, trait.Replacements,
			"the modifier's replacements must have been moved onto the trait by the time the prompt is up")
		return false
	}))

	c.False(attachModifierClones("", []*unison.Table[*Node[*gurps.Trait]]{sheet.Traits.Table}, entity,
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
	_, user := uxtest.UseTestLibraries(t, c)
	group := gurps.NewTraitModifier(nil, nil, true)
	group.Name = "Group"
	inner := gurps.NewTraitModifier(nil, group, false)
	inner.Name = "Inner"
	group.Children = []*gurps.TraitModifier{inner}
	library := NewTraitModifierTableDockable(filepath.Join(user.Path(false), "Sub", "mods"+gurps.TraitModifiersExt),
		[]*gurps.TraitModifier{group})
	library.table.SetSelectionMap(map[tid.TID]bool{group.ID(): true, inner.ID(): true})
	c.Equal(2, len(library.table.SelectedRows(false)),
		"both the container and the modifier within it must be selected for the check to mean anything")
	uxtest.SwapForTest(t, &promptForModifierDestination, func(_ promptOperation, _ []FileBackedDockable) (FileBackedDockable, bool) {
		return sheet, true
	})
	uxtest.SwapForTest(t, &showModifierTargetsPrompt, func(_ promptOperation, _ string, list unison.Paneler, _ ...*unison.Label) bool {
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

func TestCanReceiveModifiersNeedsATargetAProviderAndAnUndoManager(t *testing.T) {
	c := check.New(t)
	sheet := newApplyModifierTestSheet(t)
	table := sheet.Traits.Table
	c.True(canReceiveModifiers(table), "a sheet's list with traits in it can receive modifiers")

	// A granted row can't take modifiers (see gurps.CanTakeModifiers), so a list holding nothing else offers nothing.
	entity := sheet.Entity()
	goggles := gurps.NewEquipment(entity, nil, false)
	goggles.Name = "Goggles"
	g := gurps.NewGrant(feature.GrantTrait)
	nightVision := gurps.NewTrait(nil, nil, false)
	nightVision.Name = "Night Vision"
	g.SetItem(nightVision, gurps.LibraryFile{})
	goggles.Features = gurps.Features{g}
	entity.CarriedEquipment = append(entity.CarriedEquipment, goggles)
	entity.Traits = nil
	sheet.Rebuild(true)
	table = sheet.Traits.Table
	c.Equal(1, table.RootRowCount(), "the grant must put its row in the list for the check to mean anything")
	c.False(canReceiveModifiers(table), "a list holding only granted rows cannot receive modifiers")
	alpha := gurps.NewTrait(entity, nil, false)
	alpha.Name = "Alpha"
	entity.Traits = append(entity.Traits, alpha)
	sheet.Rebuild(true)
	table = sheet.Traits.Table
	c.True(canReceiveModifiers(table), "the list must receive modifiers again once it holds a row that can take them")

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
	modifier := newSwitchableEquipmentModifier("Fine @Quality@")
	forbidModifierPrompts(t)
	shown := 0
	uxtest.SwapForTest(t, &promptForNameables, slicedNameablesPrompt(func(titles []string, _ []map[string]string, _ [][]string) bool {
		shown++
		c.Equal(2, len(titles), "one prompt covers the copy each target got")
		c.Equal(0, counter.count, "nothing may be shown before the prompt is answered")
		return false
	}))
	rope := equipmentNamed("Rope", entity.CarriedEquipment)
	coin := equipmentNamed("Coin", entity.OtherEquipment)

	c.False(applyModifiersTo([]*unison.Table[*Node[*gurps.Equipment]]{carried, other},
		[]*gurps.Equipment{rope, coin}, []*gurps.EquipmentModifier{modifier}, gurps.LibraryFile{}))
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
	uxtest.SwapForTest(t, &promptForNameables, slicedNameablesPrompt(func(titles []string, nameables []map[string]string, _ [][]string) bool {
		shown++
		c.Equal(2, len(titles), "one prompt covers both modifiers")
		for _, one := range nameables {
			for k := range one {
				one[k] = answers[k]
			}
		}
		return true
	}))

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
		uxtest.SwapForTest(t, &promptForNameables, slicedNameablesPrompt(func(titles []string, _ []map[string]string, _ [][]string) bool {
			t.Errorf("the nameables prompt must not be shown, but was shown for %v", titles)
			return false
		}))
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
		modifier := gurps.NewTraitModifier(nil, nil, false)
		modifier.Name = "@Material@ Coating"
		modifier.SetDisabled(true)

		c.True(applyModifiersTo([]*unison.Table[*Node[*gurps.Trait]]{template.Traits.Table},
			[]*gurps.Trait{trait}, []*gurps.TraitModifier{modifier}, gurps.LibraryFile{}))
		c.Equal([]string{"@Material@ Coating"}, appliedModifierNames(trait.Modifiers))
		c.False(trait.Modifiers[0].IsDisabled(), "the clone is switched on")
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
		modifier := gurps.NewEquipmentModifier(nil, nil, false)
		modifier.Name = "@Material@ Coating"
		modifier.SetDisabled(true)

		c.True(applyModifiersTo([]*unison.Table[*Node[*gurps.Equipment]]{loot.Equipment.Table},
			[]*gurps.Equipment{item}, []*gurps.EquipmentModifier{modifier}, gurps.LibraryFile{}))
		c.Equal([]string{"@Material@ Coating"}, appliedModifierNames(item.Modifiers))
		c.False(item.Modifiers[0].IsDisabled(), "the clone is switched on")
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
		uxtest.SwapForTest(t, &Workspace.DocumentDock, NewDocumentDock())
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
		modifier := gurps.NewTraitModifier(nil, nil, false)
		modifier.Name = "@Material@ Coating"
		modifier.SetDisabled(true)

		c.True(applyModifiersTo([]*unison.Table[*Node[*gurps.Trait]]{library.table}, []*gurps.Trait{trait},
			[]*gurps.TraitModifier{modifier}, gurps.LibraryFile{}))
		c.Equal([]string{"@Material@ Coating"}, appliedModifierNames(trait.Modifiers))
		c.False(trait.Modifiers[0].IsDisabled(), "the clone is switched on")
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
	modifier := gurps.NewEquipmentModifier(nil, nil, false)
	modifier.Name = "Fine"
	tables := []*unison.Table[*Node[*gurps.Equipment]]{table}
	selectedName := func() string {
		live := sheet.CarriedEquipment.Table
		if row := live.FirstSelectedRowIndex(); row != -1 && len(live.SelectedRows(false)) == 1 {
			return live.RowFromIndex(row).Data().Name
		}
		return ""
	}

	c.True(applyModifiersTo(tables, []*gurps.Equipment{knife}, []*gurps.EquipmentModifier{modifier}, gurps.LibraryFile{}))
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
// promptForDestinations hand a lone choice back without a dialog and report no choice at all as nothing chosen. Neither
// can put up a dialog here, since there is no window.
func TestPromptForSingleDestinationSkipsTheDialogForOneChoice(t *testing.T) {
	c := check.New(t)
	sheet := newApplyModifierTestSheet(t)
	dest, ok := promptForSingleDestination(promptOperation{}, []FileBackedDockable{sheet})
	c.True(ok, "a lone choice must be taken without asking")
	c.Equal(any(sheet), any(dest), "the lone choice must be the one handed back")
	_, ok = promptForSingleDestination[FileBackedDockable](promptOperation{}, nil)
	c.False(ok, "no choice at all must be reported as nothing chosen")
	c.Equal([]FileBackedDockable{sheet}, promptForDestinations(promptOperation{}, []FileBackedDockable{sheet}),
		"a lone choice must come back as it is")
	c.Equal(0, len(promptForDestinations[FileBackedDockable](promptOperation{}, nil)), "no choice at all must come back empty")
}

// TestApplyModifierCommandNeedsASelectionAndADestination verifies that the command needs both a selection and an open
// destination, stays available from the library's filter field and on a filtered library, and is never offered by a
// modifier list inside an editor (see NewTableDockable).
func TestApplyModifierCommandNeedsASelectionAndADestination(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(registerActions)
	RegisterKnownFileTypes() // Docking asks for the sheet's icon, which comes from the file type registry.
	sheet := newApplyModifierTestSheet(t)
	uxtest.SwapForTest(t, &Workspace.DocumentDock, NewDocumentDock())
	modifier := gurps.NewTraitModifier(nil, nil, false)
	modifier.Name = "Ranged"
	melee := gurps.NewTraitModifier(nil, nil, false)
	melee.Name = "Melee"
	library := NewTraitModifierTableDockable("mods"+gurps.TraitModifiersExt, []*gurps.TraitModifier{modifier, melee})
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
	editorModifiers := []*gurps.TraitModifier{modifier.Clone(gurps.LibraryFile{}, sheet.Entity(), nil, gurps.Copy)}
	editor := newTraitModifiersPanel(root, sheet.Entity(), gurps.NewTrait(sheet.Entity(), nil, false), &editorModifiers,
		&gurps.TraitEditData{})
	editor.table.SelectAll()
	c.True(editor.table.HasSelection(), "the editor's modifier must be selected for the check to mean anything")
	c.False(canApply(editor.table), "a modifier list inside an editor must not offer the command")
}

// TestApplySelectedModifiersAcrossBothEquipmentLists runs the whole command for equipment modifiers with the prompts
// answered non-interactively: the docked sheet is offered and picked, the target prompt is headed for the sheet's
// equipment, and choosing Rope and Coin attaches the modifier to both as one edit that bumps the entity's modification
// time, leaving each selected in its own list. One Undo takes the modifier off both and leaves the sheet unmodified,
// and one Redo puts it back on both.
func TestApplySelectedModifiersAcrossBothEquipmentLists(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(registerActions)
	RegisterKnownFileTypes() // Docking asks for the sheet's icon, which comes from the file type registry.
	forbidModifierPrompts(t)
	sheet := newApplyModifierTestSheet(t)
	entity := sheet.Entity()
	entity.ModifiedOn = jio.Time{}
	uxtest.SwapForTest(t, &Workspace.DocumentDock, NewDocumentDock())
	Workspace.DocumentDock.DockTo(sheet, nil, side.Left)
	modifier := gurps.NewEquipmentModifier(nil, nil, false)
	modifier.Name = "Fine"
	modifier.SetDisabled(true)
	library := NewEquipmentModifierTableDockable("mods"+gurps.EquipmentModifiersExt, []*gurps.EquipmentModifier{modifier})
	library.table.SelectAll()
	var offered []string
	uxtest.SwapForTest(t, &promptForModifierDestination, func(_ promptOperation, choices []FileBackedDockable) (FileBackedDockable, bool) {
		for _, choice := range choices {
			offered = append(offered, choice.Title())
		}
		return sheet, true
	})
	var header string
	var labels []string
	uxtest.SwapForTest(t, &showModifierTargetsPrompt, func(_ promptOperation, h string, list unison.Paneler, _ ...*unison.Label) bool {
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
	c.False(equipmentNamed("Rope", entity.CarriedEquipment).Modifiers[0].IsDisabled(), "the applied copy is switched on")
	c.False(equipmentNamed("Coin", entity.OtherEquipment).Modifiers[0].IsDisabled(), "the applied copy is switched on")
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
	c.NotEqual(jio.Time{}, entity.ModifiedOn, "the change must be reported")
	c.Equal("Undo "+applyModifierAction.Title, mgr.UndoTitle(), "the edit is named for the command")
	mgr.Undo()
	c.False(mgr.CanUndo(), "the command must make exactly one edit, though it touched two lists")
	c.Equal(0, len(equipmentNamed("Rope", entity.CarriedEquipment).Modifiers), "undo must take the modifier back off")
	c.Equal(0, len(equipmentNamed("Coin", entity.OtherEquipment).Modifiers), "undo must take the modifier back off")
	c.False(sheet.Modified(), "undo must leave the sheet as it was")
	mgr.Redo()
	c.Equal([]string{"Fine"}, appliedModifierNames(equipmentNamed("Rope", entity.CarriedEquipment).Modifiers))
	c.Equal([]string{"Fine"}, appliedModifierNames(equipmentNamed("Coin", entity.OtherEquipment).Modifiers))
}

// rowInView reports whether the row at the given index of the table is within the part of the view showing it that can
// be seen, judged top to bottom only, since a row is as wide as its table and the table may be wider than the view.
func rowInView[T gurps.Node[T]](table *unison.Table[*Node[T]], row int) bool {
	view := table.ViewForRow(row)
	frame := view.RectToRoot(view.RowFrame(row))
	visible := visibleRect(view.AsPanel())
	return frame.Y >= visible.Y-viewSlop && frame.Bottom() <= visible.Bottom()+viewSlop
}

// TestApplyModifierHeadless drives the Apply Modifier command through its menus and dialogs with a sheet, a template
// and a trait modifier library open: the command is in the modifier list's context menu but not a trait list's; the
// destination prompt allows one choice, with the first selected so that Return alone accepts it; the target prompt
// names each trait with its container and allows several; canceling either prompt changes nothing; accepting attaches
// an enabled copy of the selected modifier to the chosen traits, opens their container and focuses the traits list
// with them selected and in view; Undo and Redo take the copies off and put them back, container included; and with
// only the sheet open the command goes straight to the target prompt.
func TestApplyModifierHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	screen.EnableAccessibility()
	forbidModifierPrompts(t)

	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	traitNames := []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon"}
	var delta *gurps.Trait
	screen.Do(func() {
		entity := sheet.Entity()
		entity.Traits = nil
		for _, name := range traitNames[:3] {
			trait := gurps.NewTrait(entity, nil, false)
			trait.Name = name
			entity.Traits = append(entity.Traits, trait)
		}
		// A closed container holding a trait, so that the prompt has a nested row to name and the command has a
		// container to open.
		delta = gurps.NewTrait(entity, nil, true)
		delta.Name = "Delta"
		delta.SetOpen(false)
		epsilon := gurps.NewTrait(entity, delta, false)
		epsilon.Name = "Epsilon"
		delta.Children = []*gurps.Trait{epsilon}
		entity.Traits = append(entity.Traits, delta)
		entity.Recalculate()
		sheet.Rebuild(true)
		// A new sheet counts as modified until saved, so it is marked unmodified and its timestamp cleared for the test
		// to tell whether the command changed and reported anything.
		sheet.markUnmodified()
		entity.ModifiedOn = jio.Time{}
	})
	template, ok := uxtest.OpenedByAction(t, screen, newCharacterTemplateAction).(*Template)
	if !ok {
		t.Fatal("New Character Template must open a template")
	}
	screen.Do(func() {
		// A template with nothing in it has nothing to apply a modifier to, so it would not be offered.
		trait := gurps.NewTrait(template.template, nil, false)
		trait.Name = "Template Trait"
		template.template.Traits = []*gurps.Trait{trait}
		template.Rebuild(true)
		template.markUnmodified() // So that closing it later asks nothing.
	})
	ranged := gurps.NewTraitModifier(nil, nil, false)
	ranged.Name = "Ranged"
	ranged.SetDisabled(true)
	melee := gurps.NewTraitModifier(nil, nil, false)
	melee.Name = "Melee"
	var library *TableDockable[*gurps.TraitModifier]
	screen.Do(func() {
		library = NewTraitModifierTableDockable(filepath.Join(t.TempDir(), "mods"+gurps.TraitModifiersExt),
			[]*gurps.TraitModifier{ranged, melee})
		DisplayNewDockable(library)
	})
	if library == nil {
		t.Fatal("the modifier library could not be opened")
	}
	// traitModifierNames returns the modifier names of the sheet's trait with the given name, failing the test if there
	// is none.
	traitModifierNames := func(name string) []string {
		t.Helper()
		var names []string
		found := false
		screen.Do(func() {
			gurps.Traverse(func(trait *gurps.Trait) bool {
				if trait.Name == name {
					names = appliedModifierNames(trait.Modifiers)
					found = true
				}
				return found
			}, false, false, sheet.Entity().Traits...)
		})
		if !found {
			t.Fatalf("the sheet has no trait named %q", name)
		}
		return names
	}
	nothingApplied := func(why string) {
		t.Helper()
		var canUndo, modified bool
		var modifiedOn jio.Time
		screen.Do(func() {
			canUndo = sheet.undoMgr.CanUndo()
			modified = sheet.Modified()
			modifiedOn = sheet.Entity().ModifiedOn
		})
		c.False(canUndo, "%s: there must be nothing to undo", why)
		c.False(modified, "%s: the sheet must be left unmodified", why)
		c.Equal(jio.Time{}, modifiedOn, "%s: nothing must have been reported to the sheet", why)
		for _, name := range traitNames {
			c.Equal(0, len(traitModifierNames(name)), "%s: %s must have no modifiers", why, name)
		}
	}
	windowCount := func() int {
		var count int
		screen.Do(func() { count = len(unison.Windows()) })
		return count
	}
	// selectModifier clicks the library's first row, Ranged, which selects it alone and puts the focus on the list.
	selectModifier := func() {
		t.Helper()
		screen.Click(cellScreenPoint(t, screen, wnd, library.table, 0, gurps.TraitModifierDescriptionColumn))
		var enabled bool
		screen.Do(func() {
			enabled = wnd.Focus() == library.table.AsPanel() && library.table.HasSelection() &&
				applyModifierAction.Enabled(nil)
		})
		c.True(enabled, "with a modifier selected and a destination open, the command must be available")
	}
	// destinationPrompt chooses the command from the Item menu and returns the destination prompt it puts up, along
	// with the prompt's list and the row of the sheet in it.
	destinationPrompt := func() (dialogWnd *unison.Window, dialog *unison.Dialog, list *unison.List[FileBackedDockable], sheetRow int) {
		t.Helper()
		selectModifier()
		uxtest.ChooseMenuBarItem(t, screen, wnd, "Item", applyModifierAction.Title)
		dialogWnd, dialog = modalDialog(t, screen, wnd)
		sheetRow = -1
		screen.Do(func() {
			if lists := uxtest.PanelsOfType[*unison.List[FileBackedDockable]](dialogWnd.Content()); len(lists) == 1 {
				list = lists[0]
				for i := range list.Count() {
					if list.DataAtIndex(i).AsPanel() == sheet.AsPanel() {
						sheetRow = i
					}
				}
			}
		})
		if list == nil {
			t.Fatal("the destination prompt must hold one list of destinations")
		}
		if sheetRow == -1 {
			t.Fatal("the destination prompt must offer the sheet")
		}
		return dialogWnd, dialog, list, sheetRow
	}
	// chooseSheet picks the sheet in the destination prompt and accepts it with Return.
	chooseSheet := func(dialogWnd *unison.Window, _ *unison.Dialog, list *unison.List[FileBackedDockable], sheetRow int) {
		t.Helper()
		var rowPt geom.Point
		screen.Do(func() { rowPt = screenPoint(dialogWnd, list.RectToRoot(list.RowRect(sheetRow)).Center()) })
		screen.Click(rowPt)
		var selected int
		screen.Do(func() { selected = list.Selection.FirstSet() })
		c.Equal(sheetRow, selected, "clicking the sheet's row must select it")
		screen.KeyPress(unison.KeyReturn, mod.None)
	}
	// targetPromptFor returns the target prompt that is up and its list, checking that its header names the given
	// destination.
	targetPromptFor := func(title string) (*unison.Window, *unison.Dialog, *unison.List[modifierTargetChoice[*gurps.Trait]]) {
		t.Helper()
		dialogWnd, dialog := modalDialog(t, screen, wnd)
		var targets *unison.List[modifierTargetChoice[*gurps.Trait]]
		var headerFound bool
		screen.Do(func() {
			if lists := uxtest.PanelsOfType[*unison.List[modifierTargetChoice[*gurps.Trait]]](dialogWnd.Content()); len(lists) == 1 {
				targets = lists[0]
			}
			for _, label := range uxtest.PanelsOfType[*unison.Label](dialogWnd.Content()) {
				if strings.Contains(label.String(), "Choose the traits in "+title) {
					headerFound = true
				}
			}
		})
		if targets == nil {
			t.Fatal("the target prompt must hold one list of traits")
		}
		c.True(headerFound, "the prompt's header must name the destination that was chosen, %s", title)
		return dialogWnd, dialog, targets
	}
	// targetPrompt returns the sheet's target prompt, which must be the one showing.
	targetPrompt := func() (*unison.Window, *unison.Dialog, *unison.List[modifierTargetChoice[*gurps.Trait]]) {
		t.Helper()
		return targetPromptFor(sheet.Title())
	}
	// sheetState records what the command leaves behind on the sheet.
	type sheetState struct {
		selected  []string
		deltaOpen bool
		focused   bool
	}
	currentSheetState := func() sheetState {
		var s sheetState
		screen.Do(func() {
			s.deltaOpen = delta.IsOpen()
			s.focused = wnd.Focus() == sheet.Traits.Table.AsPanel()
			for _, row := range sheet.Traits.Table.SelectedRows(false) {
				s.selected = append(s.selected, row.Data().Name)
			}
		})
		return s
	}

	// The command is offered where the modifier is, and not on a list it can't work from.
	selectModifier()
	screen.KeyPress(unison.KeyF10, mod.Shift)
	c.True(slices.Contains(contextMenuTitles(t, screen, wnd), applyModifierAction.Title),
		"the modifier list's context menu must offer the command")
	closeContextMenu(c, screen, wnd)
	screen.Click(cellScreenPoint(t, screen, wnd, sheet.Traits.Table, 0, gurps.TraitDescriptionColumn))
	screen.KeyPress(unison.KeyF10, mod.Shift)
	traitsMenu := contextMenuTitles(t, screen, wnd)
	c.True(len(traitsMenu) != 0, "the traits list must have opened a context menu for the check to mean anything")
	c.False(slices.Contains(traitsMenu, applyModifierAction.Title), "a list of traits has no modifier to apply")
	closeContextMenu(c, screen, wnd)

	// The destination prompt offers the sheet and the template, one choosable, with the first already selected so that
	// Return alone accepts it. Canceling either prompt changes nothing.
	dialogWnd, _, list, _ := destinationPrompt()
	var titles []string
	var multiple, headerFound bool
	var preselected int
	screen.Do(func() {
		multiple = list.AllowMultipleSelection()
		preselected = list.Selection.FirstSet()
		for i := range list.Count() {
			titles = append(titles, list.DataAtIndex(i).Title())
		}
		for _, label := range uxtest.PanelsOfType[*unison.Label](dialogWnd.Content()) {
			if label.String() == "Choose a destination:" {
				headerFound = true
			}
		}
	})
	c.False(multiple, "only one destination may be chosen")
	c.Equal(0, preselected, "the first destination starts out selected")
	c.True(headerFound, "the prompt asks for a destination")
	first := titles[0]
	slices.Sort(titles)
	wanted := []string{sheet.Title(), template.Title()}
	slices.Sort(wanted)
	c.Equal(wanted, titles, "the prompt offers everything open that holds traits, and nothing else")
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Equal(1, windowCount(), "the prompt has been dismissed")
	nothingApplied("canceling the destination prompt")
	destinationPrompt()
	screen.KeyPress(unison.KeyReturn, mod.None)
	targetPromptFor(first)
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Equal(1, windowCount(), "the prompt has been dismissed")
	nothingApplied("canceling the target prompt Return led to")

	// The target prompt offers the sheet's traits by name, the nested one with its container, and lets several be
	// chosen. Canceling it changes nothing.
	chooseSheet(destinationPrompt())
	dialogWnd, _, targets := targetPrompt()
	var labels []string
	screen.Do(func() {
		multiple = targets.AllowMultipleSelection()
		for i := range targets.Count() {
			labels = append(labels, targets.DataAtIndex(i).String())
		}
	})
	c.True(multiple, "more than one target may be chosen")
	c.Equal([]string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon (in Delta)"}, labels,
		"the prompt offers the sheet's traits, naming the container a nested one sits in")
	// A row hands its text over to the label described within it, so what is heard for a row is the label's text.
	var spoken []string
	tree := screen.AccessibilityTree(dialogWnd)
	tree.Walk(func(n *accessibility.Node) bool {
		if n.Role == role.ListItem {
			text := n.Name
			for _, id := range tree.UnignoredChildren(n.ID) {
				if child := tree.Node(id); child != nil {
					text += child.Name
				}
			}
			spoken = append(spoken, text)
		}
		return true
	})
	c.Equal(labels, spoken, "a screen reader hears each row by its label, container and all")
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Equal(1, windowCount(), "the prompt has been dismissed")
	nothingApplied("canceling the target prompt")

	// Accepting the target prompt with Alpha and Epsilon chosen applies the selected modifier to both, opening Delta to
	// show Epsilon.
	chooseSheet(destinationPrompt())
	_, _, targets = targetPrompt()
	screen.Do(func() { targets.Select(false, 0, 4) })
	screen.KeyPress(unison.KeyReturn, mod.None)
	var modified, enabled bool
	var modifiedOn jio.Time
	var undoTitle string
	var selectedInView, selectedRows int
	screen.Do(func() {
		modified = sheet.Modified()
		modifiedOn = sheet.Entity().ModifiedOn
		undoTitle = sheet.undoMgr.UndoTitle()
		for _, trait := range sheet.Entity().Traits {
			if trait.Name == "Alpha" && len(trait.Modifiers) == 1 {
				enabled = trait.Modifiers[0].Enabled()
			}
		}
		table := sheet.Traits.Table
		for row := range table.LastRowIndex() + 1 {
			if table.IsRowSelected(row) {
				selectedRows++
				if rowInView(table, row) {
					selectedInView++
				}
			}
		}
	})
	c.Equal(1, windowCount(), "the prompt has been dismissed")
	c.Equal([]string{"Ranged"}, traitModifierNames("Alpha"), "only the selected modifier is applied")
	c.Equal(0, len(traitModifierNames("Beta")), "Beta was not chosen")
	c.Equal(0, len(traitModifierNames("Gamma")), "Gamma was not chosen")
	c.Equal(0, len(traitModifierNames("Delta")), "Delta was not chosen")
	c.Equal([]string{"Ranged"}, traitModifierNames("Epsilon"), "only the selected modifier is applied")
	c.True(enabled, "the applied copy is switched on although the library's modifier is off")
	c.True(modified, "the sheet has been changed")
	c.NotEqual(jio.Time{}, modifiedOn, "the change has been reported to the sheet")
	c.Equal(sheetState{selected: []string{"Alpha", "Epsilon"}, deltaOpen: true, focused: true}, currentSheetState(),
		"the container is opened and the focus lands on the sheet's traits list with the targets selected")
	c.Equal(2, selectedRows, "both targets must be selected for the check to mean anything")
	c.Equal(selectedRows, selectedInView, "both targets are in view")
	c.Equal("Undo "+applyModifierAction.Title, undoTitle)
	var templateModifiers int
	screen.Do(func() { templateModifiers = len(template.template.Traits[0].Modifiers) })
	c.Equal(0, templateModifiers, "the template, which was not chosen, is left alone")

	// Edit > Undo undoes the command in one step, closing the container it opened, and Edit > Redo puts everything back
	// as the command left it.
	uxtest.ChooseMenuBarItem(t, screen, wnd, "Edit", "Undo "+applyModifierAction.Title)
	var canUndo, canRedo bool
	screen.Do(func() {
		canUndo = sheet.undoMgr.CanUndo()
		canRedo = sheet.undoMgr.CanRedo()
	})
	c.Equal(0, len(traitModifierNames("Alpha")), "undo must take the modifier back off")
	c.Equal(0, len(traitModifierNames("Epsilon")), "undo must take the modifier back off")
	c.False(currentSheetState().deltaOpen, "undo must close the container the command opened")
	c.False(canUndo, "one command must make exactly one edit")
	c.True(canRedo, "the command must be redoable")
	uxtest.ChooseMenuBarItem(t, screen, wnd, "Edit", "Redo "+applyModifierAction.Title)
	c.Equal([]string{"Ranged"}, traitModifierNames("Alpha"), "redo must put the modifier back")
	c.Equal([]string{"Ranged"}, traitModifierNames("Epsilon"), "redo must put the modifier back")
	c.Equal(sheetState{selected: []string{"Alpha", "Epsilon"}, deltaOpen: true, focused: true}, currentSheetState(),
		"redo must show the targets as the command did")
	uxtest.ChooseMenuBarItem(t, screen, wnd, "Edit", "Undo "+applyModifierAction.Title)
	var stillModified bool
	screen.Do(func() {
		stillModified = sheet.Modified()
		sheet.Entity().ModifiedOn = jio.Time{} // Undo puts the data back, not the timestamp the command's report set.
	})
	c.False(stillModified, "undoing the command must leave the sheet as it was")

	// With the template closed, the sheet is the only destination, so the command asks for its traits straight away.
	uxtest.CloseEditorWithoutPrompt(t, screen, template)
	var destinations int
	screen.Do(func() { destinations = len(modifierDestinations(traitModifierTargetKind(), AllDockables())) })
	c.Equal(1, destinations, "only the sheet must be left as a destination for the check to mean anything")
	selectModifier()
	uxtest.ChooseMenuBarItem(t, screen, wnd, "Item", applyModifierAction.Title)
	targetPrompt()
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Equal(1, windowCount(), "the prompt has been dismissed")
	nothingApplied("canceling the target prompt reached directly")
}

// TestApplyModifierHeadlessRaisesTheDestinationsWindow verifies that the command, chosen from a modifier library in a
// window of its own, ends with the destination's window in front and the focus on the list that received the modifier,
// so that the next Undo goes to the sheet rather than the library.
func TestApplyModifierHeadlessRaisesTheDestinationsWindow(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	forbidModifierPrompts(t)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	screen.Do(func() {
		entity := sheet.Entity()
		trait := gurps.NewTrait(entity, nil, false)
		trait.Name = "Alpha"
		entity.Traits = []*gurps.Trait{trait}
		entity.Recalculate()
		sheet.Rebuild(true)
		sheet.markUnmodified()
	})
	ranged := gurps.NewTraitModifier(nil, nil, false)
	ranged.Name = "Ranged"
	var library *TableDockable[*gurps.TraitModifier]
	var libraryWnd *unison.Window
	screen.Do(func() {
		library = NewTraitModifierTableDockable(filepath.Join(t.TempDir(), "mods"+gurps.TraitModifiersExt),
			[]*gurps.TraitModifier{ranged})
		var err error
		if libraryWnd, err = NewWindowForDockable(library, dgroup.Libraries); err != nil {
			t.Errorf("unable to open the library in a window of its own: %v", err)
			return
		}
		// A window packed around one modifier is too small to hold its Item menu's popup, which would then open
		// past its edge, over the workspace window, where a click on an item lands in the workspace instead.
		libraryWnd.SetFrameRect(geom.NewRect(200, 100, 900, 700))
		libraryWnd.ToFront()
	})
	if libraryWnd == nil {
		t.Fatal("the modifier library could not be opened in a window of its own")
	}
	alphaModifiers := func() []string {
		var names []string
		screen.Do(func() { names = appliedModifierNames(sheet.Entity().Traits[0].Modifiers) })
		return names
	}

	screen.Click(cellScreenPoint(t, screen, libraryWnd, library.table, 0, gurps.TraitModifierDescriptionColumn))
	c.Equal(libraryWnd, screen.FocusedWindow(), "the library's window must hold the focus for the check to mean anything")
	var enabled bool
	screen.Do(func() { enabled = library.table.HasSelection() && applyModifierAction.Enabled(nil) })
	c.True(enabled, "with a modifier selected and the sheet open, the command must be available")

	// The sheet being the only destination, the command asks for its traits straight away.
	uxtest.ChooseMenuBarItem(t, screen, libraryWnd, "Item", applyModifierAction.Title)
	dialogWnd, dialog := modalDialog(t, screen, wnd)
	var picked bool
	screen.Do(func() {
		if lists := uxtest.PanelsOfType[*unison.List[modifierTargetChoice[*gurps.Trait]]](dialogWnd.Content()); len(lists) == 1 {
			lists[0].Select(false, 0)
			picked = true
		}
	})
	if !picked {
		t.Fatal("the target prompt must hold one list of the sheet's traits")
	}
	screen.Click(screen.PanelCenter(dialogButton(t, screen, dialog, unison.ModalResponseOK)))
	c.Equal([]string{"Ranged"}, alphaModifiers(), "the modifier must have been applied")
	c.Equal(wnd, screen.FocusedWindow(), "the sheet's window must have been brought to the front")
	var focused bool
	screen.Do(func() { focused = wnd.Focus() == sheet.Traits.Table.AsPanel() })
	c.True(focused, "the focus must land on the sheet's traits list")

	// The focus being on the sheet, its window's Edit > Undo undoes the command.
	uxtest.ChooseMenuBarItem(t, screen, wnd, "Edit", "Undo "+applyModifierAction.Title)
	c.Equal(0, len(alphaModifiers()), "undo must take the modifier back off")
	screen.Do(func() { libraryWnd.Dispose() })
}

// TestRevealModifierTargetsFocusesTheFirstListWithASelection verifies that with targets in both of a sheet's equipment
// lists, the focus lands on the first list, with its target in view.
func TestRevealModifierTargetsFocusesTheFirstListWithASelection(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	forbidModifierPrompts(t)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	var rope, coin *gurps.Equipment
	screen.Do(func() {
		entity := sheet.Entity()
		rope = gurps.NewEquipment(entity, nil, false)
		rope.Name = "Rope"
		entity.CarriedEquipment = []*gurps.Equipment{rope}
		coin = gurps.NewEquipment(entity, nil, false)
		coin.Name = "Coin"
		entity.OtherEquipment = []*gurps.Equipment{coin}
		entity.Recalculate()
		sheet.Rebuild(true)
		sheet.markUnmodified()
	})
	fine := gurps.NewEquipmentModifier(nil, nil, false)
	fine.Name = "Fine"
	var applied bool
	screen.Do(func() {
		tables := []*unison.Table[*Node[*gurps.Equipment]]{sheet.CarriedEquipment.Table, sheet.OtherEquipment.Table}
		applied = applyModifiersTo(tables, []*gurps.Equipment{rope, coin}, []*gurps.EquipmentModifier{fine},
			gurps.LibraryFile{})
		revealModifierTargets(sheet, tables)
	})
	c.True(applied, "the modifier must have been applied")
	selectedNames := func(table *unison.Table[*Node[*gurps.Equipment]]) []string {
		rows := table.SelectedRows(false)
		names := make([]string, 0, len(rows))
		for _, row := range rows {
			names = append(names, row.Data().Name)
		}
		return names
	}
	var carriedSelected, otherSelected []string
	var carriedFocused, otherFocused, inView bool
	screen.Do(func() {
		carried, other := sheet.CarriedEquipment.Table, sheet.OtherEquipment.Table
		carriedSelected = selectedNames(carried)
		otherSelected = selectedNames(other)
		carriedFocused = wnd.Focus() == carried.AsPanel()
		otherFocused = wnd.Focus() == other.AsPanel()
		if row := carried.FirstSelectedRowIndex(); row != -1 {
			inView = rowInView(carried, row)
		}
	})
	c.Equal([]string{"Rope"}, carriedSelected, "the carried target must be selected")
	c.Equal([]string{"Coin"}, otherSelected, "the other target must be selected")
	c.True(carriedFocused, "the focus must land on the first list holding a target")
	c.False(otherFocused, "the focus must not move on to the second list")
	c.True(inView, "the first list's target must be in view")
	screen.Do(func() { sheet.undoMgr.Undo() })
}

// TestRevealModifierTargetsScrollsToATargetBelowTheOldEndOfTheList verifies that a target far down inside a closed
// container is scrolled into view, which needs the layout validated after opening the container, or the scroll would
// be clamped to the old end of the list.
func TestRevealModifierTargetsScrollsToATargetBelowTheOldEndOfTheList(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	forbidModifierPrompts(t)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	var last *gurps.Trait
	var lastRowBefore int
	screen.Do(func() {
		entity := sheet.Entity()
		entity.Traits = nil
		for i := range 10 {
			trait := gurps.NewTrait(entity, nil, false)
			trait.Name = fmt.Sprintf("Trait %d", i)
			entity.Traits = append(entity.Traits, trait)
		}
		bag := gurps.NewTrait(entity, nil, true)
		bag.Name = "Bag"
		bag.SetOpen(false)
		for i := range 60 {
			last = gurps.NewTrait(entity, bag, false)
			last.Name = fmt.Sprintf("Item %d", i)
			bag.Children = append(bag.Children, last)
		}
		entity.Traits = append(entity.Traits, bag)
		entity.Recalculate()
		sheet.Rebuild(true)
		sheet.markUnmodified()
		lastRowBefore = sheet.Traits.Table.LastRowIndex()
	})
	c.Equal(10, lastRowBefore, "the list must start out showing the ten traits and the closed container")
	ranged := gurps.NewTraitModifier(nil, nil, false)
	ranged.Name = "Ranged"
	var applied bool
	screen.Do(func() {
		tables := []*unison.Table[*Node[*gurps.Trait]]{sheet.Traits.Table}
		applied = applyModifiersTo(tables, []*gurps.Trait{last}, []*gurps.TraitModifier{ranged}, gurps.LibraryFile{})
		revealModifierTargets(sheet, tables)
	})
	c.True(applied, "the modifier must have been applied")
	// The state is read in a second round, once the scroll the reveal asked for has been laid out.
	var row int
	var inView bool
	screen.Do(func() {
		live := sheet.Traits.Table
		row = live.FirstSelectedRowIndex()
		if row != -1 {
			inView = rowInView(live, row)
		}
	})
	c.Equal(70, row, "the last item in the container must be the row selected")
	c.True(inView, "the selected target must have been scrolled into view")
	screen.Do(func() { sheet.undoMgr.Undo() })
}
