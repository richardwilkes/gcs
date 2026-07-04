// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/early"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xos"
)

func TestDesktopEntryStartupWMClass(t *testing.T) {
	c := check.New(t)
	early.Configure()
	entry := desktopEntry("/opt/gcs/gcs")

	// The StartupWMClass entry must be present and must match the WM_CLASS that Unison sets on our windows
	// (xos.AppIdentifier); otherwise window managers cannot associate our windows with this launcher entry. See
	// https://github.com/richardwilkes/gcs/issues/1059.
	c.Contains(entry, "\nStartupWMClass="+xos.AppIdentifier+"\n", "desktop entry StartupWMClass line:\n", entry)

	// Sanity check the rest of the entry still renders as expected.
	c.HasPrefix(entry, "[Desktop Entry]\n", "desktop entry prefix")
	c.Contains(entry, "\nExec=\"/opt/gcs/gcs\" %F\n", "desktop entry Exec line")
}
