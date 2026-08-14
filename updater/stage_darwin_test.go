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
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// makeDMG builds a real disk image holding a directory shaped like an application bundle, so that the mount, copy and
// unmount sequence is exercised against the tools it actually uses rather than against a stand-in.
//
// The image is given the same style of volume name the packager produces, since that name is exactly what the mount
// deliberately does not rely on.
func makeDMG(t *testing.T, volumeName string) string {
	t.Helper()
	for _, tool := range []string{"/usr/bin/hdiutil", "/usr/bin/ditto"} {
		if _, err := os.Stat(tool); err != nil {
			t.Skipf("%s is needed for this test", tool)
		}
	}
	src := filepath.Join(t.TempDir(), "src")
	write(t, filepath.Join(src, BundleName, "Contents", "MacOS", CmdName), "the application")
	write(t, filepath.Join(src, BundleName, "Contents", "Info.plist"), "<plist/>")
	write(t, filepath.Join(src, BundleName, "Contents", "Resources", "app.icns"), "icon data")

	dmg := filepath.Join(t.TempDir(), "test.dmg")
	cmd := exec.Command("/usr/bin/hdiutil", "create", "-quiet", "-srcfolder", src, "-volname", volumeName,
		"-format", "UDZO", "-ov", dmg)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("unable to build a disk image for this test: %v\n%s", err, out)
	}
	return dmg
}

// TestStagePayloadFromADiskImage exercises the macOS unpacking path against a real disk image: mounting it, copying the
// bundle out, and unmounting again. None of that can be exercised any other way, and getting the hdiutil invocation
// wrong would only show up when a user tried to update.
func TestStagePayloadFromADiskImage(t *testing.T) {
	if testing.Short() {
		t.Skip("mounting a disk image is slow")
	}
	c := check.New(t)
	dmg := makeDMG(t, "GCS v5.46.0")
	dst := filepath.Join(t.TempDir(), BundleName)

	c.NoError(stagePayload(t.Context(), dmg, dst))

	c.Equal("the application", read(t, filepath.Join(dst, "Contents", "MacOS", CmdName)),
		"the whole bundle must be copied out, not just its top level")
	c.Equal("icon data", read(t, filepath.Join(dst, "Contents", "Resources", "app.icns")))

	fi, err := os.Stat(filepath.Join(dst, "Contents", "MacOS", CmdName))
	c.NoError(err)
	c.True(fi.Mode().IsRegular())

	// The image must not be left mounted. A stray mount would accumulate a volume per update attempt, and would keep
	// the downloaded file alive after the directory holding it was removed.
	out, err := exec.Command("/usr/bin/hdiutil", "info").CombinedOutput()
	c.NoError(err)
	c.NotContains(string(out), dmg, "the disk image was left mounted")
}

// TestStagePayloadWithTheVolumeNameAlreadyInUse is the case the explicit mount point exists for. The packager names
// every volume "GCS v<version>", so a user who already has that disk image open would push a second mount to a
// different, unpredictable path -- and anything that guessed the path from the volume name would copy from the wrong
// place, or fail outright.
func TestStagePayloadWithTheVolumeNameAlreadyInUse(t *testing.T) {
	if testing.Short() {
		t.Skip("mounting a disk image is slow")
	}
	c := check.New(t)
	const volumeName = "GCS v5.46.0"

	// Mount one image under the contested name and keep it there for the duration.
	occupying := makeDMG(t, volumeName)
	occupyingMount := filepath.Join(t.TempDir(), "occupying")
	c.NoError(os.MkdirAll(occupyingMount, 0o755))
	c.NoError(attachDMG(t.Context(), occupying, occupyingMount))
	defer detachDMG(occupyingMount)

	// Unpacking a second image with the same volume name must still work, and must produce that image's contents.
	dmg := makeDMG(t, volumeName)
	dst := filepath.Join(t.TempDir(), BundleName)
	c.NoError(stagePayload(t.Context(), dmg, dst))
	c.Equal("the application", read(t, filepath.Join(dst, "Contents", "MacOS", CmdName)))
}

// TestStagePayloadRejectsAnImageWithoutTheApplication verifies that a disk image which is readable but does not hold
// what was expected is refused, rather than producing an empty or partial installation.
func TestStagePayloadRejectsAnImageWithoutTheApplication(t *testing.T) {
	if testing.Short() {
		t.Skip("mounting a disk image is slow")
	}
	c := check.New(t)
	src := filepath.Join(t.TempDir(), "src")
	write(t, filepath.Join(src, "SomethingElse.app", "Contents", "MacOS", "other"), "not what was asked for")
	dmg := filepath.Join(t.TempDir(), "test.dmg")
	if out, err := exec.Command("/usr/bin/hdiutil", "create", "-quiet", "-srcfolder", src, "-volname", "Other",
		"-format", "UDZO", "-ov", dmg).CombinedOutput(); err != nil {
		t.Skipf("unable to build a disk image for this test: %v\n%s", err, out)
	}

	dst := filepath.Join(t.TempDir(), BundleName)
	c.HasError(stagePayload(t.Context(), dmg, dst))
	c.False(exists(dst))
}

// TestStagePayloadRejectsSomethingThatIsNotADiskImage verifies that a corrupted download is reported rather than
// producing a confusing failure further along.
func TestStagePayloadRejectsSomethingThatIsNotADiskImage(t *testing.T) {
	c := check.New(t)
	dmg := filepath.Join(t.TempDir(), "test.dmg")
	write(t, dmg, "this is not a disk image")
	c.HasError(stagePayload(t.Context(), dmg, filepath.Join(t.TempDir(), BundleName)))
}
