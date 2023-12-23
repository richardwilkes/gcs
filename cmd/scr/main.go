/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/fatal"
)

const dirToUpdate = "/Users/rich/code/gurps_campaign/Library"

func main() {
	dice.GURPSFormat = true
	fatal.IfErr(updateSCR(dirToUpdate, "Spells, Arcane.spl", 0))
	fatal.IfErr(updateSCR(dirToUpdate, "Spells, Divine.spl", 1))
	fatal.IfErr(updateSCR(dirToUpdate, "Spells, Druidic.spl", 1))
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
