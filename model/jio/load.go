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
	"io/fs"
	"os"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xio"
)

// InvalidFileData returns a message indicating that the file contains invalid data.
func InvalidFileData() string {
	return i18n.Text("Invalid file data.")
}

// LoadFromFile loads JSON data from the specified filesystem path. 'fileSystem' may be nil, in which case os.Open() is
// used instead.
func LoadFromFile(fileSystem fs.FS, path string, result any) error {
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
	return UnmarshalRead(r, result)
}

// LoadNew allocates a new T, loads the JSON file at path from fileSystem into it and returns it. As with LoadFromFile,
// 'fileSystem' may be nil to read from the local disk.
func LoadNew[T any](fileSystem fs.FS, path string) (*T, error) {
	var result T
	if err := LoadFromFile(fileSystem, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// LoadVersionedFile loads the JSON file at path from fileSystem into data, reporting any failure to open or decode it
// as InvalidFileData(), then checks that the data version it carries is one this release can load. 'version' must
// point at the version field inside data, so that the check sees the value that was just read.
func LoadVersionedFile(fileSystem fs.FS, path string, data any, version *int) error {
	if err := LoadFromFile(fileSystem, path, data); err != nil {
		return errs.NewWithCause(InvalidFileData(), err)
	}
	return CheckVersion(*version)
}

// DecompressAndDeserialize decompresses the buffer, then loads JSON data from it.
func DecompressAndDeserialize(data []byte, result any) error {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return errs.Wrap(err)
	}
	if err = UnmarshalRead(r, result); err != nil {
		// No need to close the gzip reader on error, since this is all done in memory and has no resources to free.
		return errs.Wrap(err)
	}
	return errs.Wrap(r.Close())
}

// Unmarshal is a drop-in replacement for json.Unmarshal that applies any application-wide options.
func Unmarshal(data []byte, result any, opts ...json.Options) error {
	return UnmarshalRead(bytes.NewReader(data), result, opts...)
}

// UnmarshalRead is a drop-in replacement for json.UnmarshalRead that applies any application-wide options.
func UnmarshalRead(r io.Reader, result any, opts ...json.Options) error {
	opts = append([]json.Options{}, opts...)
	return errs.Wrap(json.UnmarshalRead(r, result, opts...))
}

// UnmarshalStringFrom decodes a JSON string from dec, hands it to parse and stores the result in dst. dst is left
// untouched if either the decode or the parse fails.
func UnmarshalStringFrom[T any](dec *jsontext.Decoder, dst *T, parse func(string) (T, error)) error {
	var s string
	if err := json.UnmarshalDecode(dec, &s); err != nil {
		return err
	}
	v, err := parse(s)
	if err != nil {
		return err
	}
	*dst = v
	return nil
}

// UnmarshalStringFromInfallible is UnmarshalStringFrom for parse functions that cannot fail.
func UnmarshalStringFromInfallible[T any](dec *jsontext.Decoder, dst *T, parse func(string) T) error {
	return UnmarshalStringFrom(dec, dst, func(s string) (T, error) { return parse(s), nil })
}
