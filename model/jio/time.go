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

package jio

import (
	"time"

	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/errs"
)

// Time is a time.Time that has been tailored to be used with JSON in this application.
type Time time.Time

// Now returns the current local time.
func Now() Time {
	return Time(time.Now())
}

// After reports whether this time is after the other time.
func (e Time) After(other Time) bool {
	return time.Time(e).After(time.Time(other))
}

func (e Time) String() string {
	return time.Time(e).In(time.Local).Format("Jan _2, 2006, 3:04 PM")
}

// MarshalJSON implements json.Marshaler.
func (e Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(e).Format(time.RFC3339))
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *Time) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		// Try older format
		t, err = time.ParseInLocation("Jan _2, 2006, 3:04 PM", s, time.Now().Location())
		if err != nil {
			return errs.New("invalid time/date format: " + s)
		}
	}
	*e = Time(t)
	return nil
}
