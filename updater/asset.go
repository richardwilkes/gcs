// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

// Package updater applies an application update in place: it downloads the release built for the running platform,
// verifies it, swaps it into the installation, and relaunches.
//
// Nothing here depends on the UI, so that it can be exercised headlessly and so that the helper process -- which
// finishes the update after the application has exited -- can run it without starting a window.
//
// User-visible text belongs in the ux package. Failures are reported as a stable Reason code the caller maps to a
// localized message, because that reason is written to a file read after the swap by a *different* build of GCS,
// possibly running in a different language.
package updater

import (
	"runtime"
	"strings"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xos"
)

const (
	// AppName is the application's name. It and the two names below are literals rather than the mutable xos globals,
	// which are set at startup: they have to match what the packager baked into the published files, and the helper
	// process has to resolve the same paths and asset names as the application that staged the update.
	AppName = "GCS"
	// CmdName is the executable name, and the base of every release asset name.
	CmdName = "gcs"
	// BundleName is the macOS application bundle's name, matching finder_app_name in packaging.yml.
	BundleName = "GCS.app"
)

// Asset is a downloadable file attached to a release.
type Asset struct {
	Name   string
	URL    string
	SHA256 string
	Size   int64
}

// platform describes the names and formats the packager produces for a GOOS.
type platform struct {
	label   string
	ext     string
	payload string
}

// platforms mirrors the naming in unison's cmd/upack/packager, whose packager_darwin.go, packager_linux.go and
// packager_windows.go each build "<exe>-<version>-<label><ext>". It is written out rather than derived from
// xos.AppCmdName so a change to the application's command name cannot silently change which file gets downloaded.
var platforms = map[string]platform{
	xos.MacOS:     {label: "macos", ext: ".dmg", payload: BundleName},
	xos.LinuxOS:   {label: "linux", ext: ".tgz", payload: CmdName},
	xos.WindowsOS: {label: "windows", ext: ".zip", payload: CmdName + ".exe"},
}

// AssetName returns the name of the release asset built for the given platform. version must be the raw release
// version, e.g. "5.46.0" -- ux.filterVersion trims a trailing ".0" for display, which would not match.
func AssetName(version, goos, goarch string) (string, error) {
	p, ok := platforms[goos]
	if !ok {
		return "", errs.Newf("no %s distribution is built for %s", AppName, goos)
	}
	if goarch != "amd64" && goarch != "arm64" {
		return "", errs.Newf("no %s distribution is built for %s/%s", AppName, goos, goarch)
	}
	if version == "" {
		return "", errs.New("a version is required")
	}
	return CmdName + "-" + version + "-" + p.label + "-" + goarch + p.ext, nil
}

// PayloadName returns the name of the single item the platform's archive holds: the application bundle on macOS, the
// executable elsewhere.
func PayloadName(goos string) (string, error) {
	p, ok := platforms[goos]
	if !ok {
		return "", errs.Newf("no %s distribution is built for %s", AppName, goos)
	}
	return p.payload, nil
}

// SelectAsset returns the asset holding the distribution for the given platform. Matching is by exact (though
// case-insensitive) name rather than by pattern, so a release missing the build for this platform is reported as
// missing instead of resolving to some other platform's file.
func SelectAsset(version, goos, goarch string, assets []Asset) (Asset, error) {
	name, err := AssetName(version, goos, goarch)
	if err != nil {
		return Asset{}, err
	}
	for _, one := range assets {
		if strings.EqualFold(one.Name, name) {
			return one, nil
		}
	}
	return Asset{}, errs.Newf("release %s has no %s", version, name)
}

// UpdateArch returns the architecture to update to. It is normally the one this build was compiled for, but an amd64
// build running under Rosetta on Apple silicon returns arm64: leaving it as amd64 would keep such a user on the
// translated build forever.
func UpdateArch(assets []Asset, version string) string {
	arch := runtime.GOARCH
	if runtime.GOOS != xos.MacOS || arch != "amd64" || !runningTranslated() {
		return arch
	}
	// Prefer the native build only if it was published; otherwise stay on the one known to exist.
	if _, err := SelectAsset(version, xos.MacOS, "arm64", assets); err != nil {
		return arch
	}
	return "arm64"
}
