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
	"hash"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

// hashIntPtr writes a presence byte, then the value when present.
func hashIntPtr[T ~int64](h hash.Hash, p *T) {
	xhash.Bool(h, p != nil)
	if p != nil {
		xhash.Num64(h, *p)
	}
}

// canonicalText returns t with the qualifier cleared when the comparison is "any", since it then means nothing.
func canonicalText(t criteria.Text) criteria.Text {
	if t.IsZero() {
		return criteria.Text{}
	}
	return t
}

// canonicalNumber returns n with the qualifier cleared when the comparison is "any".
func canonicalNumber(n criteria.Number) criteria.Number {
	if n.IsZero() {
		return criteria.Number{}
	}
	return n
}

// canonicalWeight returns w with the qualifier cleared when the comparison is "any".
func canonicalWeight(w criteria.Weight) criteria.Weight {
	if w.IsZero() {
		return criteria.Weight{}
	}
	return w
}

// canonicalQuantity returns the canonical form of a Count quantity or per-group: nil for nil and for "at least 1",
// which is what nil means, otherwise a criterion with no stale qualifier.
func canonicalQuantity(p *criteria.Number) *criteria.Number {
	if p == nil || (p.Compare == criteria.AtLeastNumber && p.Qualifier == fxp.One) {
		return nil
	}
	n := canonicalNumber(*p)
	return &n
}

// hashQuantity writes the canonical form of a Count quantity or per-group into the hasher. criteria.Number's Hash
// already ignores the qualifier of "any", so only "at least 1" needs mapping to nil.
func hashQuantity(h hash.Hash, p *criteria.Number) {
	present := p != nil && (p.Compare != criteria.AtLeastNumber || p.Qualifier != fxp.One)
	xhash.Bool(h, present)
	if present {
		p.Hash(h)
	}
}
