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
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/layoutnode"
	"github.com/richardwilkes/gcs/v5/model/paper"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/thememode"
)

// newTestTemplateDockable returns a template dockable for the given file name, first registering the bindable actions
// its toolbar needs. It is marked as loaded from that file, so it goes by the path rather than the template's name.
func newTestTemplateDockable(fileName string, data *gurps.Template) *Template {
	registerKeyBindingsOnce.Do(func() { registerActions() })
	template := NewTemplate(filepath.Join("some", "dir", fileName+gurps.TemplatesExt), data)
	template.markLoadedFromFile()
	return template
}

// TestPageInfoProviderForTemplate verifies that obtaining a template dockable's page info provider fills in the page
// title and clears the modification timestamp, replacing any values left from an earlier export. A gurps.Template has
// neither of its own, so printing, which once skipped this setup, produced footers with a blank or stale title.
func TestPageInfoProviderForTemplate(t *testing.T) {
	t.Run("stale values from an earlier export are replaced", func(t *testing.T) {
		c := check.New(t)
		data := gurps.NewTemplate()
		dockable := newTestTemplateDockable("My Template", data)
		data.ExplicitPageTitle = "Some Other Template"
		data.ExplicitModifiedOn = "Jan 1, 1970"

		provider := pageInfoProviderFor(dockable)
		c.Equal(data, provider, "the template itself is the page info provider")
		c.Equal("My Template", provider.PageTitle(), "the page title is the dockable's title")
		c.Equal("", provider.ModifiedOnString(), "the stale modification timestamp is cleared")
	})

	t.Run("renaming the dockable's file is picked up", func(t *testing.T) {
		c := check.New(t)
		data := gurps.NewTemplate()
		dockable := newTestTemplateDockable("My Template", data)
		c.Equal("My Template", pageInfoProviderFor(dockable).PageTitle())

		dockable.path = filepath.Join("elsewhere", "Renamed"+gurps.TemplatesExt)
		c.Equal("Renamed", pageInfoProviderFor(dockable).PageTitle(), "the current title is used, not the first one")
	})
}

// TestPageInfoProviderForSheet verifies that a character sheet's page info provider is handed back as-is, with its own
// title intact, since only templates need values supplied by their dockable.
func TestPageInfoProviderForSheet(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	entity.Profile.Name = "Dai Blackthorn"

	provider := pageInfoProviderFor(sheet)
	c.Equal(entity, provider, "the entity itself is the page info provider")
	c.Equal("Dai Blackthorn", provider.PageTitle(), "the entity supplies its own page title")
}

// TestPageExporterKeepsEveryBandWithinThePage verifies, over a range of list lengths and for a list stacked below a
// long one both as a band of its own and within a column, that no band of an exported page runs past the bottom of the
// page's printable area. A list with a minimum height once did: it was placed where its rows fit and only then stood at
// its deeper floor.
func TestPageExporterKeepsEveryBandWithinThePage(t *testing.T) {
	c := check.New(t)
	shapes := map[string]func() *gurps.SheetLayoutNode{
		"as bands": func() *gurps.SheetLayoutNode {
			return testContainerNode(layoutnode.Column, fxp.One,
				testBlockNode(gurps.BlockTraitsKey, fxp.One),
				testBlockNode(gurps.BlockSkillsKey, fxp.One),
			)
		},
		"in a column": func() *gurps.SheetLayoutNode {
			return testContainerNode(layoutnode.Column, fxp.One,
				testContainerNode(layoutnode.Row, fxp.One,
					testContainerNode(layoutnode.Column, fxp.One,
						testBlockNode(gurps.BlockTraitsKey, fxp.One),
						testBlockNode(gurps.BlockSkillsKey, fxp.One),
					),
					testBlockNode(gurps.BlockNotesKey, fxp.One),
				),
			)
		},
	}
	for name, shape := range shapes {
		for _, traitCount := range []int{0, 8, 16, 24, 32, 40, 48, 56} {
			entity := gurps.NewEntity()
			for i := range traitCount {
				trait := gurps.NewTrait(nil, nil, false)
				trait.Name = fmt.Sprintf("Trait %d", i+1)
				entity.Traits = append(entity.Traits, trait)
			}
			for i := range 2 {
				skill := gurps.NewSkill(nil, nil, false)
				skill.Name = fmt.Sprintf("Skill %d", i+1)
				entity.Skills = append(entity.Skills, skill)
				note := gurps.NewNote(nil, nil, false)
				note.MarkDown = fmt.Sprintf("Note %d", i+1)
				entity.Notes = append(entity.Notes, note)
			}
			layout := &gurps.SheetLayout{Root: shape()}
			keys := layout.VisibleKeys()
			for _, key := range gurps.AllBlockKeys {
				if !slices.Contains(keys, key) {
					layout.Hidden = append(layout.Hidden, key)
				}
			}
			c.True(layout.SetMinHeight(gurps.BlockSkillsKey, paper.Length{Length: 3, Units: paper.Inch}))
			layout.EnsureValidity()
			entity.SheetSettings.Layout = layout

			exporter := newPageExporter(entity)
			pageSize := exporter.PageSize()
			for i, page := range exporter.pages {
				bottom := pageSize.Height - page.insets().Bottom
				for _, band := range page.Children() {
					c.True(band.FrameRect().Bottom() <= bottom+0.5,
						"with %d traits %s, a band of page %d ends at %v, past the printable bottom of %v",
						traitCount, name, i+1, band.FrameRect().Bottom(), bottom)
				}
			}
		}
	}
}

