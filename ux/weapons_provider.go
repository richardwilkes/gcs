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
	"log/slog"
	"reflect"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/unison"
)

const (
	meleeWeaponRefKey  = "melee_weapon"
	rangedWeaponRefKey = "ranged_weapon"
)

var _ TableProvider[*gurps.Weapon] = &weaponsProvider{}

type weaponsProvider struct {
	listProvider[*gurps.Weapon]
	provider gurps.WeaponListProvider
	melee    bool
}

// NewWeaponsProvider creates a new table provider for weapons.
func NewWeaponsProvider(provider gurps.WeaponListProvider, melee, forPage bool) TableProvider[*gurps.Weapon] {
	p := &weaponsProvider{provider: provider, melee: melee}
	p.listProvider = listProvider[*gurps.Weapon]{
		dataOwner: provider,
		list:      func() []*gurps.Weapon { return provider.Weapons(melee, p.showAllWeapons(), forPage) },
		setList:   func(list []*gurps.Weapon) { provider.SetWeapons(melee, list) },
		columnIDs: p.ColumnIDs,
		headerData: func(columnID int) gurps.HeaderData {
			return gurps.WeaponHeaderData(columnID, melee, forPage)
		},
		forPage: forPage,
		reuse:   pageRowReuse(forPage),
	}
	return p
}

func (p *weaponsProvider) RefKey() string {
	if p.melee {
		return meleeWeaponRefKey
	}
	return rangedWeaponRefKey
}

// showAllWeapons reports whether a page lists every weapon rather than only those of carried, equipped items. Off a
// page there is no such choice: an item's editor always shows all of its weapons.
func (p *weaponsProvider) showAllWeapons() bool {
	settings := p.pageSheetSettings()
	return settings != nil && settings.ShowAllWeapons
}

// hideUnusedColumns reports whether a page drops the columns none of its weapons have data for.
func (p *weaponsProvider) hideUnusedColumns() bool {
	settings := p.pageSheetSettings()
	return settings != nil && settings.HideUnusedWeaponColumns
}

func (p *weaponsProvider) DragKey() *uti.DataType {
	if p.melee {
		return meleeWeaponDragKey
	}
	return rangedWeaponDragKey
}

func (p *weaponsProvider) DragSVG() *unison.SVG {
	return weaponSVG(p.melee)
}

// weaponSVG returns the SVG that should be used for the weapon type.
func weaponSVG(melee bool) *unison.SVG {
	if melee {
		return svg.MeleeWeapon
	}
	return svg.RangedWeapon
}

func (p *weaponsProvider) ItemNames() (singular, plural string) {
	if p.melee {
		return i18n.Text("Melee Attack"), i18n.Text("Melee Attacks")
	}
	return i18n.Text("Ranged Attack"), i18n.Text("Ranged Attacks")
}

func (p *weaponsProvider) Headers() []unison.TableColumnHeader[*Node[*gurps.Weapon]] {
	return DisableSorting(p.listProvider.Headers())
}

func (p *weaponsProvider) ColumnIDs() []int {
	columnIDs := make([]int, 0, 11)
	if p.forPage {
		columnIDs = append(columnIDs, gurps.WeaponDescriptionColumn)
	} else {
		columnIDs = append(columnIDs, gurps.WeaponHideColumn)
	}
	columnIDs = append(
		columnIDs,
		gurps.WeaponUsageColumn,
		gurps.WeaponSLColumn,
	)
	if p.melee {
		columnIDs = append(
			columnIDs,
			gurps.WeaponParryColumn,
			gurps.WeaponBlockColumn,
			gurps.WeaponDamageColumn,
			gurps.WeaponReachColumn,
		)
	} else {
		columnIDs = append(
			columnIDs,
			gurps.WeaponAccColumn,
			gurps.WeaponDamageColumn,
			gurps.WeaponRangeColumn,
			gurps.WeaponRoFColumn,
			gurps.WeaponShotsColumn,
			gurps.WeaponBulkColumn,
			gurps.WeaponRecoilColumn,
		)
	}
	columnIDs = append(columnIDs, gurps.WeaponSTColumn)
	if p.hideUnusedColumns() {
		columnIDs = p.removeUnusedColumns(columnIDs)
	}
	return columnIDs
}

