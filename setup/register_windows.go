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

package setup

import (
	"os"
	"path/filepath"

	"github.com/richardwilkes/toolbox/errs"
	"golang.org/x/sys/windows/registry"
)

// RegisterFileTypes registers the file extensions GCS knows how to open with the Windows OS.
func RegisterFileTypes() error {
	exePath, err := os.Executable()
	if err != nil {
		return errs.Wrap(err)
	}
	if exePath, err = filepath.Abs(exePath); err != nil {
		return errs.Wrap(err)
	}
	var k registry.Key
	k, _, err = registry.CreateKey(registry.CLASSES_ROOT, ".ggg", registry.READ|registry.WRITE)
	if err != nil {
		return errs.Wrap(err)
	}
	if err = k.Close(); err != nil {
		return errs.Wrap(err)
	}
	return nil
}
