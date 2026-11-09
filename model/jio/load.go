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
// used instead. Any 'opts' are applied to the decode.
func LoadFromFile(fileSystem fs.FS, path string, result any, opts ...json.Options) error {
	f, err := openFile(fileSystem, path)
	if err != nil {
		return err
	}
	defer xio.CloseIgnoringErrors(f)
	var r io.Reader
	if r, err = xio.NewBOMStripper(f); err != nil {
		return err
	}
	return UnmarshalRead(r, result, opts...)
}

// openFile opens the file at path from fileSystem, or from the local disk when 'fileSystem' is nil.
func openFile(fileSystem fs.FS, path string) (fs.File, error) {
	var f fs.File
	var err error
	if fileSystem == nil {
		f, err = os.Open(path)
	} else {
		f, err = fileSystem.Open(path)
	}
	if err != nil {
		return nil, errs.NewWithCause(path, err)
	}
	return f, nil
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

// ReadHeader reads the file at path from fileSystem and returns the header members found at the top level of the JSON
// object it holds, wherever they appear in it, along with the file's content, less any byte order mark. 'fileSystem'
// may be nil, in which case os.Open() is used instead. Nothing beyond the header is decoded.
func ReadHeader(fileSystem fs.FS, path string) (Header, []byte, error) {
	f, err := openFile(fileSystem, path)
	if err != nil {
		return Header{}, nil, err
	}
	defer xio.CloseIgnoringErrors(f)
	var r io.Reader
	if r, err = xio.NewBOMStripper(f); err != nil {
		return Header{}, nil, err
	}
	var buf []byte
	if buf, err = io.ReadAll(r); err != nil {
		return Header{}, nil, errs.NewWithCause(path, err)
	}
	var h Header
	dec := jsontext.NewDecoder(bytes.NewReader(buf))
	var tok jsontext.Token
	if tok, err = dec.ReadToken(); err != nil {
		return Header{}, nil, errs.Wrap(err)
	}
	if tok.Kind() != '{' {
		return Header{}, nil, errs.New("the data is not a JSON object")
	}
	var hasType, hasVersion, hasMinVersion bool
	for !hasType || !hasVersion || !hasMinVersion {
		if tok, err = dec.ReadToken(); err != nil {
			return Header{}, nil, errs.Wrap(err)
		}
		if tok.Kind() == '}' {
			break
		}
		switch tok.String() {
		case "type":
			err = json.UnmarshalDecode(dec, &h.Type)
			hasType = true
		case "version":
			err = json.UnmarshalDecode(dec, &h.Version)
			hasVersion = true
		case "min_version":
			err = json.UnmarshalDecode(dec, &h.MinVersion)
			hasMinVersion = true
		default:
			err = dec.SkipValue()
		}
		if err != nil {
			return Header{}, nil, errs.Wrap(err)
		}
	}
	return h, buf, nil
}

// LoadVersionedFile is LoadVersionedFileWithMinimum with MinimumDataVersion as the minimum.
func LoadVersionedFile(fileSystem fs.FS, path string, data any, expectedType string,
	opts ...json.Options,
) (Header, error) {
	return LoadVersionedFileWithMinimum(fileSystem, path, data, expectedType, MinimumDataVersion, opts...)
}

// LoadVersionedFileWithMinimum reads the header of the JSON file at path from fileSystem and checks it with
// CheckHeaderWithMinimum before decoding the file into data, so that data this release cannot load is never decoded.
// A failure to read the file or its header is reported as InvalidFileData(), as is a failure to decode the data, unless
// the file came from a newer release, in which case the failure is reported as needing that newer release. Any 'opts'
// are applied to the decode. The header is returned.
func LoadVersionedFileWithMinimum(fileSystem fs.FS, path string, data any, expectedType string, minimum int,
	opts ...json.Options,
) (Header, error) {
	h, buf, err := ReadHeader(fileSystem, path)
	if err != nil {
		return h, errs.NewWithCause(InvalidFileData(), err)
	}
	var newer bool
	if newer, err = CheckHeaderWithMinimum(h, expectedType, minimum); err != nil {
		return h, err
	}
	if err = Unmarshal(buf, data, opts...); err != nil {
		if newer {
			return h, errs.NewWithCause(newerVersionMessage(), err)
		}
		return h, errs.NewWithCause(InvalidFileData(), err)
	}
	return h, nil
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
