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

type equipmentModifierListProvider struct {
	modifiers []*gurps.EquipmentModifier
}

func (p *equipmentModifierListProvider) DataOwner() gurps.DataOwner {
	return nil
}

func (p *equipmentModifierListProvider) EquipmentModifierList() []*gurps.EquipmentModifier {
	return p.modifiers
}

func (p *equipmentModifierListProvider) SetEquipmentModifierList(list []*gurps.EquipmentModifier) {
	p.modifiers = list
}

// NewEquipmentModifierTableDockableFromFile loads a list of equipment modifiers from a file and creates a new
// unison.Dockable for them.
func NewEquipmentModifierTableDockableFromFile(filePath string) (unison.Dockable, error) {
	modifiers, err := gurps.NewEquipmentModifiersFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	d := NewEquipmentModifierTableDockable(filePath, modifiers)
	d.needsSaveAsPrompt = false
	return d, nil
}

// NewEquipmentModifierTableDockable creates a new unison.Dockable for equipment modifier list files.
func NewEquipmentModifierTableDockable(filePath string, modifiers []*gurps.EquipmentModifier) *TableDockable[*gurps.EquipmentModifier] {
	provider := &equipmentModifierListProvider{modifiers: modifiers}
	return NewTableDockable(filePath, gurps.EquipmentModifiersExt,
		NewEquipmentModifiersProvider(provider, false),
		func(path string) error { return gurps.SaveEquipmentModifiers(provider.EquipmentModifierList(), path) },
		NewEquipmentModifierItemID, NewEquipmentContainerModifierItemID)
}
