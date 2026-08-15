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
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// swapFixture lays out an installation, a staging directory holding the replacement, and the backup path the swap will
// use, all within one temporary directory so that the renames behave as they do in practice.
type swapFixture struct {
	target  string
	payload string
	backup  string
	workDir string
}

func newSwapFixture(t *testing.T, bundle bool) *swapFixture {
	t.Helper()
	dir := t.TempDir()
	f := &swapFixture{
		target:  filepath.Join(dir, CmdName),
		workDir: filepath.Join(dir, workDirPrefix+"abc"),
		backup:  filepath.Join(dir, "."+CmdName+backupSuffix+"123"),
	}
	f.payload = filepath.Join(f.workDir, CmdName)
	if bundle {
		f.target = filepath.Join(dir, BundleName)
		f.payload = filepath.Join(f.workDir, BundleName)
		f.backup = filepath.Join(dir, "."+BundleName+backupSuffix+"123")
		mkBundle(t, f.target, "installed")
		mkBundle(t, f.payload, "replacement")
	} else {
		write(t, f.target, "installed")
		write(t, f.payload, "replacement")
	}
	return f
}

// mkBundle builds a directory shaped like an application bundle, so the swap is exercised against a tree rather than a
// single file -- the case an ordinary rename cannot handle.
func mkBundle(t *testing.T, path, marker string) {
	t.Helper()
	write(t, filepath.Join(path, "Contents", "MacOS", CmdName), marker)
	write(t, filepath.Join(path, "Contents", "Info.plist"), "<plist/>")
}

// bundleMarker reads back what mkBundle wrote, to tell the two copies apart after a swap.
func bundleMarker(t *testing.T, path string) string {
	t.Helper()
	return read(t, filepath.Join(path, "Contents", "MacOS", CmdName))
}

// TestSwapReplacesAFile verifies the ordinary Linux and Windows case: the installed executable becomes the downloaded
// one, and the previous version is left where the record says it will be.
func TestSwapReplacesAFile(t *testing.T) {
	c := check.New(t)
	f := newSwapFixture(t, false)

	c.NoError(swap(f.target, f.payload, f.backup))

	c.Equal("replacement", read(t, f.target))
	c.Equal("installed", read(t, f.backup), "the previous version must be kept until the new one has started")
}

// TestSwapReplacesABundle verifies the macOS case, where what is replaced is a directory tree. A plain rename cannot
// replace a non-empty directory, so this is the case that would fail if the swap were written naively.
func TestSwapReplacesABundle(t *testing.T) {
	c := check.New(t)
	f := newSwapFixture(t, true)

	c.NoError(swap(f.target, f.payload, f.backup))

	c.Equal("replacement", bundleMarker(t, f.target))
	c.Equal("installed", bundleMarker(t, f.backup))
	c.True(exists(filepath.Join(f.target, "Contents", "Info.plist")), "the whole bundle must move, not just its parts")
}

// TestSwapReplacesABundleWithoutTheAtomicExchange verifies the same result on a filesystem that does not implement the
// atomic exchange, where the swap has to fall back to moving the installation aside first. This is the only path on
// Linux and Windows, and a real possibility on macOS.
func TestSwapReplacesABundleWithoutTheAtomicExchange(t *testing.T) {
	c := check.New(t)
	useRenameFallback(t)
	f := newSwapFixture(t, true)

	c.NoError(swap(f.target, f.payload, f.backup))

	c.Equal("replacement", bundleMarker(t, f.target))
	c.Equal("installed", bundleMarker(t, f.backup))
}

// useRenameFallback forces the two-rename form, which is what runs on Linux and Windows always, and on macOS whenever
// the filesystem does not implement the atomic exchange. Without this the macOS run would take the exchange and the
// rollback below would never be exercised anywhere.
func useRenameFallback(t *testing.T) {
	t.Helper()
	realExchange := exchangeFunc
	exchangeFunc = func(_, _, _ string) (bool, error) { return false, nil }
	t.Cleanup(func() { exchangeFunc = realExchange })
}

// TestSwapRestoresTheInstallationWhenTheSecondRenameFails is the case that decides whether a failed update leaves the
// user with a working application or with nothing at all. It is unreachable without forcing the failure, and it is
// exactly the situation where being wrong is unrecoverable.
func TestSwapRestoresTheInstallationWhenTheSecondRenameFails(t *testing.T) {
	c := check.New(t)
	useRenameFallback(t)
	f := newSwapFixture(t, false)

	// Let the installation move aside, then refuse to put the replacement in its place.
	realRename := renameFunc
	renameFunc = func(from, to string) error {
		if from == f.payload {
			return errors.New("refused")
		}
		return realRename(from, to)
	}
	t.Cleanup(func() { renameFunc = realRename })

	err := swap(f.target, f.payload, f.backup)

	c.HasError(err)
	c.True(exists(f.target), "the installation must be put back when the update cannot be completed")
	c.Equal("installed", read(t, f.target), "the restored installation must be the original one")
	c.Equal("replacement", read(t, f.payload), "the staged copy is left for the sweep to clear")
}

