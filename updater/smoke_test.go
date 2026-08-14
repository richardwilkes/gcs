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
	"runtime"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xos"
)

// buildStandIn compiles a small program that behaves the way the real executable does when given "-v": it prints the
// version and exits. src decides how it misbehaves instead.
func buildStandIn(t *testing.T, src string) string {
	t.Helper()
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("the Go toolchain is needed to build the stand-in executable")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module standin\n\ngo 1.26.0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	exePath := filepath.Join(dir, CmdName)
	if runtime.GOOS == xos.WindowsOS {
		exePath += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", exePath, ".")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("unable to build the stand-in executable: %v\n%s", err, out)
	}
	return exePath
}

// TestVerifyRunsAcceptsAWorkingBuild verifies the trial run passes for an executable that starts and reports the
// version that was expected.
func TestVerifyRunsAcceptsAWorkingBuild(t *testing.T) {
	c := check.New(t)
	exePath := buildStandIn(t, `package main

import "fmt"

func main() { fmt.Println("5.46.0") }
`)
	c.NoError(verifyRuns(t.Context(), exePath, "5.46.0"))
}

// TestVerifyRunsAcceptsAModifiedBuildMarker verifies the version comparison tolerates the "~" that a build made from a
// modified checkout appends, which would otherwise be read as the wrong version entirely.
func TestVerifyRunsAcceptsAModifiedBuildMarker(t *testing.T) {
	c := check.New(t)
	exePath := buildStandIn(t, `package main

import "fmt"

func main() { fmt.Println("5.46.0~") }
`)
	c.NoError(verifyRuns(t.Context(), exePath, "5.46.0"))
}

// TestVerifyRunsRejectsABuildThatFails is the case this check exists for. A downloaded release that cannot start on
// this machine -- wrong processor, a system library too old, a corrupted extraction -- must be caught here, while the
// installation is still untouched, rather than at the launch that follows the swap, where there is nothing left to
// fall back to.
func TestVerifyRunsRejectsABuildThatFails(t *testing.T) {
	c := check.New(t)
	exePath := buildStandIn(t, `package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "error while loading shared libraries: libc.so.6: version not found")
	os.Exit(1)
}
`)
	err := verifyRuns(t.Context(), exePath, "5.46.0")
	c.HasError(err)
	c.Contains(err.Error(), "will not run on this system")
	c.Contains(err.Error(), "libc.so.6", "the reason the executable gave must survive into the error")
}

// TestVerifyRunsRejectsTheWrongVersion verifies that an executable which starts but is not the release that was
// downloaded is refused, which would mean the wrong asset had been fetched or unpacked.
func TestVerifyRunsRejectsTheWrongVersion(t *testing.T) {
	c := check.New(t)
	exePath := buildStandIn(t, `package main

import "fmt"

func main() { fmt.Println("5.30.0") }
`)
	err := verifyRuns(t.Context(), exePath, "5.46.0")
	c.HasError(err)
	c.Contains(err.Error(), "5.30.0")
}

// TestVerifyRunsAcceptsSilence verifies that an executable exiting cleanly without printing anything is accepted. The
// Windows build is linked as a GUI application, so it has no console to write to and reports nothing even though it
// ran correctly -- a clean exit is the requirement, not the output.
func TestVerifyRunsAcceptsSilence(t *testing.T) {
	c := check.New(t)
	exePath := buildStandIn(t, `package main

func main() {}
`)
	c.NoError(verifyRuns(t.Context(), exePath, "5.46.0"))
}

// TestVerifyRunsRejectsSomethingUnrunnable verifies the case a corrupted download produces: a file that exists and is
// marked executable but is not a program at all.
func TestVerifyRunsRejectsSomethingUnrunnable(t *testing.T) {
	c := check.New(t)
	exePath := filepath.Join(t.TempDir(), CmdName)
	write(t, exePath, "this is definitely not an executable")
	c.NoError(os.Chmod(exePath, executableModePerm))
	c.HasError(verifyRuns(t.Context(), exePath, "5.46.0"))

	c.HasError(verifyRuns(t.Context(), filepath.Join(t.TempDir(), "does-not-exist"), "5.46.0"))
}
