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
	"github.com/richardwilkes/gcs/v5/constants"
	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// ChangeLibraryLocations brings up the dialog that allows the user to edit the library locations.
var ChangeLibraryLocations *unison.Action

func registerLibraryMenuActions() {
	ChangeLibraryLocations = &unison.Action{
		ID:              constants.ChangeLibraryLocationsItemID,
		Title:           i18n.Text("Change Library Locations"),
		ExecuteCallback: unimplemented,
	}

	settings.RegisterKeyBinding("change_library_locations", ChangeLibraryLocations)
}

func updateLibraryMenu(m unison.Menu) {
	for i := m.Count() - 1; i >= 0; i-- {
		m.RemoveItem(i)
	}
	f := m.Factory()
	m.InsertItem(-1, ChangeLibraryLocations.NewMenuItem(f))
}
