// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

// Package gurpstest provides helpers for tests that build gurps data. Since it imports gurps, the gurps package's
// internal tests cannot import it, so they keep private copies of these helpers; its external (gurps_test) tests can.
package gurpstest
