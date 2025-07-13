// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package main

import (
	"flag"

	"github.com/richardwilkes/gcs/v5/early"
	"github.com/richardwilkes/toolbox/v2/xflag"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xyaml"
	"github.com/richardwilkes/unison/cmd/upack/packager"
)

func main() {
	early.Configure()
	baseAppName := xos.AppName
	xos.AppName += " Packager"
	xos.AppCmdName += "pack"
	xos.AppIdentifier += ".pack"
	xflag.SetUsage(nil, "A tool for packaging "+baseAppName+" for distribution.", "")
	release := flag.String("release", "", "The release `version` (e.g. \"1.2.3\") to package")
	createDist := flag.Bool("dist", false, "Enable creation of a distribution package")
	xflag.Parse()
	if *release == "" {
		xos.ExitWithMsg("A release version must be specified.")
	}
	var cfg packager.Config
	xos.ExitIfErr(xyaml.Load("packaging.yml", &cfg))
	xos.ExitIfErr(packager.Package(&cfg, *release, *createDist))
	xos.Exit(0)
}
