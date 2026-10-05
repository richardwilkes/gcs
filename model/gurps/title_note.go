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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Feature = &TitleNote{}

// TitleNote holds text that is shown in parentheses after the name of the item it is attached to, the way a skill's
// specialization is. A title note on a modifier is shown after the name of the item that owns the modifier. A title note
// describes what the item is rather than its state during play, so unlike other features it can't be switchable.
type TitleNote struct {
	Type feature.Type `json:"type"`
	Text string       `json:"text,omitzero"`
}

// NewTitleNote creates a new TitleNote.
func NewTitleNote() *TitleNote {
	return &TitleNote{Type: feature.TitleNote}
}

// FeatureType implements Feature.
func (s *TitleNote) FeatureType() feature.Type {
	return s.Type
}

// Clone implements Feature.
func (s *TitleNote) Clone() Feature {
	return clonePtr(s)
}

// IsSwitchable implements Feature. A title note is never switchable.
func (s *TitleNote) IsSwitchable() bool {
	return false
}

// SetSwitchable implements Feature. A title note is never switchable, so this does nothing.
func (s *TitleNote) SetSwitchable(_ bool) {
}

// FillWithNameableKeys implements Feature.
func (s *TitleNote) FillWithNameableKeys(m, existing map[string]string) {
	nameable.Extract(m, existing, s.Text)
}

// Describe returns a plain-language description of what the title note adds, in the form the features editor uses for
// every feature. Nameable markers take their values from replacements, and those without one are left as they are. The
// text is passed through em, which may wrap it for emphasis.
func (s *TitleNote) Describe(_ *Entity, replacements map[string]string, em func(string) string) string {
	text := strings.TrimSpace(nameable.Apply(s.Text, replacements))
	if text == "" {
		return i18n.Text("Adds an empty title note")
	}
	return fmt.Sprintf(i18n.Text("Adds the title note %s"), em(text))
}

// Hash writes this object's contents into the hasher.
func (s *TitleNote) Hash(h hash.Hash) {
	if s == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, s.Type)
	xhash.StringWithLen(h, s.Text)
}

// appendTitleNotes appends the non-empty text of the title note features in the lists to parts, with the replacements
// applied.
func appendTitleNotes(parts []string, replacements map[string]string, lists ...Features) []string {
	for _, list := range lists {
		for _, f := range list {
			if note, ok := f.(*TitleNote); ok {
				if text := strings.TrimSpace(nameable.Apply(note.Text, replacements)); text != "" {
					parts = append(parts, text)
				}
			}
		}
	}
	return parts
}

// modifierTitleNotes appends the title notes of the enabled, non-container modifiers to parts, in modifier order: for
// each, its CompactName when it is shown in the title, followed by the text of its title note features.
func modifierTitleNotes[M ModifierNode[M, T], T ModifiableNode[T, M], S ~[]M](parts []string, modifiers S, replacements map[string]string, features func(M) Features) []string {
	Traverse(func(mod M) bool {
		if mod.ShowsInTitle() {
			if name := strings.TrimSpace(mod.CompactName()); name != "" {
				parts = append(parts, name)
			}
		}
		parts = appendTitleNotes(parts, replacements, features(mod))
		return false
	}, true, true, modifiers...)
	return parts
}

// writeParenthetical writes the parts, separated by "; " and wrapped in parentheses, preceded by a space. Nothing is
// written when there are no parts.
func writeParenthetical(buffer *strings.Builder, parts []string) {
	if len(parts) == 0 {
		return
	}
	buffer.WriteString(" (")
	buffer.WriteString(strings.Join(parts, "; "))
	buffer.WriteByte(')')
}

// describeTitleNote returns the clause a prerequisite adds to its description for its title note criteria, such as
// ` with the title note Pistol`, or nothing when any title note will do. A non-empty qualifier is passed through em.
func describeTitleNote(t criteria.Text, replacements map[string]string, em func(string) string) string {
	if t.Compare == criteria.AnyText {
		return ""
	}
	q := nameable.Apply(t.Qualifier, replacements)
	if q != "" {
		q = em(q)
	}
	if t.Compare == criteria.IsText && q != "" {
		return i18n.Text(" with the title note ") + q
	}
	return " " + t.Compare.DescribeWithPrefix(i18n.Text("with a title note that"), i18n.Text("with all title notes that"), q)
}
