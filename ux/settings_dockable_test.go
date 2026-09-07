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
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestSettingsDockableDoLoadPrefersRefLoader verifies that doLoad hands the whole file reference to RefLoader when one
// is set, falls back to Loader with the reference's file system and path when there is no RefLoader, and reports a
// loader's failure through the workspace's error handler under the dockable's title.
func TestSettingsDockableDoLoadPrefersRefLoader(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	ref := &gurps.NamedFileRef{
		Name:       "Elf",
		FileSystem: os.DirFS(dir),
		FilePath:   "Elf.ancestry",
		DiskPath:   filepath.Join(dir, "Elf.ancestry"),
	}
	var loaderCalls, refLoaderCalls int
	var gotFS fs.FS
	var gotPath string
	var gotRef *gurps.NamedFileRef
	d := &SettingsDockable{
		TabTitle: "Ancestry",
		Loader: func(fileSystem fs.FS, filePath string) error {
			loaderCalls++
			gotFS = fileSystem
			gotPath = filePath
			return nil
		},
		RefLoader: func(ref *gurps.NamedFileRef) error {
			refLoaderCalls++
			gotRef = ref
			return nil
		},
	}

	// With both set, RefLoader wins and receives the reference itself.
	d.doLoad(ref)
	c.Equal(1, refLoaderCalls, "RefLoader is preferred")
	c.Equal(0, loaderCalls, "Loader is not called when RefLoader is set")
	c.True(gotRef == ref, "RefLoader receives the same reference")

	// With only Loader set, it receives the reference's file system and path.
	d.RefLoader = nil
	d.doLoad(ref)
	c.Equal(1, loaderCalls, "Loader is used when there is no RefLoader")
	c.Equal(ref.FileSystem, gotFS)
	c.Equal(ref.FilePath, gotPath)

	// A failing loader is reported through the workspace's error handler.
	var gotMsg string
	var gotErr error
	swapForTest(t, &Workspace.ErrorHandler, func(msg string, err error) {
		gotMsg = msg
		gotErr = err
	})
	boom := errors.New("boom")
	d.RefLoader = func(_ *gurps.NamedFileRef) error { return boom }
	d.doLoad(ref)
	c.Equal("Unable to load Ancestry", gotMsg)
	c.True(errors.Is(gotErr, boom), "the loader's error is passed through")
	c.Equal(1, loaderCalls, "Loader is not called when RefLoader fails")
}

// TestSettingsDockableCanLoad verifies that a dockable can load with either form of loader, and not without one.
func TestSettingsDockableCanLoad(t *testing.T) {
	c := check.New(t)
	d := &SettingsDockable{}
	c.False(d.canLoad(), "no loader")
	d.Loader = func(_ fs.FS, _ string) error { return nil }
	c.True(d.canLoad(), "Loader only")
	d.RefLoader = func(_ *gurps.NamedFileRef) error { return nil }
	c.True(d.canLoad(), "both loaders")
	d.Loader = nil
	c.True(d.canLoad(), "RefLoader only")
}

