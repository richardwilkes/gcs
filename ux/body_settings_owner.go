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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

var _ BodySettingsOwner = &globalBodySettingsOwner{}

// BodySettingsOwner is the interface that a panel must implement to be able to display and edit Body settings.
type BodySettingsOwner interface {
	EntityPanel
	BodySettingsTitle() string
	BodySettings(forReset bool) *gurps.Body
	SetBodySettings(body *gurps.Body)
}

type globalBodySettingsOwner struct{}

func (g *globalBodySettingsOwner) AsPanel() *unison.Panel {
	return nil
}

func (g *globalBodySettingsOwner) Entity() *gurps.Entity {
	return nil
}

func (g *globalBodySettingsOwner) BodySettingsTitle() string {
	return i18n.Text("Default Body Type")
}

func (g *globalBodySettingsOwner) BodySettings(forReset bool) *gurps.Body {
	if forReset {
		return gurps.FactoryBody()
	}
	return gurps.GlobalSettings().Sheet.BodyType
}

func (g *globalBodySettingsOwner) SetBodySettings(body *gurps.Body) {
	gurps.GlobalSettings().Sheet.BodyType = body
}
