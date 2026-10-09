// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

// Package calculator holds the rules behind the calculators GCS offers at the table: explosions and area attacks,
// scatter, demolition, collisions and falls, jumping, throwing and hiking. Each calculator's inputs, choice tables,
// sheet readers and results live here; the ux package only lays them out, formats them and explains them. The choice
// tables are built on every call rather than held in package-level variables, so that their names follow the language
// the user chose, which is only set after the packages have initialized.
package calculator
