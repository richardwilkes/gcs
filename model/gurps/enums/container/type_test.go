// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package container_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestTypeUnmarshalTextAcceptsOldKeys verifies that a key earlier versions of GCS wrote still loads, whatever its case.
func TestTypeUnmarshalTextAcceptsOldKeys(t *testing.T) {
	c := check.New(t)
	for _, text := range []string{"race", "Race"} {
		var kind container.Type
		c.NoError(kind.UnmarshalText([]byte(text)), text)
		c.Equal(container.Ancestry, kind, text)
	}
}
