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
	"os"
	"path/filepath"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/role"
)

// TestSplitMarkdownPageRef guards the regression a reference containing a '#' anchor caused: the anchor must be
// removed before the ".md" extension is appended, otherwise the path becomes "File#Section.md" and never resolves to a
// real file.
func TestSplitMarkdownPageRef(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name       string
		ref        string
		wantPath   string
		wantAnchor string
	}{
		{name: "no extension, no anchor", ref: "Home", wantPath: "Home.md", wantAnchor: ""},
		{name: "extension, no anchor", ref: "Home.md", wantPath: "Home.md", wantAnchor: ""},
		{name: "no extension, with anchor", ref: "Home#New", wantPath: "Home.md", wantAnchor: "New"},
		{name: "extension, with anchor", ref: "Home.md#New", wantPath: "Home.md", wantAnchor: "New"},
		{name: "path with subdir and anchor", ref: "Guide/Intro#Getting Started", wantPath: "Guide/Intro.md", wantAnchor: "Getting Started"},
		{name: "uppercase extension preserved", ref: "Home.MD#New", wantPath: "Home.MD", wantAnchor: "New"},
		{name: "empty ref", ref: "", wantPath: "", wantAnchor: ""},
		{name: "anchor only", ref: "#New", wantPath: "", wantAnchor: "New"},
		{name: "trailing hash yields empty anchor", ref: "Home#", wantPath: "Home.md", wantAnchor: ""},
		{name: "encoded spaces in path", ref: "User%20Guide/Scripting%20Guide#code", wantPath: "User Guide/Scripting Guide.md", wantAnchor: "code"},
		{name: "encoded spaces without anchor", ref: "User%20Guide/Home", wantPath: "User Guide/Home.md", wantAnchor: ""},
	} {
		t.Run(tc.name, func(_ *testing.T) {
			path, anchor := splitMarkdownPageRef(tc.ref)
			c.Equal(tc.wantPath, path, tc.name)
			c.Equal(tc.wantAnchor, anchor, tc.name)
		})
	}
}

// TestAsPageRefMappingsDockable verifies that the Page Reference Mappings view is found from any of the Dockable
// values that refer to it, including the embedded SettingsDockable that SettingsDockable.Setup hands to the placement
// code. A direct type assertion misses that inner layer, so RefreshPageRefMappingsView silently found nothing to
// refresh, leaving the view showing a mapping's previous PDF until it was closed and reopened.
func TestAsPageRefMappingsDockable(t *testing.T) {
	c := check.New(t)
	d := &pageRefMappingsDockable{}
	d.Self = d
	c.Equal(d, asPageRefMappingsDockable(d), "view docked in the workspace")
	c.Equal(d, asPageRefMappingsDockable(&d.SettingsDockable), "view in a window of its own")
	other := &SettingsDockable{}
	other.Self = other
	c.Nil(asPageRefMappingsDockable(other), "some other settings view")
}

// TestPageRefMappingsSyncReflectsChangedPath verifies that syncing the Page Reference Mappings view rebuilds its rows
// from the current settings, so that pointing a key at a different PDF shows up in the row's path field. This is the
// update askUserForPageRefPath asks for after the user picks a new PDF for a mapping.
func TestPageRefMappingsSyncReflectsChangedPath(t *testing.T) {
	c := check.New(t)
	global := gurps.GlobalSettings()
	uxtest.SwapForTest(t, &global.PageRefs, gurps.PageRefs{})
	global.PageRefs.Set(&gurps.PageRef{ID: "B", Path: filepath.Join("pdfs", "Basic Set.pdf"), Offset: 2})

	d := newPageRefMappingsContent()
	c.Equal(filepath.Join("pdfs", "Basic Set.pdf"), pageRefMappingRowPath(t, d, 0))

	global.PageRefs.Set(&gurps.PageRef{ID: "B", Path: filepath.Join("pdfs", "Basic Set 4th Edition.pdf"), Offset: 2})
	d.sync()
	c.Equal(filepath.Join("pdfs", "Basic Set 4th Edition.pdf"), pageRefMappingRowPath(t, d, 0))
}

