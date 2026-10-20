// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps_test

import (
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/gurpstest"
	"github.com/richardwilkes/toolbox/v2/check"
)

// weaponCell is what a table takes from a weapon for one column.
type weaponCell struct {
	Primary   string
	Secondary string
	Tooltip   string
	HasData   bool
}

// weaponDisplay is what a page takes from a weapon: a cell per column and the hash that tells the weapons apart.
type weaponDisplay struct {
	Cells []weaponCell
	Hash  uint64
}

// displayWeapons returns what a page takes from each weapon. With hashFirst set, every hash is found before any cell,
// so that the memos are filled by calls that want no tooltip before the cells ask for theirs.
func displayWeapons(weapons []*gurps.Weapon, columns []int, hashFirst bool) []weaponDisplay {
	result := make([]weaponDisplay, len(weapons))
	if hashFirst {
		for i, w := range weapons {
			result[i].Hash = w.HashDisplayedState()
		}
	}
	for i, w := range weapons {
		for _, col := range columns {
			result[i].Cells = append(result[i].Cells, weaponCellFor(w, col))
		}
		if !hashFirst {
			result[i].Hash = w.HashDisplayedState()
		}
	}
	return result
}

func weaponCellFor(w *gurps.Weapon, col int) weaponCell {
	data := gurps.CellData{ForPage: true}
	w.CellData(col, &data)
	return weaponCell{
		Primary:   data.Primary,
		Secondary: data.Secondary,
		Tooltip:   data.Tooltip,
		HasData:   w.ColumnHasData(col),
	}
}

// newMemoTestEntity returns a large character with every kind of weapon bonus, and the columns a page shows for each
// kind of weapon.
func newMemoTestEntity() (e *gurps.Entity, columns map[bool][]int) {
	size := gurpstest.DefaultLargeEntitySize.Scaled(0.2)
	return gurpstest.NewLargeEntity(&size), map[bool][]int{true: meleePageColumns, false: rangedPageColumns}
}

func TestWeaponMemoMatchesUncached(t *testing.T) {
	c := check.New(t)
	e, columns := newMemoTestEntity()
	for _, melee := range []bool{true, false} {
		weapons := e.Weapons(melee, true, false)
		c.NotEqual(0, len(weapons), "melee %v: the character has weapons", melee)
		restore := gurps.SetWeaponMemosEnabled(false)
		want := displayWeapons(weapons, columns[melee], false)
		restore()
		withModifiers := 0
		for _, one := range want {
			for _, cell := range one.Cells {
				if strings.Contains(cell.Tooltip, "Includes modifiers from:") {
					withModifiers++
				}
			}
		}
		c.NotEqual(0, withModifiers, "melee %v: some tooltips list modifiers", melee)
		for _, hashFirst := range []bool{false, true} {
			e.DiscardCaches()
			c.Equal(want, displayWeapons(weapons, columns[melee], hashFirst), "melee %v, hash first %v: cold", melee,
				hashFirst)
			c.Equal(want, displayWeapons(weapons, columns[melee], hashFirst), "melee %v, hash first %v: warm", melee,
				hashFirst)
		}
	}
}

func TestWeaponTooltipReplayIsByteIdentical(t *testing.T) {
	c := check.New(t)
	e, columns := newMemoTestEntity()
	for _, melee := range []bool{true, false} {
		for _, w := range e.Weapons(melee, true, false) {
			for _, col := range columns[melee] {
				restore := gurps.SetWeaponMemosEnabled(false)
				want := weaponCellFor(w, col).Tooltip
				restore()
				e.DiscardCaches()
				c.Equal(want, weaponCellFor(w, col).Tooltip, "%s, column %d: first call", w, col)
				c.Equal(want, weaponCellFor(w, col).Tooltip, "%s, column %d: second call", w, col)
			}
		}
	}
}

func TestEntityWeaponsCached(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	e.Traits = nil // Drop the natural attacks
	addWeaponTrait := func(name string) *gurps.Weapon {
		trait := gurps.NewTrait(e, nil, false)
		trait.Name = name
		w := gurps.NewWeapon(trait, true)
		w.SetOwner(trait)
		trait.Weapons = []*gurps.Weapon{w}
		e.Traits = append(e.Traits, trait)
		return w
	}
	claws := addWeaponTrait("Claws")
	e.Recalculate()

	first := e.Weapons(true, false, true)
	second := e.Weapons(true, false, true)
	c.Equal([]*gurps.Weapon{claws}, first)
	c.Equal(first, second, "a second call returns the same weapons")
	c.True(&first[0] != &second[0], "each call returns a slice of its own")

	first[0] = nil
	c.Equal(2, len(append(second, claws)))
	c.Equal([]*gurps.Weapon{claws}, e.Weapons(true, false, true), "altering a returned slice leaves the kept list alone")
	c.Equal(0, len(e.Weapons(false, false, true)), "the ranged list is kept apart from the melee one")

	teeth := addWeaponTrait("Teeth")
	c.Equal([]*gurps.Weapon{claws}, e.Weapons(true, false, true), "the list is kept until the caches are discarded")
	e.DiscardCaches()
	c.Equal([]*gurps.Weapon{claws, teeth}, e.Weapons(true, false, true), "the new weapon is listed")

	restore := gurps.SetWeaponMemosEnabled(false)
	defer restore()
	fangs := addWeaponTrait("Fangs")
	c.Equal([]*gurps.Weapon{claws, fangs, teeth}, e.Weapons(true, false, true),
		"with the keeping turned off, a list kept before is not served")
}
