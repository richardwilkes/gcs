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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xfilepath"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
)

var (
	_ unison.Dockable  = &SettingsDockable{}
	_ unison.TabCloser = &SettingsDockable{}
)

// SettingsDockable holds common settings dockable data.
type SettingsDockable struct {
	unison.Panel
	TabTitle   string
	TabIcon    *unison.SVG
	Extensions []string
	Loader     func(fileSystem fs.FS, filePath string) error
	// RefLoader is an alternative to Loader for dockables that need to know where the file lives on disk. When set, it
	// is used in preference to Loader.
	RefLoader func(ref *gurps.NamedFileRef) error
	// LoadItemTitle is the title of the toolbar menu's item that loads a file chosen in a dialog. When empty, it is
	// Import…, which suits a dockable that takes the file's contents into itself; the file editors, which open the
	// chosen file in an editor of its own, call it Open… instead.
	LoadItemTitle     string
	Saver             func(filePath string) error
	Resetter          func()
	ModifiedCallback  func() bool
	WillCloseCallback func() bool
}

// settingsSpec is what distinguishes one of the views that edit a global setting in place -- the colors, fonts, menu
// keys, general settings, page reference mappings, sheet defaults and library settings -- from another; everything else
// about them is in SettingsDockable. Their tab icon is always the settings icon, so the spec does not name one. Any of
// the functions may be nil, in which case the base omits what it would have done with it: with no loader, saver or
// resetter the toolbar has no menu or reset button, and with no willClose the view closes without being asked.
type settingsSpec struct {
	title             string
	ext               string
	loader            func(fileSystem fs.FS, filePath string) error
	saver             func(filePath string) error
	resetter          func()
	willClose         func() bool
	addToStartToolbar func(toolbar *unison.Panel)
	initContent       func(content *unison.Panel)
}

// initSettings fills in the base from the spec, with self, the outer view, as what the dock resolves the panel to, then
// builds the toolbar and content and places the view in the dock. The caller must already have checked that the view is
// not open, with activateDockable or a predicate of its own, since the base cannot tell one view from another.
func (d *SettingsDockable) initSettings(self unison.Paneler, spec *settingsSpec) {
	d.Self = self
	d.TabTitle = spec.title
	d.TabIcon = svg.Settings
	if spec.ext != "" {
		d.Extensions = []string{spec.ext}
	}
	d.Loader = spec.loader
	d.Saver = spec.saver
	d.Resetter = spec.resetter
	d.WillCloseCallback = spec.willClose
	d.Setup(spec.addToStartToolbar, nil, spec.initContent)
}

