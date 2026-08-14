// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

//go:build !darwin

package updater

import (
	"context"
	"os"
	"runtime"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xio"
	"github.com/richardwilkes/toolbox/v2/xos"
)

// stagePayload unpacks the downloaded archive. Both the Linux and Windows distributions hold exactly one file, the
// executable, so this extracts that one file and refuses anything else.
func stagePayload(_ context.Context, archivePath, dstPath string) error {
	wantName, err := PayloadName(runtime.GOOS)
	if err != nil {
		return err
	}
	if runtime.GOOS == xos.WindowsOS {
		return ExtractSingleZip(archivePath, wantName, dstPath)
	}
	f, err := os.Open(archivePath) //nolint:gosec // The path is one this package just created in its own staging area
	if err != nil {
		return errs.Wrap(err)
	}
	defer xio.CloseIgnoringErrors(f)
	return ExtractSingleTGZ(f, wantName, dstPath)
}

// verifyPayload has nothing to check on these platforms. The distributions carry no signature -- the packager signs
// only on macOS -- so the SHA-256 recorded by GitHub, checked during the download, is the whole of the integrity
// guarantee here. What follows it, running the staged executable to see that it works, is what catches the rest.
func verifyPayload(_ context.Context, _, _ string) error {
	return nil
}
