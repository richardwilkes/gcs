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
	"slices"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

type poolSettingsPanel struct {
	unison.Panel
	dockable *attributeSettingsDockable
	def      *gurps.AttributeDef
}

func newPoolSettingsPanel(dockable *attributeSettingsDockable, def *gurps.AttributeDef) *poolSettingsPanel {
	p := &poolSettingsPanel{
		dockable: dockable,
		def:      def,
	}
	p.Self = p
	p.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, 0, unison.NewUniformInsets(1), false))
	p.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: align.Fill,
		HGrab:  true,
	})
	for _, threshold := range def.Thresholds {
		p.AddChild(newThresholdSettingsPanel(p, threshold))
	}
	return p
}

func (p *poolSettingsPanel) addThreshold() {
	undo := &unison.UndoEdit[[]*gurps.PoolThreshold]{
		ID:       unison.NextUndoID(),
		EditName: i18n.Text("Add Pool Threshold"),
		UndoFunc: func(e *unison.UndoEdit[[]*gurps.PoolThreshold]) {
			p.applyThresholds(e.BeforeData)
		},
		RedoFunc: func(e *unison.UndoEdit[[]*gurps.PoolThreshold]) {
			p.applyThresholds(e.AfterData)
		},
		AbsorbFunc: func(_ *unison.UndoEdit[[]*gurps.PoolThreshold], _ unison.Undoable) bool { return false },
	}
	undo.BeforeData = clonePoolThresholds(p.def.Thresholds)
	threshold := &gurps.PoolThreshold{KeyPrefix: p.dockable.targetMgr.NextPrefix()}
	p.def.Thresholds = append(p.def.Thresholds, threshold)
	newThreshold := newThresholdSettingsPanel(p, threshold)
	p.AddChild(newThreshold)
	if children := p.Children(); len(children) == 2 {
		if panel, ok := children[0].Self.(*thresholdSettingsPanel); ok {
			panel.deleteButton.SetEnabled(true)
		}
	}
	undo.AfterData = clonePoolThresholds(p.def.Thresholds)
	p.dockable.UndoManager().Add(undo)
	p.dockable.MarkModified(nil)
	p.MarkForLayoutRecursivelyUpward()
	p.dockable.ValidateLayout()
	focus := newThreshold.Children()[2]
	focus.RequestFocus()
	focus.ScrollIntoView()
}

func (p *poolSettingsPanel) deleteThreshold(target *thresholdSettingsPanel) {
	i := p.IndexOfChild(target)
	target.RemoveFromParent()
	if children := p.Children(); len(children) == 1 {
		if panel, ok := children[0].Self.(*thresholdSettingsPanel); ok {
			panel.deleteButton.SetEnabled(false)
		}
	}
	undo := &unison.UndoEdit[[]*gurps.PoolThreshold]{
		ID:       unison.NextUndoID(),
		EditName: i18n.Text("Delete Pool Threshold"),
		UndoFunc: func(e *unison.UndoEdit[[]*gurps.PoolThreshold]) {
			p.applyThresholds(e.BeforeData)
		},
		RedoFunc: func(e *unison.UndoEdit[[]*gurps.PoolThreshold]) {
			p.applyThresholds(e.AfterData)
		},
		AbsorbFunc: func(_ *unison.UndoEdit[[]*gurps.PoolThreshold], _ unison.Undoable) bool { return false },
	}
	undo.BeforeData = clonePoolThresholds(p.def.Thresholds)
	p.def.Thresholds = slices.Delete(p.def.Thresholds, i, i+1)
	undo.AfterData = clonePoolThresholds(p.def.Thresholds)
	p.dockable.UndoManager().Add(undo)
	p.dockable.MarkModified(nil)
}

func (p *poolSettingsPanel) applyThresholds(thresholds []*gurps.PoolThreshold) {
	p.def.Thresholds = clonePoolThresholds(thresholds)
	p.dockable.sync()
}

func clonePoolThresholds(in []*gurps.PoolThreshold) []*gurps.PoolThreshold {
	thresholds := make([]*gurps.PoolThreshold, len(in))
	for i, one := range in {
		thresholds[i] = one.Clone()
	}
	return thresholds
}
