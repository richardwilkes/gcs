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
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xfilepath"
	"github.com/richardwilkes/unison"
)

// fileEditorModel is what a fileEditorDockable needs of the model it edits.
type fileEditorModel[T any] interface {
	gurps.Hashable
	Clone() T
	Save(filePath string) error
	Normalize()
	ResetTargetKeyPrefixes(prefixProvider func() string)
}

// fileEditorSpec describes one kind of file editor: the file type it edits, how to make and read its model, and how to
// build its content. It is what distinguishes the ancestry editor from the name generator editor; everything else about
// them is in fileEditorDockable.
type fileEditorSpec[T fileEditorModel[T]] struct {
	// tabTitle is the editor's name, such as "Ancestry", used in the tab title and the base's own prompts.
	tabTitle string
	// icon is the tab icon.
	icon *unison.SVG
	// ext is the file extension, including the leading dot.
	ext string
	// newModel returns a new, blank model.
	newModel func() T
	// readModel reads a model from a file.
	readModel func(fileSystem fs.FS, filePath string) (T, error)
	// buildContent fills the editor's content panel from its model. It is called for the initial build and again by
	// every sync().
	buildContent func()
	// fallbackName, which may be nil, returns a name for a model that has neither a file nor a loaded name, such as the
	// name typed into an ancestry. An empty result means there is none.
	fallbackName func() string
	// openRef opens the file the reference names in an editor of this kind, activating the editor already showing it if
	// there is one; see openFileEditor. It is what the toolbar menu's Open… and library entries do.
	openRef func(ref *library.NamedFileRef) error
}

// fileEditorDockable is the common part of the editors that edit a settings file of their own -- the ancestry and name
// generator editors. It is a settings dockable, sharing the Reset button and the toolbar menu of every library's files
// with the attribute and body type editors, but since it edits a file of its own rather than a per-sheet setting, it
// also tracks the file it was loaded from and saves back to it, which makes it a FileBackedDockable as well. Each
// editor holds one document for its whole life, as a sheet does: File > New opens another editor rather than replacing
// what one holds, and opening a file finds the editor already showing it or opens one of its own. Every change to the
// model goes through an undo edit that captures the whole model, so a structural change is as undoable as a typed one.
type fileEditorDockable[T fileEditorModel[T]] struct {
	SettingsDockable
	structuralEditorBase[T]
	spec fileEditorSpec[T]
	// path is the absolute path the model was loaded from or last saved to, or "" when it is not on disk yet.
	path string
	// loadedName is the name of the file the model was loaded from when that file has no disk path, as a built-in one
	// does not, so that the title and the Save As file name can still show which one it is. It is "" otherwise.
	loadedName string
	saveButton *unison.Button
}

// init readies the editor to hold a new, blank model. The outer dockable calls it after setting Self, which is what the
// rows name as their editor, so a payload can be recognized as this editor's own; load may then replace the blank model
// with a file's, before show builds the toolbar and content and places the editor in the dock.
func (d *fileEditorDockable[T]) init(spec fileEditorSpec[T]) {
	d.spec = spec
	editor, ok := d.Self.(rowDragEditor)
	if !ok {
		editor = d
	}
	d.initEditor(editor, func(model T) T { return model.Clone() }, spec.buildContent)
	model := d.spec.newModel()
	model.ResetTargetKeyPrefixes(d.targetMgr.NextPrefix)
	d.setModel(model)
	d.TabTitle = spec.tabTitle
	d.TabIcon = spec.icon
	d.Extensions = []string{spec.ext}
	d.LoadItemTitle = i18n.Text("Open…")
	d.RefLoader = spec.openRef
	d.Resetter = d.reset
	d.ModifiedCallback = d.modified
	d.InstallCmdHandlers(SaveItemID, func(_ any) bool { return d.Modified() }, func(_ any) { d.save(false) })
	d.InstallCmdHandlers(SaveAsItemID, unison.AlwaysEnabled, func(_ any) { d.save(true) })
}

