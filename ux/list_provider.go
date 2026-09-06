// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"maps"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
)

// tagLister is implemented by the node types that carry tags.
type tagLister interface {
	TagList() []string
}

// listProvider is the part of a TableProvider that is the same for every kind of node: everything that only delegates
// to the list the rows come from, the owner of that list, and the header data for its columns. A provider embeds it
// and supplies the parts that differ between node types -- the reference and drag keys, the item names, the column
// IDs, the context menu, and how items are created and edited -- along with any of these methods it needs to do
// differently. The column IDs are taken through a callback, since they are the embedding provider's to decide.
type listProvider[T gurps.Node[T]] struct {
	table      *unison.Table[*Node[T]]
	dataOwner  gurps.DataOwnerProvider
	list       func() []T
	setList    func(list []T)
	columnIDs  func() []int
	headerData func(columnID int) gurps.HeaderData
	forPage    bool
}

// AllTags returns every tag found on the nodes in the list, at any depth, in natural order. Node types without tags
// yield nil.
func (p *listProvider[T]) AllTags() []string {
	set := make(map[string]struct{})
	gurps.Traverse(func(node T) bool {
		if tagged, ok := any(node).(tagLister); ok {
			for _, tag := range tagged.TagList() {
				set[tag] = struct{}{}
			}
		}
		return false
	}, false, false, p.list()...)
	return slices.SortedFunc(maps.Keys(set), func(a, b string) int { return xstrings.NaturalCmp(a, b, true) })
}

func (p *listProvider[T]) SetTable(table *unison.Table[*Node[T]]) {
	p.table = table
}

func (p *listProvider[T]) RootRowCount() int {
	return len(p.list())
}

func (p *listProvider[T]) RootRows() []*Node[T] {
	data := p.list()
	rows := make([]*Node[T], 0, len(data))
	for _, one := range data {
		rows = append(rows, NewNode(p.table, nil, one, p.forPage))
	}
	return rows
}

func (p *listProvider[T]) SetRootRows(rows []*Node[T]) {
	p.setList(ExtractNodeDataFromList(rows))
}

func (p *listProvider[T]) RootData() []T {
	return p.list()
}

func (p *listProvider[T]) SetRootData(data []T) {
	p.setList(data)
}

func (p *listProvider[T]) DataOwner() gurps.DataOwner {
	return p.dataOwner.DataOwner()
}

func (p *listProvider[T]) DropShouldMoveData(from, to *unison.Table[*Node[T]]) bool {
	return from == to
}

func (p *listProvider[T]) ProcessDropData(_, _ *unison.Table[*Node[T]]) {
}

func (p *listProvider[T]) AltDropSupport() *AltDropSupport {
	return nil
}

func (p *listProvider[T]) Headers() []unison.TableColumnHeader[*Node[T]] {
	ids := p.columnIDs()
	headers := make([]unison.TableColumnHeader[*Node[T]], 0, len(ids))
	for _, id := range ids {
		headers = append(headers, headerFromData[T](p.headerData(id), p.forPage))
	}
	return headers
}

func (p *listProvider[T]) SyncHeader(_ []unison.TableColumnHeader[*Node[T]]) {
}

func (p *listProvider[T]) Serialize() ([]byte, error) {
	return jio.SerializeAndCompress(p.list())
}

func (p *listProvider[T]) Deserialize(data []byte) error {
	var rows []T
	if err := jio.DecompressAndDeserialize(data, &rows); err != nil {
		return err
	}
	p.setList(rows)
	return nil
}

// insertItems adds the items to the list and the table, as InsertItems does, with the list's own accessors.
func (p *listProvider[T]) insertItems(owner Rebuildable, table *unison.Table[*Node[T]], items ...T) {
	InsertItems(owner, table, p.list, p.setList, func(_ *unison.Table[*Node[T]]) []*Node[T] { return p.RootRows() },
		items...)
}
