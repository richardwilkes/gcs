// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestPrepareProfileForContentCache verifies that every profile field placed into the deep search content cache is
// lowercased, since the search text is lowercased before the comparison is made.
func TestPrepareProfileForContentCache(t *testing.T) {
	c := check.New(t)
	content := prepareProfileForContentCache(&gurps.Profile{
		ProfileRandom: gurps.ProfileRandom{
			Name:       "Conan",
			Age:        "Thirty",
			Birthday:   "January 1",
			Eyes:       "Blue",
			Hair:       "Black",
			Skin:       "Tan",
			Handedness: "Right",
			Gender:     "Male",
		},
		PlayerName:   "Robert",
		Title:        "King Of Aquilonia",
		Organization: "The Black Dragons",
		Religion:     "Crom",
	})
	for _, one := range []string{
		"conan",
		"thirty",
		"january 1",
		"blue",
		"black",
		"tan",
		"right",
		"male",
		"robert",
		"king of aquilonia",
		"the black dragons",
		"crom",
	} {
		c.True(strings.Contains(content, one), "expected %q in the cached content, got %q", one, content)
	}
}

// TestDeepSearchOfSheetProfileIsCaseInsensitive verifies that a deep search of a character sheet matches profile
// fields regardless of case. The search text is lowercased before comparison, so a character named "Conan" must be
// found when searching for "conan", even though the file name doesn't match.
func TestDeepSearchOfSheetProfileIsCaseInsensitive(t *testing.T) {
	c := check.New(t)
	RegisterGCSFileTypes() // The deep search resolves the file's extension through the file type registry.
	dir := t.TempDir()
	entity := gurps.NewEntity()
	entity.Profile.Name = "Conan"
	entity.Profile.PlayerName = "Robert"
	entity.Profile.Organization = "The Black Dragons"
	fileName := "sheet1" + gurps.SheetExt
	c.NoError(entity.Save(filepath.Join(dir, fileName)))

	lib := gurps.NewLibrary("Test", "test", "", "test", dir)
	n := &Navigator{
		searchIndex: -1,
		deepSearch:  map[string]bool{gurps.SheetExt: true},
	}
	node := NewFileNode(lib, fileName, nil)
	for _, one := range []struct {
		name string
		text string
	}{
		{name: "character name", text: "conan"},
		{name: "partial character name", text: "cona"},
		{name: "player name", text: "robert"},
		{name: "organization", text: "the black dragons"},
	} {
		n.searchResult = nil
		n.search(one.text, []*NavigatorNode{node})
		c.Equal(1, len(n.searchResult), one.name)
	}

	// Something that isn't present must still not match.
	n.searchResult = nil
	n.search("belit", []*NavigatorNode{node})
	c.Equal(0, len(n.searchResult))
}
