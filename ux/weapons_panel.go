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
	"github.com/google/uuid"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/unison"
)

type weaponsPanel struct {
	unison.Panel
	entity      *gurps.Entity
	weaponOwner gurps.WeaponOwner
	weaponType  gurps.WeaponType
	allWeapons  *[]*gurps.Weapon
	weapons     []*gurps.Weapon
	provider    TableProvider[*gurps.Weapon]
	table       *unison.Table[*Node[*gurps.Weapon]]
}

func newWeaponsPanel(cmdRoot Rebuildable, weaponOwner gurps.WeaponOwner, weaponType gurps.WeaponType, weapons *[]*gurps.Weapon) *weaponsPanel {
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
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	p.SetBorder(unison.NewLineBorder(gurps.HeaderColor, 0, unison.NewUniformInsets(1), false))
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, unison.ContentColor.Paint(gc, rect, unison.Fill))
	}
	p.provider = NewWeaponsProvider(p, p.weaponType, false)
	p.table = newEditorTable(p.AsPanel(), p.provider)
	p.table.RefKey = weaponType.Key() + "-" + uuid.New().String()
	var id int
	switch weaponType {
	case gurps.MeleeWeaponType:
		id = NewMeleeWeaponItemID
	case gurps.RangedWeaponType:
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

func (p *weaponsPanel) Weapons(weaponType gurps.WeaponType) []*gurps.Weapon {
	return gurps.ExtractWeaponsOfType(weaponType, *p.allWeapons)
}

func (p *weaponsPanel) SetWeapons(weaponType gurps.WeaponType, list []*gurps.Weapon) {
	melee, ranged := gurps.SeparateWeapons(*p.allWeapons)
	switch weaponType {
	case gurps.MeleeWeaponType:
		melee = list
	case gurps.RangedWeaponType:
		ranged = list
	}
	*p.allWeapons = append(append(make([]*gurps.Weapon, 0, len(melee)+len(ranged)), melee...), ranged...)
	sel := p.table.CopySelectionMap()
	p.table.SyncToModel()
	p.table.SetSelectionMap(sel)
}
