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
	"log/slog"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/desktop"
	"github.com/richardwilkes/toolbox/i18n"
	xfs "github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

type pageRefMappingsDockable struct {
	SettingsDockable
	content *unison.Panel
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

// OpenPageReference opens the given page reference. Returns true if the the user asked to cancel further processing.
func OpenPageReference(ref, highlight string, promptContext map[string]bool) bool {
	switch {
	case unison.HasURLPrefix(ref):
		if err := desktop.Open(ref); err != nil {
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
	ref = ref[3:]
	if ref != "" {
		if !strings.HasSuffix(strings.ToLower(ref), gurps.MarkdownExt) {
			ref += gurps.MarkdownExt
		}
		for _, lib := range gurps.GlobalSettings().LibrarySet.List() {
			filePath := filepath.Join(lib.Path(), "Markdown", ref)
			if xfs.FileIsReadable(filePath) {
				OpenFile(filePath, 0)
				return
			}
		}
	}
	unison.ErrorDialogWithMessage(i18n.Text("Unable to open markdown"), ref+"\n"+i18n.Text("does not exist in the Markdown directory in any library."))
}

func openPDFPageReference(ref, highlight string, promptContext map[string]bool) bool {
	if promptContext == nil {
		promptContext = make(map[string]bool)
	}
	i := len(ref) - 1
	for i >= 0 {
		ch := ref[i]
		if ch >= '0' && ch <= '9' {
			i--
		} else {
			i++
			break
		}
	}
	if i > 0 {
		page, err := strconv.Atoi(ref[i:])
		if err != nil {
			unison.ErrorDialogWithMessage(i18n.Text("Unable to open ")+ref, i18n.Text("Does it exist?"))
			return false
		}
		key := ref[:i]
		s := gurps.GlobalSettings()
		pageRef := s.PageRefs.Lookup(key)
		if pageRef == nil && !promptContext[key] {
			pdfName := PageRefKeyToName(key)
			if pdfName != "" {
				pdfName = fmt.Sprintf(i18n.Text("\nThis key is normally mapped to a PDF named:\n%s"), pdfName)
			}
			switch unison.YesNoCancelDialog(fmt.Sprintf(i18n.Text(`There is no valid mapping for page reference key "%s".
Would you like to create one by choosing a PDF to map to this key?`), key), pdfName) {
			case unison.ModalResponseDiscard:
				promptContext[key] = true
			case unison.ModalResponseOK:
				pageRef = askUserForPageRefPath(key, 0)
			case unison.ModalResponseCancel:
				return true
			}
		}
		if pageRef != nil {
			if strings.TrimSpace(s.General.ExternalPDFCmdLine) == "" {
				pageNum := page + pageRef.Offset - 1 // The pdf package uses 0 for the first page, not 1
				if d, wasOpen := OpenFile(pageRef.Path, pageNum); d != nil {
					if pdfDockable, ok := d.(*PDFDockable); ok {
						pdfDockable.SetSearchText(highlight)
						pdfDockable.LoadPage(pageNum)
						if !wasOpen {
							pdfDockable.ClearHistory()
						}
					}
				}
			} else {
				openExternalPDF(pageRef.Path, page+pageRef.Offset)
			}
		}
	}
	return false
}

func openExternalPDF(filePath string, pageNum int) {
	cl := gurps.GlobalSettings().General.ExternalPDFCmdLine
	cl = strings.ReplaceAll(cl, "$FILE", filePath)
	cl = strings.ReplaceAll(cl, "$PAGE", strconv.Itoa(pageNum))
	cl = strings.TrimSpace(cl)
	parts, err := cmdline.Parse(cl)
	errTitle := i18n.Text("Unable to use external PDF command line")
	if err != nil {
		Workspace.ErrorHandler(errTitle, err)
		return
	}
	if len(parts) == 0 {
		unison.ErrorDialogWithMessage(errTitle, i18n.Text("invalid path"))
		return
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	if err = cmd.Start(); err != nil {
		Workspace.ErrorHandler(errTitle, err)
		return
	}
	go func() {
		if err = cmd.Wait(); err != nil {
			// Intentionally not putting up a dialog, since -- at least on Windows -- many of the viewers incorrectly
			// return a non-zero exit code.
			slog.Error("unexpected response from external PDF viewer", "error", err)
		}
	}()
}

func askUserForPageRefPath(key string, offset int) *gurps.PageRef {
	dialog := unison.NewOpenDialog()
	dialog.SetAllowsMultipleSelection(false)
	dialog.SetResolvesAliases(true)
	dialog.SetAllowedExtensions("pdf")
	dialog.SetCanChooseDirectories(false)
	dialog.SetCanChooseFiles(true)
	global := gurps.GlobalSettings()
	dialog.SetInitialDirectory(global.LastDir(gurps.DefaultLastDirKey))
	if !dialog.RunModal() {
		return nil
	}
	p := dialog.Path()
	global.SetLastDir(gurps.DefaultLastDirKey, filepath.Dir(p))
	pageRef := &gurps.PageRef{
		ID:     key,
		Path:   p,
		Offset: offset,
	}
	gurps.GlobalSettings().PageRefs.Set(pageRef)
	RefreshPageRefMappingsView()
	return pageRef
}

// RefreshPageRefMappingsView causes the Page References Mappings view to be refreshed if it is open.
func RefreshPageRefMappingsView() {
	for _, one := range AllDockables() {
		if d, ok := one.(*pageRefMappingsDockable); ok {
			d.sync()
			break
		}
	}
}

// ShowPageRefMappings shows the Page Reference Mappings.
func ShowPageRefMappings() {
	if Activate(func(d unison.Dockable) bool {
		_, ok := d.AsPanel().Self.(*pageRefMappingsDockable)
		return ok
	}) {
		return
	}
	d := &pageRefMappingsDockable{}
	d.Self = d
	d.TabTitle = i18n.Text("Page Reference Mappings")
	d.TabIcon = svg.Settings
	d.Extensions = []string{gurps.PageRefSettingsExt}
	d.Loader = d.load
	d.Saver = d.save
	d.Resetter = d.reset
	d.Setup(d.addToStartToolbar, nil, d.initContent)
	if len(d.content.Children()) > 1 {
		d.content.Children()[1].RequestFocus()
	}
}

func (d *pageRefMappingsDockable) addToStartToolbar(toolbar *unison.Panel) {
	helpButton := unison.NewSVGButton(svg.Help)
	helpButton.Tooltip = newWrappedTooltip(i18n.Text("Help"))
	helpButton.ClickCallback = func() { HandleLink(nil, "md:Help/Interface/Page Reference Mappings") }
	toolbar.AddChild(helpButton)
}

func (d *pageRefMappingsDockable) initContent(content *unison.Panel) {
	d.content = content
	d.content.SetLayout(&unison.FlexLayout{
		Columns:  5,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	d.sync()
}

func (d *pageRefMappingsDockable) reset() {
	gurps.GlobalSettings().PageRefs = gurps.PageRefs{}
	d.sync()
}

func (d *pageRefMappingsDockable) sync() {
	d.content.RemoveAllChildren()
	for _, one := range gurps.GlobalSettings().PageRefs.List() {
		d.createTrashField(one)
		d.createIDField(one)
		d.createOffsetField(one)
		d.createEditField(one)
		d.createNameField(one)
	}
	d.MarkForRedraw()
}

func (d *pageRefMappingsDockable) createIDField(ref *gurps.PageRef) {
	p := unison.NewLabel()
	p.HAlign = align.Middle
	p.OnBackgroundInk = unison.DefaultTooltipTheme.Label.OnBackgroundInk
	p.SetTitle(ref.ID)
	p.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, 0, unison.NewUniformInsets(1), false),
		unison.NewEmptyBorder(unison.Insets{
			Top:    1,
			Left:   unison.StdHSpacing,
			Bottom: 1,
			Right:  unison.StdHSpacing,
		})))
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, unison.DefaultTooltipTheme.BackgroundInk.Paint(gc, rect, paintstyle.Fill))
		p.DefaultDraw(gc, rect)
	}
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	d.content.AddChild(p)
}

func (d *pageRefMappingsDockable) createOffsetField(ref *gurps.PageRef) {
	p := NewIntegerField(nil, "", i18n.Text("Page Offset"),
		func() int { return ref.Offset },
		func(v int) {
			ref.Offset = v
			gurps.GlobalSettings().PageRefs.Set(ref)
		}, -9999, 9999, true, false)
	p.Tooltip = newWrappedTooltip(i18n.Text(`If your PDF is opening up to the wrong page when opening page references, enter an offset here to compensate.`))
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	d.content.AddChild(p)
}

func (d *pageRefMappingsDockable) createNameField(ref *gurps.PageRef) {
	p := unison.NewLabel()
	p.SetTitle(filepath.Base(ref.Path))
	p.Tooltip = newWrappedTooltip(ref.Path)
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Start,
		VAlign: align.Middle,
	})
	d.content.AddChild(p)
}

func (d *pageRefMappingsDockable) createEditField(ref *gurps.PageRef) {
	b := unison.NewSVGButton(svg.Edit)
	b.ClickCallback = func() {
		askUserForPageRefPath(ref.ID, ref.Offset)
	}
	b.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Middle,
		VAlign: align.Middle,
	})
	d.content.AddChild(b)
}

func (d *pageRefMappingsDockable) createTrashField(ref *gurps.PageRef) {
	b := unison.NewSVGButton(svg.Trash)
	b.ClickCallback = func() {
		if unison.QuestionDialog(fmt.Sprintf(i18n.Text("Are you sure you want to remove\n%s (%s)?"), ref.ID,
			filepath.Base(ref.Path)), "") == unison.ModalResponseOK {
			gurps.GlobalSettings().PageRefs.Remove(ref.ID)
			d.sync()
		}
	}
	b.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Middle,
		VAlign: align.Middle,
	})
	d.content.AddChild(b)
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
