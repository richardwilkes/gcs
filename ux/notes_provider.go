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
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/unison"
)

const noteDragKey = "note"

var (
	noteColMap = map[int]int{
		0: model.NoteTextColumn,
		1: model.NoteReferenceColumn,
	}
	_ TableProvider[*model.Note] = &notesProvider{}
)

type notesProvider struct {
	table    *unison.Table[*Node[*model.Note]]
	provider model.NoteListProvider
	forPage  bool
}

// NewNotesProvider creates a new table provider for notes.
func NewNotesProvider(provider model.NoteListProvider, forPage bool) TableProvider[*model.Note] {
	return &notesProvider{
		provider: provider,
		forPage:  forPage,
	}
}

func (p *notesProvider) RefKey() string {
	return model.BlockLayoutNotesKey
}

func (p *notesProvider) AllTags() []string {
	return nil
}

func (p *notesProvider) SetTable(table *unison.Table[*Node[*model.Note]]) {
	p.table = table
}

func (p *notesProvider) RootRowCount() int {
	return len(p.provider.NoteList())
}

func (p *notesProvider) RootRows() []*Node[*model.Note] {
	data := p.provider.NoteList()
	rows := make([]*Node[*model.Note], 0, len(data))
	for _, one := range data {
		rows = append(rows, NewNode[*model.Note](p.table, nil, noteColMap, one, p.forPage))
	}
	return rows
}

func (p *notesProvider) SetRootRows(rows []*Node[*model.Note]) {
	p.provider.SetNoteList(ExtractNodeDataFromList(rows))
}

func (p *notesProvider) RootData() []*model.Note {
	return p.provider.NoteList()
}

func (p *notesProvider) SetRootData(data []*model.Note) {
	p.provider.SetNoteList(data)
}

func (p *notesProvider) Entity() *model.Entity {
	return p.provider.Entity()
}

func (p *notesProvider) DragKey() string {
	return noteDragKey
}

func (p *notesProvider) DragSVG() *unison.SVG {
	return svg.GCSNotes
}

func (p *notesProvider) DropShouldMoveData(from, to *unison.Table[*Node[*model.Note]]) bool {
	return from == to
}

func (p *notesProvider) ProcessDropData(_, _ *unison.Table[*Node[*model.Note]]) {
}

func (p *notesProvider) AltDropSupport() *AltDropSupport {
	return nil
}

func (p *notesProvider) ItemNames() (singular, plural string) {
	return i18n.Text("Note"), i18n.Text("Notes")
}

func (p *notesProvider) Headers() []unison.TableColumnHeader[*Node[*model.Note]] {
	var headers []unison.TableColumnHeader[*Node[*model.Note]]
	for i := 0; i < len(noteColMap); i++ {
		switch noteColMap[i] {
		case model.NoteTextColumn:
			headers = append(headers, NewEditorListHeader[*model.Note](i18n.Text("Note"), "", p.forPage))
		case model.NoteReferenceColumn:
			headers = append(headers, NewEditorPageRefHeader[*model.Note](p.forPage))
		default:
			jot.Fatalf(1, "invalid note column: %d", noteColMap[i])
		}
	}
	return headers
}

func (p *notesProvider) SyncHeader(_ []unison.TableColumnHeader[*Node[*model.Note]]) {
}

func (p *notesProvider) HierarchyColumnIndex() int {
	for k, v := range noteColMap {
		if v == model.NoteTextColumn {
			return k
		}
	}
	return 0
}

func (p *notesProvider) ExcessWidthColumnIndex() int {
	return p.HierarchyColumnIndex()
}

func (p *notesProvider) OpenEditor(owner Rebuildable, table *unison.Table[*Node[*model.Note]]) {
	OpenEditor[*model.Note](table, func(item *model.Note) { EditNote(owner, item) })
}

func (p *notesProvider) CreateItem(owner Rebuildable, table *unison.Table[*Node[*model.Note]], variant ItemVariant) {
	item := model.NewNote(p.Entity(), nil, variant == ContainerItemVariant)
	InsertItems[*model.Note](owner, table, p.provider.NoteList, p.provider.SetNoteList,
		func(_ *unison.Table[*Node[*model.Note]]) []*Node[*model.Note] { return p.RootRows() }, item)
	EditNote(owner, item)
}

func (p *notesProvider) Serialize() ([]byte, error) {
	return jio.SerializeAndCompress(p.provider.NoteList())
}

func (p *notesProvider) Deserialize(data []byte) error {
	var rows []*model.Note
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
