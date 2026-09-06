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
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// editorListPanel is the bordered table an editor shows one of its item lists (modifiers, weapons) in, spanning both
// columns of the editor's layout. The list itself belongs to the editor's data; the panel only points at it. A
// concrete panel embeds this and adds the list provider methods that the table's provider reads and writes the list
// through.
type editorListPanel[T gurps.Node[T]] struct {
	unison.Panel
	owner    gurps.DataOwner
	list     *[]T
	provider TableProvider[T]
	table    *unison.Table[*Node[T]]
}

// init sets up the panel and builds its table. self is the concrete panel embedding this one, which provider must
// already be bound to. The list is recorded before the table is built, since building it reads the list through the
// provider.
func (p *editorListPanel[T]) init(self unison.Paneler, owner gurps.DataOwner, list *[]T, provider TableProvider[T], refKey string) {
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
	p.table = newEditorTable(p.AsPanel(), provider)
	p.table.RefKey = refKey
}

// DataOwner implements gurps.DataOwnerProvider.
func (p *editorListPanel[T]) DataOwner() gurps.DataOwner {
	return p.owner
}

// setList replaces the list and re-syncs the table to it, keeping the selection.
func (p *editorListPanel[T]) setList(list []T) {
	*p.list = list
	syncTablePreservingSelection(p.table)
}

// installNewItemHandler makes the command create a new item of the given variant in this panel's table. The handler
// is installed on cmdRoot, the editor holding the panel, which is also what the creation is undone and rebuilt through.
func (p *editorListPanel[T]) installNewItemHandler(cmdRoot Rebuildable, id int, variant ItemVariant) {
	cmdRoot.AsPanel().InstallCmdHandlers(id, unison.AlwaysEnabled,
		func(_ any) { p.provider.CreateItem(cmdRoot, p.table, variant) })
}
