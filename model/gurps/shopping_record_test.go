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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

func mustTime(t *testing.T, s string) jio.Time {
	t.Helper()
	when, err := jio.NewTimeFrom(s)
	if err != nil {
		t.Fatal(err)
	}
	return when
}

func TestShoppingRecordRoundTrip(t *testing.T) {
	c := check.New(t)
	entry := &ShoppingRecordEntry{
		When:      mustTime(t, "2026-10-13T10:05:00Z"),
		Name:      "Sold small knives",
		Kind:      ShoppingSold,
		Items:     []ShoppingRecordItem{{Name: "Small Knife", Quantity: fxp.FromInteger(2), Value: fxp.FromInteger(30)}},
		Money:     []ShoppingRecordItem{{Name: "Silver Eagle", Quantity: fxp.FromInteger(1), Value: fxp.FromInteger(20)}, {Name: "Copper Penny", Quantity: fxp.FromInteger(11), Value: fxp.FromInteger(1)}},
		Exchanged: fxp.FromInteger(31),
	}
	data, err := jio.Marshal(&EntityData{ShoppingRecord: []*ShoppingRecordEntry{entry}})
	c.NoError(err)
	var loaded EntityData
	c.NoError(jio.Unmarshal(data, &loaded))
	c.Equal([]*ShoppingRecordEntry{entry}, loaded.ShoppingRecord)
}

func TestShoppingRecordCloneIsIndependent(t *testing.T) {
	c := check.New(t)
	entry := &ShoppingRecordEntry{
		When:  mustTime(t, "2026-10-13T10:05:00Z"),
		Kind:  ShoppingBought,
		Items: []ShoppingRecordItem{{Name: "Rope", Quantity: fxp.FromInteger(1)}},
	}
	clone := CloneShoppingRecordList([]*ShoppingRecordEntry{entry})
	clone[0].Items[0].Quantity = fxp.FromInteger(9)
	clone[0].Name = "changed"
	c.Equal(fxp.FromInteger(1), entry.Items[0].Quantity)
	c.Equal("", entry.Name)
}

func TestAddShoppingRecordKeepsMostRecentFirst(t *testing.T) {
	c := check.New(t)
	var e Entity
	older := &ShoppingRecordEntry{When: mustTime(t, "2026-10-12T16:30:00Z"), Kind: ShoppingSold}
	newer := &ShoppingRecordEntry{When: mustTime(t, "2026-10-13T10:05:00Z"), Kind: ShoppingBought}
	e.AddShoppingRecord(older)
	e.AddShoppingRecord(newer)
	c.Equal([]*ShoppingRecordEntry{newer, older}, e.ShoppingRecord)
}