// removeUnusedColumns removes any columns from the provided list that have no meaningful data in any of the weapons
// that will be displayed.
func (p *weaponsProvider) removeUnusedColumns(columnIDs []int) []int {
	weapons := p.RootData()
	if len(weapons) == 0 {
		return columnIDs
	}
	filtered := columnIDs[:0]
	for _, id := range columnIDs {
		used := false
		for _, w := range weapons {
			if w.ColumnHasData(id) {
				used = true
				break
			}
		}
		if used {
			filtered = append(filtered, id)
		}
	}
	return filtered
}

func (p *weaponsProvider) HierarchyColumnID() int {
	return -1
}

func (p *weaponsProvider) ExcessWidthColumnID() int {
	if p.forPage {
		return gurps.WeaponDescriptionColumn
	}
	return gurps.WeaponUsageColumn
}

func (p *weaponsProvider) OpenEditor(owner Rebuildable, table *unison.Table[*Node[*gurps.Weapon]]) {
	if p.forPage {
		OpenEditor(table, func(item *gurps.Weapon) {
			if sheet := unison.Ancestor[*Sheet](table.AsPanel()); sheet != nil {
				switch itemOwner := item.Owner.(type) {
				case *gurps.Equipment:
					if !openWeaponEditor(sheet, sheet.CarriedEquipment, item, itemOwner,
						func(owner Rebuildable, eq *gurps.Equipment) *editor[*gurps.Equipment, *gurps.EquipmentEditData] {
							return EditEquipment(owner, eq, true)
						}) {
						openWeaponEditor(sheet, sheet.OtherEquipment, item, itemOwner,
							func(owner Rebuildable, eq *gurps.Equipment) *editor[*gurps.Equipment, *gurps.EquipmentEditData] {
								return EditEquipment(owner, eq, false)
							})
					}
				case *gurps.Skill:
					openWeaponEditor(sheet, sheet.Skills, item, itemOwner, EditSkill)
				case *gurps.Spell:
					openWeaponEditor(sheet, sheet.Spells, item, itemOwner, EditSpell)
				case *gurps.Trait:
					openWeaponEditor(sheet, sheet.Traits, item, itemOwner, EditTrait)
				default:
					// Should not happen.
					slog.Warn("unexpected weapon owner type", "type", reflect.TypeOf(item.Owner).String())
				}
			}
		})
	} else {
		OpenEditor(table, func(item *gurps.Weapon) { EditWeapon(owner, item) })
	}
}

// openWeaponEditor opens the editor of the item on the sheet's list that owns the weapon and, from that, the editor of
// the item editor's copy of the weapon, going through the editor of the item editor's copy of the modifier that holds
// the weapon, if one does. It returns false if the list doesn't hold the item.
func openWeaponEditor[T gurps.Node[T], D gurps.EditorData[T]](sheet *Sheet, pageList *PageList[T], item *gurps.Weapon, itemOwner T, editFunc func(Rebuildable, T) *editor[T, D]) bool {
	node := findRow(pageList.Table.RootRows(), func(data T) bool { return data == itemOwner })
	if node == nil {
		return false
	}
	pageList.Table.DiscloseRow(node, false)
	pageList.Table.ClearSelection()
	pageList.Table.SelectByIndex(pageList.Table.RowToIndex(node))
	OpenEditor(pageList.Table, func(data T) {
		if editor := editFunc(sheet, data); editor != nil {
			switch mod := item.HoldingModifier().(type) {
			case *gurps.TraitModifier:
				openModifierWeaponEditor(editor, mod.ID(), item, EditTraitModifier)
			case *gurps.EquipmentModifier:
				openModifierWeaponEditor(editor, mod.ID(), item, EditEquipmentModifier)
			default:
				openWeaponInEditor(editor, item)
			}
		}
	})
	return true
}

