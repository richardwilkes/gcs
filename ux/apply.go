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
	"slices"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
)

// transferKind classifies a document by what may arrive in it. Rows landing on a sheet are applied to it, settling
// every choice they carry; rows landing on a template are kept as authored; anything else is treated as a library.
type transferKind uint8

const (
	transferLibrary transferKind = iota
	transferSheet
	transferTemplate
)

// transferKindOf returns the kind of document the panel belongs to. Character sheets and loot sheets are both sheets.
func transferKindOf(panel unison.Paneler) transferKind {
	if xreflect.IsNil(panel) {
		return transferLibrary
	}
	switch unison.AncestorOrSelf[unison.Dockable](panel).(type) {
	case *Sheet, *LootSheet:
		return transferSheet
	case *Template:
		return transferTemplate
	default:
		return transferLibrary
	}
}

// applyOptions selects the steps of applyTransfer that rows arriving in a document go through.
type applyOptions struct {
	// resolvePickers asks the user to settle the template choices the rows carry, which the destination can't hold.
	resolvePickers bool
	// askAncestry offers to disable the character's existing ancestry when the rows bring one of their own.
	askAncestry bool
	// promptForChoices asks the user to settle the rows' modifiers and nameables.
	promptForChoices bool
	// randomize offers to reapply the profile randomization when the rows bring an ancestry.
	randomize bool
	// suppressRandomizePrompt randomizes without asking first.
	suppressRandomizePrompt bool
	// clearPreconfigured clears the Preconfigured flag, which only means something in a template.
	clearPreconfigured bool
	// stripPickers removes the template choices the rows carry, once the user agrees to it, since the destination can't
	// hold them and has no way to settle them either.
	stripPickers bool
	// merge folds the points of rows that duplicate a row already present into that row.
	merge bool
}

// applyOptionsFor returns the steps that rows moving from the source's document into the destination's go through.
// Rows arriving on a sheet from anywhere but another sheet are fully applied; from another sheet they are a plain copy,
// save for settling any template choices a sheet can't hold and the ancestry question. Rows arriving on a template are
// kept as authored, save that those from a library have their modifiers and nameables prompted for. Rows arriving in a
// library are kept as they are, save for the template choices, which only a template can hold.
func applyOptionsFor(source, destination unison.Paneler) applyOptions {
	from := transferKindOf(source)
	switch transferKindOf(destination) {
	case transferSheet:
		if from == transferSheet {
			return applyOptions{resolvePickers: true, askAncestry: true, clearPreconfigured: true, merge: true}
		}
		return applyOptions{
			resolvePickers:     true,
			askAncestry:        true,
			promptForChoices:   true,
			randomize:          true,
			clearPreconfigured: true,
			merge:              true,
		}
	case transferTemplate:
		return applyOptions{promptForChoices: from == transferLibrary, merge: true}
	default:
		return applyOptions{stripPickers: true, clearPreconfigured: true}
	}
}

// applyPart holds the rows of one type being applied to a document, along with the table they are going into and where
// in it they are to be placed.
type applyPart[T gurps.Node[T]] struct {
	table *unison.Table[*Node[T]]
	rows  []T
	// parent is the container the rows go into; the zero value places them at the top level.
	parent T
	// index is the position among the parent's children the rows go to; -1 places them after the last of them.
	index int
	// placed holds the rows that were placed, which leaves out any that were merged into a row already present.
	placed []T
}

// applyPartOps is what applyTransfer needs of each part, whatever its row type.
type applyPartOps interface {
	resolvePickers() bool
	hasPickerData() bool
	stripPickers()
	promptForModifiers() bool
	promptForNameables() bool
	place(merge bool)
	clearPreconfigured()
	undoData() tableRestorer
	changed() (unison.Paneler, Rebuildable)
}

// applyParts holds everything being applied to a document: the rows of each type, and the body type a template may
// bring along with them.
type applyParts struct {
	bodyType  *gurps.Body
	traits    applyPart[*gurps.Trait]
	skills    applyPart[*gurps.Skill]
	spells    applyPart[*gurps.Spell]
	equipment applyPart[*gurps.Equipment]
	notes     applyPart[*gurps.Note]
}

