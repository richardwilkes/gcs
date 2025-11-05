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
	"bytes"
	"compress/gzip"
	"encoding/json/v2"
	"io"
	"io/fs"
	"os"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xio"
)

// Load JSON data from the specified filesystem path. 'fileSystem' may be nil, in which case os.Open() is used instead.
func Load(fileSystem fs.FS, path string, result any) error {
	var f fs.File
	var err error
	if fileSystem == nil {
		f, err = os.Open(path)
	} else {
		f, err = fileSystem.Open(path)
	}
	if err != nil {
		return errs.NewWithCause(path, err)
	}
	defer xio.CloseIgnoringErrors(f)
	var r io.Reader
	if r, err = xio.NewBOMStripper(f); err != nil {
		return err
	}
	return errs.Wrap(json.UnmarshalRead(r, result))
}

// DecompressAndDeserialize decompresses the buffer, then loads JSON data from it.
func DecompressAndDeserialize(data []byte, result any) error {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return errs.Wrap(err)
	}
	return errs.Wrap(json.UnmarshalRead(r, result))
}
