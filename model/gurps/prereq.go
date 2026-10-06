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

// describeText returns the comparison and qualifier of t, such as `is Fire` or `contains "Fi"`, the qualifier shown as
// describeValue shows it.
func describeText(t criteria.Text, replacements map[string]string, em func(string) string) string {
	return describeComparison(t.Compare.String(), t.Compare, []string{nameable.Apply(t.Qualifier, replacements)}, em)
}

// describeComparison returns words, which say how the comparison compares, followed by the qualifiers, joined with
// "or", each shown as describeValue shows it. One qualifier after "is" may be bare; several are always quoted, so that
// they can't be read as one value holding "or", or one as several. "is anything" takes no qualifier.
func describeComparison(words string, compare criteria.StringComparison, qualifiers []string,
	em func(string) string,
) string {
	if compare.EnsureValid() == criteria.AnyText {
		return words
	}
	bare := compare == criteria.IsText && len(qualifiers) == 1
	parts := make([]string, len(qualifiers))
	for i, q := range qualifiers {
		parts[i] = describeValue(q, bare, em)
	}
	text := parts[len(parts)-1]
	if len(parts) > 1 {
		text = i18n.Text("%s or %s", strings.Join(parts[:len(parts)-1], i18n.Text(", ")), text)
	}
	return words + " " + text
}

// describeValue returns a qualifier as a sentence shows it, passed through em: bare when bare is set and it reads
// plainly, without a comma, a double quote or space at either end that would blur where it starts and stops, and
// otherwise in double quotes, or in single quotes when it holds a double quote. An empty one reads "", and one that is
// only space is quoted as it is, since it matches as it is.
func describeValue(q string, bare bool, em func(string) string) string {
	switch {
	case q == "":
		return `""`
	case bare && q == strings.TrimSpace(q) && !strings.ContainsAny(q, `,"`):
		return em(q)
	case strings.Contains(q, `"`):
		return "'" + em(q) + "'"
	default:
		return `"` + em(q) + `"`
	}
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
