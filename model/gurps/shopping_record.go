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
	"slices"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/toolbox/v2/tid"
)

// ShoppingKind identifies whether a shopping record entry is a sale or a purchase.
type ShoppingKind string

// Shopping kinds.
const (
	ShoppingSold   ShoppingKind = "sold"
	ShoppingBought ShoppingKind = "bought"
)

// ShoppingRecordItem is a snapshot of an item, or of a money item, as it was when a transaction was made. It keeps no
// link back to the equipment, so later changes to the sheet don't alter the record. Value is the full (100%) price of
// one unit, including the contents of a container. The Quantity of an item is always positive. The Quantity of a money
// item is signed: positive when the character received the money and negative when they handed it over.
type ShoppingRecordItem struct {
	Name     string  `json:"name"`
	Quantity fxp.Int `json:"quantity"`
	Value    fxp.Int `json:"value"`
}

// ShoppingRecordEntry holds one sale or purchase made from this sheet. Items lists what was sold or bought, and Money
// the money that changed hands, so a purchase that got change lists both the money paid and the change. Exchanged is
// the unsigned amount of money that changed hands, net of any change.
type ShoppingRecordEntry struct {
	ID        tid.TID              `json:"id"`
	When      jio.Time             `json:"when"`
	Name      string               `json:"name,omitzero"`
	Kind      ShoppingKind         `json:"kind"`
	Items     []ShoppingRecordItem `json:"items,omitempty"`
	Money     []ShoppingRecordItem `json:"money,omitempty"`
	Exchanged fxp.Int              `json:"exchanged"`
}

// CloneShoppingRecordList creates a clone of the provided ShoppingRecordEntry list.
func CloneShoppingRecordList(list []*ShoppingRecordEntry) []*ShoppingRecordEntry {
	clone := make([]*ShoppingRecordEntry, len(list))
	for i, one := range list {
		entry := *one
		entry.Items = slices.Clone(one.Items)
		entry.Money = slices.Clone(one.Money)
		clone[i] = &entry
	}
	return clone
}

// SortShoppingRecordList sorts the provided ShoppingRecordEntry list into the order GCS presents and stores them in,
// which is most recent first. Times are kept only to the second, so entries made in the same second keep the order they
// had.
func SortShoppingRecordList(list []*ShoppingRecordEntry) {
	slices.SortStableFunc(list, func(a, b *ShoppingRecordEntry) int { return b.When.Compare(a.When) })
}

// AddShoppingRecord adds an entry to this sheet's shopping record, giving it an ID if it has none, and keeps the record
// in presentation order. The entry goes ahead of any made in the same second.
func (e *Entity) AddShoppingRecord(entry *ShoppingRecordEntry) {
	entry.ensureID()
	e.ShoppingRecord = slices.Insert(e.ShoppingRecord, 0, entry)
	SortShoppingRecordList(e.ShoppingRecord)
}

// SetShoppingRecord sets a new shopping record list, made of copies of the entries in presentation order.
func (e *Entity) SetShoppingRecord(list []*ShoppingRecordEntry) {
	e.ShoppingRecord = CloneShoppingRecordList(list)
	SortShoppingRecordList(e.ShoppingRecord)
}

func (s *ShoppingRecordEntry) ensureID() {
	if !tid.IsKindAndValid(s.ID, kinds.ShoppingRecord) {
		s.ID = tid.MustNewTID(kinds.ShoppingRecord)
	}
}

// normalizeShoppingRecord drops the missing entries of a record loaded from disk, gives an ID to each entry without a
// valid one, and puts the record in presentation order.
func normalizeShoppingRecord(list []*ShoppingRecordEntry) []*ShoppingRecordEntry {
	list = slices.DeleteFunc(list, func(entry *ShoppingRecordEntry) bool { return entry == nil })
	for _, entry := range list {
		entry.ensureID()
	}
	SortShoppingRecordList(list)
	return list
}
