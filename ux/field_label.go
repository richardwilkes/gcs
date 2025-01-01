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

// NewFieldLeadingLabel creates a new label appropriate for the first label in a row before a field.
func NewFieldLeadingLabel(text string, small bool) *unison.Label {
	label := unison.NewLabel()
	adjustForSmall(label, small)
	label.SetTitle(text)
	label.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.End,
		VAlign: align.Middle,
	})
	return label
}

// NewFieldInteriorLeadingLabel creates a new label appropriate for the label in the interior of a row before a field.
func NewFieldInteriorLeadingLabel(text string, small bool) *unison.Label {
	label := unison.NewLabel()
	adjustForSmall(label, small)
	label.SetTitle(text)
	label.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.End,
		VAlign: align.Middle,
	})
	label.SetBorder(unison.NewEmptyBorder(unison.Insets{Left: unison.StdHSpacing}))
	return label
}

// NewFieldTrailingLabel creates a new label appropriate for after a field.
func NewFieldTrailingLabel(text string, small bool) *unison.Label {
	label := unison.NewLabel()
	adjustForSmall(label, small)
	label.SetTitle(text)
	label.SetLayoutData(&unison.FlexLayoutData{
		VAlign: align.Middle,
	})
	return label
}

func adjustForSmall(label *unison.Label, small bool) {
	if small {
		fd := label.Font.Descriptor()
		fd.Size *= 0.8
		label.Font = fd.Font()
	}
}
