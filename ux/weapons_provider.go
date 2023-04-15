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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

var _ TableProvider[*gurps.Weapon] = &weaponsProvider{}

type weaponsProvider struct {
	table      *unison.Table[*Node[*gurps.Weapon]]
	provider   gurps.WeaponListProvider
	weaponType gurps.WeaponType
	forPage    bool
}

// NewWeaponsProvider creates a new table provider for weapons.
func NewWeaponsProvider(provider gurps.WeaponListProvider, weaponType gurps.WeaponType, forPage bool) TableProvider[*gurps.Weapon] {
	return &weaponsProvider{
		provider:   provider,
		weaponType: weaponType,
		forPage:    forPage,
	}
}

func (p *weaponsProvider) RefKey() string {
	return p.weaponType.Key()
}

func (p *weaponsProvider) AllTags() []string {
	return nil
}

func (p *weaponsProvider) SetTable(table *unison.Table[*Node[*gurps.Weapon]]) {
	p.table = table
}

func (p *weaponsProvider) RootRowCount() int {
	return len(p.provider.Weapons(p.weaponType))
}

func (p *weaponsProvider) RootRows() []*Node[*gurps.Weapon] {
	data := p.provider.Weapons(p.weaponType)
	rows := make([]*Node[*gurps.Weapon], 0, len(data))
	for _, one := range data {
		rows = append(rows, NewNode[*gurps.Weapon](p.table, nil, one, p.forPage))
	}
	return rows
}

func (p *weaponsProvider) SetRootRows(rows []*Node[*gurps.Weapon]) {
	p.provider.SetWeapons(p.weaponType, ExtractNodeDataFromList(rows))
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

func (p *weaponsProvider) DropShouldMoveData(from, to *unison.Table[*Node[*gurps.Weapon]]) bool {
	return from == to
}

func (p *weaponsProvider) ProcessDropData(_, _ *unison.Table[*Node[*gurps.Weapon]]) {
}

func (p *weaponsProvider) AltDropSupport() *AltDropSupport {
	return nil
}

func (p *weaponsProvider) ItemNames() (singular, plural string) {
	return p.weaponType.String(), p.weaponType.AltString()
}

func (p *weaponsProvider) Headers() []unison.TableColumnHeader[*Node[*gurps.Weapon]] {
	ids := p.ColumnIDs()
	headers := make([]unison.TableColumnHeader[*Node[*gurps.Weapon]], 0, len(ids))
	for _, id := range ids {
		switch id {
		case gurps.WeaponDescriptionColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Weapon](p.weaponType.String(), "", p.forPage))
		case gurps.WeaponUsageColumn:
			var title string
			switch {
			case p.forPage:
				title = i18n.Text("Usage")
			case p.weaponType == gurps.MeleeWeaponType:
				title = i18n.Text("Melee Weapon Usage")
			default:
				title = i18n.Text("Ranged Weapon Usage")
			}
			headers = append(headers, NewEditorListHeader[*gurps.Weapon](title, "", p.forPage))
		case gurps.WeaponSLColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Weapon](i18n.Text("SL"), i18n.Text("Skill Level"), p.forPage))
		case gurps.WeaponParryColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Weapon](i18n.Text("Parry"), "", p.forPage))
		case gurps.WeaponBlockColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Weapon](i18n.Text("Block"), "", p.forPage))
		case gurps.WeaponDamageColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Weapon](i18n.Text("Damage"), "", p.forPage))
		case gurps.WeaponReachColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Weapon](i18n.Text("Reach"), "", p.forPage))
		case gurps.WeaponSTColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Weapon](i18n.Text("ST"), i18n.Text("Minimum Strength"), p.forPage))
		case gurps.WeaponAccColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Weapon](i18n.Text("Acc"), i18n.Text("Accuracy Bonus"), p.forPage))
		case gurps.WeaponRangeColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Weapon](i18n.Text("Range"), "", p.forPage))
		case gurps.WeaponRoFColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Weapon](i18n.Text("RoF"), i18n.Text("Rate of Fire"), p.forPage))
		case gurps.WeaponShotsColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Weapon](i18n.Text("Shots"), "", p.forPage))
		case gurps.WeaponBulkColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Weapon](i18n.Text("Bulk"), "", p.forPage))
		case gurps.WeaponRecoilColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Weapon](i18n.Text("Recoil"), "", p.forPage))
		}
	}
	return DisableSorting(headers)
}

func (p *weaponsProvider) SyncHeader(_ []unison.TableColumnHeader[*Node[*gurps.Weapon]]) {
}

func (p *weaponsProvider) ColumnIDs() []int {
	columnIDs := make([]int, 0, 11)
	if p.forPage {
		columnIDs = append(columnIDs, gurps.WeaponDescriptionColumn)
	}
	columnIDs = append(columnIDs,
		gurps.WeaponUsageColumn,
		gurps.WeaponSLColumn,
	)
	switch p.weaponType {
	case gurps.MeleeWeaponType:
		columnIDs = append(columnIDs,
			gurps.WeaponParryColumn,
			gurps.WeaponBlockColumn,
			gurps.WeaponDamageColumn,
			gurps.WeaponReachColumn,
		)
	case gurps.RangedWeaponType:
		columnIDs = append(columnIDs,
			gurps.WeaponAccColumn,
			gurps.WeaponDamageColumn,
			gurps.WeaponRangeColumn,
			gurps.WeaponRoFColumn,
			gurps.WeaponShotsColumn,
			gurps.WeaponBulkColumn,
			gurps.WeaponRecoilColumn,
		)
	}
	return append(columnIDs, gurps.WeaponSTColumn)
}

func (p *weaponsProvider) HierarchyColumnID() int {
	return -1
}

func (p *weaponsProvider) ExcessWidthColumnID() int {
	return gurps.WeaponDescriptionColumn
}

func (p *weaponsProvider) OpenEditor(owner Rebuildable, table *unison.Table[*Node[*gurps.Weapon]]) {
	if !p.forPage {
		OpenEditor[*gurps.Weapon](table, func(item *gurps.Weapon) { EditWeapon(owner, item) })
	}
}

func (p *weaponsProvider) CreateItem(owner Rebuildable, table *unison.Table[*Node[*gurps.Weapon]], _ ItemVariant) {
	if !p.forPage {
		wpn := gurps.NewWeapon(p.provider.WeaponOwner(), p.weaponType)
		InsertItems[*gurps.Weapon](owner, table,
			func() []*gurps.Weapon { return p.provider.Weapons(p.weaponType) },
			func(list []*gurps.Weapon) { p.provider.SetWeapons(p.weaponType, list) },
			func(_ *unison.Table[*Node[*gurps.Weapon]]) []*Node[*gurps.Weapon] { return p.RootRows() },
			wpn)
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

func (p *weaponsProvider) ContextMenuItems() []ContextMenuItem {
	var list []ContextMenuItem
	switch p.weaponType {
	case gurps.MeleeWeaponType:
		list = append(list, ContextMenuItem{i18n.Text("New Melee Weapon"), NewMeleeWeaponItemID})
	case gurps.RangedWeaponType:
		list = append(list, ContextMenuItem{i18n.Text("New Ranged Weapon"), NewRangedWeaponItemID})
	}
	return AppendDefaultContextMenuItems(list)
}
