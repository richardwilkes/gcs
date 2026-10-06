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
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestModelDoesNotDependOnUnison verifies that nothing beneath model, its tests included, depends on unison, whether
// directly or through another package, so that the model can be used without a user interface. Whatever the model
// needs from one goes through gurps.Host instead. Each operating system CI builds for is checked, since the model has
// per-platform sources, as is the race build tag, which selects the TestRace wrappers.
func TestModelDoesNotDependOnUnison(t *testing.T) {
	c := check.New(t)
	// The go command on the PATH is the one build.sh and CI run the tests with. The test binary is given no GOROOT it
	// could use to find the toolchain that built it.
	goCmd, err := exec.LookPath("go")
	if err != nil {
		t.Skip("the go tool is not available")
	}
	const unisonPath = "github.com/richardwilkes/unison"
	for _, goos := range []string{"darwin", "linux", "windows"} {
		for _, tags := range []string{"", "race"} {
			cmd := exec.Command(goCmd, "list", "-test", "-tags", tags, "-f", `{{.ImportPath}}|{{join .Deps " "}}`,
				"./model/...")
			cmd.Env = append(os.Environ(), "GOOS="+goos)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			var out []byte
			out, err = cmd.Output()
			c.NoError(err, "GOOS=%s tags=%q: %s", goos, tags, stderr.String())
			packages := 0
			for line := range strings.Lines(string(out)) {
				pkg, deps, _ := strings.Cut(strings.TrimSpace(line), "|")
				packages++
				for dep := range strings.FieldsSeq(deps) {
					if dep == unisonPath || strings.HasPrefix(dep, unisonPath+"/") {
						t.Errorf("GOOS=%s tags=%q: %s depends on %s", goos, tags, pkg, dep)
						break
					}
				}
			}
			c.True(packages > 0, "GOOS=%s tags=%q: no packages were listed", goos, tags)
		}
	}
}
