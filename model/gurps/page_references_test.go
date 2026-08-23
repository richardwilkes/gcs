// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestPageRefsUnmarshalSkipsNullEntries verifies that a null in place of a page reference -- which a hand-edited or
// damaged file may hold, and which decodes without error as a nil *PageRef -- is skipped rather than dereferenced.
func TestPageRefsUnmarshalSkipsNullEntries(t *testing.T) {
	c := check.New(t)
	var refs PageRefs
	c.NotPanics(func() { c.NoError(jio.Unmarshal([]byte(`{"B":null}`), &refs)) })
	c.True(refs.IsZero())
	c.Equal(0, len(refs.List()))

	// A null alongside a usable entry must cost only the null.
	c.NotPanics(func() {
		c.NoError(jio.Unmarshal([]byte(`{"B":null,"BX":{"path":"/refs/basic.pdf","offset":2}}`), &refs))
	})
	list := refs.List()
	c.Equal(1, len(list))
	c.Equal("BX", list[0].ID)
	c.Equal("/refs/basic.pdf", list[0].Path)
	c.Equal(2, list[0].Offset)
}

// TestNewPageRefsFromFSWithNullEntry verifies that importing a page reference file holding a null reports usable data
// rather than crashing the app. This is reachable from the page reference mappings dockable's import.
func TestNewPageRefsFromFSWithNullEntry(t *testing.T) {
	c := check.New(t)
	fileSystem := fstest.MapFS{
		"refs.refs": &fstest.MapFile{
			Data: []byte(`{"B":null,"BX":{"path":"/refs/basic.pdf","offset":2}}`),
		},
	}
	var refs *PageRefs
	var err error
	c.NotPanics(func() { refs, err = NewPageRefsFromFS(fileSystem, "refs.refs") })
	c.NoError(err)
	c.NotNil(refs)
	c.Equal(1, len(refs.List()))
	c.Nil(refs.Lookup("B"))
}

// TestLoadSettingsOrDefaultsWithNullPageRefEntry verifies that a settings file holding a null in place of a page
// reference is survivable. Decoding one used to panic, which bypassed loadSettingsOrDefaults' recovery entirely and
// crashed GCS at startup rather than leaving it with usable settings.
func TestLoadSettingsOrDefaultsWithNullPageRefEntry(t *testing.T) {
	c := check.New(t)
	countErrorLogging(t)
	p := filepath.Join(t.TempDir(), "settings.json")
	c.NoError(os.WriteFile(p, []byte(`{"last_seen_gcs_version":"1.2.3","page_refs":{"B":null,`+
		`"BX":{"path":"/refs/basic.pdf","offset":2}}}`), 0o600))
	var settings Settings
	c.NotPanics(func() { settings = loadSettingsOrDefaults(p) })
	c.Equal("1.2.3", settings.LastSeenGCSVersion, "the rest of the settings file still loaded")
	list := settings.PageRefs.List()
	c.Equal(1, len(list), "the null entry was skipped")
	c.Equal("BX", list[0].ID, "the usable page reference entry survived")
}
