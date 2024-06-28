// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package message

import "github.com/richardwilkes/toolbox/i18n"

// UnexpectedFileData returns a message indicating that the file does not contain the expected data.
func UnexpectedFileData() string {
	return i18n.Text("This file does not contain the expected data.")
}

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
