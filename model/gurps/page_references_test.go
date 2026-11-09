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
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// A null in place of a page reference -- which a hand-edited or damaged file may hold, and which decodes without error
// as a nil *PageRef -- is skipped rather than dereferenced.
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

// Importing a page reference file holding a null must report usable data rather than crashing the app. This is
// reachable from the page reference mappings dockable's import.
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

// TestPageRefsFileRoundTrip verifies that a page reference file loads whether it holds the bare map written before such
// files carried a header, even when a book key of its is one of the header's member names, or the current form, and
// that saving writes the current form with the header first and the references intact.
func TestPageRefsFileRoundTrip(t *testing.T) {
	const ref = `{"path":"/refs/basic.pdf","offset":2}`
	for name, tc := range map[string]struct {
		data string
		want []string
	}{
		"bare": {data: `{"B":` + ref + `}`, want: []string{"B"}},
		"bare with header names as keys": {
			data: `{"file_type":` + ref + `,"version":` + ref + `}`,
			want: []string{"file_type", "version"},
		},
		"current": {
			data: fmt.Sprintf(`{"file_type":%q,"version":%d,"refs":{"B":%s}}`, PageRefsDataKind, jio.CurrentDataVersion, ref),
			want: []string{"B"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			c := check.New(t)
			p := filepath.Join(t.TempDir(), "refs.refs")
			c.NoError(os.WriteFile(p, []byte(tc.data), 0o600))
			refs, err := NewPageRefsFromFS(nil, p)
			c.NoError(err)
			c.Equal(tc.want, pageRefIDs(c, refs))

			c.NoError(refs.Save(p))
			var saved []byte
			saved, err = os.ReadFile(p)
			c.NoError(err)
			c.True(strings.HasPrefix(string(saved), headerPrefix(PageRefsDataKind)), "the header is written first: %s",
				saved)
			refs, err = NewPageRefsFromFS(nil, p)
			c.NoError(err)
			c.Equal(tc.want, pageRefIDs(c, refs), "the references survive a save and load")
		})
	}
}

// pageRefIDs returns the IDs of the page references, checking that each points at the test's PDF.
func pageRefIDs(c check.Checker, refs *PageRefs) []string {
	list := refs.List()
	ids := make([]string, 0, len(list))
	for _, one := range list {
		c.Equal("/refs/basic.pdf", one.Path)
		c.Equal(2, one.Offset)
		ids = append(ids, one.ID)
	}
	return ids
}

// A settings file holding a null in place of a page reference or a library must be survivable. Decoding one used to
// panic, which bypassed loadSettingsOrDefaults' recovery entirely and crashed GCS at startup rather than leaving it
// with usable settings.
func TestLoadSettingsOrDefaultsWithNullPageRefEntry(t *testing.T) {
	c := check.New(t)
	countLogs(t, slog.LevelError)
	for _, tc := range []struct {
		name  string
		data  string
		check func(settings *Settings)
	}{
		{
			name: "page reference",
			data: `"page_refs":{"B":null,"BX":{"path":"/refs/basic.pdf","offset":2}}`,
			check: func(settings *Settings) {
				list := settings.PageRefs.List()
				c.Equal(1, len(list), "the null entry was skipped")
				c.Equal("BX", list[0].ID, "the usable page reference entry survived")
			},
		},
		{
			name: "library",
			data: `"libraries":{"a/b":null,"someone/repo":{"title":"Good","path":"/libs/good"}}`,
			check: func(settings *Settings) {
				c.NotNil(settings.Libraries.Lookup("someone/repo"), "the usable library entry survived")
				c.Nil(settings.Libraries.Lookup("a/b"), "the null entry was skipped")
			},
		},
	} {
		p := filepath.Join(t.TempDir(), "settings.json")
		c.NoError(os.WriteFile(p, []byte(`{"last_seen_gcs_version":"1.2.3",`+tc.data+`}`), 0o600))
		var settings Settings
		c.NotPanics(func() { settings = loadSettingsOrDefaults(p) }, tc.name)
		c.Equal("1.2.3", settings.LastSeenGCSVersion, "%s: the rest of the settings file still loaded", tc.name)
		tc.check(&settings)
	}
}
