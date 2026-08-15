// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package updater

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xos"
)

// installed builds a directory shaped the way an installation is on the running platform and returns what an update to
// it would replace.
func installed(t *testing.T, dir string) Target {
	t.Helper()
	exePath := filepath.Join(dir, CmdName)
	if runtime.GOOS == xos.MacOS {
		exePath = filepath.Join(dir, BundleName, "Contents", "MacOS", CmdName)
		write(t, filepath.Join(dir, BundleName, "Contents", "Info.plist"), "<plist/>")
		write(t, filepath.Join(dir, BundleName, "Contents", "Resources", "app.icns"), "icon data")
	}
	write(t, exePath, "the application")
	if err := os.Chmod(exePath, executableModePerm); err != nil {
		t.Fatal(err)
	}
	target, err := ResolveTarget(exePath, runtime.GOOS)
	if err != nil {
		t.Fatal(err)
	}
	return target
}

// TestStageHelperProducesAnExecutableCopy covers what every platform needs from the helper, whatever shape the copy
// takes: the path handed back is a copy of the running program, it lives in the staging directory rather than in the
// installation about to be replaced, and it can be executed.
func TestStageHelperProducesAnExecutableCopy(t *testing.T) {
	c := check.New(t)
	target := installed(t, t.TempDir())
	workDir := t.TempDir()

	helper, err := stageHelper(t.Context(), &target, workDir)
	c.NoError(err)

	c.True(strings.HasPrefix(helper, workDir+string(filepath.Separator)),
		"the helper must be staged rather than run from the installation being replaced")
	fi, err := os.Stat(helper)
	c.NoError(err)
	c.True(fi.Mode().IsRegular())
	c.Equal("the application", read(t, helper))
	if runtime.GOOS != xos.WindowsOS {
		c.True(fi.Mode().Perm()&0o100 != 0, "the helper must be executable")
	}
}

// TestStageHelperCanBeStarted is a regression test for an update that could be prepared but never applied.
//
// The helper was staged as a bare "helper", with no extension. On Windows that copy cannot be started at all: the
// program name is resolved against PATHEXT even when it is an absolute path, so os/exec rejected it with "executable
// file not found in %PATH%" and every Windows update failed at the final step, after the user had already waited
// through the whole download.
//
// exec.LookPath is what this checks against because it applies exactly the rule Start does, and it does not require a
// real executable image -- so the stub the fixture writes is enough to catch a name that could never be run.
func TestStageHelperCanBeStarted(t *testing.T) {
	c := check.New(t)
	target := installed(t, t.TempDir())

	helper, err := stageHelper(t.Context(), &target, t.TempDir())
	c.NoError(err)

	_, err = exec.LookPath(helper)
	c.NoError(err, "the staged helper must be startable by the path stageHelper hands back")
}

// TestStageHelperLeavesTheInstallationAlone confirms the property the whole design rests on at this point: preparing an
// update touches nothing outside the staging directory, so an abort here can cost the user nothing.
func TestStageHelperLeavesTheInstallationAlone(t *testing.T) {
	c := check.New(t)
	installDir := t.TempDir()
	target := installed(t, installDir)
	before := listing(t, installDir)

	_, err := stageHelper(t.Context(), &target, t.TempDir())
	c.NoError(err)

	c.Equal(before, listing(t, installDir))
}

// listing returns every path under dir, relative to it, so that two moments can be compared.
func listing(t *testing.T, dir string) []string {
	t.Helper()
	var paths []string
	if err := filepath.WalkDir(dir, func(path string, _ os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			return relErr
		}
		paths = append(paths, rel)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return paths
}
