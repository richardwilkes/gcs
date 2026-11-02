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
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/richardwilkes/canvas/pdf"
	"github.com/richardwilkes/canvas/stream"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xfilepath"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/thememode"
	"github.com/richardwilkes/unison/printing"
)

var _ Rebuildable = &pageExporter{}

type pageExporter struct {
	unison.Panel
	entity    *gurps.Entity
	provider  gurps.PageInfoProvider
	targetMgr *TargetMgr
	// lists holds, by block key, the list built for the first rows of the block, and handedOut how many of its
	// continuations have been handed out for the rows that go on past it (see exportPageList).
	lists       map[string]splittableList
	handedOut   map[string]int
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

// pageInfoProviderFor returns the dockable's page info provider after any setup it needs. A gurps.Template has no title
// or modification timestamp of its own, so its dockable supplies them, replacing any stale values from an earlier
// export. Every path that builds pages must go through here.
func pageInfoProviderFor(dockable ExportDockable) gurps.PageInfoProvider {
	if tmplDockable, ok := dockable.(*Template); ok {
		tmplDockable.template.ExplicitPageTitle = tmplDockable.Title()
		tmplDockable.template.ExplicitModifiedOn = ""
	}
	return dockable.PageInfoProvider()
}

// Print the given dockable.
func Print(dockable ExportDockable) {
	p := pageInfoProviderFor(dockable)
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
		unison.InvokeTask(func() { Workspace.ErrorHandler(i18n.Text("Printing '%s' failed", title), err) })
	}
}

// ExportPage exports the given dockable to the specified file type, one of "pdf", "webp", "png", or "jpeg".
func ExportPage(ext string, dockable ExportDockable) {
	dockable.AsPanel().Window().ShowCursor()
	backingFilePath := dockable.BackingFilePath()
	filePath, ok := chooseFileToSave(filepath.Dir(backingFilePath), xfilepath.BaseName(backingFilePath), ext,
		gurps.DefaultLastDirKey)
	if !ok {
		return
	}
	exporter := newPageExporter(pageInfoProviderFor(dockable))
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
		Workspace.ErrorHandler(i18n.Text("Unable to export as %s!", ext), err)
	}
}

func newPageExporter(provider gurps.PageInfoProvider) *pageExporter {
	p := &pageExporter{
		provider:  provider,
		lists:     make(map[string]splittableList),
		handedOut: make(map[string]int),
	}
	p.targetMgr = NewTargetMgr(p)
	var first *Page
	switch kind := provider.(type) {
	case *gurps.Entity:
		p.entity = kind
		first = NewPage(p.entity)
	case *gurps.Loot:
		first = createLootTopBlock(kind, p.targetMgr)
	default:
		first = NewPage(p.provider)
	}
	p.pages = paginateBands(gurps.SheetSettingsFor(p.entity).Layout.Root, p.exportLeaf, func() *Page {
		page := first
		first = nil
		if page == nil {
			page = NewPage(p.provider)
		}
		p.AddChild(page)
		return page
	}, p.PageSize())
	return p
}

