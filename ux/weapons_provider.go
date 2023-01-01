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
	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

var _ TableProvider[*model.Weapon] = &weaponsProvider{}

type weaponsProvider struct {
	table      *unison.Table[*Node[*model.Weapon]]
	provider   model.WeaponListProvider
	weaponType model.WeaponType
	forPage    bool
}

// NewWeaponsProvider creates a new table provider for weapons.
func NewWeaponsProvider(provider model.WeaponListProvider, weaponType model.WeaponType, forPage bool) TableProvider[*model.Weapon] {
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

func (p *weaponsProvider) SetTable(table *unison.Table[*Node[*model.Weapon]]) {
	p.table = table
}

func (p *weaponsProvider) RootRowCount() int {
	return len(p.provider.Weapons(p.weaponType))
}

func (p *weaponsProvider) RootRows() []*Node[*model.Weapon] {
	data := p.provider.Weapons(p.weaponType)
	rows := make([]*Node[*model.Weapon], 0, len(data))
	for _, one := range data {
		rows = append(rows, NewNode[*model.Weapon](p.table, nil, one, p.forPage))
	}
	return rows
}

func (p *weaponsProvider) SetRootRows(rows []*Node[*model.Weapon]) {
	p.provider.SetWeapons(p.weaponType, ExtractNodeDataFromList(rows))
}

func (p *weaponsProvider) RootData() []*model.Weapon {
	return p.provider.Weapons(p.weaponType)
}

func (p *weaponsProvider) SetRootData(data []*model.Weapon) {
	p.provider.SetWeapons(p.weaponType, data)
}

func (p *weaponsProvider) Entity() *model.Entity {
	return p.provider.Entity()
}

func (p *weaponsProvider) DragKey() string {
	return p.weaponType.Key()
}

func (p *weaponsProvider) DragSVG() *unison.SVG {
	return p.weaponType.SVG()
}

func (p *weaponsProvider) DropShouldMoveData(from, to *unison.Table[*Node[*model.Weapon]]) bool {
	return from == to
}

func (p *weaponsProvider) ProcessDropData(_, _ *unison.Table[*Node[*model.Weapon]]) {
}

func (p *weaponsProvider) AltDropSupport() *AltDropSupport {
	return nil
}

func (p *weaponsProvider) ItemNames() (singular, plural string) {
	return p.weaponType.String(), p.weaponType.AltString()
}

func (p *weaponsProvider) Headers() []unison.TableColumnHeader[*Node[*model.Weapon]] {
	ids := p.ColumnIDs()
	headers := make([]unison.TableColumnHeader[*Node[*model.Weapon]], 0, len(ids))
	for _, id := range ids {
		switch id {
		case model.WeaponDescriptionColumn:
			headers = append(headers, NewEditorListHeader[*model.Weapon](p.weaponType.String(), "", p.forPage))
		case model.WeaponUsageColumn:
			var title string
			switch {
			case p.forPage:
				title = i18n.Text("Usage")
			case p.weaponType == model.MeleeWeaponType:
				title = i18n.Text("Melee Weapon Usage")
			default:
				title = i18n.Text("Ranged Weapon Usage")
			}
			headers = append(headers, NewEditorListHeader[*model.Weapon](title, "", p.forPage))
		case model.WeaponSLColumn:
			headers = append(headers, NewEditorListHeader[*model.Weapon](i18n.Text("SL"), i18n.Text("Skill Level"), p.forPage))
		case model.WeaponParryColumn:
			headers = append(headers, NewEditorListHeader[*model.Weapon](i18n.Text("Parry"), "", p.forPage))
		case model.WeaponBlockColumn:
			headers = append(headers, NewEditorListHeader[*model.Weapon](i18n.Text("Block"), "", p.forPage))
		case model.WeaponDamageColumn:
			headers = append(headers, NewEditorListHeader[*model.Weapon](i18n.Text("Damage"), "", p.forPage))
		case model.WeaponReachColumn:
			headers = append(headers, NewEditorListHeader[*model.Weapon](i18n.Text("Reach"), "", p.forPage))
		case model.WeaponSTColumn:
			headers = append(headers, NewEditorListHeader[*model.Weapon](i18n.Text("ST"), i18n.Text("Minimum Strength"), p.forPage))
		case model.WeaponAccColumn:
			headers = append(headers, NewEditorListHeader[*model.Weapon](i18n.Text("Acc"), i18n.Text("Accuracy Bonus"), p.forPage))
		case model.WeaponRangeColumn:
			headers = append(headers, NewEditorListHeader[*model.Weapon](i18n.Text("Range"), "", p.forPage))
		case model.WeaponRoFColumn:
			headers = append(headers, NewEditorListHeader[*model.Weapon](i18n.Text("RoF"), i18n.Text("Rate of Fire"), p.forPage))
		case model.WeaponShotsColumn:
			headers = append(headers, NewEditorListHeader[*model.Weapon](i18n.Text("Shots"), "", p.forPage))
		case model.WeaponBulkColumn:
			headers = append(headers, NewEditorListHeader[*model.Weapon](i18n.Text("Bulk"), "", p.forPage))
		case model.WeaponRecoilColumn:
			headers = append(headers, NewEditorListHeader[*model.Weapon](i18n.Text("Recoil"), "", p.forPage))
		}
	}
	return headers
}

func (p *weaponsProvider) SyncHeader(_ []unison.TableColumnHeader[*Node[*model.Weapon]]) {
}

func (p *weaponsProvider) ColumnIDs() []int {
	columnIDs := make([]int, 0, 11)
	if p.forPage {
		columnIDs = append(columnIDs, model.WeaponDescriptionColumn)
	}
	columnIDs = append(columnIDs,
		model.WeaponUsageColumn,
		model.WeaponSLColumn,
	)
	switch p.weaponType {
	case model.MeleeWeaponType:
		columnIDs = append(columnIDs,
			model.WeaponParryColumn,
			model.WeaponBlockColumn,
			model.WeaponDamageColumn,
			model.WeaponReachColumn,
		)
	case model.RangedWeaponType:
		columnIDs = append(columnIDs,
			model.WeaponAccColumn,
			model.WeaponDamageColumn,
			model.WeaponRangeColumn,
			model.WeaponRoFColumn,
			model.WeaponShotsColumn,
			model.WeaponBulkColumn,
			model.WeaponRecoilColumn,
		)
	}
	return append(columnIDs, model.WeaponSTColumn)
}

func (p *weaponsProvider) HierarchyColumnID() int {
	return -1
}

func (p *weaponsProvider) ExcessWidthColumnID() int {
	return model.WeaponDescriptionColumn
}

func (p *weaponsProvider) OpenEditor(owner Rebuildable, table *unison.Table[*Node[*model.Weapon]]) {
	if !p.forPage {
		OpenEditor[*model.Weapon](table, func(item *model.Weapon) { EditWeapon(owner, item) })
	}
}

func (p *weaponsProvider) CreateItem(owner Rebuildable, table *unison.Table[*Node[*model.Weapon]], _ ItemVariant) {
	if !p.forPage {
		wpn := model.NewWeapon(p.provider.WeaponOwner(), p.weaponType)
		InsertItems[*model.Weapon](owner, table,
			func() []*model.Weapon { return p.provider.Weapons(p.weaponType) },
			func(list []*model.Weapon) { p.provider.SetWeapons(p.weaponType, list) },
			func(_ *unison.Table[*Node[*model.Weapon]]) []*Node[*model.Weapon] { return p.RootRows() },
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
	var rows []*model.Weapon
	if err := jio.DecompressAndDeserialize(data, &rows); err != nil {
		return err
	}
	p.provider.SetWeapons(p.weaponType, rows)
	return nil
}

func (p *weaponsProvider) ContextMenuItems() []ContextMenuItem {
	var list []ContextMenuItem
	switch p.weaponType {
	case model.MeleeWeaponType:
		list = append(list, ContextMenuItem{i18n.Text("New Melee Weapon"), NewMeleeWeaponItemID})
	case model.RangedWeaponType:
		list = append(list, ContextMenuItem{i18n.Text("New Ranged Weapon"), NewRangedWeaponItemID})
	}
	return AppendDefaultContextMenuItems(list)
}
