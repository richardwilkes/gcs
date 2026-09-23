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
	// Nothing from here until the pickers have been dealt with may modify the sheet: canceling a picker abandons the
	// entire operation, which must leave the sheet exactly as it was. That includes the Ancestry question below, which
	// is asked here to preserve the order the questions are presented in, but not acted upon until the operation is
	// known to be going through.
	disableExistingAncestries := false
	templateAncestries := gurps.ActiveAncestries(ExtractNodeDataFromList(parts.traits))
	if len(templateAncestries) != 0 {
		entityAncestries := gurps.ActiveAncestries(e.Traits)
		if len(entityAncestries) != 0 {
			disableExistingAncestries = unison.YesNoDialog(fmt.Sprintf(i18n.Text(`The template contains an Ancestry (%s).
Disable your character's existing Ancestry (%s)?`),
				templateAncestries[0].Name, entityAncestries[0].Name), "") == unison.ModalResponseOK
		}
	}
	if !processPickers(parts) {
		return false // A picker was canceled, so the sheet has been left untouched.
	}
	// The sheet is modified from this point on.
	if parts.body != nil {
		e.SheetSettings.BodyType = parts.body.Clone(e, nil)
	}
	if disableExistingAncestries {
		for _, one := range gurps.ActiveAncestryTraits(e.Traits) {
			one.Disabled = true
		}
	}
	// Skills and spells merge points with identical existing rows during appendRows, and the merge match includes the
	// nameable replacements. Since they have no modifiers to toggle, resolve their nameables up front so the merge
	// compares against the final replacements; otherwise re-applying a template would compare empty replacements
	// against the already-resolved existing rows and add duplicates instead of merging.
	ProcessNameables(sheet.Skills.Table, ExtractNodeDataFromList(parts.skills), true)
	ProcessNameables(sheet.Spells.Table, ExtractNodeDataFromList(parts.spells), true)
	appendRows(sheet.Traits.Table, parts.traits)
	appendRows(sheet.Skills.Table, parts.skills)
	appendRows(sheet.Spells.Table, parts.spells)
	appendRows(sheet.CarriedEquipment.Table, parts.equipment)
	appendRows(sheet.Notes.Table, parts.notes)
	rebuildAsModified(sheet, true)
	ProcessModifiersForSelection(sheet.Traits.Table, true)
	ProcessModifiersForSelection(sheet.CarriedEquipment.Table, true)
	ProcessNameablesForSelection(sheet.Traits.Table, true)
	ProcessNameablesForSelection(sheet.CarriedEquipment.Table, true)
	ProcessNameablesForSelection(sheet.Notes.Table, true)
	maybeClearPreconfiguredFlag(sheet.Traits.Table, sheet.Traits.Table.RootRows())
	maybeClearPreconfiguredFlag(sheet.Skills.Table, sheet.Skills.Table.RootRows())
	maybeClearPreconfiguredFlag(sheet.Spells.Table, sheet.Spells.Table.RootRows())
	maybeClearPreconfiguredFlag(sheet.CarriedEquipment.Table, sheet.CarriedEquipment.Table.RootRows())
	maybeClearPreconfiguredFlag(sheet.Notes.Table, sheet.Notes.Table.RootRows())

	if len(templateAncestries) != 0 && gurps.GlobalSettings().General.AutoFillProfile {
		randomize := true
		if !suppressRandomizePrompt {
			randomize = unison.YesNoDialog(i18n.Text("Would you like to apply the initial randomization again?"), "") == unison.ModalResponseOK
		}
		if randomize {
			e.Profile.ApplyRandomizers(e)
			updateRandomizedProfileFieldsWithoutUndo(sheet)
			rebuildAsModified(sheet, true)
		}
	}
	sheet.Window().ToFront()
	sheet.RequestFocus()
	return true
}
