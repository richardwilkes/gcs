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

package editors

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/weapon"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/gcs/v5/ui/widget/ntable"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/unison"
)

var (
	meleeWeaponColMap = map[int]int{
		0: gurps.WeaponUsageColumn,
		1: gurps.WeaponSLColumn,
		2: gurps.WeaponParryColumn,
		3: gurps.WeaponBlockColumn,
		4: gurps.WeaponDamageColumn,
		5: gurps.WeaponReachColumn,
		6: gurps.WeaponSTColumn,
	}
	meleeWeaponForPageColMap = map[int]int{
		0: gurps.WeaponDescriptionColumn,
		1: gurps.WeaponUsageColumn,
		2: gurps.WeaponSLColumn,
		3: gurps.WeaponParryColumn,
		4: gurps.WeaponBlockColumn,
		5: gurps.WeaponDamageColumn,
		6: gurps.WeaponReachColumn,
		7: gurps.WeaponSTColumn,
	}
	rangedWeaponColMap = map[int]int{
		0: gurps.WeaponUsageColumn,
		1: gurps.WeaponSLColumn,
		2: gurps.WeaponAccColumn,
		3: gurps.WeaponDamageColumn,
		4: gurps.WeaponRangeColumn,
		5: gurps.WeaponRoFColumn,
		6: gurps.WeaponShotsColumn,
		7: gurps.WeaponBulkColumn,
		8: gurps.WeaponRecoilColumn,
		9: gurps.WeaponSTColumn,
	}
	rangedWeaponforPageColMap = map[int]int{
		0:  gurps.WeaponDescriptionColumn,
		1:  gurps.WeaponUsageColumn,
		2:  gurps.WeaponSLColumn,
		3:  gurps.WeaponAccColumn,
		4:  gurps.WeaponDamageColumn,
		5:  gurps.WeaponRangeColumn,
		6:  gurps.WeaponRoFColumn,
		7:  gurps.WeaponShotsColumn,
		8:  gurps.WeaponBulkColumn,
		9:  gurps.WeaponRecoilColumn,
		10: gurps.WeaponSTColumn,
	}
	_ ntable.TableProvider[*gurps.Weapon] = &weaponsProvider{}
)

type weaponsProvider struct {
	table      *unison.Table[*ntable.Node[*gurps.Weapon]]
	colMap     map[int]int
	provider   gurps.WeaponListProvider
	weaponType weapon.Type
	forPage    bool
}

// NewWeaponsProvider creates a new table provider for weapons.
func NewWeaponsProvider(provider gurps.WeaponListProvider, weaponType weapon.Type, forPage bool) ntable.TableProvider[*gurps.Weapon] {
	p := &weaponsProvider{
		provider:   provider,
		weaponType: weaponType,
		forPage:    forPage,
	}
	switch {
	case weaponType == weapon.Melee && forPage:
		p.colMap = meleeWeaponForPageColMap
	case weaponType == weapon.Melee:
		p.colMap = meleeWeaponColMap
	case weaponType == weapon.Ranged && forPage:
		p.colMap = rangedWeaponforPageColMap
	case weaponType == weapon.Ranged:
		p.colMap = rangedWeaponColMap
	default:
		jot.Fatalf(1, "unknown weapon type: %d", weaponType)
	}
	return p
}

func (p *weaponsProvider) SetTable(table *unison.Table[*ntable.Node[*gurps.Weapon]]) {
	p.table = table
}

func (p *weaponsProvider) RootRowCount() int {
	return len(p.provider.Weapons(p.weaponType))
}

func (p *weaponsProvider) RootRows() []*ntable.Node[*gurps.Weapon] {
	data := p.provider.Weapons(p.weaponType)
	rows := make([]*ntable.Node[*gurps.Weapon], 0, len(data))
	for _, one := range data {
		rows = append(rows, ntable.NewNode[*gurps.Weapon](p.table, nil, p.colMap, one, p.forPage))
	}
	return rows
}

func (p *weaponsProvider) SetRootRows(rows []*ntable.Node[*gurps.Weapon]) {
	p.provider.SetWeapons(p.weaponType, ntable.ExtractNodeDataFromList(rows))
}

func (p *weaponsProvider) RootData() []*gurps.Weapon {
	return p.provider.Weapons(p.weaponType)
}

func (p *weaponsProvider) SetRootData(data []*gurps.Weapon) {
	p.provider.SetWeapons(p.weaponType, data)
}

func (p *weaponsProvider) Entity() *gurps.Entity {
	return p.provider.Entity()
}

func (p *weaponsProvider) DragKey() string {
	return p.weaponType.Key()
}

func (p *weaponsProvider) DragSVG() *unison.SVG {
	return p.weaponType.SVG()
}

func (p *weaponsProvider) DropShouldMoveData(from, to *unison.Table[*ntable.Node[*gurps.Weapon]]) bool {
	return from == to
}

func (p *weaponsProvider) ItemNames() (singular, plural string) {
	return p.weaponType.String(), p.weaponType.AltString()
}

