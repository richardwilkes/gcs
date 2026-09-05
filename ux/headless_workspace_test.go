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
	"image/png"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
)

// startHeadlessWorkspace starts a headless unison session running the GCS workspace -- the menu bar, the navigator and
// the document dock in a single window -- exactly as Start does, minus the update checks and the handoff service. It
// returns the screen that drives the session and the workspace window. The session is stopped when the test ends, and
// the process-wide state the workspace touches is put back afterwards: the Workspace global, the settings path, the
// libraries (swapped for fresh temporary ones, see useTestLibraries), the recent files and last-used directories that
// saving and opening files record (see preserveRecentFilesAndLastDirs) and the workspace-restoration setting, which is
// turned off so that the dock does not try to restore whatever the settings hold.
//
// Stop clears every window's AllowCloseCallback and WillCloseCallback and stops every modal loop before it quits, so a
// dockable left open with unsaved changes cannot put up a save prompt that would hang the shutdown, and the workspace's
// own close handler, which saves the global settings, never runs. Even so, a test should leave the dockables it opened
// closed or unmodified, so that a failure elsewhere is not hidden behind whatever teardown makes of them.
//
// Anything the session recorded through Errors() -- a panic in a callback, a wait that was abandoned because the
// application never went quiet -- fails the test when it ends. Sessions run one at a time and own most of unison's
// mutable globals while they run, so a test using this must not call t.Parallel.
func startHeadlessWorkspace(t *testing.T, c check.Checker) (*unison.HeadlessScreen, *unison.Window) {
	t.Helper()
	savedWorkspace := Workspace
	t.Cleanup(func() { Workspace = savedWorkspace })
	savedSettingsPath := gurps.SettingsPath
	gurps.SettingsPath = filepath.Join(t.TempDir(), "settings.json")
	t.Cleanup(func() { gurps.SettingsPath = savedSettingsPath })
	general := gurps.GlobalSettings().General
	savedRestore := general.RestoreWorkspaceOnStart
	general.RestoreWorkspaceOnStart = false
	t.Cleanup(func() { general.RestoreWorkspaceOnStart = savedRestore })
	useTestLibraries(t, c)
	preserveRecentFilesAndLastDirs(t)
	// main registers the file types before starting the UI; the navigator looks its folder icons up in that registry.
	RegisterKnownFileTypes()

	var wnd *unison.Window
	screen, err := unison.StartHeadless(unison.HeadlessConfig{Width: 1400, Height: 900},
		unison.StartupFinishedCallback(func() {
			w, wndErr := unison.NewWindow("GCS")
			if wndErr != nil {
				t.Errorf("unable to create the workspace window: %v", wndErr)
				return
			}
			registerWindowDragTypes(w)
			SetupMenuBar(w)
			InitWorkspace(w)
			wnd = w
		}))
	if err != nil {
		t.Fatalf("unable to start the headless session: %v", err)
	}
	// Registered ahead of Stop so that it runs after the session has ended, by which time everything the session is
	// ever going to record has been recorded.
	t.Cleanup(func() {
		for _, one := range screen.Errors() {
			t.Errorf("the headless session recorded an error: %v", one)
		}
	})
	t.Cleanup(screen.Stop)
	if wnd == nil {
		t.Fatal("the workspace window was not created")
	}
	// The workspace reports errors with a modal dialog once it has finished initializing. A dialog nobody is going to
	// dismiss would leave the test stranded, so report them as test failures instead.
	screen.Do(func() {
		Workspace.ErrorHandler = func(msg string, err error) { t.Errorf("unexpected error: %s: %v", msg, err) }
	})
	return screen, wnd
}

// preserveRecentFilesAndLastDirs puts the global settings' recent files list and last-used directories back when the
// test ends. Saving or opening a file through the workspace records the file in the one and its directory in the
// other, and the global settings are process-wide, so a temporary path left in either would be offered to every test
// that runs afterwards.
func preserveRecentFilesAndLastDirs(t *testing.T) {
	t.Helper()
	global := gurps.GlobalSettings()
	savedRecentFiles := slices.Clone(global.RecentFiles)
	savedLastDirs := maps.Clone(global.LastDirs)
	t.Cleanup(func() {
		global.RecentFiles = savedRecentFiles
		global.LastDirs = savedLastDirs
	})
}

// captureScreen writes what the screen shows to name.png in the directory named by GCS_HEADLESS_CAPTURE_DIR, for the
// person running the test to look at. When nothing is named there, nobody is going to look, so nothing is captured.
func captureScreen(t *testing.T, c check.Checker, screen *unison.HeadlessScreen, name string) {
	t.Helper()
	dir := os.Getenv("GCS_HEADLESS_CAPTURE_DIR")
	if dir == "" {
		return
	}
	img := screen.Capture()
	if img == nil {
		t.Fatal("the screen could not be captured")
	}
	path := filepath.Join(dir, name+".png")
	f, err := os.Create(path) //nolint:gosec // G703: writing into the directory named in the environment is the point
	if err != nil {
		t.Fatalf("unable to create %s: %v", path, err)
	}
	c.NoError(png.Encode(f, img), "encoding %s", path)
	c.NoError(f.Close(), "closing %s", path)
}

