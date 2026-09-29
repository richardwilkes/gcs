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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/promptstep"
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
	// normalizeChoices strips the template choice containers among the rows of everything a choice container doesn't
	// use (see gurps.NormalizeTemplateChoiceContainers), as a template does to the rows it loads.
	normalizeChoices bool
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
	// clearPreconfigured clears the Preconfigured flag, which means nothing on a sheet.
	clearPreconfigured bool
	// stripPickers removes the template choices the rows carry, once the user agrees to it, since the destination can't
	// hold them and has no way to settle them either.
	stripPickers bool
	// merge folds the points of rows that duplicate a row already present into that row.
	merge bool
}

// applyOptionsFor returns the steps that rows moving from the source's document into the destination's go through. Rows
// arriving on a sheet from anywhere but another sheet are fully applied; from another sheet they are a plain copy, save
// for settling any template choices a sheet can't hold, the ancestry question and the offer to randomize. Rows arriving
// on a template are kept as authored, save that their choice containers are normalized and those from a library have
// their modifiers and nameables prompted for. Rows arriving in a library are kept as they are, Preconfigured flag
// included, save for the template choices, which only a template can hold.
func applyOptionsFor(source, destination unison.Paneler) applyOptions {
	from := transferKindOf(source)
	switch transferKindOf(destination) {
	case transferSheet:
		if from == transferSheet {
			return applyOptions{
				resolvePickers:     true,
				askAncestry:        true,
				randomize:          true,
				clearPreconfigured: true,
				merge:              true,
			}
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
		return applyOptions{normalizeChoices: true, promptForChoices: from == transferLibrary, merge: true}
	default:
		return applyOptions{stripPickers: true}
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
	normalizeChoices()
	resolvePickers(op promptOperation) bool
	pickerContainers() []string
	stripPickers()
	modifierTargetCount() int
	promptForModifiers(op promptOperation, done, total int) (asked int, ok bool)
	promptForNameables(op promptOperation) bool
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

// promptForModifiers puts up the modifier prompts for the rows of every part, numbering them across all of the parts
// rather than starting over with each, since the user sees them as one run. Returns false if a prompt was canceled.
func (a *applyParts) promptForModifiers(op promptOperation) bool {
	total := 0
	a.each(func(part applyPartOps) { total += part.modifierTargetCount() })
	done := 0
	return a.all(func(part applyPartOps) bool {
		asked, ok := part.promptForModifiers(op, done, total)
		done += asked
		return ok
	})
}

// pickerContainers returns the names of the containers among the parts' rows that carry template picker data.
func (a *applyParts) pickerContainers() []string {
	var names []string
	a.each(func(part applyPartOps) { names = append(names, part.pickerContainers()...) })
	return names
}

func (p *applyPart[T]) normalizeChoices() {
	gurps.NormalizeTemplateChoiceContainers(p.rows...)
}

func (p *applyPart[T]) resolvePickers(op promptOperation) bool {
	revised, abort := processPickerRows(op, p.rows)
	if abort {
		return false
	}
	p.rows = revised
	return true
}

func (p *applyPart[T]) pickerContainers() []string {
	var names []string
	gurps.Traverse(func(row T) bool {
		if gurps.IsTemplateChoiceContainer(row) {
			names = append(names, row.String())
		}
		return false
	}, false, false, p.rows...)
	return names
}

func (p *applyPart[T]) stripPickers() {
	gurps.ClearTemplatePickerData(p.rows...)
}

func (p *applyPart[T]) modifierTargetCount() int {
	return len(modifierTargets(p.rows))
}

// promptForModifiers and promptForNameables put up the prompts for the rows' modifiers and nameable keys. Neither
// rebuilds or reports anything: the rows aren't in a table yet, and applyTransfer does both once the answers are in.
// promptForModifiers also returns how many rows it asked about, counted before any answer can change that.
func (p *applyPart[T]) promptForModifiers(op promptOperation, done, total int) (asked int, ok bool) {
	targets := modifierTargets(p.rows)
	return len(targets), promptForModifierTargets(op, targets, done, total)
}

func (p *applyPart[T]) promptForNameables(op promptOperation) bool {
	return processNameables(op, p.rows)
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
	gurps.ClearPreconfigured(p.placed...)
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
// the modifiers, the nameables, the ancestry question and the randomization. Canceling any of them discards the rows,
// leaving the destination exactly as it was, and returns false. Only once every answer is in is the destination
// changed, all at once, and the change is recorded as a single undo edit with the given name. The operation describes
// the transfer, such as "Copying Broadsword to Sir Bob", and each question shows it (see newOperationLabel), since the
// questions are otherwise asked with nothing around them to say what they are part of.
func applyTransfer(destination unison.Paneler, parts *applyParts, opts applyOptions, op promptOperation, editName string) bool {
	// The incoming rows are clones, so normalizing them changes nothing elsewhere, even if a later prompt is canceled. It
	// comes first so that no prompt asks about modifiers or nameables a choice container is about to lose.
	if opts.normalizeChoices {
		parts.each(applyPartOps.normalizeChoices)
	}
	if opts.resolvePickers && !promptForPickers(op, parts) {
		return false
	}
	var pickerContainers []string
	if opts.stripPickers {
		pickerContainers = parts.pickerContainers()
	}
	stripPickers := len(pickerContainers) != 0
	if stripPickers && !confirmTemplatePickerDataRemoval(op, pickerContainers) {
		return false
	}
	if opts.promptForChoices &&
		!(parts.promptForModifiers(op) &&
			parts.all(func(part applyPartOps) bool { return part.promptForNameables(op) })) {
		return false
	}
	// The ancestries arriving are only known once the template choices have been settled, since an ancestry may be one
	// of the options of a choice and not be chosen. They are named by their containers rather than by the ancestries
	// the containers link to, since many containers may link to the same ancestry, each with details of its own, and
	// the names are only final once the nameables have been settled.
	sheet, isSheet := unison.AncestorOrSelf[unison.Dockable](destination).(*Sheet)
	var entity *gurps.Entity
	var incomingAncestries []*gurps.Trait
	if isSheet {
		entity = sheet.Entity()
		incomingAncestries = gurps.ActiveAncestryTraits(parts.traits.rows)
	}
	disableExistingAncestries := false
	if opts.askAncestry && len(incomingAncestries) != 0 {
		if existing := gurps.ActiveAncestryTraits(entity.Traits); len(existing) != 0 {
			disableExistingAncestries = askToDisableExistingAncestry(op, ancestryNames(incomingAncestries),
				ancestryNames(existing))
		}
	}
	randomize := opts.randomize && len(incomingAncestries) != 0 && gurps.GlobalSettings().General.AutoFillProfile &&
		(opts.suppressRandomizePrompt || askToRandomizeAgain(op))

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
		// The randomizers work from the character as it now is -- the ancestry's scripts derive height and weight from
		// ST, which the rows just placed may have raised -- so the entity is brought up to date with them first rather
		// than being left for the rebuild below.
		entity.Recalculate()
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

// ancestryNames returns the names of the ancestry containers, with their nameables replaced.
func ancestryNames(ancestries []*gurps.Trait) []string {
	names := make([]string, len(ancestries))
	for i, one := range ancestries {
		names[i] = one.NameWithReplacements()
	}
	return names
}

// askToDisableExistingAncestry asks whether the character's existing ancestries should be disabled in favor of the ones
// arriving. It is held in a variable so that tests can substitute a non-interactive implementation.
var askToDisableExistingAncestry = func(op promptOperation, incoming, existing []string) bool {
	primary := i18n.Text("An ancestry is being added")
	if len(incoming) > 1 {
		primary = i18n.Text("Ancestries are being added")
	}
	question := i18n.Text("Disable the character's existing ancestry?")
	if len(existing) > 1 {
		question = i18n.Text("Disable all of the character's existing ancestries?")
	}
	detail := fmt.Sprintf(i18n.Text("Adding: %s\nExisting: %s\n\n%s"), joinNames(incoming), joinNames(existing),
		question)
	return runPromptDialog(op.at(promptstep.Ancestry), newOperationMessagePanel(op, primary, detail),
		unison.NewNoButtonInfo(), unison.NewYesButtonInfo()) == unison.ModalResponseOK
}

// askToRandomizeAgain asks whether the profile randomization should be applied again for the ancestry arriving. It is
// held in a variable so that tests can substitute a non-interactive implementation.
var askToRandomizeAgain = func(op promptOperation) bool {
	return runPromptDialog(op.at(promptstep.Randomize), newOperationMessagePanel(op,
		i18n.Text("Randomize the character's profile again?"),
		i18n.Text("An ancestry is being added. Randomizing replaces the character's name, gender,\nage, birthday, height, weight, eyes, hair, skin and handedness with new values.")),
		unison.NewNoButtonInfo(), unison.NewYesButtonInfo()) == unison.ModalResponseOK
}

// confirmTemplatePickerDataRemoval asks whether rows carrying template picker data may have that data removed so they
// can be added to something other than a template. The containers are the names of the rows that carry it. It is held
// in a variable so that tests can substitute a non-interactive implementation.
var confirmTemplatePickerDataRemoval = func(op promptOperation, containers []string) bool {
	remove := unison.NewOKButtonInfo()
	remove.Title = i18n.Text("Remove Choices")
	// Unlike the other prompts, this one carries an icon: continuing throws data away, which a warning should mark.
	dialog, err := newPromptDialog(op.at(promptstep.RemoveChoices), unison.DefaultDialogTheme.WarningIcon,
		unison.DefaultDialogTheme.WarningIconInk, newOperationMessagePanel(op,
			i18n.Text("Template choices can only be kept in a template"),
			fmt.Sprintf(i18n.Text(`These rows carry template choices, which only a template can hold:
%s

Continuing removes the choices, leaving the rows otherwise as they are.`), nameList(containers))),
		unison.NewCancelButtonInfo(), remove)
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
