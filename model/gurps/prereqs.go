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
	"encoding/json/jsontext"
	"encoding/json/v2"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Prereqs holds a list of prerequisites.
type Prereqs []Prereq

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (p *Prereqs) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var v []jsontext.Value
	if err := json.UnmarshalDecode(dec, &v); err != nil {
		return errs.Wrap(err)
	}
	*p = make([]Prereq, len(v))
	for i, one := range v {
		var typeData struct {
			Type prereq.Type `json:"type"`
		}
		if err := json.Unmarshal(one, &typeData); err != nil {
			return errs.Wrap(err)
		}
		var pr Prereq
		switch typeData.Type {
		case prereq.List:
			pr = &PrereqList{}
		case prereq.Trait:
			pr = &TraitPrereq{}
		case prereq.Attribute:
			pr = &AttributePrereq{}
		case prereq.ContainedQuantity:
			pr = &ContainedQuantityPrereq{}
		case prereq.ContainedWeight:
			pr = &ContainedWeightPrereq{}
		case prereq.EquippedEquipment:
			pr = &EquippedEquipmentPrereq{}
		case prereq.Skill:
			pr = &SkillPrereq{}
		case prereq.Spell:
			pr = &SpellPrereq{}
		case prereq.Script:
			pr = &ScriptPrereq{}
		default:
			return errs.Newf(i18n.Text("Unknown prerequisite type: %s"), typeData.Type)
		}
		if err := json.Unmarshal(one, &pr); err != nil {
			return errs.Wrap(err)
		}
		(*p)[i] = pr
	}
	return nil
}