// dragRowAheadOf drags the handle of the from'th child of rows to the upper half of the to'th, which asks for the row
// to be inserted ahead of that one, the way a user reorders an editor list. Both rows must be in view for the drag to
// land where it is aimed, so the span from a little above the target row to the dragged row's handle is scrolled into
// view first; the headroom keeps the pointer clear of the edge that the drop target scrolls at. The test fails if the
// span does not fit the view, so the two rows must be near enough to each other for it to.
func dragRowAheadOf(t *testing.T, screen *unison.HeadlessScreen, wnd *unison.Window, rows *unison.Panel, from, to int) {
	t.Helper()
	var handle *unison.Panel
	var target geom.Point
	var count int
	var spanFits, handleVisible, targetVisible bool
	screen.Do(func() {
		children := rows.Children()
		count = len(children)
		if from < 0 || from >= count || to < 0 || to >= count {
			return
		}
		handles := panelsOfType[*DragHandle](children[from])
		if len(handles) == 0 {
			return
		}
		handle = handles[0].AsPanel()
		targetRow := children[to]
		span := targetRow.FrameRect()
		span.Y -= 40
		span.Height += 40
		span = span.Union(rows.RectFromRoot(handle.RectToRoot(handle.ContentRect(false))))
		view := rows.ScrollRoot().ContentView()
		spanFits = span.Height <= view.ContentRect(false).Height
		rows.ScrollRectIntoView(span)
		rows.ValidateScrollRoot()
		handleVisible = fullyVisible(handle)
		targetRect := targetRow.RectToRoot(targetRow.ContentRect(false))
		visible := visibleRect(targetRow)
		targetVisible = visible.Y == targetRect.Y && visible.Height >= 16
		target = screenPoint(wnd, geom.NewPoint(visible.CenterX(), visible.Y+8))
	})
	if from < 0 || from >= count || to < 0 || to >= count {
		t.Fatalf("the list has %d rows, so row %d cannot be dragged ahead of row %d", count, from, to)
	}
	if handle == nil {
		t.Fatalf("row %d has no drag handle", from)
	}
	if !spanFits || !handleVisible || !targetVisible {
		t.Fatalf("rows %d and %d must both be in view for the drag (span fits: %v, handle visible: %v, "+
			"target visible: %v)", to, from, spanFits, handleVisible, targetVisible)
	}
	screen.Drag(screen.PanelCenter(handle), target, 10)
}

// buttonWithSVG returns the first button within root whose icon is the given SVG, or nil if there is none. Toolbar and
// row buttons in GCS have no title, so the icon is what identifies them.
func buttonWithSVG(root *unison.Panel, icon *unison.SVG) *unison.Button {
	for _, b := range panelsOfType[*unison.Button](root) {
		if drawable, ok := b.Drawable.(*unison.DrawableSVG); ok && drawable.SVG == icon {
			return b
		}
	}
	return nil
}

// buttonWithTooltip returns the first button within root whose tooltip reads exactly text, or nil if there is none.
// Where several buttons share an icon, the tooltip is what tells a user -- and so a test -- which is which.
func buttonWithTooltip(root *unison.Panel, text string) *unison.Button {
	for _, b := range panelsOfType[*unison.Button](root) {
		if b.Tooltip != nil && tooltipText(b.Tooltip) == text {
			return b
		}
	}
	return nil
}

// In a headless session the menus are unison's pure-Go, in-window kind. The window's root panel holds, in this order,
// any open popup menus (newest first), the menu bar, a tooltip if one is showing, and the content. A menu panel, bar or
// popup alike, holds one scroll panel whose content has one child panel per item of the menu, in the menu's own order
// and with the separators included, so the panel for an item is found by its index in the unison.Menu.

// menuBarPanel returns the in-window menu bar of wnd. It is the one child of the root, other than the content, that
// spans the full width of the window at the top; popups are packed to the width of their items.
func menuBarPanel(wnd *unison.Window) *unison.Panel {
	root := wnd.Content().Parent()
	if root == nil {
		return nil
	}
	width := root.FrameRect().Width
	for _, child := range root.Children() {
		if child == wnd.Content() {
			continue
		}
		if r := child.FrameRect(); r.X == 0 && r.Y == 0 && r.Width == width {
			return child
		}
	}
	return nil
}

// openMenuPopup returns the most recently opened popup menu in wnd, or nil if none is open.
func openMenuPopup(wnd *unison.Window) *unison.Panel {
	root := wnd.Content().Parent()
	if root == nil || len(root.Children()) == 0 {
		return nil
	}
	first := root.Children()[0]
	if first == wnd.Content() || first == menuBarPanel(wnd) {
		return nil
	}
	return first
}

