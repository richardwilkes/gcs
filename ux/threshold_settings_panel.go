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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/threshold"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/check"
)

type thresholdSettingsPanel struct {
	unison.Panel
	pool         *poolSettingsPanel
	threshold    *gurps.PoolThreshold
	deleteButton *unison.Button
}

func newThresholdSettingsPanel(pool *poolSettingsPanel, thresh *gurps.PoolThreshold) *thresholdSettingsPanel {
	p := &thresholdSettingsPanel{
		pool:      pool,
		threshold: thresh,
	}
	p.Self = p
	configureEditorRow(p.AsPanel(), 3, false)
	p.AddChild(NewDragHandle(editorRowDragKey, &editorRowDragData{
		editor: pool.dockable,
		row:    p.AsPanel(),
		title:  i18n.Text("Pool Threshold Drag"),
		move: func(to int) bool {
			return moveEntry(&pool.def.Thresholds, slices.Index(pool.def.Thresholds, thresh), to)
		},
	}))
	p.AddChild(p.createButtons())
	p.AddChild(p.createContent())
	return p
}

func (p *thresholdSettingsPanel) createButtons() *unison.Panel {
	p.deleteButton = unison.NewSVGButton(unison.TrashSVG)
	p.deleteButton.ClickCallback = func() { p.pool.deleteThreshold(p) }
	p.deleteButton.Tooltip = newWrappedTooltip(i18n.Text("Remove pool threshold"))
	p.deleteButton.SetEnabled(len(p.pool.def.Thresholds) > 1)
	return newEditorRowButtonColumn(p.deleteButton)
}

func (p *thresholdSettingsPanel) createContent() *unison.Panel {
	content := unison.NewPanel()
	content.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	content.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})

	mgr := p.pool.dockable.targetMgr
	field := addLabelAndTargetedStringField(content, mgr, p.threshold.KeyPrefix+"state", i18n.Text("State"),
		i18n.Text("A short description of the threshold state"), prototypeMinIDWidth,
		func() string { return p.threshold.State },
		func(s string) { p.threshold.State = s })
	field.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill})
	addLabelAndScriptField(content, mgr, p.threshold.KeyPrefix+"threshold", i18n.Text("Threshold"),
		i18n.Text("The value where the threshold takes effect, which may be a number or a script expression"),
		func() string { return p.threshold.Value },
		func(s string) { p.threshold.Value = s }, false)

	for _, op := range threshold.Ops[1:] {
		content.AddChild(unison.NewPanel())
		content.AddChild(p.createOpCheckBox(op))
	}

	addLabelAndScriptField(content, mgr, p.threshold.KeyPrefix+"explanation", i18n.Text("Explanation"),
		i18n.Text("A explanation of the effects of the threshold state. This field will be interpreted as markdown and may have scripts embedded in it by wrapping each script in <script>your script goes here</script> tags."),
		func() string { return p.threshold.Explanation },
		func(value string) {
			p.threshold.Explanation = value
			content.MarkForLayoutAndRedraw()
			MarkModified(content)
		}, true)

	return content
}

func (p *thresholdSettingsPanel) createOpCheckBox(op threshold.Op) *CheckBox {
	c := NewCheckBox(p.pool.dockable.targetMgr, p.threshold.KeyPrefix+op.Key(), op.String(),
		func() check.Enum { return check.FromBool(p.threshold.ContainsOp(op)) },
		func(state check.Enum) {
			if state == check.On {
				p.threshold.AddOp(op)
			} else {
				p.threshold.RemoveOp(op)
			}
		})
	c.Tooltip = newWrappedTooltip(op.AltString())
	return c
}
