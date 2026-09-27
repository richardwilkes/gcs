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
	"maps"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
)

// The Apply Modifier command's prompts are held in variables so that tests can substitute non-interactive
// implementations.
var (
	promptForModifierDestination = promptForSingleDestination[FileBackedDockable]
	showModifierTargetsPrompt    = showListQuestionDialog
)

// modifierTargetList is a destination's table of rows modifiers can be attached to, named for the target prompt when
// the destination has more than one (a sheet's carried and other equipment).
type modifierTargetList[T gurps.Node[T]] struct {
	name  string
	table *unison.Table[*Node[T]]
}

// modifierTargetKind holds what differs between applying trait modifiers and equipment modifiers.
type modifierTargetKind[T gurps.ModifiableNode[T, M], M gurps.ModifierNode[M, T]] struct {
	// lists returns the dockable's lists that can currently receive the modifiers (see canReceiveModifiers). They are
	// looked up each time, since a rebuild can replace a list.
	lists  func(d unison.Dockable) []modifierTargetList[T]
	header string // The target prompt's header, with a %s for the destination's title.
}

// traitModifierTargetKind returns the kind that applies trait modifiers: to the traits of a sheet or template, or to
// a trait library list.
func traitModifierTargetKind() modifierTargetKind[*gurps.Trait, *gurps.TraitModifier] {
	return modifierTargetKind[*gurps.Trait, *gurps.TraitModifier]{
		lists: func(d unison.Dockable) []modifierTargetList[*gurps.Trait] {
			switch d := d.AsPanel().Self.(type) {
			case *Sheet:
				return usableModifierTargetLists(pageTargetList("", d.Traits))
			case *Template:
				return usableModifierTargetLists(pageTargetList("", d.Traits))
			case *TableDockable[*gurps.Trait]:
				return usableModifierTargetLists(modifierTargetList[*gurps.Trait]{table: d.table})
			default:
				return nil
			}
		},
		header: i18n.Text("Choose the traits in %s that should receive the selected modifiers:"),
	}
}

// equipmentModifierTargetKind returns the kind that applies equipment modifiers: to the carried and other equipment
// of a sheet, the equipment of a template or loot sheet, or to an equipment library list.
func equipmentModifierTargetKind() modifierTargetKind[*gurps.Equipment, *gurps.EquipmentModifier] {
	return modifierTargetKind[*gurps.Equipment, *gurps.EquipmentModifier]{
		lists: func(d unison.Dockable) []modifierTargetList[*gurps.Equipment] {
			switch d := d.AsPanel().Self.(type) {
			case *Sheet:
				return usableModifierTargetLists(
					pageTargetList(i18n.Text("Carried Equipment"), d.CarriedEquipment),
					pageTargetList(i18n.Text("Other Equipment"), d.OtherEquipment),
				)
			case *Template:
				return usableModifierTargetLists(pageTargetList("", d.Equipment))
			case *LootSheet:
				return usableModifierTargetLists(pageTargetList("", d.Equipment))
			case *TableDockable[*gurps.Equipment]:
				return usableModifierTargetLists(modifierTargetList[*gurps.Equipment]{table: d.table})
			default:
				return nil
			}
		},
		header: i18n.Text("Choose the equipment in %s that should receive the selected modifiers:"),
	}
}

// pageTargetList returns the page list's table under the given name, or no table for a list not yet built.
func pageTargetList[T gurps.Node[T]](name string, list *PageList[T]) modifierTargetList[T] {
	if list == nil {
		return modifierTargetList[T]{name: name}
	}
	return modifierTargetList[T]{name: name, table: list.Table}
}

// usableModifierTargetLists returns the lists whose tables can take modifiers right now (see canReceiveModifiers).
func usableModifierTargetLists[T gurps.Node[T]](lists ...modifierTargetList[T]) []modifierTargetList[T] {
	var result []modifierTargetList[T]
	for _, list := range lists {
		if canReceiveModifiers(list.table) {
			result = append(result, list)
		}
	}
	return result
}

// canReceiveModifiers reports whether the Apply Modifier command can attach modifiers to the table's rows: it has
// rows, it isn't filtered (the same refusal a drop makes), it carries its provider, which the undo data is collected
// from, and it sits under an undo manager, which a list the sheet's layout leaves off the page does not.
func canReceiveModifiers[T gurps.Node[T]](table *unison.Table[*Node[T]]) bool {
	if table == nil || table.IsFiltered() || len(table.RootRows()) == 0 || unison.UndoManagerFor(table) == nil {
		return false
	}
	_, ok := table.ClientData()[TableProviderClientKey].(TableProvider[T])
	return ok
}

