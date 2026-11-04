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
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestCollegeListLoadsSorted verifies that colleges come back sorted from both the list and the legacy string formats,
// as tags do.
func TestCollegeListLoadsSorted(t *testing.T) {
	c := check.New(t)
	var list CollegeList
	c.NoError(json.Unmarshal([]byte(`["Water", " Air ", "", "Fire"]`), &list))
	c.Equal(CollegeList{"Air", "Fire", "Water"}, list, "list format")
	c.NoError(json.Unmarshal([]byte(`"Water or Air/Fire"`), &list))
	c.Equal(CollegeList{"Air", "Fire", "Water"}, list, "legacy string format")
}
