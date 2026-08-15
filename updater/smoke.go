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
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/richardwilkes/toolbox/v2/errs"
)

// smokeTestTimeout bounds the trial run. Printing a version string is near-instant; anything approaching this means the
// executable is not doing what it was asked, which is itself a reason not to install it.
const smokeTestTimeout = 30 * time.Second

// verifyRuns proves the staged application actually runs on this machine before anything is swapped in.
//
// This is the one check that covers the failures nothing else can see. A checksum establishes that the right bytes
// arrived, and a signature that the right developer produced them; neither says the result will start here. This
// catches a build for the wrong processor, a Linux build wanting a newer glibc than the system has, a macOS bundle the
// kernel refuses to execute, and a corrupted extraction whose checksum passed because the checksum itself was wrong.
//
// The cost of skipping it is severe and asymmetric: those failures all succeed at swap time and only fail at launch,
// which is the one point where there is nothing left to roll back to -- the user is simply left without a working
// application. Running it here means such a release is reported as unusable while the installation is still untouched.
//
// The "-v" flag prints the version and exits during flag parsing, well before any window is created, so this never
// puts anything on screen.
func verifyRuns(ctx context.Context, exePath, wantVersion string) error {
	cctx, cancel := context.WithTimeout(ctx, smokeTestTimeout)
	defer cancel()
	cmd := exec.CommandContext(cctx, exePath, "-v") //nolint:gosec // The path is one this package just staged
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = strings.TrimSpace(stdout.String())
		}
		if detail != "" {
			return errs.NewWithCause("the downloaded version will not run on this system: "+detail, err)
		}
		return errs.NewWithCause("the downloaded version will not run on this system", err)
	}
	// A clean exit is the requirement. The version string is checked only when there is one to check: the Windows
	// build is linked as a GUI application and has no console to write to, so it exits zero having printed nothing.
	reported := strings.TrimSpace(stdout.String())
	if reported == "" || wantVersion == "" {
		return nil
	}
	// A build made from a modified checkout has a "~" appended, so match on the prefix rather than for equality.
	if !strings.HasPrefix(reported, wantVersion) {
		return errs.Newf("the downloaded version reports itself as %s rather than %s", reported, wantVersion)
	}
	return nil
}
