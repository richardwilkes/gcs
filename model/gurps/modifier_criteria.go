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
	"hash"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
)

// ModifierCriteria holds what a prerequisite asks of the modifiers of the item it matches. They are met by a single
// enabled modifier, containers left out, that meets them all.
type ModifierCriteria struct {
	ModifierNameCriteria  criteria.Text `json:"modifier_name,omitzero"`
	ModifierNotesCriteria criteria.Text `json:"modifier_notes,omitzero"`
	ModifierTagsCriteria  criteria.Text `json:"modifier_tags,omitzero"`
}

// criteriaModifier is what ModifierCriteria checks of a modifier.
type criteriaModifier[M Node[M]] interface {
	Node[M]
	NameWithReplacements() string
	LocalNotesWithReplacements() string
	TagList() []string
}

// hasModifierCriteria returns true if any of the modifier criteria asks for something.
func (c *ModifierCriteria) hasModifierCriteria() bool {
	return c.ModifierNameCriteria.Compare != criteria.AnyText || c.ModifierNotesCriteria.Compare != criteria.AnyText ||
		c.ModifierTagsCriteria.Compare != criteria.AnyText
}

// modifiersMeet returns true if there are no modifier criteria, or one of the enabled modifiers, containers left out,
// meets them all.
func modifiersMeet[M criteriaModifier[M]](c *ModifierCriteria, replacements map[string]string, modifiers []M) bool {
	if !c.hasModifierCriteria() {
		return true
	}
	found := false
	Traverse(func(mod M) bool {
		found = c.ModifierNameCriteria.Matches(replacements, mod.NameWithReplacements()) &&
			c.ModifierNotesCriteria.Matches(replacements, mod.LocalNotesWithReplacements()) &&
			c.ModifierTagsCriteria.MatchesList(replacements, mod.TagList()...)
		return found
	}, true, true, modifiers...)
	return found
}

// describeModifier returns the clause that describes the modifier criteria, such as " with a modifier whose name is
// Fine", or an empty string when they ask for nothing.
func (c *ModifierCriteria) describeModifier(replacements map[string]string, em func(string) string) string {
	if !c.hasModifierCriteria() {
		return ""
	}
	var parts []string
	for _, part := range []string{
		describeWhose(i18n.Text("whose name"), c.ModifierNameCriteria, replacements, em),
		describeWhose(i18n.Text("whose notes"), c.ModifierNotesCriteria, replacements, em),
		describeTags(c.ModifierTagsCriteria, replacements, em),
	} {
		if part != "" {
			parts = append(parts, strings.TrimSpace(part))
		}
	}
	return i18n.Text(" with a modifier ") + strings.Join(parts, i18n.Text(" and "))
}

// fillWithModifierNameableKeys adds the nameable keys of the modifier criteria to m.
func (c *ModifierCriteria) fillWithModifierNameableKeys(m, existing map[string]string) {
	nameable.Extract(m, existing, c.ModifierNameCriteria.Qualifier, c.ModifierNotesCriteria.Qualifier,
		c.ModifierTagsCriteria.Qualifier)
}

// hashModifier writes the modifier criteria into the hasher.
func (c *ModifierCriteria) hashModifier(h hash.Hash) {
	c.ModifierNameCriteria.Hash(h)
	c.ModifierNotesCriteria.Hash(h)
	c.ModifierTagsCriteria.Hash(h)
}
