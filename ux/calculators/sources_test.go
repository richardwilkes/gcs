// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package calculators

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// TestCalculatorSheetSources drives the Source popup of every calculator that takes a character's numbers, through the
// life of the sheet the calculators were opened from: the sheet must be preselected with the controls for what it
// supplies locked and filled, choosing Manual must unlock them and keep the numbers, choosing the sheet must lock them
// again, a change to the sheet's numbers must reach every calculator, and closing the sheet must drop each calculator
// back to Manual.
func TestCalculatorSheetSources(t *testing.T) {
	screen, wnd := uxtest.StartHeadlessWorkspace(t, check.New(t))
	sheet := openNewCharacterSheet(t, screen)
	calc := openCalculator(t, screen)

	type source struct {
		name   string
		tab    calculatorTab
		picker *sheetSourcePicker
		// supplied holds the controls for what a sheet supplies.
		supplied []unison.Paneler
		// numbers returns some of what a sheet supplies as the calculator holds it, and fromSheet the same as the sheet
		// has it.
		numbers   func() []fxp.Int
		fromSheet func(entity *gurps.Entity) []fxp.Int
	}
	collision, jumping, throwing, hiking, explosion := calc.collision, calc.jumping, calc.throwing, calc.hiking, calc.explosion
	sources := []source{
		{
			name:   "collisions",
			tab:    collision,
			picker: &collision.mover.sheetSourcePicker,
			supplied: []unison.Paneler{
				collision.mover.hpField, collision.mover.stField, collision.mover.smField, collision.mover.acrobaticsField,
				collision.mover.swimmingField, collision.mover.armorDRField, collision.mover.innateDRField,
			},
			numbers: func() []fxp.Int { return []fxp.Int{collision.mover.hp, collision.mover.st} },
			fromSheet: func(entity *gurps.Entity) []fxp.Int {
				return []fxp.Int{entity.Attributes.Maximum(gurps.HitPointsID), entity.Attributes.Maximum(gurps.StrengthID)}
			},
		},
		{
			name:   "jumping",
			tab:    jumping,
			picker: &jumping.source,
			supplied: []unison.Paneler{
				jumping.basicMoveField, jumping.liftingSTField, jumping.weightField, jumping.encumbrancePopup,
				jumping.jumpingSkillField, jumping.enhancedMoveField, jumping.superJumpField,
			},
			numbers: func() []fxp.Int { return []fxp.Int{jumping.basicMove, jumping.liftingST, fxp.Int(jumping.weight)} },
			fromSheet: func(entity *gurps.Entity) []fxp.Int {
				return []fxp.Int{
					entity.ResolveAttributeCurrent(gurps.BasicMoveID), entity.LiftingStrength(), fxp.Int(entity.Profile.Weight),
				}
			},
		},
		{
			name:   "throwing",
			tab:    throwing,
			picker: &throwing.source,
			supplied: []unison.Paneler{
				throwing.stField, throwing.strikingSTField, throwing.throwingPopup, throwing.throwingArtPopup,
			},
			numbers: func() []fxp.Int { return []fxp.Int{throwing.st, throwing.strikingST} },
			fromSheet: func(entity *gurps.Entity) []fxp.Int {
				return []fxp.Int{entity.LiftingStrength() - entity.LiftingStrengthBonus, entity.StrikingStrength()}
			},
		},
		{
			name:   "hiking",
			tab:    hiking,
			picker: &hiking.source,
			supplied: []unison.Paneler{
				hiking.moveField, hiking.enhancedMoveField, hiking.encumbrancePopup, hiking.fpField, hiking.fitnessPopup,
				hiking.recoverEnergyPopup,
			},
			numbers: func() []fxp.Int { return []fxp.Int{fxp.FromInteger(hiking.move), hiking.fp} },
			fromSheet: func(entity *gurps.Entity) []fxp.Int {
				return []fxp.Int{
					fxp.FromInteger(entity.Move(entity.EncumbranceLevel(false))),
					entity.Attributes.Maximum(gurps.FatiguePointsID),
				}
			},
		},
		{
			name:   "explosions",
			tab:    explosion,
			picker: &explosion.target.sheetSourcePicker,
			supplied: []unison.Paneler{
				explosion.target.smField, explosion.target.hpField, explosion.target.torsoDRField, explosion.target.drField,
			},
			numbers: func() []fxp.Int { return []fxp.Int{explosion.target.hp} },
			fromSheet: func(entity *gurps.Entity) []fxp.Int {
				return []fxp.Int{entity.Attributes.Maximum(gurps.HitPointsID)}
			},
		},
	}

	type state struct {
		sheet     *ux.Sheet
		index     int
		unlocked  []bool
		numbers   []fxp.Int
		fromSheet []fxp.Int
	}
	current := func(one *source) state {
		var s state
		screen.Do(func() {
			s.sheet = one.picker.sheet
			s.index = one.picker.popup.SelectedIndex()
			for _, control := range one.supplied {
				s.unlocked = append(s.unlocked, control.AsPanel().Enabled())
			}
			s.numbers = one.numbers()
			s.fromSheet = one.fromSheet(sheet.Entity())
		})
		return s
	}
	all := func(one *source, unlocked bool) []bool {
		return slices.Repeat([]bool{unlocked}, len(one.supplied))
	}

	for i := range sources {
		one := &sources[i]
		t.Run(one.name, func(t *testing.T) {
			c := check.New(t)
			selectCalculatorTab(t, screen, calc, one.tab)
			s := current(one)
			c.Equal(sheet, s.sheet, "the active sheet must be preselected as the source")
			c.Equal(1, s.index, "the Source popup must show the sheet")
			c.Equal(s.fromSheet, s.numbers, "the numbers must come from the sheet")
			c.Equal(all(one, false), s.unlocked, "the controls for what the sheet supplies must be locked")

			uxtest.ChoosePopupItem(t, screen, wnd, one.picker.popup, 0)
			s = current(one)
			c.Nil(s.sheet, "choosing Manual must drop the sheet")
			c.Equal(s.fromSheet, s.numbers, "the numbers last read from the sheet must be kept")
			c.Equal(all(one, true), s.unlocked, "the controls for what is typed in must be unlocked")

			uxtest.ChoosePopupItem(t, screen, wnd, one.picker.popup, 1)
			s = current(one)
			c.Equal(sheet, s.sheet, "choosing the sheet must make it the source again")
			c.Equal(all(one, false), s.unlocked, "the controls for what the sheet supplies must be locked again")
		})
	}

	// A change to the sheet's numbers reaches every calculator drawing on it, whichever tab is showing.
	raise := func(by fxp.Int) {
		screen.Do(func() {
			entity := sheet.Entity()
			for _, id := range []string{gurps.StrengthID, gurps.BasicMoveID, gurps.FatiguePointsID} {
				attr := entity.Attributes.Set[id]
				attr.SetMaximum(attr.Maximum() + by)
			}
			entity.Recalculate()
			ux.NotifySheetSourceUsers(sheet)
		})
	}
	before := make([][]fxp.Int, len(sources))
	for i := range sources {
		before[i] = current(&sources[i]).numbers
	}
	raise(fxp.Three)
	for i := range sources {
		one := &sources[i]
		t.Run(one.name+" follows the sheet", func(t *testing.T) {
			c := check.New(t)
			s := current(one)
			c.NotEqual(before[i], s.fromSheet, "the sheet's numbers must have changed")
			c.Equal(s.fromSheet, s.numbers, "the numbers must follow the sheet's as they change")
		})
	}
	raise(-fxp.Three) // The sheet is as it was again, so it closes without asking to be saved.

	// Nothing announces a sheet closing, so a calculator notices when its tab is next shown; the first of them here is
	// not the one that was showing as the sheet closed.
	uxtest.CloseEditorWithoutPrompt(t, screen, sheet)
	for i := range sources {
		one := &sources[i]
		t.Run(one.name+" drops the closed sheet", func(t *testing.T) {
			c := check.New(t)
			selectCalculatorTab(t, screen, calc, one.tab)
			s := current(one)
			c.Nil(s.sheet, "closing the sheet must drop it as the source")
			c.Equal(0, s.index, "the Source popup must show Manual once the sheet is gone")
			c.Equal(all(one, true), s.unlocked, "the controls must be unlocked once the sheet is gone")
		})
	}

	uxtest.CloseEditorWithoutPrompt(t, screen, calc)
}

