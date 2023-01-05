/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ux

import (
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/desktop"
	"github.com/richardwilkes/toolbox/i18n"
	xfs "github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
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

// OpenPageReference opens the given page reference in the given window, which should contain a workspace. May pass nil
// for wnd to let it pick the first such window it discovers. Returns true if the the user asked to cancel further
// processing.
func OpenPageReference(wnd *unison.Window, ref, highlight string, promptContext map[string]bool) bool {
	lowerRef := strings.ToLower(ref)
	switch {
	case strings.HasPrefix(lowerRef, "http://") || strings.HasPrefix(lowerRef, "https://"):
		if err := desktop.Open(ref); err != nil {
			unison.ErrorDialogWithError(i18n.Text("Unable to open link"), err)
		}
		return false
	case strings.HasPrefix(lowerRef, "md:"):
		openMarkdownPageReference(wnd, ref)
		return false
	default:
		return openPDFPageReference(wnd, ref, highlight, promptContext)
	}
}

func openMarkdownPageReference(wnd *unison.Window, ref string) {
	ref = ref[3:]
	if ref != "" {
		if !strings.HasSuffix(strings.ToLower(ref), ".md") {
			ref += ".md"
		}
		for _, lib := range model.GlobalSettings().LibrarySet.List() {
			filePath := filepath.Join(lib.Path(), "Markdown", ref)
			if xfs.FileIsReadable(filePath) {
				OpenFile(wnd, filePath)
				return
			}
		}
	}
	unison.ErrorDialogWithMessage(i18n.Text("Unable to open markdown"), ref+"\n"+i18n.Text("does not exist in the Markdown directory in any library."))
}

func openPDFPageReference(wnd *unison.Window, ref, highlight string, promptContext map[string]bool) bool {
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
			return false
		}
		key := ref[:i]
		s := model.GlobalSettings()
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
				if d, wasOpen := OpenFile(wnd, pageRef.Path); d != nil {
					if pdfDockable, ok := d.(*PDFDockable); ok {
						pdfDockable.SetSearchText(highlight)
						pdfDockable.LoadPage(page + pageRef.Offset - 1) // The pdf package uses 0 for the first page, not 1
						if !wasOpen {
							pdfDockable.ClearHistory()
						}
					}
				}
			} else {
				var parts []string
				parts, err = cmdline.Parse(strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(s.General.ExternalPDFCmdLine, "$FILE", pageRef.Path), "$PAGE", strconv.Itoa(page+pageRef.Offset))))
				errTitle := i18n.Text("Unable to use external PDF command line")
				if err != nil {
					unison.ErrorDialogWithError(errTitle, err)
					return false
				}
				if len(parts) == 0 {
					unison.ErrorDialogWithMessage(errTitle, i18n.Text("invalid path"))
					return false
				}
				cmd := exec.Command(parts[0], parts[1:]...)
				if err = cmd.Start(); err != nil {
					unison.ErrorDialogWithError(errTitle, err)
					return false
				}
				go func() {
					if err = cmd.Wait(); err != nil {
						unison.InvokeTask(func() {
							unison.ErrorDialogWithError(i18n.Text("Unexpected response from command"), err)
						})
					}
				}()
			}
		}
	}
	return false
}

func askUserForPageRefPath(key string, offset int) *model.PageRef {
	dialog := unison.NewOpenDialog()
	dialog.SetAllowsMultipleSelection(false)
	dialog.SetResolvesAliases(true)
	dialog.SetAllowedExtensions("pdf")
	dialog.SetCanChooseDirectories(false)
	dialog.SetCanChooseFiles(true)
	global := model.GlobalSettings()
	dialog.SetInitialDirectory(global.LastDir(model.DefaultLastDirKey))
	if !dialog.RunModal() {
		return nil
	}
	p := dialog.Path()
	global.SetLastDir(model.DefaultLastDirKey, filepath.Dir(p))
	pageRef := &model.PageRef{
		ID:     key,
		Path:   p,
		Offset: offset,
	}
	model.GlobalSettings().PageRefs.Set(pageRef)
	RefreshPageRefMappingsView()
	return pageRef
}

// RefreshPageRefMappingsView causes the Page References Mappings view to be refreshed if it is open.
func RefreshPageRefMappingsView() {
	ws := AnyWorkspace()
	if ws == nil {
		return
	}
	ws.DocumentDock.RootDockLayout().ForEachDockContainer(func(container *unison.DockContainer) bool {
		for _, one := range container.Dockables() {
			if d, ok := one.(*pageRefMappingsDockable); ok {
				d.sync()
				return true
			}
		}
		return false
	})
}

