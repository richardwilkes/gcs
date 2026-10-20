// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/encumbrance"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// syncCountingProvider counts the syncs of the page list it is installed in: PageList.Sync is the only caller of
// SyncHeader.
type syncCountingProvider[T gurps.Node[T]] struct {
	TableProvider[T]
	syncs int
}

func (p *syncCountingProvider[T]) SyncHeader(headers []unison.TableColumnHeader[*Node[T]]) {
	p.syncs++
	p.TableProvider.SyncHeader(headers)
}

// A full rebuild syncs every list it keeps as it lays the page out, and the deep sync that follows used to sync each
// of them a second time, which on a sheet with large lists was half the cost of the rebuild. A list is synced once
// per update, whichever kind of update it is.
func TestFullRebuildSyncsEachListOnce(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Kept"
	entity.Traits = []*gurps.Trait{trait}
	sheet.Rebuild(true)
	counter := &syncCountingProvider[*gurps.Trait]{TableProvider: sheet.Traits.provider}
	sheet.Traits.provider = counter
	list := sheet.Traits

	sheet.Rebuild(true)
	c.Equal(list, sheet.Traits, "the list must have been kept, or the count means nothing")
	c.Equal(1, counter.syncs, "a full rebuild syncs a list it keeps once")
	counter.syncs = 0
	sheet.Rebuild(false)
	c.Equal(1, counter.syncs, "a partial rebuild syncs the list once")
	counter.syncs = 0
	sheet.MarkModified(sheet)
	c.Equal(1, counter.syncs, "marking the sheet modified syncs the list once")

	data := gurps.NewTemplate()
	data.Traits = []*gurps.Trait{gurps.NewTrait(nil, nil, false)}
	template := newTestTemplateDockable("Syncs", data)
	templateCounter := &syncCountingProvider[*gurps.Trait]{TableProvider: template.Traits.provider}
	template.Traits.provider = templateCounter
	templateList := template.Traits
	template.Rebuild(true)
	c.Equal(templateList, template.Traits, "the template's list must have been kept")
	c.Equal(1, templateCounter.syncs, "a full rebuild of a template syncs a list it keeps once")
}

// The encumbrance block finds the current level when it is synced rather than as it is drawn, so a block built for a
// character who is already encumbered has to find it as it is built.
func TestEncumbrancePanelStartsAtTheCurrentLevel(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	// Basic lift at ST 10 is 20 lb, so 30 lb is Light.
	eq := gurps.NewEquipment(entity, nil, false)
	eq.BaseWeight = "30 lb"
	entity.CarriedEquipment = append(entity.CarriedEquipment, eq)
	entity.Recalculate()
	panel := NewEncumbrancePanel(entity)
	c.Equal(int(encumbrance.Light), panel.current, "the block starts at the character's level without a sync")
	c.False(panel.overloaded)
}
