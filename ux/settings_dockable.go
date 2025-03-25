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
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	xfs "github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
)

// Known dockable kinds
const (
	SheetDockableKind     = "sheet"
	TemplateDockableKind  = "template"
	LootSheetDockableKind = "loot"
	ListDockableKind      = "list"
)

var (
	_ unison.Dockable  = &SettingsDockable{}
	_ unison.TabCloser = &SettingsDockable{}
)

// SettingsDockable holds common settings dockable data.
type SettingsDockable struct {
	unison.Panel
	TabTitle          string
	TabIcon           *unison.SVG
	Extensions        []string
	Loader            func(fileSystem fs.FS, filePath string) error
	Saver             func(filePath string) error
	Resetter          func()
	ModifiedCallback  func() bool
	WillCloseCallback func() bool
}

// Setup the dockable and display it.
func (d *SettingsDockable) Setup(addToStartToolbar, addToEndToolbar, initContent func(*unison.Panel)) {
	d.SetLayout(&unison.FlexLayout{Columns: 1})
	toolbar := d.createToolbar(addToStartToolbar, addToEndToolbar)
	d.AddChild(toolbar)
	content := unison.NewPanel()
	content.SetBorder(unison.NewEmptyBorder(unison.NewUniformInsets(unison.StdHSpacing * 2)))
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
func (d *SettingsDockable) TitleIcon(suggestedSize unison.Size) unison.Drawable {
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

// MarkModified implements widget.ModifiableRoot
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
	toolbar := unison.NewPanel()
	toolbar.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, 0, unison.Insets{Bottom: 1},
		false), unison.NewEmptyBorder(unison.StdInsets())))
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
	if d.Loader != nil || d.Saver != nil {
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
	toolbar.SetLayout(&unison.FlexLayout{
		Columns:  len(toolbar.Children()),
		HSpacing: unison.StdHSpacing,
	})
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
	if d.Loader != nil {
		m.InsertItem(-1, f.NewItem(id, i18n.Text("Import…"), unison.KeyBinding{}, nil, d.handleImport))
		id++
	}
	if d.Saver != nil {
		m.InsertItem(-1, f.NewItem(id, i18n.Text("Export…"), unison.KeyBinding{}, nil, d.handleExport))
		id++
	}
	if d.Loader != nil {
		libraries := gurps.GlobalSettings().Libraries()
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
		d.doLoad(ref.FileSystem, ref.FilePath)
	}))
}

func (d *SettingsDockable) doLoad(fileSystem fs.FS, filePath string) {
	if err := d.Loader(fileSystem, filePath); err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to load ")+d.TabTitle, err)
	}
}

func (d *SettingsDockable) handleImport(_ unison.MenuItem) {
	dialog := unison.NewOpenDialog()
	dialog.SetAllowsMultipleSelection(false)
	dialog.SetResolvesAliases(true)
	dialog.SetAllowedExtensions(d.Extensions...)
	dialog.SetCanChooseDirectories(false)
	dialog.SetCanChooseFiles(true)
	global := gurps.GlobalSettings()
	dialog.SetInitialDirectory(global.LastDir(gurps.SettingsLastDirKey))
	if dialog.RunModal() {
		p := dialog.Path()
		dir := filepath.Dir(p)
		global.SetLastDir(gurps.SettingsLastDirKey, dir)
		d.doLoad(os.DirFS(dir), filepath.Base(p))
	}
}

func (d *SettingsDockable) handleExport(_ unison.MenuItem) {
	dialog := unison.NewSaveDialog()
	dialog.SetAllowedExtensions(d.Extensions[0])
	global := gurps.GlobalSettings()
	dialog.SetInitialDirectory(global.LastDir(gurps.SettingsLastDirKey))
	dialog.SetInitialFileName(xfs.SanitizeName(xfs.BaseName(d.Title())))
	if dialog.RunModal() {
		if filePath, ok := unison.ValidateSaveFilePath(dialog.Path(), d.Extensions[0], false); ok {
			global.SetLastDir(gurps.SettingsLastDirKey, filepath.Dir(filePath))
			if err := d.Saver(filePath); err != nil {
				Workspace.ErrorHandler(i18n.Text("Unable to save ")+d.TabTitle, err)
			}
		}
	}
}
