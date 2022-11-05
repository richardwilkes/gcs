/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ux

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/unison"
)

// CellCache holds data for a table row's cell to reduce the need to constantly recreate them.
type CellCache struct {
	Panel unison.Paneler
	Data  gurps.CellData
	Width float32
}

// Matches returns true if the provided width and data match the current contents.
func (c *CellCache) Matches(width float32, data *gurps.CellData) bool {
	return c != nil && c.Panel != nil && c.Width == width && c.Data == *data
}