// exportLeaf returns the leaf function that builds the blocks of a band for the page currently being filled. The
// startAt map holds, per block key, the first row of a list that has yet to be placed and, for a block that isn't a
// list, whether it has been placed at all. A block that has already had its say hands back a page spacer, so that a
// band that spreads over several pages doesn't repeat on the later ones what is already on the earlier ones, while the
// slot the block held is kept for whatever continues beside it.
func (p *pageExporter) exportLeaf(startAt map[string]int) layoutLeafFunc {
	return func(key string) unison.Paneler {
		switch key {
		case gurps.BlockReactionsKey:
			if p.entity == nil {
				return nil
			}
			return exportPageList(p, key, startAt, func() *PageList[*gurps.ConditionalModifier] {
				return NewReactionsPageList(p.entity)
			})
		case gurps.BlockConditionalModifiersKey:
			if p.entity == nil {
				return nil
			}
			return exportPageList(p, key, startAt, func() *PageList[*gurps.ConditionalModifier] {
				return NewConditionalModifiersPageList(p.entity)
			})
		case gurps.BlockMeleeKey:
			if p.entity == nil {
				return nil
			}
			return exportPageList(p, key, startAt, func() *PageList[*gurps.Weapon] {
				return NewMeleeWeaponsPageList(p.entity)
			})
		case gurps.BlockRangedKey:
			if p.entity == nil {
				return nil
			}
			return exportPageList(p, key, startAt, func() *PageList[*gurps.Weapon] {
				return NewRangedWeaponsPageList(p.entity)
			})
		case gurps.BlockTraitsKey:
			return exportPageList(p, key, startAt, func() *PageList[*gurps.Trait] {
				return NewTraitsPageList(p, p.provider)
			})
		case gurps.BlockSkillsKey:
			return exportPageList(p, key, startAt, func() *PageList[*gurps.Skill] {
				return NewSkillsPageList(p, p.provider)
			})
		case gurps.BlockSpellsKey:
			return exportPageList(p, key, startAt, func() *PageList[*gurps.Spell] {
				return NewSpellsPageList(p, p.provider)
			})
		case gurps.BlockEquipmentKey:
			return exportPageList(p, key, startAt, func() *PageList[*gurps.Equipment] {
				return NewCarriedEquipmentPageList(p, p.provider)
			})
		case gurps.BlockOtherEquipmentKey:
			return exportPageList(p, key, startAt, func() *PageList[*gurps.Equipment] {
				return NewOtherEquipmentPageList(p, p.provider)
			})
		case gurps.BlockNotesKey:
			return exportPageList(p, key, startAt, func() *PageList[*gurps.Note] {
				return NewNotesPageList(p, p.provider)
			})
		default:
			// The blocks that aren't lists only exist for a character, and are built anew for each page, since the one
			// that was built for a page that had no room for it is thrown away rather than moved.
			if p.entity == nil {
				return nil
			}
			if startAt[key] == placedBlockMarker {
				return newPageSpacer()
			}
			panel := newSheetBlockPanel(key, p.entity, p.targetMgr)
			if xreflect.IsNil(panel) {
				return nil
			}
			panel.AsPanel().ClientData()[pageKey] = key
			return panel
		}
	}
}

// exportPageList hands back the panel that shows the rows of a list from the first one startAt says has yet to be
// placed: a list built afresh for a band's first rows, kept while the band is placed, and the next continuation of that
// list for the rows of each later pass, so that every page shows the one table. A later pass starts a page with the
// band alone on it, so it always places some of the continuation. Nothing comes back for a list with no rows at all,
// and a page spacer once every row of it has been placed on an earlier page.
func exportPageList[T gurps.Node[T]](p *pageExporter, key string, startAt map[string]int, build func() *PageList[T]) unison.Paneler {
	start := startAt[key]
	list, ok := p.lists[key]
	if start == 0 || !ok {
		list = build()
		p.lists[key] = list
	}
	return pageListLeaf(key, list, start, p.handedOut, true)
}

func (p *pageExporter) exportAsPDFBytes() ([]byte, error) {
	s := stream.NewMemoryWStream()
	if err := p.exportAsPDF(s); err != nil {
		return nil, err
	}
	return s.Bytes(), nil
}

// exportAsPDFFile writes the PDF in full before it touches the file (see writeExportFile), so that a failure leaves
// whatever the file held before.
func (p *pageExporter) exportAsPDFFile(filePath string) error {
	data, err := p.exportAsPDFBytes()
	if err != nil {
		return err
	}
	return writeExportFile(filePath, data)
}

// writeExportFile writes the data to a temporary file beside the file and renames it over the file once it is complete
// (see xos.WriteSafeFile), so that a failure partway through, such as the volume filling, leaves whatever the file held
// before rather than a truncated file.
func writeExportFile(filePath string, data []byte) error {
	if err := xos.WriteSafeFile(filePath, func(w io.Writer) error {
		_, err := w.Write(data)
		return err
	}); err != nil {
		return errs.NewWithCause(filePath, err)
	}
	return nil
}

func (p *pageExporter) exportAsPDF(s stream.WStream) error {
	savedColorMode := p.saveTheme()
	defer p.restoreTheme(savedColorMode)
	title := p.provider.PageTitle()
	return unison.CreatePDF(s, &pdf.Metadata{
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
		if err = writeExportFile(fmt.Sprintf("%s-%d%s", filePathBase, pageNumber, extension), data); err != nil {
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
	invalidateDisplayData()
	gurps.DiscardGlobalResolveCache()
}
