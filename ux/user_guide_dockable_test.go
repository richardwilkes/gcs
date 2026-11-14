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
	"go/ast"
	"go/token"
	"strconv"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/userguide"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// TestUserGuideNavigatesInPlace verifies that each way of moving through the User Guide -- a link to another page, to a
// heading on another page or up out of a folder, a link to a heading on the same page, a row of the table of contents,
// chosen or arrived at by a key extending the selection, and a search match -- shows the place in the same view,
// records it as one history entry and selects its page's row, alone, in the table of contents. Arriving at a group's
// row that way records nothing, since a group has no page to show, and leaves that row alone selected. A key that
// moves the cell cursor through the table of contents leaves it in the column it was in, as in any other table.
func TestUserGuideNavigatesInPlace(t *testing.T) {
	c := check.New(t)
	// enterCell moves the cell cursor into the cell of the row the person is on.
	enterCell := func(d *userGuideDockable) {
		d.toc.KeyDownCallback(unison.KeyRight, 0, false)
		c.Equal(0, d.toc.LeadColumnIndex(), "Right must move the cell cursor into the row's cell")
	}
	for _, tc := range []struct {
		name       string
		from       string
		follow     func(d *userGuideDockable)
		wantPage   string
		wantAnchor string
		wantGroup  string // the title of the group whose row is left selected, when the place is one
		inCell     bool   // the cell cursor is left in the first column, rather than at row level
	}{
		{
			name:     "link to another page",
			from:     userguide.HomePage,
			follow:   func(d *userGuideDockable) { d.handleLink(nil, "Traits") },
			wantPage: "Traits.md",
		},
		{
			name:       "link to a heading on another page",
			from:       "Traits.md",
			follow:     func(d *userGuideDockable) { d.handleLink(nil, "Scripting%20Guide#rules-and-limits") },
			wantPage:   "Scripting Guide.md",
			wantAnchor: "rules-and-limits",
		},
		{
			name:     "link up out of a folder",
			from:     "Scripting/Dice.md",
			follow:   func(d *userGuideDockable) { d.handleLink(nil, "../Scripting%20Guide") },
			wantPage: "Scripting Guide.md",
		},
		{
			name:       "link to a heading on the same page",
			from:       userguide.HomePage,
			follow:     func(d *userGuideDockable) { d.markdown.AnchorLinkHandler("table-of-contents") },
			wantPage:   userguide.HomePage,
			wantAnchor: "table-of-contents",
		},
		{
			name: "table of contents row",
			from: userguide.HomePage,
			follow: func(d *userGuideDockable) {
				node := findUserGuideTOCNode(d.toc.RootRows(), "Traits.md")
				if node == nil {
					c.True(false, "Traits must be in the table of contents")
					return
				}
				d.toc.DiscloseRow(node, false)
				d.toc.ClearSelection()
				d.toc.SelectByIndex(d.toc.RowToIndex(node))
			},
			wantPage: "Traits.md",
		},
		{
			name:     "table of contents selection extended down a row",
			from:     "Traits.md",
			follow:   func(d *userGuideDockable) { d.toc.KeyDownCallback(unison.KeyDown, mod.Shift, false) },
			wantPage: "Skills.md",
		},
		{
			name:     "table of contents selection extended up to the first row",
			from:     "Traits.md",
			follow:   func(d *userGuideDockable) { d.toc.KeyDownCallback(unison.KeyHome, mod.Shift, false) },
			wantPage: userguide.HomePage,
		},
		{
			name:      "table of contents selection extended up onto a group",
			from:      "Traits.md",
			follow:    func(d *userGuideDockable) { d.toc.KeyDownCallback(unison.KeyUp, mod.Shift, false) },
			wantPage:  "Traits.md",
			wantGroup: "Abilities",
		},
		{
			name: "table of contents cell cursor moved down a row",
			from: "Traits.md",
			follow: func(d *userGuideDockable) {
				enterCell(d)
				d.toc.KeyDownCallback(unison.KeyDown, 0, false)
			},
			wantPage: "Skills.md",
			inCell:   true,
		},
		{
			name: "table of contents cell cursor extending the selection down a row",
			from: "Traits.md",
			follow: func(d *userGuideDockable) {
				enterCell(d)
				d.toc.KeyDownCallback(unison.KeyDown, mod.Shift, false)
			},
			wantPage: "Skills.md",
			inCell:   true,
		},
		{
			name: "table of contents cell cursor extending the selection up onto a group",
			from: "Traits.md",
			follow: func(d *userGuideDockable) {
				enterCell(d)
				d.toc.KeyDownCallback(unison.KeyUp, mod.Shift, false)
			},
			wantPage:  "Traits.md",
			wantGroup: "Abilities",
			inCell:    true,
		},
		{
			name: "search match",
			from: "Traits.md",
			follow: func(d *userGuideDockable) {
				for _, match := range d.findMatches("table of contents") {
					if match.key == "Home.md#table-of-contents" {
						match.show()
						return
					}
				}
				c.True(false, "the search must find Home's Table of Contents heading")
			},
			wantPage:   userguide.HomePage,
			wantAnchor: "table-of-contents",
		},
	} {
		d := newUserGuideDockable()
		d.navigate(tc.from, "", true)
		before := len(d.history)
		tc.follow(d)
		wantHistory := before + 1
		wantSelected := tc.wantPage
		if tc.wantGroup != "" {
			wantHistory = before
			wantSelected = tc.wantGroup
		}
		c.Equal(tc.wantPage, d.page, "%s: the page shown", tc.name)
		c.Equal(wantHistory, len(d.history), "%s: the history entries", tc.name)
		c.Equal(len(d.history)-1, d.historyPos, "%s: the current history entry", tc.name)
		c.Equal(tc.wantPage, d.history[d.historyPos].page, "%s: the history entry's page", tc.name)
		c.Equal(tc.wantAnchor, d.history[d.historyPos].anchor, "%s: the history entry's anchor", tc.name)
		rows := d.toc.SelectedRows(false)
		selected := make([]string, 0, len(rows))
		for _, row := range rows {
			if row.page == "" {
				selected = append(selected, row.title)
			} else {
				selected = append(selected, row.page)
			}
		}
		c.Equal([]string{wantSelected}, selected, "%s: the table of contents selection", tc.name)
		wantColumn := -1
		if tc.inCell {
			wantColumn = 0
		}
		c.Equal(wantColumn, d.toc.LeadColumnIndex(), "%s: the column the cell cursor is in", tc.name)
	}
}

