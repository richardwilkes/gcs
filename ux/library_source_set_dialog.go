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
	"path"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xfilepath"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/check"
)

// sourceSetRow is one row of a source set's checklist: a library, a folder or a file the set may name, or an entry of
// the set that names nothing the Library Explorer lists.
type sourceSetRow struct {
	entry gurps.LibraryFile // What checking the row adds to the set
	label string
	// depth is 0 for a library, one more for each folder above a folder or file, and 1 for an entry naming nothing,
	// which is shown beneath a heading of its own.
	depth   int
	state   check.Enum
	enabled bool
	missing bool // The row is an entry naming nothing
}

// sourceSetChecklist is what the source set editor shows of a set, and the rules for checking its rows, apart from the
// widgets that show them. Ahead of the rest is a row for each entry of the set that names nothing the Library Explorer
// lists, then come a row for each registered library and beneath it, indented by depth, a row for each folder holding
// files of the list type at any depth and for each such file; a folder or file the set names is shown even when it
// isn't one of those. A row the set names is checked. A row beneath one the set names is checked too, and can't be
// changed, since the set takes it in already. A library or folder the set names something beneath is mixed: it is
// checked only when it is named itself, which takes in the files added to it later. Which rows there are is settled
// when the checklist is made and again when the extension changes, so that checking and unchecking rows only ever
// changes their states. With no listing, there are no rows.
type sourceSetChecklist struct {
	listing *navigatorListing
	libs    *library.Libraries
	ext     string
	entries []gurps.LibraryFile // Normalized (see gurps.NormalizeLibrarySourceEntries)
	// listed holds what the listing lists in a registered library, true for a library or folder and false for a file.
	listed map[gurps.LibraryFile]bool
	rows   []sourceSetRow
}

// newSourceSetChecklist returns the checklist of a set holding the entries, for the files with the extension that the
// registered libraries hold, as the Library Explorer listed them.
func newSourceSetChecklist(listing *navigatorListing, libs *library.Libraries, ext string,
	entries []gurps.LibraryFile,
) *sourceSetChecklist {
	c := &sourceSetChecklist{
		listing: listing,
		libs:    libs,
		ext:     ext,
		entries: gurps.NormalizeLibrarySourceEntries(entries),
		listed:  make(map[gurps.LibraryFile]bool),
	}
	c.forEachLibrary(func(key string, entries []*navigatorEntry) {
		c.listed[gurps.LibraryFile{Library: key}] = true
		var walk func(list []*navigatorEntry)
		walk = func(list []*navigatorEntry) {
			for _, entry := range list {
				c.listed[gurps.LibraryFile{Library: key, Path: entry.path}] = entry.isDir
				walk(entry.entries)
			}
		}
		walk(entries)
	})
	c.buildRows()
	return c
}

// forEachLibrary calls f with the key and listed entries of each library of the listing still registered, in listing
// order, as resolveLibraryFilterFiles goes over them.
func (c *sourceSetChecklist) forEachLibrary(f func(key string, entries []*navigatorEntry)) {
	if c.listing == nil {
		return
	}
	for _, one := range c.listing.libraries {
		if key := one.library.Key(); c.libs.Lookup(key) == one.library {
			f(key, one.entries)
		}
	}
}

func (c *sourceSetChecklist) buildRows() {
	c.rows = nil
	if c.listing == nil {
		return
	}
	for _, entry := range c.entries {
		if _, ok := c.listed[entry]; !ok {
			c.rows = append(c.rows, sourceSetRow{
				entry: entry, label: describeLibraryEntry(c.libs, entry), depth: 1,
				missing: true,
			})
		}
	}
	c.forEachLibrary(func(key string, entries []*navigatorEntry) {
		label := key
		if lib := c.libs.Lookup(key); lib != nil {
			label = lib.Data().Title
		}
		c.rows = append(c.rows, sourceSetRow{entry: gurps.LibraryFile{Library: key}, label: label})
		c.rows = append(c.rows, c.entryRows(key, entries, 1)...)
	})
	c.updateStates()
}

// entryRows returns the rows of the listed entries of the library with the key, and those beneath them, at the depth.
func (c *sourceSetChecklist) entryRows(key string, entries []*navigatorEntry, depth int) []sourceSetRow {
	var rows []sourceSetRow
	for _, entry := range entries {
		f := gurps.LibraryFile{Library: key, Path: entry.path}
		name := path.Base(entry.path)
		if entry.isDir {
			beneath := c.entryRows(key, entry.entries, depth+1)
			if len(beneath) != 0 || slices.ContainsFunc(c.entries, f.Covers) {
				rows = append(rows, sourceSetRow{entry: f, label: name, depth: depth})
				rows = append(rows, beneath...)
			}
		} else if ofType := strings.EqualFold(path.Ext(name), c.ext); ofType || slices.Contains(c.entries, f) {
			if ofType {
				name = xfilepath.TrimExtension(name) // As the Library Explorer shows it
			}
			rows = append(rows, sourceSetRow{entry: f, label: name, depth: depth})
		}
	}
	return rows
}