// openModifierWeaponEditor opens, from the item's editor, the editor of its copy of the modifier with the given ID and,
// from that, the editor of the modifier editor's copy of the weapon. The copies keep the IDs of what they copy, so the
// modifier is found in the item editor's list of modifiers by its ID.
func openModifierWeaponEditor[N gurps.Node[N], DN gurps.EditorData[N], M gurps.Node[M], DM gurps.EditorData[M]](itemEditor *editor[N, DN], id tid.TID, item *gurps.Weapon, edit func(Rebuildable, M) *editor[M, DM]) {
	for _, child := range itemEditor.content.Children() {
		list, ok := child.Self.(interface {
			listTable() *unison.Table[*Node[M]]
		})
		if !ok {
			continue
		}
		table := list.listTable()
		if node := findRow(table.RootRows(), func(data M) bool { return data.ID() == id }); node != nil {
			table.DiscloseRow(node, false)
			table.ClearSelection()
			table.SelectByIndex(table.RowToIndex(node))
			OpenEditor(table, func(m M) {
				if modEditor := edit(itemEditor, m); modEditor != nil {
					openWeaponInEditor(modEditor, item)
				}
			})
		}
		return
	}
}

// openWeaponInEditor selects the editor's copy of the weapon, which has the weapon's ID, in the editor's list of that
// kind of weapon and opens the copy's editor. It does nothing if the editor has no such list or the list no such copy.
func openWeaponInEditor[N gurps.Node[N], D gurps.EditorData[N]](e *editor[N, D], item *gurps.Weapon) {
	p := e.rangedWeapons
	if item.IsMelee() {
		p = e.meleeWeapons
	}
	if p == nil {
		return
	}
	wt := p.table
	if row := findRow(wt.RootRows(), func(w *gurps.Weapon) bool { return w.TID == item.TID }); row != nil {
		wt.ClearSelection()
		wt.SelectByIndex(wt.RowToIndex(row))
		OpenEditor(wt, func(w *gurps.Weapon) { EditWeapon(e, w) })
	}
}

// findRow returns the first row whose data match accepts, looking through the rows and, at any depth, the rows within
// them, those of closed containers included, or nil if there is none.
func findRow[T gurps.Node[T]](rows []*Node[T], match func(T) bool) *Node[T] {
	for _, row := range rows {
		if match(row.data) {
			return row
		}
		if row.CanHaveChildren() {
			if found := findRow(row.Children(), match); found != nil {
				return found
			}
		}
	}
	return nil
}

func (p *weaponsProvider) CreateItem(owner Rebuildable, table *unison.Table[*Node[*gurps.Weapon]], _ ItemVariant) {
	if !p.forPage {
		w := gurps.NewWeapon(p.provider.WeaponOwner(), p.melee)
		p.insertItems(owner, table, w)
		EditWeapon(owner, w)
	}
}

func (p *weaponsProvider) Serialize() ([]byte, error) {
	if p.forPage {
		return nil, errs.New("not allowed")
	}
	return p.listProvider.Serialize()
}

func (p *weaponsProvider) Deserialize(data []byte) error {
	if p.forPage {
		return errs.New("not allowed")
	}
	return p.listProvider.Deserialize(data)
}

func (p *weaponsProvider) ContextMenuItems() []ContextMenuItem {
	var list []ContextMenuItem
	if p.melee {
		list = append(list, contextMenuItemFor(newMeleeWeaponAction))
	} else {
		list = append(list, contextMenuItemFor(newRangedWeaponAction))
	}
	return AppendDefaultContextMenuItems(list)
}
