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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// newChoiceTrait returns a trait container carrying template choices, holding a child for each of the given names.
func newChoiceTrait(name string, optionNames ...string) *gurps.Trait {
	choices := gurps.NewTrait(nil, nil, true)
	choices.Name = name
	choices.TemplatePicker.Type = picker.Count
	choices.TemplatePicker.Qualifier.Compare = criteria.EqualsNumber
	choices.TemplatePicker.Qualifier.Qualifier = fxp.One
	for _, optionName := range optionNames {
		option := gurps.NewTrait(nil, choices, false)
		option.Name = optionName
		choices.Children = append(choices.Children, option)
	}
	return choices
}

// newTestTemplateWithTraits returns a template dockable holding the given traits, all of them selected.
func newTestTemplateWithTraits(traits ...*gurps.Trait) *Template {
	data := gurps.NewTemplate()
	data.Traits = traits
	template := newTestTemplateDockable("Source", data)
	template.Traits.Table.SelectAll()
	return template
}

// pickFirstOption substitutes a non-interactive picker prompt that settles each top-level trait carrying template
// choices by taking its first option in its place, just as the real prompt dissolves the container into the options
// chosen. The count of times the prompt was used is returned.
func pickFirstOption(t *testing.T) *int {
	t.Helper()
	calls := 0
	swapForTest(t, &promptForPickers, func(parts *applyParts) bool {
		calls++
		var revised []*gurps.Trait
		for _, row := range parts.traits.rows {
			if row.Container() && !row.TemplatePicker.IsZero() {
				option := row.Children[0]
				option.SetParent(row.Parent())
				revised = append(revised, option)
			} else {
				revised = append(revised, row)
			}
		}
		parts.traits.rows = revised
		return true
	})
	return &calls
}

// TestCopyFromTemplateToSheetResolvesPickers verifies that rows copied from a template onto a sheet have their
// template choices settled on the way, since only a template may hold them: the choice container dissolves into the
// options chosen.
func TestCopyFromTemplateToSheetResolvesPickers(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	originalTraits := len(entity.Traits)
	template := newTestTemplateWithTraits(newChoiceTrait("Pick One", "First", "Second"))
	calls := pickFirstOption(t)

	copySelectionTo(template.Traits.Table, []*Sheet{sheet})

	c.Equal(1, *calls, "the choices must have been put to the user")
	c.Equal(originalTraits+1, len(entity.Traits), "only the chosen option must arrive")
	c.Equal("First", entity.Traits[len(entity.Traits)-1].Name, "the chosen option must take the container's place")
	c.False(gurps.HasTemplatePickerData(entity.Traits...), "no template choices may reach the sheet")
}

// TestCopyFromTemplateToTemplateKeepsPickers verifies that rows copied from one template to another arrive exactly as
// authored: nothing is put to the user, and the template choices are kept.
func TestCopyFromTemplateToTemplateKeepsPickers(t *testing.T) {
	c := check.New(t)
	choices := newChoiceTrait("Pick One", "First", "Second")
	choices.Children[0].Modifiers = []*gurps.TraitModifier{newSwitchableTraitModifier("Retractable")}
	source := newTestTemplateWithTraits(choices)
	destinationData := gurps.NewTemplate()
	destination := newTestTemplateDockable("Destination", destinationData)
	calls := pickFirstOption(t)
	prompts := captureModifierPrompts(t)

	copySelectionTo(source.Traits.Table, []*Template{destination})

	c.Equal(0, *calls, "a template may hold choices, so none must be put to the user")
	c.Equal(0, len(*prompts), "a copy between templates must not prompt for modifiers")
	c.Equal(1, len(destinationData.Traits))
	c.True(gurps.HasTemplatePickerData(destinationData.Traits...), "the choices must have been kept")
}

