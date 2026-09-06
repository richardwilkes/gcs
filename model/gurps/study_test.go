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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/study"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestCloneStudyList verifies the shared study deep-copy: an empty list clones to nil and each session is copied so the
// clone can be edited without touching the source.
func TestCloneStudyList(t *testing.T) {
	c := check.New(t)
	c.Nil(cloneStudyList(nil), "an absent list clones to nil")
	c.Nil(cloneStudyList([]*Study{}), "an empty list clones to nil")

	list := []*Study{{Type: study.Self, Hours: fxp.Four, Note: "Reading"}, {Type: study.Job, Hours: fxp.Two}}
	clone := cloneStudyList(list)
	c.Equal(len(list), len(clone), "every session is carried over")
	for i := range list {
		c.True(list[i] != clone[i], "session %d is a distinct object", i)
		c.Equal(*list[i], *clone[i], "session %d holds the same data", i)
	}
	clone[0].Hours = fxp.One
	c.Equal(fxp.Four, list[0].Hours, "editing the clone leaves the source alone")
}
