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

type traitListProvider struct {
	traits []*gurps.Trait
}

func (p *traitListProvider) DataOwner() gurps.DataOwner {
	return nil
}

func (p *traitListProvider) TraitList() []*gurps.Trait {
	return p.traits
}

func (p *traitListProvider) SetTraitList(list []*gurps.Trait) {
	p.traits = list
}

// NewTraitTableDockableFromFile loads a list of traits from a file and creates a new unison.Dockable for them.
func NewTraitTableDockableFromFile(filePath string) (unison.Dockable, error) {
	traits, err := gurps.NewTraitsFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	d := NewTraitTableDockable(filePath, traits)
	d.needsSaveAsPrompt = false
	return d, nil
}

// NewTraitTableDockable creates a new unison.Dockable for trait list files.
func NewTraitTableDockable(filePath string, traits []*gurps.Trait) *TableDockable[*gurps.Trait] {
	provider := &traitListProvider{traits: traits}
	return NewTableDockable(filePath, gurps.TraitsExt, NewTraitsProvider(provider, false),
		func(path string) error { return gurps.SaveTraits(provider.TraitList(), path) },
		NewTraitItemID, NewTraitContainerItemID)
}
