// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/study"
	"github.com/richardwilkes/toolbox/i18n"
)

// Study holds data about a single study session.
type Study struct {
	Type  study.Type `json:"type"`
	Hours fxp.Int    `json:"hours"`
	Note  string     `json:"note,omitempty"`
}

// Clone creates a copy of the TemplatePicker.
func (s *Study) Clone() *Study {
	clone := *s
	return &clone
}

// ResolveStudyHours returns the resolved total study hours.
func ResolveStudyHours(s []*Study) fxp.Int {
	var total fxp.Int
	for _, one := range s {
		total += one.Hours.Mul(one.Type.Multiplier())
	}
	return total
}

// StudyHoursProgressText returns the progress text or an empty string.
func StudyHoursProgressText(hours fxp.Int, needed study.Level, force bool) string {
	if hours <= 0 {
		hours = 0
		if !force {
			return ""
		}
	}
	studyNeeded := "200"
	if needed != study.Standard {
		studyNeeded = needed.Key()
	}
	return fmt.Sprintf(i18n.Text("Studied %v of %s hours"), hours, studyNeeded)
}
