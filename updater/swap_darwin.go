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
	"errors"
	"log/slog"
	"time"

	"github.com/richardwilkes/toolbox/v2/errs"
	"golang.org/x/sys/unix"
)

// exchange swaps the staged copy and the installed one in a single atomic operation.
//
// macOS has a rename that exchanges two existing paths rather than replacing one with the other, which suits this
// exactly. It removes the window the two-rename form leaves -- there is no instant at which the application is not at
// its usual path, so no crash or power loss can leave the user without one -- and it sidesteps the fact that an
// ordinary rename cannot replace a non-empty directory, which is what an application bundle is.
//
// It also does the work of the backup for free: after the exchange, the previous version is sitting at the staged
// copy's path, so it only has to be moved to where the record says the backup lives.
//
// Returns false if the exchange is not available on this filesystem, leaving the caller to fall back to the two-rename
// form.
func exchange(payload, target, backup string) (bool, error) {
	var err error
	for i := range renameAttempts {
		if err = unix.RenamexNp(payload, target, unix.RENAME_SWAP); err == nil {
			slog.Info("replacing the installed version", "target", target, "payload", payload, "backup", backup)
			// The previous version is now where the staged copy was. Move it to the recorded backup path so that the
			// startup sweep can find it, and so that it is not removed along with the staging directory.
			if renameErr := renameWithRetry(payload, backup); renameErr != nil {
				// The update itself succeeded; only the tidying did not. Say so and carry on rather than undoing a
				// good installation over a misplaced backup.
				slog.Warn("unable to move the previous version aside after the update", "error", renameErr)
			}
			return true, nil
		}
		switch {
		case errors.Is(err, unix.ENOTSUP), errors.Is(err, unix.EINVAL):
			// The filesystem does not support the exchange. Not an error worth reporting: the two-rename form below
			// works everywhere.
			return false, nil
		case errors.Is(err, unix.EPERM), errors.Is(err, unix.EACCES):
			// Refused rather than transiently unavailable, most likely because macOS wants explicit permission for one
			// application to modify another. Retrying cannot help, and there is no way to ask from here.
			return false, errs.NewWithCause("not allowed to replace the installed application", err)
		}
		if i < renameAttempts-1 {
			time.Sleep(renameDelay)
		}
	}
	slog.Warn("unable to exchange the installed application atomically; falling back", "error", err)
	return false, nil
}
