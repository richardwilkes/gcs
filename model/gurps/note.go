/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package gurps

import (
	"context"
	"io/fs"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
)

var _ Node[*Note] = &Note{}

// Columns that can be used with the note method .CellData()
const (
	NoteTextColumn = iota
	NoteReferenceColumn
)

const (
	noteListTypeKey = "note_list"
	noteTypeKey     = "note"
)

// Note holds a note.
type Note struct {
	NoteData
	Entity *Entity
}

type noteListData struct {
	Type    string  `json:"type"`
	Version int     `json:"version"`
	Rows    []*Note `json:"rows"`
}

// NewNotesFromFile loads an Note list from a file.
func NewNotesFromFile(fileSystem fs.FS, filePath string) ([]*Note, error) {
	var data noteListData
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &data); err != nil {
		return nil, errs.NewWithCause(invalidFileDataMsg(), err)
	}
	if data.Type != noteListTypeKey {
		return nil, errs.New(unexpectedFileDataMsg())
	}
	if err := CheckVersion(data.Version); err != nil {
		return nil, err
	}
	return data.Rows, nil
}

// SaveNotes writes the Note list to the file as JSON.
func SaveNotes(notes []*Note, filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, &noteListData{
		Type:    noteListTypeKey,
		Version: CurrentDataVersion,
		Rows:    notes,
	})
}

// NewNote creates a new Note.
func NewNote(entity *Entity, parent *Note, container bool) *Note {
	n := &Note{
		NoteData: NoteData{
			ContainerBase: newContainerBase[*Note](noteTypeKey, container),
		},
		Entity: entity,
	}
	n.Text = n.Kind()
	n.parent = parent
	return n
}

// Clone implements Node.
func (n *Note) Clone(entity *Entity, parent *Note, preserveID bool) *Note {
	other := NewNote(entity, parent, n.Container())
	if preserveID {
		other.ID = n.ID
	}
	other.IsOpen = n.IsOpen
	other.ThirdParty = n.ThirdParty
	other.NoteEditData.CopyFrom(n)
	if n.HasChildren() {
		other.Children = make([]*Note, 0, len(n.Children))
		for _, child := range n.Children {
			other.Children = append(other.Children, child.Clone(entity, other, preserveID))
		}
	}
	return other
}

// MarshalJSON implements json.Marshaler.
func (n *Note) MarshalJSON() ([]byte, error) {
	type calc struct {
		ResolvedNotes string `json:"resolved_text,omitempty"`
	}
	n.ClearUnusedFieldsForType()
	data := struct {
		NoteData
		Calc *calc `json:"calc,omitempty"`
	}{
		NoteData: n.NoteData,
	}
	notes := n.resolveText()
	if notes != n.Text {
		data.Calc = &calc{ResolvedNotes: notes}
	}
	return json.Marshal(&data)
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *Note) UnmarshalJSON(data []byte) error {
	n.NoteData = NoteData{}
	if err := json.Unmarshal(data, &n.NoteData); err != nil {
		return err
	}
	n.ClearUnusedFieldsForType()
	if n.Container() {
		for _, one := range n.Children {
			one.parent = n
		}
	}
	return nil
}

func (n *Note) String() string {
	return n.resolveText()
}

func (n *Note) resolveText() string {
	return EvalEmbeddedRegex.ReplaceAllStringFunc(n.Text, n.Entity.EmbeddedEval)
}

// NoteHeaderData returns the header data information for the given note column.
func NoteHeaderData(columnID int) HeaderData {
	var data HeaderData
	switch columnID {
	case NoteTextColumn:
		data.Title = i18n.Text("Note")
		data.Primary = true
	case NoteReferenceColumn:
		data.Title = HeaderBookmark
		data.TitleIsImageKey = true
		data.Detail = PageRefTooltipText()
	}
	return data
}

// CellData returns the cell data information for the given column.
func (n *Note) CellData(columnID int, data *CellData) {
	switch columnID {
	case NoteTextColumn:
		data.Type = cell.Markdown
		data.Primary = n.resolveText()
	case NoteReferenceColumn, PageRefCellAlias:
		data.Type = cell.PageRef
		data.Primary = n.PageRef
		if n.PageRefHighlight != "" {
			data.Secondary = n.PageRefHighlight
		} else {
			data.Secondary = n.resolveText()
		}
	}
}

// Depth returns the number of parents this node has.
func (n *Note) Depth() int {
	count := 0
	p := n.parent
	for p != nil {
		count++
		p = p.parent
	}
	return count
}

// OwningEntity returns the owning Entity.
func (n *Note) OwningEntity() *Entity {
	return n.Entity
}

// SetOwningEntity sets the owning entity and configures any sub-components as needed.
func (n *Note) SetOwningEntity(entity *Entity) {
	n.Entity = entity
	if n.Container() {
		for _, child := range n.Children {
			child.SetOwningEntity(entity)
		}
	}
}

// Enabled returns true if this node is enabled.
func (n *Note) Enabled() bool {
	return true
}

// FillWithNameableKeys adds any nameable keys found to the provided map.
func (n *Note) FillWithNameableKeys(m map[string]string) {
	Extract(n.Text, m)
}

// ApplyNameableKeys replaces any nameable keys found with the corresponding values in the provided map.
func (n *Note) ApplyNameableKeys(m map[string]string) {
	n.Text = Apply(n.Text, m)
}
