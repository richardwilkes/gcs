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

// In a headless session the menus are unison's pure-Go kind. The window's root panel holds the menu bar, a tooltip if
// one is showing, and the content. Each open popup menu is shown in a borderless window of its own, which unison keeps
// out of its window list and hands back through HeadlessScreen.MenuWindows. A menu panel, bar or popup alike, holds
// one scroll panel whose content has one child panel per item of the menu, in the menu's own order and with the
// separators included, so the panel for an item is found by its index in the unison.Menu.

// MenuBarPanel returns the in-window menu bar of wnd, or nil if it has none.
func MenuBarPanel(wnd *unison.Window) *unison.Panel {
	root := wnd.Content().Parent()
	if root == nil {
		return nil
	}
	if bar, ok := root.Self.(interface{ MenuBar() *unison.Panel }); ok {
		return bar.MenuBar()
	}
	return nil
}

// OpenMenuPopup returns the panel of the popup menu open for wnd, or nil if none is open. When several are open, as
// when a sub-menu is showing, the outermost is returned; HeadlessScreen.MenuWindows hands back the whole stack. May be
// called from the UI thread or the test's goroutine.
func OpenMenuPopup(wnd *unison.Window) *unison.Panel {
	screen := unison.ActiveHeadlessScreen()
	if screen == nil || wnd == nil {
		return nil
	}
	var popup *unison.Panel
	screen.Do(func() {
		if menus := screen.MenuWindows(wnd); len(menus) != 0 {
			popup = menus[0].Content()
		}
	})
	return popup
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
// title in the bar, then each title of path in turn in the popup that the previous click opened, so every title but the
// last names a sub-menu. Since every injection waits for the application to go quiet, the item's handler has finished
// by the time this returns -- or, for a handler that puts up a modal dialog, the dialog is up and idle.
func ChooseMenuBarItem(t testing.TB, screen *unison.HeadlessScreen, wnd *unison.Window, menuTitle string, path ...string) {
	t.Helper()
	if len(path) == 0 {
		t.Fatal("no item title to choose")
	}
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

	for depth, itemTitle := range path {
		var itemPanel *unison.Panel
		var subMenu unison.Menu
		screen.Do(func() {
			items := MenuItemPanels(innermostMenuPopup(wnd))
			for i := range menu.Count() {
				if item := menu.ItemAtIndex(i); !item.IsSeparator() && item.Title() == itemTitle && i < len(items) {
					itemPanel = items[i]
					subMenu = item.SubMenu()
					return
				}
			}
		})
		if itemPanel == nil {
			t.Fatalf("no %q item in the open %q menu", itemTitle, menu.Title())
		}
		if depth < len(path)-1 && subMenu == nil {
			t.Fatalf("%q in the %q menu has no sub-menu", itemTitle, menu.Title())
		}
		screen.Click(screen.PanelCenter(itemPanel))
		menu = subMenu
	}
}

// innermostMenuPopup returns the panel of the innermost popup menu open for wnd, which is the one a click on a sub-menu
// item opened last, or nil if none is open. Must be called from the UI thread.
func innermostMenuPopup(wnd *unison.Window) *unison.Panel {
	screen := unison.ActiveHeadlessScreen()
	if screen == nil || wnd == nil {
		return nil
	}
	if menus := screen.MenuWindows(wnd); len(menus) != 0 {
		return menus[len(menus)-1].Content()
	}
	return nil
}

// ChoosePopupItem chooses the item at index from a unison.PopupMenu (or anything wrapping one, such as ux.Popup) the
// way a user would: it clicks the popup, which opens an in-window menu holding one item per entry in the popup's own
// order, then clicks that menu's item. The popup's selection callback has run by the time this returns.
func ChoosePopupItem(t testing.TB, screen *unison.HeadlessScreen, wnd *unison.Window, popup unison.Paneler, index int) {
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
