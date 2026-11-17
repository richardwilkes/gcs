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
	"testing"

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

// TestPageRefsFileRoundTrip verifies that a page reference file loads whether it holds the bare map written before such
// files carried a header, even when a book key of its is one of the header's member names, or the current form, and
// that saving it keeps the references intact.
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

// TestPageRefsSet verifies that a page reference is stored as a copy, so that neither the one given nor one looked up
// can alter it, and that one without an ID is refused.
func TestPageRefsSet(t *testing.T) {
	c := check.New(t)
	count := countLogs(t, slog.LevelError)
	pdf := filepath.Join(t.TempDir(), "basic.pdf")
	c.NoError(os.WriteFile(pdf, nil, 0o600))
	var refs PageRefs
	ref := &PageRef{ID: "B", Path: pdf, Offset: 2}
	refs.Set(ref)
	ref.Offset = 5
	found := refs.Lookup("B")
	c.NotNil(found)
	c.Equal(2, found.Offset, "changing the reference given doesn't change the one stored")
	found.Offset = 7
	c.Equal(2, refs.Lookup("B").Offset, "changing a reference looked up doesn't change the one stored")
	refs.Set(&PageRef{Path: pdf})
	c.Equal(1, len(refs.List()), "a reference without an ID is refused")
	c.Equal(int32(1), count.Load(), "and the refusal is logged")
}
