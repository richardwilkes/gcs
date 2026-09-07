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
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/xfilepath"
	"github.com/richardwilkes/unison"
)

// FileBackedDockable defines methods a Dockable that is based on a file should implement.
type FileBackedDockable interface {
	unison.Dockable
	unison.TabCloser
	BackingFilePath() string
	SetBackingFilePath(p string)
}

// fileLoadable is a dockable whose content can be loaded from a file, which openDockableFromFile tells it about.
type fileLoadable interface {
	unison.Dockable
	markLoadedFromFile()
}

// fileBackedPanel is the root panel of a dockable that shows a file. It holds the path the content came from, or will
// be saved to, and, for editable content, the hash of that content as of the last load or save. It provides the
// FileBackedDockable, KeyedDockable and unison.TabCloser methods those dockables share, so each only has to define the
// ones it does differently.
//
// The unison.Panel lives inside it rather than beside it in the dockable: the panel has a Tooltip field and a String
// method, and if the two were embedded side by side, the Tooltip and String methods here would be ambiguous with them
// and so not be promoted to the dockable at all.
type fileBackedPanel struct {
	unison.Panel
	// dockable is the dockable this is embedded in, which the shared methods need for the calls that take the whole
	// dockable, such as updating its title, and for the methods the dockable overrides, such as BackingFilePath.
	dockable  FileBackedDockable
	path      string
	extension string
	// saver writes the content to the given path. It is nil for the dockables that only view a file.
	saver func(filePath string) error
	// hashable is the content whose hash decides whether there are unsaved changes. It is nil for the dockables that
	// only view a file, which are never modified.
	hashable          gurps.Hashable
	hash              uint64
	needsSaveAsPrompt bool
}

// initFileViewer sets up the panel for a dockable that only views the file at the given path.
func (p *fileBackedPanel) initFileViewer(dockable FileBackedDockable, filePath string) {
	p.dockable = dockable
	p.path = filePath
}

// initFileEditor sets up the panel for a dockable that edits the given content, which the saver writes to a file with
// the given extension. The content must be complete, since its hash is taken now. The dockable starts out needing a
// Save As prompt, since the content may not exist on disk yet; openDockableFromFile clears that for the ones it loads.
func (p *fileBackedPanel) initFileEditor(dockable FileBackedDockable, filePath, extension string, saver func(filePath string) error, hashable gurps.Hashable) {
	p.initFileViewer(dockable, filePath)
	p.extension = extension
	p.saver = saver
	p.hashable = hashable
	p.hash = gurps.Hash64(hashable)
	p.needsSaveAsPrompt = true
}

// DockKey implements KeyedDockable.
func (p *fileBackedPanel) DockKey() string {
	return filePrefix + p.path
}

// TitleIcon implements unison.Dockable.
func (p *fileBackedPanel) TitleIcon(suggestedSize geom.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  gurps.FileInfoFor(p.path).SVG,
		Size: suggestedSize,
	}
}

// Title implements unison.Dockable.
func (p *fileBackedPanel) Title() string {
	return xfilepath.BaseName(p.dockable.BackingFilePath())
}

func (p *fileBackedPanel) String() string {
	return p.dockable.Title()
}

// Tooltip implements unison.Dockable.
func (p *fileBackedPanel) Tooltip() string {
	return p.dockable.BackingFilePath()
}

// BackingFilePath implements FileBackedDockable.
func (p *fileBackedPanel) BackingFilePath() string {
	return p.path
}

// SetBackingFilePath implements FileBackedDockable.
func (p *fileBackedPanel) SetBackingFilePath(filePath string) {
	p.path = filePath
	UpdateTitleForDockable(p.dockable)
}

// Modified implements unison.Dockable.
func (p *fileBackedPanel) Modified() bool {
	return p.hashable != nil && p.hash != gurps.Hash64(p.hashable)
}

// MayAttemptClose implements unison.TabCloser.
func (p *fileBackedPanel) MayAttemptClose() bool {
	return MayAttemptCloseOfGroup(p.dockable)
}

// AttemptClose implements unison.TabCloser.
func (p *fileBackedPanel) AttemptClose() bool {
	if AttemptSaveForDockable(p.dockable) {
		return AttemptCloseForDockable(p.dockable)
	}
	return false
}

// markLoadedFromFile records that the content is the file at the path, so that saving it need not ask where it goes.
func (p *fileBackedPanel) markLoadedFromFile() {
	p.needsSaveAsPrompt = false
}

// save writes the content to its file, first asking where that should be if forceSaveAs is set or the content has never
// been saved. It returns true if the content was saved.
func (p *fileBackedPanel) save(forceSaveAs bool) bool {
	var success bool
	if forceSaveAs || p.needsSaveAsPrompt {
		success = SaveDockableAs(p.dockable, p.extension, p.saver, func(filePath string) {
			p.markUnmodified()
			p.path = filePath
		})
	} else {
		success = SaveDockable(p.dockable, p.saver, p.markUnmodified)
	}
	if success {
		p.needsSaveAsPrompt = false
	}
	return success
}

func (p *fileBackedPanel) markUnmodified() {
	p.hash = gurps.Hash64(p.hashable)
}

// openDockableFromFile loads the content of the file at the path with load, creates the dockable for it with create,
// and marks the dockable as having been loaded from that file, so that saving it need not ask where it goes.
func openDockableFromFile[M any, D fileLoadable](filePath string, load func(fileSystem fs.FS, filePath string) (M, error), create func(filePath string, content M) D) (unison.Dockable, error) {
	content, err := load(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	d := create(filePath, content)
	d.markLoadedFromFile()
	return d, nil
}

// unsavedFileName returns the file name a sheet that has never been saved goes by: its name, or the fallback when it
// has none, with the extension.
func unsavedFileName(name, fallback, ext string) string {
	if name = strings.TrimSpace(name); name == "" {
		name = fallback
	}
	return name + ext
}
