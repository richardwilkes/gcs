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
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/toolbox/v2/xflag"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

// pageRefMappingsMinSize is the smallest the Page Reference Mappings view lets its content be before scrolling it
// instead: room for a row's key, offset and buttons, and enough of its path to tell one PDF from another.
var pageRefMappingsMinSize = geom.NewSize(400, 200)

const (
	// pageRefMappingColumns is the number of widgets in a row of the mappings grid: the key, its page offset, the
	// PDF's path, which is clicked to choose another, and the button that removes the mapping.
	pageRefMappingColumns = 4
	// pageRefMappingHeadings is the number of labels in the heading row of the mappings grid; the PDF's spans the
	// button column, which needs no heading.
	pageRefMappingHeadings = 3
)

type pageRefMappingsDockable struct {
	SettingsDockable
	// rows is the grid of mappings, rebuilt by sync: a heading row and then a row per mapping, or a note on how
	// mappings come to be when there are none.
	rows *unison.Panel
	// keys holds the key of each row of the grid, in order.
	keys []string
}

// ExtractPageReferences extracts any page references from the string.
func ExtractPageReferences(s string) []string {
	var list []string
	for _, one := range strings.FieldsFunc(s, func(ch rune) bool { return ch == ',' || ch == ';' }) {
		if one = strings.TrimSpace(one); one != "" {
			list = append(list, one)
		}
	}
	return list
}

// OpenPageReference opens the given page reference. Returns true if the user asked to cancel further processing.
func OpenPageReference(ref, highlight string, promptContext map[string]bool) bool {
	switch {
	case unison.HasURLPrefix(ref):
		if err := unison.OpenBrowser(ref); err != nil {
			Workspace.ErrorHandler(i18n.Text("Unable to open ")+ref, err)
		}
		return false
	case strings.HasPrefix(ref, "md:"):
		openMarkdownPageReference(ref)
		return false
	default:
		return openPDFPageReference(ref, highlight, promptContext)
	}
}

func openMarkdownPageReference(ref string) {
	path, anchor := splitMarkdownPageRef(ref[3:])
	if path != "" {
		for _, lib := range gurps.GlobalSettings().Libraries.List() {
			filePath := filepath.Join(lib.Path(false), "Markdown", path)
			if xos.FileIsReadable(filePath) {
				openMarkdownFileAtAnchor(filePath, anchor)
				return
			}
		}
		for _, lib := range gurps.GlobalSettings().Libraries.List() {
			filePath := filepath.Join(lib.Path(false), path)
			if xos.FileIsReadable(filePath) {
				openMarkdownFileAtAnchor(filePath, anchor)
				return
			}
		}
	}
	unison.ErrorDialogWithMessage(i18n.Text("Unable to open markdown"), path+"\n"+i18n.Text("does not exist in the Markdown directory in any library."))
}

// splitMarkdownPageRef splits a markdown page reference (with the "md:" prefix already removed) into the file path and
// the optional anchor following a '#'. The path is URL-decoded, so encoded and plain references resolve to the same
// file, and gets the ".md" extension if it lacks one; splitting the anchor off first makes "File#Section" resolve to
// "File.md" rather than "File#Section.md". The anchor is returned as-is; unison decodes it when scrolling to it.
func splitMarkdownPageRef(ref string) (path, anchor string) {
	path, anchor, _ = strings.Cut(ref, "#")
	if unescaped, err := url.PathUnescape(path); err == nil {
		path = unescaped
	}
	if path != "" && !strings.HasSuffix(strings.ToLower(path), uti.Markdown.Extensions[0]) {
		path += uti.Markdown.Extensions[0]
	}
	return path, anchor
}

func openMarkdownFileAtAnchor(filePath, anchor string) {
	dockable, _ := OpenFile(filePath, gurps.PageInfo{})
	if anchor == "" {
		return
	}
	if md, ok := dockable.(*MarkdownDockable); ok {
		// Defer until the freshly-displayed dockable has laid out, so the heading positions are known.
		unison.InvokeTask(func() { md.ScrollToAnchor(anchor) })
	}
}

