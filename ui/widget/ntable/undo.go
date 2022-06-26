/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ntable

import (
	"github.com/google/uuid"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/unison"
)

type tableDragUndoEditData[T gurps.NodeConstraint[T]] struct {
	From       *unison.Table[*Node[T]]
	To         *unison.Table[*Node[T]]
	FromData   []byte
	FromSelMap map[uuid.UUID]bool
	ToData     []byte
	ToSelMap   map[uuid.UUID]bool
	Move       bool
}

func newTableDragUndoEditData[T gurps.NodeConstraint[T]](from, to *unison.Table[*Node[T]], move bool) *tableDragUndoEditData[T] {
	data, err := collectTableData(to)
	if err != nil {
		jot.Error(err)
		return nil
	}
	undo := &tableDragUndoEditData[T]{
		From:     from,
		To:       to,
		ToData:   data,
		ToSelMap: to.CopySelectionMap(),
		Move:     move,
	}
	if move && from != to {
		if data, err = collectTableData(from); err != nil {
			jot.Error(err)
			return nil
		}
		undo.FromData = data
		undo.FromSelMap = from.CopySelectionMap()
	}
	return undo
}

func (t *tableDragUndoEditData[T]) apply() {
	applyTableData(t.To, t.ToData, t.ToSelMap)
	if t.Move && t.From != t.To {
		applyTableData(t.From, t.FromData, t.FromSelMap)
	}
}

func collectTableData[T gurps.NodeConstraint[T]](table *unison.Table[*Node[T]]) ([]byte, error) {
	provider, ok := table.ClientData()[tableProviderClientKey].(TableProvider[T])
	if !ok {
		return nil, errs.New("unable to locate provider")
	}
	return provider.Serialize()
}

func applyTableData[T gurps.NodeConstraint[T]](table *unison.Table[*Node[T]], data []byte, selMap map[uuid.UUID]bool) {
	provider, ok := table.ClientData()[tableProviderClientKey].(TableProvider[T])
	if !ok {
		jot.Error(errs.New("unable to locate provider"))
		return
	}
	if err := provider.Deserialize(data); err != nil {
		jot.Error(err)
		return
	}
	table.SyncToModel()
	widget.MarkModified(table)
	table.SetSelectionMap(selMap)
}