// modifierDestinations returns the dockables among those given that hold at least one list the kind's modifiers can
// be applied to right now.
func modifierDestinations[T gurps.ModifiableNode[T, M], M gurps.ModifierNode[M, T]](kind modifierTargetKind[T, M], dockables []unison.Dockable) []FileBackedDockable {
	var result []FileBackedDockable
	for _, d := range dockables {
		if fbd, ok := d.AsPanel().Self.(FileBackedDockable); ok && len(kind.lists(d)) != 0 {
			result = append(result, fbd)
		}
	}
	return result
}

// installApplyModifierHandler installs the Apply Modifier command on target for the given table of modifiers. It is
// available while the table has a selection and some open dockable can receive the modifiers. The handler goes on the
// dockable rather than the table so that it stays available while the focus is in the filter field, as the standard
// table commands do.
func installApplyModifierHandler[T gurps.ModifiableNode[T, M], M gurps.ModifierNode[M, T]](target unison.Paneler, table *unison.Table[*Node[M]], kind modifierTargetKind[T, M]) {
	target.AsPanel().InstallCmdHandlers(ApplyModifierItemID,
		func(_ any) bool { return table.HasSelection() && len(modifierDestinations(kind, AllDockables())) != 0 },
		func(_ any) { applySelectedModifiers(table, kind) })
}

// applySelectedModifiers applies the modifiers selected in source to the rows the user picks, prompting first for the
// open dockable and then for the rows within it. It is the keyboard counterpart of dropping the modifiers onto rows.
func applySelectedModifiers[T gurps.ModifiableNode[T, M], M gurps.ModifierNode[M, T]](source *unison.Table[*Node[M]], kind modifierTargetKind[T, M]) {
	rows := source.SelectedRows(true)
	if len(rows) == 0 {
		return
	}
	modifiers := make([]M, 0, len(rows))
	for _, row := range rows {
		modifiers = append(modifiers, row.Data())
	}
	from := libraryFileFromTable(source)
	dest, ok := promptForModifierDestination(modifierDestinations(kind, AllDockables()))
	if !ok {
		return
	}
	picked := showModifierTargetsDialog(fmt.Sprintf(kind.header, dest.Title()), modifierTargetChoices(kind.lists(dest)))
	if len(picked) == 0 {
		return
	}
	// The choices are in prompt order, one list after another, so each table's targets are contiguous.
	var tables []*unison.Table[*Node[T]]
	targets := make([]T, 0, len(picked))
	for _, choice := range picked {
		if len(tables) == 0 || tables[len(tables)-1] != choice.table {
			tables = append(tables, choice.table)
		}
		targets = append(targets, choice.target)
	}
	if applyModifiersTo(tables, targets, modifiers, from) {
		revealModifierTargets(dest, tables)
	}
}

// applyModifiersTo attaches clones of the modifiers to the targets, which live in the given tables of one dockable, as
// a single undoable edit (see attachModifierClones), and selects the targets, opening the containers holding them
// first. The containers' open state is not part of the undo data (see gurps.SetNodeOpen), so the ones the command
// opened are closed again before an undo and reopened before a redo, as the move commands do for theirs. Returns false
// if the user canceled a prompt, in which case nothing has been changed.
func applyModifiersTo[T gurps.ModifiableNode[T, M], M gurps.ModifierNode[M, T]](tables []*unison.Table[*Node[T]], targets []T, modifiers []M, from gurps.LibraryFile) bool {
	if len(tables) == 0 || len(targets) == 0 || len(modifiers) == 0 {
		return false
	}
	// Taken before the change, since attaching the modifiers can rebuild the owner and replace the tables, and an
	// orphan can't find its manager.
	mgr := unison.UndoManagerFor(tables[0])
	var before *tablesUndoData
	if mgr != nil {
		before = newTablesUndoDataForTables(tables)
	}
	var dataOwner gurps.DataOwner
	if provider, ok := tables[0].ClientData()[TableProviderClientKey].(gurps.DataOwnerProvider); ok &&
		!xreflect.IsNil(provider) {
		dataOwner = provider.DataOwner()
	}
	if !attachModifierClones(tables, dataOwner, targets, modifiers, from) {
		return false
	}
	opened := selectModifierTargets(tables, targets)
	live := liveTable(tables[0])
	if dropRebuilder(live) == nil {
		// Without an entity there was no rebuild to report the change.
		MarkModified(live)
	}
	if mgr != nil {
		mgr.Add(&unison.UndoEdit[*tablesUndoData]{
			ID:       unison.NextUndoID(),
			EditName: applyModifierAction.Title,
			UndoFunc: func(e *unison.UndoEdit[*tablesUndoData]) {
				setContainersOpen(opened, false)
				e.BeforeData.Apply()
			},
			RedoFunc: func(e *unison.UndoEdit[*tablesUndoData]) {
				setContainersOpen(opened, true)
				e.AfterData.Apply()
			},
			AbsorbFunc: func(_ *unison.UndoEdit[*tablesUndoData], _ unison.Undoable) bool { return false },
			BeforeData: before,
			AfterData:  newTablesUndoDataForTables(tables), // NewTableUndoEditData looks up the live tables itself.
		})
	}
	return true
}