func openPDFPageReference(ref, highlight string, promptContext map[string]bool) bool {
	if promptContext == nil {
		promptContext = make(map[string]bool)
	}
	var pageInfo gurps.PageInfo
	// A separating colon is mandatory if the key would otherwise end in a number, or if the page portion is a page
	// label in the PDF rather than a number, such as "iv".
	key, label, hadColon := strings.Cut(ref, ":")
	if !hadColon {
		key = strings.TrimRight(ref, "0123456789")
		label = ref[len(key):]
	}
	pageInfo.Label = label
	if key != "" && pageInfo.Label != "" {
		// Basic Set page references use `B` for pages in the Campaigns book, too.
		if key == "B" {
			if pageNum, err := strconv.Atoi(pageInfo.Label); err == nil && pageNum >= 338 {
				key = "BX"
			}
		}
		s := gurps.GlobalSettings()
		pageRef := s.PageRefs.Lookup(key)
		if pageRef == nil && hadColon {
			pageRef = s.PageRefs.Lookup(key + ":")
		}
		if pageRef == nil && !promptContext[key] {
			pdfName := PageRefKeyToName(key)
			if pdfName == "" && hadColon {
				pdfName = PageRefKeyToName(key + ":")
			}
			if pdfName != "" {
				pdfName = i18n.Text("\nThis key is normally mapped to a PDF named:\n%s", pdfName)
			}
			switch unison.YesNoCancelDialog(i18n.Text(`There is no valid mapping for page reference key "%s".
Would you like to create one by choosing a PDF to map to this key?`, key), pdfName) {
			case unison.ModalResponseDiscard:
				promptContext[key] = true
			case unison.ModalResponseOK:
				remap := pageRefToRemap(key, hadColon)
				pageRef = askUserForPageRefPath(remap.ID, remap.Path, remap.Offset)
			case unison.ModalResponseCancel:
				return true
			}
		}
		if pageRef != nil {
			pageInfo.Offset = pageRef.Offset
			if strings.TrimSpace(s.General.ExternalPDFCmdLine) == "" {
				if d, wasOpen := OpenFile(pageRef.Path, pageInfo); d != nil {
					if pdfDockable, ok := d.(*PDFDockable); ok {
						pdfDockable.SetSearchText(highlight)
						pdfDockable.LoadPage(pageInfo)
						if !wasOpen {
							pdfDockable.ClearHistory()
						}
					}
				}
			} else {
				openExternalPDF(pageRef.Path, highlight, pageInfo)
			}
		}
	}
	return false
}

var variablesRegexp = regexp.MustCompile(`\$(\w+)`)

