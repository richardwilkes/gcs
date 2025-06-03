// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"context"
	"hash"
	"io/fs"
	"maps"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/tid"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
	"github.com/richardwilkes/unison/enums/align"
)

var (
	_ Node[*Note]       = &Note{}
	_ EditorData[*Note] = &NoteEditData{}
)

// Columns that can be used with the note method .CellData()
const (
	NoteTextColumn = iota
	NoteReferenceColumn
	NoteLibSrcColumn
)

// Note holds a note.
type Note struct {
	NoteData
	owner DataOwner
}

// NoteData holds the Note data that is written to disk.
type NoteData struct {
	SourcedID
	NoteEditData
	ThirdParty map[string]any `json:"third_party,omitempty"`
	Children   []*Note        `json:"children,omitempty"` // Only for containers
	parent     *Note
}

// NoteEditData holds the Note data that can be edited by the UI detail editor.
type NoteEditData struct {
	NoteSyncData
	Replacements map[string]string `json:"replacements,omitempty"`
}

// NoteSyncData holds the note sync data that is common to both containers and non-containers.
type NoteSyncData struct {
	MarkDown         string `json:"markdown,omitempty"`
	PageRef          string `json:"reference,omitempty"`
	PageRefHighlight string `json:"reference_highlight,omitempty"`
}

type noteListData struct {
	Version int     `json:"version"`
	Rows    []*Note `json:"rows"`
}

// NewNotesFromFile loads an Note list from a file.
func NewNotesFromFile(fileSystem fs.FS, filePath string) ([]*Note, error) {
	var data noteListData
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &data); err != nil {
		return nil, errs.NewWithCause(InvalidFileData(), err)
	}
	if err := jio.CheckVersion(data.Version); err != nil {
		return nil, err
	}
	return data.Rows, nil
}

// SaveNotes writes the Note list to the file as JSON.
func SaveNotes(notes []*Note, filePath string) error {
	return jio.SaveToFile(context.Background(), filePath, &noteListData{
		Version: jio.CurrentDataVersion,
		Rows:    notes,
	})
}

// NewNote creates a new Note.
func NewNote(owner DataOwner, parent *Note, container bool) *Note {
	var n Note
	n.TID = tid.MustNewTID(noteKind(container))
	n.parent = parent
	n.owner = owner
	n.MarkDown = n.Kind()
	n.SetOpen(container)
	return &n
}

func noteKind(container bool) byte {
	if container {
		return kinds.NoteContainer
	}
	return kinds.Note
}

// ID returns the local ID of this data.
func (n *Note) ID() tid.TID {
	return n.TID
}

// Container returns true if this is a container.
func (n *Note) Container() bool {
	return tid.IsKind(n.TID, kinds.NoteContainer)
}

// HasChildren returns true if this node has children.
func (n *Note) HasChildren() bool {
	return n.Container() && len(n.Children) > 0
}

// NodeChildren returns the children of this node, if any.
func (n *Note) NodeChildren() []*Note {
	return n.Children
}

// SetChildren sets the children of this node.
func (n *Note) SetChildren(children []*Note) {
	n.Children = children
}

// Parent returns the parent.
func (n *Note) Parent() *Note {
	return n.parent
}

// SetParent sets the parent.
func (n *Note) SetParent(parent *Note) {
	n.parent = parent
}

// IsOpen returns true if this node is currently open.
func (n *Note) IsOpen() bool {
	return IsNodeOpen(n)
}

// SetOpen sets the current open state for this node.
func (n *Note) SetOpen(open bool) {
	SetNodeOpen(n, open)
}

