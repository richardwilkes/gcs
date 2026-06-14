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
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/drag"
	"github.com/richardwilkes/unison/enums/imgfmt"
	"github.com/richardwilkes/unison/enums/mod"
)

// Drag & drop data types for the in-app drag payloads GCS supports. Each is a private data type, unique to this
// instance of the application, used to identify what is being dragged.
var (
	traitDragKey               = unison.CreatePrivateDataType("gcs.trait")
	traitModifierDragKey       = unison.CreatePrivateDataType("gcs.trait-modifier")
	equipmentDragKey           = unison.CreatePrivateDataType("gcs.equipment")
	equipmentModifierDragKey   = unison.CreatePrivateDataType("gcs.equipment-modifier")
	noteDragKey                = unison.CreatePrivateDataType("gcs.note")
	skillDragKey               = unison.CreatePrivateDataType("gcs.skill")
	spellDragKey               = unison.CreatePrivateDataType("gcs.spell")
	reactionModifierDragKey    = unison.CreatePrivateDataType("gcs.reaction-modifier")
	conditionalModifierDragKey = unison.CreatePrivateDataType("gcs.conditional-modifier")
	meleeWeaponDragKey         = unison.CreatePrivateDataType("gcs.melee-weapon")
	rangedWeaponDragKey        = unison.CreatePrivateDataType("gcs.ranged-weapon")
	attributeSettingsDragKey   = unison.CreatePrivateDataType("gcs.attr")
	hitLocationDragKey         = unison.CreatePrivateDataType("gcs.body")
)

var (
	// panelDragData holds the data for an in-progress panel-based drag (initiated by a DragHandle). It is refreshed at
	// the start of each drag.
	panelDragData any
	// draggedTableData holds the data for an in-progress table row drag. It mirrors the data unison tracks internally
	// so that our alternate drop handlers can access the dragged rows. It is refreshed at the start of each drag.
	draggedTableData any
)

// allDragDataTypes is the complete set of in-app drag data types GCS uses for drag & drop. Every window registers all
// of them because any dockable (and therefore any of its drop targets) may be hosted by any window, and the current
// unison API only delivers drops for the data types a window has registered for.
var allDragDataTypes = []*uti.DataType{
	traitDragKey,
	traitModifierDragKey,
	equipmentDragKey,
	equipmentModifierDragKey,
	noteDragKey,
	skillDragKey,
	spellDragKey,
	reactionModifierDragKey,
	conditionalModifierDragKey,
	meleeWeaponDragKey,
	rangedWeaponDragKey,
	attributeSettingsDragKey,
	hitLocationDragKey,
}

// registerWindowDragTypes registers the supplied window as a target for every kind of drag payload GCS supports: all of
// the in-app drag keys plus readable image files and URLs (the latter being required for OS-level file/URL drops, e.g.
// dropping an image onto the portrait panel, to be delivered).
func registerWindowDragTypes(wnd *unison.Window) {
	if wnd == nil {
		return
	}
	imgUTIs := imgfmt.AllReadableUTIs()
	dockUTIs := unison.DockDragTypes()
	types := make([]*uti.DataType, 0, len(allDragDataTypes)+len(imgUTIs)+2+len(dockUTIs))
	types = append(types, allDragDataTypes...)
	types = append(types, imgUTIs...)
	types = append(types, uti.FileURL, uti.URL)
	types = append(types, dockUTIs...)
	wnd.RegisterForDragTypes(types...)
}

// installPanelDragDrop wires the supplied panel-based drag handlers to a panel using the current unison drag callbacks.
// The over and drop handlers receive the in-progress panel drag data, keyed by drag data type.
func installPanelDragDrop(panel *unison.Panel, dataType *uti.DataType,
	over func(where geom.Point, data any) bool,
	exit func(),
	drop func(where geom.Point, data any),
) {
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

// hasAnyDragDataType reports whether the drag carries a payload for any of the supplied in-app drag data types. It is
// used by container panels to decline drags they don't handle (e.g. a dock tab drag) so that the drop can propagate to
// the underlying dock.
func hasAnyDragDataType(di drag.Info, dataTypes ...*uti.DataType) bool {
	for _, dt := range dataTypes {
		if di.HasDataType(dt.UTI) {
			return true
		}
	}
	return false
}
