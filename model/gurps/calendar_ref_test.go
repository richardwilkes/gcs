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
	"strings"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestCalendarRefSaveRoundTrip verifies that a saved calendar starts with the header and loads back to the same
// configuration, so that -convert can rewrite calendar files without changing what they describe.
func TestCalendarRefSaveRoundTrip(t *testing.T) {
	c := check.New(t)
	original, err := NewCalendarRefFromFS(embeddedFS, "embedded_data/Gregorian.calendar")
	c.NoError(err)
	p := filepath.Join(t.TempDir(), "Gregorian"+CalendarExt)
	c.NoError(original.Save(p))
	var saved []byte
	saved, err = os.ReadFile(p)
	c.NoError(err)
	c.True(strings.HasPrefix(string(saved), headerPrefix(CalendarDataKind)), "the header is written first: %s", saved)
	reloaded, err := NewCalendarRefFromFS(nil, p)
	c.NoError(err)
	c.Equal(original.Name, reloaded.Name)
	c.Equal(original.Calendar.Config(), reloaded.Calendar.Config())
}
