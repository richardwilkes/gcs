// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
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
	"encoding/json/jsontext"
	"encoding/json/v2"
	"io"
	"os"
	"path/filepath"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xos"
)

// SaveToFile writes the data as JSON to the given path. Parent directories will be created automatically, if needed.
func SaveToFile(path string, data any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return errs.Wrap(err)
	}
	if err := xos.WriteSafeFile(path, func(w io.Writer) error {
		if err := MarshalWrite(w, data, jsontext.WithIndent("\t"), json.Deterministic(true)); err != nil {
			return err
		}
		// Add a newline at the end of the file for POSIX compliance.
		if _, err := w.Write([]byte{'\n'}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return errs.NewWithCause(path, err)
	}
	return nil
}

// SerializeAndCompress writes the data as JSON into a buffer, then compresses it. Unlike with SaveToFile, no attempt is
// made to make the output "pretty".
func SerializeAndCompress(data any) ([]byte, error) {
	var buffer bytes.Buffer
	gz := gzip.NewWriter(&buffer)
	if err := MarshalWrite(gz, data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, errs.Wrap(err)
	}
	return buffer.Bytes(), nil
}

// Marshal drop-in replacement for json.Marshal that applies any application-wide options
func Marshal(data any, opts ...json.Options) ([]byte, error) {
	var buffer bytes.Buffer
	if err := MarshalWrite(&buffer, data, opts...); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// MarshalWrite drop-in replacement for json.MarshalWrite that applies any application-wide options
func MarshalWrite(w io.Writer, data any, opts ...json.Options) error {
	// NOTE: Any application-wide JSON Marshal options belong here.
	opts = append([]json.Options{}, opts...)
	return errs.Wrap(json.MarshalWrite(w, data, opts...))
}