// TestPageExporterFitsAnEmptyCharacterOnOnePage verifies that a character with no content exports as a single page
// whose every band came from the block layout.
func TestPageExporterFitsAnEmptyCharacterOnOnePage(t *testing.T) {
	c := check.New(t)
	exporter := newPageExporter(gurps.NewEntity())
	c.Equal(1, len(exporter.pages), "a character with no content takes a single page")
	children := exporter.pages[0].Children()
	c.True(len(children) > 1, "the page holds the bands of the layout")
	for i, child := range children {
		c.NotNil(layoutNodeOf(child), "band %d must record the layout node it was built from", i)
	}
}

// TestPageExporterMovesABlockThatDoesNotFit verifies the user's requirement that a block which isn't a list is never
// split: one that can't fit in what is left of a page is moved to the next one whole, and only appears there.
func TestPageExporterMovesABlockThatDoesNotFit(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	for i := range 5 {
		trait := gurps.NewTrait(nil, nil, false)
		trait.Name = fmt.Sprintf("Trait %d", i+1)
		entity.Traits = append(entity.Traits, trait)
	}
	// A layout of just the traits list and a body block too tall for any page to hold along with anything else.
	layout := &gurps.SheetLayout{
		Root: testContainerNode(layoutnode.Column, fxp.One,
			testBlockNode(gurps.BlockTraitsKey, fxp.One),
			testBlockNode(gurps.BlockBodyKey, fxp.One),
		),
	}
	for _, key := range gurps.AllBlockKeys {
		if key != gurps.BlockTraitsKey && key != gurps.BlockBodyKey {
			layout.Hidden = append(layout.Hidden, key)
		}
	}
	c.True(layout.SetMinHeight(gurps.BlockBodyKey, paper.Length{Length: 20, Units: paper.Inch}))
	layout.EnsureValidity()
	entity.SheetSettings.Layout = layout

	exporter := newPageExporter(entity)
	c.Equal(2, len(exporter.pages), "the body block can't share the first page with the traits")
	c.Equal(1, countBlockPanels(exporter, gurps.BlockBodyKey), "and must appear exactly once")
	c.Equal(gurps.BlockTraitsKey, pageKeyOf(exporter.pages[0].Children()[0]))
	c.Equal(1, len(exporter.pages[1].Children()))
	c.Equal(gurps.BlockBodyKey, pageKeyOf(exporter.pages[1].Children()[0]),
		"the body block must be the first thing on the second page")
}

