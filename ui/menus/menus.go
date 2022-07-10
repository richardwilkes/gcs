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

package menus

import (
	"sync"

	"github.com/richardwilkes/gcs/v5/constants"
	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/gcs/v5/ui/about"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

var registerKeyBindingsOnce sync.Once

// Setup the menu bar for the window.
func Setup(wnd *unison.Window) {
	registerKeyBindingsOnce.Do(func() {
		registerFileMenuActions()
		registerEditMenuActions()
		registerItemMenuActions()
		registerLibraryMenuActions()
		registerSettingsMenuActions()
		registerHelpMenuActions()
	})
	settings.Global().KeyBindings.MakeCurrent()
	unison.DefaultMenuFactory().BarForWindow(wnd, func(bar unison.Menu) {
		unison.InsertStdMenus(bar, about.Show, nil, nil)
		std := bar.Item(unison.PreferencesItemID)
		if std != nil {
			std.Menu().RemoveItem(std.Index())
		}
		setupFileMenu(bar)
		setupEditMenu(bar)
		f := bar.Factory()
		i := insertMenu(bar, bar.Item(unison.EditMenuID).Index()+1, createItemMenu(f))
		i = insertMenu(bar, i, f.NewMenu(constants.LibraryMenuID, i18n.Text("Library"), updateLibraryMenu))
		insertMenu(bar, i, createSettingsMenu(f))
		setupHelpMenu(bar)
	})
}

// TODO: Implement each call site
func unimplemented(a *unison.Action, _ any) {
	unison.ErrorDialogWithMessage("Unimplemented Action:", a.Title)
}

func insertSeparator(parent unison.Menu, atIndex int) int {
	parent.InsertSeparator(atIndex, false)
	return atIndex + 1
}

func insertItem(parent unison.Menu, atIndex int, item unison.MenuItem) int {
	parent.InsertItem(atIndex, item)
	return atIndex + 1
}

func insertMenu(parent unison.Menu, atIndex int, menu unison.Menu) int {
	parent.InsertMenu(atIndex, menu)
	return atIndex + 1
}
