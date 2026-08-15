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
	"strconv"
	"strings"
	"time"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xos"
)

// Kind distinguishes what an update replaces.
type Kind byte

const (
	// KindExecutable means the target is the executable file itself, as on Linux and Windows.
	KindExecutable Kind = iota
	// KindBundle means the target is a macOS application bundle directory.
	KindBundle
)

// Target describes what an update replaces and what to run afterwards.
type Target struct {
	// Path is what gets swapped: the .app directory on macOS, the executable elsewhere.
	Path string
	// Parent is the directory holding Path. Staging happens here, so that the swap is a rename within a single
	// directory on a single filesystem.
	Parent string
	// Exec is the executable to run, which is inside Path for a bundle and equal to it otherwise.
	Exec string
	Kind Kind
}

// ResolveTarget determines what an update running from exePath would replace.
//
// goos is a parameter rather than being read from runtime.GOOS so that the rules for every platform can be exercised
// from any host; callers outside of tests should use CurrentTarget.
//
// exePath is expected to already be absolute and symlink-free. Note that this reports what *would* be replaced; whether
// replacing it is actually allowed is Preflight's job.
func ResolveTarget(exePath, goos string) (Target, error) {
	if exePath == "" {
		return Target{}, errs.New("an executable path is required")
	}
	if goos != xos.MacOS {
		return Target{Path: exePath, Parent: filepath.Dir(exePath), Exec: exePath, Kind: KindExecutable}, nil
	}
	// On macOS the executable normally lives at <name>.app/Contents/MacOS/<name>, and it is the bundle as a whole that
	// has to be replaced -- swapping just the executable would leave the bundle's signature broken. Walk up looking for
	// the enclosing bundle rather than assuming a fixed depth.
	dir := filepath.Dir(exePath)
	for {
		if strings.HasSuffix(dir, ".app") {
			return Target{Path: dir, Parent: filepath.Dir(dir), Exec: exePath, Kind: KindBundle}, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached the filesystem root without finding a bundle: a bare executable, not an installed application.
			return Target{Path: exePath, Parent: filepath.Dir(exePath), Exec: exePath, Kind: KindExecutable}, nil
		}
		dir = parent
	}
}

// CurrentTarget determines what an update to the running application would replace. Symlinks are resolved first, so
// that a symlink on the user's path is left alone and the real installation is what gets replaced.
func CurrentTarget() (Target, error) {
	exePath, err := os.Executable()
	if err != nil {
		return Target{}, errs.Wrap(err)
	}
	if exePath, err = filepath.EvalSymlinks(exePath); err != nil {
		return Target{}, errs.Wrap(err)
	}
	if exePath, err = filepath.Abs(exePath); err != nil {
		return Target{}, errs.Wrap(err)
	}
	return ResolveTarget(exePath, runtime.GOOS)
}

// BackupPath returns the path to move the existing installation aside to. The unique suffix means a leftover from an
// earlier attempt can never collide, and it is what lets the startup sweep recognize these.
//
// Two properties of the name matter on macOS. It no longer ends in ".app", so a bundle moved aside is not a second live
// Launch Services registration for com.trollworks.gcs that document associations could resolve to; and it starts with a
// dot, so the Finder does not show the user a mysterious extra item mid-update.
func (t *Target) BackupPath(unique string) string {
	return filepath.Join(t.Parent, "."+filepath.Base(t.Path)+backupSuffix+unique)
}

// FreeBackupPath returns a backup path that nothing currently occupies.
//
// The timestamp alone makes a collision all but impossible, but "all but" is not good enough here: the swap moves the
// installation to this path, and on most systems a rename over an existing file destroys it silently. If that file were
// a backup from an earlier update that had not yet been cleaned up, the user's only other copy of the application would
// disappear without a word.
func (t *Target) FreeBackupPath() string {
	base := strconv.FormatInt(time.Now().UnixNano(), 36)
	path := t.BackupPath(base)
	for i := 1; ; i++ {
		if _, err := os.Lstat(path); err != nil {
			return path
		}
		path = t.BackupPath(base + "-" + strconv.Itoa(i))
	}
}

// PayloadPath returns where the replacement lives inside a staging directory.
func (t *Target) PayloadPath(workDir string) string {
	return filepath.Join(workDir, filepath.Base(t.Path))
}

// ExecWithin returns the executable to run for an installation rooted at path, which may be a staging copy rather than
// the installed one. This is what makes it possible to smoke-test the replacement before committing to it.
func (t *Target) ExecWithin(path string) string {
	if t.Kind == KindBundle {
		rel, err := filepath.Rel(t.Path, t.Exec)
		if err != nil {
			return filepath.Join(path, "Contents", "MacOS", CmdName)
		}
		return filepath.Join(path, rel)
	}
	return path
}
