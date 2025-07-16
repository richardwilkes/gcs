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
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xfilepath"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/thememode"
	"github.com/richardwilkes/unison/printing"
)

var _ Rebuildable = &pageExporter{}

const pageKey = "pageKey"

type pageExporter struct {
	unison.Panel
	entity      *gurps.Entity
	provider    gurps.PageInfoProvider
	targetMgr   *TargetMgr
	pages       []*Page
	currentPage int
}

// ExportDockable is an interface for dockables that can be exported to a file.
type ExportDockable interface {
	FileBackedDockable
	PageInfoProvider() gurps.PageInfoProvider
}

// InstallExportCmdHandlers installs the export command handlers on the given dockable.
func InstallExportCmdHandlers(dockable ExportDockable) {
	p := dockable.AsPanel()
	p.InstallCmdHandlers(ExportAsPDFItemID, unison.AlwaysEnabled, func(_ any) { ExportPage("pdf", dockable) })
	p.InstallCmdHandlers(ExportAsWEBPItemID, unison.AlwaysEnabled, func(_ any) { ExportPage("webp", dockable) })
	p.InstallCmdHandlers(ExportAsPNGItemID, unison.AlwaysEnabled, func(_ any) { ExportPage("png", dockable) })
	p.InstallCmdHandlers(ExportAsJPEGItemID, unison.AlwaysEnabled, func(_ any) { ExportPage("jpeg", dockable) })
	p.InstallCmdHandlers(PrintItemID, unison.AlwaysEnabled, func(_ any) { Print(dockable) })
}

// Print the given dockable.
func Print(dockable ExportDockable) {
	p := dockable.PageInfoProvider()
	data, err := newPageExporter(p).exportAsPDFBytes()
	if err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to create print data!"), err)
		return
	}
	dialog := printMgr.NewJobDialog(lastPrinter, "application/pdf", nil)
	if dialog.RunModal() {
		go doPrint(p.PageTitle(), dialog.Printer(), dialog.JobAttributes(), data)
	}
	if p := dialog.Printer(); p != nil {
		lastPrinter = p.PrinterID
	}
}

func doPrint(title string, printer *printing.Printer, jobAttributes *printing.JobAttributes, data []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err := printer.Print(ctx, title, "application/pdf", bytes.NewBuffer(data), len(data), jobAttributes); err != nil {
		unison.InvokeTask(func() { Workspace.ErrorHandler(fmt.Sprintf(i18n.Text("Printing '%s' failed"), title), err) })
	}
}

// ExportPage exports the given dockable to the specified file type, one of "pdf", "webp", "png", or "jpeg".
func ExportPage(ext string, dockable ExportDockable) {
	if tmplDockable, ok := dockable.(*Template); ok {
		tmplDockable.template.ExplicitPageTitle = tmplDockable.Title()
		tmplDockable.template.ExplicitModifiedOn = ""
	}
	dockable.AsPanel().Window().ShowCursor()
	dialog := unison.NewSaveDialog()
	backingFilePath := dockable.BackingFilePath()
	dialog.SetInitialDirectory(filepath.Dir(backingFilePath))
	dialog.SetAllowedExtensions(ext)
	dialog.SetInitialFileName(xfilepath.SanitizeName(xfilepath.BaseName(backingFilePath)))
	if dialog.RunModal() {
		if filePath, ok := unison.ValidateSaveFilePath(dialog.Path(), ext, false); ok {
			gurps.GlobalSettings().SetLastDir(gurps.DefaultLastDirKey, filepath.Dir(filePath))
			exporter := newPageExporter(dockable.PageInfoProvider())
			var err error
			switch ext {
			case "pdf":
				err = exporter.exportAsPDFFile(filePath)
			case "webp":
				err = exporter.exportAsWEBPs(filePath)
			case "png":
				err = exporter.exportAsPNGs(filePath)
			case "jpeg":
				err = exporter.exportAsJPEGs(filePath)
			default:
				err = errs.New("unsupported export format: " + ext)
			}
			if err != nil {
				Workspace.ErrorHandler(fmt.Sprintf(i18n.Text("Unable to export as %s!"), ext), err)
			}
		}
	}
}

