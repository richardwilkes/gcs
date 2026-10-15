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
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
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
		ID:        tid.MustNewTID(kinds.ShoppingRecord),
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
		ID:    tid.MustNewTID(kinds.ShoppingRecord),
		When:  mustTime(t, "2026-10-13T10:05:00Z"),
		Kind:  ShoppingBought,
		Items: []ShoppingRecordItem{{Name: "Rope", Quantity: fxp.FromInteger(1)}},
		Money: []ShoppingRecordItem{{Name: "Copper Penny", Quantity: fxp.FromInteger(-3)}},
	}
	clone := CloneShoppingRecordList([]*ShoppingRecordEntry{entry})
	c.Equal([]*ShoppingRecordEntry{entry}, clone)
	clone[0].Items[0].Quantity = fxp.FromInteger(9)
	clone[0].Money[0].Quantity = fxp.FromInteger(9)
	clone[0].Name = "changed"
	clone[0].ID = tid.MustNewTID(kinds.ShoppingRecord)
	c.Equal(fxp.FromInteger(1), entry.Items[0].Quantity)
	c.Equal(fxp.FromInteger(-3), entry.Money[0].Quantity)
	c.Equal("", entry.Name)
	c.NotEqual(entry.ID, clone[0].ID)
}

func TestAddShoppingRecordKeepsMostRecentFirst(t *testing.T) {
	c := check.New(t)
	var e Entity
	older := &ShoppingRecordEntry{When: mustTime(t, "2026-10-12T16:30:00Z"), Kind: ShoppingSold}
	newer := &ShoppingRecordEntry{When: mustTime(t, "2026-10-13T10:05:00Z"), Kind: ShoppingBought}
	e.AddShoppingRecord(newer)
	e.AddShoppingRecord(older)
	c.Equal([]*ShoppingRecordEntry{newer, older}, e.ShoppingRecord)
	c.True(tid.IsKindAndValid(older.ID, kinds.ShoppingRecord), "an entry without an ID is given one")
	c.True(tid.IsKindAndValid(newer.ID, kinds.ShoppingRecord))
	c.NotEqual(older.ID, newer.ID)

	id := newer.ID
	e.AddShoppingRecord(newer)
	c.Equal(id, newer.ID, "an entry with an ID keeps it")
}

func TestShoppingRecordEntriesOfTheSameSecondStayNewestFirst(t *testing.T) {
	c := check.New(t)
	var e Entity
	when := mustTime(t, "2026-10-13T10:05:00Z")
	first := &ShoppingRecordEntry{When: when, Name: "first"}
	second := &ShoppingRecordEntry{When: when, Name: "second"}
	third := &ShoppingRecordEntry{When: when, Name: "third"}
	earlier := &ShoppingRecordEntry{When: mustTime(t, "2026-10-13T10:04:59Z"), Name: "earlier"}
	e.AddShoppingRecord(first)
	e.AddShoppingRecord(earlier)
	e.AddShoppingRecord(second)
	e.AddShoppingRecord(third)
	c.Equal([]*ShoppingRecordEntry{third, second, first, earlier}, e.ShoppingRecord)

	SortShoppingRecordList(e.ShoppingRecord)
	c.Equal([]*ShoppingRecordEntry{third, second, first, earlier}, e.ShoppingRecord, "sorting again keeps the order")

	e.SetShoppingRecord([]*ShoppingRecordEntry{earlier, third, second, first})
	names := make([]string, 0, len(e.ShoppingRecord))
	for _, entry := range e.ShoppingRecord {
		names = append(names, entry.Name)
	}
	c.Equal([]string{"third", "second", "first", "earlier"}, names)
}

