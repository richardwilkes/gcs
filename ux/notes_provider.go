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
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/unison"
)

var _ TableProvider[*gurps.Note] = &notesProvider{}

type notesProvider struct {
	listProvider[*gurps.Note]
}

// NewNotesProvider creates a new table provider for notes.
func NewNotesProvider(provider gurps.NoteListProvider, forPage bool) TableProvider[*gurps.Note] {
	p := &notesProvider{}
	p.listProvider = listProvider[*gurps.Note]{
		dataOwner:  provider,
		list:       provider.NoteList,
		setList:    provider.SetNoteList,
		columnIDs:  p.ColumnIDs,
		headerData: gurps.NotesHeaderData,
		newItem:    gurps.NewNote,
		edit:       EditNote,
		forPage:    forPage,
	}
	return p
}

func (p *notesProvider) RefKey() string {
	return gurps.BlockNotesKey
}

func (p *notesProvider) DragKey() *uti.DataType {
	return noteDragKey
}

func (p *notesProvider) DragSVG() *unison.SVG {
	return svg.GCSNotes
}

func (p *notesProvider) ItemNames() (singular, plural string) {
	return i18n.Text("Note"), i18n.Text("Notes")
}

func (p *notesProvider) ColumnIDs() []int {
	columnIDs := []int{gurps.NoteTextColumn}
	if !p.forPage {
		columnIDs = append(columnIDs, gurps.NoteTagsColumn)
	}
	return p.appendReferenceColumns(columnIDs, gurps.NoteReferenceColumn, gurps.NoteLibSrcColumn)
}

func (p *notesProvider) HierarchyColumnID() int {
	return gurps.NoteTextColumn
}

func (p *notesProvider) ExcessWidthColumnID() int {
	return gurps.NoteTextColumn
}

func (p *notesProvider) ContextMenuItems() []ContextMenuItem {
	return AppendDefaultContextMenuItems([]ContextMenuItem{
		contextMenuItemFor(newNoteAction),
		contextMenuItemFor(newNoteContainerAction),
	})
}
