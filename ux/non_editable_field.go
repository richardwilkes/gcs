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
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/role"
)

// NonEditableField holds the data for a non-editable field.
type NonEditableField struct {
	*unison.Label
	syncer func(*NonEditableField)
}

// NewNonEditableField creates a new start-aligned non-editable field that uses the standard field font.
func NewNonEditableField(syncer func(*NonEditableField)) *NonEditableField {
	return newNonEditableField(syncer, align.Start)
}

func newNonEditableField(syncer func(*NonEditableField), hAlign align.Enum) *NonEditableField {
	f := &NonEditableField{}
	f.init(f, syncer, hAlign)
	return f
}

// init sets the field up around a new label, with self as what the field's panel resolves to, so that a widget built
// by embedding the field can point it at itself.
func (f *NonEditableField) init(self unison.Paneler, syncer func(*NonEditableField), hAlign align.Enum) {
	f.Label = unison.NewLabel()
	f.syncer = syncer
	f.Self = self
	// Self is the field, not a *Label, so without an explicit role unison would not take it for static text, and
	// unison.SetFocusForReading would not make it a tab stop.
	f.Accessibility.Role = role.Label
	f.Font = unison.FieldFont
	f.HAlign = hAlign
	f.SetBorder(nonEditableFieldBorder(unison.ThemeSurfaceEdge))
	f.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	f.Sync()
}

// nonEditableFieldBorder returns the border of a non-editable field, a line in the ink within the same insets as the
// border an editable field has while it is not focused, so that the two line up in a row.
func nonEditableFieldBorder(ink unison.Ink) unison.Border {
	return unison.NewCompoundBorder(unison.NewLineBorder(ink, geom.Size{}, geom.NewUniformInsets(1), false),
		unison.NewEmptyBorder(geom.Insets{Top: 3, Left: 3, Bottom: 2, Right: 3}))
}

// Sync the field to the current value.
func (f *NonEditableField) Sync() {
	f.syncer(f)
}

// ProvideAccessibility implements unison.AccessibilityProvider. It hides the field while it is empty and unnamed, which
// unison does for a label only when it derives the label's role, and this field's role is explicit.
func (f *NonEditableField) ProvideAccessibility(b *unison.AccessibilityBuilder) {
	f.Label.ProvideAccessibility(b)
	ignoreEmptyStaticText(b, f.Text.String())
}
