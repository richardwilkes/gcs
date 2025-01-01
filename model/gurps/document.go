// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"bytes"
	"compress/gzip"
	"io"

	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/xio"
)

// Document holds the raw data for a document.
type Document struct {
	Name       string `json:"name"`
	Ext        string `json:"ext"`
	Content    []byte `json:"content"`
	Compressed bool   `json:"compressed,omitempty"`
}

// NewDocument creates a new document.
func NewDocument(name, ext string, content []byte, compress bool) *Document {
	if compress {
		var buffer bytes.Buffer
		gz := gzip.NewWriter(&buffer)
		_, err := gz.Write(content)
		if err != nil {
			err = errs.NewWithCause("unable to compress data", err)
		} else {
			if err = gz.Close(); err != nil {
				err = errs.NewWithCause("unable to compress data", err)
			}
		}
		if err != nil {
			compress = false
		} else {
			content = buffer.Bytes()
		}
	}
	return &Document{
		Name:       name,
		Ext:        ext,
		Content:    content,
		Compressed: compress,
	}
}

// UncompressedContent expands the content, if needed, and returns the uncompressed data.
func (d *Document) UncompressedContent() ([]byte, error) {
	if !d.Compressed {
		return d.Content, nil
	}
	gz, err := gzip.NewReader(bytes.NewReader(d.Content))
	if err != nil {
		return nil, errs.Wrap(err)
	}
	defer xio.CloseIgnoringErrors(gz)
	var data []byte
	if data, err = io.ReadAll(gz); err != nil {
		return nil, errs.Wrap(err)
	}
	return data, nil
}
