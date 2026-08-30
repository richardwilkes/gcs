// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestResolvePathRestoresCase verifies that resolvePath returns a path in the case it has on disk, with its symlinks
// resolved, no matter how it was typed. A temporary directory on macOS sits beneath /var, a symlink to /private/var, so
// the symlink half is exercised without any setup of its own.
func TestResolvePathRestoresCase(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	actual := filepath.Join(dir, "MyLib", "Sub")
	c.NoError(os.MkdirAll(actual, 0o750))
	typed := filepath.Join(dir, "mylib", "sub")
	if _, err := os.Stat(typed); err != nil {
		t.Skip("the filesystem is case-sensitive, so there is no case to restore")
	}

	want, err := filepath.EvalSymlinks(actual)
	c.NoError(err)
	c.NotEqual(actual, want, "a temporary directory is expected to sit beneath a symlink")

	got, err := resolvePath(typed)
	c.NoError(err)
	c.Equal(want, got, "the case on disk must be restored and the symlinks resolved")

	got, err = resolvePath(actual)
	c.NoError(err)
	c.Equal(want, got, "a path already in the right case must come back the same")

	_, err = resolvePath(filepath.Join(dir, "missing"))
	c.HasError(err, "a path that does not exist cannot be resolved")
}
