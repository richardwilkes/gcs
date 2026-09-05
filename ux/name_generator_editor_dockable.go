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
	"github.com/richardwilkes/unison"
)

var (
	_ FileBackedDockable = &nameGeneratorEditorDockable{}
	_ saveable           = &nameGeneratorEditorDockable{}
	_ structuralEditor   = &nameGeneratorEditorDockable{}
)

// nameGeneratorEditorDockable edits a name generator file. Everything about loading, saving, undo and rebuilding the
// content is in fileEditorDockable, which it shares with the ancestry editor; what is particular to a name generator
// is its content, which starts with a sample of the names it generates.
type nameGeneratorEditorDockable struct {
	fileEditorDockable[*gurps.NameGenerator]
	samples *nameGeneratorSamplesPanel
}

// newNameGeneratorDocument is what File > New Name Generator does: it opens an editor holding a new, empty generator.
// Every choice of the item opens another editor; those already open are left as they are.
func newNameGeneratorDocument() {
	newNameGeneratorEditorDockable().show()
}

// openNameGeneratorRef opens the name generator the reference names in an editor of its own, or activates the editor
// already showing it; see openFileEditor.
func openNameGeneratorRef(ref *gurps.NamedFileRef) (*nameGeneratorEditorDockable, error) {
	return openFileEditor(ref, newNameGeneratorEditorDockable)
}

// openNameGeneratorFile opens a name generator file the way the file type registry expects. The workspace has already
// looked for a dockable showing the file, so this is reached for one that is not open yet.
func openNameGeneratorFile(filePath string) (unison.Dockable, error) {
	d, err := openNameGeneratorRef(diskFileRef(filePath))
	if err != nil {
		return nil, err
	}
	return d, nil
}

// OpenNameGeneratorInEditor opens the named generator in a name generator editor: the one already showing it, if there
// is one, or a new one. A name that no library holds is ignored, since there is nothing to open.
func OpenNameGeneratorInEditor(name string) {
	ref := nameGeneratorRefNamed(name)
	if ref == nil {
		return
	}
	if _, err := openNameGeneratorRef(ref); err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to load Name Generator"), err)
	}
}

// nameGeneratorRefNamed returns the reference to the name generator with the given base name, or nil if no library
// holds one.
func nameGeneratorRefNamed(name string) *gurps.NamedFileRef {
	for _, one := range gurps.AvailableNameGenerators(gurps.GlobalSettings().Libraries) {
		if one.FileRef.Name == name {
			return one.FileRef
		}
	}
	return nil
}

// isNameGeneratorEditor returns true if the dockable is the name generator editor.
func isNameGeneratorEditor(d unison.Dockable) bool {
	_, ok := d.AsPanel().Self.(*nameGeneratorEditorDockable)
	return ok
}

// newNameGeneratorEditorDockable creates a name generator editor holding a new, empty generator, ready for load to
// replace that with a file's and for show to build its toolbar and content and place it in the dock.
func newNameGeneratorEditorDockable() *nameGeneratorEditorDockable {
	d := &nameGeneratorEditorDockable{}
	d.Self = d
	d.init(fileEditorSpec[*gurps.NameGenerator]{
		tabTitle:     i18n.Text("Name Generator"),
		icon:         svg.Naming,
		ext:          gurps.NamesExt,
		newModel:     gurps.NewNameGenerator,
		readModel:    gurps.ReadNameGeneratorFromFS,
		buildContent: d.buildContent,
		openRef: func(ref *gurps.NamedFileRef) error {
			_, err := openNameGeneratorRef(ref)
			return err
		},
	})
	return d
}

// buildContent fills the content with the sample names and the editor for the generator.
func (d *nameGeneratorEditorDockable) buildContent() {
	d.samples = newNameGeneratorSamplesPanel(d)
	d.content.AddChild(d.samples)
	d.content.AddChild(newNameGeneratorPanel(d, d.model, nil))
}
