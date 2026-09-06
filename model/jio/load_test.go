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
	Version int    `json:"version"`
	Name    string `json:"name"`
}

func TestLoadVersionedFile(t *testing.T) {
	c := check.New(t)
	fileSystem := fstest.MapFS{
		"current.json": &fstest.MapFile{Data: fmt.Appendf(nil, `{"version":%d,"name":"ok"}`, jio.CurrentDataVersion)},
		"newer.json":   &fstest.MapFile{Data: fmt.Appendf(nil, `{"version":%d,"name":"too new"}`, jio.CurrentDataVersion+1)},
		"older.json":   &fstest.MapFile{Data: fmt.Appendf(nil, `{"version":%d,"name":"too old"}`, jio.MinimumDataVersion-1)},
		"broken.json":  &fstest.MapFile{Data: []byte(`{"version":`)},
	}

	var data versionedTestData
	c.NoError(jio.LoadVersionedFile(fileSystem, "current.json", &data, &data.Version))
	c.Equal(jio.CurrentDataVersion, data.Version)
	c.Equal("ok", data.Name)

	// The version check must see the value that was just read, not a stale one, and its own message must come through
	// rather than being reported as invalid data.
	data = versionedTestData{}
	err := jio.LoadVersionedFile(fileSystem, "newer.json", &data, &data.Version)
	c.HasError(err)
	c.NotContains(err.Error(), jio.InvalidFileData())
	c.Contains(err.Error(), "newer version")

	data = versionedTestData{}
	err = jio.LoadVersionedFile(fileSystem, "older.json", &data, &data.Version)
	c.HasError(err)
	c.NotContains(err.Error(), jio.InvalidFileData())
	c.Contains(err.Error(), "older version")

	// Decode failures and missing files are both reported as invalid file data, with the cause attached.
	data = versionedTestData{}
	err = jio.LoadVersionedFile(fileSystem, "broken.json", &data, &data.Version)
	c.HasError(err)
	c.Contains(err.Error(), jio.InvalidFileData())
	c.Contains(err.Error(), "unexpected EOF")

	err = jio.LoadVersionedFile(fileSystem, "missing.json", &data, &data.Version)
	c.HasError(err)
	c.Contains(err.Error(), jio.InvalidFileData())
	c.Contains(err.Error(), "missing.json")
}
