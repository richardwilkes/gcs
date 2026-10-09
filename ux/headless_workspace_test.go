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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/mod"
)

// showInTestWindow opens a window of the given width holding the panels, one per row, each filling the width. The
// window is disposed of when the test ends.
func showInTestWindow(t *testing.T, screen *unison.HeadlessScreen, width float32, panels ...unison.Paneler) *unison.Window {
	t.Helper()
	var wnd *unison.Window
	screen.Do(func() {
		w, err := unison.NewWindow("Test")
		if err != nil {
			t.Errorf("unable to create the window: %v", err)
			return
		}
		w.Content().SetLayout(&unison.FlexLayout{Columns: 1})
		for _, one := range panels {
			one.AsPanel().SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
			w.Content().AddChild(one)
		}
		w.Pack()
		frame := w.FrameRect()
		frame.Width = width
		w.SetFrameRect(frame)
		w.Content().ValidateLayout()
		w.ToFront()
		wnd = w
	})
	if wnd == nil {
		t.Fatal("the window was not created")
	}
	t.Cleanup(func() { screen.Do(wnd.Dispose) })
	return wnd
}

// dragRowAheadOf drags the handle of the from'th child of rows to the upper half of the to'th, which asks for the row
// to be inserted ahead of that one. Both rows must be in view for the drag to land where it is aimed, so the span from
// a little above the target row to the dragged row's handle is scrolled into view first; the headroom keeps the pointer
// clear of the edge that the drop target scrolls at. The test fails if the span does not fit the view, so the two rows
// must be near enough to each other for it to.
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
		handles := uxtest.PanelsOfType[*DragHandle](children[from])
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
		targetVisible = visible.Y-targetRect.Y <= viewSlop && visible.Height >= 16
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
	for _, b := range uxtest.PanelsOfType[*unison.Button](root) {
		if drawable, ok := b.Drawable.(*unison.DrawableSVG); ok && drawable.SVG == icon {
			return b
		}
	}
	return nil
}

// buttonWithTooltip returns the first button within root whose tooltip reads exactly text, or nil if there is none.
// Where several buttons share an icon, the tooltip is what tells them apart.
func buttonWithTooltip(root *unison.Panel, text string) *unison.Button {
	for _, b := range uxtest.PanelsOfType[*unison.Button](root) {
		if b.Tooltip != nil && tooltipText(b.Tooltip) == text {
			return b
		}
	}
	return nil
}

// modalDialog returns the dialog window currently up alongside wnd and the unison.Dialog behind it, failing the test if
// there is no such window. A dialog runs a nested modal loop, and every injection waits for the application to go quiet
// inside that loop, which is what makes it safe to look for the dialog right after the click that opened it.
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

// saveDialogFields returns the file name field of the pure-Go save dialog in dialogWnd, the name it offers and the name
// of the directory its popup shows, failing the test if the window is not the save dialog or has no file name field.
func saveDialogFields(t *testing.T, screen *unison.HeadlessScreen, dialogWnd *unison.Window) (field *unison.Field, fileName, dirName string) {
	t.Helper()
	var title string
	screen.Do(func() {
		title = dialogWnd.Title()
		if fields := uxtest.PanelsOfType[*unison.Field](dialogWnd.Content()); len(fields) == 1 {
			field = fields[0]
			fileName = field.Text()
		}
		// The directory popup is a PopupMenu of an unexported item type, so it is found by the methods it has.
		if popups := uxtest.PanelsOfType[interface {
			Text() string
			ItemCount() int
		}](dialogWnd.Content()); len(popups) != 0 {
			dirName = popups[0].Text()
		}
	})
	if title != "Save…" {
		t.Fatalf("expected the save dialog, found one titled %q", title)
	}
	if field == nil {
		t.Fatal("the save dialog has no file name field")
	}
	return field, fileName, dirName
}

// loadSavedFile reads the file at path with read, which is handed the file's directory and base name the way the
// model's readers expect, failing the test if the file does not exist or does not parse.
func loadSavedFile[T any](t *testing.T, c check.Checker, path string, read func(fs.FS, string) (T, error)) T {
	t.Helper()
	_, err := os.Stat(path)
	c.NoError(err, "the file must exist at %s", path)
	loaded, err := read(os.DirFS(filepath.Dir(path)), filepath.Base(path))
	if err != nil {
		t.Fatalf("the file at %s must parse: %v", path, err)
	}
	return loaded
}

