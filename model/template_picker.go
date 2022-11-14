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

package model

import (
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/i18n"
)

var _ json.Omitter = &TemplatePicker{}

// TemplatePickerProvider defines the methods a TemplatePicker provider has.
type TemplatePickerProvider interface {
	TemplatePickerData() *TemplatePicker
}

// TemplatePicker holds the data necessary to allow a template choice to be made.
type TemplatePicker struct {
	Type      TemplatePickerType `json:"type"`
	Qualifier NumericCriteria    `json:"qualifier"`
}

// Clone creates a copy of the TemplatePicker.
func (t *TemplatePicker) Clone() *TemplatePicker {
	if t.ShouldOmit() {
		return &TemplatePicker{}
	}
	picker := *t
	return &picker
}

// ShouldOmit implements json.Omitter.
func (t *TemplatePicker) ShouldOmit() bool {
	return t == nil || t.Type == NotApplicableTemplatePickerType
}

// Description returns a description of the picker action.
func (t *TemplatePicker) Description() string {
	if t.ShouldOmit() {
		return ""
	}
	switch t.Type {
	case CountTemplatePickerType:
		return fmt.Sprintf(i18n.Text("Pick %s"), t.Qualifier.AltString())
	case PointsTemplatePickerType:
		points := i18n.Text("points")
		if t.Qualifier.Qualifier == fxp.One {
			points = i18n.Text("point")
		}
		return fmt.Sprintf(i18n.Text("Pick %s %s worth"), t.Qualifier.AltString(), points)
	default:
		return ""
	}
}
