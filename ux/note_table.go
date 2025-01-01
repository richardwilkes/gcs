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
	"os"
	"path/filepath"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/unison"
)

type noteListProvider struct {
	notes []*gurps.Note
}

func (p *noteListProvider) DataOwner() gurps.DataOwner {
	return nil
}

func (p *noteListProvider) NoteList() []*gurps.Note {
	return p.notes
}

func (p *noteListProvider) SetNoteList(list []*gurps.Note) {
	p.notes = list
}

// NewNoteTableDockableFromFile loads a list of notes from a file and creates a new unison.Dockable for them.
func NewNoteTableDockableFromFile(filePath string) (unison.Dockable, error) {
	notes, err := gurps.NewNotesFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	d := NewNoteTableDockable(filePath, notes)
	d.needsSaveAsPrompt = false
	return d, nil
}

// NewNoteTableDockable creates a new unison.Dockable for note list files.
func NewNoteTableDockable(filePath string, notes []*gurps.Note) *TableDockable[*gurps.Note] {
	provider := &noteListProvider{notes: notes}
	d := NewTableDockable(filePath, gurps.NotesExt, NewNotesProvider(provider, false),
		func(path string) error { return gurps.SaveNotes(provider.NoteList(), path) },
		NewNoteItemID, NewNoteContainerItemID)
	InstallContainerConversionHandlers(d, d, d.table)
	return d
}