// newApplyParts returns parts holding just the one given part.
func newApplyParts[T gurps.Node[T]](part applyPart[T]) *applyParts {
	var parts applyParts
	switch p := any(part).(type) {
	case applyPart[*gurps.Trait]:
		parts.traits = p
	case applyPart[*gurps.Skill]:
		parts.skills = p
	case applyPart[*gurps.Spell]:
		parts.spells = p
	case applyPart[*gurps.Equipment]:
		parts.equipment = p
	case applyPart[*gurps.Note]:
		parts.notes = p
	}
	return &parts
}

// all calls fn for each part in turn, stopping at the first to return false, in which case false is returned.
func (a *applyParts) all(fn func(part applyPartOps) bool) bool {
	for _, part := range []applyPartOps{&a.traits, &a.skills, &a.spells, &a.equipment, &a.notes} {
		if !fn(part) {
			return false
		}
	}
	return true
}

// each calls fn for each part in turn.
func (a *applyParts) each(fn func(part applyPartOps)) {
	a.all(func(part applyPartOps) bool {
		fn(part)
		return true
	})
}

// hasPickerData returns true if any of the parts' rows carry template picker data.
func (a *applyParts) hasPickerData() bool {
	return !a.all(func(part applyPartOps) bool { return !part.hasPickerData() })
}

func (p *applyPart[T]) resolvePickers() bool {
	revised, abort := processPickerRows(p.rows)
	if abort {
		return false
	}
	p.rows = revised
	return true
}

func (p *applyPart[T]) hasPickerData() bool {
	return gurps.HasTemplatePickerData(p.rows...)
}

func (p *applyPart[T]) stripPickers() {
	gurps.ClearTemplatePickerData(p.rows...)
}

// promptForModifiers and promptForNameables are handed no owner, since the rows aren't in any table yet and so there is
// nothing to rebuild when an answer is given.
func (p *applyPart[T]) promptForModifiers() bool {
	return ProcessModifiers(nil, p.rows)
}

func (p *applyPart[T]) promptForNameables() bool {
	return ProcessNameables(nil, p.rows)
}

// place puts the rows into their table and leaves them selected. With merge, the points of any row that duplicates one
// already present are first folded into that row instead (see mergePoints), and the row is left out.
func (p *applyPart[T]) place(merge bool) {
	if p.table == nil || len(p.rows) == 0 {
		return
	}
	provider, ok := p.table.ClientData()[TableProviderClientKey].(TableProvider[T])
	if !ok {
		return
	}
	selMap := make(map[tid.TID]bool)
	p.placed = p.rows
	if merge {
		// The rows aren't among their parent's children yet, so they are merged as top-level rows: mergePoints takes a
		// merged row out of its parent's children when it has one, which would leave a row not yet placed there in the
		// list to be placed, its points counted twice. The parent is set again when the rows are placed below.
		var noParent T
		SetParents(p.rows, noParent)
		p.placed = mergeIncoming(p.table, p.rows, selMap)
	}
	var siblings []T
	if xreflect.IsNil(p.parent) {
		siblings = provider.RootData()
	} else {
		siblings = p.parent.NodeChildren()
	}
	index := p.index
	if index < 0 || index > len(siblings) {
		index = len(siblings)
	}
	SetParents(p.placed, p.parent)
	siblings = slices.Insert(slices.Clone(siblings), index, p.placed...)
	if xreflect.IsNil(p.parent) {
		provider.SetRootData(siblings)
	} else {
		p.parent.SetChildren(siblings)
		// Opened so that the rows show, since only rows that are showing can be selected, and the provider's own
		// processing below works from the selection.
		p.parent.SetOpen(true)
	}
	p.table.SyncToModel()
	for _, one := range p.placed {
		selMap[one.ID()] = true
	}
	p.table.SetSelectionMap(selMap)
	provider.ProcessDropData(nil, p.table)
	p.table.ScrollRowCellIntoView(p.table.LastSelectedRowIndex(), 0)
	p.table.ScrollRowCellIntoView(p.table.FirstSelectedRowIndex(), 0)
}

func (p *applyPart[T]) clearPreconfigured() {
	gurps.Traverse(func(row T) bool {
		if tl, ok := any(row).(gurps.Preconfigurable); ok {
			tl.SetPreconfigured(false)
		}
		return false
	}, false, false, p.placed...)
}

