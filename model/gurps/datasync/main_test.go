// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package datasync

import (
	"os"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/gurpstest"
	"github.com/richardwilkes/toolbox/v2/uti"
)

// TestMain raises the per-script execution time limit for the tests (see gurpstest.RaiseScriptExecTimeLimit) and
// registers the trait list file type, standing in for the registration the ux package makes at launch. A conversion
// relies on it to recognize the primary data files and a sync to tell what a library file holds.
func TestMain(m *testing.M) {
	gurpstest.RaiseScriptExecTimeLimit()
	(&gurps.FileInfo{
		Name:      "Traits",
		UTI:       &uti.DataType{Extensions: []string{gurps.TraitsExt}},
		IsGCSData: true,
	}).Register()
	os.Exit(m.Run())
}
