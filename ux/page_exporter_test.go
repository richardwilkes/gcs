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
	"path/filepath"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

// newTestTemplateDockable returns a template dockable for the given file path. Building the toolbar reaches for the
// bindable actions, so they are registered first.
func newTestTemplateDockable(fileName string, data *gurps.Template) *Template {
	registerKeyBindingsOnce.Do(func() { registerActions() })
	return NewTemplate(filepath.Join("some", "dir", fileName+gurps.TemplatesExt), data)
}

// TestPageInfoProviderForTemplate verifies that obtaining a template dockable's page info provider fills in the page
// title and clears the modification timestamp. A gurps.Template returns only these explicitly-set values, so a page
// building path that skipped this setup — printing used to — produced footers with a blank title, or with a stale
// title left behind by an earlier export in the same session.
func TestPageInfoProviderForTemplate(t *testing.T) {
	c := check.New(t)

	t.Run("the title comes from the dockable", func(_ *testing.T) {
		data := gurps.NewTemplate()
		dockable := newTestTemplateDockable("My Template", data)
		c.Equal("", data.PageTitle(), "a template has no page title until one is supplied")

		provider := pageInfoProviderFor(dockable)
		c.Equal(data, provider, "the template itself is the page info provider")
		c.Equal("My Template", provider.PageTitle(), "the page title is the dockable's title")
		c.Equal("", provider.ModifiedOnString(), "a template has no modification timestamp to show")
	})

	t.Run("stale values from an earlier export are replaced", func(_ *testing.T) {
		data := gurps.NewTemplate()
		dockable := newTestTemplateDockable("My Template", data)
		data.ExplicitPageTitle = "Some Other Template"
		data.ExplicitModifiedOn = "Jan 1, 1970"

		provider := pageInfoProviderFor(dockable)
		c.Equal("My Template", provider.PageTitle(), "the stale page title is replaced")
		c.Equal("", provider.ModifiedOnString(), "the stale modification timestamp is cleared")
	})

	t.Run("renaming the dockable's file is picked up", func(_ *testing.T) {
		data := gurps.NewTemplate()
		dockable := newTestTemplateDockable("My Template", data)
		c.Equal("My Template", pageInfoProviderFor(dockable).PageTitle())

		dockable.path = filepath.Join("elsewhere", "Renamed"+gurps.TemplatesExt)
		c.Equal("Renamed", pageInfoProviderFor(dockable).PageTitle(), "the current title is used, not the first one")
	})
}

// TestPageInfoProviderForSheet verifies that a character sheet's page info provider is handed back as-is, with its own
// title intact, since only templates need values supplied by their dockable.
func TestPageInfoProviderForSheet(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	entity.Profile.Name = "Dai Blackthorn"

	provider := pageInfoProviderFor(sheet)
	c.Equal(entity, provider, "the entity itself is the page info provider")
	c.Equal("Dai Blackthorn", provider.PageTitle(), "the entity supplies its own page title")
}
