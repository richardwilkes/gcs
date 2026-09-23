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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

// newTemplateChoiceTrait returns a trait container carrying template choices, holding a child for each of the given
// names.
func newTemplateChoiceTrait(name string, childNames ...string) *gurps.Trait {
	container := gurps.NewTrait(nil, nil, true)
	container.Name = name
	container.TemplatePicker.Type = picker.Count
	container.TemplatePicker.Qualifier.Compare = criteria.EqualsNumber
	container.TemplatePicker.Qualifier.Qualifier = fxp.One
	children := make([]*gurps.Trait, 0, len(childNames))
	for _, childName := range childNames {
		child := gurps.NewTrait(nil, container, false)
		child.Name = childName
		children = append(children, child)
	}
	container.Children = children
	return container
}

// stubTemplatePickerStripDialog substitutes a non-interactive confirmation that answers with the given response. The
// count of confirmations presented is returned.
func stubTemplatePickerStripDialog(t *testing.T, response bool) *int {
	t.Helper()
	asked := 0
	swapForTest(t, &templatePickerStripDialog, func() bool {
		asked++
		return response
	})
	return &asked
}

// stubTemplatePickerChoices substitutes a non-interactive template choice dialog that keeps the children at the given
// indexes, standing in for the user checking exactly those boxes. The count of dialogs presented is returned.
func stubTemplatePickerChoices(t *testing.T, keep ...int) *int {
	t.Helper()
	shown := 0
	swapForTest(t, &templatePickerChoiceHook, func(_ any, _ int) ([]int, bool) {
		shown++
		return keep, false
	})
	return &shown
}

// TestDecideTemplatePickerCopyDestinations verifies what each kind of destination does about the template choices a
// copy carries: a template keeps them, a character sheet makes them, and anywhere else gives them up, but only once
// that has been confirmed.
func TestDecideTemplatePickerCopyDestinations(t *testing.T) {
	c := check.New(t)
	asked := stubTemplatePickerStripDialog(t, true)

	template := newTestTemplateDockable("Destination", gurps.NewTemplate())
	c.Equal(templatePickerCopyKeep, decideTemplatePickerCopy(template), "a template accepts template choices")
	c.Equal(0, *asked, "a template must not be asked to give up the choices")

	sheet := newTestSheetForTemplate(t)
	c.Equal(templatePickerCopyResolve, decideTemplatePickerCopy(sheet), "a sheet makes the choices on the spot")
	c.Equal(0, *asked, "a partial template application must present no confirmation of its own")

	list := newTestTraitTableDockable()
	c.Equal(templatePickerCopyStrip, decideTemplatePickerCopy(list), "a library list can only take the rows without")
	c.Equal(1, *asked, "giving up the choices must be confirmed first")

	asked = stubTemplatePickerStripDialog(t, false)
	c.Equal(templatePickerCopyRefuse, decideTemplatePickerCopy(list), "declining must abandon the copy")
	c.Equal(1, *asked)
}

// TestGuardTemplatePickerDropOntoLibraryList verifies that dragging a container carrying template choices into a
// library list asks before anything is inserted, and is turned away when the answer is no, while an ordinary drag is
// left alone.
func TestGuardTemplatePickerDropOntoLibraryList(t *testing.T) {
	c := check.New(t)
	asked := stubTemplatePickerStripDialog(t, true)
	dockable := newTestTraitTableDockable()
	di := &fakeDragInfo{types: []string{traitDragKey.UTI}}
	defer func() { pendingTemplatePickerAction = templatePickerCopyKeep }()

	plain := gurps.NewTrait(nil, nil, false)
	plain.Name = "Claws"
	swapForTest(t, &draggedTableData, any(newDragData(plain)))
	c.True(guardTemplatePickerDrop(dockable.table, dockable.provider, di), "a drag without choices must be let through")
	c.Equal(0, *asked, "a drag without choices must present no dialog")
	c.Equal(templatePickerCopyKeep, pendingTemplatePickerAction)

	swapForTest(t, &draggedTableData, any(newDragData(newTemplateChoiceTrait("Pick One", "First", "Second"))))
	c.True(guardTemplatePickerDrop(dockable.table, dockable.provider, di), "accepting must let the drop through")
	c.Equal(1, *asked, "giving up the choices must be confirmed first")
	c.Equal(templatePickerCopyStrip, pendingTemplatePickerAction, "the drop must be told to strip the choices")

	c.True(guardTemplatePickerDrop(dockable.table, dockable.provider,
		&fakeDragInfo{types: []string{noteDragKey.UTI}}), "a drag of some other type is not ours to question")
	c.Equal(1, *asked, "a drag of some other type must present no dialog")

	asked = stubTemplatePickerStripDialog(t, false)
	c.False(guardTemplatePickerDrop(dockable.table, dockable.provider, di), "declining must refuse the drop")
	c.Equal(1, *asked)
	c.Equal(templatePickerCopyRefuse, pendingTemplatePickerAction)
}