// ShowPageRefMappings shows the Page Reference Mappings.
func ShowPageRefMappings() {
	ws, dc, found := Activate(func(d unison.Dockable) bool {
		_, ok := d.(*pageRefMappingsDockable)
		return ok
	})
	if !found && ws != nil {
		d := &pageRefMappingsDockable{}
		d.Self = d
		d.TabTitle = i18n.Text("Page Reference Mappings")
		d.TabIcon = svg.Settings
		d.Extensions = []string{model.PageRefSettingsExt}
		d.Loader = d.load
		d.Saver = d.save
		d.Resetter = d.reset
		d.Setup(ws, dc, d.addToStartToolbar, nil, d.initContent)
		if len(d.content.Children()) > 1 {
			d.content.Children()[1].RequestFocus()
		}
	}
}

func (d *pageRefMappingsDockable) addToStartToolbar(toolbar *unison.Panel) {
	helpButton := unison.NewSVGButton(svg.Help)
	helpButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Help"))
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
	model.GlobalSettings().PageRefs = model.PageRefs{}
	d.sync()
}

func (d *pageRefMappingsDockable) sync() {
	d.content.RemoveAllChildren()
	for _, one := range model.GlobalSettings().PageRefs.List() {
		d.createTrashField(one)
		d.createIDField(one)
		d.createOffsetField(one)
		d.createEditField(one)
		d.createNameField(one)
	}
	d.MarkForRedraw()
}

func (d *pageRefMappingsDockable) createIDField(ref *model.PageRef) {
	p := unison.NewLabel()
	p.Text = ref.ID
	p.HAlign = unison.MiddleAlignment
	p.OnBackgroundInk = unison.DefaultTooltipTheme.Label.OnBackgroundInk
	p.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.NewUniformInsets(1), false),
		unison.NewEmptyBorder(unison.Insets{
			Top:    1,
			Left:   unison.StdHSpacing,
			Bottom: 1,
			Right:  unison.StdHSpacing,
		})))
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, unison.DefaultTooltipTheme.BackgroundInk.Paint(gc, rect, unison.Fill))
		p.DefaultDraw(gc, rect)
	}
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
	})
	d.content.AddChild(p)
}

func (d *pageRefMappingsDockable) createOffsetField(ref *model.PageRef) {
	p := NewIntegerField(nil, "", i18n.Text("Page Offset"),
		func() int { return ref.Offset },
		func(v int) {
			ref.Offset = v
			model.GlobalSettings().PageRefs.Set(ref)
		}, -9999, 9999, true, false)
	p.Tooltip = unison.NewTooltipWithText(i18n.Text(`If your PDF is opening up to the wrong page when opening
page references, enter an offset here to compensate.`))
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
	})
	d.content.AddChild(p)
}

func (d *pageRefMappingsDockable) createNameField(ref *model.PageRef) {
	p := unison.NewLabel()
	p.Text = filepath.Base(ref.Path)
	p.Tooltip = unison.NewTooltipWithText(ref.Path)
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.StartAlignment,
		VAlign: unison.MiddleAlignment,
	})
	d.content.AddChild(p)
}

func (d *pageRefMappingsDockable) createEditField(ref *model.PageRef) {
	b := unison.NewSVGButton(svg.Edit)
	b.ClickCallback = func() {
		askUserForPageRefPath(ref.ID, ref.Offset)
	}
	b.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.MiddleAlignment,
		VAlign: unison.MiddleAlignment,
	})
	d.content.AddChild(b)
}

func (d *pageRefMappingsDockable) createTrashField(ref *model.PageRef) {
	b := unison.NewSVGButton(svg.Trash)
	b.ClickCallback = func() {
		if unison.QuestionDialog(fmt.Sprintf(i18n.Text("Are you sure you want to remove\n%s (%s)?"), ref.ID,
			filepath.Base(ref.Path)), "") == unison.ModalResponseOK {
			model.GlobalSettings().PageRefs.Remove(ref.ID)
			parent := b.Parent()
			index := parent.IndexOfChild(b)
			for i := index; i > index-4; i-- {
				parent.RemoveChildAtIndex(i)
			}
			parent.MarkForLayoutAndRedraw()
		}
	}
	b.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.MiddleAlignment,
		VAlign: unison.MiddleAlignment,
	})
	d.content.AddChild(b)
}

func (d *pageRefMappingsDockable) load(fileSystem fs.FS, filePath string) error {
	s, err := model.NewPageRefsFromFS(fileSystem, filePath)
	if err != nil {
		return err
	}
	model.GlobalSettings().PageRefs = *s
	d.sync()
	return nil
}

func (d *pageRefMappingsDockable) save(filePath string) error {
	return model.GlobalSettings().PageRefs.Save(filePath)
}