// TestPageExporterForTemplateAndLoot verifies that the kinds of sheet that have no character behind them still export.
func TestPageExporterForTemplateAndLoot(t *testing.T) {
	c := check.New(t)
	data := gurps.NewTemplate()
	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Template Trait"
	data.Traits = []*gurps.Trait{trait}
	exporter := newPageExporter(pageInfoProviderFor(newTestTemplateDockable("My Template", data)))
	c.Equal(1, len(exporter.pages))
	c.Equal(1, countBlockPanels(exporter, gurps.BlockTraitsKey), "the template's traits are on the page")
	c.Equal(0, countBlockPanels(exporter, gurps.BlockBodyKey), "a block that only a character has is not")

	lootExporter := newPageExporter(gurps.NewLoot())
	c.Equal(1, len(lootExporter.pages))
	c.True(len(lootExporter.pages[0].Children()) > 0, "the loot sheet's top block is on the page")
	c.Equal(0, countBlockPanels(lootExporter, gurps.BlockBodyKey))
}

// drawnRowRanges returns the row range each page draws of the block with the given key, in page order.
func drawnRowRanges(pages []*Page, key string) [][2]int {
	var ranges [][2]int
	for _, page := range pages {
		for _, panel := range blockPanels(page.AsPanel(), key) {
			if helper, ok := panel.Self.(pageHelper); ok {
				start, endBefore := helper.CurrentDrawRowRange()
				ranges = append(ranges, [2]int{start, endBefore})
			}
		}
	}
	return ranges
}

// countBlockPanels returns the number of panels showing the block with the given key across all of the pages.
func countBlockPanels(p *pageExporter, key string) int {
	count := 0
	for _, page := range p.pages {
		count += len(blockPanels(page.AsPanel(), key))
	}
	return count
}

// blockPanels returns the panels showing the block with the given key within the given panel.
func blockPanels(panel *unison.Panel, key string) []*unison.Panel {
	return uxtest.PanelsMatching(panel, func(p *unison.Panel) bool { return pageKeyOf(p) == key })
}

// TestPageExporterSplitsBothListsOfARow verifies that a list too long for a single page is split across as many as it
// takes, with every row drawn exactly once, both when the list beside it in its band is empty and when that one is long
// as well, in which case each is split at its own row, the way the character sheet has always paginated a row of two
// blocks.
func TestPageExporterSplitsBothListsOfARow(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		name   string
		traits int
		skills int
		split  []string // The keys of the lists that must be split
	}{
		{name: "beside an empty list", traits: 200, split: []string{gurps.BlockTraitsKey}},
		{
			name:   "beside another long list",
			traits: 150,
			skills: 150,
			split:  []string{gurps.BlockTraitsKey, gurps.BlockSkillsKey},
		},
	} {
		entity := gurps.NewEntity()
		for i := range one.traits {
			trait := gurps.NewTrait(nil, nil, false)
			trait.Name = fmt.Sprintf("Trait %d", i+1)
			entity.Traits = append(entity.Traits, trait)
		}
		for i := range one.skills {
			skill := gurps.NewSkill(entity, nil, false)
			skill.Name = fmt.Sprintf("Skill %d", i+1)
			entity.Skills = append(entity.Skills, skill)
		}
		entity.Recalculate()
		// A new character may come with a trait of its own already, so the counts are taken from the character itself.
		rowCounts := map[string]int{
			gurps.BlockTraitsKey: len(entity.Traits),
			gurps.BlockSkillsKey: len(entity.Skills),
		}

		exporter := newPageExporter(entity)
		for _, key := range one.split {
			ranges := drawnRowRanges(exporter.pages, key)
			c.True(len(ranges) > 1, "%s: the %s list must be split across pages", one.name, key)
			next := 0
			for i, span := range ranges {
				c.Equal(next, span[0], "%s: the %s rows drawn on page %d must pick up where the page before left off",
					one.name, key, i+1)
				c.True(span[1] > span[0], "%s: page %d must draw at least one %s row", one.name, i+1, key)
				next = span[1]
			}
			c.Equal(rowCounts[key], next, "%s: every %s row must be drawn", one.name, key)
		}
	}
}

