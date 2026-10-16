// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package condition

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/i18n"
)

// eachMember calls fn, in collection order, for each member of m's collection whose where passes, with the where's
// result, until fn returns false. Self is left out unless m includes it. A collection the host doesn't have, a where
// that ends Failed, and running out of budget all return the Failed result to report. The where's result is only
// passed in Collect mode, through one copy per call, so that YesNo allocates nothing per member.
func (e *evaluator) eachMember(m *Members, f frame, fn func(item Item, where *Result) bool) *Result {
	items, ok := e.host.Collection(m.Of, e.self, f.subject)
	if !ok {
		r := failure(fmt.Sprintf(i18n.Text("unknown collection %q"), m.Of))
		return &r
	}
	var where *Result
	if e.collect {
		where = new(Result)
	}
	for _, item := range items {
		if item == e.self && !m.IncludeSelf {
			continue
		}
		if !e.spend(1) {
			r := failure(reasonTooComplex())
			return &r
		}
		mf := frame{subject: item, of: &m.Of}
		var o outcome
		if where != nil {
			*where = e.evalRef(m.Where, mf)
			o = outcome{reason: where.Reason, state: where.State}
		} else {
			o = e.testRef(m.Where, mf)
		}
		switch {
		case o.state.Passes():
			if !fn(item, where) {
				return nil
			}
		case o.state == Failed:
			r := failure(o.reason)
			return &r
		default:
		}
	}
	return nil
}

// eval implements Node.
func (c *Count) eval(e *evaluator, f frame) Result {
	collect := e.collect
	keep := collect || c.GroupBy != ""
	have := 0
	var members []Item
	var wheres []Result
	if fail := e.eachMember(&c.Members, f, func(item Item, where *Result) bool {
		have++
		if keep {
			members = append(members, item)
			if collect {
				wheres = append(wheres, *where)
			}
		}
		return true
	}); fail != nil {
		return *fail
	}
	var used []bool
	if c.GroupBy != "" {
		if c.Exclusive {
			var fail *Result
			if have, used, fail = e.exclusiveGroupedCount(c, members); fail != nil {
				return *fail
			}
		} else {
			var fail *Result
			if have, used, fail = e.groupBy(c, members); fail != nil {
				return *fail
			}
		}
	}
	r := Result{Have: fxp.FromInteger(have), HaveKind: ValueNumber, Need: fxp.One}
	met := have >= 1
	if c.Quantity != nil {
		met = c.Quantity.Matches(r.Have)
		r.Need = c.Quantity.Qualifier
	}
	r.State = metState(met, c.Not)
	if collect {
		c.collect(&r, members, wheres, used)
	}
	return r
}

// collect fills in the hits and children of a Collect result: each counted member is a hit, and the result of its
// where a child.
func (c *Count) collect(r *Result, members []Item, wheres []Result, used []bool) {
	for i, item := range members {
		if len(used) != 0 && !used[i] {
			continue
		}
		r.hits = append(r.hits, Hit{Node: c, Item: item})
		if !c.Where.IsZero() {
			r.Children = append(r.Children, wheres[i])
		}
	}
	if r.State == Met && !c.Not {
		r.Specificity = 1
	}
}

// groupBy groups the matched members by the values of the GroupBy field and returns the number of values whose
// members pass PerGroup, along with which members those values hold.
func (e *evaluator) groupBy(c *Count, members []Item) (have int, used []bool, fail *Result) {
	index := make(map[string]int)
	var groups [][]int
	for i, item := range members {
		values, bad := e.groupValues(item, c.GroupBy)
		if bad != nil {
			return 0, nil, bad
		}
		for _, value := range values {
			g, ok := index[value]
			if !ok {
				g = len(groups)
				index[value] = g
				groups = append(groups, nil)
			}
			groups[g] = append(groups[g], i)
		}
	}
	used = make([]bool, len(members))
	for _, g := range groups {
		if c.PerGroup == nil || c.PerGroup.Matches(fxp.FromInteger(len(g))) {
			have++
			for _, i := range g {
				used[i] = true
			}
		}
	}
	return have, used, nil
}

// groupValues returns the distinct values a member's group field puts it under. Text values ignore case. An Absent
// field, and an empty value, join no group, while a field the member's kind lacks fails the count.
func (e *evaluator) groupValues(item Item, key string) ([]string, *Result) {
	v, ok := e.host.Field(item, key)
	if !ok {
		r := failure(fmt.Sprintf(i18n.Text("no field %q"), key))
		return nil, &r
	}
	var values []string
	switch v.Kind {
	case ValueText:
		values = []string{strings.ToLower(v.Text)}
	case ValueList:
		values = make([]string, len(v.List))
		for i, one := range v.List {
			values[i] = strings.ToLower(one)
		}
	case ValueNumber, ValueWeight:
		values = []string{v.Number.String()}
	case ValueBool:
		values = []string{strconv.FormatBool(v.Bool)}
	default:
	}
	var distinct []string
	for _, one := range values {
		if one != "" && !slices.Contains(distinct, one) {
			distinct = append(distinct, one)
		}
	}
	return distinct, nil
}