// menuItemPanels returns the panels of the items of a menu panel, bar or popup, one per item in the menu's order.
func menuItemPanels(menuPanel *unison.Panel) []*unison.Panel {
	if menuPanel == nil || len(menuPanel.Children()) == 0 {
		return nil
	}
	scroller, ok := menuPanel.Children()[0].Self.(*unison.ScrollPanel)
	if !ok || scroller.Content() == nil {
		return nil
	}
	return scroller.Content().AsPanel().Children()
}

// chooseMenuBarItem chooses an item from one of the menus in wnd's menu bar the way a user would: it clicks the menu's
// title in the bar, then clicks the item in the popup that opens. Since the item's handler runs from the event loop
// once the menu has closed, and every injection waits for the application to go quiet, the handler has finished by the
// time this returns -- or, for a handler that puts up a modal dialog, the dialog is up and idle.
func chooseMenuBarItem(t *testing.T, screen *unison.HeadlessScreen, wnd *unison.Window, menuTitle, itemTitle string) {
	t.Helper()
	var titlePanel *unison.Panel
	var menu unison.Menu
	screen.Do(func() {
		bar := unison.DefaultMenuFactory().BarForWindowNoCreate(wnd)
		if bar == nil {
			return
		}
		items := menuItemPanels(menuBarPanel(wnd))
		for i := range bar.Count() {
			item := bar.ItemAtIndex(i)
			if item.Title() == menuTitle && item.SubMenu() != nil && i < len(items) {
				menu = item.SubMenu()
				titlePanel = items[i]
				return
			}
		}
	})
	if titlePanel == nil {
		t.Fatalf("no %q menu in the menu bar", menuTitle)
	}
	screen.Click(screen.PanelCenter(titlePanel))

	var itemPanel *unison.Panel
	screen.Do(func() {
		items := menuItemPanels(openMenuPopup(wnd))
		for i := range menu.Count() {
			if item := menu.ItemAtIndex(i); !item.IsSeparator() && item.Title() == itemTitle && i < len(items) {
				itemPanel = items[i]
				return
			}
		}
	})
	if itemPanel == nil {
		t.Fatalf("no %q item in the open %q menu", itemTitle, menuTitle)
	}
	screen.Click(screen.PanelCenter(itemPanel))
}

// choosePopupItem chooses the item at index from a unison.PopupMenu (or anything wrapping one, such as Popup) the way a
// user would: it clicks the popup, which opens an in-window menu holding one item per entry of the popup in the popup's
// own order, then clicks that menu's item. The popup's selection callback has run by the time this returns, since every
// injection waits for the application to go quiet.
func choosePopupItem(t *testing.T, screen *unison.HeadlessScreen, wnd *unison.Window, popup unison.Paneler, index int) {
	t.Helper()
	screen.Click(screen.PanelCenter(popup))
	var itemPanel *unison.Panel
	screen.Do(func() {
		if items := menuItemPanels(openMenuPopup(wnd)); index >= 0 && index < len(items) {
			itemPanel = items[index]
		}
	})
	if itemPanel == nil {
		t.Fatalf("no item %d in the menu the popup opened", index)
	}
	screen.Click(screen.PanelCenter(itemPanel))
}

// modalDialog returns the dialog window currently up alongside wnd and the unison.Dialog behind it, failing the test if
// there is no such window. A dialog runs a nested modal loop, so the call that put it up has not returned yet; every
// injection waits for the application to go quiet inside that loop, which is what makes it safe to look for the dialog
// right after the click that opened it.
func modalDialog(t *testing.T, screen *unison.HeadlessScreen, wnd *unison.Window) (*unison.Window, *unison.Dialog) {
	t.Helper()
	var dialogWnd *unison.Window
	var dialog *unison.Dialog
	screen.Do(func() {
		for _, w := range unison.Windows() {
			if w == wnd {
				continue
			}
			if d, ok := w.ClientData()[unison.DialogClientDataKey].(*unison.Dialog); ok {
				dialogWnd = w
				dialog = d
				return
			}
		}
	})
	if dialogWnd == nil {
		t.Fatal("no dialog is open")
	}
	return dialogWnd, dialog
}

// visibleRect returns the part of p's content area that is within view, in the root coordinate space of its window:
// its whole content area when no scroll panel encloses it, and otherwise the part of it inside the scroll panel's view
// port. An empty rect means nothing of p can be seen, and so nothing of it can be clicked.
func visibleRect(p *unison.Panel) geom.Rect {
	r := p.RectToRoot(p.ContentRect(false))
	scroller := p.ScrollRoot()
	if scroller == nil {
		return r
	}
	view := scroller.ContentView()
	return r.Intersect(view.RectToRoot(view.ContentRect(false)))
}

// fullyVisible reports whether all of p's content area is within view; see visibleRect.
func fullyVisible(p *unison.Panel) bool {
	return visibleRect(p) == p.RectToRoot(p.ContentRect(false))
}

// screenPoint converts a point in the root coordinate space of wnd into the screen's logical coordinate space, which
// is what the injection methods take.
func screenPoint(wnd *unison.Window, pt geom.Point) geom.Point {
	return pt.Add(wnd.ContentRect().Point)
}
