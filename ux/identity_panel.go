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
	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

const (
	identityPanelFieldPrefix     = "identity:"
	identityPanelNameFieldRefKey = identityPanelFieldPrefix + "name"
)

// IdentityPanel holds the contents of the identity block on the sheet.
type IdentityPanel struct {
	unison.Panel
	entity    *gurps.Entity
	targetMgr *TargetMgr
	prefix    string
}

// NewIdentityPanel creates a new identity panel.
func NewIdentityPanel(entity *gurps.Entity, targetMgr *TargetMgr) *IdentityPanel {
	p := &IdentityPanel{
		entity:    entity,
		targetMgr: targetMgr,
		prefix:    identityPanelFieldPrefix,
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: 4,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
	})
	p.SetBorder(unison.NewCompoundBorder(&TitledBorder{Title: i18n.Text("Identity")}, unison.NewEmptyBorder(unison.Insets{
		Top:    1,
		Left:   2,
		Bottom: 1,
		Right:  2,
	})))
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) { drawBandedBackground(p, gc, rect, 0, 2, nil) }
	InstallTintFunc(p, colors.TintIdentity)

	title := i18n.Text("Name")
	nameField := NewStringPageField(p.targetMgr, identityPanelNameFieldRefKey, title,
		func() string { return p.entity.Profile.Name },
		func(s string) { p.entity.Profile.Name = s })
	p.AddChild(NewPageLabelWithRandomizer(title,
		i18n.Text("Randomize the name using the current ancestry"), func() {
			p.entity.Profile.Name = p.entity.Ancestry().RandomName(
				gurps.AvailableNameGenerators(gurps.GlobalSettings().Libraries()), p.entity.Profile.Gender)
			SetTextAndMarkModified(nameField.Field, p.entity.Profile.Name)
		}))
	nameField.ClientData()[SkipDeepSync] = true
	p.AddChild(nameField)

	title = i18n.Text("Title")
	p.AddChild(NewPageLabelEnd(title))
	titleField := NewStringPageField(p.targetMgr, p.prefix+"title", title,
		func() string { return p.entity.Profile.Title },
		func(s string) { p.entity.Profile.Title = s })
	titleField.ClientData()[SkipDeepSync] = true
	p.AddChild(titleField)

	title = i18n.Text("Organization")
	p.AddChild(NewPageLabelEnd(title))
	orgField := NewStringPageField(p.targetMgr, p.prefix+"org", title,
		func() string { return p.entity.Profile.Organization },
		func(s string) { p.entity.Profile.Organization = s })
	orgField.ClientData()[SkipDeepSync] = true
	p.AddChild(orgField)
	return p
}
