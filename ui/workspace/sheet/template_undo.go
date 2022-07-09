package sheet

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ui/widget/ntable"
	"github.com/richardwilkes/toolbox/log/jot"
)

// ApplyTemplateUndoEditData holds the sheet table data for an undo.
type ApplyTemplateUndoEditData struct {
	sheet     *Sheet
	traits    ntable.PreservedTableData[*gurps.Trait]
	skills    ntable.PreservedTableData[*gurps.Skill]
	spells    ntable.PreservedTableData[*gurps.Spell]
	equipment ntable.PreservedTableData[*gurps.Equipment]
	notes     ntable.PreservedTableData[*gurps.Note]
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
