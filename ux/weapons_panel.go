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

// weaponsPanel shows the melee or the ranged half of the weapons an item or a modifier holds. Both halves share the one
// list in the editor's data, so each panel splits it on the way in and merges its half back on the way out. A
// modifier's weapons belong to the item it modifies, so the holder says what they belong to.
type weaponsPanel struct {
	editorListPanel[*gurps.Weapon]
	holder gurps.WeaponHolder
	melee  bool
}

func newWeaponsPanel(cmdRoot Rebuildable, holder gurps.WeaponHolder, melee bool, weapons *[]*gurps.Weapon) *weaponsPanel {
	p := &weaponsPanel{
		holder: holder,
		melee:  melee,
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
	p.init(p, cmdRoot, holder.DataOwner(), weapons, NewWeaponsProvider(p, melee, false), refKey)
	p.attach = holder.AttachWeapons
	p.installNewItemHandler(cmdRoot, id, NoItemVariant)
	return p
}

// WeaponOwner implements gurps.WeaponListProvider. It is nil for the weapons of a modifier that modifies nothing.
func (p *weaponsPanel) WeaponOwner() gurps.WeaponOwner {
	return p.holder.WeaponOwner()
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
