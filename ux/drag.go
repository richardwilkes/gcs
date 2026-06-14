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
	"sync"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/drag"
	"github.com/richardwilkes/unison/enums/imgfmt"
	"github.com/richardwilkes/unison/enums/mod"
)

var (
	dragDataTypesLock sync.Mutex
	dragDataTypes     = make(map[string]*uti.DataType)

	// panelDragData holds the data for an in-progress panel-based drag (initiated by a DragHandle). It is keyed by the
	// drag key and refreshed at the start of each drag.
	panelDragData map[string]any

	// draggedTableData holds the data for an in-progress table row drag. It mirrors the data unison tracks internally so
	// that our alternate drop handlers can access the dragged rows. It is refreshed at the start of each drag.
	draggedTableData any
)

// dragDataType returns the registered uti.DataType for the given drag key, registering it on first use. The previous
// versions of unison used bare strings to identify drag payloads; the current API uses uti.DataType, so we map our
// keys onto stable, lazily-registered data types.
func dragDataType(key string) *uti.DataType {
	dragDataTypesLock.Lock()
	defer dragDataTypesLock.Unlock()
	dt, ok := dragDataTypes[key]
	if !ok {
		dt = uti.Register(&uti.DataType{UTI: "private.gcs.drag." + key})
		dragDataTypes[key] = dt
	}
	return dt
}

// allDragDataKeys is the complete set of in-app drag keys GCS uses for drag & drop. Every window registers all of them
// because any dockable (and therefore any of its drop targets) may be hosted by any window, and the current unison API
// only delivers drops for the data types a window has registered for.
var allDragDataKeys = []string{
	traitDragKey,
	traitModifierDragKey,
	equipmentDragKey,
	equipmentModifierDragKey,
	noteDragKey,
	gurps.SkillID,
	gurps.SpellID,
	reactionModifierDragKey,
	conditionalModifierDragKey,
	meleeWeaponDragKey,
	rangedWeaponDragKey,
	attributeSettingsDragDataKey,
	hitLocationDragDataKey,
}

// registerWindowDragTypes registers the supplied window as a target for every kind of drag payload GCS supports: all of
// the in-app drag keys plus readable image files and URLs (the latter being required for OS-level file/URL drops, e.g.
// dropping an image onto the portrait panel, to be delivered).
func registerWindowDragTypes(wnd *unison.Window) {
	if wnd == nil {
		return
	}
	types := make([]*uti.DataType, 0, len(allDragDataKeys)+len(imgfmt.AllReadableUTIs())+2)
	for _, key := range allDragDataKeys {
		types = append(types, dragDataType(key))
	}
	types = append(types, imgfmt.AllReadableUTIs()...)
	types = append(types, uti.FileURL, uti.URL)
	types = append(types, unison.DockDragTypes()...)
	wnd.RegisterForDragTypes(types...)
}

// installPanelDragDrop wires the supplied panel-based drag handlers to a panel using the current unison drag callbacks.
// The over, exit, and drop handlers retain their map[string]any payload signatures; the payload is supplied from the
// in-progress panel drag data.
func installPanelDragDrop(panel *unison.Panel, key string,
	over func(where geom.Point, data map[string]any) bool,
	exit func(),
	drop func(where geom.Point, data map[string]any),
) {
	dataType := dragDataType(key)
	panel.CanAcceptDropCallback = func(di drag.Info) bool { return di.HasDataType(dataType.UTI) }
	update := func(di drag.Info, where geom.Point, _ mod.Modifiers) drag.Op {
		if di.HasDataType(dataType.UTI) {
			over(where, panelDragData)
			return drag.Move
		}
		return drag.None
	}
	panel.DragEnteredCallback = update
	panel.DragUpdatedCallback = update
	panel.DragExitedCallback = exit
	panel.DropCallback = func(di drag.Info, where geom.Point, _ mod.Modifiers) bool {
		if di.HasDataType(dataType.UTI) {
			drop(where, panelDragData)
			return true
		}
		return false
	}
}

// hasAnyDragDataKey reports whether the drag carries a payload for any of the supplied in-app drag keys. It is used by
// container panels to decline drags they don't handle (e.g. a dock tab drag) so that the drop can propagate to the
// underlying dock.
func hasAnyDragDataKey(di drag.Info, keys ...string) bool {
	for _, key := range keys {
		if di.HasDataType(dragDataType(key).UTI) {
			return true
		}
	}
	return false
}
