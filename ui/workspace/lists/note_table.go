package lists

import (
	"os"
	"path/filepath"

	"github.com/richardwilkes/gcs/v5/constants"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/ui/workspace/editors"
	"github.com/richardwilkes/unison"
)

type noteListProvider struct {
	notes []*gurps.Note
}

func (p *noteListProvider) Entity() *gurps.Entity {
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
	return NewTableDockable(filePath, library.NotesExt, editors.NewNotesProvider(provider, false),
		func(path string) error { return gurps.SaveNotes(provider.NoteList(), path) },
		constants.NewNoteItemID, constants.NewNoteContainerItemID)
}
