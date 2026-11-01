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
	"maps"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// editorListPanel is the bordered table an editor shows one of its item lists (modifiers, weapons) in, spanning both
// columns of the editor's layout. The list itself belongs to the editor's data; the panel only points at it.
type editorListPanel[T gurps.Node[T]] struct {
	unison.Panel
	owner    gurps.DataOwner
	list     *[]T
	provider TableProvider[T]
	table    *unison.Table[*Node[T]]
	// attach, when set, points the nodes of a replacement list, or a node joining the list, at what they belong to,
	// which the serialized form undo restores them from doesn't record.
	attach func(list []T)
}

// init sets up the panel and builds its table. self is the concrete panel embedding this one, which provider must
// already be bound to, since building the table reads the list through the provider.
func (p *editorListPanel[T]) init(self unison.Paneler, cmdRoot Rebuildable, owner gurps.DataOwner, list *[]T, provider TableProvider[T], refKey string) {
	p.Self = self
	p.owner = owner
	p.list = list
	p.provider = provider
	p.SetLayout(&unison.FlexLayout{Columns: 1})
	p.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: align.Fill,
		HGrab:  true,
	})
	p.SetBorder(unison.NewLineBorder(unison.ThemeAboveSurface, geom.Size{}, geom.NewUniformInsets(1), false))
	p.table = newEditorTable(p.AsPanel(), provider, !inReadOnlyEditor(cmdRoot))
	p.table.RefKey = refKey
}

// makeReadOnly leaves the panel on, so that the rows of its table can still be selected and opened. In a read-only
// editor, the table is built without the commands that change the list.
func (p *editorListPanel[T]) makeReadOnly() {}

// listTable returns the table the panel shows its list in.
func (p *editorListPanel[T]) listTable() *unison.Table[*Node[T]] {
	return p.table
}

// DataOwner implements gurps.DataOwnerProvider.
func (p *editorListPanel[T]) DataOwner() gurps.DataOwner {
	return p.owner
}

// setList replaces the list and re-syncs the table to it, keeping the selection.
func (p *editorListPanel[T]) setList(list []T) {
	if p.attach != nil {
		p.attach(list)
	}
	*p.list = list
	p.syncRows()
}

// syncRows re-syncs the table to the list, keeping the selection, so rows whose items changed get their new heights.
func (p *editorListPanel[T]) syncRows() {
	syncTablePreservingSelection(p.table)
}

// attachNode points a node joining the list at what the list's nodes belong to, as setList does for a replacement list.
func (p *editorListPanel[T]) attachNode(node T) {
	if p.attach != nil {
		p.attach([]T{node})
	}
}

// reattach points the list's nodes back at what they belong to, since copying the editor's data from another node
// leaves them pointed at that node.
func (p *editorListPanel[T]) reattach() {
	if p.attach != nil {
		p.attach(*p.list)
	}
}

// replacementsHolder holds nameable replacements that can be replaced as a whole, as the data of an item editor does.
type replacementsHolder interface {
	nameable.Accesser
	nameable.Setter
}

// editedModifierReplacements is the data of an item editor, whose replacements take those an older version saved on the
// modifiers joining the editor's list (see attachEditedModifiers), along with the values each of them added, by its ID,
// so that they leave the list and return to it with the modifier (see followEditedModifiers).
type editedModifierReplacements struct {
	data  replacementsHolder
	added map[tid.TID]*legacyAddition
}

// legacyAddition is what a modifier joining an item editor's list added to the replacements of the editor's data.
type legacyAddition struct {
	values map[string]string
	// out is whether the modifier has left the list, taking the values out of the replacements.
	out bool
}

// attachEditedModifiers points the modifiers at the item being edited, as gurps.AttachModifiers does, except that the
// replacements an older version saved on a modifier, which SetTarget would move into the item, are moved into those of
// the editor's data, keeping any value it already has, so that they are applied or canceled with the rest of the edit.
// legacy returns where a modifier keeps its own replacements.
func attachEditedModifiers[T gurps.ModifiableNode[T, M], M gurps.ModifierNode[M, T]](item T, r *editedModifierReplacements, list []M, legacy func(M) *map[string]string) {
	replacements := r.data.NameableReplacements()
	changed := false
	gurps.Traverse(func(mod M) bool {
		if from := legacy(mod); len(*from) != 0 {
			for k, v := range *from {
				if _, exists := replacements[k]; !exists {
					if replacements == nil {
						replacements = make(map[string]string, len(*from))
					}
					replacements[k] = v
					r.note(mod.ID(), k, v)
					changed = true
				}
			}
			*from = nil
		}
		return false
	}, false, false, list...)
	if changed {
		r.data.SetNameableReplacements(replacements)
	}
	gurps.AttachModifiers(item, list)
}

// note records that the modifier with the ID added the value under the key.
func (r *editedModifierReplacements) note(id tid.TID, key, value string) {
	one := r.added[id]
	if one == nil {
		if r.added == nil {
			r.added = make(map[tid.TID]*legacyAddition)
		}
		one = &legacyAddition{values: make(map[string]string)}
		r.added[id] = one
	}
	one.values[key] = value
}

// followEditedModifiers takes the values a modifier added to the replacements of the editor's data (see
// attachEditedModifiers) back out once the modifier is no longer in the list, as when the drop or paste that brought it
// is undone, keeping any that have been changed since, and puts them back should it return, as on redo, unless the data
// has a value of its own for them by then. list is the editor's whole list.
func followEditedModifiers[M gurps.Node[M]](r *editedModifierReplacements, list []M) {
	if len(r.added) == 0 {
		return
	}
	present := make(map[tid.TID]bool)
	gurps.Traverse(func(mod M) bool {
		present[mod.ID()] = true
		return false
	}, false, false, list...)
	replacements := maps.Clone(r.data.NameableReplacements())
	changed := false
	for id, one := range r.added {
		if present[id] != one.out {
			continue
		}
		one.out = !one.out
		for k, v := range one.values {
			current, exists := replacements[k]
			switch {
			case one.out && exists && current == v:
				delete(replacements, k)
				changed = true
			case !one.out && !exists:
				if replacements == nil {
					replacements = make(map[string]string, len(one.values))
				}
				replacements[k] = v
				changed = true
			}
		}
	}
	if changed {
		if len(replacements) == 0 {
			replacements = nil
		}
		r.data.SetNameableReplacements(replacements)
	}
}

// installNewItemHandler makes the command create a new item of the given variant in this panel's table. The handler is
// installed on cmdRoot, the editor holding the panel, which the creation is also undone and rebuilt through.
func (p *editorListPanel[T]) installNewItemHandler(cmdRoot Rebuildable, id int, variant ItemVariant) {
	cmdRoot.AsPanel().InstallCmdHandlers(id, func(_ any) bool { return !inReadOnlyEditor(cmdRoot) },
		func(_ any) { p.provider.CreateItem(cmdRoot, p.table, variant) })
}