func newPageExporter(provider gurps.PageInfoProvider) *pageExporter {
	p := &pageExporter{provider: provider}
	p.targetMgr = NewTargetMgr(p)
	pageSize := p.PageSize()
	r := geom.Rect{Size: pageSize}
	var page *Page
	switch kind := provider.(type) {
	case *gurps.Entity:
		p.entity = kind
		page, _, _ = createPageTopBlock(p.entity, p.targetMgr)
	case *gurps.Loot:
		page = createLootTopBlock(kind, p.targetMgr)
	default:
		page = NewPage(p.provider)
	}
	p.AddChild(page)
	p.pages = append(p.pages, page)
	for _, col := range gurps.SheetSettingsFor(p.entity).BlockLayout.ByRow() {
		startAt := make(map[string]int)
		for {
			rowPanel := unison.NewPanel()
			rowPanel.SetLayoutData(&unison.FlexLayoutData{
				HAlign: align.Fill,
				HGrab:  true,
			})
			for _, c := range col {
				switch c {
				case gurps.BlockLayoutReactionsKey:
					if p.entity != nil {
						addRowPanel(rowPanel, NewReactionsPageList(p.entity), gurps.BlockLayoutReactionsKey, startAt)
					}
				case gurps.BlockLayoutConditionalModifiersKey:
					if p.entity != nil {
						addRowPanel(rowPanel, NewConditionalModifiersPageList(p.entity),
							gurps.BlockLayoutConditionalModifiersKey, startAt)
					}
				case gurps.BlockLayoutMeleeKey:
					if p.entity != nil {
						addRowPanel(rowPanel, NewMeleeWeaponsPageList(p.entity), gurps.BlockLayoutMeleeKey, startAt)
					}
				case gurps.BlockLayoutRangedKey:
					if p.entity != nil {
						addRowPanel(rowPanel, NewRangedWeaponsPageList(p.entity), gurps.BlockLayoutRangedKey, startAt)
					}
				case gurps.BlockLayoutTraitsKey:
					addRowPanel(rowPanel, NewTraitsPageList(p, provider), gurps.BlockLayoutTraitsKey, startAt)
				case gurps.BlockLayoutSkillsKey:
					addRowPanel(rowPanel, NewSkillsPageList(p, provider), gurps.BlockLayoutSkillsKey, startAt)
				case gurps.BlockLayoutSpellsKey:
					addRowPanel(rowPanel, NewSpellsPageList(p, provider), gurps.BlockLayoutSpellsKey, startAt)
				case gurps.BlockLayoutEquipmentKey:
					addRowPanel(rowPanel, NewCarriedEquipmentPageList(p, provider), gurps.BlockLayoutEquipmentKey,
						startAt)
				case gurps.BlockLayoutOtherEquipmentKey:
					addRowPanel(rowPanel, NewOtherEquipmentPageList(p, provider), gurps.BlockLayoutOtherEquipmentKey,
						startAt)
				case gurps.BlockLayoutNotesKey:
					addRowPanel(rowPanel, NewNotesPageList(p, provider), gurps.BlockLayoutNotesKey, startAt)
				}
			}
			children := rowPanel.Children()
			if len(children) == 0 {
				break
			}
			rowPanel.SetLayout(&unison.FlexLayout{
				Columns:      len(children),
				HSpacing:     1,
				HAlign:       align.Fill,
				EqualColumns: true,
			})
			page.AddChild(rowPanel)
			page.SetFrameRect(r)
			page.MarkForLayoutRecursively()
			page.ValidateLayout()
			_, pref, _ := page.Sizes(geom.Size{Width: r.Width})
			excess := pref.Height - pageSize.Height
			if excess <= 0 {
				break // Not extending off the page, so move to the next row
			}
			remaining := (pageSize.Height - page.insets().Bottom) - rowPanel.FrameRect().Y
			startNewPage := false
			data := make([]*pageState, len(children))
			for i, child := range children {
				data[i] = newPageState(child)
				if remaining < data[i].minimum {
					startNewPage = true
				}
			}
			if startNewPage {
				// At least one of the columns can't fit at least the header plus the next row, so start a new page
				page.RemoveChild(rowPanel)
				page = NewPage(p.provider)
				p.AddChild(page)
				p.pages = append(p.pages, page)
				page.AddChild(rowPanel)
				page.SetFrameRect(r)
				page.MarkForLayoutRecursively()
				page.ValidateLayout()
				_, pref, _ = page.Sizes(geom.Size{Width: r.Width})
				if excess = pref.Height - pageSize.Height; excess <= 0 {
					break // Not extending off the page, so move to the next row
				}
				remaining = (pageSize.Height - page.insets().Bottom) - rowPanel.FrameRect().Y
			}
			startNewPage = false
			for _, one := range data {
				allowed := remaining - one.overhead
				start, endBefore := one.helper.CurrentDrawRowRange()
				startAt[one.key()] = len(one.heights) // Assume all remaining fit
				for i := start; i < endBefore; i++ {
					allowed -= one.heights[i] + 1
					if allowed < 0 {
						// No more fit, mark it
						one.helper.SetDrawRowRange(start, max(i, start+1))
						if i == start {
							// I have to guard against the case where a single row is so large it can't fit on a single
							// page on its own. In this case, I just let it flow off the end and drop that extra
							// content.
							//
							// TODO: In the future, see if I can do sub-row partitioning.
							i = start + 1
						}
						startAt[one.key()] = i
						startNewPage = true
						break
					}
				}
			}
			if startNewPage {
				// We've filled the page, so add another
				page = NewPage(p.provider)
				p.AddChild(page)
				p.pages = append(p.pages, page)
			}
		}
	}
	for _, page = range p.pages {
		page.Force = true
		page.SetFrameRect(r)
		page.MarkForLayoutRecursively()
		page.ValidateLayout()
	}
	return p
}

