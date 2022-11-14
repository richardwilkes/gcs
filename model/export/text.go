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

package export

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/xio/fs"
)

// ToText exports the files to a text representation.
func ToText(tmplPath string, fileList []string) error {
	for _, one := range fileList {
		switch strings.ToLower(filepath.Ext(one)) {
		case library.SheetExt:
			entity, err := gurps.NewEntityFromFile(os.DirFS(filepath.Dir(one)), filepath.Base(one))
			if err != nil {
				return err
			}
			if err = LegacyExport(entity, tmplPath, fs.TrimExtension(one)+filepath.Ext(tmplPath)); err != nil {
				return err
			}
		default:
			jot.Warn("ignoring: " + one)
		}
	}
	return nil
}
