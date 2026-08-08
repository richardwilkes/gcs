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
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestConvertWalkerUnreadableRoot verifies that the conversion walker tolerates a root it can't stat. filepath.WalkDir
// hands the callback a nil DirEntry along with the error in that case, which previously panicked on the immediate
// d.Name() call rather than skipping the path.
func TestConvertWalkerUnreadableRoot(t *testing.T) {
	c := check.New(t)
	pathSet := make(map[string]struct{})
	extSet := map[string]struct{}{SheetExt: {}}
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	c.NotPanics(func() {
		c.NoError(filepath.WalkDir(missing, convertWalker(pathSet, extSet)), "a missing root is skipped, not reported")
	}, "a missing root must not panic")
	c.Equal(0, len(pathSet), "nothing is collected from a missing root")

	// A readable root still collects the files with matching extensions.
	dir := t.TempDir()
	wanted := filepath.Join(dir, "sheet"+SheetExt)
	c.NoError(os.WriteFile(wanted, []byte("{}"), 0o600))
	c.NoError(os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("{}"), 0o600))
	c.NoError(filepath.WalkDir(dir, convertWalker(pathSet, extSet)))
	c.Equal(1, len(pathSet), "only the file with a matching extension is collected")
}

// TestConvertWalkerIgnoresExtensionCase verifies that the conversion walker collects data files whose extension is
// upper- or mixed-case. The switch that dispatches the conversion lowercases the extension, so a file the walker
// declines to collect here is silently skipped by both --convert and --sync, with nothing said about it.
func TestConvertWalkerIgnoresExtensionCase(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	wanted := []string{
		"lower" + SheetExt,
		"upper" + strings.ToUpper(SheetExt),
		"Mixed.GcS",
		"template" + strings.ToUpper(TemplatesExt),
	}
	for _, name := range append(slices.Clone(wanted), "notes.TXT") {
		c.NoError(os.WriteFile(filepath.Join(dir, name), []byte("{}"), 0o600))
	}

	pathSet := make(map[string]struct{})
	extSet := map[string]struct{}{SheetExt: {}, TemplatesExt: {}}
	c.NoError(filepath.WalkDir(dir, convertWalker(pathSet, extSet)))

	// The collected paths have been resolved through any symlinks, so compare the names rather than the paths.
	collected := make(map[string]struct{}, len(pathSet))
	for p := range pathSet {
		collected[filepath.Base(p)] = struct{}{}
	}
	for _, name := range wanted {
		_, exists := collected[name]
		c.True(exists, "%s is collected", name)
	}
	c.Equal(len(wanted), len(collected), "the file with an unrelated extension is not collected")
}
