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
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// NonEditableField holds the data for a non-editable field.
type NonEditableField struct {
	*unison.Label
	syncer func(*NonEditableField)
}

// NewNonEditableField creates a new start-aligned non-editable field that uses the same font and size as the field.
func NewNonEditableField(syncer func(*NonEditableField)) *NonEditableField {
	return newNonEditableField(syncer, align.Start)
}

func newNonEditableField(syncer func(*NonEditableField), hAlign align.Enum) *NonEditableField {
	f := &NonEditableField{
		Label:  unison.NewLabel(),
		syncer: syncer,
	}
	f.Self = f
	f.Font = unison.FieldFont
	f.HAlign = hAlign
	f.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, 0,
		unison.NewUniformInsets(1), false), unison.NewEmptyBorder(unison.Insets{Top: 3, Left: 3, Bottom: 2, Right: 3})))
	f.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	f.Sync()
	return f
}

// Sync the field to the current value.
func (f *NonEditableField) Sync() {
	f.syncer(f)
}