func openExternalPDF(filePath, highlight string, pageInfo gurps.PageInfo) {
	errTitle := i18n.Text("Unable to use external PDF command line")
	input := strings.TrimSpace(gurps.GlobalSettings().General.ExternalPDFCmdLine)
	parts, err := xflag.SplitCommandLineWithoutEscapes(input)
	if err != nil {
		Workspace.ErrorHandler(errTitle, err)
		return
	}
	if len(parts) == 0 {
		unison.ErrorDialogWithMessage(errTitle, i18n.Text("invalid path"))
		return
	}
	for i, part := range parts {
		parts[i] = variablesRegexp.ReplaceAllStringFunc(part, func(s string) string {
			switch s {
			case "$FILE":
				return filePath
			case "$TEXT":
				return highlight
			case "$PAGE":
				return pageInfo.ToExternalForm()
			default:
				return s
			}
		})
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	if err = cmd.Start(); err != nil {
		Workspace.ErrorHandler(errTitle, err)
		return
	}
	go func() {
		if err = cmd.Wait(); err != nil {
			// Intentionally not putting up a dialog, since -- at least on Windows -- many viewers incorrectly return
			// a non-zero exit code.
			slog.Error("unexpected response from external PDF viewer", "error", err)
		}
	}()
}

// pageRefToRemap returns the mapping that a PDF chosen for a page reference key with no usable mapping is to replace:
// the one held for the key, whose PDF can no longer be read, so that its page offset is kept, or else a new one under
// the key as the known keys and imported mappings spell it. hadColon is whether a colon separated the key from the
// page, since only then can the key be one spelled with a colon, which a key ending in a digit always is.
func pageRefToRemap(key string, hadColon bool) *gurps.PageRef {
	ids := []string{key}
	if hadColon {
		if last := key[len(key)-1]; last >= '0' && last <= '9' {
			ids = []string{key + ":", key}
		} else {
			ids = append(ids, key+":")
		}
	}
	list := gurps.GlobalSettings().PageRefs.List()
	for _, id := range ids {
		for _, one := range list {
			if one.ID == id {
				return one
			}
		}
	}
	return &gurps.PageRef{ID: ids[0]}
}

// askUserForPageRefPath asks the user for the PDF to map the key to, starting in the directory of the PDF it is
// currently mapped to, if any, and records the mapping. It returns the mapping, or nil if the user chose nothing.
func askUserForPageRefPath(key, current string, offset int) *gurps.PageRef {
	p, ok := chooseFileToOpenNear(current, gurps.DefaultLastDirKey, "pdf")
	if !ok {
		return nil
	}
	pageRef := &gurps.PageRef{
		ID:     key,
		Path:   p,
		Offset: offset,
	}
	gurps.GlobalSettings().PageRefs.Set(pageRef)
	RefreshPageRefMappingsView()
	return pageRef
}

// asPageRefMappingsDockable returns the Page Reference Mappings view for the given Dockable, or nil if it isn't one.
// The Dockable is resolved through its panel's Self, since the value a caller has in hand may be an inner layer rather
// than the view itself -- SettingsDockable.Setup, for example, hands its own embedded SettingsDockable to the
// placement code -- which a direct type assertion would not see.
func asPageRefMappingsDockable(d unison.Dockable) *pageRefMappingsDockable {
	m, _ := d.AsPanel().Self.(*pageRefMappingsDockable) //nolint:errcheck // The nil case is handled by returning nil
	return m
}

// RefreshPageRefMappingsView refreshes the Page Reference Mappings view if it is open.
func RefreshPageRefMappingsView() {
	for _, one := range AllDockables() {
		if d := asPageRefMappingsDockable(one); d != nil {
			d.sync()
			break
		}
	}
}

// ShowPageRefMappings shows the Page Reference Mappings.
func ShowPageRefMappings() {
	if ActivateDockableOfType[*pageRefMappingsDockable](nil) {
		return
	}
	d := &pageRefMappingsDockable{}
	d.ContentWidthBehavior = behavior.Follow
	d.ContentMinSize = pageRefMappingsMinSize
	d.initSettings(d, &settingsSpec{
		title:             i18n.Text("Page Reference Mappings"),
		ext:               gurps.PageRefSettingsExt,
		loader:            d.load,
		saver:             d.save,
		resetter:          d.reset,
		addToStartToolbar: d.addToStartToolbar,
		initContent:       d.initContent,
	})
}

func (d *pageRefMappingsDockable) addToStartToolbar(toolbar *unison.Panel) {
	addHelpButton(toolbar, "md:User%20Guide/Page%20References")
}

// initContent builds the content as a titled section holding the grid of mappings, which sync fills, so that the rows
// can be rebuilt without disturbing the title.
func (d *pageRefMappingsDockable) initContent(content *unison.Panel) {
	initSettingsSections(content)
	section := newSettingsSection(content, i18n.Text("PDF Mappings"), 1)
	d.rows = unison.NewPanel()
	d.rows.SetLayout(&unison.FlexLayout{
		Columns:  pageRefMappingColumns,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	d.rows.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	section.AddChild(d.rows)
	d.sync()
}

func (d *pageRefMappingsDockable) reset() {
	gurps.GlobalSettings().PageRefs = gurps.PageRefs{}
	d.sync()
}

// sync rebuilds the grid from the current settings. A widget in the grid holding the focus goes with the rows, which
// would leave the window to give the focus to its first tab stop, so the focus is handed to the widget in its place:
// see refocus.
func (d *pageRefMappingsDockable) sync() {
	focusIndex, focusKey := d.focusedChild()
	d.rows.RemoveAllChildren()
	list := gurps.GlobalSettings().PageRefs.List()
	d.keys = make([]string, len(list))
	if len(list) == 0 {
		d.addEmptyNote()
	} else {
		d.addHeadings()
		for i, one := range list {
			d.keys[i] = one.ID
			d.createIDField(one)
			d.createOffsetField(one)
			d.createPathField(one)
			d.createTrashField(one)
		}
	}
	d.rows.MarkForLayoutRecursivelyUpward()
	d.MarkForRedraw()
	if focusIndex >= 0 {
		d.refocus(focusIndex, focusKey)
	}
}

// focusedChild returns the index of the grid's child holding the focus, or -1 when none does, along with the key of the
// row that child is in, if it is in one.
func (d *pageRefMappingsDockable) focusedChild() (index int, key string) {
	if wnd := d.rows.Window(); wnd != nil {
		if focus := wnd.CurrentFocus(); focus != nil {
			for i, child := range d.rows.Children() {
				if unison.AncestorIsOrSelf(focus, child) {
					if i >= pageRefMappingHeadings {
						key = d.keys[(i-pageRefMappingHeadings)/pageRefMappingColumns]
					}
					return i, key
				}
			}
		}
	}
	return -1, ""
}

// refocus gives the focus held by the child at index of the grid before it was rebuilt to the child in its place: for a
// row's widget, the one in the same column of the row for the same key, or of the row now in its place, or the last,
// when the key has none. When no child there can take the focus, the view's first tab stop is given it instead.
func (d *pageRefMappingsDockable) refocus(index int, key string) {
	if index >= pageRefMappingHeadings && len(d.keys) != 0 {
		row := (index - pageRefMappingHeadings) / pageRefMappingColumns
		if i := slices.Index(d.keys, key); i >= 0 {
			row = i
		}
		index = pageRefMappingHeadings + min(row, len(d.keys)-1)*pageRefMappingColumns +
			(index-pageRefMappingHeadings)%pageRefMappingColumns
	}
	if children := d.rows.Children(); index < len(children) && children[index].Focusable() {
		children[index].RequestFocus()
	} else {
		d.RequestFocus()
	}
}

// addHeadings adds the row naming the columns. The key's is centered like the keys beneath it, and the PDF's spans the
// button column, which needs no heading.
func (d *pageRefMappingsDockable) addHeadings() {
	headings := []struct {
		title string
		data  unison.FlexLayoutData
	}{
		{i18n.Text("Key"), unison.FlexLayoutData{HAlign: align.Middle}},
		{i18n.Text("Page Offset"), unison.FlexLayoutData{}},
		{i18n.Text("PDF"), unison.FlexLayoutData{HSpan: pageRefMappingColumns - pageRefMappingHeadings + 1}},
	}
	for _, heading := range headings {
		label := unison.NewLabel()
		label.SetTitle(heading.title)
		data := heading.data
		data.VAlign = align.Middle
		label.SetLayoutData(&data)
		d.rows.AddChild(label)
	}
}

// addEmptyNote fills the grid with a note on how mappings come to be, in place of a heading row over nothing. The lines
// sit in a panel of their own, so that the grid holds one child rather than a heading's worth.
func (d *pageRefMappingsDockable) addEmptyNote() {
	text := i18n.Text("No PDFs have been mapped yet. GCS asks which PDF to use for a key the first time a page reference with that key is opened, and keeps the answer here.")
	note := unison.NewPanel()
	note.SetLayout(&unison.FlexLayout{Columns: 1})
	note.SetLayoutData(&unison.FlexLayoutData{HSpan: pageRefMappingColumns})
	for _, line := range unison.NewTextWrappedLines(text, &unison.TextDecoration{
		Font:            settingsNoteFont(),
		OnBackgroundInk: unison.DefaultLabelTheme.OnBackgroundInk,
	}, pageRefMappingsMinSize.Width-unison.StdHSpacing*4) {
		label := unison.NewLabel()
		label.Text = line
		note.AddChild(label)
	}
	d.rows.AddChild(note)
}

// createIDField adds the mapping's key, with the name of the book it is normally the key for, when that is known, so
// that a key mapped to the wrong PDF is easy to spot.
func (d *pageRefMappingsDockable) createIDField(ref *gurps.PageRef) {
	p := unison.NewLabel()
	p.HAlign = align.Middle
	p.OnBackgroundInk = unison.DefaultTooltipTheme.Label.OnBackgroundInk
	p.SetTitle(ref.ID)
	if name := PageRefKeyToName(ref.ID); name != "" {
		hint := i18n.Text("Normally the key for %s", name)
		p.Tooltip = newWrappedTooltip(hint)
		describeWithHint(p, func() string { return hint })
	}
	p.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{},
		geom.NewUniformInsets(1), false), unison.NewEmptyBorder(geom.Insets{
		Top:    1,
		Left:   unison.StdHSpacing,
		Bottom: 1,
		Right:  unison.StdHSpacing,
	})))
	p.DrawCallback = func(gc *unison.Canvas, rect geom.Rect) {
		gc.DrawRect(rect, unison.DefaultTooltipTheme.BackgroundInk.Paint(gc, rect, paintstyle.Fill))
		p.DefaultDraw(gc, rect)
	}
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	d.rows.AddChild(p)
}