// TestCopyFromSheetToSheetIsAPlainCopy verifies that rows copied from one sheet to another aren't applied again: their
// modifiers were settled when they arrived on the first sheet.
func TestCopyFromSheetToSheetIsAPlainCopy(t *testing.T) {
	c := check.New(t)
	source := newTestSheetForTemplate(t)
	trait := gurps.NewTrait(source.Entity(), nil, false)
	trait.Name = "Claws"
	trait.Modifiers = []*gurps.TraitModifier{newSwitchableTraitModifier("Retractable")}
	source.Entity().Traits = []*gurps.Trait{trait}
	source.Rebuild(true)
	source.Traits.Table.SelectAll()
	destination := newTestSheetForTemplate(t)
	originalTraits := len(destination.Entity().Traits)
	prompts := captureModifierPrompts(t)

	copySelectionTo(source.Traits.Table, []*Sheet{destination})

	c.Equal(0, len(*prompts), "a copy between sheets must not prompt for modifiers")
	c.Equal(originalTraits+1, len(destination.Entity().Traits), "the trait must have been copied")
}

// newAncestryTrait returns an enabled ancestry container for the named built-in ancestry.
func newAncestryTrait(name string) *gurps.Trait {
	trait := gurps.NewTrait(nil, nil, true)
	trait.Name = name
	trait.ContainerType = container.Ancestry
	trait.Ancestry = name
	return trait
}

// TestAncestryQuestionOnlyForChosenAncestry verifies that the question of disabling the character's existing ancestry,
// and that of randomizing again, are only put when an ancestry actually arrives: one that is an option of a template
// choice and isn't chosen doesn't bring either.
func TestAncestryQuestionOnlyForChosenAncestry(t *testing.T) {
	c := check.New(t)
	for _, chooseAncestry := range []bool{false, true} {
		sheet := newTestSheetForTemplate(t)
		entity := sheet.Entity()
		existing := newAncestryTrait("Human")
		existing.SetDataOwner(entity)
		entity.Traits = append(entity.Traits, existing)
		sheet.Rebuild(true)
		choices := newChoiceTrait("Pick One", "Plain Option")
		ancestry := newAncestryTrait("Human")
		ancestry.SetParent(choices)
		if chooseAncestry {
			choices.Children = append([]*gurps.Trait{ancestry}, choices.Children...)
		} else {
			choices.Children = append(choices.Children, ancestry)
		}
		template := newTestTemplateWithTraits(choices)
		pickFirstOption(t)
		asked := 0
		swapForTest(t, &askToDisableExistingAncestry, func(_, _ string) bool {
			asked++
			return false
		})
		randomizeAsked := 0
		swapForTest(t, &askToRandomizeAgain, func() bool {
			randomizeAsked++
			return false
		})

		copySelectionTo(template.Traits.Table, []*Sheet{sheet})

		if chooseAncestry {
			c.Equal(1, asked, "a chosen ancestry must bring the question")
		} else {
			c.Equal(0, asked, "an ancestry that wasn't chosen must not bring the question")
			c.Equal(0, randomizeAsked, "an ancestry that wasn't chosen must not bring randomization")
		}
	}
}

// TestAncestryQuestionNamesTheContainers verifies that the question of disabling the character's existing ancestry
// names the ancestry containers, not the ancestry they link to, since many containers may link to the same one, and
// that the name of the arriving container is the one its nameables settled on.
func TestAncestryQuestionNamesTheContainers(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	existing := newAncestryTrait("Human")
	existing.Name = "Northern Human"
	existing.SetDataOwner(entity)
	entity.Traits = append(entity.Traits, existing)
	sheet.Rebuild(true)
	arriving := newAncestryTrait("Human")
	arriving.Name = "Desert Human (@Tribe@)"
	template := newTestTemplateWithTraits(arriving)
	swapForTest(t, &promptForNameables, func(_ []string, nameables []map[string]string, _ [][]string) bool {
		for _, m := range nameables {
			for k := range m {
				m[k] = "Sand"
			}
		}
		return true
	})
	var incomingName, existingName string
	swapForTest(t, &askToDisableExistingAncestry, func(incoming, existing string) bool {
		incomingName = incoming
		existingName = existing
		return false
	})
	swapForTest(t, &askToRandomizeAgain, func() bool { return false })

	copySelectionTo(template.Traits.Table, []*Sheet{sheet})

	c.Equal("Desert Human (Sand)", incomingName, "the arriving container must be named, with its nameables settled")
	c.Equal("Northern Human", existingName, "the existing container must be named")
}

