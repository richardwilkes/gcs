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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

// CanApplyTemplate returns true if a template can be applied.
func CanApplyTemplate() bool {
	return len(OpenSheets(nil)) > 0
}

func installTemplateApplyCmdHandler(template *Template) {
	template.InstallCmdHandlers(
		ApplyTemplateItemID,
		func(_ any) bool { return CanApplyTemplate() },
		func(_ any) { ApplyTemplate(template, true) },
	)
}

// ApplyTemplateFile loads the specified template file and applies it to one or more open sheets.
func ApplyTemplateFile(filePath string) {
	t, err := NewTemplateFromFile(filePath)
	if err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to load template"), err)
		return
	}
	if CanApplyTemplate() {
		if t, ok := t.(*Template); ok {
			ApplyTemplate(t, false)
		}
	}
}

// ApplyTemplate applies a template to one or more open sheets.
func ApplyTemplate(template *Template, suppressRandomizePrompt bool) {
	for _, sheet := range PromptForDestination(OpenSheets(nil)) {
		ApplyTemplateToSheet(template, sheet, suppressRandomizePrompt)
	}
}

// templateParts holds what a template contributes to a sheet: its body type, and its rows cloned for insertion. A
// partial application carries only some of them, which is why each is separately nil-able rather than the parts being
// read back off the template as they are needed.
type templateParts struct {
	body      *gurps.Body
	traits    []*Node[*gurps.Trait]
	skills    []*Node[*gurps.Skill]
	spells    []*Node[*gurps.Spell]
	equipment []*Node[*gurps.Equipment]
	notes     []*Node[*gurps.Note]
}

// ApplyTemplateToSheet applies a template to a sheet.
func ApplyTemplateToSheet(template *Template, sheet *Sheet, suppressRandomizePrompt bool) bool {
	return ApplyTemplateToSheetWithPickers(template, sheet, suppressRandomizePrompt, processTemplateParts)
}

// ApplyTemplateToSheetWithPickers applies a template to a sheet, using processPickers to resolve any template pickers
// the template's rows contain. The picker processing is passed in so that headless tests, which have no way to respond
// to the dialogs it would otherwise present, can substitute their own.
func ApplyTemplateToSheetWithPickers(template *Template, sheet *Sheet, suppressRandomizePrompt bool, processPickers processTemplatePartsFunc) bool {
	parts := templateParts{
		body:      template.template.BodyType,
		traits:    cloneRows(sheet.Traits.Table, template.Traits.Table.RootRows()),
		skills:    cloneRows(sheet.Skills.Table, template.Skills.Table.RootRows()),
		spells:    cloneRows(sheet.Spells.Table, template.Spells.Table.RootRows()),
		equipment: cloneRows(sheet.CarriedEquipment.Table, template.Equipment.Table.RootRows()),
		notes:     cloneRows(sheet.Notes.Table, template.Notes.Table.RootRows()),
	}
	return ApplyTemplateParts(&parts, sheet, processPickers, suppressRandomizePrompt)
}

// ApplyTemplateParts applies already-gathered template parts to a sheet, recording one undo edit for the whole of it.
// Nothing is recorded when a picker is canceled, since the sheet is then left exactly as it was.
func ApplyTemplateParts(parts *templateParts, sheet *Sheet, processPickers processTemplatePartsFunc, suppressRandomizePrompt bool) bool {
	var undo *unison.UndoEdit[*ApplyTemplateUndoEditData]
	mgr := unison.UndoManagerFor(sheet)
	if mgr != nil {
		if beforeData, err := NewApplyTemplateUndoEditData(sheet); err != nil {
			errs.Log(err)
			mgr = nil
		} else {
			undo = &unison.UndoEdit[*ApplyTemplateUndoEditData]{
				ID:         unison.NextUndoID(),
				EditName:   i18n.Text("Apply Template"),
				UndoFunc:   func(e *unison.UndoEdit[*ApplyTemplateUndoEditData]) { e.BeforeData.Apply() },
				RedoFunc:   func(e *unison.UndoEdit[*ApplyTemplateUndoEditData]) { e.AfterData.Apply() },
				AbsorbFunc: func(_ *unison.UndoEdit[*ApplyTemplateUndoEditData], _ unison.Undoable) bool { return false },
				BeforeData: beforeData,
			}
		}
	}

	if !applyTemplateParts(parts, sheet, processPickers, suppressRandomizePrompt) {
		return false
	}

	if mgr != nil && undo != nil {
		var err error
		if undo.AfterData, err = NewApplyTemplateUndoEditData(sheet); err != nil {
			errs.Log(err)
		} else {
			mgr.Add(undo)
		}
	}
	return true
}

