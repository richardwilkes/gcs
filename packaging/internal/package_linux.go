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

package internal

import "github.com/richardwilkes/toolbox/errs"

// Just here to silence the linter
var (
	_ image.Image = appImg
	_ image.Image = docImg
)

func platformPackage() error {
	return errs.New("no implementation " + shortAppVersion())
}