// TestUserGuideBackAndForwardRestoreTheView verifies that Back and Forward return to the page each history entry holds,
// showing at the top of the view what was there when it was left rather than its anchor or top, even when the page has
// been wrapped to another width or shown at another scale since.
func TestUserGuideBackAndForwardRestoreTheView(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	toggleSideBar := func(d *userGuideDockable) { d.side.toggle(d.content, d.markdown) }
	// Each case makes its change after Equipment is left and reverts it after Traits is left, so that each page is
	// returned to at another width or scale than it was left at, but for the case that changes nothing.
	for _, tc := range []struct {
		name   string
		change func(d *userGuideDockable)
		revert func(d *userGuideDockable)
	}{
		{name: "nothing changed", change: func(*userGuideDockable) {}, revert: func(*userGuideDockable) {}},
		{
			name:   "scale changed",
			change: func(d *userGuideDockable) { d.PerformCmd(nil, Scale200ItemID) },
			revert: func(d *userGuideDockable) { d.PerformCmd(nil, ScaleDefaultItemID) },
		},
		{name: "side bar toggled", change: toggleSideBar, revert: toggleSideBar},
	} {
		screen.Do(func() { showUserGuide("Equipment.md", "") })
		screen.Sync()
		d := uxtest.SoleEditor[*userGuideDockable](t, screen, func(unison.Dockable) bool { return true })
		// headingAtTop scrolls the page shown, without moving in the history, so that the heading is at the top of the
		// view, returning where that is.
		headingAtTop := func(id string) (y float32) {
			screen.Do(func() {
				d.markdown.ScrollToAnchor(id)
				_, y = d.docScroll.Position()
			})
			return y
		}
		// arrive moves through the history, returning the page shown and how far the view is from showing the heading
		// at its top.
		arrive := func(move func(), id string) (page string, off float32) {
			screen.Do(func() {
				move()
				page = d.page
				_, y := d.docScroll.Position()
				d.markdown.ScrollToAnchor(id)
				_, want := d.docScroll.Position()
				off = y - want
			})
			return page, off
		}
		c.True(headingAtTop("managing-equipment") > 0, "%s: the Equipment heading must be below the top", tc.name)
		screen.Do(func() {
			d.handleLink(nil, "Traits")
			tc.change(d)
		})
		screen.Sync()
		c.True(headingAtTop("managing-traits") > 0, "%s: the Traits heading must be below the top", tc.name)

		page, off := arrive(d.back, "managing-equipment")
		c.Equal("Equipment.md", page, "%s: after Back", tc.name)
		c.True(off > -1 && off < 1, "%s: Back returns to the heading at the top of the view, but is %v from it",
			tc.name, off)
		screen.Do(func() { tc.revert(d) })
		screen.Sync()
		page, off = arrive(d.forward, "managing-traits")
		c.Equal("Traits.md", page, "%s: after Forward", tc.name)
		c.True(off > -1 && off < 1, "%s: Forward returns to the heading at the top of the view, but is %v from it",
			tc.name, off)

		uxtest.CloseEditorWithoutPrompt(t, screen, d)
	}
}

