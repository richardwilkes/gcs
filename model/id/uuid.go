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

package id

import (
	"github.com/google/uuid"
	"github.com/richardwilkes/toolbox/log/jot"
)

// NewUUID creates a new UUID.
func NewUUID() uuid.UUID {
	id, err := uuid.NewRandom()
	if err != nil {
		jot.Error(err)
		// continue on... the id will be garbage, but we can live with that... and this should not be possible anyway
	}
	return id
}
