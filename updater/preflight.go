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
	"path/filepath"
	"runtime"
	"strings"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xos"
)

// Blocker identifies why this installation cannot update itself. These are checked before anything is downloaded, so
// the user is told that an automatic update is not possible without first waiting for thirty megabytes to arrive.
type Blocker string

const (
	// BlockerDevBuild means this is a development build, which has no release to update from.
	BlockerDevBuild Blocker = "dev-build"
	// BlockerRenamedExecutable means the executable is not named what the distribution names it, so replacing it would
	// not produce a working installation -- on Linux the desktop integration keys off the name.
	BlockerRenamedExecutable Blocker = "renamed-executable"
	// BlockerNotABundle means macOS is running a bare executable rather than an installed application bundle.
	BlockerNotABundle Blocker = "not-a-bundle"
	// BlockerTranslocated means macOS is running a read-only copy made by App Translocation, typically because the
	// application was launched straight from the disk image.
	BlockerTranslocated Blocker = "translocated"
	// BlockerPackageManaged means the installation belongs to a package manager, which should be the one to update it.
	BlockerPackageManaged Blocker = "package-managed"
	// BlockerHomebrew means the installation came from the Homebrew cask.
	BlockerHomebrew Blocker = "homebrew"
	// BlockerReadOnly means the installation cannot be written to. This is the ordinary outcome for an application in
	// a location that needs administrator rights.
	BlockerReadOnly Blocker = "read-only"
	// BlockerNoAsset means the release has no build for this operating system and processor.
	BlockerNoAsset Blocker = "no-asset"
	// BlockerNoDigest means the release's build for this platform carries no checksum, so it cannot be verified.
	BlockerNoDigest Blocker = "no-digest"
)

// Unavailable reports that an automatic update is not possible, and why. Detail carries the technical specifics for the
// log; the Blocker is what the caller turns into something to show the user.
type Unavailable struct {
	Blocker Blocker
	Detail  string
}

func (u *Unavailable) Error() string {
	if u.Detail != "" {
		return string(u.Blocker) + ": " + u.Detail
	}
	return string(u.Blocker)
}

// Plan is a checked, ready-to-run update.
type Plan struct {
	Target      Target
	Asset       Asset
	FromVersion string
	ToVersion   string
}

// packageManagedPrefixes are locations owned by a system package manager. Replacing a file under one of these would
// leave the package database describing something that is no longer there, and the next package operation would either
// undo the update or refuse to proceed.
var packageManagedPrefixes = []string{
	"/usr/", "/opt/", "/snap/", "/nix/store/", "/var/lib/flatpak/", "/app/",
}

// sandboxEnvVars are set by the packaging systems that run an application from a managed, usually read-only, image.
var sandboxEnvVars = []string{"SNAP", "FLATPAK_ID", "APPIMAGE", "APPDIR", "container"}

// caskroomPaths are where Homebrew records a cask installation, on Apple silicon and Intel respectively.
var caskroomPaths = []string{"/opt/homebrew/Caskroom/" + CmdName, "/usr/local/Caskroom/" + CmdName}

// caskInstallPath is where the Homebrew cask puts the application. A cask moves the bundle into /Applications exactly
// as a person dragging it there would, so the installed copy carries no marking of its own -- the only evidence is the
// Caskroom entry alongside it.
const caskInstallPath = "/Applications/" + BundleName

// checkHomebrew refuses to replace an installation Homebrew is managing. Updating it behind Homebrew's back would leave
// `brew` still describing the old version, free to reinstall it over the new one on the next upgrade.
//
// Both conditions are required. The Caskroom existing only means the cask is installed somewhere; a second copy the
// user keeps elsewhere is theirs to update, and refusing that would be wrong.
func checkHomebrew(targetPath string) error {
	if targetPath != caskInstallPath {
		return nil
	}
	for _, caskroom := range caskroomPaths {
		if _, err := os.Stat(caskroom); err == nil {
			return &Unavailable{Blocker: BlockerHomebrew, Detail: caskroom + " manages " + targetPath}
		}
	}
	return nil
}