// TestUserGuideWrapsAtTheDefaultWidthUntilLaidOut verifies that a page the guide shows before its view has been laid
// out, as when it first opens, is wrapped at the width a markdown file is wrapped at then, which a window of its own is
// sized to fit, rather than to the view, which has no width yet.
func TestUserGuideWrapsAtTheDefaultWidthUntilLaidOut(t *testing.T) {
	c := check.New(t)
	d := newUserGuideDockable()
	d.navigate("Equipment.md", "", true)
	_, pref, _ := d.markdown.Sizes(geom.Size{})
	c.True(pref.Width >= unison.DefaultMarkdownWidth, "the page asks for %v of width, short of the %v it is wrapped at",
		pref.Width, float32(unison.DefaultMarkdownWidth))
}

// TestUserGuideCaretReadsOnFromThePlaceShown verifies that the keys moving the reading caret go on from the place the
// guide was taken to -- a Help button's heading, or where Back and Forward return to -- rather than scrolling the view
// back to the top of the page.
func TestUserGuideCaretReadsOnFromThePlaceShown(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	screen.Do(func() { showUserGuide("Traits.md", "") })
	screen.Sync()
	d := uxtest.SoleEditor[*userGuideDockable](t, screen, func(unison.Dockable) bool { return true })
	screen.Do(func() {
		bar := d.docScroll.Bar(false)
		d.docScroll.SetPosition(0, (bar.Max()-bar.Extent())/2)
	})
	for _, tc := range []struct {
		name   string
		arrive func()
	}{
		{name: "a Help button's heading", arrive: func() { showUserGuide("Equipment.md", "equipment-editor-fields") }},
		{name: "Back", arrive: d.back},
		{name: "Forward", arrive: d.forward},
	} {
		var before, after float32
		screen.Do(tc.arrive)
		screen.Sync()
		screen.Do(func() { _, before = d.docScroll.Position() })
		c.True(before > 0, "%s: the place must be below the top of its page", tc.name)
		screen.KeyPress(unison.KeyDown, 0)
		screen.Do(func() { _, after = d.docScroll.Position() })
		c.Equal(before, after, "%s: Down reads on from the place shown", tc.name)
	}

	uxtest.CloseEditorWithoutPrompt(t, screen, d)
}

