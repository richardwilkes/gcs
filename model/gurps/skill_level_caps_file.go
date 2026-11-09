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
	"io/fs"

	"github.com/richardwilkes/gcs/v5/model/jio"
)

// skillLevelCapsData is the on-disk form of a skill level caps file. It carries no fallback for unknown members, since
// only the caps are handed on by NewSkillLevelCapsFromFile, so any other top-level member would be dropped on save in
// any case.
type skillLevelCapsData struct {
	jio.Header
	Caps SkillLevelCaps `json:"skill_level_caps,omitempty"`
}

// NewSkillLevelCapsFromFile loads skill level caps from a file, which may hold none. They are made valid the same way
// the ones in the sheet settings are.
func NewSkillLevelCapsFromFile(fileSystem fs.FS, filePath string) (SkillLevelCaps, error) {
	var data skillLevelCapsData
	if _, err := jio.LoadVersionedFile(fileSystem, filePath, &data, SkillLevelCapsDataKind); err != nil {
		return nil, err
	}
	data.Caps.EnsureValidity()
	return data.Caps, nil
}

// SaveSkillLevelCaps writes skill level caps to a file.
func SaveSkillLevelCaps(caps SkillLevelCaps, filePath string) error {
	return jio.SaveToFile(filePath, &skillLevelCapsData{
		Header: jio.NewHeader(SkillLevelCapsDataKind),
		Caps:   caps,
	})
}
