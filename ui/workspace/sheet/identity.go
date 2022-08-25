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

package sheet

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/ancestry"
	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// IdentityPanel holds the contents of the identity block on the sheet.
type IdentityPanel struct {
	unison.Panel
	entity    *gurps.Entity
	targetMgr *widget.TargetMgr
	prefix    string
}

// NewIdentityPanel creates a new identity panel.
func NewIdentityPanel(entity *gurps.Entity, targetMgr *widget.TargetMgr) *IdentityPanel {
	p := &IdentityPanel{
		entity:    entity,
		targetMgr: targetMgr,
		prefix:    targetMgr.NextPrefix(),
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: 4,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
	})
	p.SetBorder(unison.NewCompoundBorder(&widget.TitledBorder{Title: i18n.Text("Identity")}, unison.NewEmptyBorder(unison.Insets{
		Top:    1,
		Left:   2,
		Bottom: 1,
		Right:  2,
	})))
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, unison.ContentColor.Paint(gc, rect, unison.Fill))
	}

	title := i18n.Text("Name")
	field := widget.NewStringPageField(p.targetMgr, p.prefix+"name", title,
		func() string { return p.entity.Profile.Name },
		func(s string) { p.entity.Profile.Name = s })
	p.AddChild(widget.NewPageLabelWithRandomizer(title,
		i18n.Text("Randomize the name using the current ancestry"), func() {
			p.entity.Profile.Name = p.entity.Ancestry().RandomName(
				ancestry.AvailableNameGenerators(settings.Global().Libraries()), p.entity.Profile.Gender)
			SetTextAndMarkModified(field.Field, p.entity.Profile.Name)
		}))
	p.AddChild(field)

	title = i18n.Text("Title")
	p.AddChild(widget.NewPageLabelEnd(title))
	p.AddChild(widget.NewStringPageField(p.targetMgr, p.prefix+"title", title,
		func() string { return p.entity.Profile.Title },
		func(s string) { p.entity.Profile.Title = s }))

	title = i18n.Text("Organization")
	p.AddChild(widget.NewPageLabelEnd(title))
	p.AddChild(widget.NewStringPageField(p.targetMgr, p.prefix+"org", title,
		func() string { return p.entity.Profile.Organization },
		func(s string) { p.entity.Profile.Organization = s }))
	return p
}
