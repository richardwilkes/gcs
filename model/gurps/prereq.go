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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// Prereq holds data necessary to track a prerequisite.
type Prereq interface {
	nameable.Filler
	PrereqType() prereq.Type
	// ParentList returns the owning parent list, if any.
	ParentList() *PrereqList
	// SetParentList sets the owning parent list.
	SetParentList(list *PrereqList)
	// Clone creates a new copy of this Prereq.
	Clone(parent *PrereqList) Prereq
	// Satisfied returns true if this Prereq is satisfied by the specified Entity. 'buffer', if not nil, receives a
	// description of what was unsatisfied, with 'prefix' written before each line. 'hasEquipmentPenalty', if not nil,
	// is only ever set to true, by an unmet equipped-equipment prerequisite and by an unsatisfied list one contributed
	// to (see PrereqList.Satisfied); that is what earns a skill or spell the missing-equipment penalty.
	Satisfied(entity *Entity, exclude any, buffer *xbytes.InsertBuffer, prefix string, hasEquipmentPenalty *bool) bool
	// Describe returns a plain-language description of what this Prereq requires, whether or not it is met, naming
	// attributes and giving weights as the entity, which may be nil, defines them. Names and qualifiers are passed
	// through em, which may wrap them for emphasis; pass an identity func for plain text.
	Describe(entity *Entity, replacements map[string]string, em func(string) string) string
	// Hash writes this object's contents into the hasher.
	Hash(h hash.Hash)
}

// HasText returns the appropriate text for has.
func HasText(has bool) string {
	if has {
		return i18n.Text("Has")
	}
	return i18n.Text("Does not have")
}

// plainText is the em func for descriptions that want no emphasis.
func plainText(s string) string {
	return s
}

// describeText returns the comparison and qualifier of t, such as `is Fire` or `contains "Fi"`. A non-empty qualifier
// is passed through em, and is quoted unless the comparison is "is" and the qualifier reads plainly (see
// describeComparison).
func describeText(t criteria.Text, replacements map[string]string, em func(string) string) string {
	return describeComparison(t.Compare.String(), t.Compare, []string{nameable.Apply(t.Qualifier, replacements)}, em)
}

// describeComparison returns words, which say how the comparison compares, followed by the qualifiers, joined with
// "or". Each is passed through em, and quoted unless the comparison is "is" and it reads plainly, without a comma, a
// quote or space at either end that would blur where it starts and stops. One that is empty, or only space, reads as
// "". "is anything" takes no qualifier, and no qualifiers reads as one empty one.
func describeComparison(words string, compare criteria.StringComparison, qualifiers []string,
	em func(string) string,
) string {
	if compare.EnsureValid() == criteria.AnyText {
		return words
	}
	if len(qualifiers) == 0 {
		qualifiers = []string{""}
	}
	parts := make([]string, len(qualifiers))
	for i, q := range qualifiers {
		switch {
		case strings.TrimSpace(q) == "":
			parts[i] = `""`
		case compare == criteria.IsText && q == strings.TrimSpace(q) && !strings.ContainsAny(q, `,"`):
			parts[i] = em(q)
		default:
			parts[i] = `"` + em(q) + `"`
		}
	}
	text := parts[len(parts)-1]
	if len(parts) > 1 {
		text = i18n.Text("%s or %s", strings.Join(parts[:len(parts)-1], i18n.Text(", ")), text)
	}
	return words + " " + text
}

// describeName returns how a prerequisite names what it looks for: the bare name for "is", "of any name" when any name
// will do, and a "whose name" clause otherwise.
func describeName(t criteria.Text, replacements map[string]string, em func(string) string) string {
	switch {
	case t.Compare == criteria.AnyText:
		return i18n.Text("of any name")
	case t.Compare == criteria.IsText && t.Qualifier != "":
		return em(nameable.Apply(t.Qualifier, replacements))
	default:
		return i18n.Text("whose name ") + describeText(t, replacements, em)
	}
}

// describeSpecialization returns how a specialization narrows a skill: the bare specialization in parentheses for
// "is", nothing when any will do, and a "with a specialization that" clause otherwise, followed by the same kind of
// clause for the optional specialization when it is set.
func describeSpecialization(specialization, optional criteria.Text, replacements map[string]string, em func(string) string) string {
	var text string
	switch {
	case specialization.Compare == criteria.AnyText:
	case specialization.Compare == criteria.IsText && specialization.Qualifier != "":
		text = " (" + em(nameable.Apply(specialization.Qualifier, replacements)) + ")"
	default:
		text = i18n.Text(" with a specialization that ") + describeText(specialization, replacements, em)
	}
	if optional.Compare != criteria.AnyText {
		text += i18n.Text(" with an optional specialization that ") + describeText(optional, replacements, em)
	}
	return text
}

// describeTags returns a clause for the tags t matches, such as " tagged Weapon", or nothing when any tags will do.
func describeTags(t criteria.Text, replacements map[string]string, em func(string) string) string {
	if t.Compare == criteria.AnyText {
		return ""
	}
	q := nameable.Apply(t.Qualifier, replacements)
	if q != "" {
		q = em(q)
	}
	if t.Compare == criteria.IsText && q != "" {
		return i18n.Text(" tagged ") + q
	}
	return " " + t.Compare.DescribeWithPrefix(i18n.Text("with a tag that"), i18n.Text("with all tags that"), q)
}