// TestPageRefMappingsInitialFocus verifies that the Page Reference Mappings view starts out focused on the first row's
// offset field, which is what SettingsDockable.Setup selects. The view once overrode that with a hardcoded child index
// that landed on the row's unfocusable ID label, so the window lost its focus and keyboard input went nowhere until
// something was clicked. An empty view, whose grid holds only the note that stands in for its rows, falls back to the
// toolbar rather than being left with nothing focused.
func TestPageRefMappingsInitialFocus(t *testing.T) {
	c := check.New(t)
	global := gurps.GlobalSettings()
	uxtest.SwapForTest(t, &global.PageRefs, gurps.PageRefs{})
	build := func() (d *pageRefMappingsDockable, toolbar, content *unison.Panel) {
		d = &pageRefMappingsDockable{}
		d.Self = d
		content = unison.NewPanel()
		d.initContent(content)
		toolbar = unison.NewPanel()
		d.addToStartToolbar(toolbar)
		return d, toolbar, content
	}

	d, toolbar, content := build()
	c.Equal(0, len(pageRefMappingRows(d)), "no mappings means no rows")
	c.Equal(1, len(d.rows.Children()), "the note stands in for the rows")
	c.Nil(firstFocusableInSubtree(d.rows, nil), "the note cannot take the focus")
	c.NotNil(firstContentFocusTarget(toolbar, content), "the toolbar's help button takes the focus instead")

	global.PageRefs.Set(&gurps.PageRef{ID: "B", Path: filepath.Join("pdfs", "Basic Set.pdf"), Offset: 2})
	d, toolbar, content = build()
	rows := pageRefMappingRows(d)
	if len(rows) != 1 {
		t.Fatalf("one mapping must make one row of key label, offset field, path field and remove button, but made %d", len(rows))
	}
	c.Nil(firstFocusableInSubtree(rows[0][0], nil), "the ID label the old code targeted cannot take the focus")

	target := firstContentFocusTarget(toolbar, content)
	c.True(rows[0][1] == target, "the first row's offset field is the initial focus target")
	_, ok := target.Self.(*IntegerField)
	c.True(ok, "the initial focus target is an offset field")
}

// TestPageRefMappingsSyncKeepsTheFocus verifies that rebuilding the rows, which takes away the widget holding the focus,
// hands the focus to the widget in the same column of the row for the same key, or of the row now in its place when the
// key is gone, or to the view's first tab stop once no row is left, rather than leaving it outside the view, where the
// window would give it to its first tab stop.
func TestPageRefMappingsSyncKeepsTheFocus(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	global := gurps.GlobalSettings()
	uxtest.SwapForTest(t, &global.PageRefs, gurps.PageRefs{})
	mapBoth := func() {
		global.PageRefs = gurps.PageRefs{}
		global.PageRefs.Set(&gurps.PageRef{ID: "B", Path: filepath.Join("pdfs", "Basic Set.pdf")})
		global.PageRefs.Set(&gurps.PageRef{ID: "M", Path: filepath.Join("pdfs", "Magic.pdf")})
	}
	screen.Do(mapBoth)
	screen.Do(ShowPageRefMappings)
	var d *pageRefMappingsDockable
	screen.Do(func() {
		for _, one := range AllDockables() {
			if m := asPageRefMappingsDockable(one); m != nil {
				d = m
			}
		}
	})
	if d == nil {
		t.Fatal("the Page Reference Mappings view must be open")
	}
	const offsetColumn, pathColumn, trashColumn = 1, 2, 3
	// The focus starts in column of row, and is wanted in wantColumn of wantRow afterwards, or in the view's first tab
	// stop when wantRow is -1.
	for _, tc := range []struct {
		name       string
		row        int
		column     int
		change     func()
		wantRow    int
		wantColumn int
	}{
		{
			name:   "choosing another PDF keeps it on the path",
			row:    1,
			column: pathColumn,
			change: func() {
				global.PageRefs.Set(&gurps.PageRef{ID: "M", Path: filepath.Join("pdfs", "Magic 2.pdf")})
				RefreshPageRefMappingsView()
			},
			wantRow:    1,
			wantColumn: pathColumn,
		},
		{
			name:   "removing a mapping moves it to the row in its place",
			column: trashColumn,
			change: func() {
				global.PageRefs.Remove("B")
				d.sync()
			},
			wantColumn: trashColumn,
		},
		{
			name:   "removing the last row's mapping moves it to the row above",
			row:    1,
			column: trashColumn,
			change: func() {
				global.PageRefs.Remove("M")
				d.sync()
			},
			wantColumn: trashColumn,
		},
		{
			name:    "resetting moves it to the view's first tab stop",
			column:  offsetColumn,
			change:  func() { d.reset() },
			wantRow: -1,
		},
	} {
		t.Run(tc.name, func(_ *testing.T) {
			screen.Do(func() {
				mapBoth()
				d.sync()
				pageRefMappingRows(d)[tc.row][tc.column].RequestFocus()
			})
			screen.Do(tc.change)
			var focus, want *unison.Panel
			screen.Do(func() {
				focus = d.Window().CurrentFocus()
				if tc.wantRow < 0 {
					want = firstFocusableInSubtree(d.AsPanel(), nil)
				} else {
					want = pageRefMappingRows(d)[tc.wantRow][tc.wantColumn]
				}
			})
			c.True(focus == want, "the focus must be on the %v, but is on the %v", want, focus)
		})
	}
	uxtest.CloseEditorWithoutPrompt(t, screen, d)
}

