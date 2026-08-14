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
	"runtime"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xos"
)

// noEnv stands in for a process with none of the sandbox markers set.
func noEnv(string) (string, bool) { return "", false }

// linuxAssets is a release publishing every build, so that the platform checks are what a test exercises rather than a
// missing asset.
func linuxAssets(version string) []Asset {
	var assets []Asset
	for _, goos := range []string{xos.MacOS, xos.LinuxOS, xos.WindowsOS} {
		for _, goarch := range []string{"amd64", "arm64"} {
			name, err := AssetName(version, goos, goarch)
			if err != nil {
				continue
			}
			assets = append(assets, Asset{Name: name, URL: "https://example.com/" + name, SHA256: "abc", Size: 100})
		}
	}
	return assets
}

// blockerOf extracts the Blocker from a refusal, failing the test if the error is not one.
func blockerOf(t *testing.T, err error) Blocker {
	t.Helper()
	var unavailable *Unavailable
	if !errors.As(err, &unavailable) {
		t.Fatalf("expected an *Unavailable, got %v", err)
	}
	return unavailable.Blocker
}

// TestPreflightAcceptsAWritableInstallation verifies the case that must work: an ordinary installation in a directory
// the user can write to, with the matching build published.
func TestPreflightAcceptsAWritableInstallation(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	exePath := filepath.Join(dir, CmdName)
	write(t, exePath, "binary")

	plan, err := Preflight(exePath, xos.LinuxOS, "amd64", "5.45.2", "5.46.0", linuxAssets("5.46.0"), noEnv)
	c.NoError(err)
	c.NotNil(plan)
	c.Equal("5.46.0", plan.ToVersion)
	c.Equal("5.45.2", plan.FromVersion)
	c.Equal("gcs-5.46.0-linux-amd64.tgz", plan.Asset.Name)
	c.Equal(exePath, plan.Target.Path)
}

// TestPreflightRefusesADevBuild verifies that a build with no release behind it never tries to update. The stamped
// version is "0.0" for anything not built by the release workflow.
func TestPreflightRefusesADevBuild(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	exePath := filepath.Join(dir, CmdName)
	write(t, exePath, "binary")

	for _, version := range []string{"0.0", ""} {
		_, err := Preflight(exePath, xos.LinuxOS, "amd64", version, "5.46.0", linuxAssets("5.46.0"), noEnv)
		c.HasError(err, version)
		c.Equal(BlockerDevBuild, blockerOf(t, err), version)
	}
}

// TestPreflightRefusesARenamedExecutable verifies that a copy the user renamed is left alone. On Linux the desktop
// integration only installs itself when the executable is named "gcs", so replacing a differently-named file would
// produce an installation that half works.
func TestPreflightRefusesARenamedExecutable(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	exePath := filepath.Join(dir, "gcs-5.45.2")
	write(t, exePath, "binary")

	_, err := Preflight(exePath, xos.LinuxOS, "amd64", "5.45.2", "5.46.0", linuxAssets("5.46.0"), noEnv)
	c.HasError(err)
	c.Equal(BlockerRenamedExecutable, blockerOf(t, err))
}

// TestPreflightRefusesPackageManagedLocations verifies that an installation belonging to a package manager is left for
// that package manager to update. Replacing it would leave the package database describing a file that is no longer
// there.
func TestPreflightRefusesPackageManagedLocations(t *testing.T) {
	c := check.New(t)
	for _, dir := range []string{
		"/usr/bin", "/usr/local/bin", "/opt/gcs", "/snap/gcs/current/bin", "/nix/store/abc123-gcs/bin",
		"/var/lib/flatpak/app/com.trollworks.gcs/current", "/app/bin",
	} {
		_, err := Preflight(filepath.Join(dir, CmdName), xos.LinuxOS, "amd64", "5.45.2", "5.46.0",
			linuxAssets("5.46.0"), noEnv)
		c.HasError(err, dir)
		c.Equal(BlockerPackageManaged, blockerOf(t, err), dir)
	}
}

// TestPreflightRefusesSandboxedRuntimes verifies detection of the packaging systems that run the application from a
// managed image, which do not necessarily install under a recognizable path.
func TestPreflightRefusesSandboxedRuntimes(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	exePath := filepath.Join(dir, CmdName)
	write(t, exePath, "binary")

	for _, name := range sandboxEnvVars {
		env := func(lookup string) (string, bool) {
			if lookup == name {
				return "set", true
			}
			return "", false
		}
		_, err := Preflight(exePath, xos.LinuxOS, "amd64", "5.45.2", "5.46.0", linuxAssets("5.46.0"), env)
		c.HasError(err, name)
		c.Equal(BlockerPackageManaged, blockerOf(t, err), name)
	}
}

