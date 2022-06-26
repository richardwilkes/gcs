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

package widget

// Known dockable kinds
const (
	SheetDockableKind    = "sheet"
	TemplateDockableKind = "template"
	ListDockableKind     = "list"
)

// DockableKind defines the method required DockableKinds.
type DockableKind interface {
	DockableKind() string
}
