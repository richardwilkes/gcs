// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
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

	"github.com/richardwilkes/toolbox/i18n"
)

// HitLocationChoice holds a single hit location choice.
type HitLocationChoice struct {
	Key   string
	Title string
}

// HitLocationChoices collects the available choices for hit locations for the given entity.
func HitLocationChoices(entity *Entity, currentKey string, forEquipmentModifier bool) (choices []*HitLocationChoice, current *HitLocationChoice) {
	bodyType := BodyFor(entity)
	list := bodyType.UniqueHitLocations(entity)
	choices = make([]*HitLocationChoice, 0, len(list)+3)
	if forEquipmentModifier {
		choices = append(choices, &HitLocationChoice{Title: i18n.Text("to this armor")})
	}
	choices = append(choices, &HitLocationChoice{Key: AllID, Title: i18n.Text("to all locations")})
	prefix := i18n.Text("to the")
	for _, one := range list {
		choices = append(choices, &HitLocationChoice{Key: one.LocID, Title: prefix + " " + one.ChoiceName})
	}
	for _, choice := range choices {
		if choice.Key == currentKey {
			return choices, choice
		}
	}
	current = &HitLocationChoice{
		Key:   currentKey,
		Title: fmt.Sprintf(prefix+i18n.Text("unrecognized key (%s) for body type (%s)"), currentKey, bodyType.Name),
	}
	return append(choices, current), current
}

func (c *HitLocationChoice) String() string {
	return c.Title
}