// Clone implements Node.
func (n *Note) Clone(from LibraryFile, owner DataOwner, parent *Note, preserveID bool) *Note {
	other := NewNote(owner, parent, n.Container())
	other.AdjustSource(from, n.SourcedID, preserveID)
	other.SetOpen(n.IsOpen())
	other.ThirdParty = n.ThirdParty
	other.CopyFrom(n)
	if n.HasChildren() {
		other.Children = make([]*Note, 0, len(n.Children))
		for _, child := range n.Children {
			other.Children = append(other.Children, child.Clone(from, owner, other, preserveID))
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
	if notes != n.MarkDown {
		data.Calc = &calc{ResolvedNotes: notes}
	}
	return json.Marshal(&data)
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *Note) UnmarshalJSON(data []byte) error {
	var localData struct {
		NoteData
		// Old data fields
		Type     string `json:"type"`
		ExprText string `json:"text"`
		IsOpen   bool   `json:"open"`
	}
	if err := json.Unmarshal(data, &localData); err != nil {
		return err
	}
	setOpen := false
	if !tid.IsValid(localData.TID) {
		// Fixup old data that used UUIDs instead of TIDs
		localData.TID = tid.MustNewTID(noteKind(strings.HasSuffix(localData.Type, containerKeyPostfix)))
		setOpen = localData.IsOpen
	}
	n.NoteData = localData.NoteData
	if n.MarkDown == "" && localData.ExprText != "" {
		n.MarkDown = EmbeddedExprToScript(localData.ExprText)
	}
	n.ClearUnusedFieldsForType()
	if n.Container() {
		for _, one := range n.Children {
			one.parent = n
		}
	}
	if setOpen {
		SetNodeOpen(n, true)
	}
	return nil
}

// TextWithReplacements returns the text with any replacements applied.
func (n *Note) TextWithReplacements() string {
	return nameable.Apply(n.MarkDown, n.Replacements)
}

func (n *Note) String() string {
	return n.resolveText()
}

func (n *Note) resolveText() string {
	return ResolveText(EntityFromNode(n), ScriptSelfProvider{}, n.TextWithReplacements())
}

// NotesHeaderData returns the header data information for the given note column.
func NotesHeaderData(columnID int) HeaderData {
	var data HeaderData
	switch columnID {
	case NoteTextColumn:
		data.Title = i18n.Text("Note")
		data.Primary = true
	case NoteReferenceColumn:
		data.Title = HeaderBookmark
		data.TitleIsImageKey = true
		data.Detail = PageRefTooltip()
	case NoteLibSrcColumn:
		data.Title = HeaderDatabase
		data.TitleIsImageKey = true
		data.Detail = LibSrcTooltip()
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
	case NoteLibSrcColumn:
		data.Type = cell.Text
		data.Alignment = align.Middle
		if !toolbox.IsNil(n.owner) {
			state, _ := n.owner.SourceMatcher().Match(n)
			data.Primary = state.AltString()
			data.Tooltip = state.String()
			if state != srcstate.Custom {
				data.Tooltip += "\n" + n.Source.String()
			}
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

// DataOwner returns the data owner.
func (n *Note) DataOwner() DataOwner {
	return n.owner
}

// SetDataOwner sets the data owner and configures any sub-components as needed.
func (n *Note) SetDataOwner(owner DataOwner) {
	n.owner = owner
	if n.Container() {
		for _, child := range n.Children {
			child.SetDataOwner(owner)
		}
	}
}

// Enabled returns true if this node is enabled.
func (n *Note) Enabled() bool {
	return true
}

// NameableReplacements returns the replacements to be used with Nameables.
func (n *Note) NameableReplacements() map[string]string {
	if n == nil {
		return nil
	}
	return n.Replacements
}

// FillWithNameableKeys adds any nameable keys found to the provided map.
func (n *Note) FillWithNameableKeys(m, existing map[string]string) {
	if existing == nil {
		existing = n.Replacements
	}
	nameable.Extract(n.MarkDown, m, existing)
}

// ApplyNameableKeys replaces any nameable keys found with the corresponding values in the provided map.
func (n *Note) ApplyNameableKeys(m map[string]string) {
	needed := make(map[string]string)
	n.FillWithNameableKeys(needed, nil)
	n.Replacements = nameable.Reduce(needed, m)
}

// CanConvertToFromContainer returns true if this node can be converted to/from a container.
func (n *Note) CanConvertToFromContainer() bool {
	return !n.Container() || !n.HasChildren()
}

// ConvertToContainer converts this node to a container.
func (n *Note) ConvertToContainer() {
	n.TID = tid.TID(kinds.NoteContainer) + n.TID[1:]
}

// ConvertToNonContainer converts this node to a non-container.
func (n *Note) ConvertToNonContainer() {
	n.TID = tid.TID(kinds.Note) + n.TID[1:]
}

// Kind returns the kind of data.
func (n *Note) Kind() string {
	if n.Container() {
		return i18n.Text("Note Container")
	}
	return i18n.Text("Note")
}

// ClearUnusedFieldsForType zeroes out the fields that are not applicable to this type (container vs not-container).
func (n *Note) ClearUnusedFieldsForType() {
	if !n.Container() {
		n.Children = nil
	}
}

// GetSource returns the source of this data.
func (n *Note) GetSource() Source {
	return n.Source
}

// ClearSource clears the source of this data.
func (n *Note) ClearSource() {
	n.Source = Source{}
}

// SyncWithSource synchronizes this data with the source.
func (n *Note) SyncWithSource() {
	if !toolbox.IsNil(n.owner) {
		if state, data := n.owner.SourceMatcher().Match(n); state == srcstate.Mismatched {
			if other, ok := data.(*Note); ok {
				n.NoteSyncData = other.NoteSyncData
			}
		}
	}
}

// Hash writes this object's contents into the hasher. Note that this only hashes the data that is considered to be
// "source" data, i.e. not expected to be modified by the user after copying from a library.
func (n *Note) Hash(h hash.Hash) {
	n.hash(h)
}

func (n *NoteSyncData) hash(h hash.Hash) {
	hashhelper.String(h, n.MarkDown)
	hashhelper.String(h, n.PageRef)
	hashhelper.String(h, n.PageRefHighlight)
}

// CopyFrom implements node.EditorData.
func (n *NoteEditData) CopyFrom(other *Note) {
	n.copyFrom(&other.NoteEditData)
}

// ApplyTo implements node.EditorData.
func (n *NoteEditData) ApplyTo(other *Note) {
	other.copyFrom(n)
}

func (n *NoteEditData) copyFrom(other *NoteEditData) {
	*n = *other
	n.Replacements = maps.Clone(other.Replacements)
}