// show builds the toolbar and content from the model as it stands and places the editor in the dock, so a file to be
// edited is loaded first.
func (d *fileEditorDockable[T]) show() {
	d.Setup(d.addToStartToolbar, nil, d.initContent)
}

// Title implements unison.Dockable. The dock resolves the embedded SettingsDockable to this dockable, so this, rather
// than the base's TabTitle, is what the tab shows. The TabTitle is still what the base uses in its own prompts.
func (d *fileEditorDockable[T]) Title() string {
	return fmt.Sprintf("%s: %s", d.TabTitle, d.displayName())
}

// displayName returns the name the model is known by: the base name of its file once it has one, since that is what
// ancestries and traits select these files by; otherwise the name of the built-in file it was loaded from; otherwise
// whatever the spec's fallback offers; and failing that, a placeholder.
func (d *fileEditorDockable[T]) displayName() string {
	if d.path != "" {
		return xfilepath.BaseName(d.path)
	}
	if d.loadedName != "" {
		return d.loadedName
	}
	if d.spec.fallbackName != nil {
		if name := d.spec.fallbackName(); name != "" {
			return name
		}
	}
	return i18n.Text("Untitled")
}

// Tooltip implements unison.Dockable.
func (d *fileEditorDockable[T]) Tooltip() string {
	return d.path
}

// BackingFilePath implements FileBackedDockable. Until the model has been saved, this is a relative pseudo-path made
// from its display name, as a new sheet's is: no such file exists, so the workspace offers to save it when closing, and
// Save As uses it for the initial file name.
func (d *fileEditorDockable[T]) BackingFilePath() string {
	if d.path != "" {
		return d.path
	}
	return xfilepath.SanitizeName(d.displayName()) + d.spec.ext
}

// SetBackingFilePath implements FileBackedDockable.
func (d *fileEditorDockable[T]) SetBackingFilePath(p string) {
	d.path = p
	UpdateTitleForDockable(d)
}

// MarkModified implements ModifiableRoot. It does what the base does, but hands this dockable rather than the embedded
// base to the title update, so that a window of its own shows this dockable's title rather than the base's TabTitle.
func (d *fileEditorDockable[T]) MarkModified(_ unison.Paneler) {
	d.Modified()
	UpdateTitleForDockable(d)
	DeepSync(d)
}

// AttemptClose implements unison.TabCloser. Offering to save here, rather than through the base's WillCloseCallback,
// is what gives every path through the workspace's close logic exactly one prompt.
func (d *fileEditorDockable[T]) AttemptClose() bool {
	if AttemptSaveForDockable(d) {
		return AttemptCloseForDockable(d)
	}
	return false
}

func (d *fileEditorDockable[T]) modified() bool {
	modified := d.modelModified()
	if d.saveButton != nil {
		d.saveButton.SetEnabled(modified)
	}
	return modified
}

func (d *fileEditorDockable[T]) addToStartToolbar(toolbar *unison.Panel) {
	d.toolbar = toolbar
	d.saveButton = unison.NewSVGButton(svg.Save)
	d.saveButton.Tooltip = newWrappedTooltip(i18n.Text("Save"))
	d.saveButton.SetEnabled(false)
	d.saveButton.ClickCallback = func() { d.save(false) }
	toolbar.AddChild(d.saveButton)

	saveAsButton := unison.NewSVGButton(svg.SaveAs)
	saveAsButton.Tooltip = newWrappedTooltip(i18n.Text("Save As…"))
	saveAsButton.ClickCallback = func() { d.save(true) }
	toolbar.AddChild(saveAsButton)
}

// save writes the model to its file, prompting for one when it has none yet or when a Save As was asked for. A new
// file is offered the User Library's ancestries folder, which is where ancestries and the name generators they use
// live, and the file name is taken from the display name.
func (d *fileEditorDockable[T]) save(forceSaveAs bool) bool {
	if forceSaveAs || d.path == "" {
		return saveDockableAs(d, d.spec.ext, func() string { return gurps.GlobalSettings().Libraries.User().AncestriesPath() },
			gurps.SettingsLastDirKey, d.model.Save, func(p string) {
				d.path = p
				d.loadedName = ""
				d.markSaved()
			})
	}
	return SaveDockable(d, d.model.Save, d.markSaved)
}

