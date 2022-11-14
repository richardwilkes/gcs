/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ux

import (
	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/toolbox/log/jot"
)

// ApplyTemplateUndoEditData holds the sheet table data for an undo.
type ApplyTemplateUndoEditData struct {
	sheet     *Sheet
	traits    PreservedTableData[*model.Trait]
	skills    PreservedTableData[*model.Skill]
	spells    PreservedTableData[*model.Spell]
	equipment PreservedTableData[*model.Equipment]
	notes     PreservedTableData[*model.Note]
}

// NewApplyTemplateUndoEditData creates a new undo that preserves the current sheet table data.
func NewApplyTemplateUndoEditData(sheet *Sheet) (*ApplyTemplateUndoEditData, error) {
	var data ApplyTemplateUndoEditData
	data.sheet = sheet
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
	if err := a.traits.Apply(a.sheet.Traits.Table); err != nil {
		jot.Warn(err)
	}
	if err := a.skills.Apply(a.sheet.Skills.Table); err != nil {
		jot.Warn(err)
	}
	if err := a.spells.Apply(a.sheet.Spells.Table); err != nil {
		jot.Warn(err)
	}
	if err := a.equipment.Apply(a.sheet.CarriedEquipment.Table); err != nil {
		jot.Warn(err)
	}
	if err := a.notes.Apply(a.sheet.Notes.Table); err != nil {
		jot.Warn(err)
	}
	a.sheet.Rebuild(true)
}
