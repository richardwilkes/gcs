// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// A portrait whose bytes the current build cannot decode must not be discarded, since a different build may be able to
// decode it and the data would otherwise be silently dropped by the next save.
func TestProfilePortraitDataSurvivesDecodeFailure(t *testing.T) {
	c := check.New(t)
	data := []byte("this is not a valid image")
	var p gurps.Profile
	p.PortraitData = data
	c.Nil(p.Portrait())
	c.Equal(data, p.PortraitData)

	// A second call must not decode again, but must still leave the data alone.
	c.Nil(p.Portrait())
	c.Equal(data, p.PortraitData)

	// The data must still be written out when saved.
	var buffer bytes.Buffer
	c.NoError(jio.Save(&buffer, &p))
	c.True(strings.Contains(buffer.String(), `"portrait":`))

	// Replacing the data must clear the undecodable state so a valid image can be loaded afterwards.
	p.SetPortraitData(nil)
	c.Nil(p.PortraitData)
	c.Nil(p.Portrait())
}
