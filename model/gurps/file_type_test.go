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
	"maps"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/uti"
)

// TestFilteredExtensionLists verifies that each of the filtered extension lists selects only the registered
// extensions carrying its flag and returns them in natural sort order, so that a type registered under several
// extensions contributes each of them and a flag on one type never leaks into another list. The model tests don't run
// the ux-layer file-type registration, so the registry is seeded here and restored afterward.
func TestFilteredExtensionLists(t *testing.T) {
	c := check.New(t)
	savedRegistry := maps.Clone(fileTypeRegistry)
	savedKnown := KnownFileTypes
	t.Cleanup(func() {
		fileTypeRegistry = savedRegistry
		KnownFileTypes = savedKnown
	})
	fileTypeRegistry = make(map[string]*FileInfo)
	KnownFileTypes = nil

	(&FileInfo{
		Name:             "GCS Data",
		UTI:              &uti.DataType{Extensions: []string{".test10", ".test2"}},
		IsGCSData:        true,
		IsDeepSearchable: true,
	}).Register()
	(&FileInfo{
		Name: "Openable Only",
		UTI:  &uti.DataType{Extensions: []string{".test1"}},
	}).Register()
	(&FileInfo{
		Name:      "Special",
		UTI:       &uti.DataType{Extensions: []string{".special"}},
		IsSpecial: true,
		IsGCSData: true,
	}).Register()

	c.Equal([]string{".test1", ".test2", ".test10"}, AcceptableExtensions())
	c.Equal([]string{".test2", ".test10"}, DeepSearchableExtensions())
	c.Equal([]string{".special", ".test2", ".test10"}, GCSExtensions())
}
