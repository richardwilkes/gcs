// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps_test

import (
	"encoding/json/jsontext"
	"fmt"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/filternode"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/levelcap"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestEveryFeatureTypeDecodesToItsConcreteType verifies that the polymorphic feature decoder has a concrete type for
// every feature type this build knows about, so that none of them is quietly preserved as an UnknownFeature, and that
// the reserved Unknown type itself, which has no concrete representation, is.
func TestEveryFeatureTypeDecodesToItsConcreteType(t *testing.T) {
	for _, one := range feature.Types {
		t.Run(one.Key(), func(t *testing.T) {
			c := check.New(t)
			var features gurps.Features
			c.NoError(jio.Unmarshal(fmt.Appendf(nil, `[{"type":%q}]`, one.Key()), &features))
			c.Equal(1, len(features))
			if len(features) != 1 {
				return
			}
			_, isUnknown := features[0].(*gurps.UnknownFeature)
			c.Equal(one == feature.Unknown, isUnknown, "unknown wrapper used")
			c.Equal(one, features[0].FeatureType())
			if one.IsWeaponBonus() {
				_, isWeaponBonus := features[0].(*gurps.WeaponBonus)
				c.True(isWeaponBonus, "weapon bonus types decode to a WeaponBonus")
			}
		})
	}
}

// TestEveryPrereqTypeDecodesToItsConcreteType is the prerequisite counterpart of
// TestEveryFeatureTypeDecodesToItsConcreteType.
func TestEveryPrereqTypeDecodesToItsConcreteType(t *testing.T) {
	for _, one := range prereq.Types {
		t.Run(one.Key(), func(t *testing.T) {
			c := check.New(t)
			var prereqs gurps.Prereqs
			c.NoError(jio.Unmarshal(fmt.Appendf(nil, `[{"type":%q}]`, one.Key()), &prereqs))
			c.Equal(1, len(prereqs))
			if len(prereqs) != 1 {
				return
			}
			_, isUnknown := prereqs[0].(*gurps.UnknownPrereq)
			c.Equal(one == prereq.Unknown, isUnknown, "unknown wrapper used")
			c.Equal(one, prereqs[0].PrereqType())
		})
	}
}

// TestTypedListRejectsMalformedElements verifies that the shared decoder reports, rather than swallows, JSON that is
// not a list of objects or whose element can't be decoded into its concrete type.
func TestTypedListRejectsMalformedElements(t *testing.T) {
	c := check.New(t)
	var features gurps.Features
	c.HasError(jio.Unmarshal([]byte(`{"type":"attribute_bonus"}`), &features), "not a list")
	c.HasError(jio.Unmarshal([]byte(`[1]`), &features), "element is not an object")
	c.HasError(jio.Unmarshal([]byte(`[{"type":"attribute_bonus","amount":"not a number"}]`), &features),
		"element does not decode into its concrete type")
	var prereqs gurps.Prereqs
	c.HasError(jio.Unmarshal([]byte(`[{"type":"trait_prereq","has":"maybe"}]`), &prereqs),
		"element does not decode into its concrete type")
}

// TestEveryTypedListElementKeepsUnknownMembers verifies that every type a typed list decodes into, features,
// prerequisites, level cap nodes and filter nodes alike, keeps the members this release does not know: the element
// writes them back unchanged, and so does a clone of it, even once more has been decoded into the element.
func TestEveryTypedListElementKeepsUnknownMembers(t *testing.T) {
	for _, one := range feature.Types {
		checkKeepsUnknownMembers[gurps.Features](t, "feature/"+one.Key(), gurps.Feature.Clone)
	}
	for _, one := range prereq.Types {
		checkKeepsUnknownMembers[gurps.Prereqs](t, "prereq/"+one.Key(),
			func(p gurps.Prereq) gurps.Prereq { return p.Clone(nil) })
	}
	for _, one := range levelcap.Types {
		checkKeepsUnknownMembers[gurps.LevelCaps](t, "levelcap/"+one.Key(),
			func(n gurps.LevelCap) gurps.LevelCap { return n.Clone(nil) })
	}
	for _, one := range filternode.Types {
		checkKeepsUnknownMembers[gurps.FilterNodes](t, "filternode/"+one.Key(),
			func(n gurps.FilterNode) gurps.FilterNode { return n.Clone(nil) })
	}
}

// checkKeepsUnknownMembers runs a subtest, named by the list kind and the element's type key, that decodes a list of
// type L holding one element of that type with a member this release does not know, and checks that the element and a
// clone of it write the member back unchanged.
func checkKeepsUnknownMembers[L ~[]E, E any](t *testing.T, name string, clone func(E) E) {
	t.Run(name, func(t *testing.T) {
		c := check.New(t)
		const future = `{"a":[1,2]}`
		_, key, _ := strings.Cut(name, "/")
		var list L
		c.NoError(jio.Unmarshal(fmt.Appendf(nil, `[{"type":%q,"zzz_future":%s}]`, key, future), &list))
		c.Equal(1, len(list))
		if len(list) != 1 {
			return
		}
		c.Equal(future, memberJSON(c, list[0], "zzz_future"), "the element")
		dup := clone(list[0])
		c.NoError(jio.Unmarshal([]byte(`{"zzz_more":2}`), list[0]))
		c.Equal(future, memberJSON(c, dup, "zzz_future"), "the clone")
		c.Equal("", memberJSON(c, dup, "zzz_more"), "the clone keeps nothing decoded into the element later")
	})
}

// memberJSON returns the compact JSON of the named member of the object v writes, or "" when it has none.
func memberJSON(c check.Checker, v any, name string) string {
	data, err := jio.Marshal(v)
	c.NoError(err)
	var members map[string]jsontext.Value
	c.NoError(jio.Unmarshal(data, &members))
	value := members[name]
	if len(value) == 0 {
		return ""
	}
	c.NoError(value.Compact())
	return string(value)
}
