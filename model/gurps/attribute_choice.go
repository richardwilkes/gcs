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
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// AttributeFlags provides flags that can be set to extend the defined attribute choice list.
type AttributeFlags byte

// Possible AttributeFlags.
const (
	BlankFlag AttributeFlags = 1 << iota
	TenFlag
	SizeFlag
	DodgeFlag
	ParryFlag
	BlockFlag
	SkillFlag
)

// AttributeChoice holds a single attribute choice.
type AttributeChoice struct {
	Key   string
	Title string
}

// AttributeChoices collects the available choices for attributes for the given entity, or nil.
func AttributeChoices(entity *Entity, prefix string, flags AttributeFlags, currentKey string) (choices []*AttributeChoice, current *AttributeChoice) {
	if prefix != "" && !strings.HasSuffix(prefix, " ") {
		prefix += " "
	}
	list := AttributeDefsFor(entity).List(true)
	choices = make([]*AttributeChoice, 0, len(list)+8)
	if flags&BlankFlag != 0 {
		choices = append(choices, &AttributeChoice{})
	}
	if flags&TenFlag != 0 {
		choices = append(choices, &AttributeChoice{Key: "10", Title: prefix + "10"})
	}
	addedDodge := false
	for _, def := range list {
		if def.DefID == DodgeID {
			if flags&DodgeFlag == 0 {
				continue
			}
			addedDodge = true
		}
		choices = append(choices, &AttributeChoice{Key: def.DefID, Title: prefix + def.Name})
	}
	if flags&SizeFlag != 0 {
		choices = append(choices, &AttributeChoice{Key: SizeModifierID, Title: prefix + i18n.Text("Size Modifier")})
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
		Title: fmt.Sprintf(prefix+i18n.Text("unrecognized key (%s)"), currentKey),
	}
	return append(choices, current), current
}

func (c *AttributeChoice) String() string {
	return c.Title
}
