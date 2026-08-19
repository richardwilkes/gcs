// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import "github.com/richardwilkes/toolbox/v2/i18n"

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

// SwitchHeaderTooltip returns the standard tooltip text for the switch column header.
func SwitchHeaderTooltip() string {
	return i18n.Text("Whether the switchable features of an item are currently on. Features marked as switchable only take effect while the item's switch is on. Items with no switchable features have nothing to switch, so this column is blank for them.")
}

// SwitchCellTooltip returns the standard tooltip text for a switch column cell.
func SwitchCellTooltip() string {
	return i18n.Text("Click to toggle whether this item's switchable features are on. Features marked as switchable only take effect while the item's switch is on. Hold down the Option/Alt key while clicking to also apply the change to everything contained within this item.")
}

// SwitchedOnTooltip returns the standard tooltip text for the "Switched On" editor field.
func SwitchedOnTooltip() string {
	return i18n.Text("Whether the switchable features of this item are currently on. This only matters if at least one feature is marked as switchable, since features marked as switchable only take effect while the item's switch is on.")
}

// SwitchableTooltip returns the standard tooltip text for the "switchable" feature checkbox.
func SwitchableTooltip() string {
	return i18n.Text("When checked, this feature only takes effect while the switch of the trait, skill, spell or piece of equipment it belongs to is on. For a feature on a modifier, that is the item the modifier belongs to. The switch can be toggled from the character sheet or the item's editor.")
}
