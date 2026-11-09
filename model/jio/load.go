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
// used instead. Any 'opts' are applied to the decode. The file's header, if any, is not checked; see LoadVersionedFile
// for that.
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

// readFile returns the content of the file at path from fileSystem, less any byte order mark.
func readFile(fileSystem fs.FS, path string) ([]byte, error) {
	f, err := openFile(fileSystem, path)
	if err != nil {
		return nil, err
	}
	defer xio.CloseIgnoringErrors(f)
	var r io.Reader
	if r, err = xio.NewBOMStripper(f); err != nil {
		return nil, err
	}
	var buf []byte
	if buf, err = io.ReadAll(r); err != nil {
		return nil, errs.NewWithCause(path, err)
	}
	return buf, nil
}

// ReadHeader returns the header members found among the top-level members of the JSON object in data, wherever they
// appear in it. Nothing else is decoded: the other members are skipped, so a nested object's members are never taken
// for the header. A "file_type" member counts only when its value is a string and a "version" member only when its
// value is a number, so a file written before headers existed as a bare map, such as a page reference file, is read as
// one even if a key of its happens to be one of those names. Data that is not a JSON object is an error.
func ReadHeader(data []byte) (Header, error) {
	dec := jsontext.NewDecoder(bytes.NewReader(data))
	tok, err := dec.ReadToken()
	if err != nil {
		return Header{}, errs.Wrap(err)
	}
	if tok.Kind() != '{' {
		return Header{}, errs.New("the data is not a JSON object")
	}
	var h Header
	var hasFileType, hasVersion bool
	for !hasFileType || !hasVersion {
		if tok, err = dec.ReadToken(); err != nil {
			return Header{}, errs.Wrap(err)
		}
		if tok.Kind() == '}' {
			break
		}
		switch name := tok.String(); {
		case name == "file_type" && dec.PeekKind() == '"':
			err = json.UnmarshalDecode(dec, &h.FileType)
			hasFileType = true
		case name == "version" && dec.PeekKind() == '0':
			err = json.UnmarshalDecode(dec, &h.Version)
			hasVersion = true
		default:
			err = dec.SkipValue()
		}
		if err != nil {
			return Header{}, errs.Wrap(err)
		}
	}
	return h, nil
}

// ReadVersionedFile reads the file at path from fileSystem and checks its header with CheckHeader before anything else
// is decoded, returning the header and the file's content for the caller to decode. 'fileSystem' may be nil, in which
// case os.Open() is used instead. A failure to read the file or find a header in it is reported as InvalidFileData(),
// while a header that CheckHeader refuses is reported with its own message.
func ReadVersionedFile(fileSystem fs.FS, path, expectedFileType string, minimum int) (Header, []byte, error) {
	buf, err := readFile(fileSystem, path)
	if err != nil {
		return Header{}, nil, errs.NewWithCause(InvalidFileData(), err)
	}
	var h Header
	if h, err = ReadHeader(buf); err != nil {
		return Header{}, nil, errs.NewWithCause(InvalidFileData(), err)
	}
	if err = CheckHeader(h, expectedFileType, minimum); err != nil {
		return h, nil, err
	}
	return h, buf, nil
}

// LoadVersionedFile is LoadVersionedFileWithMinimum with MinimumDataVersion as the minimum.
func LoadVersionedFile(fileSystem fs.FS, path string, data any, expectedFileType string, opts ...json.Options) error {
	return LoadVersionedFileWithMinimum(fileSystem, path, data, expectedFileType, MinimumDataVersion, opts...)
}

// LoadVersionedFileWithMinimum reads the JSON file at path from fileSystem with ReadVersionedFile, so that data this
// release cannot load is never decoded, then decodes it into data, reporting a failure to do so as InvalidFileData().
// Any 'opts' are applied to the decode.
func LoadVersionedFileWithMinimum(fileSystem fs.FS, path string, data any, expectedFileType string, minimum int,
	opts ...json.Options,
) error {
	_, buf, err := ReadVersionedFile(fileSystem, path, expectedFileType, minimum)
	if err != nil {
		return err
	}
	if err = Unmarshal(buf, data, opts...); err != nil {
		return errs.NewWithCause(InvalidFileData(), err)
	}
	return nil
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
