// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

// Package updater applies an application update in place: it downloads the release built for the running platform,
// verifies it, swaps it into the installation, and relaunches.
//
// Nothing here depends on the UI, so that it can be exercised headlessly and so that the helper process -- which
// finishes the update after the application has exited -- can run it without starting a window.
//
// User-visible text belongs in the ux package. Failures are reported as a stable Reason code the caller maps to a
// localized message, because that reason is written to a file read after the swap by a *different* build of GCS,
// possibly running in a different language.
package updater
