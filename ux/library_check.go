// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
)

// libraryCheckTimeout bounds a check made on demand from the Library Explorer. It is far shorter than the background
// checks allow themselves, since the user is waiting on this one.
const libraryCheckTimeout = time.Minute

// libraryUpdateButtonsEnabled returns true if the Library Explorer's update and release-notes buttons should be enabled
// for the given library: when a check has turned up a release to offer, or when no check has completed and there is a
// repository to ask, so that the buttons remain a way to get at a library's releases whatever the periodic checks are
// set to. Clicking either button in that second state makes the check first (see checkLibraryReleases).
func libraryUpdateButtonsEnabled(lib *gurps.Library) bool {
	_, releases := lib.AvailableReleases()
	return (len(releases) != 0 && releases[0].HasUpdate()) || lib.NeedsUpgradeCheck()
}

// checkLibraryReleases makes sure the releases of the given libraries are known, returning true once they are.
// Libraries whose releases are already known aren't asked about again, so when the periodic checks have been running
// this returns at once. Otherwise a small window reports on the check while it runs, which the user may cancel, and the
// result is false if they did or if a library still has no answer once the checks have finished, the failure having
// been reported. Must be called on the UI thread; it runs a modal loop while waiting on the checks.
func checkLibraryReleases(libs []*gurps.Library) bool {
	var pending []*gurps.Library
	for _, lib := range libs {
		if lib.NeedsUpgradeCheck() {
			pending = append(pending, lib)
		}
	}
	if len(pending) == 0 {
		return true
	}
	ctx, cancel := context.WithTimeout(context.Background(), libraryCheckTimeout)
	defer cancel()
	wnd, _, err := newLibraryProgressWindow(i18n.Text("Checking…"), libraryCheckTitle(pending),
		unison.NewProgressBar(0), cancel)
	if err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to check for library updates"), err)
		return false
	}
	go func() {
		client := &http.Client{}
		var wg sync.WaitGroup
		for _, lib := range pending {
			wg.Go(func() { lib.CheckForAvailableUpgrade(ctx, client) })
		}
		wg.Wait()
		// The teardown is posted to the UI thread rather than done here, since StopModal mutates the modal state that
		// the UI thread is spinning on inside RunModal without any locking.
		unison.InvokeTask(func() { wnd.StopModal(unison.ModalResponseOK) })
	}()
	wnd.RunModal()
	if errors.Is(ctx.Err(), context.Canceled) {
		return false // The user asked for this, so there is nothing to report
	}
	for _, lib := range pending {
		if lib.NeedsUpgradeCheck() {
			unison.ErrorDialogWithMessage(fmt.Sprintf(i18n.Text("Unable to check for %s updates"), lib.Data().Title),
				xstrings.Wrap("", i18n.Text("The library's releases could not be retrieved. Check its settings and your network connection, then try again."), 100))
			return false
		}
	}
	return true
}

// libraryCheckTitle describes what the check is looking at.
func libraryCheckTitle(libs []*gurps.Library) string {
	if len(libs) == 1 {
		return fmt.Sprintf(i18n.Text("Checking for %s updates…"), libs[0].Data().Title)
	}
	return i18n.Text("Checking for library updates…")
}

// reportNoLibraryReleases tells the user that a library they asked to update, or to see the release notes of, has no
// release to offer. Until a check has been made the buttons are enabled on the strength of the repository alone, so
// this is the first they hear of it.
func reportNoLibraryReleases(lib *gurps.Library) {
	unison.WarningDialogWithMessage(fmt.Sprintf(i18n.Text("No releases are available for %s"), lib.Data().Title),
		xstrings.Wrap("", fmt.Sprintf(i18n.Text("The library's repository has no release that this version of %s can use."),
			xos.AppName), 100))
}