func (c *sourceSetChecklist) updateStates() {
	for i := range c.rows {
		row := &c.rows[i]
		row.enabled = true
		switch {
		case slices.Contains(c.entries, row.entry):
			row.state = check.On
		case slices.ContainsFunc(c.entries, func(entry gurps.LibraryFile) bool { return entry.Covers(row.entry) }):
			row.state = check.On
			row.enabled = false
		case slices.ContainsFunc(c.entries, row.entry.Covers):
			row.state = check.Mixed
		default:
			row.state = check.Off
		}
	}
}

// toggle unchecks the row at the index when it is checked, taking its entry out of the set, and otherwise checks it,
// putting its entry in the set in place of those beneath it. A row that can't be changed is left alone.
func (c *sourceSetChecklist) toggle(index int) {
	if index < 0 || index >= len(c.rows) || !c.rows[index].enabled {
		return
	}
	entry := c.rows[index].entry
	if c.rows[index].state == check.On {
		c.entries = slices.DeleteFunc(slices.Clone(c.entries), func(one gurps.LibraryFile) bool { return one == entry })
	} else {
		c.entries = gurps.NormalizeLibrarySourceEntries(append(slices.Clone(c.entries), entry))
	}
	c.updateStates()
}

// setExtension has the checklist show the files with the extension in place of those it showed. The set's entries for
// files are dropped, since they are files of the other type: those the listing lists as files, and those naming
// nothing that end with the extension the checklist had.
func (c *sourceSetChecklist) setExtension(ext string) {
	if strings.EqualFold(ext, c.ext) {
		return
	}
	c.entries = slices.DeleteFunc(slices.Clone(c.entries), func(entry gurps.LibraryFile) bool {
		if isDir, ok := c.listed[entry]; ok {
			return !isDir
		}
		return entry.Path != "" && strings.EqualFold(path.Ext(entry.Path), c.ext)
	})
	c.ext = ext
	c.buildRows()
}

// result returns the set's entries as they now stand, normalized.
func (c *sourceSetChecklist) result() []gurps.LibraryFile {
	return slices.Clone(c.entries)
}

// builtInLibrarySourceSetTitles returns what the source set popup calls the list type's sets that aren't saved ones:
// its set of every file and an unsaved set (see libraryFilterSetTitle). No saved set of the type may share them.
func builtInLibrarySourceSetTitles(kind libraryFilterKind) []string {
	return []string{libraryFilterSetTitle(kind, nil, false), libraryFilterSetTitle(kind, &gurps.LibrarySourceSet{}, true)}
}

// checkLibrarySourceSet returns whether a set of the list type may have the name, which is when it isn't blank, isn't
// one of the type's built-in titles (see builtInLibrarySourceSetTitles) and no set of the type other than except, the
// set being edited, if any, has it, all ignoring case and surrounding whitespace; whether the set can be accepted, which
// also needs entries, its count of entries, not to be 0; and the message the editor shows when it can't be. A blank
// name gets no message, since the name field's watermark already says what is wanted.
func checkLibrarySourceSet(settings *gurps.Settings, kind libraryFilterKind, name string, entries int,
	except *gurps.LibrarySourceSet,
) (nameValid, accept bool, message string) {
	name = strings.TrimSpace(name)
	inUse := name != "" && (settings.LibrarySourceSetNameInUse(kind.key, name, except) ||
		slices.ContainsFunc(builtInLibrarySourceSetTitles(kind), func(title string) bool {
			return strings.EqualFold(title, name)
		}))
	nameValid = name != "" && !inUse
	accept = nameValid && entries != 0
	switch {
	case inUse:
		message = i18n.Text("Another set of this type already has this name.")
	case entries == 0:
		message = i18n.Text("Check at least one library, folder or file.")
	}
	return nameValid, accept, message
}

