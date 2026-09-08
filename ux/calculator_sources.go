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
	"slices"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

// sheetSource is an entry in a Source popup: the sheet the numbers come from, or none for numbers that are typed in.
type sheetSource struct {
	name  string
	sheet *Sheet
}

func (s sheetSource) String() string {
	return s.name
}

// sheetSourcePicker owns a Source popup and the sheet it names, and drops a sheet that has been closed. What is read
// from the sheet is the caller's business: selected runs when the user picks a source, so that it can seed suggestions
// as well as read values, while refreshed runs when the current sheet's numbers must be re-read. changed is what every
// other control of the calculator runs once it has stored its value.
type sheetSourcePicker struct {
	popup      *unison.PopupMenu[sheetSource]
	sheet      *Sheet
	selected   func(sheet *Sheet)
	refreshed  func()
	changed    func()
	rebuilding bool
}

// addRow adds the Source popup, labeled with the given text, as a row of its own.
func (p *sheetSourcePicker) addRow(rows *calculatorContent, label string) {
	row := rows.addRow(2)
	addPlainLabel(row, label)
	p.popup = unison.NewPopupMenu[sheetSource]()
	p.popup.WillShowMenuCallback = func(_ *unison.PopupMenu[sheetSource]) { p.rebuild() }
	p.popup.SelectionChangedCallback = func(popup *unison.PopupMenu[sheetSource]) {
		if p.rebuilding {
			return
		}
		if source, ok := popup.Selected(); ok {
			p.sheet = source.sheet
			if p.selected != nil {
				p.selected(source.sheet)
			}
		}
		p.changed()
	}
	row.AddChild(p.popup)
	p.rebuild()
}

// preselect makes the sheet the source before the popup has ever been shown, reading the numbers from it as choosing it
// would. It is for a calculator opened from a sheet, which starts out with that sheet chosen.
func (p *sheetSourcePicker) preselect(sheet *Sheet) {
	p.sheet = sheet
	if p.selected != nil {
		p.selected(sheet)
	}
	p.rebuild()
}

// rebuild fills the Source popup with the sheets that are open right now, keeping the current sheet selected if it is
// still among them and otherwise dropping back to typed-in numbers.
func (p *sheetSourcePicker) rebuild() {
	p.rebuilding = true
	defer func() { p.rebuilding = false }()
	p.popup.RemoveAllItems()
	p.popup.AddItem(sheetSource{name: i18n.Text("Manual")})
	sheets := OpenSheets(nil)
	names := sheetSourceNames(sheets)
	selected := 0
	for i, sheet := range sheets {
		p.popup.AddItem(sheetSource{name: names[i], sheet: sheet})
		if sheet == p.sheet {
			selected = i + 1
		}
	}
	if selected == 0 {
		p.sheet = nil
	}
	p.popup.SelectIndex(selected)
}

// refresh drops the sheet if it has been closed, and otherwise re-reads its numbers.
func (p *sheetSourcePicker) refresh() {
	if p.sheet == nil {
		return
	}
	if !slices.Contains(OpenSheets(nil), p.sheet) {
		p.sheet = nil
		p.rebuild()
		return
	}
	if p.refreshed != nil {
		p.refreshed()
	}
}

// entity returns the entity the numbers come from, or nil when they are typed in.
func (p *sheetSourcePicker) entity() *gurps.Entity {
	if p.sheet == nil {
		return nil
	}
	return p.sheet.Entity()
}

// sheetSourceNames returns a name for each sheet: its title, or its full path when another open sheet has the same
// title.
func sheetSourceNames(sheets []*Sheet) []string {
	counts := make(map[string]int, len(sheets))
	for _, sheet := range sheets {
		counts[sheet.String()]++
	}
	names := make([]string, len(sheets))
	for i, sheet := range sheets {
		names[i] = sheet.String()
		if counts[names[i]] > 1 {
			if path := sheet.BackingFilePath(); path != "" {
				names[i] = path
			}
		}
	}
	return names
}

// skillLevelOrDefault returns the entity's level in the named skill, falling back to its default from the attribute
// when the skill is not on the sheet. It is zero when even the default cannot be worked out.
func skillLevelOrDefault(entity *gurps.Entity, name, defaultAttrID string, modifier int) int {
	if sk := entity.BestSkillNamed(name, "", false, nil); sk != nil {
		return sk.CalculateLevel(nil).Level.AsInteger[int]()
	}
	def := &gurps.SkillDefault{DefaultType: defaultAttrID, Modifier: fxp.FromInteger(modifier)}
	level := def.SkillLevelFast(entity, nil, false, nil, true)
	if level == fxp.Min {
		return 0
	}
	return level.AsInteger[int]()
}

// basicLiftFor returns the Basic Lift the entity has with the given ST, or the one a character typed in has with it
// when the entity is nil, worked out under the global default sheet settings.
func basicLiftFor(entity *gurps.Entity, st fxp.Int) fxp.Weight {
	if entity != nil {
		return entity.BasicLiftForST(st)
	}
	return gurps.BasicLiftForST(st, gurps.SheetSettingsFor(nil).DamageProgression)
}

// thrustFor returns the thrust damage for the given ST under the entity's damage progression, or the global default one
// when the entity is nil.
func thrustFor(entity *gurps.Entity, st fxp.Int) dice.Dice {
	return gurps.SheetSettingsFor(entity).DamageProgression.Thrust(st.AsInteger[int]())
}

// newSourcedWeightField returns a weight field shown in the weight units the sheet settings of the entity that source
// names prefer, or the global default ones when it names none. The units follow the source as it changes, so a field
// whose source has just changed is synced to show the new ones.
func newSourcedWeightField(undoTitle string, source func() *gurps.Entity, get func() fxp.Weight, set func(fxp.Weight), minValue, maxValue fxp.Weight) *WeightField {
	return newUnitsField(nil, "", undoTitle, get, set,
		func(value fxp.Weight) string {
			return gurps.SheetSettingsFor(source()).DefaultWeightUnits.Format(value)
		},
		func(s string) (fxp.Weight, error) {
			return fxp.WeightFromString(s, gurps.SheetSettingsFor(source()).DefaultWeightUnits)
		}, minValue, maxValue, false)
}

// torsoDR returns the entity's total DR on the torso and the part of it that comes from armor.
func torsoDR(entity *gurps.Entity) (total, armor int) {
	body := entity.SheetSettings.BodyType
	if body == nil {
		return 0, 0
	}
	torso := body.LookupLocationByID(entity, gurps.TorsoID)
	if torso == nil {
		return 0, 0
	}
	return torso.DR(entity, nil, nil)[gurps.AllID], torso.ArmorDR(entity, nil)[gurps.AllID]
}

// sheetSourceUser is implemented by the dockables that draw their numbers from open character sheets.
type sheetSourceUser interface {
	sheetChanged(sheet *Sheet)
}

// UpdateCalculatorsForSheet brings every calculator that draws numbers from sheets up to date with the one that has
// just changed. Nothing announces a sheet closing, so this is also one of the places a source that names a closed sheet
// is noticed and dropped; the others are the Source popup, just before it opens, and every change to a control.
func UpdateCalculatorsForSheet(sheet *Sheet) {
	for _, d := range AllDockables() {
		if user, ok := d.AsPanel().Self.(sheetSourceUser); ok {
			user.sheetChanged(sheet)
		}
	}
}