// TestCopyOfAncestryBetweenSheetsOffersRandomization verifies that an ancestry copied from one sheet to another brings
// the offer to randomize the profile again, just as one arriving from a template or library does.
func TestCopyOfAncestryBetweenSheetsOffersRandomization(t *testing.T) {
	c := check.New(t)
	source := newTestSheetForTemplate(t)
	ancestry := newAncestryTrait("Human")
	ancestry.SetDataOwner(source.Entity())
	source.Entity().Traits = []*gurps.Trait{ancestry}
	source.Rebuild(true)
	source.Traits.Table.SelectAll()
	destination := newTestSheetForTemplate(t)
	offered := 0
	swapForTest(t, &askToRandomizeAgain, func() bool {
		offered++
		return false
	})

	copySelectionTo(source.Traits.Table, []*Sheet{destination})

	c.Equal(1, offered, "copying an ancestry between sheets must offer to randomize the profile again")
}

// TestRandomizationSeesTheArrivingTraits verifies that the profile is randomized against the character as it is once
// the rows have arrived. The Human ancestry derives height from ST, and a template raising ST by 10 moves the range of
// heights it can give from 63-73 inches to 83-93, so a height drawn against the ST the character had beforehand is
// told apart from one drawn against the ST it has now.
func TestRandomizationSeesTheArrivingTraits(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	strong := gurps.NewTrait(nil, nil, false)
	strong.Name = "Increased Strength"
	bonus := gurps.NewAttributeBonus(gurps.StrengthID)
	bonus.Amount = fxp.FromInteger(10)
	strong.Features = gurps.Features{bonus}
	data := gurps.NewTemplate()
	data.Traits = []*gurps.Trait{newAncestryTrait("Human"), strong}
	template := newTestTemplateDockable("Strong", data)

	c.True(template.applyTemplateToSheet(sheet, true), "the template must be applied")

	entity := sheet.Entity()
	c.Equal(fxp.FromInteger(20), entity.ResolveAttributeCurrent(gurps.StrengthID), "the template must raise ST to 20")
	height := fxp.Int(entity.Profile.Height)
	c.True(height >= fxp.FromInteger(83) && height <= fxp.FromInteger(93),
		"the height must have been drawn for ST 20, not ST 10, got "+height.String()+" inches")
}

// TestCanceledModifierPromptLeavesSheetUntouched verifies that canceling the modifier prompt of a copy onto a sheet
// abandons the copy, leaving the sheet exactly as it was and nothing to undo.
func TestCanceledModifierPromptLeavesSheetUntouched(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	originalTraits := len(entity.Traits)
	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Claws"
	trait.Modifiers = []*gurps.TraitModifier{newSwitchableTraitModifier("Retractable")}
	source := newLibraryStyleTraitsTable(trait)
	source.SelectAll()
	swapForTest(t, &promptForTraitModifiers, func(_ string, _ []*gurps.TraitModifier) (changed, canceled bool) {
		return false, true
	})

	copySelectionTo(source, []*Sheet{sheet})

	c.Equal(originalTraits, len(entity.Traits), "nothing may have been added to the sheet")
	c.Equal(originalTraits, len(sheet.Traits.Table.RootRows()), "the traits list must show the original rows")
	c.False(sheet.Modified(), "the sheet must not be left marked as modified")
	c.False(sheet.undoMgr.CanUndo(), "nothing was changed, so there must be nothing to undo")
}

// TestCanceledNameablesPromptLeavesSheetUntouched verifies that canceling the nameables prompt, the last question put
// to the user before a template's rows and body type are applied, abandons the template, leaving the sheet exactly as
// it was and nothing to undo.
func TestCanceledNameablesPromptLeavesSheetUntouched(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	originalBodyName := entity.SheetSettings.BodyType.Name
	originalTraits := len(entity.Traits)
	template := newTestTemplateWithBodyType("Template Body")
	template.template.Traits[0].Name = "Phobia (@Subject@)"
	shown := 0
	swapForTest(t, &promptForNameables, func(_ []string, _ []map[string]string, _ [][]string) bool {
		shown++
		return false
	})

	c.False(template.applyTemplateToSheet(sheet, true), "a canceled prompt must report the template was not applied")
	c.Equal(1, shown, "the nameables must have been put to the user")
	c.Equal(originalBodyName, entity.SheetSettings.BodyType.Name, "the body type must not have been replaced")
	c.Equal(originalTraits, len(entity.Traits), "the template's trait must not have been added")
	c.False(sheet.Modified(), "the sheet must not be left marked as modified")
	c.False(sheet.undoMgr.CanUndo(), "nothing was changed, so there must be nothing to undo")
}

