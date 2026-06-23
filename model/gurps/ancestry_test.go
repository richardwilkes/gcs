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
	"testing"
	"testing/fstest"

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestNewAncestryFromInvalidFile verifies that loading an ancestry file containing invalid JSON returns an error rather
// than crashing the application. See issue #1007, where a missing comma in an ancestry file caused GCS to exit when
// randomizing a value.
func TestNewAncestryFromInvalidFile(t *testing.T) {
	c := check.New(t)

	// A well-formed ancestry file loads without error.
	const valid = `{
	"type": "ancestry",
	"version": 5,
	"name": "Human"
}`
	fileSystem := fstest.MapFS{"Human.ancestry": {Data: []byte(valid)}}
	a, err := NewAncestryFromFile(fileSystem, "Human.ancestry")
	c.NoError(err, "valid ancestry file should load")
	c.Equal("Human", a.Name, "name should be loaded")

	// The same content with a missing comma (matching the report in issue #1007) must produce an error instead of
	// panicking or exiting.
	const invalid = `{
	"type": "ancestry"
	"version": 5,
	"name": "Human"
}`
	fileSystem = fstest.MapFS{"Human.ancestry": {Data: []byte(invalid)}}
	a, err = NewAncestryFromFile(fileSystem, "Human.ancestry")
	c.HasError(err, "invalid ancestry file should return an error, not crash")
	c.Equal((*Ancestry)(nil), a, "no ancestry should be returned on failure")
}
