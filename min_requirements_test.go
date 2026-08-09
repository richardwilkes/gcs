// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/errs"
)

const minRequirementsPath = "cmd/min_requirements.sh"

// TestExtractArchivesSurvivesUnmountableDMG runs the script's extract_archives function, as committed, against a .dmg
// that can't be mounted. The function is deliberately tolerant of unusable archives everywhere else, and it runs under
// `set -euo pipefail`, so a failing mount must not take the entire analysis down with it.
func TestExtractArchivesSurvivesUnmountableDMG(t *testing.T) {
	c := check.New(t)
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash is not available")
	}
	dir := t.TempDir()
	c.NoError(os.WriteFile(filepath.Join(dir, "corrupt.dmg"), []byte("not a disk image"), 0o600))
	out, err := runExtractArchives(c, bash, dir, "exit 1")
	c.NoError(err, "a .dmg that can't be mounted aborted the analysis: "+out)
	c.Contains(out, "finished", "extract_archives never returned")
	c.Contains(out, "warning: unable to mount")
}

// TestExtractArchivesDetachesMountedDMG verifies the other half of the mount handling: when the mount succeeds, the
// mount point is parsed out of hdiutil's output and handed back to hdiutil to detach, so the test run doesn't leave
// disk images attached.
func TestExtractArchivesDetachesMountedDMG(t *testing.T) {
	c := check.New(t)
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash is not available")
	}
	dir := t.TempDir()
	c.NoError(os.WriteFile(filepath.Join(dir, "good.dmg"), []byte("disk image"), 0o600))
	out, err := runExtractArchives(c, bash, dir,
		`printf '/dev/disk4s1\tApple_HFS\t/Volumes/GCS\n'`+"\nexit 0")
	c.NoError(err, out)
	c.Contains(out, "finished")
	var log []byte
	log, err = os.ReadFile(filepath.Join(dir, "hdiutil.log"))
	c.NoError(err)
	c.Contains(string(log), "detach -quiet /Volumes/GCS")
}

// runExtractArchives runs the committed extract_archives function against dir with a stand-in for hdiutil whose
// behavior is supplied by attachScript, which runs only for the "attach" verb. Every invocation is logged to
// hdiutil.log within dir.
func runExtractArchives(c check.Checker, bash, dir, attachScript string) (string, error) {
	fn, err := extractShellFunction(minRequirementsPath, "extract_archives")
	c.NoError(err)
	binDir := filepath.Join(dir, "bin")
	c.NoError(os.Mkdir(binDir, 0o755))
	c.NoError(os.WriteFile(filepath.Join(binDir, "hdiutil"), []byte("#!/usr/bin/env bash\n"+
		`echo "$@" >>"`+filepath.Join(dir, "hdiutil.log")+"\"\n"+
		"if [ \"$1\" = attach ]; then\n"+attachScript+"\nfi\nexit 0\n"), 0o700))
	scriptPath := filepath.Join(dir, "harness.sh")
	// The same options the script itself runs under, since they are what makes a failing mount fatal.
	c.NoError(os.WriteFile(scriptPath, []byte("set -euo pipefail\n"+fn+"\nextract_archives \"$1\"\necho finished\n"),
		0o600))
	cmd := exec.Command(bash, scriptPath, dir)
	cmd.Env = []string{"PATH=" + binDir + string(os.PathListSeparator) + os.Getenv("PATH")}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// extractShellFunction returns the text of the named shell function, from its opening line through the closing brace
// in the first column, so tests can exercise a piece of a script that would otherwise run the whole thing.
func extractShellFunction(path, name string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	for i, line := range lines {
		if line != name+"() {" {
			continue
		}
		for j := i + 1; j < len(lines); j++ {
			if lines[j] == "}" {
				return strings.Join(lines[i:j+1], "\n"), nil
			}
		}
		return "", errs.Newf("%s: no closing brace found for the %s function", path, name)
	}
	return "", errs.Newf("%s: no %s function found", path, name)
}
