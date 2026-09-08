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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
)

var (
	_ unison.Dockable            = &Calculator{}
	_ unison.TabCloser           = &Calculator{}
	_ unison.UndoManagerProvider = &Calculator{}
	_ sheetSourceUser            = &Calculator{}
)

// calculatorTab is one of the calculators the Calculator dockable holds, each shown on a tab of its own. A calculator
// belongs to no document: whatever numbers it needs are either typed in or taken from any open character sheet.
type calculatorTab interface {
	// title returns what the calculator's tab is labeled.
	title() string
	// panel returns the calculator's content, which is shown while its tab is selected.
	panel() *unison.Panel
	// preselect makes the sheet the source of whichever numbers the calculator most naturally takes from the character
	// it was opened for. It runs once, before the calculator is first shown; changed follows it.
	preselect(sheet *Sheet)
	// sheetChanged tells the calculator that the sheet's numbers have changed. One drawing its numbers from that sheet
	// re-reads them; the rest do nothing.
	sheetChanged(sheet *Sheet)
	// changed re-reads the calculator's sources, brings its dependent controls into line, and recomputes its results.
	// Every control runs it once it has stored its value, and the Calculator runs it whenever the calculator's tab is
	// selected, since sheets may have come and gone while another tab was showing.
	changed()
}

// Calculator holds the calculators for the rules that take some working out at the table, one to a tab: explosions &
// area attacks, scatter and demolition, which come from the same rules, then collisions & falls, jumping, throwing and
// hiking. The explosions come first because they make the tallest panel, so the dockable opens sized for the worst
// case. Only one is open at a time; the sheet it is opened from, if any, is preselected as the source of the numbers
// each calculator most naturally takes from a character.
type Calculator struct {
	unison.Panel
	undoMgr    *unison.UndoManager
	scroll     *unison.ScrollPanel
	slot       *unison.Panel
	tabBar     *tabBar
	tabs       []calculatorTab
	collision  *collisionCalculator
	jumping    *jumpingCalculator
	throwing   *throwingCalculator
	hiking     *hikingCalculator
	explosion  *explosionCalculator
	scatter    *scatterCalculator
	demolition *demolitionCalculator
	scale      int
}

// DisplayCalculator brings the calculators forward, opening them if they are not already open. preselect, when not nil,
// is the sheet each calculator starts out taking its numbers from; it is ignored when the calculators are already open,
// so that re-choosing the menu item or clicking a sheet's calculator button never disturbs what has been entered.
func DisplayCalculator(preselect *Sheet) {
	if activateDockable[*Calculator](nil) {
		return
	}
	c := &Calculator{scale: gurps.GlobalSettings().General.InitialEditorUIScale}
	c.Self = c
	c.undoMgr = unison.NewUndoManager(100, func(err error) { errs.Log(err) })
	c.SetLayout(&unison.FlexLayout{Columns: 1})

	c.collision = newCollisionCalculator()
	c.jumping = newJumpingCalculator()
	c.throwing = newThrowingCalculator()
	c.hiking = newHikingCalculator()
	c.explosion = newExplosionCalculator()
	c.scatter = newScatterCalculator()
	c.demolition = newDemolitionCalculator()
	c.tabs = []calculatorTab{c.explosion, c.scatter, c.demolition, c.collision, c.jumping, c.throwing, c.hiking}

	c.slot = unison.NewPanel()
	c.slot.SetLayout(&columnLayout{})
	c.scroll = unison.NewScrollPanel()
	c.scroll.SetContent(c.slot, behavior.HintedFill, behavior.Fill)
	c.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})

	c.tabBar = newTabBar()
	c.tabBar.SelectionChangedCallback = c.showTab
	for _, tab := range c.tabs {
		c.tabBar.addTab(tab.title())
	}

	c.AddChild(c.createToolbar())
	c.AddChild(c.tabBar)
	c.AddChild(c.scroll)
	for _, tab := range c.tabs {
		if preselect != nil {
			tab.preselect(preselect)
		}
		tab.changed()
	}
	c.tabBar.selectTab(0)
	PlaceInDock(c, dgroup.Editors, false)
	c.slot.RequestFocus()
	// Taking the focus scrolls the control that got it into view, and the dock has not sized the calculators yet, so
	// that scrolls them off the top and side of a view that is still too small. The view is put back at the start,
	// where a first look belongs.
	c.scroll.SetPosition(0, 0)
}

// showTab puts the calculator at the given index into the slot beneath the tab bar, in place of whatever was there,
// and brings it up to date, since sheets may have come and gone while another tab was showing.
func (c *Calculator) showTab(index int) {
	tab := c.tabs[index]
	tab.changed()
	fillSlot(c.slot, tab.panel())
	c.slot.MarkForLayoutRecursively()
	c.slot.MarkForLayoutRecursivelyUpward()
	c.scroll.SetPosition(0, 0)
	c.slot.ValidateScrollRoot()
	c.MarkForRedraw()
}

// sheetChanged implements sheetSourceUser.
func (c *Calculator) sheetChanged(sheet *Sheet) {
	for _, tab := range c.tabs {
		tab.sheetChanged(sheet)
	}
}

func (c *Calculator) createToolbar() *unison.Panel {
	toolbar := newToolbar()
	toolbar.AddChild(NewDefaultInfoPop())
	addUIScaleField(toolbar, func() int { return gurps.GlobalSettings().General.InitialEditorUIScale },
		func() int { return c.scale }, func(scale int) { c.scale = scale }, false, c.scroll)
	finishToolbarLayout(toolbar)
	return toolbar
}

// TitleIcon implements unison.Dockable
func (c *Calculator) TitleIcon(suggestedSize geom.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  svg.Calculator,
		Size: suggestedSize,
	}
}

// Title implements unison.Dockable
func (c *Calculator) Title() string {
	return i18n.Text("Calculators")
}

func (c *Calculator) String() string {
	return c.Title()
}

// Tooltip implements unison.Dockable
func (c *Calculator) Tooltip() string {
	return ""
}

// Modified implements unison.Dockable
func (c *Calculator) Modified() bool {
	return false
}

// MayAttemptClose implements unison.TabCloser
func (c *Calculator) MayAttemptClose() bool {
	return true
}

// AttemptClose implements unison.TabCloser
func (c *Calculator) AttemptClose() bool {
	return AttemptCloseForDockable(c)
}

// UndoManager implements unison.UndoManagerProvider
func (c *Calculator) UndoManager() *unison.UndoManager {
	return c.undoMgr
}
