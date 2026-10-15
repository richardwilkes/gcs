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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

// listFilterKind is a list type that can have saved filters: the key they are stored under, what the list type is
// called, and the fields a filter for it may test.
type listFilterKind struct {
	key    string
	title  string
	fields func() []filterFieldInfo
}

// listFilterKinds returns the list types that can have saved filters, in the order the Saved Filters view shows them.
func listFilterKinds() []listFilterKind {
	return []listFilterKind{
		newListFilterKind(gurps.TraitsExt, i18n.Text("Traits"), gurps.TraitFilterFields),
		newListFilterKind(gurps.TraitModifiersExt, i18n.Text("Trait Modifiers"), gurps.TraitModifierFilterFields),
		newListFilterKind(gurps.SkillsExt, i18n.Text("Skills"), gurps.SkillFilterFields),
		newListFilterKind(gurps.SpellsExt, i18n.Text("Spells"), gurps.SpellFilterFields),
		newListFilterKind(gurps.EquipmentExt, i18n.Text("Equipment"), gurps.EquipmentFilterFields),
		newListFilterKind(gurps.EquipmentModifiersExt, i18n.Text("Equipment Modifiers"),
			gurps.EquipmentModifierFilterFields),
		newListFilterKind(gurps.NotesExt, i18n.Text("Notes"), gurps.NoteFilterFields),
	}
}

// newListFilterKind returns the list type of the files with the extension. Its fields have no list to draw
// suggestions from.
func newListFilterKind[T gurps.Node[T]](ext, title string, fields func() []*gurps.FilterField[T]) listFilterKind {
	return listFilterKind{
		key:    gurps.ListFilterKeyForExtension(ext),
		title:  title,
		fields: func() []filterFieldInfo { return filterFieldInfos(fields(), nil) },
	}
}

// listFilterRows builds the rows of the Saved Filters view: one per list type, each followed by one per filter. Each
// list type and filter keeps its row ID from one build to the next, so that the selection survives a rebuild.
type listFilterRows struct {
	ids   map[any]tid.TID
	kinds map[*gurps.ListFilter]listFilterKind
}

// build returns the rows for the saved filters as they now stand.
func (r *listFilterRows) build() []*listFilterRow {
	ids := r.ids
	r.ids = make(map[any]tid.TID)
	idFor := func(key any) tid.TID {
		id, ok := ids[key]
		if !ok {
			id = tid.MustNewTID(kinds.ListFilter)
		}
		r.ids[key] = id
		return id
	}
	r.kinds = make(map[*gurps.ListFilter]listFilterKind)
	var rows []*listFilterRow
	for _, kind := range listFilterKinds() {
		filters := gurps.GlobalSettings().ListFilters[kind.key]
		var summary string
		switch len(filters) {
		case 0:
			summary = i18n.Text("No saved filters")
		case 1:
			summary = i18n.Text("1 filter")
		default:
			summary = i18n.Text("%d filters", len(filters))
		}
		rows = append(rows, &listFilterRow{id: idFor(kind.key), key: kind.key, name: kind.title, summary: summary})
		fields := kind.fields()
		for _, f := range filters {
			r.kinds[f] = kind
			rows = append(rows, &listFilterRow{
				id:      idFor(f),
				filter:  f,
				name:    f.Name,
				summary: summarizeListFilter(f, fields),
			})
		}
	}
	return rows
}

// kindOf returns the list type of a filter shown in the last build.
func (r *listFilterRows) kindOf(f *gurps.ListFilter) listFilterKind {
	return r.kinds[f]
}

// selectionFor returns the table selection that selects the rows of the filters.
func (r *listFilterRows) selectionFor(filters []*gurps.ListFilter) map[tid.TID]bool {
	sel := make(map[tid.TID]bool, len(filters))
	for _, f := range filters {
		sel[r.ids[f]] = true
	}
	return sel
}

// listFilterRow is a row of the Saved Filters view: the list type with the key, or one of its filters when filter is
// set.
type listFilterRow struct {
	id      tid.TID
	key     string
	filter  *gurps.ListFilter
	name    string
	summary string
}

