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
	"os"
	"path/filepath"

	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/xio/fs/safe"
)

// SaveToFile writes the data as JSON to the given path. Parent directories will be created automatically, if needed.
func SaveToFile(ctx context.Context, path string, data any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return errs.Wrap(err)
	}
	if err := safe.WriteFileWithMode(path, func(w io.Writer) error {
		return Save(ctx, w, data)
	}, 0o640); err != nil {
		return errs.NewWithCause(path, err)
	}
	return nil
}

// Save writes the data as JSON.
func Save(ctx context.Context, w io.Writer, data any) error {
	encoder := json.NewEncoder(w)
	encoder.SetContext(ctx)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "\t")
	return errs.Wrap(encoder.Encode(data))
}

// SerializeAndCompress writes the data as JSON into a buffer, then compresses it.
func SerializeAndCompress(data any) ([]byte, error) {
	var buffer bytes.Buffer
	gz := gzip.NewWriter(&buffer)
	if err := json.NewEncoder(gz).Encode(data); err != nil {
		return nil, errs.Wrap(err)
	}
	if err := gz.Close(); err != nil {
		return nil, errs.Wrap(err)
	}
	return buffer.Bytes(), nil
}
