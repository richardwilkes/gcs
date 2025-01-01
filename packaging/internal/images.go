// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package internal

import (
	"bytes"
	_ "embed"
	"image"

	"github.com/richardwilkes/toolbox/errs"
)

//go:embed embedded/app-1024.png
var appImgBytes []byte
var appImg image.Image

//go:embed embedded/doc-1024.png
var docImgBytes []byte
var docImg image.Image

func loadBaseImages() error {
	var err error
	if appImg, _, err = image.Decode(bytes.NewBuffer(appImgBytes)); err != nil {
		return errs.Wrap(err)
	}
	if docImg, _, err = image.Decode(bytes.NewBuffer(docImgBytes)); err != nil {
		return errs.Wrap(err)
	}
	return nil
}