// showLibrarySourceSetDialog edits the set, and which list type it is for, returning true when the user accepted. The
// set and the key change only then. It can't be accepted until the set has a name it may have and at least one entry
// (see checkLibrarySourceSet, which except, the set being edited, if any, is passed on to).
func showLibrarySourceSetDialog(title string, key *string, set *gurps.LibrarySourceSet,
	except *gurps.LibrarySourceSet,
) bool {
	kinds := libraryFilterKinds()
	kindIndex := max(slices.IndexFunc(kinds, func(kind libraryFilterKind) bool { return kind.key == *key }), 0)
	name := set.Name
	listing := libraryFilterListing()
	libs := gurps.GlobalSettings().Libraries
	checklist := newSourceSetChecklist(listing, libs, kinds[kindIndex].ext, set.Entries)

	content := unison.NewPanel()
	content.SetLayout(&unison.FlexLayout{Columns: 2, HSpacing: unison.StdHSpacing, VSpacing: unison.StdVSpacing})
	content.AddChild(NewFieldLeadingLabel(i18n.Text("Name"), false))
	nameField := NewStringField(nil, "", i18n.Text("Name"), func() string { return name },
		func(s string) { name = s })
	nameField.Watermark = i18n.Text(`Give this set a name, like "My Campaign"`)
	nameField.SetMinimumTextWidthUsing(strings.Repeat("M", 20))
	content.AddChild(nameField)

	content.AddChild(NewFieldLeadingLabel(i18n.Text("Type"), false))
	typePopup := unison.NewPopupMenu[string]()
	for _, kind := range kinds {
		typePopup.AddItem(kind.title)
	}
	typePopup.SelectIndex(kindIndex)
	content.AddChild(typePopup)

	list := unison.NewPanel()
	list.SetLayout(&unison.FlexLayout{Columns: 1, VSpacing: unison.StdVSpacing})
	list.SetBorder(unison.NewEmptyBorder(unison.StdInsets()))
	scroll := unison.NewScrollPanel()
	scroll.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.NewUniformInsets(1), false))
	scroll.SetContent(list, behavior.Fill, behavior.Fill)
	// Sized by hint rather than to fit, since a library's files may number in the hundreds.
	scroll.SetLayoutData(&unison.FlexLayoutData{
		SizeHint: geom.NewSize(500, 400),
		MinSize:  geom.NewSize(300, 200),
		HSpan:    2,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
		HGrab:    true,
		VGrab:    true,
	})
	content.AddChild(scroll)

	dialog, err := unison.NewDialog(nil, nil, content,
		[]*unison.DialogButtonInfo{unison.NewCancelButtonInfo(), unison.NewOKButtonInfo()})
	if err != nil {
		reportUIError(i18n.Text("Unable to create the source set editor"), err)
		return false
	}
	dialog.Window().SetTitle(title)
	setMessage := addDialogNotice(dialog)
	nameField.ValidateCallback = func() bool {
		valid, accept, message := checkLibrarySourceSet(gurps.GlobalSettings(), kinds[kindIndex], name,
			len(checklist.entries), except)
		dialog.Button(unison.ModalResponseOK).SetEnabled(accept)
		setMessage(message)
		return valid
	}

	var boxes []*unison.CheckBox
	// syncBoxes brings the boxes into line with the checklist's rows, which keeps the scroll position and the focus,
	// since the rows themselves don't change.
	syncBoxes := func() {
		for i, box := range boxes {
			box.State = checklist.rows[i].state
			box.SetEnabled(checklist.rows[i].enabled)
			box.MarkForRedraw()
		}
	}
	render := func() {
		list.RemoveAllChildren()
		boxes = make([]*unison.CheckBox, 0, len(checklist.rows))
		if listing == nil {
			notice := NewWrappingLabel()
			notice.SetTitle(i18n.Text("The Library Explorer is still scanning the libraries, so there is nothing to choose from yet."))
			list.AddChild(notice)
		}
		headed := false
		for i, row := range checklist.rows {
			if row.missing && !headed {
				heading := unison.NewLabel()
				heading.Font = unison.EmphasizedSystemFont
				heading.SetTitle(i18n.Text("Can't be found"))
				list.AddChild(heading)
				headed = true
			}
			box := unison.NewCheckBox()
			if row.depth == 0 {
				box.Font = unison.EmphasizedSystemFont
			}
			box.SetTitle(row.label)
			box.SetBorder(unison.NewEmptyBorder(geom.Insets{Left: float32(row.depth) * unison.StdHSpacing * 4}))
			box.ClickCallback = func() {
				checklist.toggle(i)
				syncBoxes()
				nameField.Validate()
			}
			list.AddChild(box)
			boxes = append(boxes, box)
		}
		syncBoxes()
		list.MarkForLayoutRecursivelyUpward()
		scroll.MarkForRedraw()
	}
	render()
	typePopup.SelectionChangedCallback = func(popup *unison.PopupMenu[string]) {
		i := popup.SelectedIndex()
		if i < 0 || i >= len(kinds) || i == kindIndex {
			return
		}
		kindIndex = i
		checklist.setExtension(kinds[i].ext)
		render()
		scroll.SetPosition(0, 0)
		nameField.Validate()
	}
	nameField.Validate() // Here to update the OK button and the message.
	nameField.RequestFocus()
	nameField.SelectAll()
	if dialog.RunModal() != unison.ModalResponseOK {
		return false
	}
	set.Name = strings.TrimSpace(name)
	set.Entries = checklist.result()
	*key = kinds[kindIndex].key
	return true
}
