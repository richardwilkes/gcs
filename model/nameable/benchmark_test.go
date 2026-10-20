// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package nameable_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/nameable"
)

// BenchmarkApply measures replacing the markers in the kinds of strings a sheet's names hold. Run it with:
//
//	GOEXPERIMENT=simd go test ./model/nameable -run '^$' -bench BenchmarkApply -benchmem
func BenchmarkApply(b *testing.B) {
	replacements := map[string]string{"Skill": "Broadsword", "Attack": "Swung", "Specialty": "Fire"}
	for _, one := range []struct {
		name string
		text string
	}{
		{"plain", "Broadsword"},
		{"marker", "Ritual Magic (@Specialty@)"},
		{"multiple", "Targeted Attack (@Skill@ @Attack@/Neck)"},
		{"legacy", "@Skill: Broadsword, Shortsword, etc.@"},
		{"unfilled", "Hobby Skill (@Hobby@)"},
		{"options", "@Skill|Broadsword|Shortsword|*@"},
	} {
		b.Run(one.name, func(b *testing.B) {
			for b.Loop() {
				nameable.Apply(one.text, replacements)
			}
		})
	}
}
