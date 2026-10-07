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

// Portrait data the user interface cannot decode must still be written out when saved, since a different build may be
// able to decode it, and replacing the data must discard whatever the user interface cached for the old data.
func TestProfilePortraitData(t *testing.T) {
	c := check.New(t)
	data := []byte("this is not a valid image")
	var p gurps.Profile
	p.PortraitData = data
	p.PortraitCache = "decoded"

	var buffer bytes.Buffer
	c.NoError(jio.MarshalWrite(&buffer, &p))
	c.True(strings.Contains(buffer.String(), `"portrait":`))
	c.Equal(data, p.PortraitData)

	p.SetPortraitData(nil)
	c.Nil(p.PortraitData)
	c.Nil(p.PortraitCache)
}
