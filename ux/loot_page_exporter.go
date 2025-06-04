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
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/thememode"
)

var _ Rebuildable = &lootPageExporter{}

type lootPageExporter struct {
	unison.Panel
	loot        *gurps.Loot
	targetMgr   *TargetMgr
	pages       []*Page
	currentPage int
}

func newLootPageExporter(loot *gurps.Loot) *lootPageExporter {
	p := &lootPageExporter{loot: loot}
	p.targetMgr = NewTargetMgr(p)
	pageSize := p.PageSize()
	r := unison.Rect{Size: pageSize}
	page := createLootTopBlock(loot, p.targetMgr)
	p.AddChild(page)
	p.pages = append(p.pages, page)
	for _, col := range [][]string{{gurps.BlockLayoutOtherEquipmentKey}, {gurps.BlockLayoutNotesKey}} {
		startAt := make(map[string]int)
		for {
			rowPanel := unison.NewPanel()
			rowPanel.SetLayoutData(&unison.FlexLayoutData{
				HAlign: align.Fill,
				HGrab:  true,
			})
			for _, c := range col {
				switch c {
				case gurps.BlockLayoutOtherEquipmentKey:
					addRowPanel(rowPanel, NewOtherEquipmentPageList(p, loot), gurps.BlockLayoutOtherEquipmentKey, startAt)
				case gurps.BlockLayoutNotesKey:
					addRowPanel(rowPanel, NewNotesPageList(p, loot), gurps.BlockLayoutNotesKey, startAt)
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
			_, pref, _ := page.Sizes(unison.Size{Width: r.Width})
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
				page = NewPage(loot)
				p.AddChild(page)
				p.pages = append(p.pages, page)
				page.AddChild(rowPanel)
				page.SetFrameRect(r)
				page.MarkForLayoutRecursively()
				page.ValidateLayout()
				_, pref, _ = page.Sizes(unison.Size{Width: r.Width})
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
				page = NewPage(loot)
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

func (p *lootPageExporter) exportAsPDFBytes() ([]byte, error) {
	stream := unison.NewMemoryStream()
	defer stream.Close()
	if err := p.exportAsPDF(stream); err != nil {
		return nil, err
	}
	return stream.Bytes(), nil
}

func (p *lootPageExporter) exportAsPDFFile(filePath string) error {
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

func (p *lootPageExporter) exportAsPDF(stream unison.Stream) error {
	savedColorMode := p.saveTheme()
	defer p.restoreTheme(savedColorMode)
	return unison.CreatePDF(stream, &unison.PDFMetaData{
		Title:           p.loot.Name,
		Author:          toolbox.CurrentUserName(),
		Subject:         p.loot.Name,
		Keywords:        "GCS Loot Sheet",
		Creator:         "GCS",
		RasterDPI:       300,
		EncodingQuality: 101,
	}, p)
}

func (p *lootPageExporter) exportAsPNGs(filePathBase string) error {
	return p.exportAsImages(filePathBase, ".png", func(img *unison.Image) ([]byte, error) {
		return img.ToPNG(6)
	})
}

func (p *lootPageExporter) exportAsWEBPs(filePathBase string) error {
	return p.exportAsImages(filePathBase, ".webp", func(img *unison.Image) ([]byte, error) {
		return img.ToWebp(80, true)
	})
}

func (p *lootPageExporter) exportAsJPEGs(filePathBase string) error {
	return p.exportAsImages(filePathBase, ".jpeg", func(img *unison.Image) ([]byte, error) {
		return img.ToJPEG(80)
	})
}

func (p *lootPageExporter) exportAsImages(filePathBase, extension string, f func(img *unison.Image) ([]byte, error)) error {
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

func (p *lootPageExporter) saveTheme() thememode.Enum {
	savedColorMode := unison.CurrentThemeMode()
	unison.SetThemeMode(thememode.Light)
	unison.ThemeChanged()
	unison.RebuildDynamicColors()
	return savedColorMode
}

func (p *lootPageExporter) restoreTheme(colorMode thememode.Enum) {
	unison.SetThemeMode(colorMode)
	unison.ThemeChanged()
	unison.RebuildDynamicColors()
}

// HasPage implements unison.PageProvider.
func (p *lootPageExporter) HasPage(pageNumber int) bool {
	p.currentPage = pageNumber
	return pageNumber > 0 && pageNumber <= len(p.pages)
}

// PageSize implements unison.PageProvider.
func (p *lootPageExporter) PageSize() unison.Size {
	pageSettings := p.loot.PageSettings()
	w, h := pageSettings.Orientation.Dimensions(gurps.MustParsePageSize(pageSettings.Size))
	return unison.NewSize(w.Pixels(), h.Pixels())
}

// DrawPage implements unison.PageProvider.
func (p *lootPageExporter) DrawPage(canvas *unison.Canvas, pageNumber int) error {
	p.currentPage = pageNumber
	if pageNumber > 0 && pageNumber <= len(p.pages) {
		page := p.pages[pageNumber-1]
		page.Draw(canvas, page.ContentRect(true))
		return nil
	}
	return errs.New("invalid page number")
}

func (p *lootPageExporter) String() string {
	return ""
}

func (p *lootPageExporter) Rebuild(_ bool) {
	gurps.DiscardGlobalResolveCache()
}