// TestPageRefMappingsFlagsMissingPDF verifies that a mapping whose PDF cannot be read is flagged with a warning icon on
// its path and a tooltip saying why, since PageRefs.Lookup treats such a mapping as none at all, and that one whose PDF
// can be read is not: with its whole path in view, it has no tooltip at all. A file that is there but cannot be read, or
// a folder, is not called missing, since the person told so would go looking for it rather than fixing what is wrong.
func TestPageRefMappingsFlagsMissingPDF(t *testing.T) {
	c := check.New(t)
	global := gurps.GlobalSettings()
	uxtest.SwapForTest(t, &global.PageRefs, gurps.PageRefs{})
	dir := t.TempDir()
	present := filepath.Join(dir, "Basic Set.pdf")
	c.NoError(os.WriteFile(present, []byte("%PDF-1.4\n"), 0o600))
	unreadable := filepath.Join(dir, "Unreadable.pdf")
	c.NoError(os.WriteFile(unreadable, []byte("%PDF-1.4\n"), 0o200))
	// Windows has no permission to withhold reading from a file's owner, so there the file can be read after all.
	fi, err := os.Stat(unreadable)
	c.NoError(err)
	ownerCanRead := err == nil && fi.Mode().Perm()&0o400 != 0
	cases := []struct {
		id      string
		path    string
		problem string
	}{
		{id: "B", path: present},
		{id: "D", path: dir, problem: "This is a folder, not a PDF. Choose a PDF or remove the mapping."},
		{id: "M", path: filepath.Join(dir, "Magic.pdf"), problem: "This file could not be found. Choose another PDF or remove the mapping."},
		{id: "U", path: unreadable, problem: "This file cannot be read. Check its permissions, choose another PDF or remove the mapping."},
		{id: "X", problem: "No PDF has been chosen. Choose one or remove the mapping."},
	}
	for _, tc := range cases {
		global.PageRefs.Set(&gurps.PageRef{ID: tc.id, Path: tc.path})
	}

	d := newPageRefMappingsContent()
	rows := pageRefMappingRows(d)
	if len(rows) != len(cases) {
		t.Fatalf("%d mappings must make as many rows, but made %d", len(cases), len(rows))
	}
	for i, tc := range cases {
		if tc.id == "U" && ownerCanRead {
			continue
		}
		field := pageRefMappingPathField(t, rows[i])
		c.Equal(tc.problem, field.problem, "the problem with the PDF for %s", tc.id)
		if tc.problem == "" {
			c.Nil(field.Drawable, "a PDF that can be read is not flagged")
			c.Nil(field.Tooltip, "nor, with its whole path in view, has it anything to say")
		} else {
			c.NotNil(field.Drawable, "a PDF for %s that cannot be read is flagged", tc.id)
			c.NotNil(field.Tooltip, "and says why")
		}
	}
}

// TestPageRefMappingsPathChoosesThePDF verifies that a row's path is the button that chooses another PDF for its key,
// which took a button of its own before, so it is a tab stop that says what pressing it does.
func TestPageRefMappingsPathChoosesThePDF(t *testing.T) {
	c := check.New(t)
	global := gurps.GlobalSettings()
	uxtest.SwapForTest(t, &global.PageRefs, gurps.PageRefs{})
	global.PageRefs.Set(&gurps.PageRef{ID: "B", Path: filepath.Join("pdfs", "Basic Set.pdf")})

	rows := pageRefMappingRows(newPageRefMappingsContent())
	if len(rows) != 1 {
		t.Fatalf("one mapping must make one row, but made %d", len(rows))
	}
	field := pageRefMappingPathField(t, rows[0])
	c.NotNil(field.clickCallback, "clicking the path chooses another PDF")
	c.True(field.Focusable(), "the path can be reached by keyboard to do the same")
	c.Equal(role.Button, field.Accessibility.Role)
}

