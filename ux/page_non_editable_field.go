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
	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// NonEditablePageField holds the data for a non-editable page field.
type NonEditablePageField struct {
	*unison.Label
	syncer func(*NonEditablePageField)
}

// NewNonEditablePageField creates a new start-aligned non-editable field using the page field font.
func NewNonEditablePageField(syncer func(*NonEditablePageField)) *NonEditablePageField {
	return newNonEditablePageField(syncer, align.Start)
}

// NewNonEditablePageFieldEnd creates a new end-aligned non-editable field using the page field font.
func NewNonEditablePageFieldEnd(syncer func(*NonEditablePageField)) *NonEditablePageField {
	return newNonEditablePageField(syncer, align.End)
}

// NewNonEditablePageFieldFor creates a new start-aligned non-editable field showing what text returns on each sync.
func NewNonEditablePageFieldFor(text func() string) *NonEditablePageField {
	return NewNonEditablePageField(syncTitleFrom(text))
}

// NewNonEditablePageFieldEndFor creates a new end-aligned non-editable field showing what text returns on each sync.
func NewNonEditablePageFieldEndFor(text func() string) *NonEditablePageField {
	return NewNonEditablePageFieldEnd(syncTitleFrom(text))
}

// NewNonEditablePageFieldCenter creates a new center-aligned non-editable field using the page field font.
func NewNonEditablePageFieldCenter(syncer func(*NonEditablePageField)) *NonEditablePageField {
	return newNonEditablePageField(syncer, align.Middle)
}

func newNonEditablePageField(syncer func(*NonEditablePageField), hAlign align.Enum) *NonEditablePageField {
	f := &NonEditablePageField{
		Label:  unison.NewLabel(),
		syncer: syncer,
	}
	f.Self = f
	f.Font = fonts.PageFieldPrimary
	f.HAlign = hAlign
	f.SetBorder(unison.NewEmptyBorder(geom.Insets{
		Left:   1,
		Bottom: 1,
		Right:  1,
	})) // Match normal fields
	f.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	f.Sync()
	return f
}

func syncTitleFrom(text func() string) func(*NonEditablePageField) {
	return func(f *NonEditablePageField) { f.SetTitleIfChanged(text()) }
}

// Sync the field to the current value.
func (f *NonEditablePageField) Sync() {
	f.syncer(f)
}

// SetTitleIfChanged sets the field's text and marks it for layout within its dockable when the text differs, returning
// true if it did.
func (f *NonEditablePageField) SetTitleIfChanged(text string) bool {
	if text == f.Text.String() {
		return false
	}
	f.SetTitle(text)
	MarkForLayoutWithinDockable(f)
	return true
}
