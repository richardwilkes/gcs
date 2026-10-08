// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.
// Package calculators holds the Calculators dockable: the GURPS rules that take some working out at the table, one to
// a tab, each able to take a character's numbers from any open sheet. Linking the package in installs the dockable as
// what ux.OpenCalculators opens, which the Calculators menu item and the sheets' calculator buttons run.
package calculators

import "github.com/richardwilkes/gcs/v5/ux"

func init() {
	ux.OpenCalculators = Display
}
