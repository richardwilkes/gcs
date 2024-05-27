// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"github.com/google/uuid"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wpn"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

type weaponsPanel struct {
	unison.Panel
	entity      *gurps.Entity
	weaponOwner gurps.WeaponOwner
	weaponType  wpn.Type
	allWeapons  *[]*gurps.Weapon
	weapons     []*gurps.Weapon
	provider    TableProvider[*gurps.Weapon]
	table       *unison.Table[*Node[*gurps.Weapon]]
}

func newWeaponsPanel(cmdRoot Rebuildable, weaponOwner gurps.WeaponOwner, weaponType wpn.Type, weapons *[]*gurps.Weapon) *weaponsPanel {
	p := &weaponsPanel{
		weaponOwner: weaponOwner,
		weaponType:  weaponType,
		allWeapons:  weapons,
		weapons:     gurps.ExtractWeaponsOfType(weaponType, *weapons),
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{Columns: 1})
	p.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: align.Fill,
		HGrab:  true,
	})
	p.SetBorder(unison.NewLineBorder(unison.ThemeAboveSurface, 0, unison.NewUniformInsets(1), false))
	p.provider = NewWeaponsProvider(p, p.weaponType, false)
	p.table = newEditorTable(p.AsPanel(), p.provider)
	p.table.RefKey = weaponType.Key() + "-" + uuid.New().String()
	var id int
	switch weaponType {
	case wpn.Melee:
		id = NewMeleeWeaponItemID
	case wpn.Ranged:
		id = NewRangedWeaponItemID
	default:
		return p
	}
	cmdRoot.AsPanel().InstallCmdHandlers(id, unison.AlwaysEnabled,
		func(_ any) { p.provider.CreateItem(cmdRoot, p.table, NoItemVariant) })
	return p
}

func (p *weaponsPanel) Entity() *gurps.Entity {
	return p.entity
}

func (p *weaponsPanel) WeaponOwner() gurps.WeaponOwner {
	return p.weaponOwner
}

func (p *weaponsPanel) Weapons(weaponType wpn.Type) []*gurps.Weapon {
	return gurps.ExtractWeaponsOfType(weaponType, *p.allWeapons)
}

func (p *weaponsPanel) SetWeapons(weaponType wpn.Type, list []*gurps.Weapon) {
	melee, ranged := gurps.SeparateWeapons(*p.allWeapons)
	switch weaponType {
	case wpn.Melee:
		melee = list
	case wpn.Ranged:
		ranged = list
	default:
	}
	*p.allWeapons = append(append(make([]*gurps.Weapon, 0, len(melee)+len(ranged)), melee...), ranged...)
	sel := p.table.CopySelectionMap()
	p.table.SyncToModel()
	p.table.SetSelectionMap(sel)
}
