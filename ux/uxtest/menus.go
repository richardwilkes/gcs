// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.
package uxtest

import (
	"testing"

	"github.com/richardwilkes/unison"
)

// In a headless session the menus are unison's pure-Go, in-window kind. The window's root panel holds, in this order,
// any open popup menus (newest first), the menu bar, a tooltip if one is showing, and the content. A menu panel, bar or
// popup alike, holds one scroll panel whose content has one child panel per item of the menu, in the menu's own order
// and with the separators included, so the panel for an item is found by its index in the unison.Menu.

// MenuBarPanel returns the in-window menu bar of wnd. It is the one child of the root, other than the content, that
// spans the full width of the window at the top; popups are packed to the width of their items.
func MenuBarPanel(wnd *unison.Window) *unison.Panel {
	root := wnd.Content().Parent()
	if root == nil {
		return nil
	}
	width := root.FrameRect().Width
	for _, child := range root.Children() {
		if child == wnd.Content() {
			continue
		}
		if r := child.FrameRect(); r.X == 0 && r.Y == 0 && r.Width == width {
			return child
		}
	}
	return nil
}

// OpenMenuPopup returns the most recently opened popup menu in wnd, or nil if none is open.
func OpenMenuPopup(wnd *unison.Window) *unison.Panel {
	root := wnd.Content().Parent()
	if root == nil || len(root.Children()) == 0 {
		return nil
	}
	first := root.Children()[0]
	if first == wnd.Content() || first == MenuBarPanel(wnd) {
		return nil
	}
	return first
}

// MenuItemPanels returns the panels of the items of a menu panel, bar or popup, one per item in the menu's order.
func MenuItemPanels(menuPanel *unison.Panel) []*unison.Panel {
	if menuPanel == nil || len(menuPanel.Children()) == 0 {
		return nil
	}
	scroller, ok := menuPanel.Children()[0].Self.(*unison.ScrollPanel)
	if !ok || scroller.Content() == nil {
		return nil
	}
	return scroller.Content().AsPanel().Children()
}

// ChooseMenuBarItem chooses an item from one of the menus in wnd's menu bar the way a user would: it clicks the menu's
// title in the bar, then the item in the popup that opens. Since every injection waits for the application to go quiet,
// the item's handler has finished by the time this returns -- or, for a handler that puts up a modal dialog, the dialog
// is up and idle.
func ChooseMenuBarItem(t *testing.T, screen *unison.HeadlessScreen, wnd *unison.Window, menuTitle, itemTitle string) {
	t.Helper()
	var titlePanel *unison.Panel
	var menu unison.Menu
	screen.Do(func() {
		bar := unison.DefaultMenuFactory().BarForWindowNoCreate(wnd)
		if bar == nil {
			return
		}
		items := MenuItemPanels(MenuBarPanel(wnd))
		for i := range bar.Count() {
			item := bar.ItemAtIndex(i)
			if item.Title() == menuTitle && item.SubMenu() != nil && i < len(items) {
				menu = item.SubMenu()
				titlePanel = items[i]
				return
			}
		}
	})
	if titlePanel == nil {
		t.Fatalf("no %q menu in the menu bar", menuTitle)
	}
	screen.Click(screen.PanelCenter(titlePanel))

	var itemPanel *unison.Panel
	screen.Do(func() {
		items := MenuItemPanels(OpenMenuPopup(wnd))
		for i := range menu.Count() {
			if item := menu.ItemAtIndex(i); !item.IsSeparator() && item.Title() == itemTitle && i < len(items) {
				itemPanel = items[i]
				return
			}
		}
	})
	if itemPanel == nil {
		t.Fatalf("no %q item in the open %q menu", itemTitle, menuTitle)
	}
	screen.Click(screen.PanelCenter(itemPanel))
}

// ChoosePopupItem chooses the item at index from a unison.PopupMenu (or anything wrapping one, such as ux.Popup) the
// way a user would: it clicks the popup, which opens an in-window menu holding one item per entry in the popup's own
// order, then clicks that menu's item. The popup's selection callback has run by the time this returns.
func ChoosePopupItem(t *testing.T, screen *unison.HeadlessScreen, wnd *unison.Window, popup unison.Paneler, index int) {
	t.Helper()
	screen.Click(screen.PanelCenter(popup))
	var itemPanel *unison.Panel
	screen.Do(func() {
		if items := MenuItemPanels(OpenMenuPopup(wnd)); index >= 0 && index < len(items) {
			itemPanel = items[index]
		}
	})
	if itemPanel == nil {
		t.Fatalf("no item %d in the menu the popup opened", index)
	}
	screen.Click(screen.PanelCenter(itemPanel))
}
