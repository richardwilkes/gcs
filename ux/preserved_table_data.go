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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

// PreservedTableData holds the data and selection state of a table in a serialized form, along with the objects the
// data was taken from, so that putting it back can reuse them (see Restore).
type PreservedTableData[T gurps.Node[T]] struct {
	data       []byte
	selMap     map[tid.TID]bool
	identities gurps.NodeIdentities
}

// Collect the data and selection state from a table.
func (d *PreservedTableData[T]) Collect(table *unison.Table[*Node[T]]) error {
	provider, ok := table.ClientData()[TableProviderClientKey].(TableProvider[T])
	if !ok {
		return errs.New("unable to locate provider")
	}
	data, err := provider.Serialize()
	if err != nil {
		return err
	}
	d.data = data
	d.selMap = table.CopySelectionMap()
	d.identities = make(gurps.NodeIdentities)
	gurps.CollectNodeIdentities(d.identities, provider.RootData())
	return nil
}

// Restore puts the data and selection state back into a table without reporting the change to anything. The reporting
// is left to the caller, which rebuilds the table's owner -- or, for a table with no owner, marks the table as
// modified -- but only one of the two, since a rebuild already recalculates the entity and re-syncs every table. That
// lets an undo which has to put several tables back report the change once for all of them (see restoredTables).
//
// The nodes come back as the objects the table already holds for their IDs or, for those it no longer holds, the
// objects the data was collected from, rather than the fresh ones deserializing makes. Other undo edits on the stack
// -- an editor's, a toggle's -- change the objects they were made with, and would otherwise be changing objects no
// list contains once an undo or redo of this edit had passed. The table's own objects come first so that the restore
// converges on what is on screen.
func (d *PreservedTableData[T]) Restore(table *unison.Table[*Node[T]]) error {
	provider, ok := table.ClientData()[TableProviderClientKey].(TableProvider[T])
	if !ok {
		return errs.New("unable to locate provider")
	}
	identities := make(gurps.NodeIdentities, len(d.identities))
	maps.Copy(identities, d.identities)
	gurps.CollectNodeIdentities(identities, provider.RootData())
	if err := provider.Deserialize(d.data); err != nil {
		return err
	}
	provider.SetRootData(gurps.RecycleNodes(identities, provider.RootData()))
	table.SyncToModel()
	table.SetSelectionMap(d.selMap)
	return nil
}
