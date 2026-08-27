// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package updatecheck_test

import (
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/updatecheck"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestInterval verifies the repeat interval each option asks for. Only the recurring options produce a non-zero
// interval; a zero tells the caller not to schedule a repeating check at all.
func TestInterval(t *testing.T) {
	c := check.New(t)
	c.Equal(time.Duration(0), updatecheck.AtLaunch.Interval())
	c.Equal(time.Hour, updatecheck.Hourly.Interval())
	c.Equal(24*time.Hour, updatecheck.Daily.Interval())
	c.Equal(time.Duration(0), updatecheck.Never.Interval())
}

// TestChecksAtLaunch verifies that every option other than Never asks for a check when the application starts.
func TestChecksAtLaunch(t *testing.T) {
	c := check.New(t)
	c.True(updatecheck.AtLaunch.ChecksAtLaunch())
	c.True(updatecheck.Hourly.ChecksAtLaunch())
	c.True(updatecheck.Daily.ChecksAtLaunch())
	c.False(updatecheck.Never.ChecksAtLaunch())
}

// TestZeroValueIsAtLaunch verifies that the zero value -- what a settings file written before this setting existed
// loads as -- is the same as the default, so old settings files quietly adopt the intended behavior.
func TestZeroValueIsAtLaunch(t *testing.T) {
	c := check.New(t)
	c.Equal(updatecheck.AtLaunch, updatecheck.Option(0))
	c.Equal(updatecheck.AtLaunch, updatecheck.DefaultOption)
}

// TestExtractOption verifies that a known key round-trips and that an unrecognized one falls back to the default
// rather than yielding an invalid value.
func TestExtractOption(t *testing.T) {
	c := check.New(t)
	c.Equal(updatecheck.Daily, updatecheck.ExtractOption("daily"))
	c.Equal(updatecheck.AtLaunch, updatecheck.ExtractOption("garbage"))
}
