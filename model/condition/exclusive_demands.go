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
	"github.com/richardwilkes/gcs/v5/model/criteria"
)

// demand is a count that consumes members from a pool: it needs q of its candidates, or, when grouped, values each
// holding p members. Every member counts toward at most one demand of a pool, and toward one value only.
type demand struct {
	count        *Count
	pos          map[int]int   // pool member to its index in cands
	memberValues map[int][]int // grouped: pool member to the values it has, in its own order
	cands        []int         // pool members whose where passed, in collection order
	wheres       []Result      // the where result of each candidate, in Collect mode
	values       [][]int       // grouped: the candidates of each value, in order of first appearance
	q            int
	p            int
	grouped      bool
	// Where the demand sits in its pool's network: the node and sink edge of a plain demand, or of the node a
	// per-group-1 demand's values feed; and for a demand needing p >= 2 members per value, each value's node and sink
	// edge.
	node       int
	sink       int
	valueNodes []int
	valueSinks []int
}

// owner records where a pool member went: nowhere when d is nil, else to d, and for a grouped demand to one value.
type owner struct {
	d     *demand
	value int
}

// need returns how many a quantity, which Validate has made at least a whole number, asks for, nil meaning 1.
func need(q *criteria.Number) int {
	if q == nil {
		return 1
	}
	return q.Qualifier.AsInteger[int]()
}

// newDemand returns a demand for c over the given candidates.
func newDemand(c *Count, cands []int, wheres []Result) *demand {
	d := &demand{count: c, cands: cands, wheres: wheres, pos: make(map[int]int, len(cands)), q: need(c.Quantity)}
	for i, m := range cands {
		d.pos[m] = i
	}
	return d
}

// group sorts the candidates of a grouped demand into the values of its GroupBy field, items being the pool's
// members. It returns the Failed result to report when a candidate's kind lacks the field.
func (d *demand) group(e *evaluator, items []Item) *Result {
	d.grouped = true
	d.p = need(d.count.PerGroup)
	d.memberValues = make(map[int][]int)
	index := make(map[string]int)
	for _, m := range d.cands {
		names, fail := e.groupValues(items[m], d.count.GroupBy)
		if fail != nil {
			return fail
		}
		for _, name := range names {
			v, ok := index[name]
			if !ok {
				v = len(d.values)
				index[name] = v
				d.values = append(d.values, nil)
			}
			d.values[v] = append(d.values[v], m)
			d.memberValues[m] = append(d.memberValues[m], v)
		}
	}
	return nil
}

// tally counts what owners give d: members for a plain demand, and for a grouped one the members of each value and the
// number of values holding at least p.
func (d *demand) tally(owners []owner) (members int, perValue []int, filled int) {
	if d.grouped {
		perValue = make([]int, len(d.values))
	}
	for _, m := range d.cands {
		if o := owners[m]; o.d == d {
			members++
			if d.grouped {
				perValue[o.value]++
			}
		}
	}
	for _, n := range perValue {
		if n >= d.p {
			filled++
		}
	}
	return members, perValue, filled
}

// surplus gives each unused candidate to the first demand, in tree order, that has it, and for a grouped demand to
// the first of its values already filled, so that Have and the hits show every member that counts. It runs last and
// never changes a state.
func surplus(demands []*demand, owners []owner) {
	perValues := make([][]int, len(demands))
	for i, d := range demands {
		_, perValues[i], _ = d.tally(owners)
	}
	for m := range owners {
		if owners[m].d != nil {
			continue
		}
	next:
		for i, d := range demands {
			if _, ok := d.pos[m]; !ok {
				continue
			}
			if !d.grouped {
				owners[m] = owner{d: d, value: -1}
				break
			}
			for _, v := range d.memberValues[m] {
				if perValues[i][v] >= d.p {
					owners[m] = owner{d: d, value: v}
					perValues[i][v]++
					break next
				}
			}
		}
	}
}
