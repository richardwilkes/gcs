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
	"github.com/richardwilkes/gcs/v5/model/gurps/gid"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/gcs/v5/ui/widget/ntable"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/unison"
)

var (
	noteColMap = map[int]int{
		0: gurps.NoteTextColumn,
		1: gurps.NoteReferenceColumn,
	}
	_ ntable.TableProvider[*gurps.Note] = &notesProvider{}
)

type notesProvider struct {
	table    *unison.Table[*ntable.Node[*gurps.Note]]
	provider gurps.NoteListProvider
	forPage  bool
}

// NewNotesProvider creates a new table provider for notes.
func NewNotesProvider(provider gurps.NoteListProvider, forPage bool) ntable.TableProvider[*gurps.Note] {
	return &notesProvider{
		provider: provider,
		forPage:  forPage,
	}
}

func (p *notesProvider) SetTable(table *unison.Table[*ntable.Node[*gurps.Note]]) {
	p.table = table
}

func (p *notesProvider) RootRowCount() int {
	return len(p.provider.NoteList())
}

func (p *notesProvider) RootRows() []*ntable.Node[*gurps.Note] {
	data := p.provider.NoteList()
	rows := make([]*ntable.Node[*gurps.Note], 0, len(data))
	for _, one := range data {
		rows = append(rows, ntable.NewNode[*gurps.Note](p.table, nil, noteColMap, one, p.forPage))
	}
	return rows
}

func (p *notesProvider) SetRootRows(rows []*ntable.Node[*gurps.Note]) {
	p.provider.SetNoteList(ntable.ExtractNodeDataFromList(rows))
}

func (p *notesProvider) RootData() []*gurps.Note {
	return p.provider.NoteList()
}

func (p *notesProvider) SetRootData(data []*gurps.Note) {
	p.provider.SetNoteList(data)
}

func (p *notesProvider) Entity() *gurps.Entity {
	return p.provider.Entity()
}

func (p *notesProvider) DragKey() string {
	return gid.Note
}

func (p *notesProvider) DragSVG() *unison.SVG {
	return res.GCSNotesSVG
}

func (p *notesProvider) DropShouldMoveData(from, to *unison.Table[*ntable.Node[*gurps.Note]]) bool {
	return from == to
}

func (p *notesProvider) ProcessDropData(_, _ *unison.Table[*ntable.Node[*gurps.Note]]) {
}

func (p *notesProvider) ItemNames() (singular, plural string) {
	return i18n.Text("Note"), i18n.Text("Notes")
}

func (p *notesProvider) Headers() []unison.TableColumnHeader[*ntable.Node[*gurps.Note]] {
	var headers []unison.TableColumnHeader[*ntable.Node[*gurps.Note]]
	for i := 0; i < len(noteColMap); i++ {
		switch noteColMap[i] {
		case gurps.NoteTextColumn:
			headers = append(headers, NewHeader[*gurps.Note](i18n.Text("Note"), "", p.forPage))
		case gurps.NoteReferenceColumn:
			headers = append(headers, NewPageRefHeader[*gurps.Note](p.forPage))
		default:
			jot.Fatalf(1, "invalid note column: %d", noteColMap[i])
		}
	}
	return headers
}

func (p *notesProvider) SyncHeader(_ []unison.TableColumnHeader[*ntable.Node[*gurps.Note]]) {
}

func (p *notesProvider) HierarchyColumnIndex() int {
	for k, v := range noteColMap {
		if v == gurps.NoteTextColumn {
			return k
		}
	}
	return 0
}

func (p *notesProvider) ExcessWidthColumnIndex() int {
	return p.HierarchyColumnIndex()
}

func (p *notesProvider) OpenEditor(owner widget.Rebuildable, table *unison.Table[*ntable.Node[*gurps.Note]]) {
	ntable.OpenEditor[*gurps.Note](table, func(item *gurps.Note) { EditNote(owner, item) })
}

func (p *notesProvider) CreateItem(owner widget.Rebuildable, table *unison.Table[*ntable.Node[*gurps.Note]], variant ntable.ItemVariant) {
	item := gurps.NewNote(p.Entity(), nil, variant == ntable.ContainerItemVariant)
	ntable.InsertItems[*gurps.Note](owner, table, p.provider.NoteList, p.provider.SetNoteList,
		func(_ *unison.Table[*ntable.Node[*gurps.Note]]) []*ntable.Node[*gurps.Note] { return p.RootRows() }, item)
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