// TestPageRefMappingRemovalQuestion verifies that the question asked before removing a mapping names its key and the
// file name of its PDF, or only its key when no PDF has been chosen, rather than the "." filepath.Base makes of an empty
// path.
func TestPageRefMappingRemovalQuestion(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		path string
		want string
	}{
		{path: filepath.Join("pdfs", "Basic Set.pdf"), want: "Are you sure you want to remove\nB (Basic Set.pdf)?"},
		{want: "Are you sure you want to remove\nB?"},
	} {
		c.Equal(tc.want, pageRefRemovalQuestion(&gurps.PageRef{ID: "B", Path: tc.path, Offset: 2}), "path %q", tc.path)
	}
}

// TestPageRefMappingsKeyNamesItsBook verifies that a key known to belong to a book says which in its tooltip, so that
// a key mapped to the wrong PDF is easy to spot, while a key not known to belong to any book has nothing to say.
func TestPageRefMappingsKeyNamesItsBook(t *testing.T) {
	c := check.New(t)
	global := gurps.GlobalSettings()
	uxtest.SwapForTest(t, &global.PageRefs, gurps.PageRefs{})
	global.PageRefs.Set(&gurps.PageRef{ID: "DF8:", Path: filepath.Join("pdfs", "Dungeon Fantasy 8.pdf")})
	global.PageRefs.Set(&gurps.PageRef{ID: "ZZ", Path: filepath.Join("pdfs", "House Rules.pdf")})

	d := newPageRefMappingsContent()
	rows := pageRefMappingRows(d)
	if len(rows) != 2 {
		t.Fatalf("two mappings must make two rows, but made %d", len(rows))
	}
	c.Equal("DF8:", pageRefMappingKeyLabel(t, rows[0]).String())
	c.NotNil(rows[0][0].Tooltip, "DF8: is the key for Dungeon Fantasy 8, which its tooltip names")
	c.Equal("ZZ", pageRefMappingKeyLabel(t, rows[1]).String())
	c.Nil(rows[1][0].Tooltip, "ZZ is no book's key, so there is nothing to say")
}

// TestPageReferenceWithoutKeyOpensNothing verifies that a page reference with no key, such as one of only digits, opens
// nothing and does not ask the caller to stop, so that any further references are still opened.
func TestPageReferenceWithoutKeyOpensNothing(t *testing.T) {
	c := check.New(t)
	for _, ref := range []string{"123", ":12", ""} {
		var canceled bool
		c.NotPanics(func() { canceled = OpenPageReference(ref, "", nil) }, ref)
		c.False(canceled, ref)
	}
}