// TestPageExporterContinuesAListWithViewsOfItsTable verifies that the pages a list is split over show the one table:
// the panel each later page holds for the list, under the block's key, is a continuation of the list on the first page
// holding a view of its table of its own, sharing its columns. Building a fresh list for each page fails it.
func TestPageExporterContinuesAListWithViewsOfItsTable(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	for i := range 200 {
		trait := gurps.NewTrait(nil, nil, false)
		trait.Name = fmt.Sprintf("Trait %d", i+1)
		entity.Traits = append(entity.Traits, trait)
	}
	layout := &gurps.SheetLayout{
		Root: testContainerNode(layoutnode.Column, fxp.One, testBlockNode(gurps.BlockTraitsKey, fxp.One)),
	}
	for _, key := range gurps.AllBlockKeys {
		if key != gurps.BlockTraitsKey {
			layout.Hidden = append(layout.Hidden, key)
		}
	}
	layout.EnsureValidity()
	entity.SheetSettings.Layout = layout

	exporter := newPageExporter(entity)
	c.True(len(exporter.pages) > 2, "the traits must take more than two pages")
	first := blockPanels(exporter.pages[0].AsPanel(), gurps.BlockTraitsKey)
	if len(first) != 1 {
		t.Fatalf("the first page must hold the traits list once, not %d times", len(first))
	}
	list, ok := first[0].Self.(*PageList[*gurps.Trait])
	if !ok {
		t.Fatal("the first page must hold the traits list itself")
	}
	views := []*unison.Table[*Node[*gurps.Trait]]{list.Table}
	for i, page := range exporter.pages[1:] {
		number := i + 2
		panels := blockPanels(page.AsPanel(), gurps.BlockTraitsKey)
		if len(panels) != 1 {
			c.Errorf("page %d must hold the traits once under the block's key, not %d times", number, len(panels))
			continue
		}
		cont, isCont := panels[0].Self.(*pageListContinuation[*gurps.Trait])
		if !isCont {
			c.Errorf("page %d must hold a continuation of the traits list", number)
			continue
		}
		c.True(unison.SameTable(list.Table, cont.Table), "page %d must show a view of the first page's table", number)
		c.False(slices.Contains(views, cont.Table), "page %d must show a view of its own", number)
		c.Equal(list.Table.Columns, cont.Table.Columns, "page %d must show the first page's columns", number)
		views = append(views, cont.Table)
	}
}

// TestPageExporterPaginatesABandGroup verifies that a band that is a Column of its own -- a band group -- paginates the
// way the two bands it was made from would have: the list at the top of it gives up as many rows as fit and the block
// below it follows on the page where the list finishes, since placeColumn works down a stack of blocks one at a time.
func TestPageExporterPaginatesABandGroup(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	for i := range 200 {
		trait := gurps.NewTrait(nil, nil, false)
		trait.Name = fmt.Sprintf("Trait %d", i+1)
		entity.Traits = append(entity.Traits, trait)
	}
	for i := range 3 {
		note := gurps.NewNote(entity, nil, false)
		note.MarkDown = fmt.Sprintf("Note %d", i+1)
		entity.Notes = append(entity.Notes, note)
	}
	// A layout of a single band holding the traits list with the notes list stacked below it.
	layout := &gurps.SheetLayout{
		Root: testContainerNode(layoutnode.Column, fxp.One,
			testContainerNode(layoutnode.Column, fxp.One,
				testBlockNode(gurps.BlockTraitsKey, fxp.One),
				testBlockNode(gurps.BlockNotesKey, fxp.One),
			),
		),
	}
	for _, key := range gurps.AllBlockKeys {
		if key != gurps.BlockTraitsKey && key != gurps.BlockNotesKey {
			layout.Hidden = append(layout.Hidden, key)
		}
	}
	layout.EnsureValidity()
	c.Equal(1, len(layout.Root.Children), "the band group must have survived validation as a single band")
	c.Equal(layoutnode.Column, layout.Root.Children[0].Type)
	entity.SheetSettings.Layout = layout

	exporter := newPageExporter(entity)
	c.True(len(exporter.pages) > 1, "the traits don't fit on a single page")
	c.Equal(columnBandPanel, bandKindOf(exporter.pages[0].Children()[0]), "the band is placed as a column")
	ranges := drawnRowRanges(exporter.pages, gurps.BlockTraitsKey)
	c.True(len(ranges) > 1, "the traits list must be split across pages")
	next := 0
	for i, one := range ranges {
		c.Equal(next, one[0], "the rows drawn on page %d must pick up where the page before it left off", i+1)
		c.True(one[1] > one[0], "page %d must draw at least one row", i+1)
		next = one[1]
	}
	c.Equal(len(entity.Traits), next, "every trait must be drawn")

	c.Equal(1, countBlockPanels(exporter, gurps.BlockNotesKey), "the notes list must appear exactly once")
	c.Equal([][2]int{{0, len(entity.Notes)}}, drawnRowRanges(exporter.pages, gurps.BlockNotesKey),
		"the notes list must draw all of its rows")
	last := len(exporter.pages) - 1
	c.Equal(1, len(blockPanels(exporter.pages[last].AsPanel(), gurps.BlockNotesKey)),
		"the notes list must follow the traits on the page the traits finish on")
	c.Equal(1, len(blockPanels(exporter.pages[last].AsPanel(), gurps.BlockTraitsKey)),
		"the last page must hold the end of the traits list")
}

