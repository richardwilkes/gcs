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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

// PreservedTableData holds the data and selection state of a table in a serialized form.
type PreservedTableData[T gurps.Node[T]] struct {
	data   []byte
	selMap map[tid.TID]bool
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
	return nil
}

// Apply the data and selection state to a table, marking it as modified so that the change is reflected everywhere
// that shows the derived state.
func (d *PreservedTableData[T]) Apply(table *unison.Table[*Node[T]]) error {
	if err := d.Restore(table); err != nil {
		return err
	}
	MarkModified(table)
	return nil
}

// Restore puts the data and selection state back into a table without reporting the change to anything. Callers that
// go on to rebuild the table's owner should use this rather than Apply and leave the reporting to the rebuild, which
// recalculates the entity and re-syncs every table on its own; doing both would perform all of that work twice. Note
// that whether a rebuild is needed can only be determined once the data is back in place, since the set of columns a
// list requires depends on its content.
func (d *PreservedTableData[T]) Restore(table *unison.Table[*Node[T]]) error {
	provider, ok := table.ClientData()[TableProviderClientKey].(TableProvider[T])
	if !ok {
		return errs.New("unable to locate provider")
	}
	if err := provider.Deserialize(d.data); err != nil {
		return err
	}
	table.SyncToModel()
	table.SetSelectionMap(d.selMap)
	return nil
}