type pageHelper interface {
	OverheadHeight() float32
	RowHeights() []float32
	CurrentDrawRowRange() (start, endBefore int)
	SetDrawRowRange(start, endBefore int)
}

type pageState struct {
	child    *unison.Panel
	helper   pageHelper
	current  float32
	overhead float32
	minimum  float32
	heights  []float32
}

func newPageState(child unison.Paneler) *pageState {
	panel := child.AsPanel()
	helper, ok := panel.Self.(pageHelper)
	if !ok {
		panic("child does not implement pageHelper")
	}
	state := &pageState{
		child:    panel,
		helper:   helper,
		current:  panel.FrameRect().Height,
		overhead: helper.OverheadHeight(),
		heights:  helper.RowHeights(),
	}
	state.minimum = state.overhead
	start, _ := state.helper.CurrentDrawRowRange()
	if len(state.heights) > start {
		state.minimum += state.heights[start] + 1
	}
	return state
}

func (s *pageState) key() string {
	if str, ok := s.child.ClientData()[pageKey].(string); ok {
		return str
	}
	return ""
}

func addRowPanel[T gurps.NodeTypes](rowPanel *unison.Panel, list *PageList[T], key string, startAtMap map[string]int) {
	list.ClientData()[pageKey] = key
	count := list.RowCount()
	startAt := startAtMap[key]
	if count > startAt {
		list.SetDrawRowRange(startAt, count)
		rowPanel.AddChild(list)
	}
}

func (p *pageExporter) exportAsPDFBytes() ([]byte, error) {
	stream := unison.NewMemoryStream()
	defer stream.Close()
	if err := p.exportAsPDF(stream); err != nil {
		return nil, err
	}
	return stream.Bytes(), nil
}

func (p *pageExporter) exportAsPDFFile(filePath string) error {
	if err := os.Remove(filePath); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return errs.Wrap(err)
	}
	stream, err := unison.NewFileStream(filePath)
	if err != nil {
		return err
	}
	defer stream.Close()
	return p.exportAsPDF(stream)
}