// TestSourcePopupOffersTheOpenSheets verifies that a Source popup is filled as it is opened: sheets opened since the
// calculators were are offered, by their paths when they share a title and by the title otherwise, and a source that
// has been closed since is dropped.
func TestSourcePopupOffersTheOpenSheets(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	calc := openCalculator(t, screen)
	selectCalculatorTab(t, screen, calc, calc.jumping)
	picker := &calc.jumping.source

	dir := t.TempDir()
	paths := make(map[*ux.Sheet]string)
	for _, folder := range []string{"one", "two"} {
		path := filepath.Join(dir, folder, "Twin"+gurps.SheetExt)
		c.NoError(os.MkdirAll(filepath.Dir(path), 0o750))
		c.NoError(gurps.NewEntity().Save(path))
		var twin *ux.Sheet
		screen.Do(func() {
			d, _ := ux.OpenFile(path, gurps.PageInfo{})
			if opened, isSheet := d.(*ux.Sheet); isSheet {
				twin = opened
			}
		})
		if twin == nil {
			t.Fatalf("%s must open as a character sheet", path)
		}
		paths[twin] = path
	}

	// open clicks the popup, which shows its menu, and returns what the popup then offers and the source it names.
	open := func() (names []string, sheets []*ux.Sheet, source *ux.Sheet) {
		screen.Click(screen.PanelCenter(picker.popup))
		screen.Do(func() {
			for i := range picker.popup.ItemCount() {
				if item, ok := picker.popup.ItemAt(i); ok {
					names = append(names, item.name)
					sheets = append(sheets, item.sheet)
				}
			}
			source = picker.sheet
		})
		return names, sheets, source
	}

	names, sheets, _ := open()
	var items []*unison.Panel
	screen.Do(func() { items = uxtest.MenuItemPanels(uxtest.OpenMenuPopup(wnd)) })
	if len(sheets) != 3 || len(items) != 3 {
		t.Fatalf("the Source popup must offer Manual and the two sheets opened since, but offers %q", names)
	}
	c.Equal([]string{"Manual", paths[sheets[1]], paths[sheets[2]]}, names,
		"sheets that share a title must be told apart by their paths")
	screen.Click(screen.PanelCenter(items[2]))
	var source *ux.Sheet
	screen.Do(func() { source = picker.sheet })
	c.Equal(sheets[2], source, "choosing a sheet opened since must make it the source")

	// Nothing announces the source closing, so it is still the source until the popup is opened.
	uxtest.CloseEditorWithoutPrompt(t, screen, sheets[2])
	names, _, source = open()
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Equal([]string{"Manual", "Twin"}, names, "a sheet whose title is its own must go by it")
	c.Nil(source, "a source closed since must be dropped as the popup is opened")

	uxtest.CloseEditorWithoutPrompt(t, screen, sheets[1])
	uxtest.CloseEditorWithoutPrompt(t, screen, calc)
}
