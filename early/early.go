// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package early

import "github.com/richardwilkes/toolbox/v2/xos"

// Configure the app data. This is here so that the helper utilities can utilize it as well.
func Configure() {
	xos.AppName = "GCS"
	xos.AppCmdName = "gcs"
	xos.License = "Mozilla Public License, version 2.0"
	xos.CopyrightStartYear = "1998"
	xos.CopyrightHolder = "Richard A. Wilkes"
	xos.AppIdentifier = "com.trollworks.gcs"
}
