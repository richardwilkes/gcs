// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

// Package uxtest drives the GCS workspace headlessly for the tests of ux and of the packages built on it, such as
// ux/calculators. It starts a headless unison session running the workspace, finds and closes the dockables a test
// opens, works the in-window menus and popups the way a user would, and audits what a screen reader would be told.
//
// It imports nothing from ux, since a package's own tests may not import a package that imports it. What it needs of
// the workspace -- how to stand it up in a window, and how to list the dockables it holds -- each test package hands
// it through Main, from its TestMain.
package uxtest