// CloneForTarget implements unison.TableRowData. Rows can't be dragged, so it is never called.
func (r *listFilterRow) CloneForTarget(_ unison.Paneler, _ *listFilterRow) *listFilterRow {
	return nil
}

// ID implements unison.TableRowData.
func (r *listFilterRow) ID() tid.TID {
	return r.id
}

// Parent implements unison.TableRowData.
func (r *listFilterRow) Parent() *listFilterRow {
	return nil
}

// SetParent implements unison.TableRowData. The rows are flat, so it does nothing.
func (r *listFilterRow) SetParent(_ *listFilterRow) {}

// CanHaveChildren implements unison.TableRowData.
func (r *listFilterRow) CanHaveChildren() bool {
	return false
}

// Children implements unison.TableRowData.
func (r *listFilterRow) Children() []*listFilterRow {
	return nil
}

// SetChildren implements unison.TableRowData. The rows are flat, so it does nothing.
func (r *listFilterRow) SetChildren(_ []*listFilterRow) {}

// CellDataForSort implements unison.TableRowData.
func (r *listFilterRow) CellDataForSort(col int) string {
	if col == 0 {
		return r.name
	}
	return r.summary
}

// ColumnCell implements unison.TableRowData. A list type's name is bold, a filter's is indented beneath it, and a
// filter's summary wraps to the width of its column.
func (r *listFilterRow) ColumnCell(_, col int, foreground, _ unison.Ink, _, _, _ bool) unison.Paneler {
	label := NewWrappingLabel()
	label.setText(r.CellDataForSort(col), foreground)
	switch {
	case r.filter == nil && col == 0:
		label.Font = unison.EmphasizedSystemFont
	case r.filter != nil && col == 0:
		label.SetBorder(unison.NewEmptyBorder(geom.Insets{Left: unison.StdHSpacing * 2}))
	}
	return label
}

// IsOpen implements unison.TableRowData. The rows are flat, so none is open.
func (r *listFilterRow) IsOpen() bool {
	return false
}

// SetOpen implements unison.TableRowData. The rows are flat, so it does nothing.
func (r *listFilterRow) SetOpen(_ bool) {}

// summarizeListFilter returns a plain-text account of what a filter requires, its groups named for how they combine
// what they hold.
func summarizeListFilter(f *gurps.ListFilter, fields []filterFieldInfo) string {
	byKey := make(map[string]filterFieldInfo, len(fields))
	for _, field := range fields {
		byKey[field.key] = field
	}
	lookup := func(key string) (title string, kind gurps.FilterFieldKind, plural, ok bool) {
		info, ok := byKey[key]
		return info.title, info.kind, info.plural, ok
	}
	units := gurps.SheetSettingsFor(nil).DefaultWeightUnits
	var describe func(node gurps.FilterNode) string
	list := func(group *gurps.FilterGroup) string {
		if group.Empty() {
			return i18n.Text("No conditions")
		}
		parts := make([]string, len(group.Children))
		for i, child := range group.Children {
			parts[i] = describe(child)
		}
		return strings.Join(parts, i18n.Text("; "))
	}
	describe = func(node gurps.FilterNode) string {
		switch one := node.(type) {
		case *gurps.FilterGroup:
			switch groupModeOf(one) {
			case filterAnyOf:
				return i18n.Text("Any of (%s)", list(one))
			case filterNoneOf:
				return i18n.Text("None of (%s)", list(one))
			case filterNotAllOf:
				return i18n.Text("Not all of (%s)", list(one))
			default:
				return i18n.Text("All of (%s)", list(one))
			}
		case *gurps.FilterCondition:
			return stripEm.Replace(one.Describe(lookup, units, emphasize))
		case *gurps.UnknownFilterNode:
			return one.Describe()
		}
		return ""
	}
	// A root that requires all of what it holds reads as a plain list, which is how most filters read.
	if f.Root.All && !f.Root.Not {
		return list(f.Root)
	}
	return describe(f.Root)
}