// TestPageExporterKeepsTheSlotsOfFinishedBlocks verifies that a list continuing onto later pages keeps the place it had
// on the first: beside a block that finished on the first page, whether a list or not, it keeps its width, the finished
// block's slot in the Row standing empty, while below one in a Column it moves up to the top of the page. Handing back
// nothing rather than a page spacer for a finished block, or a Row dropping its spacers, widens the continuing list to
// the full page, and a Column keeping them leaves an empty slot above it.
func TestPageExporterKeepsTheSlotsOfFinishedBlocks(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		name      string
		container layoutnode.Type
		finished  string // The key of the block that finishes on the first page
		keepsSlot bool   // Whether the finished block's slot is kept on the later pages
	}{
		{name: "beside a list", container: layoutnode.Row, finished: gurps.BlockTraitsKey, keepsSlot: true},
		{name: "beside a block", container: layoutnode.Row, finished: gurps.BlockLiftingKey, keepsSlot: true},
		{name: "below a list", container: layoutnode.Column, finished: gurps.BlockTraitsKey},
	} {
		entity := gurps.NewEntity()
		for i := range 3 {
			trait := gurps.NewTrait(nil, nil, false)
			trait.Name = fmt.Sprintf("Trait %d", i+1)
			entity.Traits = append(entity.Traits, trait)
		}
		for i := range 200 {
			skill := gurps.NewSkill(entity, nil, false)
			skill.Name = fmt.Sprintf("Skill %d", i+1)
			entity.Skills = append(entity.Skills, skill)
		}
		entity.Recalculate()
		layout := &gurps.SheetLayout{
			Root: testContainerNode(layoutnode.Column, fxp.One,
				testContainerNode(one.container, fxp.One,
					testBlockNode(one.finished, fxp.One),
					testBlockNode(gurps.BlockSkillsKey, fxp.One),
				),
			),
		}
		for _, key := range gurps.AllBlockKeys {
			if key != one.finished && key != gurps.BlockSkillsKey {
				layout.Hidden = append(layout.Hidden, key)
			}
		}
		layout.EnsureValidity()
		entity.SheetSettings.Layout = layout

		exporter := newPageExporter(entity)
		c.True(len(exporter.pages) > 1, "%s: the skills must continue onto a second page", one.name)
		c.Equal(1, countBlockPanels(exporter, one.finished), "%s: the %s block must appear exactly once", one.name,
			one.finished)
		c.Equal(1, len(blockPanels(exporter.pages[0].AsPanel(), one.finished)),
			"%s: the %s block must finish on the first page", one.name, one.finished)
		first := blockPanels(exporter.pages[0].AsPanel(), gurps.BlockSkillsKey)
		if len(first) != 1 {
			c.Errorf("%s: the skills must start on the first page", one.name)
			continue
		}
		width := first[0].FrameRect().Width
		for i, page := range exporter.pages[1:] {
			number := i + 2
			lists := blockPanels(page.AsPanel(), gurps.BlockSkillsKey)
			if len(lists) != 1 {
				c.Errorf("%s: page %d must hold the skills list once, not %d times", one.name, number, len(lists))
				continue
			}
			list := lists[0]
			c.Equal(width, list.FrameRect().Width, "%s: the skills list must keep its width on page %d", one.name,
				number)
			spacers := uxtest.PanelsMatching(page.AsPanel(), isPageSpacer)
			if one.keepsSlot {
				c.Equal(1, len(spacers), "%s: page %d must hold one empty slot", one.name, number)
				siblings := list.Parent().Children()
				c.True(len(siblings) == 2 && isPageSpacer(siblings[0]) && pageKeyOf(siblings[0]) == "" &&
					siblings[1] == list, "%s: on page %d the %s block's slot must stand empty left of the skills",
					one.name, number, one.finished)
			} else {
				c.Equal(0, len(spacers), "%s: page %d must hold no empty slot", one.name, number)
				c.Equal([]*unison.Panel{list}, page.Children(), "%s: the skills list must start page %d", one.name,
					number)
			}
		}
	}
}

