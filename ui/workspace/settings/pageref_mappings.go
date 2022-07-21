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

package settings

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/gcs/v5/ui/workspace"
	"github.com/richardwilkes/gcs/v5/ui/workspace/external"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

type pageRefMappingsDockable struct {
	Dockable
	content *unison.Panel
}

// ExtractPageReferences extracts any page references from the string.
func ExtractPageReferences(s string) []string {
	var list []string
	for _, one := range strings.FieldsFunc(s, func(ch rune) bool { return ch == ',' || ch == ';' || ch == ' ' }) {
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
		s := settings.Global()
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
				dialog := unison.NewOpenDialog()
				dialog.SetAllowsMultipleSelection(false)
				dialog.SetResolvesAliases(true)
				dialog.SetAllowedExtensions("pdf")
				if dialog.RunModal() {
					pageRef = &settings.PageRef{
						ID:   key,
						Path: dialog.Paths()[0],
					}
					s.PageRefs.Set(pageRef)
					RefreshPageRefMappingsView()
				}
			case unison.ModalResponseCancel:
				return true
			}
		}
		if pageRef != nil {
			if d, wasOpen := workspace.OpenFile(wnd, pageRef.Path); d != nil {
				if pdfDockable, ok := d.(*external.PDFDockable); ok {
					pdfDockable.SetSearchText(highlight)
					pdfDockable.LoadPage(page + pageRef.Offset - 1) // The pdf package uses 0 for the first page, not 1
					if !wasOpen {
						pdfDockable.ClearHistory()
					}
				}
			}
		}
	}
	return false
}

// RefreshPageRefMappingsView causes the Page References Mappings view to be refreshed if it is open.
func RefreshPageRefMappingsView() {
	ws := workspace.Any()
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
	ws, dc, found := workspace.Activate(func(d unison.Dockable) bool {
		_, ok := d.(*pageRefMappingsDockable)
		return ok
	})
	if !found && ws != nil {
		d := &pageRefMappingsDockable{}
		d.Self = d
		d.TabTitle = i18n.Text("Page Reference Mappings")
		d.Extensions = []string{".refs"}
		d.Loader = d.load
		d.Saver = d.save
		d.Resetter = d.reset
		d.Setup(ws, dc, nil, nil, d.initContent)
		if len(d.content.Children()) > 1 {
			d.content.Children()[1].RequestFocus()
		}
	}
}

func (d *pageRefMappingsDockable) initContent(content *unison.Panel) {
	d.content = content
	d.content.SetLayout(&unison.FlexLayout{
		Columns:  4,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	d.sync()
}

func (d *pageRefMappingsDockable) reset() {
	settings.Global().PageRefs = settings.PageRefs{}
	d.sync()
}

func (d *pageRefMappingsDockable) sync() {
	d.content.RemoveAllChildren()
	for _, one := range settings.Global().PageRefs.List() {
		d.createIDField(one)
		d.createOffsetField(one)
		d.createNameField(one)
		d.createTrashField(one)
	}
	d.MarkForRedraw()
}

func (d *pageRefMappingsDockable) createIDField(ref *settings.PageRef) {
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

func (d *pageRefMappingsDockable) createOffsetField(ref *settings.PageRef) {
	p := widget.NewIntegerField(nil, "", i18n.Text("Page Offset"),
		func() int { return ref.Offset },
		func(v int) {
			ref.Offset = v
			settings.Global().PageRefs.Set(ref)
		}, -9999, 9999, true, false)
	p.Tooltip = unison.NewTooltipWithText(i18n.Text(`If your PDF is opening up to the wrong page when opening
page references, enter an offset here to compensate.`))
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
	})
	d.content.AddChild(p)
}

func (d *pageRefMappingsDockable) createNameField(ref *settings.PageRef) {
	p := unison.NewLabel()
	p.Text = filepath.Base(ref.Path)
	p.Tooltip = unison.NewTooltipWithText(ref.Path)
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
		HGrab:  true,
	})
	d.content.AddChild(p)
}

func (d *pageRefMappingsDockable) createTrashField(ref *settings.PageRef) {
	b := unison.NewSVGButton(res.TrashSVG)
	b.ClickCallback = func() {
		if unison.QuestionDialog(fmt.Sprintf(i18n.Text("Are you sure you want to remove\n%s (%s)?"), ref.ID,
			filepath.Base(ref.Path)), "") == unison.ModalResponseOK {
			settings.Global().PageRefs.Remove(ref.ID)
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
	s, err := settings.NewPageRefsFromFS(fileSystem, filePath)
	if err != nil {
		return err
	}
	settings.Global().PageRefs = *s
	d.sync()
	return nil
}

func (d *pageRefMappingsDockable) save(filePath string) error {
	return settings.Global().PageRefs.Save(filePath)
}
