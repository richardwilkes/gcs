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
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/richardwilkes/toolbox/v2/errs"
)

const (
	// renameAttempts and renameDelay bound the retries around each rename.
	//
	// These exist for Windows, where antivirus scanners and shell extensions routinely take a brief handle on a
	// freshly written executable and a rename fails with an access error that is gone a moment later. They are
	// harmless on the other platforms, where the first attempt always succeeds.
	renameAttempts = 10
	renameDelay    = 200 * time.Millisecond
)

// renameFunc and exchangeFunc are the operations swap is built from. They are variables so that a test can make a
// chosen step fail and check that the recovery actually restores the installation -- a path that is otherwise
// unreachable, and the one that matters most, since getting it wrong means leaving the user with no application at all.
//
// Substituting exchangeFunc also makes the two-rename fallback reachable on macOS, where the atomic exchange would
// normally be taken. That fallback is not test-only there: it is what runs on any filesystem whose driver does not
// implement the exchange.
var (
	renameFunc   = os.Rename
	exchangeFunc = exchange
)

// swap replaces the installation at target with the staged copy at payload, moving what was there to backup.
//
// The installation is either the old one or the new one at every moment a crash could occur. On macOS this is a single
// atomic exchange, so there is no moment at all where the application is missing. Elsewhere it is two renames within
// one directory, leaving a window of microseconds between them; a crash inside that window leaves the backup intact,
// which the startup repair recovers from.
//
// If the second rename fails, the first is undone before returning, so a failure leaves the installation exactly as it
// was found.
func swap(target, payload, backup string) error {
	if err := validateSwap(target, payload, backup); err != nil {
		return err
	}
	if swapped, err := exchangeFunc(payload, target, backup); swapped || err != nil {
		return err
	}
	if err := renameWithRetry(target, backup); err != nil {
		// Nothing has moved, so there is nothing to undo.
		return errs.NewWithCause("unable to move the existing version aside", err)
	}
	if err := renameWithRetry(payload, target); err != nil {
		if restoreErr := renameWithRetry(backup, target); restoreErr != nil {
			// The destination name is free and the source exists, so this should not be reachable. If it ever is, the
			// installation is missing and only the startup repair can put it back, which is why the backup's location
			// is recorded before any of this begins.
			errs.Log(errs.NewWithCause("unable to restore the existing version after a failed update", restoreErr),
				"backup", backup, "target", target)
		}
		return errs.NewWithCause("unable to put the new version in place", err)
	}
	return nil
}

// validateSwap checks the preconditions that make the renames below safe, before any of them are attempted.
func validateSwap(target, payload, backup string) error {
	if target == "" || payload == "" || backup == "" {
		return errs.New("the update is missing the paths it needs")
	}
	if filepath.Dir(target) != filepath.Dir(backup) {
		// A backup in another directory could be on another filesystem, which would turn the move into a copy: no
		// longer atomic, and no longer something that can be undone by renaming it back.
		return errs.New("the backup must be beside what it replaces")
	}
	if _, err := os.Lstat(backup); err == nil {
		// The installation is about to be renamed to this path, and on most systems that would replace whatever is
		// already there without a word. If it happened to be a backup from an earlier update, the user's only other
		// copy of the application would be destroyed. The name carries a timestamp, so reaching this means something
		// is wrong; refusing costs one update, and being wrong here costs the fallback.
		return errs.New("something is already where the previous version would be moved to")
	}
	payloadInfo, err := os.Lstat(payload)
	if err != nil {
		return errs.NewWithCause("the prepared update is missing", err)
	}
	targetInfo, err := os.Lstat(target)
	if err != nil {
		return errs.NewWithCause("the installed version is missing", err)
	}
	if payloadInfo.IsDir() != targetInfo.IsDir() {
		// A bundle replacing a bare executable, or the reverse, means something upstream resolved the wrong thing.
		return errs.New("the prepared update does not match what it would replace")
	}
	return nil
}

// renameWithRetry renames a path, retrying briefly before giving up. See renameAttempts for why.
func renameWithRetry(from, to string) error {
	var err error
	for i := range renameAttempts {
		if err = renameFunc(from, to); err == nil {
			return nil
		}
		if i < renameAttempts-1 {
			time.Sleep(renameDelay)
		}
	}
	return errs.Wrap(err)
}

// logSwap records what was done, so that a report of an update going wrong has something behind it.
func logSwap(target, payload, backup string) {
	slog.Info("replacing the installed version", "target", target, "payload", payload, "backup", backup)
}