// TestShowSettingsViewsOpenOnceAndActivate drives every entry point that goes through initSettings inside a headless
// workspace: each opens its view with the title, extension, loader, saver, resetter and toolbar the spec named, and a
// second call activates the open view rather than opening another. The sheet settings are also opened for a sheet, to
// check that the owner is part of a per-sheet view's identity: the defaults view and the sheet's must both stay open.
func TestShowSettingsViewsOpenOnceAndActivate(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	user := gurps.GlobalSettings().Libraries.User()
	type view struct {
		name          string
		show          func()
		is            func(unison.Dockable) bool
		title         string
		ext           string
		loads         bool
		resets        bool
		toolbarButton bool
		asksOnClose   bool
	}
	views := []view{
		{
			name:   "colors",
			show:   ShowColorSettings,
			is:     func(d unison.Dockable) bool { _, isOne := d.AsPanel().Self.(*colorSettingsDockable); return isOne },
			title:  "Colors",
			ext:    gurps.ColorSettingsExt,
			loads:  true,
			resets: true,
			// The color mode popup is in the toolbar.
			toolbarButton: true,
		},
		{
			name:   "fonts",
			show:   ShowFontSettings,
			is:     func(d unison.Dockable) bool { _, isOne := d.AsPanel().Self.(*fontSettingsDockable); return isOne },
			title:  "Fonts",
			ext:    gurps.FontSettingsExt,
			loads:  true,
			resets: true,
		},
		{
			name:   "menu keys",
			show:   ShowMenuKeySettings,
			is:     func(d unison.Dockable) bool { _, isOne := d.AsPanel().Self.(*menuKeySettingsDockable); return isOne },
			title:  "Menu Keys",
			ext:    gurps.KeySettingsExt,
			loads:  true,
			resets: true,
		},
		{
			name:          "general",
			show:          ShowGeneralSettings,
			is:            func(d unison.Dockable) bool { _, isOne := d.AsPanel().Self.(*generalSettingsDockable); return isOne },
			title:         "General Settings",
			ext:           gurps.GeneralSettingsExt,
			loads:         true,
			resets:        true,
			toolbarButton: true,
			asksOnClose:   true,
		},
		{
			name:          "page reference mappings",
			show:          ShowPageRefMappings,
			is:            func(d unison.Dockable) bool { return asPageRefMappingsDockable(d) != nil },
			title:         "Page Reference Mappings",
			ext:           gurps.PageRefSettingsExt,
			loads:         true,
			resets:        true,
			toolbarButton: true,
		},
		{
			name: "default sheet settings",
			show: func() { ShowSheetSettings(nil) },
			is: func(d unison.Dockable) bool {
				s, isOne := d.AsPanel().Self.(*sheetSettingsDockable)
				return isOne && s.owner == nil
			},
			title:         "Default Sheet Settings",
			ext:           gurps.SheetSettingsExt,
			loads:         true,
			resets:        true,
			toolbarButton: true,
		},
		{
			name: "sheet's sheet settings",
			show: func() { ShowSheetSettings(sheet) },
			is: func(d unison.Dockable) bool {
				s, isOne := d.AsPanel().Self.(*sheetSettingsDockable)
				return isOne && s.owner == sheet
			},
			title:         sheetSettingsTabTitle(sheet),
			ext:           gurps.SheetSettingsExt,
			loads:         true,
			resets:        true,
			toolbarButton: true,
		},
		{
			name: "library",
			show: func() { ShowLibrarySettings(user) },
			is: func(d unison.Dockable) bool {
				l, isOne := d.AsPanel().Self.(*librarySettingsDockable)
				return isOne && l.library == user
			},
			title:         librarySettingsTitle(user.Config().Title),
			toolbarButton: true,
			asksOnClose:   true,
		},
	}
	for _, one := range views {
		screen.Do(one.show)
		d := soleEditor[unison.Dockable](t, screen, one.is)
		var base *SettingsDockable
		var isSettings bool
		var title string
		var exts []string
		var loads, saves, resets, asksOnClose, hasToolbarButton bool
		screen.Do(func() {
			// The view embeds the base, so the base is the view's own panel rather than an ancestor of it.
			base, isSettings = settingsDockableOf(d)
			if !isSettings {
				return
			}
			title = d.Title()
			exts = base.Extensions
			loads = base.Loader != nil
			saves = base.Saver != nil
			resets = base.Resetter != nil
			asksOnClose = base.WillCloseCallback != nil
			hasToolbarButton = toolbarStartsWithViewContent(base)
		})
		c.True(isSettings, "%s: the view is a settings dockable", one.name)
		c.Equal(one.title, title, "%s: title", one.name)
		if one.ext != "" {
			c.Equal([]string{one.ext}, exts, "%s: extensions", one.name)
		} else {
			c.Equal(0, len(exts), "%s: no extensions", one.name)
		}
		c.Equal(one.loads, loads, "%s: loader", one.name)
		c.Equal(one.loads, saves, "%s: saver", one.name)
		c.Equal(one.resets, resets, "%s: resetter", one.name)
		c.Equal(one.asksOnClose, asksOnClose, "%s: will-close callback", one.name)
		c.Equal(one.toolbarButton, hasToolbarButton, "%s: start-of-toolbar content", one.name)
	}
	// Opening every view again activates the one already open rather than adding another, and the view asked for last
	// is the one showing.
	for _, one := range views {
		screen.Do(one.show)
		d := soleEditor[unison.Dockable](t, screen, one.is)
		var current bool
		screen.Do(func() {
			if dc := unison.Ancestor[*unison.DockContainer](d.AsPanel()); dc != nil {
				if cur := dc.CurrentDockable(); cur != nil {
					current = cur.AsPanel().Self == d.AsPanel().Self
				}
			}
		})
		c.True(current, "%s: showing it again activates the open view", one.name)
	}
	var total int
	screen.Do(func() {
		for _, d := range AllDockables() {
			if _, isSettings := settingsDockableOf(d); isSettings {
				total++
			}
		}
	})
	c.Equal(len(views), total, "each view is open exactly once")
}

// settingsDockableOf returns the SettingsDockable a dockable embeds, if it is one of the settings views.
func settingsDockableOf(d unison.Dockable) (*SettingsDockable, bool) {
	switch v := d.AsPanel().Self.(type) {
	case *colorSettingsDockable:
		return &v.SettingsDockable, true
	case *fontSettingsDockable:
		return &v.SettingsDockable, true
	case *menuKeySettingsDockable:
		return &v.SettingsDockable, true
	case *generalSettingsDockable:
		return &v.SettingsDockable, true
	case *pageRefMappingsDockable:
		return &v.SettingsDockable, true
	case *sheetSettingsDockable:
		return &v.SettingsDockable, true
	case *librarySettingsDockable:
		return &v.SettingsDockable, true
	default:
		return nil, false
	}
}

// toolbarStartsWithViewContent reports whether the settings view's toolbar starts with something the view itself
// added. What the base adds -- the reset and menu buttons -- comes after a grabbing spacer when the view added
// anything, and is all there is otherwise, so a toolbar whose first child is the spacer, or which is empty, holds
// nothing from the view.
func toolbarStartsWithViewContent(base *SettingsDockable) bool {
	children := base.Children()[0].Children()
	if len(children) == 0 {
		return false
	}
	data, ok := children[0].LayoutData().(*unison.FlexLayoutData)
	return !ok || !data.HGrab || children[0].Self != children[0]
}