// dragged returns the table's rows as the data a drag of all of them hands over.
func dragged(table *unison.Table[*Node[*gurps.Trait]]) *unison.TableDragData[*Node[*gurps.Trait]] {
	return &unison.TableDragData[*Node[*gurps.Trait]]{Table: table, Rows: table.RootRows()}
}

// TestDropOnSheetIsAppliedWhereItLanded verifies that a drop onto a sheet goes through the same application as a copy,
// with the rows ending up where they were dropped -- here, inside a container, ahead of the row already there -- and
// that the whole of it can be undone in one step.
func TestDropOnSheetIsAppliedWhereItLanded(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	group := gurps.NewTrait(entity, nil, true)
	group.Name = "Group"
	existing := gurps.NewTrait(entity, group, false)
	existing.Name = "Existing"
	group.Children = []*gurps.Trait{existing}
	entity.Traits = append(entity.Traits, group)
	sheet.Rebuild(true)
	roots := sheet.Traits.Table.RootRows()
	groupRow := roots[len(roots)-1]
	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Claws"
	trait.Modifiers = []*gurps.TraitModifier{newSwitchableTraitModifier("Retractable")}
	from := newLibraryStyleTraitsTable(trait)
	var childrenWhenPrompted int
	swapForTest(t, &promptForTraitModifiers, func(_ string, modifiers []*gurps.TraitModifier) (changed, canceled bool) {
		childrenWhenPrompted = len(group.Children)
		modifiers[0].Disabled = false
		return true, false
	})

	c.True(applyDrop(dragged(from), sheet.Traits.Table, groupRow, 0), "the drop must go through")

	c.Equal(1, childrenWhenPrompted, "nothing may be added to the sheet while the user is being asked")
	c.Equal(2, len(group.Children), "the dropped row must have been placed once")
	c.Equal("Claws", group.Children[0].Name, "the dropped row must be where it was dropped")
	c.Equal(group, group.Children[0].Parent(), "the dropped row must belong to the container it was dropped into")
	c.False(group.Children[0].Modifiers[0].Disabled, "the answer to the modifier prompt must have been kept")
	c.True(sheet.undoMgr.CanUndo(), "the drop must be undoable")

	sheet.undoMgr.Undo()
	group = entity.Traits[len(entity.Traits)-1]
	c.Equal(1, len(group.Children), "undo must take the dropped row back out")
	c.Equal("Existing", group.Children[0].Name)
}

// TestDropOnSheetWithCanceledPickerLeavesSheetUntouched verifies that canceling the template choices of rows dropped
// onto a sheet refuses the drop, leaving the sheet as it was.
func TestDropOnSheetWithCanceledPickerLeavesSheetUntouched(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	originalTraits := len(entity.Traits)
	source := newTestTemplateWithTraits(newChoiceTrait("Pick One", "First", "Second"))
	calls := 0
	swapForTest(t, &promptForPickers, func(_ *applyParts) bool {
		calls++
		return false
	})

	c.False(applyDrop(dragged(source.Traits.Table), sheet.Traits.Table, nil, -1), "the drop must be refused")

	c.Equal(1, calls, "the choices must have been put to the user")
	c.Equal(originalTraits, len(entity.Traits), "nothing may have been added")
	c.False(sheet.Modified(), "the sheet must not be marked as modified")
	c.False(sheet.undoMgr.CanUndo(), "nothing was changed, so there must be nothing to undo")
}

