/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ux

import (
	"os"
	"path/filepath"

	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/unison"
)

type equipmentModifierListProvider struct {
	modifiers []*model.EquipmentModifier
}

func (p *equipmentModifierListProvider) Entity() *model.Entity {
	return nil
}

func (p *equipmentModifierListProvider) EquipmentModifierList() []*model.EquipmentModifier {
	return p.modifiers
}

func (p *equipmentModifierListProvider) SetEquipmentModifierList(list []*model.EquipmentModifier) {
	p.modifiers = list
}

// NewEquipmentModifierTableDockableFromFile loads a list of equipment modifiers from a file and creates a new
// unison.Dockable for them.
func NewEquipmentModifierTableDockableFromFile(filePath string) (unison.Dockable, error) {
	modifiers, err := model.NewEquipmentModifiersFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	d := NewEquipmentModifierTableDockable(filePath, modifiers)
	d.needsSaveAsPrompt = false
	return d, nil
}

// NewEquipmentModifierTableDockable creates a new unison.Dockable for equipment modifier list files.
func NewEquipmentModifierTableDockable(filePath string, modifiers []*model.EquipmentModifier) *TableDockable[*model.EquipmentModifier] {
	provider := &equipmentModifierListProvider{modifiers: modifiers}
	return NewTableDockable(filePath, model.EquipmentModifiersExt,
		NewEquipmentModifiersProvider(provider, false),
		func(path string) error { return model.SaveEquipmentModifiers(provider.EquipmentModifierList(), path) },
		NewEquipmentModifierItemID, NewEquipmentContainerModifierItemID)
}
