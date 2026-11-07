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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/toolbox/v2/i18n"
)

// AttributeFlags provides flags that can be set to extend the defined attribute choice list.
type AttributeFlags uint16

// Possible AttributeFlags.
const (
	BlankFlag AttributeFlags = 1 << iota
	TenFlag
	SizeFlag
	DodgeFlag
	ParryFlag
	BlockFlag
	SkillFlag
	// MoveFlag adds a choice standing in for Move when the definitions have no move attribute; see Entity.Move.
	MoveFlag
	// NoPoolsFlag leaves out the pools, which the bonuses of their thresholds can't alter.
	NoPoolsFlag
)

// AttributeChoice holds a single attribute choice.
type AttributeChoice struct {
	Key   string
	Title string
}

// AttributeChoices returns the attribute choices for the entity, which may be nil, along with the choice matching
// currentKey; see AttributeChoicesFor.
func AttributeChoices(entity *Entity, prefix string, flags AttributeFlags, currentKey string) (choices []*AttributeChoice, current *AttributeChoice) {
	return AttributeChoicesFor(AttributeDefsFor(entity), prefix, flags, currentKey)
}

// AttributeChoicesFor returns the attribute choices for the given definitions, along with the choice matching
// currentKey; if none matches, an "unrecognized key" choice for it is appended. A defined dodge attribute is listed
// only with DodgeFlag, and a stand-in for Move is added with MoveFlag only when no move attribute is defined.
func AttributeChoicesFor(defs *AttributeDefs, prefix string, flags AttributeFlags, currentKey string) (choices []*AttributeChoice, current *AttributeChoice) {
	if prefix != "" && !strings.HasSuffix(prefix, " ") {
		prefix += " "
	}
	list := defs.List(true)
	choices = make([]*AttributeChoice, 0, len(list)+8)
	if flags&BlankFlag != 0 {
		choices = append(choices, &AttributeChoice{})
	}
	if flags&TenFlag != 0 {
		choices = append(choices, &AttributeChoice{Key: "10", Title: prefix + "10"})
	}
	addedDodge := false
	addedMove := false
	for _, def := range list {
		switch {
		case def.DefID == DodgeID:
			if flags&DodgeFlag == 0 {
				continue
			}
			addedDodge = true
		case def.DefID == MoveID:
			addedMove = true
		case flags&NoPoolsFlag != 0 && (def.Type == attribute.Pool || def.Type == attribute.PoolRef):
			continue
		}
		choices = append(choices, &AttributeChoice{Key: def.DefID, Title: prefix + def.Name})
	}
	if flags&SizeFlag != 0 {
		choices = append(choices, &AttributeChoice{Key: SizeModifierID, Title: prefix + i18n.Text("Size Modifier")})
	}
	if flags&MoveFlag != 0 && !addedMove {
		choices = append(choices, &AttributeChoice{Key: MoveID, Title: prefix + i18n.Text("Move")})
	}
	if flags&DodgeFlag != 0 && !addedDodge {
		choices = append(choices, &AttributeChoice{Key: DodgeID, Title: prefix + i18n.Text("Dodge")})
	}
	if flags&ParryFlag != 0 {
		choices = append(choices, &AttributeChoice{Key: ParryID, Title: prefix + i18n.Text("Parry")})
	}
	if flags&BlockFlag != 0 {
		choices = append(choices, &AttributeChoice{Key: BlockID, Title: prefix + i18n.Text("Block")})
	}
	if flags&SkillFlag != 0 {
		choices = append(choices, &AttributeChoice{Key: SkillID, Title: prefix + i18n.Text("Skill")})
	}
	for _, choice := range choices {
		if choice.Key == currentKey {
			return choices, choice
		}
	}
	current = &AttributeChoice{
		Key:   currentKey,
		Title: prefix + i18n.Text("unrecognized key (%s)", currentKey),
	}
	return append(choices, current), current
}

func (c *AttributeChoice) String() string {
	return c.Title
}