func (p *applyPart[T]) undoData() tableRestorer {
	if p.table != nil {
		if data := NewTableUndoEditData(p.table); data != nil {
			return data
		}
	}
	return nil
}

// changed returns the table the rows went into, along with its owner, if any, for reporting the change (see
// restoredTables). A row that merged into one already present changed that row, so the table is returned even when no
// row was placed; nothing is returned only when there were no rows at all.
func (p *applyPart[T]) changed() (unison.Paneler, Rebuildable) {
	if p.table == nil || len(p.rows) == 0 {
		return nil, nil
	}
	if owner, ok := p.table.ClientData()[TableOwnerClientKey].(Rebuildable); ok {
		return p.table, owner
	}
	return p.table, nil
}

// applyTransfer applies the parts to the destination -- a sheet having a template applied to it, or the table rows are
// being copied or dropped into -- going through the steps opts selects. It works in two phases. First, every question
// is put to the user against the incoming rows alone, before they have been added to anything: the template choices,
// the ancestry question, the modifiers, the nameables and the randomization. Canceling any of them discards the rows,
// leaving the destination exactly as it was, and returns false. Only once every answer is in is the destination
// changed, all at once, and the change is recorded as a single undo edit with the given name.
func applyTransfer(destination unison.Paneler, parts *applyParts, opts applyOptions, editName string) bool {
	if opts.resolvePickers && !promptForPickers(parts) {
		return false
	}
	// The ancestries arriving are only known once the template choices have been settled, since an ancestry may be one
	// of the options of a choice and not be chosen.
	sheet, isSheet := unison.AncestorOrSelf[unison.Dockable](destination).(*Sheet)
	var entity *gurps.Entity
	var incomingAncestries []*gurps.Ancestry
	if isSheet {
		entity = sheet.Entity()
		incomingAncestries = gurps.ActiveAncestries(parts.traits.rows)
	}
	disableExistingAncestries := false
	if opts.askAncestry && len(incomingAncestries) != 0 {
		if existing := gurps.ActiveAncestries(entity.Traits); len(existing) != 0 {
			disableExistingAncestries = askToDisableExistingAncestry(incomingAncestries[0].Name, existing[0].Name)
		}
	}
	stripPickers := opts.stripPickers && parts.hasPickerData()
	if stripPickers && !confirmTemplatePickerDataRemoval() {
		return false
	}
	if opts.promptForChoices &&
		!(parts.all(func(part applyPartOps) bool { return part.promptForModifiers() }) &&
			parts.all(func(part applyPartOps) bool { return part.promptForNameables() })) {
		return false
	}
	randomize := opts.randomize && len(incomingAncestries) != 0 && gurps.GlobalSettings().General.AutoFillProfile &&
		(opts.suppressRandomizePrompt || askToRandomizeAgain())

	// Every answer is in, so the destination is changed from this point on. The undo manager is found first, since the
	// rebuild that reports the change can replace the table it would otherwise be found through.
	mgr := unison.UndoManagerFor(destination)
	before := newApplyUndoEditData(sheet, parts)
	if stripPickers {
		parts.each(applyPartOps.stripPickers)
	}
	if entity != nil && parts.bodyType != nil {
		entity.SheetSettings.BodyType = parts.bodyType.Clone(entity, nil)
	}
	if disableExistingAncestries {
		for _, one := range gurps.ActiveAncestryTraits(entity.Traits) {
			one.Disabled = true
		}
	}
	// The change is reported once for the whole of it, after every table has been changed (see restoredTables). A sheet
	// is always rebuilt, since its body type may have changed even when no rows were placed.
	var changed restoredTables
	if isSheet {
		changed.add(nil, sheet)
	}
	parts.each(func(part applyPartOps) {
		part.place(opts.merge)
		if opts.clearPreconfigured {
			part.clearPreconfigured()
		}
		changed.add(part.changed())
	})
	if randomize {
		entity.Profile.ApplyRandomizers(entity)
		updateRandomizedProfileFieldsWithoutUndo(sheet)
	}
	changed.report()
	if mgr != nil {
		mgr.Add(&unison.UndoEdit[*applyUndoEditData]{
			ID:         unison.NextUndoID(),
			EditName:   editName,
			UndoFunc:   func(e *unison.UndoEdit[*applyUndoEditData]) { e.BeforeData.Apply() },
			RedoFunc:   func(e *unison.UndoEdit[*applyUndoEditData]) { e.AfterData.Apply() },
			AbsorbFunc: func(_ *unison.UndoEdit[*applyUndoEditData], _ unison.Undoable) bool { return false },
			BeforeData: before,
			AfterData:  newApplyUndoEditData(sheet, parts),
		})
	}
	return true
}