// sheetDropFixture prepares a character sheet and the pieces needed to drive a drag of trait rows onto it.
func sheetDropFixture(t *testing.T, c check.Checker, dragged *gurps.Trait) (*Sheet, *unison.Table[*Node[*gurps.Trait]], TableProvider[*gurps.Trait], *fakeDragInfo) {
	t.Helper()
	sheet := newTestSheetForTemplate(t)
	table := sheet.Traits.Table
	provider, ok := table.ClientData()[TableProviderClientKey].(TableProvider[*gurps.Trait])
	c.True(ok, "the sheet's traits table must know its provider")
	swapForTest(t, &draggedTableData, any(newDragData(dragged)))
	t.Cleanup(func() { pendingTemplatePickerAction = templatePickerCopyKeep })
	return sheet, table, provider, &fakeDragInfo{types: []string{traitDragKey.UTI}}
}

// TestGuardTemplatePickerDropOntoSheetAppliesPartialTemplate verifies that dragging a container carrying template
// choices onto a character sheet is performed as a partial template application by the guard itself, leaving only what
// was picked on the sheet, and that the drop unison would otherwise have made is turned away.
func TestGuardTemplatePickerDropOntoSheetAppliesPartialTemplate(t *testing.T) {
	c := check.New(t)
	shown := stubTemplatePickerChoices(t, 1)
	asked := stubTemplatePickerStripDialog(t, false)
	sheet, table, provider, di := sheetDropFixture(t, c, newTemplateChoiceTrait("Pick One", "First", "Second"))
	entity := sheet.Entity()
	originalTraits := len(entity.Traits) // A new entity may come with traits of its own, such as the natural attacks.

	c.False(guardTemplatePickerDrop(table, provider, di),
		"the drop must be turned away, since the partial template application has already added the rows")
	c.Equal(1, *shown, "the choices must have been presented")
	c.Equal(0, *asked, "the partial application path must present no dialog of its own")
	c.Equal(templatePickerCopyKeep, pendingTemplatePickerAction,
		"nothing may be left for the did-drop callback to act on")
	c.Equal(originalTraits+1, len(entity.Traits), "only the picked trait must have been added")
	added := entity.Traits[len(entity.Traits)-1]
	c.Equal("Second", added.Name)
	c.False(added.Container(), "the container carrying the choices must not have come along")
	c.True(added.TemplatePicker.IsZero(), "no template choices may have landed on the sheet")
}

// TestGuardTemplatePickerDropOntoSheetCanceledLeavesSheetUntouched verifies that canceling the choices of a drag onto a
// character sheet adds nothing and records nothing to undo.
func TestGuardTemplatePickerDropOntoSheetCanceledLeavesSheetUntouched(t *testing.T) {
	c := check.New(t)
	swapForTest(t, &templatePickerChoiceHook, func(_ any, _ int) ([]int, bool) { return nil, true })
	sheet, table, provider, di := sheetDropFixture(t, c, newTemplateChoiceTrait("Pick One", "First", "Second"))
	entity := sheet.Entity()
	originalTraits := len(entity.Traits)
	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	mgr.Clear()

	c.False(guardTemplatePickerDrop(table, provider, di), "a canceled choice must turn the drop away")
	c.Equal(originalTraits, len(entity.Traits), "nothing must have been added")
	c.False(mgr.CanUndo(), "an application that left no trace must record no undo edit")
}

