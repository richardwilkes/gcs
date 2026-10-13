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
)

// ShoppingKind identifies whether a shopping record entry is a sale or a purchase.
type ShoppingKind string

// Shopping kinds.
const (
	ShoppingSold   ShoppingKind = "sold"
	ShoppingBought ShoppingKind = "bought"
)

// ShoppingRecordItem is a snapshot of an item, or of a money item, as it was when a transaction was made. It keeps no
// link back to the equipment, so later changes to the sheet don't alter the record.
type ShoppingRecordItem struct {
	Name     string  `json:"name"`
	Quantity fxp.Int `json:"quantity"`
	Value    fxp.Int `json:"value"`
}

// ShoppingRecordEntry holds one sale or purchase made from this sheet.
type ShoppingRecordEntry struct {
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
// which is most recent first.
func SortShoppingRecordList(list []*ShoppingRecordEntry) {
	slices.SortFunc(list, func(a, b *ShoppingRecordEntry) int { return b.When.Compare(a.When) })
}

// AddShoppingRecord adds an entry to this sheet's shopping record, keeping the record in presentation order.
func (e *Entity) AddShoppingRecord(entry *ShoppingRecordEntry) {
	e.ShoppingRecord = append(e.ShoppingRecord, entry)
	SortShoppingRecordList(e.ShoppingRecord)
}
