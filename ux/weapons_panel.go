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
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/toolbox/v2/tid"
)

// weaponsPanel shows the melee or the ranged half of an item's weapons. Both halves share the one list in the editor's
// data, so each panel splits it on the way in and merges its half back on the way out.
type weaponsPanel struct {
	editorListPanel[*gurps.Weapon]
	weaponOwner gurps.WeaponOwner
	melee       bool
}

func newWeaponsPanel(cmdRoot Rebuildable, weaponOwner gurps.WeaponOwner, melee bool, weapons *[]*gurps.Weapon) *weaponsPanel {
	p := &weaponsPanel{
		weaponOwner: weaponOwner,
		melee:       melee,
	}
	var id int
	var refKey string
	if melee {
		id = NewMeleeWeaponItemID
		refKey = string(tid.MustNewTID(kinds.WeaponMelee))
	} else {
		id = NewRangedWeaponItemID
		refKey = string(tid.MustNewTID(kinds.WeaponRanged))
	}
	p.init(p, weaponOwner.DataOwner(), weapons, NewWeaponsProvider(p, melee, false), refKey)
	p.attach = func(list []*gurps.Weapon) {
		for _, one := range list {
			one.SetOwner(weaponOwner)
		}
	}
	p.installNewItemHandler(cmdRoot, id, NoItemVariant)
	return p
}

func (p *weaponsPanel) WeaponOwner() gurps.WeaponOwner {
	return p.weaponOwner
}

func (p *weaponsPanel) Weapons(melee, _, excludeHidden bool) []*gurps.Weapon {
	return gurps.ExtractWeaponsOfType(melee, excludeHidden, *p.list)
}

// SetWeapons replaces this panel's kind of weapon in the shared list. The replacements take the places the old list
// gave that kind, so that the other kind stays where it was, and any left over go on the end.
func (p *weaponsPanel) SetWeapons(melee bool, list []*gurps.Weapon) {
	old := *p.list
	merged := make([]*gurps.Weapon, 0, len(old)+len(list))
	next := 0
	for _, w := range old {
		switch {
		case w.IsMelee() != melee:
			merged = append(merged, w)
		case next < len(list):
			merged = append(merged, list[next])
			next++
		}
	}
	p.setList(append(merged, list[next:]...))
}
