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
	"encoding/json/v2"
	"strings"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestSourcePathSeparatorNormalization verifies that source paths are stored and round-tripped using forward slashes,
// regardless of the separator they were authored with. See issue #1005, where files created on Windows stored source
// paths with backslash separators (e.g. `Basic Set\Basic Set Traits.adq`). Those paths could not be located on other
// platforms, causing the library source match status to show as a question mark.
func TestSourcePathSeparatorNormalization(t *testing.T) {
	c := check.New(t)

	// A path loaded with backslash separators is normalized to forward slashes.
	const windowsAuthored = `{"id":"a","source":{"library":"Master Library","path":"Basic Set\\Basic Set Traits.adq","id":"x"}}`
	var loaded SourcedID
	c.NoError(json.Unmarshal([]byte(windowsAuthored), &loaded), "backslash path should load")
	c.Equal("Basic Set/Basic Set Traits.adq", loaded.Source.Path, "path should be normalized on load")

	// An already-normalized path is left unchanged on load.
	const unixAuthored = `{"id":"a","source":{"library":"Master Library","path":"Basic Set/Basic Set Traits.adq","id":"x"}}`
	var loaded2 SourcedID
	c.NoError(json.Unmarshal([]byte(unixAuthored), &loaded2), "forward-slash path should load")
	c.Equal("Basic Set/Basic Set Traits.adq", loaded2.Source.Path, "forward-slash path should be unchanged")

	// A source carrying a backslash path in memory is always written out with forward slashes.
	inMemory := SourcedID{
		TID: "a",
		Source: Source{
			LibraryFile: LibraryFile{Library: "Master Library", Path: `Basic Set\Basic Set Traits.adq`},
			TID:         "x",
		},
	}
	data, err := json.Marshal(&inMemory)
	c.NoError(err, "source should marshal")
	c.False(strings.Contains(string(data), `\`), "marshaled output must not contain backslashes")
	c.True(strings.Contains(string(data), "Basic Set/Basic Set Traits.adq"), "marshaled output should use forward slashes")

	// Round-tripping the marshaled output preserves the normalized path.
	var roundTripped SourcedID
	c.NoError(json.Unmarshal(data, &roundTripped), "marshaled source should load")
	c.Equal("Basic Set/Basic Set Traits.adq", roundTripped.Source.Path, "round-tripped path should remain normalized")
}
