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
	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
)

// FilterFieldLookup returns the title and kind of the field with the key, whether the title names something plural,
// and false for ok when the list type has no such field. The titles are phrased to follow "must" or "must not", as in
// "have a name".
type FilterFieldLookup func(key string) (title string, kind FilterFieldKind, plural, ok bool)

// Describe returns a plain-language description of what the condition requires, such as `Must have a name that
// contains "sword"`, taking the field's title and kind from lookup and giving weights in units. What follows a plural
// title agrees with it, as in `Must have notes that contain "cheap"`. The qualifier is passed through em, which may
// wrap it for emphasis; pass an identity func for plain text.
func (c *FilterCondition) Describe(lookup FilterFieldLookup, units fxp.WeightUnit, em func(string) string) string {
	title, kind, plural, ok := lookup(c.Field)
	if !ok {
		return i18n.Text("Condition on unknown field %q; it can't be checked", c.Field)
	}
	var criterion string
	switch kind {
	case FilterFieldText:
		if !plural {
			criterion = describeText(c.Text, nil, em)
			break
		}
		criterion = describeComparison(c.Text.Compare.PluralClause(), c.Text.Compare,
			[]string{c.Text.Compare.EffectiveQualifier(nameable.Apply(c.Text.Qualifier, nil))}, em)
	case FilterFieldList:
		return c.describeList(title, em)
	case FilterFieldNumber:
		criterion = describeNumber(c.Number.Compare, em(c.Number.Qualifier.Comma()), plural)
	case FilterFieldWeight:
		criterion = describeNumber(c.Weight.Compare, em(units.Format(c.Weight.Qualifier)), plural)
	default:
		return c.describeHaving(title)
	}
	switch {
	case plural && c.Not:
		return i18n.Text("Must not %s %s", title, criterion)
	case plural:
		return i18n.Text("Must %s %s", title, criterion)
	case c.Not:
		return i18n.Text("Must not %s that %s", title, criterion)
	default:
		return i18n.Text("Must %s that %s", title, criterion)
	}
}

// describeHaving returns the description of a condition that says only whether the title holds, as a yes/no field's
// does.
func (c *FilterCondition) describeHaving(title string) string {
	if c.Not {
		return i18n.Text("Must not %s", title)
	}
	return i18n.Text("Must %s", title)
}

// describeNumber returns the comparison and the qualifier, as text: for a plural title, a clause that agrees with it,
// such as "that are at least 5", and otherwise the words that follow "that", such as "is at least 5".
func describeNumber(compare criteria.NumericComparison, qualifier string, plural bool) string {
	if !plural {
		return compare.DescribeWith(qualifier)
	}
	if compare.EnsureValid() == criteria.AnyNumber {
		return compare.PluralClause()
	}
	return compare.PluralClause() + " " + qualifier
}

// describeList returns the description of a condition on a list of values, which is matched value by value against
// each of the comma-separated qualifiers: a comparison holds when at least one value matches one of them, and a "not"
// comparison when none does. Either way, the list must hold something.
func (c *FilterCondition) describeList(title string, em func(string) string) string {
	compare := c.Text.Compare.EnsureValid()
	qualifiers := criteria.SplitQualifiers(nameable.Apply(c.Text.Qualifier, nil))
	text := describeComparison(compare.ListClause(), compare.Positive(), qualifiers, em)
	if c.Not {
		return i18n.Text("Must not %s %s", title, text)
	}
	return i18n.Text("Must %s %s", title, text)
}

// Describe returns what the editor says of a node this version of GCS doesn't understand.
func (n *UnknownFilterNode) Describe() string {
	return i18n.Text("Unknown filter node type %q; it will be preserved, but can't be checked", n.Kind)
}