func (d *pageRefMappingsDockable) createOffsetField(ref *gurps.PageRef) {
	p := NewIntegerField(nil, "", i18n.Text("Page Offset"),
		func() int { return ref.Offset },
		func(v int) {
			ref.Offset = v
			gurps.GlobalSettings().PageRefs.Set(ref)
		}, -9999, 9999, true, false)
	p.Tooltip = newWrappedTooltip(i18n.Text(`If your PDF is opening up to the wrong page when opening page references, enter an offset here to compensate.`))
	// Named explicitly, since the key it follows would otherwise name it.
	p.Accessibility.Name = i18n.Text("Page offset for %s", ref.ID)
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	d.rows.AddChild(p)
}

// createPathField adds the PDF's path, shortened in its middle when the row has less room than it needs, which is
// clicked to choose another PDF. A path that no longer leads to a readable file is marked as a problem, since
// PageRefs.Lookup treats such a mapping as none at all and the next page reference using the key asks for a PDF again.
func (d *pageRefMappingsDockable) createPathField(ref *gurps.PageRef) {
	p := NewPathField(ref.Path)
	if problem := pageRefPathProblem(ref.Path); problem != "" {
		p.SetProblem(problem)
	}
	p.SetClickCallback(func() { askUserForPageRefPath(ref.ID, ref.Path, ref.Offset) })
	describeWithHint(p, func() string { return i18n.Text("Choose the PDF for %s", ref.ID) })
	d.rows.AddChild(p)
}