// TestPageReferencePromptRecordsTheChosenPDF verifies how the PDF chosen for a page reference that has no usable
// mapping is recorded: in place of a mapping for its key whose PDF can no longer be read, keeping that mapping's page
// offset, and otherwise under the key as the known keys and imported mappings spell it, with the colon that a key
// ending in a digit is written with.
func TestPageReferencePromptRecordsTheChosenPDF(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	global := gurps.GlobalSettings()
	uxtest.SwapForTest(t, &global.PageRefs, gurps.PageRefs{})
	uxtest.SwapForTest(t, &global.General.ExternalPDFCmdLine, "")
	dir := t.TempDir()
	chosen := filepath.Join(dir, "Basic Set.pdf")
	c.NoError(os.WriteFile(chosen, []byte(pdfTextSamplePDF), 0o600))
	gone := filepath.Join(dir, "Moved.pdf")
	screen.Do(func() { global.SetLastDir(gurps.DefaultLastDirKey, dir) })
	for _, tc := range []struct {
		name   string
		stored []*gurps.PageRef
		ref    string
		want   []gurps.PageRef
	}{
		{
			name:   "in place of a mapping whose PDF is gone, keeping its offset",
			stored: []*gurps.PageRef{{ID: "B", Path: gone, Offset: 2}},
			ref:    "B12",
			want:   []gurps.PageRef{{ID: "B", Path: chosen, Offset: 2}},
		},
		{
			name: "under a key ending in a digit with its colon",
			ref:  "PY120:12",
			want: []gurps.PageRef{{ID: "PY120:", Path: chosen}},
		},
		{
			name:   "in place of a mapping for a key ending in a digit whose PDF is gone",
			stored: []*gurps.PageRef{{ID: "PY120:", Path: gone, Offset: 3}},
			ref:    "PY120:12",
			want:   []gurps.PageRef{{ID: "PY120:", Path: chosen, Offset: 3}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			screen.Do(func() {
				global.PageRefs = gurps.PageRefs{}
				for _, one := range tc.stored {
					global.PageRefs.Set(one)
				}
			})
			// The prompt and the open-file dialog each run a nested modal loop, so the call is posted rather than run
			// through Do and its results are read once both have been dismissed.
			returned := false
			c.True(screen.Post(func() {
				OpenPageReference(tc.ref, "", nil)
				returned = true
			}))
			screen.Sync()
			_, prompt := modalDialog(t, screen, wnd)
			var yes *unison.Button
			screen.Do(func() { yes = prompt.Button(unison.ModalResponseOK) })
			screen.Click(screen.PanelCenter(yes))
			screen.Sync()
			dialogWnd, dialog := modalDialog(t, screen, wnd)
			fileList, rowCount := openDialogFileList(t, screen, dialogWnd)
			if rowCount != 1 {
				t.Fatalf("expected the dialog to list the one PDF in %s, found %d rows", dir, rowCount)
			}
			screen.Click(fileListRowPoint(screen, dialogWnd, fileList, 0))
			var openButton *unison.Button
			screen.Do(func() { openButton = dialog.Button(unison.ModalResponseOK) })
			screen.Click(screen.PanelCenter(openButton))
			var got []gurps.PageRef
			var pdf uxtest.Closable
			screen.Do(func() {
				for _, one := range global.PageRefs.List() {
					got = append(got, *one)
				}
				pdf, _ = LocateFileBackedDockable(chosen).(uxtest.Closable) //nolint:errcheck // Checked below
			})
			c.True(returned, "the page reference has been opened")
			c.Equal(tc.want, got)
			if pdf == nil {
				t.Fatalf("the chosen PDF must have been opened")
			}
			uxtest.CloseEditorWithoutPrompt(t, screen, pdf)
		})
	}
}

// newPageRefMappingsContent builds the content of a Page Reference Mappings view from the current settings.
func newPageRefMappingsContent() *pageRefMappingsDockable {
	d := &pageRefMappingsDockable{}
	d.Self = d
	d.initContent(unison.NewPanel())
	return d
}

// pageRefMappingRows returns the rows of the view's grid, each the key label, offset field, path field and remove
// button of one mapping, or none when the grid holds the note that stands in for them.
func pageRefMappingRows(d *pageRefMappingsDockable) [][]*unison.Panel {
	children := d.rows.Children()
	if len(children) <= pageRefMappingHeadings {
		return nil
	}
	var rows [][]*unison.Panel
	for i := pageRefMappingHeadings; i+pageRefMappingColumns <= len(children); i += pageRefMappingColumns {
		rows = append(rows, children[i:i+pageRefMappingColumns])
	}
	return rows
}

// pageRefMappingKeyLabel returns the row's key label, failing the test if the row holds something else there.
func pageRefMappingKeyLabel(t *testing.T, row []*unison.Panel) *unison.Label {
	t.Helper()
	label, ok := row[0].Self.(*unison.Label)
	if !ok {
		t.Fatalf("the row's key must be a label, not a %T", row[0].Self)
	}
	return label
}

// pageRefMappingPathField returns the row's path field, failing the test if the row holds something else there.
func pageRefMappingPathField(t *testing.T, row []*unison.Panel) *PathField {
	t.Helper()
	field, ok := row[2].Self.(*PathField)
	if !ok {
		t.Fatalf("the row's path must be a path field, not a %T", row[2].Self)
	}
	return field
}

// pageRefMappingRowPath returns the path shown by the given row of the Page Reference Mappings view, or an empty string
// if that row isn't present.
func pageRefMappingRowPath(t *testing.T, d *pageRefMappingsDockable, row int) string {
	t.Helper()
	rows := pageRefMappingRows(d)
	if row >= len(rows) {
		return ""
	}
	return pageRefMappingPathField(t, rows[row]).path
}
