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
	"fmt"

	"github.com/richardwilkes/gcs/v5/early"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/v2/xflag"
	"github.com/richardwilkes/toolbox/v2/xos"
)

func main() {
	early.Configure()
	xos.AppName += " Prerequisite Counts"
	xos.AppCmdName = "prereq-counts"
	dice.GURPSFormat = true
	xflag.SetUsage(nil, "A tool for updating the prerequisite field data in GURPS spell files.", "")
	xflag.Parse()
	for _, one := range flag.Args() {
		fmt.Printf("Processing %s\n", one)
		spells, err := gurps.NewSpellsFromFile(nil, one)
		xos.ExitIfErr(err)
		gurps.Traverse(func(spell *gurps.Spell) bool {
			spell.PrereqCount = gurps.CountPrereqsForSpell(spell, spells)
			return false
		}, false, true, spells...)
		xos.ExitIfErr(gurps.SaveSpells(spells, one))
	}
}
