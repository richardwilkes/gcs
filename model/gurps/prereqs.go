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

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
)

// Prereqs holds a list of prerequisites.
type Prereqs []Prereq

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (p *Prereqs) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	list, err := unmarshalTypedList(dec, prereq.ExtractKnownType, allocPrereq,
		func(kind string, raw jsontext.Value) Prereq { return NewUnknownPrereq(kind, raw) })
	if err != nil {
		return err
	}
	*p = list
	return nil
}

// allocPrereq returns an empty Prereq of the concrete type that represents prereqType, or nil if there is none.
func allocPrereq(prereqType prereq.Type) Prereq {
	switch prereqType {
	case prereq.List:
		return &PrereqList{}
	case prereq.Trait:
		return &TraitPrereq{}
	case prereq.Attribute:
		return &AttributePrereq{}
	case prereq.ContainedQuantity:
		return &ContainedQuantityPrereq{}
	case prereq.ContainedWeight:
		return &ContainedWeightPrereq{}
	case prereq.EquippedEquipment:
		return &EquippedEquipmentPrereq{}
	case prereq.Skill:
		return &SkillPrereq{}
	case prereq.Spell:
		return &SpellPrereq{}
	case prereq.Script:
		return &ScriptPrereq{}
	default:
		return nil
	}
}
