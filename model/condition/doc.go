// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

// Package condition evaluates condition trees, the composable rules behind prerequisites, list filters and selectors:
// groups, checks on an item's fields, counts of a collection's members, compares of two numbers, and scripts. The
// package knows nothing of the items it tests; a Host supplies their collections, fields, attributes and scripts.
package condition
