// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"fmt"
	"hash"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
)

var _ json.Omitter = &TemplatePicker{}

// TemplatePickerProvider defines the methods a TemplatePicker provider has.
type TemplatePickerProvider interface {
	TemplatePickerData() *TemplatePicker
}

// TemplatePicker holds the data necessary to allow a template choice to be made.
type TemplatePicker struct {
	Type      picker.Type     `json:"type"`
	Qualifier criteria.Number `json:"qualifier"`
}

// Clone creates a copy of the TemplatePicker.
func (t *TemplatePicker) Clone() *TemplatePicker {
	if t.ShouldOmit() {
		return &TemplatePicker{}
	}
	p := *t
	return &p
}

// ShouldOmit implements json.Omitter.
func (t *TemplatePicker) ShouldOmit() bool {
	return t == nil || t.Type == picker.NotApplicable
}

// Description returns a description of the picker action.
func (t *TemplatePicker) Description() string {
	if t.ShouldOmit() {
		return ""
	}
	switch t.Type {
	case picker.Count:
		return fmt.Sprintf(i18n.Text("Pick %s"), t.Qualifier.AltString())
	case picker.Points:
		points := i18n.Text("points")
		if t.Qualifier.Qualifier == fxp.One {
			points = i18n.Text("point")
		}
		return fmt.Sprintf(i18n.Text("Pick %s %s worth"), t.Qualifier.AltString(), points)
	default:
		return ""
	}
}

// Hash writes this object's contents into the hasher.
func (t *TemplatePicker) Hash(h hash.Hash) {
	if t.ShouldOmit() {
		hashhelper.Num8(h, uint8(255))
		return
	}
	hashhelper.Num8(h, t.Type)
	t.Qualifier.Hash(h)
}
