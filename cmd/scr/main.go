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
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/early"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xfilepath"
	"github.com/richardwilkes/toolbox/v2/xflag"
	"github.com/richardwilkes/toolbox/v2/xio"
	"github.com/richardwilkes/toolbox/v2/xos"
)

const dirToUpdate = "/Users/rich/code/gurps_campaign/Library"

var fileSet = []struct {
	name     string
	extraSCR int
}{
	{"Spells, Arcane.spl", 0},
	{"Spells, Divine.spl", 1},
	{"Spells, Druidic.spl", 1},
}

func main() {
	early.Configure()
	xos.AppName += " SCR Utility"
	xos.AppCmdName = "scr"
	dice.GURPSFormat = true
	xflag.SetUsage(nil, "A tool for updating or exporting SCR data in GURPS spell files.", "")
	export := flag.Bool("export", false, "Export SCR data to CSV files")
	xflag.Parse()
	if *export {
		for _, one := range fileSet {
			xos.ExitIfErr(exportSCR(dirToUpdate, one.name, one.extraSCR))
		}
	} else {
		for _, one := range fileSet {
			xos.ExitIfErr(updateSCR(dirToUpdate, one.name, one.extraSCR))
		}
	}
}

func updateSCR(dir, fileName string, extraSCR int) error {
	spells, err := gurps.NewSpellsFromFile(os.DirFS(dir), fileName)
	if err != nil {
		return err
	}
	gurps.Traverse(func(spell *gurps.Spell) bool {
		var tags []string
		for _, tag := range spell.Tags {
			if !strings.HasPrefix(tag, "SCR") {
				tags = append(tags, tag)
			}
		}
		tags = append(tags, fmt.Sprintf("SCR %d", extraSCR+gurps.CountPrereqsForSpell(spell, spells, 1, false)))
		slices.Sort(tags)
		spell.Tags = tags
		return false
	}, false, true, spells...)
	return gurps.SaveSpells(spells, filepath.Join(dir, fileName))
}

func exportSCR(dir, fileName string, extraSCR int) error {
	spells, err := gurps.NewSpellsFromFile(os.DirFS(dir), fileName)
	if err != nil {
		return err
	}
	var out *os.File
	if out, err = os.Create(xfilepath.BaseName(fileName) + ".csv"); err != nil {
		return errs.Wrap(err)
	}
	w := csv.NewWriter(out)
	gurps.Traverse(func(spell *gurps.Spell) bool {
		scr := extraSCR + gurps.CountPrereqsForSpell(spell, spells, 1, false)
		if err = w.Write([]string{spell.NameWithReplacements(), fmt.Sprintf("%d", scr)}); err != nil {
			err = errs.Wrap(err)
			return true
		}
		return false
	}, false, true, spells...)
	if err == nil {
		w.Flush()
	}
	if err != nil {
		xio.CloseIgnoringErrors(out)
		return err
	}
	return errs.Wrap(out.Close())
}
