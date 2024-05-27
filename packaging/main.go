// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package main

import (
	"github.com/richardwilkes/gcs/v5/packaging/internal"
	"github.com/richardwilkes/toolbox/atexit"
	"github.com/richardwilkes/toolbox/fatal"
)

func main() {
	fatal.IfErr(internal.Package())
	atexit.Exit(0)
}
