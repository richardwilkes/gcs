// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

const noteDragKey = "note"

var _ TableProvider[*gurps.Note] = &notesProvider{}

type notesProvider struct {
	table    *unison.Table[*Node[*gurps.Note]]
	provider gurps.NoteListProvider
	forPage  bool
}

// NewNotesProvider creates a new table provider for notes.
func NewNotesProvider(provider gurps.NoteListProvider, forPage bool) TableProvider[*gurps.Note] {
	return &notesProvider{
		provider: provider,
		forPage:  forPage,
	}
}

func (p *notesProvider) RefKey() string {
	return gurps.BlockLayoutNotesKey
}

func (p *notesProvider) AllTags() []string {
	return nil
}

func (p *notesProvider) SetTable(table *unison.Table[*Node[*gurps.Note]]) {
	p.table = table
}

func (p *notesProvider) RootRowCount() int {
	return len(p.provider.NoteList())
}

func (p *notesProvider) RootRows() []*Node[*gurps.Note] {
	data := p.provider.NoteList()
	rows := make([]*Node[*gurps.Note], 0, len(data))
	for _, one := range data {
		rows = append(rows, NewNode(p.table, nil, one, p.forPage))
	}
	return rows
}

func (p *notesProvider) SetRootRows(rows []*Node[*gurps.Note]) {
	p.provider.SetNoteList(ExtractNodeDataFromList(rows))
}

func (p *notesProvider) RootData() []*gurps.Note {
	return p.provider.NoteList()
}

func (p *notesProvider) SetRootData(data []*gurps.Note) {
	p.provider.SetNoteList(data)
}

func (p *notesProvider) DataOwner() gurps.DataOwner {
	return p.provider.DataOwner()
}

func (p *notesProvider) DragKey() string {
	return noteDragKey
}

func (p *notesProvider) DragSVG() *unison.SVG {
	return svg.GCSNotes
}

func (p *notesProvider) DropShouldMoveData(from, to *unison.Table[*Node[*gurps.Note]]) bool {
	return from == to
}

func (p *notesProvider) ProcessDropData(_, _ *unison.Table[*Node[*gurps.Note]]) {
}

func (p *notesProvider) AltDropSupport() *AltDropSupport {
	return nil
}

func (p *notesProvider) ItemNames() (singular, plural string) {
	return i18n.Text("Note"), i18n.Text("Notes")
}

func (p *notesProvider) Headers() []unison.TableColumnHeader[*Node[*gurps.Note]] {
	ids := p.ColumnIDs()
	headers := make([]unison.TableColumnHeader[*Node[*gurps.Note]], 0, len(ids))
	for _, id := range ids {
		headers = append(headers, headerFromData[*gurps.Note](gurps.NotesHeaderData(id), p.forPage))
	}
	return headers
}

func (p *notesProvider) SyncHeader(_ []unison.TableColumnHeader[*Node[*gurps.Note]]) {
}

func (p *notesProvider) ColumnIDs() []int {
	columnIDs := []int{
		gurps.NoteTextColumn,
		gurps.NoteReferenceColumn,
	}
	if p.forPage {
		if entity := p.DataOwner().OwningEntity(); entity == nil || !entity.SheetSettings.HideSourceMismatch {
			columnIDs = append(columnIDs, gurps.NoteLibSrcColumn)
		}
	}
	return columnIDs
}

func (p *notesProvider) HierarchyColumnID() int {
	return gurps.NoteTextColumn
}

func (p *notesProvider) ExcessWidthColumnID() int {
	return gurps.NoteTextColumn
}

func (p *notesProvider) OpenEditor(owner Rebuildable, table *unison.Table[*Node[*gurps.Note]]) {
	OpenEditor(table, func(item *gurps.Note) { EditNote(owner, item) })
}

func (p *notesProvider) CreateItem(owner Rebuildable, table *unison.Table[*Node[*gurps.Note]], variant ItemVariant) {
	item := gurps.NewNote(p.DataOwner(), nil, variant == ContainerItemVariant)
	InsertItems(owner, table, p.provider.NoteList, p.provider.SetNoteList,
		func(_ *unison.Table[*Node[*gurps.Note]]) []*Node[*gurps.Note] { return p.RootRows() }, item)
	EditNote(owner, item)
}

func (p *notesProvider) Serialize() ([]byte, error) {
	return jio.SerializeAndCompress(p.provider.NoteList())
}

func (p *notesProvider) Deserialize(data []byte) error {
	var rows []*gurps.Note
	if err := jio.DecompressAndDeserialize(data, &rows); err != nil {
		return err
	}
	p.provider.SetNoteList(rows)
	return nil
}

func (p *notesProvider) ContextMenuItems() []ContextMenuItem {
	var list []ContextMenuItem
	list = append(list,
		ContextMenuItem{i18n.Text("New Note"), NewNoteItemID},
		ContextMenuItem{i18n.Text("New Note Container"), NewNoteContainerItemID},
	)
	return AppendDefaultContextMenuItems(list)
}