// TestDropOnLibraryAsksBeforeRemovingPickers verifies that rows carrying template choices dropped onto a library are
// only let in once the user agrees to the choices being removed, and that the choice containers then also lose their
// source, since only a template may hold choices and a template is never a source.
func TestDropOnLibraryAsksBeforeRemovingPickers(t *testing.T) {
	c := check.New(t)
	choices := newChoiceTrait("Pick One", "First", "Second")
	choices.Source = gurps.Source{Library: "lib", Path: "x.adq", TID: choices.ID()}
	source := newTestTemplateWithTraits(choices)

	for _, accept := range []bool{false, true} {
		to := newLibraryStyleTraitsTable()
		provider, ok := to.ClientData()[TableProviderClientKey].(TableProvider[*gurps.Trait])
		c.True(ok)
		asked := 0
		swapForTest(t, &confirmTemplatePickerDataRemoval, func() bool {
			asked++
			return accept
		})

		c.Equal(accept, applyDrop(dragged(source.Traits.Table), to, nil, -1))

		c.Equal(1, asked, "the user must have been asked")
		if accept {
			c.Equal(1, len(provider.RootData()), "an accepted drop must be let in")
			dropped := provider.RootData()[0]
			c.False(gurps.HasTemplatePickerData(dropped), "the choices must have been removed")
			c.Equal(gurps.Source{}, dropped.Source, "the container that lost its choices must lose its source too")
			c.Equal(2, len(dropped.Children), "the options must be kept")
		} else {
			c.Equal(0, len(provider.RootData()), "a refused drop must leave the list alone")
		}
		c.True(gurps.HasTemplatePickerData(choices), "the template's own row must keep its choices")
	}
}

// TestEditorUndoRestoresSourceClearedByTemplatePicker verifies that undoing an edit that gave a container template
// choices, and with them cost it its source, puts the source back along with everything else, and that redoing it
// takes the source away again. An edit that leaves the source alone must leave it alone when undone and redone, too.
func TestEditorUndoRestoresSourceClearedByTemplatePicker(t *testing.T) {
	c := check.New(t)
	trait := gurps.NewTrait(nil, nil, true)
	trait.Name = "Advantages"
	source := gurps.Source{Library: "lib", Path: "x.adq", TID: trait.ID()}
	trait.Source = source
	data := gurps.NewTemplate()
	data.Traits = []*gurps.Trait{trait}
	template := newTestTemplateDockable("Source", data)
	mgr := unison.UndoManagerFor(template)
	c.NotNil(mgr, "the template must have an undo manager")

	e, _ := buildEditorContent(template, trait, initTraitEditor)
	_, tp := e.editorData.TemplatePickerData()
	tp.Type = picker.Count
	tp.Qualifier.Compare = criteria.EqualsNumber
	tp.Qualifier.Qualifier = fxp.One
	e.applyEdits()
	c.False(trait.TemplatePicker.IsZero(), "the choices must have been applied")
	c.Equal(gurps.Source{}, trait.Source, "a container given choices must lose its source")

	mgr.Undo()
	c.True(trait.TemplatePicker.IsZero(), "undo must take the choices away")
	c.Equal(source, trait.Source, "undo must give the source back")

	mgr.Redo()
	c.False(trait.TemplatePicker.IsZero(), "redo must bring the choices back")
	c.Equal(gurps.Source{}, trait.Source, "redo must take the source away again")

	mgr.Undo()
	e, _ = buildEditorContent(template, trait, initTraitEditor)
	e.editorData.Name = "Renamed"
	e.applyEdits()
	c.Equal(source, trait.Source, "an edit giving no choices must leave the source alone")
	mgr.Undo()
	c.Equal("Advantages", trait.Name)
	c.Equal(source, trait.Source, "undoing an edit that left the source alone must leave it alone")
	mgr.Redo()
	c.Equal(source, trait.Source, "redoing an edit that left the source alone must leave it alone")
}

// TestEditorApplyClearsSourceOfTemplatePicker verifies that a container given template choices loses its source, while
// one without them keeps it.
func TestEditorApplyClearsSourceOfTemplatePicker(t *testing.T) {
	c := check.New(t)
	source := gurps.Source{Library: "lib", Path: "x.adq"}

	plain := gurps.NewTrait(nil, nil, true)
	plain.Source = source
	clearSourceOfTemplatePicker(plain)
	c.Equal(source, plain.Source, "a container without choices must keep its source")

	choices := newChoiceTrait("Pick One", "First")
	choices.Source = source
	clearSourceOfTemplatePicker(choices)
	c.Equal(gurps.Source{}, choices.Source, "a container with choices must lose its source")
}
