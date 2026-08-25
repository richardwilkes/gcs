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
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/richardwilkes/canvas/pdf"
	"github.com/richardwilkes/canvas/stream"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/layoutnode"
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

// pageInfoProviderFor returns the dockable's page info provider, first performing any last-minute setup it needs. A
// gurps.Template has no title or modification timestamp of its own, so both must be supplied by the dockable holding
// it; without this, a template's page footers come out blank, or carry stale values left behind by an earlier export.
// Every path that builds pages goes through here so that no such path can skip the setup.
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
		unison.InvokeTask(func() { Workspace.ErrorHandler(fmt.Sprintf(i18n.Text("Printing '%s' failed"), title), err) })
	}
}

// ExportPage exports the given dockable to the specified file type, one of "pdf", "webp", "png", or "jpeg".
func ExportPage(ext string, dockable ExportDockable) {
	dockable.AsPanel().Window().ShowCursor()
	dialog := unison.NewSaveDialog()
	backingFilePath := dockable.BackingFilePath()
	dialog.SetInitialDirectory(filepath.Dir(backingFilePath))
	dialog.SetAllowedExtensions(ext)
	dialog.SetInitialFileName(xfilepath.SanitizeName(xfilepath.BaseName(backingFilePath)))
	if dialog.RunModal() {
		if filePath, ok := unison.ValidateSaveFilePath(dialog.Path(), ext, false); ok {
			gurps.GlobalSettings().SetLastDir(gurps.DefaultLastDirKey, filepath.Dir(filePath))
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
				Workspace.ErrorHandler(fmt.Sprintf(i18n.Text("Unable to export as %s!"), ext), err)
			}
		}
	}
}

// maxBandPlacementPasses caps the number of pages a single band of the layout is allowed to spread itself over. Each
// pass places at least one row or one block, so no band can legitimately need anything close to this many; the cap is
// only here so that a band that somehow fails to make progress can't hang the export.
const maxBandPlacementPasses = 1000

// placedBlockMarker is what the startAt map holds for a block that isn't a list once it has been placed. Such a block
// is placed whole or not at all, so there is no row to pick up from, only the fact that it has had its say.
const placedBlockMarker = 1

func newPageExporter(provider gurps.PageInfoProvider) *pageExporter {
	p := &pageExporter{provider: provider}
	p.targetMgr = NewTargetMgr(p)
	pageSize := p.PageSize()
	r := geom.Rect{Size: pageSize}
	var page *Page
	switch kind := provider.(type) {
	case *gurps.Entity:
		p.entity = kind
		page = NewPage(p.entity)
	case *gurps.Loot:
		page = createLootTopBlock(kind, p.targetMgr)
	default:
		page = NewPage(p.provider)
	}
	p.AddChild(page)
	p.pages = append(p.pages, page)
	// Each band of the layout is placed on its own, one page at a time. A band that doesn't fit in what is left of the
	// page it starts on is built again for the next page, this time holding only the parts of it that have yet to be
	// placed, and the band after it picks up wherever the one before it left off. The new page is only brought into
	// being once there is something to put on it, so that a band that ends exactly at the bottom of a page can't leave
	// a blank one behind it.
	needNewPage := false
	for i, node := range gurps.SheetSettingsFor(p.entity).Layout.Root.Children {
		startAt := make(map[string]int)
		finished := false
		for range maxBandPlacementPasses {
			built := buildLayoutNode(node, p.exportLeaf(startAt))
			if xreflect.IsNil(built) {
				finished = true // Nothing is left of the band that hasn't already been placed
				break
			}
			if needNewPage {
				page = NewPage(p.provider)
				p.AddChild(page)
				p.pages = append(p.pages, page)
				needNewPage = false
			}
			band := built.AsPanel()
			page.AddChild(band)
			layoutPage(page, r)
			// A band that has the page to itself is allowed to overflow it, since moving it on wouldn't give it any
			// more room than it already has and it would never be placed at all.
			alone := len(page.Children()) == 1
			placer := newBandPlacer(startAt)
			used, done := placer.placeBand(band, availableHeightFor(page, band, pageSize), alone)
			if used == 0 && !done && !alone {
				// Nothing at all fit in what was left of this page, so the band starts over on a page of its own.
				page.RemoveChild(band)
				page = NewPage(p.provider)
				p.AddChild(page)
				p.pages = append(p.pages, page)
				page.AddChild(band)
				layoutPage(page, r)
				placer = newBandPlacer(startAt)
				_, done = placer.placeBand(band, availableHeightFor(page, band, pageSize), true)
			}
			if !placer.applyBand(band) {
				page.RemoveChild(band)
			}
			layoutPage(page, r)
			if done {
				finished = true
				break
			}
			needNewPage = true
		}
		if !finished {
			errs.Log(errs.New("gave up trying to fit a block layout band onto the exported pages"), "band", i)
		}
	}
	for _, page = range p.pages {
		page.Force = true
		layoutPage(page, r)
	}
	return p
}

