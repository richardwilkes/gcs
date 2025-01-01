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
	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// NonEditablePageField holds the data for a non-editable page field.
type NonEditablePageField struct {
	*unison.Label
	syncer func(*NonEditablePageField)
}

// NewNonEditablePageField creates a new start-aligned non-editable field that uses the same font and size as the page
// field.
func NewNonEditablePageField(syncer func(*NonEditablePageField)) *NonEditablePageField {
	return newNonEditablePageField(syncer, align.Start)
}

// NewNonEditablePageFieldEnd creates a new end-aligned non-editable field that uses the same font and size as the page
// field.
func NewNonEditablePageFieldEnd(syncer func(*NonEditablePageField)) *NonEditablePageField {
	return newNonEditablePageField(syncer, align.End)
}

// NewNonEditablePageFieldCenter creates a new center-aligned non-editable field that uses the same font and size as the
// page field.
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
	f.SetBorder(unison.NewEmptyBorder(unison.Insets{
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

// Sync the field to the current value.
func (f *NonEditablePageField) Sync() {
	f.syncer(f)
}