// Preflight decides whether the running installation can replace itself with the given release, and returns the plan to
// do it. Every check here is local and immediate: nothing is downloaded until this succeeds.
//
// The inputs are parameters rather than being read from globals so that each refusal can be reached from a test.
func Preflight(exePath, goos, goarch, fromVersion, toVersion string, assets []Asset, lookupEnv func(string) (string, bool)) (*Plan, error) {
	if IsDevVersion(fromVersion) {
		return nil, &Unavailable{Blocker: BlockerDevBuild, Detail: "version is " + fromVersion}
	}
	target, err := ResolveTarget(exePath, goos)
	if err != nil {
		return nil, err
	}
	if err = checkInstallation(&target, exePath, goos, lookupEnv); err != nil {
		return nil, err
	}
	asset, err := SelectAsset(toVersion, goos, goarch, assets)
	if err != nil {
		return nil, &Unavailable{Blocker: BlockerNoAsset, Detail: err.Error()}
	}
	if asset.SHA256 == "" {
		// GitHub only began publishing asset checksums in 2025, so this cannot happen for any release new enough to be
		// worth installing. It is refused rather than downgraded to an unverified install, because an update that
		// silently skips its integrity check is worse than no automatic update at all.
		return nil, &Unavailable{Blocker: BlockerNoDigest, Detail: asset.Name + " has no checksum"}
	}
	return &Plan{Target: target, Asset: asset, FromVersion: fromVersion, ToVersion: toVersion}, nil
}

// checkInstallation verifies that this installation is one that may replace itself.
func checkInstallation(target *Target, exePath, goos string, lookupEnv func(string) (string, bool)) error {
	wantExe := CmdName
	if goos == xos.WindowsOS {
		wantExe += ".exe"
	}
	if !strings.EqualFold(filepath.Base(exePath), wantExe) {
		return &Unavailable{Blocker: BlockerRenamedExecutable, Detail: filepath.Base(exePath) + " is not " + wantExe}
	}
	if goos == xos.MacOS {
		// App Translocation runs the application from a randomized read-only copy, which happens whenever a quarantined
		// application is launched without first being moved out of the disk image it arrived in. The copy's containing
		// directory can be writable, so the write probe below would happily "update" something that is thrown away the
		// moment the application quits.
		if strings.Contains(exePath, "/AppTranslocation/") {
			return &Unavailable{Blocker: BlockerTranslocated, Detail: exePath}
		}
		if target.Kind != KindBundle {
			return &Unavailable{Blocker: BlockerNotABundle, Detail: exePath}
		}
		if err := checkHomebrew(target.Path); err != nil {
			return err
		}
	}
	if goos == xos.LinuxOS {
		for _, name := range sandboxEnvVars {
			if value, ok := lookupEnv(name); ok && value != "" {
				return &Unavailable{Blocker: BlockerPackageManaged, Detail: name + " is set"}
			}
		}
	}
	if goos != xos.WindowsOS {
		normalized := filepath.ToSlash(target.Path)
		for _, prefix := range packageManagedPrefixes {
			if strings.HasPrefix(normalized, prefix) {
				return &Unavailable{Blocker: BlockerPackageManaged, Detail: target.Path + " is under " + prefix}
			}
		}
	}
	if err := probeWritable(target.Parent); err != nil {
		return &Unavailable{Blocker: BlockerReadOnly, Detail: err.Error()}
	}
	return nil
}

// probeWritable verifies that the directory can actually be written to, by writing to it. Checking permission bits is
// not equivalent: it does not account for access control lists, for a read-only mount -- which is what running straight
// from a disk image looks like -- or for Windows, where the bits carry almost no information at all.
func probeWritable(dir string) error {
	probe, err := os.MkdirTemp(dir, workDirPrefix+"probe-")
	if err != nil {
		return errs.Wrap(err)
	}
	if err = os.Remove(probe); err != nil {
		return errs.Wrap(err)
	}
	return nil
}

// CurrentPreflight runs Preflight for the running application against the given release.
func CurrentPreflight(toVersion string, assets []Asset) (*Plan, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, errs.Wrap(err)
	}
	if exePath, err = filepath.EvalSymlinks(exePath); err != nil {
		return nil, errs.Wrap(err)
	}
	if exePath, err = filepath.Abs(exePath); err != nil {
		return nil, errs.Wrap(err)
	}
	return Preflight(exePath, runtime.GOOS, UpdateArch(assets, toVersion), xos.AppVersion, toVersion, assets,
		os.LookupEnv)
}
