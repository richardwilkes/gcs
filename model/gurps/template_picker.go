// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

// TemplatePickerProvider provides the method used to get a list of valid picker type and to access the picker data
type TemplatePickerProvider interface {
	// TemplatePickerData returns a list of valid picker types and valid, non-nil, pointer to a TemplatePicker struct
	TemplatePickerData() ([]picker.Type, *TemplatePicker)
}

// TemplatePicker holds the data necessary to allow a template choice to be made.
type TemplatePicker struct {
	Type      picker.Type     `json:"type"`
	Qualifier criteria.Number `json:"qualifier,omitzero"`
}

// IsZero implements json.isZero.
func (t TemplatePicker) IsZero() bool {
	return t.Type == picker.NotApplicable
}

func (t TemplatePicker) String() string {
	if t.IsZero() {
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
	case picker.Quantity:
		return fmt.Sprintf(i18n.Text("Pick %s in quantity"), t.Qualifier.AltString())
	case picker.Cost:
		return fmt.Sprintf(i18n.Text("Pick %s worth of cost"), t.Qualifier.AltString())
	case picker.Weight:
		return fmt.Sprintf(i18n.Text("Pick %s worth of weight"), t.Qualifier.AltString())
	default:
		return ""
	}
}

// Hash writes this object's contents into the hasher.
func (t TemplatePicker) Hash(h hash.Hash) {
	xhash.Num8(h, t.Type)
	t.Qualifier.Hash(h)
}
