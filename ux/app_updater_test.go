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
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// The updater finds a release's files, and the installation it replaces, by the names the packager gave them, so the
// names it is configured with have to be the ones in packaging.yml, and the port its helper waits on has to be the one
// the single-instance handoff holds.
func TestAppUpdaterConfigMatchesThePackaging(t *testing.T) {
	c := check.New(t)
	data, err := os.ReadFile(filepath.Join("..", "packaging.yml"))
	c.NoError(err)
	for key, want := range map[string]string{
		"executable_name": appUpdateCmdName,
		"finder_app_name": strings.TrimSuffix(appUpdateBundleName, ".app"),
	} {
		c.True(regexp.MustCompile(`(?m)^\s*`+key+`:\s*`+regexp.QuoteMeta(want)+`\s*$`).Match(data),
			"packaging.yml must set %s to %s", key, want)
	}
	cfg := appUpdaterConfig()
	c.Equal(appUpdateCmdName, cfg.CmdName)
	c.Equal(appUpdateBundleName, cfg.BundleName)
	c.Equal(handoffPort, cfg.HandoffPort)
}
