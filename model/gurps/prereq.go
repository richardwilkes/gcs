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
// is passed through em, and is quoted unless the comparison is "is". "is" and "is not" drop space at either end of it,
// as they do when matching.
func describeText(t criteria.Text, replacements map[string]string, em func(string) string) string {
	q := t.Compare.EffectiveQualifier(nameable.Apply(t.Qualifier, replacements))
	if q == "" {
		return t.Compare.Describe(q)
	}
	if t.Compare == criteria.IsText {
		return t.Compare.String() + " " + em(q)
	}
	return t.Compare.Describe(em(q))
}

// describeName returns how a prerequisite names what it looks for: the bare name for "is", or "" when that is blank
// once its markers are replaced, "of any name" when any name will do, and a "whose name" clause otherwise.
func describeName(t criteria.Text, replacements map[string]string, em func(string) string) string {
	switch t.Compare {
	case criteria.AnyText:
		return i18n.Text("of any name")
	case criteria.IsText:
		if q := strings.TrimSpace(nameable.Apply(t.Qualifier, replacements)); q != "" {
			return em(q)
		}
		return `""`
	default:
		return i18n.Text("whose name ") + describeText(t, replacements, em)
	}
}

// describeSpecialization returns how a specialization narrows a skill: the bare specialization in parentheses for
// "is", "without a specialization" when "is" names none once its markers are replaced, since that picks only a skill
// without one, nothing when any will do, and a "with a specialization that" clause otherwise, followed by the same kind
// of clause for the optional specialization when it is set.
func describeSpecialization(specialization, optional criteria.Text, replacements map[string]string, em func(string) string) string {
	var text string
	switch specialization.Compare {
	case criteria.AnyText:
	case criteria.IsText:
		if q := strings.TrimSpace(nameable.Apply(specialization.Qualifier, replacements)); q != "" {
			text = " (" + em(q) + ")"
		} else {
			text = i18n.Text(" without a specialization")
		}
	default:
		text = i18n.Text(" with a specialization that ") + describeText(specialization, replacements, em)
	}
	switch {
	case optional.Compare == criteria.AnyText:
	case optional.Compare == criteria.IsText && strings.TrimSpace(nameable.Apply(optional.Qualifier, replacements)) == "":
		text += i18n.Text(" without an optional specialization")
	default:
		text += i18n.Text(" with an optional specialization that ") + describeText(optional, replacements, em)
	}
	return text
}

// describeTags returns a clause for the tags t matches, such as " tagged Weapon", or nothing when any tags will do. A
// blank "is" picks only what has no tags, and a blank "is not" only what has some.
func describeTags(t criteria.Text, replacements map[string]string, em func(string) string) string {
	q := t.Compare.EffectiveQualifier(nameable.Apply(t.Qualifier, replacements))
	switch {
	case t.Compare == criteria.AnyText:
		return ""
	case t.Compare == criteria.IsText && q == "":
		return i18n.Text(" without tags")
	case t.Compare == criteria.IsNotText && q == "":
		return i18n.Text(" with at least one tag")
	case t.Compare == criteria.IsText:
		return i18n.Text(" tagged ") + em(q)
	}
	if q != "" {
		q = em(q)
	}
	return " " + t.Compare.DescribeWithPrefix(i18n.Text("with a tag that"), i18n.Text("with all tags that"), q)
}
