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
	"io/fs"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

var _ GroupedCloser = &bodySettingsDockable{}

type bodySettingsDockable struct {
	undoableSettingsDockable[*gurps.Body]
	owner BodySettingsOwner
}

// ShowBodySettings shows the Body Settings. Pass in globalBodySettings to edit the defaults or a sheet to edit the
// sheet's.
func ShowBodySettings(owner BodySettingsOwner) {
	if Activate(func(d unison.Dockable) bool { return isBodySettingsFor(d, owner) }) {
		return
	}
	d := newBodySettingsDockable(owner)
	d.TabTitle = owner.BodySettingsTitle()
	d.TabIcon = svg.BodyType
	d.Extensions = []string{gurps.BodyExt, gurps.BodyExtAlt}
	d.Loader = d.load
	d.Resetter = d.reset
	d.show()
}

// newBodySettingsDockable returns a dockable holding a copy of the owner's body type, ready for its content to be built.
func newBodySettingsDockable(owner BodySettingsOwner) *bodySettingsDockable {
	d := &bodySettingsDockable{owner: owner}
	d.Self = d
	d.init(d, undoableSettingsSpec[*gurps.Body]{
		helpLink:     "md:User%20Guide/Body%20Type",
		clone:        d.cloneBody,
		buildContent: func() { d.content.AddChild(newBodySettingsPanel(d)) },
		apply:        d.apply,
	}, d.cloneBody(owner.BodySettings(false)))
	return d
}

// isBodySettingsFor returns true if the dockable is the body settings dockable belonging to the given owner. Owners are
// compared by identity, so the caller must always pass the same owner for a given set of settings -- see the comment on
// globalBodySettings for the defaults.
func isBodySettingsFor(d unison.Dockable, owner BodySettingsOwner) bool {
	s, ok := d.AsPanel().Self.(*bodySettingsDockable)
	return ok && s.owner == owner
}

func (d *bodySettingsDockable) CloseWithGroup(other unison.Paneler) bool {
	return d.owner == other
}

func (d *bodySettingsDockable) Entity() *gurps.Entity {
	return d.owner.Entity()
}

// cloneBody returns a deep copy of the body type, as a top-level table for this dockable's entity.
func (d *bodySettingsDockable) cloneBody(body *gurps.Body) *gurps.Body {
	return body.Clone(d.Entity(), nil)
}

// moveHitLocation moves the location to the given insertion position within the table that owns it and reports whether
// anything changed. It is what dropping a dragged hit location does.
func (d *bodySettingsDockable) moveHitLocation(loc *gurps.HitLocation, to int) bool {
	table := loc.OwningTable()
	if !moveEntry(&table.Locations, slices.Index(table.Locations, loc), to) {
		return false
	}
	table.Update(d.Entity())
	return true
}

func (d *bodySettingsDockable) reset() {
	d.replaceModel(i18n.Text("Reset Body Type"), d.cloneBody(d.owner.BodySettings(true)))
}

func (d *bodySettingsDockable) load(fileSystem fs.FS, filePath string) error {
	bodyType, err := gurps.NewBodyFromFile(fileSystem, filePath)
	if err != nil {
		return err
	}
	d.replaceModel(i18n.Text("Load Body Type"), bodyType)
	return nil
}

func (d *bodySettingsDockable) apply() {
	d.Window().FocusNext() // Intentionally move the focus to ensure any pending edits are flushed
	d.owner.SetBodySettings(d.cloneBody(d.model))
}
