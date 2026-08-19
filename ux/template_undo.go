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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/errs"
)

// ApplyTemplateUndoEditData holds the sheet table data for an undo.
type ApplyTemplateUndoEditData struct {
	sheet     *Sheet
	profile   gurps.ProfileRandom
	bodyType  *gurps.Body
	traits    PreservedTableData[*gurps.Trait]
	skills    PreservedTableData[*gurps.Skill]
	spells    PreservedTableData[*gurps.Spell]
	equipment PreservedTableData[*gurps.Equipment]
	notes     PreservedTableData[*gurps.Note]
}

// NewApplyTemplateUndoEditData creates a new undo that preserves the current sheet table data.
func NewApplyTemplateUndoEditData(sheet *Sheet) (*ApplyTemplateUndoEditData, error) {
	var data ApplyTemplateUndoEditData
	data.sheet = sheet
	entity := sheet.Entity()
	data.profile = entity.Profile.ProfileRandom
	if entity.SheetSettings.BodyType != nil {
		data.bodyType = entity.SheetSettings.BodyType.Clone(entity, nil)
	}
	if err := data.traits.Collect(sheet.Traits.Table); err != nil {
		return nil, err
	}
	if err := data.skills.Collect(sheet.Skills.Table); err != nil {
		return nil, err
	}
	if err := data.spells.Collect(sheet.Spells.Table); err != nil {
		return nil, err
	}
	if err := data.equipment.Collect(sheet.CarriedEquipment.Table); err != nil {
		return nil, err
	}
	if err := data.notes.Collect(sheet.Notes.Table); err != nil {
		return nil, err
	}
	return &data, nil
}

// Apply the data.
func (a *ApplyTemplateUndoEditData) Apply() {
	entity := a.sheet.Entity()
	entity.Profile.ProfileRandom = a.profile
	if a.bodyType != nil {
		entity.SheetSettings.BodyType = a.bodyType.Clone(entity, nil)
	}
	updateRandomizedProfileFieldsWithoutUndo(a.sheet)
	// The tables are only restored here, not marked as modified: the rebuild below recalculates the entity, re-syncs
	// every table, refreshes the search results and restores the focus and scroll position, so reporting each table as
	// it is put back would perform all of that work five more times than necessary for a single undo.
	if err := a.traits.Restore(a.sheet.Traits.Table); err != nil {
		errs.Log(err)
	}
	if err := a.skills.Restore(a.sheet.Skills.Table); err != nil {
		errs.Log(err)
	}
	if err := a.spells.Restore(a.sheet.Spells.Table); err != nil {
		errs.Log(err)
	}
	if err := a.equipment.Restore(a.sheet.CarriedEquipment.Table); err != nil {
		errs.Log(err)
	}
	if err := a.notes.Restore(a.sheet.Notes.Table); err != nil {
		errs.Log(err)
	}
	rebuildAsModified(a.sheet, true)
}
