// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import "github.com/richardwilkes/toolbox/i18n"

// InvalidFileData returns a message indicating that the file contains invalid data.
func InvalidFileData() string {
	return i18n.Text("Invalid file data.")
}

// NoAdditionalModifiers returns a message indicating that there are no additional modifiers.
func NoAdditionalModifiers() string {
	return i18n.Text("No additional modifiers")
}

// IncludesModifiersFrom returns a message indicating that the current modifiers include modifiers from another source.
func IncludesModifiersFrom() string {
	return i18n.Text("Includes modifiers from")
}

// PageRefTooltip returns the standard tooltip text for a page reference.
func PageRefTooltip() string {
	return i18n.Text(`A reference to the book and page the item appears on e.g. B22 would refer to "Basic Set", page 22`)
}

// LibSrcTooltip returns the standard tooltip text for the library source indicator.
func LibSrcTooltip() string {
	return i18n.Text("Indicates whether the data matches the source library it came from")
}

// ModifierEnabledTooltip returns the standard tooltip text for the modifier enabled indicator.
func ModifierEnabledTooltip() string {
	return i18n.Text("Whether this modifier is enabled. Modifiers that are not enabled do not apply any features they may normally contribute.")
}