// markSaved records that the model as it stands is what its file holds.
func (d *fileEditorDockable[T]) markSaved() {
	d.originalHash = gurps.Hash64(d.model)
	d.modified() // refreshes the Save button
}

// load fills the editor with the model in the file, in place of the blank one init gave it. It is for an editor that
// has not been shown yet, so nothing is rebuilt and nothing is posted to the undo stack: the editor starts out holding
// the file's model, unmodified, with nothing to undo, as a sheet opened from a file does. A file that fails to load
// changes nothing. The recorded path is empty for a built-in file, so that Save on one of those prompts for a location
// rather than trying to write into the application.
func (d *fileEditorDockable[T]) load(ref *library.NamedFileRef) error {
	model, err := d.spec.readModel(ref.FileSystem, ref.FilePath)
	if err != nil {
		return err
	}
	model.Normalize()
	model.ResetTargetKeyPrefixes(d.targetMgr.NextPrefix)
	d.setModel(model)
	d.path = ref.DiskPath
	if d.path == "" {
		d.loadedName = ref.Name
	} else {
		d.loadedName = ""
	}
	return nil
}

// showsFile reports whether the editor holds the file the reference names, whether that is a file on disk or a
// built-in one.
func (d *fileEditorDockable[T]) showsFile(ref *library.NamedFileRef) bool {
	if ref.DiskPath != "" {
		return d.path == ref.DiskPath
	}
	return d.path == "" && d.loadedName == ref.Name
}

// reset replaces the model with a new, blank one, undoably. The editor keeps its file: a reset is an edit to what the
// file holds, to be saved or not, rather than a change of document, so the editor shows as modified until it is saved
// or undone.
func (d *fileEditorDockable[T]) reset() {
	d.editStructure(fmt.Sprintf(i18n.Text("Reset %s"), d.TabTitle), func() {
		d.model = d.spec.newModel()
		d.model.ResetTargetKeyPrefixes(d.targetMgr.NextPrefix)
	}, "")
}

// diskFileRef returns a reference to a file on disk of the kind a library scan produces, so that a file chosen by the
// user can be loaded the same way as one from a library. The path is made absolute when it can be.
func diskFileRef(filePath string) *library.NamedFileRef {
	if abs, err := filepath.Abs(filePath); err == nil {
		filePath = abs
	}
	dir := filepath.Dir(filePath)
	return &library.NamedFileRef{
		Name:       xfilepath.BaseName(filePath),
		FileSystem: os.DirFS(dir),
		FilePath:   filepath.Base(filePath),
		DiskPath:   filePath,
	}
}

// findDockable returns the first open dockable of the given type that match accepts, and whether there was one.
func findDockable[T unison.Dockable](match func(T) bool) (T, bool) {
	for _, d := range AllDockables() {
		if candidate, ok := d.AsPanel().Self.(T); ok && match(candidate) {
			return candidate, true
		}
	}
	var zero T
	return zero, false
}

// fileEditor is what openFileEditor needs of an editor. The ancestry and name generator editors both provide it through
// their embedded fileEditorDockable.
type fileEditor interface {
	unison.Dockable
	load(ref *library.NamedFileRef) error
	show()
	showsFile(ref *library.NamedFileRef) bool
}

// openFileEditor opens the file the reference names in an editor of the kind newEditor makes. An editor already showing
// the file is activated and returned rather than a second one being opened, so a file is never open in two editors at
// once. Otherwise the file is loaded into a new editor, which is shown only if the load succeeds: a file that fails to
// load opens nothing, and the error is returned for the caller to report.
func openFileEditor[D fileEditor](ref *library.NamedFileRef, newEditor func() D) (D, error) {
	if d, ok := findDockable(func(d D) bool { return d.showsFile(ref) }); ok {
		ActivateDockable(d)
		return d, nil
	}
	d := newEditor()
	if err := d.load(ref); err != nil {
		var zero D
		return zero, err
	}
	d.show()
	return d, nil
}