// TestUserGuideKeepsTheFocusOnItsPage verifies that the focus is put on the page shown when a Help button brings the
// open guide forward, and when the side bar is hidden while its table of contents has the focus, rather than on the
// guide's first control or outside the guide.
func TestUserGuideKeepsTheFocusOnItsPage(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	screen.Do(func() { showUserGuide("", "") })
	d := uxtest.SoleEditor[*userGuideDockable](t, screen, func(unison.Dockable) bool { return true })
	// focusWithin reports whether the focus is on the panel or within it.
	focusWithin := func(panel unison.Paneler) (within bool) {
		screen.Do(func() {
			focus := wnd.CurrentFocus()
			within = focus != nil && panel.AsPanel().HasInSelfOrDescendants(func(p *unison.Panel) bool {
				return p == focus
			})
		})
		return within
	}

	screen.Do(func() { wnd.SetFocus(Workspace.Navigator) })
	c.False(focusWithin(d), "the focus must start outside the guide")
	screen.Do(func() { showUserGuide("Traits.md", "") })
	c.True(focusWithin(d.markdown), "a Help button bringing the open guide forward puts the focus on its page")

	screen.Do(func() { d.toc.RequestFocus() })
	c.True(focusWithin(d.toc), "the table of contents must have the focus")
	screen.Click(screen.PanelCenter(d.sideBarButton))
	var hidden bool
	screen.Do(func() { hidden = d.side.scroll.Parent() == nil })
	c.True(hidden, "the click must hide the side bar")
	c.True(focusWithin(d.markdown), "hiding the side bar while it has the focus puts the focus on the page")

	uxtest.CloseEditorWithoutPrompt(t, screen, d)
}

// TestUserGuideImagesLoadFromTheEmbeddedGuide verifies that the images a page shows, written in Markdown or as the
// <img> tags its screenshots use, are read from the guide built into GCS, relative to the page's folder, rather than
// standing in as broken images or, for a tag, not being shown at all.
func TestUserGuideImagesLoadFromTheEmbeddedGuide(t *testing.T) {
	c := check.New(t)
	d := newUserGuideDockable()
	for _, tc := range []struct {
		page   string
		images int
		shows  string
	}{
		{page: userguide.HomePage, images: 4, shows: "the toolbar's icons"},
		{page: "Saved Filters.md", images: 13, shows: "7 icons and 6 <img> screenshots"},
	} {
		d.navigate(tc.page, "", true)
		images := uxtest.PanelsOfType[*unison.DrawablePanel](d.markdown.AsPanel())
		c.True(len(images) >= tc.images, "%s shows %s, but only %d images were found", tc.page, tc.shows, len(images))
		for i, image := range images {
			broken := false
			if drawable, ok := image.Drawable.(*unison.DrawableSVG); ok {
				broken = drawable.SVG == unison.BrokenImageSVG
			}
			c.False(broken, "%s: image %d (%q) must load", tc.page, i, image.Accessibility.Name)
		}
	}
}

// TestUserGuideLinkLiteralsNamePages verifies that every "md:User%20Guide/" literal in the source, such as those behind
// the Help buttons, names a page of the guide with its exact case, as the links between the guide's own pages must.
func TestUserGuideLinkLiteralsNamePages(t *testing.T) {
	c := check.New(t)
	const prefix = "md:User%20Guide/"
	found := 0
	walkModuleSources(c, func(fileSet *token.FileSet, rel string, file *ast.File) {
		if strings.HasSuffix(rel, "_test.go") {
			return
		}
		ast.Inspect(file, func(node ast.Node) bool {
			lit, ok := node.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			value, err := strconv.Unquote(lit.Value)
			if err != nil || !strings.HasPrefix(value, prefix) {
				return true
			}
			found++
			_, _, ok = userguide.ResolveLink(userguide.HomePage, strings.TrimPrefix(value, prefix))
			c.True(ok, "%s:%d: %q names no page of the User Guide with that exact case", rel,
				fileSet.Position(lit.Pos()).Line, value)
			return true
		})
	})
	c.True(found > 10, "the walk should have found the Help buttons' links, but only saw %d", found)
}
