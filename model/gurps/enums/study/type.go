// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package study

import (
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
)

// Multiplier returns the amount to multiply hours spent by for effective study hours.
func (enum Type) Multiplier() fxp.Int {
	switch enum.EnsureValid() {
	case Self:
		return fxp.Half
	case Job:
		return fxp.Quarter
	case Teacher:
		return fxp.One
	case Intensive:
		return fxp.Two
	default:
		errs.Log(errs.New("unknown study type"), "type", int(enum))
		return fxp.One
	}
}

// Limitations returns a list of strings describing the limitations when doing this type of study.
func (enum Type) Limitations() []string {
	switch enum.EnsureValid() {
	case Self:
		return []string{
			i18n.Text("Maximum of 12 hours per day with no job"),
			i18n.Text("Maximum of 8 hours per day with a part-time job"),
			i18n.Text("Maximum of 4 hours per day with a full-time job"),
		}
	case Job:
		return []string{
			i18n.Text("Maximum of 8 hours per day with a full-time job"),
			i18n.Text("Maximum of 4 hours per day with a part-time job"),
		}
	case Teacher:
		return []string{
			i18n.Text("Maximum of 8 hours per day"),
			i18n.Text("The teacher must have the Teaching skill at 12+"),
			i18n.Text("The teacher must have the skill/spell/trait being taught at the same or higher level than you, or have as many or more points spent in it as you do"),
		}
	case Intensive:
		return []string{
			i18n.Text("Maximum of 16 hours per day"),
			i18n.Text("The teacher must have the Teaching skill at 12+"),
			i18n.Text("The teacher must have the skill/spell/trait being taught at a higher level than you and have more points spent in it than you do"),
		}
	default:
		errs.Log(errs.New("unknown study type"), "type", int(enum))
		return nil
	}
}