func (p *pageExporter) exportAsPDF(stream unison.Stream) error {
	savedColorMode := p.saveTheme()
	defer p.restoreTheme(savedColorMode)
	title := p.provider.PageTitle()
	return unison.CreatePDF(stream, &unison.PDFMetaData{
		Title:           title,
		Author:          xos.CurrentUserName(),
		Subject:         title,
		Keywords:        p.provider.PageKeywords(),
		Creator:         xos.AppName,
		RasterDPI:       300,
		EncodingQuality: 101,
	}, p)
}

func (p *pageExporter) exportAsPNGs(filePathBase string) error {
	return p.exportAsImages(filePathBase, ".png", func(img *unison.Image) ([]byte, error) {
		return img.ToPNG(6)
	})
}

func (p *pageExporter) exportAsWEBPs(filePathBase string) error {
	return p.exportAsImages(filePathBase, ".webp", func(img *unison.Image) ([]byte, error) {
		return img.ToWebp(80, true)
	})
}

func (p *pageExporter) exportAsJPEGs(filePathBase string) error {
	return p.exportAsImages(filePathBase, ".jpeg", func(img *unison.Image) ([]byte, error) {
		return img.ToJPEG(80)
	})
}

func (p *pageExporter) exportAsImages(filePathBase, extension string, f func(img *unison.Image) ([]byte, error)) error {
	filePathBase = strings.TrimSuffix(filePathBase, extension)
	savedColorMode := p.saveTheme()
	defer p.restoreTheme(savedColorMode)
	resolution := gurps.GlobalSettings().General.ImageResolution
	pageNumber := 1
	for p.HasPage(pageNumber) {
		size := p.PageSize()
		var drawErr error
		img, err := unison.NewImageFromDrawing(int(size.Width), int(size.Height), resolution, func(c *unison.Canvas) {
			drawErr = p.DrawPage(c, pageNumber)
		})
		if err != nil {
			return err
		}
		if drawErr != nil {
			return drawErr
		}
		var data []byte
		if data, err = f(img); err != nil {
			return err
		}
		if err = os.WriteFile(fmt.Sprintf("%s-%d%s", filePathBase, pageNumber, extension), data, 0o640); err != nil {
			return err
		}
		pageNumber++
	}
	return nil
}

func (p *pageExporter) saveTheme() thememode.Enum {
	savedColorMode := unison.CurrentThemeMode()
	unison.SetThemeMode(thememode.Light)
	unison.ThemeChanged()
	unison.RebuildDynamicColors()
	return savedColorMode
}

func (p *pageExporter) restoreTheme(colorMode thememode.Enum) {
	unison.SetThemeMode(colorMode)
	unison.ThemeChanged()
	unison.RebuildDynamicColors()
}

// HasPage implements unison.PageProvider.
func (p *pageExporter) HasPage(pageNumber int) bool {
	p.currentPage = pageNumber
	return pageNumber > 0 && pageNumber <= len(p.pages)
}

// PageSize implements unison.PageProvider.
func (p *pageExporter) PageSize() geom.Size {
	sheetSettings := gurps.SheetSettingsFor(p.entity)
	w, h := sheetSettings.Page.Orientation.Dimensions(gurps.MustParsePageSize(sheetSettings.Page.Size))
	return geom.NewSize(w.Pixels(), h.Pixels())
}

// DrawPage implements unison.PageProvider.
func (p *pageExporter) DrawPage(canvas *unison.Canvas, pageNumber int) error {
	p.currentPage = pageNumber
	if pageNumber > 0 && pageNumber <= len(p.pages) {
		page := p.pages[pageNumber-1]
		page.Draw(canvas, page.ContentRect(true))
		return nil
	}
	return errs.New("invalid page number")
}

func (p *pageExporter) String() string {
	return ""
}

func (p *pageExporter) Rebuild(_ bool) {
	gurps.DiscardGlobalResolveCache()
}