// TestPageExporterLaysPagesOutAtTheCharactersPageSize verifies that a character's pages are laid out at the page size
// its own sheet settings call for. Measuring the page before the exporter knows which character it is exporting fails
// it, laying the pages out at the size the global settings call for instead.
func TestPageExporterLaysPagesOutAtTheCharactersPageSize(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	entity.SheetSettings.Page.Orientation = paper.Landscape
	if gurps.SheetSettingsFor(nil).Page.Orientation == paper.Landscape {
		entity.SheetSettings.Page.Orientation = paper.Portrait
	}
	exporter := newPageExporter(entity)
	size := exporter.PageSize()
	for i, page := range exporter.pages {
		c.Equal(size, page.FrameRect().Size, "page %d must be laid out at the character's page size", i+1)
	}
}

// TestPageExporterWritesAnImagePerPage verifies that exporting as images writes one image per page, numbered from 1
// after the file name chosen less its extension, and puts the theme back once it has drawn the pages in the light one.
func TestPageExporterWritesAnImagePerPage(t *testing.T) {
	c := check.New(t)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().General.ImageResolution, gurps.ImageResolutionMin)
	saved := unison.CurrentThemeMode()
	t.Cleanup(func() {
		unison.SetThemeMode(saved)
		unison.ThemeChanged()
		unison.RebuildDynamicColors()
	})
	unison.SetThemeMode(thememode.Dark)
	entity := gurps.NewEntity()
	for i := range 100 {
		trait := gurps.NewTrait(nil, nil, false)
		trait.Name = fmt.Sprintf("Trait %d", i+1)
		entity.Traits = append(entity.Traits, trait)
	}
	exporter := newPageExporter(entity)
	c.True(len(exporter.pages) > 1, "the traits must take more than one page")
	dir := t.TempDir()
	c.NoError(exporter.exportAsPNGs(filepath.Join(dir, "Sheet.png")))
	c.Equal(thememode.Dark, unison.CurrentThemeMode(), "the theme must be put back")

	want := make([]string, 0, len(exporter.pages))
	for i := range exporter.pages {
		want = append(want, fmt.Sprintf("Sheet-%d.png", i+1))
	}
	slices.Sort(want)
	entries, err := os.ReadDir(dir)
	c.NoError(err)
	got := make([]string, 0, len(entries))
	for _, entry := range entries {
		got = append(got, entry.Name())
	}
	c.Equal(want, got, "one image per page, named after the file chosen")
	data, err := os.ReadFile(filepath.Join(dir, want[0]))
	c.NoError(err)
	_, err = png.DecodeConfig(bytes.NewReader(data))
	c.NoError(err, "the image must be a PNG")
}