// askToDisableExistingAncestry asks whether the character's existing ancestry should be disabled in favor of the one
// arriving. It is held in a variable so that tests can substitute a non-interactive implementation.
var askToDisableExistingAncestry = func(incoming, existing string) bool {
	return unison.YesNoDialog(fmt.Sprintf(i18n.Text(`An Ancestry (%s) is being added.
Disable your character's existing Ancestry (%s)?`), incoming, existing), "") == unison.ModalResponseOK
}

// askToRandomizeAgain asks whether the profile randomization should be applied again for the ancestry arriving. It is
// held in a variable so that tests can substitute a non-interactive implementation.
var askToRandomizeAgain = func() bool {
	return unison.YesNoDialog(i18n.Text("Would you like to apply the initial randomization again?"), "") ==
		unison.ModalResponseOK
}

// confirmTemplatePickerDataRemoval asks whether rows carrying template picker data may have that data removed so they
// can be added to something other than a template. It is held in a variable so that tests can substitute a
// non-interactive implementation.
var confirmTemplatePickerDataRemoval = func() bool {
	remove := unison.NewOKButtonInfo()
	remove.Title = i18n.Text("Remove Choices")
	dialog, err := unison.NewDialog(unison.DefaultDialogTheme.ErrorIcon, unison.DefaultDialogTheme.ErrorIconInk,
		unison.NewMessagePanel(i18n.Text("Template choices can only be kept in a template"),
			i18n.Text(`Some of the rows being added carry template choices, which only a template can hold.
Continuing removes the choices, leaving the rows otherwise as they are.`)),
		[]*unison.DialogButtonInfo{unison.NewCancelButtonInfo(), remove})
	if err != nil {
		errs.Log(err)
		return false
	}
	return dialog.RunModal() == unison.ModalResponseOK
}

// applyUndoEditData holds what applyTransfer can change: the tables the parts go into and, for a character sheet, the
// randomized profile fields and the body type.
type applyUndoEditData struct {
	sheet    *Sheet
	profile  gurps.ProfileRandom
	bodyType *gurps.Body
	tables   []tableRestorer
}

// newApplyUndoEditData collects the undo edit data for the tables the parts go into and, when sheet isn't nil, for the
// character sheet.
func newApplyUndoEditData(sheet *Sheet, parts *applyParts) *applyUndoEditData {
	data := &applyUndoEditData{sheet: sheet}
	if sheet != nil {
		entity := sheet.Entity()
		data.profile = entity.Profile.ProfileRandom
		if entity.SheetSettings.BodyType != nil {
			data.bodyType = entity.SheetSettings.BodyType.Clone(entity, nil)
		}
	}
	parts.each(func(part applyPartOps) {
		if restorer := part.undoData(); restorer != nil {
			data.tables = append(data.tables, restorer)
		}
	})
	return data
}

// Apply the data.
func (a *applyUndoEditData) Apply() {
	if a.sheet != nil {
		entity := a.sheet.Entity()
		entity.Profile.ProfileRandom = a.profile
		if a.bodyType != nil {
			entity.SheetSettings.BodyType = a.bodyType.Clone(entity, nil)
		}
		updateRandomizedProfileFieldsWithoutUndo(a.sheet)
	}
	// Every table is put back before any of them is reported, so that the document is updated just once.
	var restored restoredTables
	for _, one := range a.tables {
		restored.add(one.restore())
	}
	restored.report()
}

// newAppendPart returns a part holding clones of the rows, to be placed after the table's last top-level row.
func newAppendPart[T gurps.Node[T]](table *unison.Table[*Node[T]], rows []*Node[T]) applyPart[T] {
	clones := make([]T, 0, len(rows))
	for _, row := range rows {
		clones = append(clones, row.CloneForTarget(table, nil).Data())
	}
	return applyPart[T]{table: table, rows: clones, index: -1}
}
