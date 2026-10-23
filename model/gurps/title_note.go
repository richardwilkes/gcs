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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Feature = &TitleNote{}

// TitleNote holds text that is shown in parentheses after the name of the item it is attached to, the way a skill's
// specialization is. A title note on a modifier is shown after the name of the item that owns the modifier. A title
// note describes what the item is rather than its state during play, so unlike other features it can't be switchable.
type TitleNote struct {
	Type feature.Type `json:"type"`
	Text string       `json:"text,omitzero"`
	// BeforeBuiltIn places the note ahead of the item's built-in parts, such as a skill's specialization.
	BeforeBuiltIn bool `json:"before_built_in,omitzero"`
}

// NewTitleNote creates a new TitleNote.
func NewTitleNote() *TitleNote {
	return &TitleNote{Type: feature.TitleNote}
}

// FeatureType implements Feature.
func (n *TitleNote) FeatureType() feature.Type {
	return n.Type
}

// Clone implements Feature.
func (n *TitleNote) Clone() Feature {
	return clonePtr(n)
}

// IsSwitchable implements Feature. A title note is never switchable.
func (n *TitleNote) IsSwitchable() bool {
	return false
}

// SetSwitchable implements Feature. A title note is never switchable, so this does nothing.
func (n *TitleNote) SetSwitchable(_ bool) {
}

// FillWithNameableKeys implements Feature.
func (n *TitleNote) FillWithNameableKeys(m, existing map[string]string) {
	nameable.Extract(m, existing, n.Text)
}

// Describe implements Feature.
func (n *TitleNote) Describe(_ *Entity, replacements map[string]string, em func(string) string) string {
	text := strings.TrimSpace(nameable.Apply(n.Text, replacements))
	if text == "" {
		return i18n.Text("Adds an empty title note")
	}
	if n.BeforeBuiltIn {
		return i18n.Text("Adds the title note %s before the built-in notes", em(text))
	}
	return i18n.Text("Adds the title note %s", em(text))
}

// Hash writes this object's contents into the hasher.
func (n *TitleNote) Hash(h hash.Hash) {
	if n == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, n.Type)
	xhash.StringWithLen(h, n.Text)
	xhash.Bool(h, n.BeforeBuiltIn)
}

// appendTitleNotes appends the non-empty text of the title note features whose BeforeBuiltIn is before to parts, with
// the replacements applied.
func appendTitleNotes(parts []string, replacements map[string]string, before bool, features Features) []string {
	for _, f := range features {
		if note, ok := f.(*TitleNote); ok && note.BeforeBuiltIn == before {
			if text := strings.TrimSpace(nameable.Apply(note.Text, replacements)); text != "" {
				parts = append(parts, text)
			}
		}
	}
	return parts
}

// modifierTitleNotes appends the title notes of the enabled, non-container modifiers to parts, in modifier order: for
// each, its CompactName (with a leveled trait modifier's level, which title note criteria therefore see too) when it
// is shown in the title and before is false, followed by the text of its title note features whose BeforeBuiltIn is
// before. It walks the modifiers itself rather than through Traverse, which allocates, since
// every item's name passes through here.
func modifierTitleNotes[M ModifierNode[M, T], T ModifiableNode[T, M], S ~[]M](
	parts []string,
	before bool,
	modifiers S,
	replacements map[string]string,
	features func(M) Features,
) []string {
	for _, mod := range modifiers {
		switch {
		case !mod.Enabled():
		case mod.Container():
			parts = modifierTitleNotes(parts, before, mod.NodeChildren(), replacements, features)
		default:
			if !before && mod.ShowsInTitle() {
				if name := strings.TrimSpace(mod.CompactName()); name != "" {
					parts = append(parts, name)
				}
			}
			parts = appendTitleNotes(parts, replacements, before, features(mod))
		}
	}
	return parts
}

// parenthetical returns the parts, separated by "; " and wrapped in parentheses, preceded by a space, or "" when there
// are none.
func parenthetical(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	return " (" + strings.Join(parts, "; ") + ")"
}

// describeTitleNote returns the clause a description adds for its title note criteria, such as ` with the title note
// Pistol`, or ` whose owner has the title note Pistol` when the notes are the owner's, or nothing when any title note
// will do. A non-empty qualifier is passed through em.
func describeTitleNote(t criteria.Text, owner bool, replacements map[string]string, em func(string) string) string {
	none, some, is := i18n.Text(" without title notes"), i18n.Text(" with at least one title note"),
		i18n.Text(" with the title note ")
	prefix, notPrefix := i18n.Text("with a title note that"), i18n.Text("with all title notes that")
	if owner {
		none, some, is = i18n.Text(" whose owner has no title notes"),
			i18n.Text(" whose owner has at least one title note"), i18n.Text(" whose owner has the title note ")
		prefix, notPrefix = i18n.Text("whose owner has a title note that"), i18n.Text("all of whose owner's title notes")
	}
	q := t.Compare.EffectiveQualifier(nameable.Apply(t.Qualifier, replacements))
	switch {
	case t.Compare == criteria.AnyText:
		return ""
	case t.Compare == criteria.IsText && q == "":
		return none
	case t.Compare == criteria.IsNotText && q == "":
		return some
	case t.Compare == criteria.IsText:
		return is + em(q)
	}
	if q != "" {
		q = em(q)
	}
	return " " + t.Compare.DescribeWithPrefix(prefix, notPrefix, q)
}
