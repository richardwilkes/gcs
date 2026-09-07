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

// ancestryEditorDockable edits an ancestry file. Loading, saving, undo and rebuilding come from fileEditorDockable;
// particular to an ancestry are its content and the name generators its rows offer.
type ancestryEditorDockable struct {
	fileEditorDockable[*gurps.Ancestry]
	// nameGeneratorLookup is a field so that tests can supply choices without a library on disk.
	nameGeneratorLookup func() []string
	// nameGeneratorChoices is what nameGeneratorLookup returned when the content was last built, which is what the
	// rows offer. Every rebuild refreshes it, so a newly saved generator is on offer the next time the ancestry is
	// edited.
	nameGeneratorChoices []string
}

// newAncestryDocument opens an editor holding a new, blank ancestry. Each use opens another editor, leaving any
// already open alone.
func newAncestryDocument() {
	newAncestryEditorDockable().show()
}

// openAncestryRef opens the ancestry the reference names in an editor of its own, or activates the editor already
// showing it.
func openAncestryRef(ref *gurps.NamedFileRef) (*ancestryEditorDockable, error) {
	return openFileEditor(ref, newAncestryEditorDockable)
}

// openAncestryFile opens an ancestry file for the file type registry, which is reached only for a file that isn't
// already showing in a dockable.
func openAncestryFile(filePath string) (unison.Dockable, error) {
	d, err := openAncestryRef(diskFileRef(filePath))
	if err != nil {
		return nil, err
	}
	return d, nil
}

func isAncestryEditor(d unison.Dockable) bool {
	_, ok := d.AsPanel().Self.(*ancestryEditorDockable)
	return ok
}

// newAncestryEditorDockable creates an ancestry editor holding a new, blank ancestry; load replaces that with a
// file's contents and show places it in the dock.
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

// availableNameGeneratorNames returns the base names of the name generators in all libraries, in library order.
func availableNameGeneratorNames() []string {
	refs := gurps.AvailableNameGenerators(gurps.GlobalSettings().Libraries)
	names := make([]string, 0, len(refs))
	for _, ref := range refs {
		names = append(names, ref.FileRef.Name)
	}
	return names
}

func (d *ancestryEditorDockable) buildContent() {
	d.nameGeneratorChoices = d.nameGeneratorLookup()
	d.content.AddChild(newAncestryEditorPanel(d))
}