func (p *weaponsProvider) Headers() []unison.TableColumnHeader[*ntable.Node[*gurps.Weapon]] {
	var headers []unison.TableColumnHeader[*ntable.Node[*gurps.Weapon]]
	for i := 0; i < len(p.colMap); i++ {
		switch p.colMap[i] {
		case gurps.WeaponDescriptionColumn:
			headers = append(headers, NewHeader[*gurps.Weapon](p.weaponType.String(), "", p.forPage))
		case gurps.WeaponUsageColumn:
			var title string
			switch {
			case p.forPage:
				title = i18n.Text("Usage")
			case p.weaponType == weapon.Melee:
				title = i18n.Text("Melee Weapon Usage")
			default:
				title = i18n.Text("Ranged Weapon Usage")
			}
			headers = append(headers, NewHeader[*gurps.Weapon](title, "", p.forPage))
		case gurps.WeaponSLColumn:
			headers = append(headers, NewHeader[*gurps.Weapon](i18n.Text("SL"), i18n.Text("Skill Level"), p.forPage))
		case gurps.WeaponParryColumn:
			headers = append(headers, NewHeader[*gurps.Weapon](i18n.Text("Parry"), "", p.forPage))
		case gurps.WeaponBlockColumn:
			headers = append(headers, NewHeader[*gurps.Weapon](i18n.Text("Block"), "", p.forPage))
		case gurps.WeaponDamageColumn:
			headers = append(headers, NewHeader[*gurps.Weapon](i18n.Text("Damage"), "", p.forPage))
		case gurps.WeaponReachColumn:
			headers = append(headers, NewHeader[*gurps.Weapon](i18n.Text("Reach"), "", p.forPage))
		case gurps.WeaponSTColumn:
			headers = append(headers, NewHeader[*gurps.Weapon](i18n.Text("ST"), i18n.Text("Minimum Strength"), p.forPage))
		case gurps.WeaponAccColumn:
			headers = append(headers, NewHeader[*gurps.Weapon](i18n.Text("Acc"), i18n.Text("Accuracy Bonus"), p.forPage))
		case gurps.WeaponRangeColumn:
			headers = append(headers, NewHeader[*gurps.Weapon](i18n.Text("Range"), "", p.forPage))
		case gurps.WeaponRoFColumn:
			headers = append(headers, NewHeader[*gurps.Weapon](i18n.Text("RoF"), i18n.Text("Rate of Fire"), p.forPage))
		case gurps.WeaponShotsColumn:
			headers = append(headers, NewHeader[*gurps.Weapon](i18n.Text("Shots"), "", p.forPage))
		case gurps.WeaponBulkColumn:
			headers = append(headers, NewHeader[*gurps.Weapon](i18n.Text("Bulk"), "", p.forPage))
		case gurps.WeaponRecoilColumn:
			headers = append(headers, NewHeader[*gurps.Weapon](i18n.Text("Recoil"), "", p.forPage))
		default:
			jot.Fatalf(1, "invalid weapon column: %d", p.colMap[i])
		}
	}
	return headers
}

func (p *weaponsProvider) SyncHeader(_ []unison.TableColumnHeader[*ntable.Node[*gurps.Weapon]]) {
}

func (p *weaponsProvider) HierarchyColumnIndex() int {
	return -1
}

func (p *weaponsProvider) ExcessWidthColumnIndex() int {
	for k, v := range p.colMap {
		if v == gurps.WeaponDescriptionColumn {
			return k
		}
	}
	return 0
}

func (p *weaponsProvider) OpenEditor(owner widget.Rebuildable, table *unison.Table[*ntable.Node[*gurps.Weapon]]) {
	if !p.forPage {
		ntable.OpenEditor[*gurps.Weapon](table, func(item *gurps.Weapon) { EditWeapon(owner, item) })
	}
}

func (p *weaponsProvider) CreateItem(owner widget.Rebuildable, table *unison.Table[*ntable.Node[*gurps.Weapon]], _ ntable.ItemVariant) {
	if !p.forPage {
		wpn := gurps.NewWeapon(p.provider.WeaponOwner(), p.weaponType)
		ntable.InsertItem[*gurps.Weapon](owner, table, wpn,
			func() []*gurps.Weapon { return p.provider.Weapons(p.weaponType) },
			func(list []*gurps.Weapon) { p.provider.SetWeapons(p.weaponType, list) },
			func(_ *unison.Table[*ntable.Node[*gurps.Weapon]]) []*ntable.Node[*gurps.Weapon] { return p.RootRows() })
		EditWeapon(owner, wpn)
	}
}

func (p *weaponsProvider) Serialize() ([]byte, error) {
	if p.forPage {
		return nil, errs.New("not allowed")
	}
	return jio.SerializeAndCompress(p.provider.Weapons(p.weaponType))
}

func (p *weaponsProvider) Deserialize(data []byte) error {
	if p.forPage {
		return errs.New("not allowed")
	}
	var rows []*gurps.Weapon
	if err := jio.DecompressAndDeserialize(data, &rows); err != nil {
		return err
	}
	p.provider.SetWeapons(p.weaponType, rows)
	return nil
}
