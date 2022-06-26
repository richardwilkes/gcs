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

package widget

import (
	"github.com/richardwilkes/unison"
)

// NewFieldLeadingLabel creates a new label appropriate for the first label in a row before a field.
func NewFieldLeadingLabel(text string) *unison.Label {
	label := unison.NewLabel()
	label.Text = text
	label.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.EndAlignment,
		VAlign: unison.MiddleAlignment,
	})
	return label
}

// NewFieldInteriorLeadingLabel creates a new label appropriate for the label in the interior of a row before a field.
func NewFieldInteriorLeadingLabel(text string) *unison.Label {
	label := unison.NewLabel()
	label.Text = text
	label.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.EndAlignment,
		VAlign: unison.MiddleAlignment,
	})
	label.SetBorder(unison.NewEmptyBorder(unison.Insets{Left: unison.StdHSpacing}))
	return label
}

// NewFieldTrailingLabel creates a new label appropriate for after a field.
func NewFieldTrailingLabel(text string) *unison.Label {
	label := unison.NewLabel()
	label.Text = text
	label.SetLayoutData(&unison.FlexLayoutData{
		VAlign: unison.MiddleAlignment,
	})
	return label
}