// exportLeaf returns the leaf function that builds the blocks of a band for the page currently being filled. The
// startAt map holds, per block key, the first row of a list that has yet to be placed and, for a block that isn't a
// list, whether it has been placed at all. A block that has already had its say hands back nothing, so that a band
// that spreads over several pages doesn't repeat on the later ones what is already on the earlier ones.
func (p *pageExporter) exportLeaf(startAt map[string]int) layoutLeafFunc {
	return func(key string) unison.Paneler {
		switch key {
		case gurps.BlockReactionsKey:
			if p.entity == nil {
				return nil
			}
			return preparePageListForExport(NewReactionsPageList(p.entity), key, startAt)
		case gurps.BlockConditionalModifiersKey:
			if p.entity == nil {
				return nil
			}
			return preparePageListForExport(NewConditionalModifiersPageList(p.entity), key, startAt)
		case gurps.BlockMeleeKey:
			if p.entity == nil {
				return nil
			}
			return preparePageListForExport(NewMeleeWeaponsPageList(p.entity), key, startAt)
		case gurps.BlockRangedKey:
			if p.entity == nil {
				return nil
			}
			return preparePageListForExport(NewRangedWeaponsPageList(p.entity), key, startAt)
		case gurps.BlockTraitsKey:
			return preparePageListForExport(NewTraitsPageList(p, p.provider), key, startAt)
		case gurps.BlockSkillsKey:
			return preparePageListForExport(NewSkillsPageList(p, p.provider), key, startAt)
		case gurps.BlockSpellsKey:
			return preparePageListForExport(NewSpellsPageList(p, p.provider), key, startAt)
		case gurps.BlockEquipmentKey:
			return preparePageListForExport(NewCarriedEquipmentPageList(p, p.provider), key, startAt)
		case gurps.BlockOtherEquipmentKey:
			return preparePageListForExport(NewOtherEquipmentPageList(p, p.provider), key, startAt)
		case gurps.BlockNotesKey:
			return preparePageListForExport(NewNotesPageList(p, p.provider), key, startAt)
		default:
			// The blocks that aren't lists only exist for a character, and are built anew for each page, since the one
			// that was built for a page that had no room for it is thrown away rather than moved.
			if p.entity == nil || startAt[key] != 0 {
				return nil
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

// preparePageListForExport readies a freshly built page list for placement on a page, handing back nothing once every
// row of it has been placed on an earlier page.
func preparePageListForExport[T gurps.Node[T]](list *PageList[T], key string, startAt map[string]int) unison.Paneler {
	count := list.RowCount()
	start := startAt[key]
	if count <= start {
		return nil
	}
	list.ClientData()[pageKey] = key
	list.SetDrawRowRange(start, count)
	return list
}

// availableHeightFor returns the height still available on the page for the given band, i.e. everything from the top of
// the band down to the bottom of the page's printable area.
func availableHeightFor(page *Page, band *unison.Panel, pageSize geom.Size) float32 {
	return (pageSize.Height - page.insets().Bottom) - band.FrameRect().Y
}

// layoutPage puts the page at the given rect and lays out everything on it.
func layoutPage(page *Page, r geom.Rect) {
	page.SetFrameRect(r)
	page.MarkForLayoutRecursively()
	page.ValidateLayout()
}

// pageKeyOf returns the block key the leaf function stamped on the panel, or an empty string if it has none.
func pageKeyOf(panel *unison.Panel) string {
	if key, ok := panel.ClientData()[pageKey].(string); ok {
		return key
	}
	return ""
}

// panelMinHeight returns the minimum height the builder recorded for the slot the panel occupies.
func panelMinHeight(panel *unison.Panel) float32 {
	if data, ok := panel.LayoutData().(*unison.FlexLayoutData); ok {
		return data.MinSize.Height
	}
	return 0
}

type pageHelper interface {
	OverheadHeight() float32
	RowHeights() []float32
	CurrentDrawRowRange() (start, endBefore int)
	SetDrawRowRange(start, endBefore int)
}

// bandPanelKind is the kind of thing one of the panels of a band is, as far as fitting it onto a page is concerned.
type bandPanelKind byte

const (
	// atomicBandPanel is a block that is placed whole or not at all.
	atomicBandPanel bandPanelKind = iota
	// listBandPanel is a block whose rows may be split across a page boundary.
	listBandPanel
	// rowBandPanel is a container whose children stand side by side.
	rowBandPanel
	// columnBandPanel is a container whose children are stacked.
	columnBandPanel
)

// bandKindOf returns what kind of thing the given panel of a band is. The panel itself is asked rather than the layout
// node recorded on it, since a container the builder was left with a single child in is dropped and its node recorded
// on the child that took its place, which would otherwise make that child look like a container.
func bandKindOf(panel *unison.Panel) bandPanelKind {
	if _, ok := panel.Self.(pageHelper); ok {
		return listBandPanel
	}
	if _, ok := panel.Layout().(*weightedRowLayout); ok {
		return rowBandPanel
	}
	if node := sheetLayoutContainerNodeOf(panel); node != nil && node.Type == layoutnode.Column {
		return columnBandPanel
	}
	return atomicBandPanel
}

// bandLeafPlacement records what came of trying to place one of the blocks of a band.
type bandLeafPlacement struct {
	helper        pageHelper // nil unless the block is a list
	start         int        // The first row of a list to be drawn
	firstUnplaced int        // The first row of a list that didn't fit
	placed        bool       // Whether anything at all was placed
}

// bandPlacer works out how much of a band fits in the space a page has left for it and then carries that decision out.
// One placer is used for one attempt at placing one band on one page.
type bandPlacer struct {
	startAt  map[string]int
	results  map[*unison.Panel]*bandLeafPlacement
	vSpacing float32
}

// newBandPlacer creates a placer that picks each block of the band up from the row the given map says it left off at,
// and that records where it leaves off in turn.
func newBandPlacer(startAt map[string]int) *bandPlacer {
	return &bandPlacer{
		startAt: startAt,
		results: make(map[*unison.Panel]*bandLeafPlacement),
		// The space a column of blocks leaves between them, matching the layout the builder gives a Column node.
		vSpacing: 1,
	}
}

// placeBand works out how much of the given panel of a band fits in avail, the height the page has left for it,
// returning the height that takes and whether everything in it was placed. Nothing is altered: the decisions are
// recorded and applyBand carries them out.
//
// A list gives up as many of its rows as fit and leaves the rest for the next page. Anything else is placed whole or
// not at all, in which case it is left to the next page instead. mayOverflow says that the panel has the page to
// itself and so must place something rather than run off the end of the pages: a list then takes its first row and a
// block that isn't a list is placed anyway, with whatever runs past the bottom of the page being lost.
//
// A minimum height is a floor the panel stands at however little of it is placed, so it is checked against avail
// before anything else is: a panel whose floor is deeper than what the page has left can't be placed here without
// running off the bottom of it, however few rows it would take, and so the whole of it is left for the next page,
// exactly as a block that can't be split is. A panel that has the page to itself is placed anyway, since no page has
// more room to offer it than this one.
func (b *bandPlacer) placeBand(panel *unison.Panel, avail float32, mayOverflow bool) (used float32, done bool) {
	minHeight := panelMinHeight(panel)
	if minHeight > avail && !mayOverflow {
		b.deferSubtree([]*unison.Panel{panel})
		return 0, false
	}
	switch bandKindOf(panel) {
	case listBandPanel:
		used, done = b.placeList(panel, avail, mayOverflow)
	case rowBandPanel:
		used, done = b.placeRow(panel, avail, mayOverflow)
	case columnBandPanel:
		used, done = b.placeColumn(panel, avail, mayOverflow)
	default:
		used, done = b.placeAtomic(panel, avail, mayOverflow)
	}
	if used > 0 {
		used = max(used, minHeight)
	}
	return used, done
}

// placeList works out how many of the rows of a list fit.
func (b *bandPlacer) placeList(panel *unison.Panel, avail float32, mayOverflow bool) (used float32, done bool) {
	helper, ok := panel.Self.(pageHelper)
	if !ok {
		return 0, true
	}
	start, endBefore := helper.CurrentDrawRowRange()
	heights := helper.RowHeights()
	endBefore = min(endBefore, len(heights))
	result := &bandLeafPlacement{
		helper:        helper,
		start:         start,
		firstUnplaced: endBefore,
		placed:        true,
	}
	b.results[panel] = result
	used = helper.OverheadHeight()
	remaining := avail - used
	for i := start; i < endBefore; i++ {
		height := heights[i] + 1
		if height > remaining {
			if i != start {
				result.firstUnplaced = i
				return used, false
			}
			if !mayOverflow {
				result.placed = false
				result.firstUnplaced = start
				return 0, false
			}
			// A row too tall to fit on a page of its own is placed anyway and allowed to run off the end of the page,
			// dropping whatever doesn't fit, since there is nowhere else for it to go.
			//
			// TODO: In the future, see if I can do sub-row partitioning.
			result.firstUnplaced = start + 1
			return used + height, start+1 >= endBefore
		}
		remaining -= height
		used += height
	}
	return used, true
}

// placeAtomic works out whether a block that can't be split fits.
func (b *bandPlacer) placeAtomic(panel *unison.Panel, avail float32, mayOverflow bool) (used float32, done bool) {
	result := &bandLeafPlacement{placed: true}
	b.results[panel] = result
	_, pref, _ := panel.Sizes(geom.Size{Width: panel.FrameRect().Width})
	height := max(pref.Height, panelMinHeight(panel))
	if height <= avail {
		return height, true
	}
	if mayOverflow {
		return avail, true
	}
	result.placed = false
	return 0, false
}

// placeColumn works out how much of a stack of blocks fits, from the top down.
func (b *bandPlacer) placeColumn(panel *unison.Panel, avail float32, mayOverflow bool) (used float32, done bool) {
	children := panel.Children()
	remaining := avail
	placedAny := false
	for i, child := range children {
		if i != 0 {
			used += b.vSpacing
			remaining -= b.vSpacing
		}
		// Only the first thing to be placed may overflow the page, since anything below it could still be moved on to
		// a page where there is more room for it.
		childUsed, childDone := b.placeBand(child, remaining, mayOverflow && !placedAny)
		used += childUsed
		remaining -= childUsed
		if childUsed > 0 {
			placedAny = true
		}
		if !childDone {
			// Nothing below a block that isn't finished may be placed, or the page after this one would have the
			// blocks of the column out of order.
			b.deferSubtree(children[i+1:])
			return used, false
		}
	}
	return used, true
}

// placeRow works out how much of a set of blocks standing side by side fits. They all start at the same height, so
// they are each given the same space, and if any of them can't take even a part of what it holds, the whole row is
// left for the next page rather than leaving a hole in the middle of it.
func (b *bandPlacer) placeRow(panel *unison.Panel, avail float32, mayOverflow bool) (used float32, done bool) {
	children := panel.Children()
	done = true
	deferred := false
	for _, child := range children {
		childUsed, childDone := b.placeBand(child, avail, mayOverflow)
		used = max(used, childUsed)
		if !childDone {
			done = false
			if childUsed == 0 {
				deferred = true
			}
		}
	}
	if deferred {
		b.deferSubtree(children)
		return 0, false
	}
	return used, done
}

// deferSubtree marks every block within the given panels as having been given nothing, taking back anything that was
// placed within them, so that applyBand leaves them for the next page.
func (b *bandPlacer) deferSubtree(panels []*unison.Panel) {
	for _, panel := range panels {
		switch bandKindOf(panel) {
		case rowBandPanel, columnBandPanel:
			b.deferSubtree(panel.Children())
		default:
			b.results[panel] = &bandLeafPlacement{}
		}
	}
}

// applyBand carries out what placeBand decided: a block that was given nothing is taken off the page, as is any
// container that leaves empty; a list that only partly fit is told to draw just the rows that did; and the point each
// block is to be picked up from on the next page is recorded. Returns false if nothing of the panel is left.
func (b *bandPlacer) applyBand(panel *unison.Panel) bool {
	switch bandKindOf(panel) {
	case rowBandPanel, columnBandPanel:
		for _, child := range slices.Clone(panel.Children()) {
			if !b.applyBand(child) {
				panel.RemoveChild(child)
			}
		}
		return len(panel.Children()) != 0
	default:
		result := b.results[panel]
		if result == nil || !result.placed {
			return false
		}
		key := pageKeyOf(panel)
		if result.helper == nil {
			if key != "" {
				b.startAt[key] = placedBlockMarker
			}
			return true
		}
		result.helper.SetDrawRowRange(result.start, result.firstUnplaced)
		if key != "" {
			b.startAt[key] = result.firstUnplaced
		}
		return true
	}
}

func (p *pageExporter) exportAsPDFBytes() ([]byte, error) {
	s := stream.NewMemoryWStream()
	if err := p.exportAsPDF(s); err != nil {
		return nil, err
	}
	return s.Bytes(), nil
}

func (p *pageExporter) exportAsPDFFile(filePath string) (err error) {
	if err = os.Remove(filePath); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return errs.Wrap(err)
	}
	s, ok := stream.NewFileWStream(filePath)
	if !ok {
		return errs.Newf("could not create %q", filePath)
	}
	defer func() {
		s.Close()
		if err == nil && s.Failed() {
			err = errs.Newf("unable to write %q", filePath)
		}
	}()
	err = p.exportAsPDF(s)
	return err
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
