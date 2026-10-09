// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.
package ux

// SheetSourceUser is implemented by the dockables that draw their numbers from open character sheets, such as the
// calculators.
type SheetSourceUser interface {
	// SheetChanged tells the dockable that the sheet's numbers have changed. Nothing announces a sheet closing, so
	// this is also when a source that names a closed sheet should be noticed and dropped.
	SheetChanged(sheet *Sheet)
}

// NotifySheetSourceUsers brings every open SheetSourceUser up to date with the sheet that has just changed.
func NotifySheetSourceUsers(sheet *Sheet) {
	for _, d := range AllDockables() {
		if user, ok := d.AsPanel().Self.(SheetSourceUser); ok {
			user.SheetChanged(sheet)
		}
	}
}
