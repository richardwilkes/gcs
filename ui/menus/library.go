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
	"fmt"

	"github.com/richardwilkes/gcs/v5/constants"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/toolbox/desktop"
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
	for i, lib := range settings.Global().LibrarySet.List() {
		if !lib.IsUser() {
			m.InsertItem(-1, newUpdateLibraryAction(constants.LibraryBaseItemID+i*2, lib).NewMenuItem(f))
		}
		m.InsertItem(-1, newShowLibraryFolderAction(constants.LibraryBaseItemID+i*2+1, lib).NewMenuItem(f))
		m.InsertSeparator(-1, false)
	}
	m.InsertItem(-1, ChangeLibraryLocations.NewMenuItem(f))
}

func newUpdateLibraryAction(id int, lib *library.Library) *unison.Action {
	action := &unison.Action{ID: id}
	avail := lib.AvailableUpdate()
	switch {
	case avail == nil:
		action.Title = fmt.Sprintf(i18n.Text("Checking for updates to %s"), lib.Title)
		action.EnabledCallback = notEnabled
	case avail.CheckFailed:
		action.Title = fmt.Sprintf(i18n.Text("Unable to access the %s repo"), lib.Title)
		action.EnabledCallback = notEnabled
	case avail.Version == "":
		action.Title = fmt.Sprintf(i18n.Text("No releases available for %s"), lib.Title)
		action.EnabledCallback = notEnabled
	default:
		currentVersion := lib.VersionOnDisk()
		if currentVersion != avail.Version {
			action.Title = fmt.Sprintf(i18n.Text("Update %s to v%s"), lib.Title, avail.Version)
		} else {
			action.Title = fmt.Sprintf(i18n.Text("%s is up to date (re-download v%s)"), lib.Title, currentVersion)
		}
		action.ExecuteCallback = unimplemented
	}
	return action
}

func newShowLibraryFolderAction(id int, lib *library.Library) *unison.Action {
	return &unison.Action{
		ID:    id,
		Title: fmt.Sprintf(i18n.Text("Show %s on Disk"), lib.Title),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			if err := desktop.Open(lib.PathOnDisk); err != nil {
				unison.ErrorDialogWithError(i18n.Text("Unable to show location on disk"), err)
			}
		},
	}
}

func notEnabled(_ *unison.Action, _ any) bool {
	return false
}