// selectModifierTargets selects the targets in the tables, opening the containers holding them first, since a table
// can only select rows it is showing, and returns the containers it opened.
func selectModifierTargets[T gurps.Node[T]](tables []*unison.Table[*Node[T]], targets []T) []T {
	selection := make(map[tid.TID]bool, len(targets))
	var opened []T
	for _, target := range targets {
		selection[target.ID()] = true
		for parent := target.Parent(); !xreflect.IsNil(parent); parent = parent.Parent() {
			if !parent.IsOpen() {
				parent.SetOpen(true)
				opened = append(opened, parent)
			}
		}
	}
	for _, table := range tables {
		live := liveTable(table)
		if len(opened) != 0 {
			live.SyncToModel()
		}
		live.SetSelectionMap(selection)
	}
	return opened
}

// attachModifierClones gives each target its own clones of the modifiers, appended to its modifiers. Both dropping
// modifiers onto rows and the Apply Modifier command end here. A modifier pointed at directly is enabled, since the
// user has just said to apply it. A container's contents are a set to choose from, so when the rows belong to an
// entity each target receiving a container is asked which of its new modifiers should be enabled (see
// ProcessModifiers); elsewhere they are left as they came and the choice is made when the row reaches a sheet (see
// applyTransfer). For an entity the nameables prompt follows, before anything is shown or reported, so that a cancel
// only has to take the clones back off and the owner is rebuilt once, with the answers in place; that rebuild is also
// what reports the change (see dropRebuilder). Elsewhere reporting is left to the caller. The tables are the ones the
// targets live in and all belong to one owner. Returns false if nothing was changed: a prompt was canceled, in which
// case every target has been put back as it was, or there was nothing to do.
func attachModifierClones[T gurps.ModifiableNode[T, M], M gurps.ModifierNode[M, T]](tables []*unison.Table[*Node[T]], dataOwner gurps.DataOwner, targets []T, modifiers []M, from gurps.LibraryFile) bool {
	if len(tables) == 0 || len(targets) == 0 || len(modifiers) == 0 {
		return false
	}
	forEntity := !xreflect.IsNil(dataOwner) && dataOwner.OwningEntity() != nil
	askAboutContainers := false
	for _, m := range modifiers {
		if m.Container() {
			askAboutContainers = forEntity
			break
		}
	}
	// The clones are grouped by target for the nameables prompt, which would otherwise show the copies of one modifier
	// as identically titled sections. What each target had before is kept so that a canceled prompt can put it back:
	// its modifier list, which is only appended to, and its replacements, which attaching a modifier saved by an older
	// version adds to.
	groups := make([]NameableGroup[M], 0, len(targets))
	originalModifiers := make([][]M, 0, len(targets))
	originalReplacements := make([]map[string]string, 0, len(targets))
	restore := func() {
		for i := range originalModifiers {
			targets[i].SetModifiers(originalModifiers[i])
			if setter, ok := any(targets[i]).(nameable.Setter); ok {
				setter.SetNameableReplacements(originalReplacements[i])
			}
		}
	}
	for _, target := range targets {
		originalModifiers = append(originalModifiers, target.ModifierList())
		originalReplacements = append(originalReplacements, maps.Clone(target.NameableReplacements()))
		clones := make([]M, 0, len(modifiers))
		for _, m := range modifiers {
			var noParent M
			clone := m.Clone(from, dataOwner, noParent, gurps.Reference)
			// A container has no switch of its own; its contents are asked about below.
			if !clone.Container() {
				clone.SetEnabled(true)
			}
			clones = append(clones, clone)
		}
		target.AddModifiers(clones...)
		if askAboutContainers && promptForClonedModifiers(xstrings.Truncate(target.String(), 40, true), clones) {
			restore()
			return false
		}
		groups = append(groups, NameableGroup[M]{Label: target.String(), Rows: clones, SharedReplacements: true})
	}
	if forEntity && !ProcessNameableGroups(groups) {
		restore()
		return false
	}
	// Looked up afresh in case a table was replaced while a prompt was up.
	for _, table := range tables {
		liveTable(table).SyncToModel()
	}
	if forEntity {
		rebuildAsModified(dropRebuilder(liveTable(tables[0])), true)
	}
	return true
}

