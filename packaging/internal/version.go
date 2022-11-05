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

import (
	"strings"

	"github.com/richardwilkes/toolbox/cmdline"
)

func shortAppVersion() string {
	shortVersion := strings.TrimSuffix(cmdline.AppVersion, ".0")
	if strings.IndexByte(shortVersion, '.') == -1 {
		return cmdline.AppVersion
	}
	return shortVersion
}
