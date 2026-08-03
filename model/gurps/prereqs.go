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
	"encoding/json/jsontext"
	"encoding/json/v2"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/toolbox/v2/errs"
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
			Type string `json:"type"`
		}
		if err := json.Unmarshal(one, &typeData); err != nil {
			return errs.Wrap(err)
		}
		// Note that the type is extracted as a string and resolved with ExtractKnownType rather than being unmarshaled
		// directly into a prereq.Type: the enum's UnmarshalText maps anything it doesn't recognize onto the first
		// value, which would turn a prerequisite written by a newer version of GCS into an empty, always-satisfied
		// PrereqList.
		prereqType, known := prereq.ExtractKnownType(typeData.Type)
		if !known {
			(*p)[i] = NewUnknownPrereq(typeData.Type, one)
			continue
		}
		var pr Prereq
		switch prereqType {
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
			// A known type that has no case above, or the Unknown type itself. Preserve rather than discard.
			(*p)[i] = NewUnknownPrereq(typeData.Type, one)
			continue
		}
		if err := json.Unmarshal(one, &pr); err != nil {
			return errs.Wrap(err)
		}
		(*p)[i] = pr
	}
	return nil
}
