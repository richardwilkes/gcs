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
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/toolbox/tid"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

type weaponsPanel struct {
	unison.Panel
	weaponOwner gurps.WeaponOwner
	allWeapons  *[]*gurps.Weapon
	weapons     []*gurps.Weapon
	provider    TableProvider[*gurps.Weapon]
	table       *unison.Table[*Node[*gurps.Weapon]]
	melee       bool
}

func newWeaponsPanel(cmdRoot Rebuildable, weaponOwner gurps.WeaponOwner, melee bool, weapons *[]*gurps.Weapon) *weaponsPanel {
	p := &weaponsPanel{
		weaponOwner: weaponOwner,
		allWeapons:  weapons,
		weapons:     gurps.ExtractWeaponsOfType(melee, false, *weapons),
		melee:       melee,
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{Columns: 1})
	p.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: align.Fill,
		HGrab:  true,
	})
	p.SetBorder(unison.NewLineBorder(unison.ThemeAboveSurface, 0, unison.NewUniformInsets(1), false))
	p.provider = NewWeaponsProvider(p, p.melee, false)
	p.table = newEditorTable(p.AsPanel(), p.provider)
	var id int
	if melee {
		id = NewMeleeWeaponItemID
		p.table.RefKey = string(tid.MustNewTID(kinds.WeaponMelee))
	} else {
		id = NewRangedWeaponItemID
		p.table.RefKey = string(tid.MustNewTID(kinds.WeaponRanged))
	}
	cmdRoot.AsPanel().InstallCmdHandlers(id, unison.AlwaysEnabled,
		func(_ any) { p.provider.CreateItem(cmdRoot, p.table, NoItemVariant) })
	return p
}

func (p *weaponsPanel) DataOwner() gurps.DataOwner {
	return p.weaponOwner.DataOwner()
}

func (p *weaponsPanel) WeaponOwner() gurps.WeaponOwner {
	return p.weaponOwner
}

func (p *weaponsPanel) Weapons(melee, excludeHidden bool) []*gurps.Weapon {
	return gurps.ExtractWeaponsOfType(melee, excludeHidden, *p.allWeapons)
}

func (p *weaponsPanel) SetWeapons(melee bool, list []*gurps.Weapon) {
	m, r := gurps.SeparateWeapons(false, *p.allWeapons)
	if melee {
		m = list
	} else {
		r = list
	}
	*p.allWeapons = append(append(make([]*gurps.Weapon, 0, len(m)+len(r)), m...), r...)
	sel := p.table.CopySelectionMap()
	p.table.SyncToModel()
	p.table.SetSelectionMap(sel)
}
