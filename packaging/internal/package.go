// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package internal

import (
	"github.com/richardwilkes/gcs/v5/early"
	"github.com/richardwilkes/gcs/v5/ux"
	"github.com/richardwilkes/toolbox/fatal"
	"github.com/richardwilkes/unison"
)

// Package performs the platform-specific packaging for GCS.
func Package() {
	early.Configure()
	fatal.IfErr(loadBaseImages())
	ux.RegisterExternalFileTypes()
	ux.RegisterGCSFileTypes()

	// The doc icons use unison's image code to generate the icons, so we need to start it up.
	unison.Start(
		unison.StartupFinishedCallback(func() {
			w, err := unison.NewWindow("")
			fatal.IfErr(err)
			w.ToFront()
			unison.InvokeTask(func() {
				fatal.IfErr(platformPackage())
				w.Dispose() // Will cause the app to quit.
			})
		}),
	)
}
