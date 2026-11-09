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
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

type loadNewTestData struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func TestLoadNew(t *testing.T) {
	c := check.New(t)
	fileSystem := fstest.MapFS{
		"good.json":   &fstest.MapFile{Data: []byte(`{"name":"widget","count":3}`)},
		"broken.json": &fstest.MapFile{Data: []byte(`{"name":`)},
	}

	data, err := jio.LoadNew[loadNewTestData](fileSystem, "good.json")
	c.NoError(err)
	c.NotNil(data)
	c.Equal("widget", data.Name)
	c.Equal(3, data.Count)

	data, err = jio.LoadNew[loadNewTestData](fileSystem, "broken.json")
	c.HasError(err)
	c.Nil(data)

	data, err = jio.LoadNew[loadNewTestData](fileSystem, "missing.json")
	c.HasError(err)
	c.Nil(data)
	c.Contains(err.Error(), "missing.json")

	// A nil file system reads from the local disk, matching LoadFromFile.
	p := filepath.Join(t.TempDir(), "disk.json")
	c.NoError(os.WriteFile(p, []byte(`{"name":"on disk","count":7}`), 0o600))
	data, err = jio.LoadNew[loadNewTestData](nil, p)
	c.NoError(err)
	c.NotNil(data)
	c.Equal("on disk", data.Name)
	c.Equal(7, data.Count)
}

type versionedTestData struct {
	jio.Header
	Name string `json:"name"`
}

func TestLoadVersionedFile(t *testing.T) {
	const kind = "test"
	current := jio.CurrentDataVersion
	newer := jio.CurrentDataVersion + 1
	for _, tc := range []struct {
		name        string
		file        string // A missing file when empty.
		wantHeader  jio.Header
		wantName    string
		contains    []string
		notContains []string
	}{
		{
			name:       "current version with a type",
			file:       fmt.Sprintf(`{"type":%q,"version":%d,"min_version":%d,"name":"ok"}`, kind, current, current),
			wantHeader: jio.NewHeader(kind),
			wantName:   "ok",
		},
		{
			name:       "current version without a type",
			file:       fmt.Sprintf(`{"version":%d,"name":"ok"}`, current),
			wantHeader: jio.Header{Version: current},
			wantName:   "ok",
		},
		{
			name:       "byte order mark",
			file:       fmt.Sprintf("\ufeff{\"version\":%d,\"name\":\"ok\"}", current),
			wantHeader: jio.Header{Version: current},
			wantName:   "ok",
		},
		{
			name:       "header members after the body",
			file:       fmt.Sprintf(`{"name":"late","min_version":%d,"version":%d,"type":%q}`, current, newer, kind),
			wantHeader: jio.Header{Type: kind, Version: newer, MinVersion: current},
			wantName:   "late",
		},
		{
			name:       "newer version still readable by this release",
			file:       fmt.Sprintf(`{"type":%q,"version":%d,"min_version":%d,"name":"future"}`, kind, newer, current),
			wantHeader: jio.Header{Type: kind, Version: newer, MinVersion: current},
			wantName:   "future",
		},
		{
			name:       "type mismatch before types were checked",
			file:       `{"type":"other","version":4,"name":"old"}`,
			wantHeader: jio.Header{Type: "other", Version: 4},
			wantName:   "old",
		},
		// The header's own messages must come through rather than being reported as invalid data, and the body must
		// not be decoded once the header has been refused.
		{
			name:        "newer version without a minimum reader version",
			file:        fmt.Sprintf(`{"version":%d,"name":"too new"}`, newer),
			contains:    []string{"newer version"},
			notContains: []string{jio.InvalidFileData()},
		},
		{
			name:        "newer minimum reader version",
			file:        fmt.Sprintf(`{"type":%q,"version":%d,"min_version":%d,"name":"too new"}`, kind, newer, newer),
			contains:    []string{"newer version"},
			notContains: []string{jio.InvalidFileData()},
		},
		{
			name:        "older version",
			file:        fmt.Sprintf(`{"version":%d,"name":"too old"}`, jio.MinimumDataVersion-1),
			contains:    []string{"older version"},
			notContains: []string{jio.InvalidFileData()},
		},
		{
			name:        "type mismatch",
			file:        fmt.Sprintf(`{"type":"other","version":%d,"name":"wrong"}`, current),
			contains:    []string{`"other"`, `"test"`},
			notContains: []string{jio.InvalidFileData()},
		},
		// A newer file that fails to decode needs the newer release, so it is not reported as invalid data.
		{
			name:        "decode failure in a newer file",
			file:        fmt.Sprintf(`{"type":%q,"version":%d,"min_version":%d,"name":7}`, kind, newer, current),
			contains:    []string{"newer version"},
			notContains: []string{jio.InvalidFileData()},
		},
		// Decode failures and missing files are reported as invalid file data, with the cause attached.
		{
			name:     "decode failure",
			file:     fmt.Sprintf(`{"version":%d,"name":7}`, current),
			contains: []string{jio.InvalidFileData()},
		},
		{
			name:     "truncated header",
			file:     `{"version":`,
			contains: []string{jio.InvalidFileData(), "unexpected EOF"},
		},
		{
			name:        "not an object",
			file:        fmt.Sprintf(`["type",%q,"version",%d,"min_version",%d]`, kind, newer, newer),
			contains:    []string{jio.InvalidFileData()},
			notContains: []string{"newer version"},
		},
		{
			name:     "missing file",
			contains: []string{jio.InvalidFileData(), "missing.json"},
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
			h, err := jio.LoadVersionedFile(fileSystem, path, &data, kind)
			if len(tc.contains) == 0 {
				c.NoError(err)
				c.Equal(tc.wantHeader, h)
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
		})
	}
}