func TestSetShoppingRecordMakesIndependentCopies(t *testing.T) {
	c := check.New(t)
	var e Entity
	entry := &ShoppingRecordEntry{
		ID:    tid.MustNewTID(kinds.ShoppingRecord),
		When:  mustTime(t, "2026-10-13T10:05:00Z"),
		Kind:  ShoppingSold,
		Items: []ShoppingRecordItem{{Name: "Rope", Quantity: fxp.One}},
	}
	list := []*ShoppingRecordEntry{entry}
	e.SetShoppingRecord(list)
	c.Equal(list, e.ShoppingRecord)
	c.True(e.ShoppingRecord[0] != entry, "the entry is copied")
	e.ShoppingRecord[0].Items[0].Quantity = fxp.Two
	e.ShoppingRecord[0].Name = "changed"
	c.Equal(fxp.One, entry.Items[0].Quantity)
	c.Equal("", entry.Name)
	e.SetShoppingRecord(nil)
	c.Equal(0, len(e.ShoppingRecord))
	c.Equal(1, len(list))
}

func TestShoppingRecordGetsIDsWhenLoaded(t *testing.T) {
	c := check.New(t)
	kept := tid.MustNewTID(kinds.ShoppingRecord)
	e := NewEntity()
	e.ShoppingRecord = []*ShoppingRecordEntry{
		{When: mustTime(t, "2026-10-12T16:30:00Z"), Name: "older", Kind: ShoppingSold},
		{ID: kept, When: mustTime(t, "2026-10-13T10:05:00Z"), Name: "newer", Kind: ShoppingBought},
		{When: mustTime(t, "2026-10-13T10:05:00Z"), Name: "same time", Kind: ShoppingSold},
	}
	data, err := jio.Marshal(e)
	c.NoError(err)
	var loaded Entity
	c.NoError(jio.Unmarshal(data, &loaded))
	c.Equal(3, len(loaded.ShoppingRecord))
	names := make([]string, 0, len(loaded.ShoppingRecord))
	for _, entry := range loaded.ShoppingRecord {
		names = append(names, entry.Name)
		c.True(tid.IsKindAndValid(entry.ID, kinds.ShoppingRecord), "every entry has an ID once loaded")
	}
	c.Equal([]string{"newer", "same time", "older"}, names, "loading sorts most recent first, keeping ties in order")
	c.Equal(kept, loaded.ShoppingRecord[0].ID, "an entry that had an ID keeps it")
	c.NotEqual(loaded.ShoppingRecord[1].ID, loaded.ShoppingRecord[2].ID)
}

// TestShoppingRecordDuplicateIDsAreReplacedWhenLoaded checks that entries of a loaded record that share an ID, as a
// hand-edited or merged file may hold, are told apart: the first keeps the ID and each later one gets its own, since the
// record's viewers identify and remove entries by ID.
func TestShoppingRecordDuplicateIDsAreReplacedWhenLoaded(t *testing.T) {
	c := check.New(t)
	shared := tid.MustNewTID(kinds.ShoppingRecord)
	e := NewEntity()
	e.ShoppingRecord = []*ShoppingRecordEntry{
		{ID: shared, When: mustTime(t, "2026-10-13T10:05:00Z"), Name: "first", Kind: ShoppingBought},
		{ID: shared, When: mustTime(t, "2026-10-12T16:30:00Z"), Name: "second", Kind: ShoppingSold},
		{ID: shared, When: mustTime(t, "2026-10-11T16:30:00Z"), Name: "third", Kind: ShoppingSold},
	}
	data, err := jio.Marshal(e)
	c.NoError(err)
	var loaded Entity
	c.NoError(jio.Unmarshal(data, &loaded))
	c.Equal(3, len(loaded.ShoppingRecord))
	seen := make(map[tid.TID]bool)
	for _, entry := range loaded.ShoppingRecord {
		c.True(tid.IsKindAndValid(entry.ID, kinds.ShoppingRecord), "every entry has a valid ID")
		c.False(seen[entry.ID], "no two entries share an ID")
		seen[entry.ID] = true
	}
	c.Equal("first", loaded.ShoppingRecord[0].Name)
	c.Equal(shared, loaded.ShoppingRecord[0].ID, "the first entry with the ID keeps it")
}
