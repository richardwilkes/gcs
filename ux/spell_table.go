// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"os"
	"path/filepath"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/unison"
)

type spellListProvider struct {
	spells []*gurps.Spell
}

func (p *spellListProvider) DataOwner() gurps.DataOwner {
	return nil
}

func (p *spellListProvider) SpellList() []*gurps.Spell {
	return p.spells
}

func (p *spellListProvider) SetSpellList(list []*gurps.Spell) {
	p.spells = list
}

// NewSpellTableDockableFromFile loads a list of spells from a file and creates a new unison.Dockable for them.
func NewSpellTableDockableFromFile(filePath string) (unison.Dockable, error) {
	spells, err := gurps.NewSpellsFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	d := NewSpellTableDockable(filePath, spells)
	d.needsSaveAsPrompt = false
	return d, nil
}

// NewSpellTableDockable creates a new unison.Dockable for spell list files.
func NewSpellTableDockable(filePath string, spells []*gurps.Spell) *TableDockable[*gurps.Spell] {
	provider := &spellListProvider{spells: spells}
	return NewTableDockable(filePath, gurps.SpellsExt, NewSpellsProvider(provider, false),
		func(path string) error { return gurps.SaveSpells(provider.SpellList(), path) },
		NewSpellItemID, NewSpellContainerItemID, NewRitualMagicSpellItemID)
}