// pageRefPathProblem returns what keeps a mapping's path from leading to a file PageRefs.Lookup will use, or an empty
// string when nothing does.
func pageRefPathProblem(path string) string {
	if xos.FileIsReadable(path) {
		return ""
	}
	if path == "" {
		return i18n.Text("No PDF has been chosen. Choose one or remove the mapping.")
	}
	fi, err := os.Stat(path)
	switch {
	case err == nil && fi.IsDir():
		return i18n.Text("This is a folder, not a PDF. Choose a PDF or remove the mapping.")
	case err == nil, errors.Is(err, fs.ErrPermission):
		return i18n.Text("This file cannot be read. Check its permissions, choose another PDF or remove the mapping.")
	default:
		return i18n.Text("This file could not be found. Choose another PDF or remove the mapping.")
	}
}

func (d *pageRefMappingsDockable) createTrashField(ref *gurps.PageRef) {
	b := unison.NewSVGButton(unison.TrashSVG)
	b.Tooltip = newWrappedTooltip(i18n.Text("Remove the mapping for %s", ref.ID))
	b.ClickCallback = func() {
		if unison.QuestionDialog(pageRefRemovalQuestion(ref), "") == unison.ModalResponseOK {
			gurps.GlobalSettings().PageRefs.Remove(ref.ID)
			d.sync()
		}
	}
	b.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Middle,
		VAlign: align.Middle,
	})
	d.rows.AddChild(b)
}

// pageRefRemovalQuestion returns the question asked before removing the mapping, naming its PDF when it has one.
func pageRefRemovalQuestion(ref *gurps.PageRef) string {
	if ref.Path == "" {
		return i18n.Text("Are you sure you want to remove\n%s?", ref.ID)
	}
	return i18n.Text("Are you sure you want to remove\n%s (%s)?", ref.ID, filepath.Base(ref.Path))
}

func (d *pageRefMappingsDockable) load(fileSystem fs.FS, filePath string) error {
	s, err := gurps.NewPageRefsFromFS(fileSystem, filePath)
	if err != nil {
		return err
	}
	gurps.GlobalSettings().PageRefs = *s
	d.sync()
	return nil
}

func (d *pageRefMappingsDockable) save(filePath string) error {
	return gurps.GlobalSettings().PageRefs.Save(filePath)
}
