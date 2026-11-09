// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package jio_test

import (
	"errors"
	"fmt"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestReadHeader verifies that the header is taken from the top-level members alone, wherever they appear, that a
// member with one of the header's names but another kind of value is passed over as any other member is, so that a
// bare map written before headers existed still reads as one, and that data which is not an object is refused.
func TestReadHeader(t *testing.T) {
	for _, tc := range []struct {
		name string
		data string
		want jio.Header
		fail bool
	}{
		{
			name: "header first",
			data: `{"file_type":"test","version":5,"name":"x"}`,
			want: jio.Header{FileType: "test", Version: 5},
		},
		{
			name: "header after other members",
			data: `{"name":"x","rows":[{"version":1}],"version":4,"file_type":"test"}`,
			want: jio.Header{FileType: "test", Version: 4},
		},
		{name: "nested object is not the header", data: `{"child":{"file_type":"other","version":9}}`},
		{name: "version alone", data: `{"version":5,"name":"x"}`, want: jio.Header{Version: 5}},
		{name: "legacy type member is ignored", data: `{"type":"advantage_list","version":4}`, want: jio.Header{Version: 4}},
		{name: "file type holding an object", data: `{"file_type":{"path":"a"},"version":{"path":"b"}}`},
		{name: "file type holding a number", data: `{"file_type":7,"version":"5"}`},
		{name: "file type holding null", data: `{"file_type":null,"version":null}`},
		{name: "empty object", data: `{}`},
		{name: "array", data: `["file_type","test","version",5]`, fail: true},
		{name: "string", data: `"file_type"`, fail: true},
		{name: "truncated", data: `{"file_type":"test","vers`, fail: true},
		{name: "fractional version", data: `{"version":5.5}`, fail: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			h, err := jio.ReadHeader([]byte(tc.data))
			if tc.fail {
				c.HasError(err)
				return
			}
			c.NoError(err)
			c.Equal(tc.want, h)
		})
	}
}

type versionedTestData struct {
	jio.Header
	Name string `json:"name"`
}

// TestLoadVersionedFile verifies that the header is checked before any of the body is decoded, that files written
// before they carried a header still load, and how each kind of failure is reported: the header's own messages come
// through as they are, while unreadable files and decode failures are reported as invalid file data with the cause
// attached.
func TestLoadVersionedFile(t *testing.T) {
	const kind = "test"
	current := jio.CurrentDataVersion
	for _, tc := range []struct {
		name        string
		file        string // A missing file when empty.
		minimum     int
		wantHeader  jio.Header
		wantName    string
		contains    []string
		notContains []string
		notExist    bool
	}{
		{
			name:       "current version with a file type",
			file:       fmt.Sprintf(`{"file_type":%q,"version":%d,"name":"ok"}`, kind, current),
			minimum:    jio.MinimumDataVersion,
			wantHeader: jio.NewHeader(kind),
			wantName:   "ok",
		},
		{
			name:       "current version without a file type",
			file:       fmt.Sprintf(`{"version":%d,"name":"ok"}`, current),
			minimum:    jio.MinimumDataVersion,
			wantHeader: jio.Header{Version: current},
			wantName:   "ok",
		},
		{
			name:       "byte order mark",
			file:       fmt.Sprintf("\ufeff{\"version\":%d,\"name\":\"ok\"}", current),
			minimum:    jio.MinimumDataVersion,
			wantHeader: jio.Header{Version: current},
			wantName:   "ok",
		},
		{
			name:       "header members after the body",
			file:       fmt.Sprintf(`{"name":"late","version":%d,"file_type":%q}`, current, kind),
			minimum:    jio.MinimumDataVersion,
			wantHeader: jio.NewHeader(kind),
			wantName:   "late",
		},
		{
			name:       "no header with no minimum",
			file:       `{"name":"bare"}`,
			wantHeader: jio.Header{},
			wantName:   "bare",
		},
		// The header's own messages must come through rather than being reported as invalid data, and the body must
		// not be decoded once the header has been refused.
		{
			name:        "file type mismatch",
			file:        fmt.Sprintf(`{"file_type":"other","version":%d,"name":"wrong"}`, current),
			minimum:     jio.MinimumDataVersion,
			contains:    []string{`"other"`, `"test"`},
			notContains: []string{jio.InvalidFileData()},
		},
		{
			name:        "newer version",
			file:        fmt.Sprintf(`{"file_type":%q,"version":%d,"name":"too new"}`, kind, current+1),
			minimum:     jio.MinimumDataVersion,
			contains:    []string{"newer version"},
			notContains: []string{jio.InvalidFileData()},
		},
		{
			// A newer release may have renamed the kind, so updating is the advice to give.
			name:        "newer version of another kind",
			file:        fmt.Sprintf(`{"file_type":"renamed","version":%d,"name":"too new"}`, current+1),
			minimum:     jio.MinimumDataVersion,
			contains:    []string{"newer version"},
			notContains: []string{jio.InvalidFileData(), `"renamed"`},
		},
		{
			name:        "older version",
			file:        fmt.Sprintf(`{"version":%d,"name":"too old"}`, jio.MinimumDataVersion-1),
			minimum:     jio.MinimumDataVersion,
			contains:    []string{"older version"},
			notContains: []string{jio.InvalidFileData()},
		},
		// Decode failures, unreadable headers and missing files are reported as invalid file data, with the cause
		// attached.
		{
			name:     "decode failure",
			file:     fmt.Sprintf(`{"version":%d,"name":7}`, current),
			minimum:  jio.MinimumDataVersion,
			contains: []string{jio.InvalidFileData()},
		},
		{
			name:     "truncated header",
			file:     `{"version":`,
			minimum:  jio.MinimumDataVersion,
			contains: []string{jio.InvalidFileData(), "unexpected EOF"},
		},
		{
			name:        "not an object",
			file:        fmt.Sprintf(`["file_type",%q,"version",%d]`, kind, current+1),
			minimum:     jio.MinimumDataVersion,
			contains:    []string{jio.InvalidFileData()},
			notContains: []string{"newer version"},
		},
		{
			name:     "missing file",
			minimum:  jio.MinimumDataVersion,
			contains: []string{jio.InvalidFileData(), "missing.json"},
			notExist: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			fileSystem := fstest.MapFS{}
			path := "missing.json"
			if tc.file != "" {
				path = "data.json"
				fileSystem[path] = &fstest.MapFile{Data: []byte(tc.file)}
			}
			var data versionedTestData
			err := jio.LoadVersionedFileWithMinimum(fileSystem, path, &data, kind, tc.minimum)
			if len(tc.contains) == 0 {
				c.NoError(err)
				c.Equal(tc.wantHeader, data.Header)
			} else {
				c.HasError(err)
				for _, one := range tc.contains {
					c.Contains(err.Error(), one)
				}
				for _, one := range tc.notContains {
					c.NotContains(err.Error(), one)
				}
			}
			c.Equal(tc.wantName, data.Name)
			c.Equal(tc.notExist, errors.Is(err, fs.ErrNotExist), "a missing file is told apart from the rest")
		})
	}
}
