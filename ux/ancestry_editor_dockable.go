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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

var (
	_ FileBackedDockable = &ancestryEditorDockable{}
	_ saveable           = &ancestryEditorDockable{}
	_ structuralEditor   = &ancestryEditorDockable{}
)

// ancestryEditorDockable edits an ancestry file. Everything about loading, saving, undo and rebuilding the content is
// in fileEditorDockable, which it shares with the name generator editor; what is particular to an ancestry is its
// content and the name generators its rows offer.
type ancestryEditorDockable struct {
	fileEditorDockable[*gurps.Ancestry]
	// nameGeneratorLookup returns the base names of the available name generators. It is a field so that tests can
	// supply choices without a library on disk.
	nameGeneratorLookup func() []string
	// nameGeneratorChoices is what nameGeneratorLookup returned when the content was last built, which is what the
	// rows offer. Every rebuild refreshes it, so a generator saved from the name generator editor is on offer as soon
	// as the ancestry is next edited.
	nameGeneratorChoices []string
}

// newAncestryDocument is what File > New Ancestry does: it opens an editor holding a new, blank ancestry. Every choice
// of the item opens another editor; those already open are left as they are.
func newAncestryDocument() {
	newAncestryEditorDockable().show()
}

// openAncestryRef opens the ancestry the reference names in an editor of its own, or activates the editor already
// showing it; see openFileEditor.
func openAncestryRef(ref *gurps.NamedFileRef) (*ancestryEditorDockable, error) {
	return openFileEditor(ref, newAncestryEditorDockable)
}

// openAncestryFile opens an ancestry file the way the file type registry expects. The workspace has already looked for
// a dockable showing the file, so this is reached for one that is not open yet.
func openAncestryFile(filePath string) (unison.Dockable, error) {
	d, err := openAncestryRef(diskFileRef(filePath))
	if err != nil {
		return nil, err
	}
	return d, nil
}

// isAncestryEditor returns true if the dockable is the ancestry editor.
func isAncestryEditor(d unison.Dockable) bool {
	_, ok := d.AsPanel().Self.(*ancestryEditorDockable)
	return ok
}

// newAncestryEditorDockable creates an ancestry editor holding a new, blank ancestry, ready for load to replace that
// with a file's and for show to build its toolbar and content and place it in the dock.
func newAncestryEditorDockable() *ancestryEditorDockable {
	d := &ancestryEditorDockable{nameGeneratorLookup: availableNameGeneratorNames}
	d.Self = d
	d.init(fileEditorSpec[*gurps.Ancestry]{
		tabTitle:     i18n.Text("Ancestry"),
		icon:         svg.Ancestry,
		ext:          gurps.AncestryExt,
		newModel:     gurps.NewAncestry,
		readModel:    gurps.NewAncestryFromFile,
		buildContent: d.buildContent,
		fallbackName: func() string { return strings.TrimSpace(d.model.Name) },
		openRef: func(ref *gurps.NamedFileRef) error {
			_, err := openAncestryRef(ref)
			return err
		},
	})
	return d
}

// availableNameGeneratorNames returns the base names of the name generators every library holds, in the order the
// libraries list them.
func availableNameGeneratorNames() []string {
	refs := gurps.AvailableNameGenerators(gurps.GlobalSettings().Libraries)
	names := make([]string, 0, len(refs))
	for _, ref := range refs {
		names = append(names, ref.FileRef.Name)
	}
	return names
}

// buildContent fills the content with the editor for the ancestry, offering the name generators available now.
func (d *ancestryEditorDockable) buildContent() {
	d.nameGeneratorChoices = d.nameGeneratorLookup()
	d.content.AddChild(newAncestryEditorPanel(d))
}
