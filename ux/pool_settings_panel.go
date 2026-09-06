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
	"slices"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
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
	p.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.NewUniformInsets(1), false))
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

// addThreshold appends a new threshold to the pool, undoably, and gives its state field the focus.
func (p *poolSettingsPanel) addThreshold() {
	threshold := &gurps.PoolThreshold{KeyPrefix: p.dockable.targetMgr.NextPrefix()}
	p.dockable.editStructure(i18n.Text("Add Pool Threshold"),
		func() { p.def.Thresholds = append(p.def.Thresholds, threshold) }, threshold.KeyPrefix+"state")
}

// deleteThreshold removes the threshold the given row edits from the pool, undoably.
func (p *poolSettingsPanel) deleteThreshold(target *thresholdSettingsPanel) {
	p.dockable.editStructure(i18n.Text("Delete Pool Threshold"), func() {
		if i := slices.Index(p.def.Thresholds, target.threshold); i != -1 {
			p.def.Thresholds = slices.Delete(p.def.Thresholds, i, i+1)
		}
	}, "")
}