// TestSwapLeavesEverythingAloneWhenTheFirstRenameFails verifies that a swap which cannot even begin changes nothing.
// This is what happens when the installation directory turns out not to be writable after all.
func TestSwapLeavesEverythingAloneWhenTheFirstRenameFails(t *testing.T) {
	c := check.New(t)
	useRenameFallback(t)
	f := newSwapFixture(t, false)

	realRename := renameFunc
	renameFunc = func(from, to string) error {
		if from == f.target {
			return errors.New("refused")
		}
		return realRename(from, to)
	}
	t.Cleanup(func() { renameFunc = realRename })

	err := swap(f.target, f.payload, f.backup)

	c.HasError(err)
	c.Equal("installed", read(t, f.target))
	c.Equal("replacement", read(t, f.payload))
	c.False(exists(f.backup), "nothing moved, so nothing should have been backed up")
}

// TestSwapRefusesBadArguments verifies the checks made before anything moves. Each of these would otherwise turn into a
// partly-completed swap, which is far harder to recover from than a refusal.
func TestSwapRefusesBadArguments(t *testing.T) {
	c := check.New(t)

	f := newSwapFixture(t, false)
	c.HasError(swap("", f.payload, f.backup), "an empty target must be refused")
	c.HasError(swap(f.target, "", f.backup), "an empty payload must be refused")
	c.HasError(swap(f.target, f.payload, ""), "an empty backup must be refused")

	// A backup in another directory could be on another filesystem, where the move is a copy rather than a rename and
	// cannot simply be undone.
	c.HasError(swap(f.target, f.payload, filepath.Join(t.TempDir(), "elsewhere")))
	c.Equal("installed", read(t, f.target), "a refused swap must change nothing")

	f2 := newSwapFixture(t, false)
	c.HasError(swap(f2.target, filepath.Join(f2.workDir, "missing"), f2.backup), "a missing payload must be refused")

	f3 := newSwapFixture(t, false)
	c.NoError(os.Remove(f3.target))
	c.HasError(swap(f3.target, f3.payload, f3.backup), "a missing installation must be refused")
}

// TestSwapRefusesToOverwriteAnExistingBackup verifies that the installation is never renamed over something already
// sitting at the backup path. On most systems that rename would destroy the existing file silently, and if it happened
// to be a backup from an earlier update, the user's only other copy of the application would go with it.
func TestSwapRefusesToOverwriteAnExistingBackup(t *testing.T) {
	c := check.New(t)
	f := newSwapFixture(t, false)
	write(t, f.backup, "an earlier version that must not be destroyed")

	c.HasError(swap(f.target, f.payload, f.backup))
	c.Equal("an earlier version that must not be destroyed", read(t, f.backup))
	c.Equal("installed", read(t, f.target), "a refused swap must change nothing")
}

// TestFreeBackupPathAvoidsWhatIsAlreadyThere verifies that the name chosen for the displaced installation is one
// nothing occupies, which is what keeps the refusal above from being reachable in practice.
func TestFreeBackupPathAvoidsWhatIsAlreadyThere(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	target := Target{Path: filepath.Join(dir, CmdName), Parent: dir, Kind: KindExecutable}

	first := target.FreeBackupPath()
	write(t, first, "taken")
	second := target.FreeBackupPath()

	c.NotEqual(first, second, "a path already in use must not be chosen again")
	c.False(exists(second))
	c.Equal(dir, filepath.Dir(second), "the backup must still be a sibling of what it replaces")
}

// TestSwapRefusesAMismatchedPayload verifies that a bundle cannot replace a bare executable, or the reverse. Reaching
// that point would mean something upstream resolved the wrong thing, and completing the swap would install something
// the system cannot run.
func TestSwapRefusesAMismatchedPayload(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	target := filepath.Join(dir, CmdName)
	write(t, target, "installed")
	payload := filepath.Join(dir, workDirPrefix+"abc", BundleName)
	mkBundle(t, payload, "replacement")

	c.HasError(swap(target, payload, filepath.Join(dir, "."+CmdName+backupSuffix+"1")))
	c.Equal("installed", read(t, target))
}
