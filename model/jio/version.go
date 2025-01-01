// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package jio

import (
	"fmt"

	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
)

// These version numbers are used both as the version of the data files written to disk as well as the major version
// number for GCS library compatibility checking. The Go release of GCS is the first release to synchronize the data
// version and the library major version.
const (
	// CurrentDataVersion holds the current version for data files written with the current release. Note that this is
	// intentionally the same for all data files that GCS processes.
	CurrentDataVersion = 5
	// FirstGoDataVersion is the version that was written by the first Go version of GCS.
	FirstGoDataVersion = 4
	// MinimumDataVersion holds the oldest version for data files that can be loaded. Note that this is intentionally
	// the same for all data files that GCS processes.
	MinimumDataVersion    = 2
	MinimumLibraryVersion = 3 // Note that as of the Go version of GCS, the data and library version are the same.
)

// CheckVersion returns an error if the data version is out of the acceptable range.
func CheckVersion(version int) error {
	if version > CurrentDataVersion {
		return errs.New(txt.Wrap("", fmt.Sprintf(i18n.Text("The data was written with a newer version of %[1]s and cannot be loaded. Please update %[1]s and try again."), cmdline.AppName), 76))
	}
	if version < MinimumDataVersion {
		return errs.New(txt.Wrap("", fmt.Sprintf(i18n.Text("The data was written with an older version of %s and cannot be loaded. You will need to load it with an earlier version that can read this version of the data and write the current format."), cmdline.AppName), 76))
	}
	return nil
}
