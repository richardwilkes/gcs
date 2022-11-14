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
	"strings"

	"github.com/richardwilkes/json"
)

// AttributeDifficulty holds an attribute ID and a difficulty.
type AttributeDifficulty struct {
	Attribute  string
	Difficulty Difficulty
	omit       bool
}

// Description returns a formatted description.
func (a *AttributeDifficulty) Description(entity *Entity) string {
	if a.Attribute == "" {
		return a.Difficulty.String()
	}
	return ResolveAttributeName(entity, a.Attribute) + "/" + a.Difficulty.String()
}

// Key returns the value that will be serialized.
func (a *AttributeDifficulty) Key() string {
	if a.Attribute == "" {
		return a.Difficulty.Key()
	}
	return a.Attribute + "/" + a.Difficulty.Key()
}

// ShouldOmit implements json.Omitter.
func (a AttributeDifficulty) ShouldOmit() bool {
	return a.omit
}

// MarshalJSON implements json.Marshaler.
func (a *AttributeDifficulty) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Key())
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *AttributeDifficulty) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parts := strings.SplitN(s, "/", 2)
	if len(parts) == 1 {
		s = parts[0]
		a.Attribute = ""
	} else {
		s = parts[1]
		a.Attribute = strings.TrimSpace(parts[0])
	}
	a.Difficulty = ExtractDifficulty(strings.TrimSpace(s))
	return nil
}

// Normalize the data. Should be called after loading from disk or the user.
func (a *AttributeDifficulty) Normalize(entity *Entity) {
	a.Difficulty = a.Difficulty.EnsureValid()
	text := strings.TrimSpace(a.Attribute)
	if text == "" {
		text = DefaultAttributeIDFor(entity)
	}
	var attr *AttributeDef
	list := AttributeDefsFor(entity).List(true)
	for _, one := range list {
		if strings.EqualFold(one.ID(), text) {
			attr = one
			break
		}
	}
	if attr == nil {
		for _, one := range list {
			if strings.EqualFold(one.Name, text) {
				attr = one
				break
			}
		}
	}
	if attr != nil {
		text = attr.ID()
	}
	a.Attribute = SanitizeID(text, true)
}
