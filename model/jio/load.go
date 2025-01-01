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
	"context"
	"io"
	"io/fs"
	"os"

	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/xio"
)

// LoadFromFile loads JSON data from the specified path.
func LoadFromFile(ctx context.Context, path string, data any) error {
	f, err := os.Open(path)
	if err != nil {
		return errs.NewWithCause(path, err)
	}
	defer xio.CloseIgnoringErrors(f)
	return Load(ctx, f, data)
}

// LoadFromFS loads JSON data from the specified filesystem path.
func LoadFromFS(ctx context.Context, fileSystem fs.FS, path string, data any) error {
	f, err := fileSystem.Open(path)
	if err != nil {
		return errs.NewWithCause(path, err)
	}
	defer xio.CloseIgnoringErrors(f)
	return Load(ctx, f, data)
}

// Load JSON data.
func Load(ctx context.Context, r io.Reader, data any) error {
	buffer, err := xio.NewBOMStripper(r)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(buffer)
	decoder.SetContext(ctx)
	decoder.UseNumber()
	if err = decoder.Decode(data); err != nil {
		return errs.Wrap(err)
	}
	return nil
}

// DecompressAndDeserialize decompresses the buffer, then loads JSON data from it.
func DecompressAndDeserialize(data []byte, result any) error {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return errs.Wrap(err)
	}
	if err = json.NewDecoder(gz).Decode(result); err != nil {
		return errs.Wrap(err)
	}
	return nil
}