// promptForClonedModifiers puts up the prompt asking which of the clones should be enabled and reports whether it was
// canceled. Clones of a kind with no prompt are not asked about.
func promptForClonedModifiers[M gurps.Node[M]](title string, clones []M) (canceled bool) {
	switch mods := any(clones).(type) {
	case []*gurps.TraitModifier:
		_, canceled = promptForTraitModifiers(title, mods)
	case []*gurps.EquipmentModifier:
		_, canceled = promptForEquipmentModifiers(title, mods)
	}
	return canceled
}

// modifierTargetChoice is one row of the target prompt: the target, the table it sits in, the row's text and the
// target's nesting depth, which sets the row's indentation.
type modifierTargetChoice[T gurps.Node[T]] struct {
	target T
	table  *unison.Table[*Node[T]]
	label  string
	depth  int
}

func (c modifierTargetChoice[T]) String() string {
	return c.label
}

// modifierTargetChoices returns a row for every target in the lists, containers included, in table order. Each is
// labeled with the target's name and where it sits -- its containers, nearest first, and the list's name when there
// is more than one list -- so that a screen reader is told what the indentation shows: "Knife (in Pouch, Backpack,
// Other Equipment)".
func modifierTargetChoices[T gurps.Node[T]](lists []modifierTargetList[T]) []modifierTargetChoice[T] {
	var choices []modifierTargetChoice[T]
	named := len(lists) > 1
	for _, list := range lists {
		rows := list.table.RootRows()
		roots := make([]T, 0, len(rows))
		for _, row := range rows {
			roots = append(roots, row.Data())
		}
		gurps.Traverse(func(node T) bool {
			var where []string
			for parent := node.Parent(); !xreflect.IsNil(parent); parent = parent.Parent() {
				where = append(where, parent.String())
			}
			depth := len(where)
			if named {
				where = append(where, list.name)
			}
			label := node.String()
			if len(where) != 0 {
				label = fmt.Sprintf(i18n.Text("%s (in %s)"), label, strings.Join(where, ", "))
			}
			choices = append(choices, modifierTargetChoice[T]{target: node, table: list.table, label: label, depth: depth})
			return false
		}, false, false, roots...)
	}
	return choices
}

// showModifierTargetsDialog puts up the target prompt of the Apply Modifier command and returns the chosen rows, in
// prompt order, or nil if there was nothing to offer, the dialog was canceled or nothing was chosen.
func showModifierTargetsDialog[T gurps.Node[T]](header string, choices []modifierTargetChoice[T]) []modifierTargetChoice[T] {
	if len(choices) == 0 {
		return nil
	}
	list := newChoiceList[modifierTargetChoice[T]](true)
	list.Factory = &modifierTargetCellFactory[T]{}
	list.Append(choices...)
	if !showModifierTargetsPrompt(header, list) {
		return nil
	}
	return pickFromList(list, choices)
}

// modifierTargetCellFactory draws the target prompt's rows as the default factory does, indented by nesting depth.
type modifierTargetCellFactory[T gurps.Node[T]] struct {
	unison.DefaultCellFactory
}

// CreateCell implements unison.CellFactory.
func (f *modifierTargetCellFactory[T]) CreateCell(owner unison.Paneler, element any, row int, foreground, background unison.Ink, selected, focused bool) unison.Paneler {
	cell := f.DefaultCellFactory.CreateCell(owner, element, row, foreground, background, selected, focused)
	if choice, ok := element.(modifierTargetChoice[T]); ok && choice.depth > 0 {
		if border, ok2 := cell.AsPanel().Border().(*unison.EmptyBorder); ok2 {
			insets := border.Insets()
			insets.Left += float32(choice.depth) * unison.FieldFont.LineHeight()
			cell.AsPanel().SetBorder(unison.NewEmptyBorder(insets))
		}
	}
	return cell
}

// revealModifierTargets brings the destination to the front and puts the focus on the first of its tables with a
// selection, with the first selected row in view. The window is raised too, since ActivateDockable only brings a
// docked destination to the front of its dock, and the library the command was chosen from may be in a window of its
// own, which would otherwise keep the focus and with it the next Undo.
func revealModifierTargets[T gurps.Node[T]](dest unison.Dockable, tables []*unison.Table[*Node[T]]) {
	if wnd := dest.AsPanel().Window(); wnd != nil {
		ActivateDockable(dest)
		wnd.ToFront()
	}
	for _, table := range tables {
		live := liveTable(table)
		row := live.FirstSelectedRowIndex()
		if row == -1 {
			continue
		}
		live.RequestFocus()
		// Opening the targets' containers only invalidated the layout, so without this the scroll range would still
		// end where the table used to.
		live.ValidateScrollRoot()
		live.ScrollRowIntoView(row)
		return
	}
}