// applyTemplateParts does the work of an application, without the undo bookkeeping ApplyTemplateParts wraps it in.
func applyTemplateParts(parts *templateParts, sheet *Sheet, processPickers processTemplatePartsFunc, suppressRandomizePrompt bool) bool {
	e := sheet.Entity()

	// Nothing until the pickers have been dealt with may modify the sheet: canceling a picker abandons the entire
	// operation, which must leave the sheet exactly as it was.
	if !processPickers(parts) {
		return false
	}

	// The sheet is modified from this point on.

	if parts.body != nil {
		e.SheetSettings.BodyType = parts.body.Clone(e, nil)
	}

	// The Ancestry question is asked about what survived the pickers, not about what the template arrived with. A
	// template whose Ancestry sits inside a picker only has one if the user picked it, and asking before the pickers
	// ran meant being asked to disable an existing Ancestry to make way for one that was then discarded.
	templateAncestries := gurps.ActiveAncestries(ExtractNodeDataFromList(parts.traits))
	if len(templateAncestries) != 0 {
		entityAncestries := gurps.ActiveAncestries(e.Traits)
		if len(entityAncestries) != 0 {
			if unison.YesNoDialog(fmt.Sprintf(i18n.Text(`The template contains an Ancestry (%s).
Disable your character's existing Ancestry (%s)?`),
				templateAncestries[0].Name, entityAncestries[0].Name), "") == unison.ModalResponseOK {
				for _, one := range gurps.ActiveAncestryTraits(e.Traits) {
					one.Disabled = true
				}
			}
		}
	}
	appendRows(sheet.Traits.Table, parts.traits)
	appendRows(sheet.Skills.Table, parts.skills)
	appendRows(sheet.Spells.Table, parts.spells)
	appendRows(sheet.CarriedEquipment.Table, parts.equipment)
	appendRows(sheet.Notes.Table, parts.notes)

	// Present the decisions the added rows carry by the same route a copy or a drop takes, which is what puts the
	// modifier prompt ahead of the nameables one (see promptForAddedRows). Every list is asked about, not just the
	// ones that were being asked about before: a skill, a spell or a note arriving from a template can carry nameable
	// keys as readily as a trait can.
	promptForAddedRows(sheet.Traits.Table)
	promptForAddedRows(sheet.Skills.Table)
	promptForAddedRows(sheet.Spells.Table)
	promptForAddedRows(sheet.CarriedEquipment.Table)
	promptForAddedRows(sheet.Notes.Table)

	// Merge the added rows into identical existing ones now that their decisions are settled, since the merge match
	// includes the nameable replacements. This is why the prompts have to come first: merging before them would
	// compare empty replacements against the already-resolved existing rows and add duplicates instead of merging.
	MergeAddedRows(sheet.Traits.Table)
	MergeAddedRows(sheet.Skills.Table)
	MergeAddedRows(sheet.Spells.Table)
	MergeAddedRows(sheet.CarriedEquipment.Table)
	MergeAddedRows(sheet.OtherEquipment.Table)
	// A sheet never allows the Preconfigured flag, so there is no question to ask about whether to clear it here --
	// unlike a copy or a drop, which can land in a document that does allow it. The sweep covers every list on the
	// sheet rather than only the ones an apply adds to, so that it stays right regardless of where rows end up.
	clearPreconfiguredFlag(sheet.Traits.Table.RootRows())
	clearPreconfiguredFlag(sheet.Skills.Table.RootRows())
	clearPreconfiguredFlag(sheet.Spells.Table.RootRows())
	clearPreconfiguredFlag(sheet.CarriedEquipment.Table.RootRows())
	clearPreconfiguredFlag(sheet.OtherEquipment.Table.RootRows())
	clearPreconfiguredFlag(sheet.Notes.Table.RootRows())

	if len(templateAncestries) != 0 && gurps.GlobalSettings().General.AutoFillProfile {
		randomize := true
		if !suppressRandomizePrompt {
			randomize = unison.YesNoDialog(i18n.Text("Would you like to apply the initial randomization again?"), "") == unison.ModalResponseOK
		}
		if randomize {
			e.Profile.ApplyRandomizers(e)
			updateRandomizedProfileFieldsWithoutUndo(sheet)
		}
	}
	// One rebuild covers the whole application. The prompts above rebuild as they are answered, which can leave this
	// sheet an orphan, so the rebuild is aimed at whatever is live rather than at the sheet this started with.
	rebuildAsModified(unison.AncestorOrSelf[Rebuildable](liveOwner(sheet)), true)

	sheet.Window().ToFront()
	sheet.RequestFocus()
	return true
}

// templatePartsFromRows gathers the rows of a copy or a drag into the parts of a template application, cloned into the
// destination sheet's tables. The rows of a single transfer are always of one type -- a drag carries a single drag key
// and a copy a single block key -- so exactly one of the parts is ever filled in. Only traits, skills and spells can
// carry template choices today, so a row type that cannot is not a partial template application and returns nil; an
// equipment case belongs here once equipment gains choices of its own.
func templatePartsFromRows[T gurps.Node[T]](sheet *Sheet, rows []*Node[T]) *templateParts {
	var parts templateParts
	switch typed := any(rows).(type) {
	case []*Node[*gurps.Trait]:
		parts.traits = cloneRows(sheet.Traits.Table, typed)
	case []*Node[*gurps.Skill]:
		parts.skills = cloneRows(sheet.Skills.Table, typed)
	case []*Node[*gurps.Spell]:
		parts.spells = cloneRows(sheet.Spells.Table, typed)
	default:
		return nil
	}
	return &parts
}

// applyPartialTemplate applies a selection of a template's rows to a sheet as a partial template application: the same
// pipeline a whole template runs, carrying only the parts the selection holds and leaving the body type alone. This is
// what a copy or a drag of rows bearing template choices onto a sheet performs, so that the one way choices are made
// is the one the template tier already uses. Returns false if a choice was canceled, in which case the sheet has been
// left untouched.
func applyPartialTemplate[T gurps.Node[T]](sheet *Sheet, rows []*Node[T]) bool {
	parts := templatePartsFromRows(sheet, rows)
	if parts == nil {
		return false
	}
	return ApplyTemplateParts(parts, sheet, processTemplateParts, false)
}