// TestPreflightRefusesAnUnwritableInstallation verifies the check that keeps an update from being downloaded only to
// discover at the last moment that it cannot be installed. The probe is a real write, since permission bits alone do
// not account for access control lists or a read-only mount.
func TestPreflightRefusesAnUnwritableInstallation(t *testing.T) {
	if runtime.GOOS == xos.WindowsOS {
		t.Skip("directory modes do not govern writability on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root can write to a read-only directory")
	}
	c := check.New(t)
	dir := filepath.Join(t.TempDir(), "readonly")
	c.NoError(os.Mkdir(dir, 0o755))
	exePath := filepath.Join(dir, CmdName)
	write(t, exePath, "binary")
	c.NoError(os.Chmod(dir, 0o500))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) }) //nolint:errcheck // Best-effort cleanup

	_, err := Preflight(exePath, xos.LinuxOS, "amd64", "5.45.2", "5.46.0", linuxAssets("5.46.0"), noEnv)
	c.HasError(err)
	c.Equal(BlockerReadOnly, blockerOf(t, err))
}

// TestPreflightRefusesATranslocatedApp verifies the App Translocation check. The randomized copy's parent directory can
// be writable, so the write probe alone would happily "update" a copy that vanishes when the application quits.
func TestPreflightRefusesATranslocatedApp(t *testing.T) {
	c := check.New(t)
	exePath := "/private/var/folders/xy/T/AppTranslocation/1234-ABCD/d/GCS.app/Contents/MacOS/gcs"
	_, err := Preflight(exePath, xos.MacOS, "arm64", "5.45.2", "5.46.0", linuxAssets("5.46.0"), noEnv)
	c.HasError(err)
	c.Equal(BlockerTranslocated, blockerOf(t, err))
}

// TestPreflightRefusesABareExecutableOnMacOS verifies that macOS only updates a real installed bundle. Replacing a bare
// executable would not produce something the Finder or Launch Services could run.
func TestPreflightRefusesABareExecutableOnMacOS(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	exePath := filepath.Join(dir, CmdName)
	write(t, exePath, "binary")

	_, err := Preflight(exePath, xos.MacOS, "arm64", "5.45.2", "5.46.0", linuxAssets("5.46.0"), noEnv)
	c.HasError(err)
	c.Equal(BlockerNotABundle, blockerOf(t, err))
}

// TestPreflightRefusesAMissingAsset verifies that a release which did not publish a build for this platform is reported
// as such, rather than some other platform's file being downloaded.
func TestPreflightRefusesAMissingAsset(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	exePath := filepath.Join(dir, CmdName)
	write(t, exePath, "binary")

	_, err := Preflight(exePath, xos.LinuxOS, "amd64", "5.45.2", "5.46.0", []Asset{
		{Name: "gcs-5.46.0-macos-arm64.dmg", SHA256: "abc"},
	}, noEnv)
	c.HasError(err)
	c.Equal(BlockerNoAsset, blockerOf(t, err))
}

// TestPreflightRefusesAnUnverifiableAsset verifies that a release whose asset carries no checksum is refused outright
// rather than installed without verification. GitHub has published checksums since 2025, so this only arises for
// releases far older than anything worth installing -- but silently skipping the check would be worse than not
// updating at all.
func TestPreflightRefusesAnUnverifiableAsset(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	exePath := filepath.Join(dir, CmdName)
	write(t, exePath, "binary")

	_, err := Preflight(exePath, xos.LinuxOS, "amd64", "5.45.2", "5.46.0", []Asset{
		{Name: "gcs-5.46.0-linux-amd64.tgz", URL: "https://example.com/x", Size: 100},
	}, noEnv)
	c.HasError(err)
	c.Equal(BlockerNoDigest, blockerOf(t, err))
}

// TestCheckHomebrewOnlyRefusesTheManagedCopy verifies both halves of the Homebrew rule. The cask moves the bundle into
// /Applications exactly as a person would, so the installed copy is indistinguishable on its own -- but a second copy
// the user keeps elsewhere is theirs, and refusing to update that would be wrong.
func TestCheckHomebrewOnlyRefusesTheManagedCopy(t *testing.T) {
	c := check.New(t)

	// Wherever the cask may or may not be installed, a copy outside /Applications is never the one it manages.
	c.NoError(checkHomebrew("/Users/someone/Applications/GCS.app"))
	c.NoError(checkHomebrew("/Volumes/Data/GCS.app"))

	// The managed location is only refused when Homebrew has actually recorded an installation there.
	err := checkHomebrew(caskInstallPath)
	installed := false
	for _, caskroom := range caskroomPaths {
		if _, statErr := os.Stat(caskroom); statErr == nil {
			installed = true
			break
		}
	}
	if installed {
		c.HasError(err, "a cask installation must be left to brew")
		c.Equal(BlockerHomebrew, blockerOf(t, err))
	} else {
		c.NoError(err, "without a Caskroom entry, /Applications is just where the user put it")
	}
}

// TestProbeWritableCleansUpAfterItself verifies the probe leaves nothing behind. It runs every time the update dialog
// is shown, so a probe that leaked a directory each time would slowly fill the user's installation directory.
func TestProbeWritableCleansUpAfterItself(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	for range 5 {
		c.NoError(probeWritable(dir))
	}
	entries, err := os.ReadDir(dir)
	c.NoError(err)
	c.Equal(0, len(entries), "the write probe left something behind")
}