// initSettingsContent gives a settings view's content panel the layout most of them share, a grid of the given number
// of columns with the standard spacing, and returns it so that a view keeping a reference to its content can take it
// from the same call.
func initSettingsContent(content *unison.Panel, columns int) *unison.Panel {
	content.SetLayout(&unison.FlexLayout{
		Columns:  columns,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	return content
}

// Setup the dockable and display it.
func (d *SettingsDockable) Setup(addToStartToolbar, addToEndToolbar, initContent func(*unison.Panel)) {
	d.SetLayout(&unison.FlexLayout{Columns: 1})
	toolbar := d.createToolbar(addToStartToolbar, addToEndToolbar)
	d.AddChild(toolbar)
	content := unison.NewPanel()
	content.SetBorder(unison.NewEmptyBorder(geom.NewUniformInsets(unison.StdHSpacing * 2)))
	initContent(content)
	scroller := unison.NewScrollPanel()
	scroller.SetContent(content, behavior.Fill, behavior.Fill)
	scroller.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	d.AddChild(scroller)
	PlaceInDock(d, dgroup.Settings, false)
	FocusFirstContent(toolbar, content)
}

// TitleIcon implements unison.Dockable
func (d *SettingsDockable) TitleIcon(suggestedSize geom.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  d.TabIcon,
		Size: suggestedSize,
	}
}

// Title implements unison.Dockable
func (d *SettingsDockable) Title() string {
	return d.TabTitle
}

// Tooltip implements unison.Dockable
func (d *SettingsDockable) Tooltip() string {
	return ""
}

// Modified implements unison.Dockable
func (d *SettingsDockable) Modified() bool {
	if d.ModifiedCallback == nil {
		return false
	}
	return d.ModifiedCallback()
}

// MarkModified implements ModifiableRoot.
func (d *SettingsDockable) MarkModified(_ unison.Paneler) {
	d.Modified()
	UpdateTitleForDockable(d)
	DeepSync(d)
}

// MayAttemptClose implements unison.TabCloser
func (d *SettingsDockable) MayAttemptClose() bool {
	return MayAttemptCloseOfGroup(d)
}

// AttemptClose implements unison.TabCloser
func (d *SettingsDockable) AttemptClose() bool {
	if !CloseGroup(d) {
		return false
	}
	if d.WillCloseCallback != nil {
		if !d.WillCloseCallback() {
			return false
		}
	}
	return AttemptCloseForDockable(d)
}

func (d *SettingsDockable) createToolbar(addToStartToolbar, addToEndToolbar func(*unison.Panel)) *unison.Panel {
	toolbar := newToolbar()
	if addToStartToolbar != nil {
		addToStartToolbar(toolbar)
	}
	index := len(toolbar.Children())
	if addToEndToolbar != nil {
		addToEndToolbar(toolbar)
	}
	if d.Resetter != nil {
		b := unison.NewSVGButton(svg.Reset)
		b.Tooltip = newWrappedTooltip(i18n.Text("Reset"))
		b.ClickCallback = d.handleReset
		toolbar.AddChild(b)
	}
	if d.canLoad() || d.Saver != nil {
		b := unison.NewSVGButton(svg.Menu)
		b.Tooltip = newWrappedTooltip(i18n.Text("Menu"))
		b.ClickCallback = func() { d.showMenu(b) }
		toolbar.AddChild(b)
	}
	if len(toolbar.Children()) != index {
		spacer := unison.NewPanel()
		spacer.SetLayoutData(&unison.FlexLayoutData{HGrab: true})
		toolbar.AddChildAtIndex(spacer, index)
	}
	finishToolbarLayout(toolbar)
	return toolbar
}

func (d *SettingsDockable) handleReset() {
	if unison.QuestionDialog(fmt.Sprintf(i18n.Text("Are you sure you want to reset the\n%s?"), d.TabTitle), "") == unison.ModalResponseOK {
		d.Resetter()
	}
}

func (d *SettingsDockable) showMenu(b *unison.Button) {
	f := unison.DefaultMenuFactory()
	id := unison.ContextMenuIDFlag
	m := f.NewMenu(id, "", nil)
	id++
	if d.canLoad() {
		m.InsertItem(-1, f.NewItem(id, d.loadItemTitle(), unison.KeyBinding{}, nil, d.handleImport))
		id++
	}
	if d.Saver != nil {
		m.InsertItem(-1, f.NewItem(id, i18n.Text("Export…"), unison.KeyBinding{}, nil, d.handleExport))
		id++
	}
	if d.canLoad() {
		libraries := gurps.GlobalSettings().Libraries
		sets := gurps.ScanForNamedFileSets(nil, "", false, libraries, d.Extensions...)
		if len(sets) != 0 {
			m.InsertSeparator(-1, false)
			for _, lib := range sets {
				m.InsertItem(-1, f.NewItem(id, lib.Name, unison.KeyBinding{},
					func(_ unison.MenuItem) bool { return false }, nil))
				id++
				for _, one := range lib.List {
					d.insertFileToLoad(m, id, one)
					id++
				}
			}
		}
	}
	m.Popup(b.RectToRoot(b.ContentRect(true)), 0)
}

func (d *SettingsDockable) insertFileToLoad(m unison.Menu, id int, ref *gurps.NamedFileRef) {
	m.InsertItem(-1, m.Factory().NewItem(id, "    "+ref.Name, unison.KeyBinding{}, nil, func(_ unison.MenuItem) {
		d.doLoad(ref)
	}))
}

// loadItemTitle returns LoadItemTitle, or Import… when none was given.
func (d *SettingsDockable) loadItemTitle() string {
	if d.LoadItemTitle != "" {
		return d.LoadItemTitle
	}
	return i18n.Text("Import…")
}

// canLoad reports whether the dockable was given either form of loader.
func (d *SettingsDockable) canLoad() bool {
	return d.Loader != nil || d.RefLoader != nil
}

// doLoad hands the whole file reference to RefLoader when one is set, and otherwise its file system and path to Loader.
// A failure is reported through the workspace's error handler.
func (d *SettingsDockable) doLoad(ref *gurps.NamedFileRef) {
	var err error
	if d.RefLoader != nil {
		err = d.RefLoader(ref)
	} else {
		err = d.Loader(ref.FileSystem, ref.FilePath)
	}
	if err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to load ")+d.TabTitle, err)
	}
}

func (d *SettingsDockable) handleImport(_ unison.MenuItem) {
	if filePath, ok := chooseFileToOpen(gurps.SettingsLastDirKey, d.Extensions...); ok {
		d.doLoad(diskFileRef(filePath))
	}
}

func (d *SettingsDockable) handleExport(_ unison.MenuItem) {
	if filePath, ok := chooseFileToSave(gurps.GlobalSettings().LastDir(gurps.SettingsLastDirKey),
		xfilepath.BaseName(d.Title()), d.Extensions[0], gurps.SettingsLastDirKey); ok {
		if err := d.Saver(filePath); err != nil {
			Workspace.ErrorHandler(i18n.Text("Unable to save ")+d.TabTitle, err)
		}
	}
}