// TestCopySelectionToSheetResolvesTemplateChoices verifies that copying a container carrying template choices onto a
// character sheet lands only what was picked, the way applying a template does.
func TestCopySelectionToSheetResolvesTemplateChoices(t *testing.T) {
	c := check.New(t)
	shown := stubTemplatePickerChoices(t, 0)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	originalTraits := len(entity.Traits) // A new entity may come with traits of its own, such as the natural attacks.
	source := newLibraryStyleTraitsTable(newTemplateChoiceTrait("Pick One", "First", "Second"))
	source.SelectAll()

	copySelectionTo(source, []*Sheet{sheet})
	c.Equal(1, *shown, "the choices must have been presented")
	c.Equal(originalTraits+1, len(entity.Traits), "only the picked trait must have been added")
	added := entity.Traits[len(entity.Traits)-1]
	c.Equal("First", added.Name)
	c.False(added.Container(), "the container carrying the choices must not have come along")
}

// TestCopySelectionToTemplateKeepsTemplateChoices verifies that a copy onto a template still carries the choices over
// untouched, since a template is the one place they can be managed.
func TestCopySelectionToTemplateKeepsTemplateChoices(t *testing.T) {
	c := check.New(t)
	shown := stubTemplatePickerChoices(t)
	asked := stubTemplatePickerStripDialog(t, false)
	templateData := gurps.NewTemplate()
	template := newTestTemplateDockable("Destination", templateData)
	source := newLibraryStyleTraitsTable(newTemplateChoiceTrait("Pick One", "First", "Second"))
	source.SelectAll()

	copySelectionTo(source, []*Template{template})
	c.Equal(0, *shown, "a template must not be asked to make the choices")
	c.Equal(0, *asked, "a template must not be asked to give up the choices")
	c.Equal(1, len(templateData.TraitList()), "the container must have been copied")
	copied := templateData.TraitList()[0]
	c.Equal("Pick One", copied.Name)
	c.False(copied.TemplatePicker.IsZero(), "the choices must have come along")
	c.Equal(2, len(copied.Children), "the choices must still have everything to pick from")
}

// TestDidDropStripsTemplateChoices verifies that a drop into a library list that was allowed through without its
// template choices lands the rows with the choices gone and everything else about them intact.
func TestDidDropStripsTemplateChoices(t *testing.T) {
	c := check.New(t)
	dockable := newTestTraitTableDockable()
	table := dockable.table
	originalRows := len(dockable.provider.RootData())

	// Stand in for the drop itself, which adds the rows and selects them before the completion callback runs.
	choices := newTemplateChoiceTrait("Pick One", "First", "Second")
	choices.LocalNotes = "Keep me"
	dockable.provider.SetRootData(append(dockable.provider.RootData(), choices))
	table.SyncToModel()
	table.SetSelectionMap(map[tid.TID]bool{choices.ID(): true})
	pendingTemplatePickerAction = templatePickerCopyStrip
	defer func() { pendingTemplatePickerAction = templatePickerCopyKeep }()

	didDropCallback(willDropCallback(nil, table, false), nil, table, false)
	c.Equal(templatePickerCopyKeep, pendingTemplatePickerAction, "the request to strip must have been consumed")
	c.Equal(originalRows+1, len(dockable.provider.RootData()), "the container must have been kept")
	landed := dockable.provider.RootData()[originalRows]
	c.Equal("Pick One", landed.Name)
	c.True(landed.TemplatePicker.IsZero(), "the choices must have been removed")
	c.Equal("Keep me", landed.LocalNotes, "everything else must be untouched")
	c.Equal(2, len(landed.Children), "the children must have come along")
}