// saveNewFileEditor saves a file editor that has no file on disk yet through its toolbar's Save button, the way a user
// would, and returns where the file went, what the file holds and the hash of the model as it was saved. A model with
// no file makes the Save button bring up the pure-Go save dialog, which must offer offeredName as the file name and
// open in the user library's ancestries folder, where both the ancestries and the name generators they use live; when
// saveAs is empty the offered name is accepted as it stands, and otherwise saveAs is typed in its place. typeName is
// what the editor edits, such as "Ancestry", which its title must show along with the file's base name; ext is the
// extension of the file it saves. The file is read back with read and must hold exactly what the editor holds.
func saveNewFileEditor[M fileEditorModel[M], T gurps.Hashable](t *testing.T, c check.Checker,
	screen *unison.HeadlessScreen, wnd *unison.Window, d *fileEditorDockable[M],
	typeName, offeredName, saveAs, ext string, read func(fs.FS, string) (T, error),
) (savedPath string, loaded T, hash uint64) {
	t.Helper()
	kind := strings.ToLower(typeName)
	var saveButton *unison.Button
	var saveEnabled bool
	screen.Do(func() {
		saveButton = d.saveButton
		saveEnabled = saveButton.Enabled()
	})
	c.True(saveEnabled, "the edited %s can be saved", kind)
	screen.Click(screen.PanelCenter(saveButton))
	dialogWnd, _ := modalDialog(t, screen, wnd)
	fileNameField, fileName, dirName := saveDialogFields(t, screen, dialogWnd)
	c.Equal(offeredName, fileName, "the save dialog offers %q as the %s's file name", offeredName, kind)
	c.Equal(library.AncestriesDirName, dirName, "the dialog opens in the user library's ancestries folder")
	screen.Click(screen.PanelCenter(fileNameField))
	name := offeredName
	if saveAs != "" {
		name = saveAs
		screen.KeyPress(unison.KeyA, mod.OSMenuCommand())
		screen.Type(saveAs)
	}
	screen.KeyPress(unison.KeyReturn, mod.None)
	savedPath = filepath.Join(gurps.GlobalSettings().Libraries.User().AncestriesPath(false), name+ext)
	var path, title, tooltip string
	var modified bool
	var windows int
	screen.Do(func() {
		windows = len(unison.Windows())
		path = d.path
		modified = d.Modified()
		saveEnabled = d.saveButton.Enabled()
		title = d.Title()
		tooltip = d.Tooltip()
		hash = gurps.Hash64(d.model)
	})
	c.Equal(1, windows, "the save dialog has been dismissed")
	c.Equal(savedPath, path, "the editor records where the file was saved")
	c.False(modified, "the saved %s is unmodified", kind)
	c.False(saveEnabled, "saving disables Save")
	c.Equal(typeName+": "+name, title, "the title follows the file's base name")
	c.Equal(savedPath, tooltip, "the tooltip shows the path")
	loaded = loadSavedFile(t, c, savedPath, read)
	c.Equal(hash, gurps.Hash64(loaded), "the file holds exactly what the editor holds")
	return savedPath, loaded, hash
}

// visibleRect returns the part of p's content area that is within view, in the root coordinate space of its window: its
// whole content area when no scroll panel encloses it, and otherwise the part of it inside the scroll panel's view
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

// viewSlop is the tolerance, in root coordinates, allowed when judging whether something lies within view. Root
// coordinates come from multiplying by the sheet's scale and adding the frame offset, which arm64 fuses into a single
// rounding and amd64 rounds twice, so an edge scrolled exactly to the edge of the view can land an ULP past it on one
// of them.
const viewSlop = 0.01

// nearlyWithin reports whether inner lies within outer, allowing viewSlop past each edge.
func nearlyWithin(inner, outer geom.Rect) bool {
	return inner.X >= outer.X-viewSlop && inner.Y >= outer.Y-viewSlop &&
		inner.Right() <= outer.Right()+viewSlop && inner.Bottom() <= outer.Bottom()+viewSlop
}

// fullyVisible reports whether all of p's content area is within view; see visibleRect.
func fullyVisible(p *unison.Panel) bool {
	return nearlyWithin(p.RectToRoot(p.ContentRect(false)), visibleRect(p))
}

// screenPoint converts a point in the root coordinate space of wnd into the screen's logical coordinate space, which
// is what the injection methods take.
func screenPoint(wnd *unison.Window, pt geom.Point) geom.Point {
	return pt.Add(wnd.ContentRect().Point)
}
