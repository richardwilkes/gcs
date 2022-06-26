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

package widget

import (
	"github.com/richardwilkes/unison"
)

// NewCheckBox creates a new check box.
func NewCheckBox(text string, checked bool, applier func(bool)) *unison.CheckBox {
	cb := unison.NewCheckBox()
	cb.Text = text
	if checked {
		cb.State = unison.OnCheckState
	}
	cb.ClickCallback = func() { applier(cb.State == unison.OnCheckState) }
	return cb
}
