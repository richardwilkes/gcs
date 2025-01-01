// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/toolbox/errs"
)

// ApplyTemplateUndoEditData holds the sheet table data for an undo.
type ApplyTemplateUndoEditData struct {
	sheet     *Sheet
	profile   gurps.ProfileRandom
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
	data.profile = sheet.Entity().Profile.ProfileRandom
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
	a.sheet.Entity().Profile.ProfileRandom = a.profile
	updateRandomizedProfileFieldsWithoutUndo(a.sheet)
	if err := a.traits.Apply(a.sheet.Traits.Table); err != nil {
		errs.Log(err)
	}
	if err := a.skills.Apply(a.sheet.Skills.Table); err != nil {
		errs.Log(err)
	}
	if err := a.spells.Apply(a.sheet.Spells.Table); err != nil {
		errs.Log(err)
	}
	if err := a.equipment.Apply(a.sheet.CarriedEquipment.Table); err != nil {
		errs.Log(err)
	}
	if err := a.notes.Apply(a.sheet.Notes.Table); err != nil {
		errs.Log(err)
	}
	a.sheet.Rebuild(true)
}
