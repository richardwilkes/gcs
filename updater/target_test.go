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
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// hostPath rewrites a path written with forward slashes into the host's own separator form.
//
// ResolveTarget takes goos as a parameter so the rules for every platform can be exercised from any host, but the path
// handling underneath it is path/filepath, which is compiled for the host rather than for the goos being exercised. The
// rules are what these tests are about, so separators are normalized on both sides of every comparison; otherwise the
// macOS cases would fail on a Windows host purely over backslashes.
func hostPath(path string) string {
	return filepath.FromSlash(path)
}

// An update running from inside an application bundle must replace the bundle rather than the executable within it.
// Swapping just the executable would leave a bundle whose code signature no longer matches its contents, which macOS
// kills on sight.
func TestResolveTargetOnMacOSFindsTheBundle(t *testing.T) {
	c := check.New(t)
	target, err := ResolveTarget(hostPath("/Applications/GCS.app/Contents/MacOS/gcs"), "darwin")
	c.NoError(err)
	c.Equal(KindBundle, target.Kind)
	c.Equal(hostPath("/Applications/GCS.app"), target.Path)
	c.Equal(hostPath("/Applications"), target.Parent)
	c.Equal(hostPath("/Applications/GCS.app/Contents/MacOS/gcs"), target.Exec)
}

// A bare executable -- a development build run straight from the build directory -- resolves to the executable itself.
// Preflight is what refuses to update it; resolution just reports what it sees.
func TestResolveTargetOnMacOSWithoutABundle(t *testing.T) {
	c := check.New(t)
	target, err := ResolveTarget(hostPath("/Users/someone/code/gcs/gcs"), "darwin")
	c.NoError(err)
	c.Equal(KindExecutable, target.Kind)
	c.Equal(hostPath("/Users/someone/code/gcs/gcs"), target.Path)
	c.Equal(hostPath("/Users/someone/code/gcs"), target.Parent)
}

// The bundle search walks up rather than assuming a fixed depth, so something nested deeper inside a bundle still
// resolves to the nearest enclosing bundle.
func TestResolveTargetOnMacOSHandlesNesting(t *testing.T) {
	c := check.New(t)
	target, err := ResolveTarget(hostPath("/Applications/GCS.app/Contents/Frameworks/Inner.app/Contents/MacOS/tool"),
		"darwin")
	c.NoError(err)
	c.Equal(KindBundle, target.Kind)
	c.Equal(hostPath("/Applications/GCS.app/Contents/Frameworks/Inner.app"), target.Path,
		"the nearest enclosing bundle is the one found")
}

// Linux and Windows replace the single executable, which is all their archives contain.
func TestResolveTargetElsewhereIsTheExecutable(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		exePath string
		goos    string
		parent  string
	}{
		{hostPath("/home/someone/bin/gcs"), "linux", hostPath("/home/someone/bin")},
		{hostPath("/usr/local/bin/gcs"), "linux", hostPath("/usr/local/bin")},
		{`C:\Apps\GCS\gcs.exe`, "windows", `C:\Apps\GCS`},
	} {
		target, err := ResolveTarget(one.exePath, one.goos)
		c.NoError(err, one.exePath)
		c.Equal(KindExecutable, target.Kind, one.exePath)
		c.Equal(one.exePath, target.Path, one.exePath)
		c.Equal(one.exePath, target.Exec, one.exePath)
		// The Windows path is the one case hostPath cannot normalize, since a host using forward slashes does not see
		// its backslashes as separators at all. Its parent is only meaningful on a Windows host.
		if one.goos != "windows" || filepath.Separator == '\\' {
			c.Equal(one.parent, target.Parent, one.exePath)
		}
	}
}

// A failed os.Executable must not produce a target rooted at the current directory.
func TestResolveTargetRejectsAnEmptyPath(t *testing.T) {
	c := check.New(t)
	_, err := ResolveTarget("", "linux")
	c.HasError(err)
	_, err = ResolveTarget("", "darwin")
	c.HasError(err)
}

// The property that keeps a displaced macOS installation from confusing Launch Services: the name it is moved aside to
// must no longer end in ".app", or /Applications briefly holds two complete bundles both claiming com.trollworks.gcs.
func TestBackupPathIsNotABundle(t *testing.T) {
	c := check.New(t)
	target, err := ResolveTarget(hostPath("/Applications/GCS.app/Contents/MacOS/gcs"), "darwin")
	c.NoError(err)
	backup := target.BackupPath("123")
	c.False(strings.HasSuffix(backup, ".app"), "got %s", backup)
	c.Equal(hostPath("/Applications"), filepath.Dir(backup), "the backup must be a sibling, so the move is a rename")
	c.True(strings.HasPrefix(filepath.Base(backup), "."), "the backup should be hidden; got %s", backup)
	c.NotEqual(backup, target.BackupPath("456"), "the unique suffix must actually vary")
}

// The same sibling property for the executable platforms. A backup landing anywhere else would make the swap a
// cross-device copy rather than a rename.
func TestBackupPathStaysBesideTheTarget(t *testing.T) {
	c := check.New(t)
	target, err := ResolveTarget(hostPath("/home/someone/bin/gcs"), "linux")
	c.NoError(err)
	backup := target.BackupPath("123")
	c.Equal(hostPath("/home/someone/bin"), filepath.Dir(backup))
	c.NotEqual(target.Path, backup)
}

// Locating the executable inside a staged copy is what makes it possible to run the replacement before committing to
// it.
func TestExecWithin(t *testing.T) {
	c := check.New(t)

	bundle, err := ResolveTarget(hostPath("/Applications/GCS.app/Contents/MacOS/gcs"), "darwin")
	c.NoError(err)
	c.Equal(hostPath("/tmp/staged/GCS.app/Contents/MacOS/gcs"), bundle.ExecWithin(hostPath("/tmp/staged/GCS.app")))

	exe, err := ResolveTarget(hostPath("/home/someone/bin/gcs"), "linux")
	c.NoError(err)
	c.Equal(hostPath("/tmp/staged/gcs"), exe.ExecWithin(hostPath("/tmp/staged/gcs")))
}

// The staged replacement is named the same as what it replaces, so the swap is a rename between two names in one
// directory.
func TestPayloadPath(t *testing.T) {
	c := check.New(t)
	bundle, err := ResolveTarget(hostPath("/Applications/GCS.app/Contents/MacOS/gcs"), "darwin")
	c.NoError(err)
	workDir := hostPath("/Applications/" + workDirPrefix + "x")
	c.Equal(hostPath("/Applications/"+workDirPrefix+"x/GCS.app"), bundle.PayloadPath(workDir))
}

// Every leftover an update can create must be matched by the patterns the startup sweep uses. A name that no pattern
// matches is one that stays in the user's Applications folder forever.
func TestSweepGlobsMatchWhatIsProduced(t *testing.T) {
	c := check.New(t)
	for _, exePath := range []string{
		hostPath("/Applications/GCS.app/Contents/MacOS/gcs"),
		hostPath("/home/someone/bin/gcs"),
	} {
		goos := "darwin"
		if !strings.Contains(exePath, ".app") {
			goos = "linux"
		}
		target, err := ResolveTarget(exePath, goos)
		c.NoError(err, exePath)
		for _, name := range []string{
			filepath.Base(target.BackupPath("1700000000000000000")),
			workDirPrefix + "987654321",
		} {
			matched := false
			for _, glob := range sweepGlobs {
				if ok, matchErr := filepath.Match(glob, name); matchErr == nil && ok {
					matched = true
					break
				}
			}
			c.True(matched, "no sweep pattern matches %s", name)
		}
	}
}
