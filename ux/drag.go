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
	"github.com/richardwilkes/unison/enums/paintstyle"
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
	editorRowDragKey           = unison.CreatePrivateDataType("gcs.editor-row")
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
	editorRowDragKey,
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

// flushDragFeedback marks the panel for redraw and immediately flushes the drawing. A native drag has no continuous
// redraw loop, so any drop feedback drawn in response to drag events must be flushed explicitly or it never appears.
// It is a variable so that headless tests, which have no window to observe drawing through, can intercept it.
var flushDragFeedback = func(panel *unison.Panel) {
	panel.MarkForRedraw()
	panel.FlushDrawing()
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
			flushDragFeedback(panel)
			return drag.Move
		}
		return drag.None
	}
	panel.DragEnteredCallback = update
	panel.DragUpdatedCallback = update
	panel.DragExitedCallback = func() {
		exit()
		flushDragFeedback(panel)
	}
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

// installDropRerouting makes a dockable's root panel accept drags for the supplied in-app drag data types and reroute
// the drag callbacks to the panel keyToPanel resolves the payload's data type to (typically the list that would hold a
// dropped item), so that a drop anywhere on the dockable lands in the right list. keyToPanel is consulted afresh for
// each drag update, since the lists may be rebuilt at any time; returning nil declines the drag. While a reroute is in
// progress, the target panel is highlighted with a translucent warning tint drawn over the root panel.
func installDropRerouting(panel *unison.Panel, keys []*uti.DataType, keyToPanel func(*uti.DataType) *unison.Panel) {
	var reroutePanel *unison.Panel
	panel.MouseDownCallback = func(_ geom.Point, _, _ int, _ mod.Modifiers) bool {
		panel.RequestFocus()
		return false
	}
	dragUpdate := func(di drag.Info, _ geom.Point, mods mod.Modifiers) drag.Op {
		reroutePanel = nil
		for _, key := range keys {
			if di.HasDataType(key.UTI) {
				if reroutePanel = keyToPanel(key); reroutePanel != nil {
					return reroutePanel.DragUpdatedCallback(di, geom.Point{Y: 100000000}, mods)
				}
				break
			}
		}
		return drag.None
	}
	panel.CanAcceptDropCallback = func(di drag.Info) bool { return hasAnyDragDataType(di, keys...) }
	panel.DragEnteredCallback = dragUpdate
	panel.DragUpdatedCallback = dragUpdate
	panel.DragExitedCallback = func() {
		if reroutePanel != nil {
			target := reroutePanel
			reroutePanel = nil
			if target.DragExitedCallback != nil {
				target.DragExitedCallback()
			}
		}
	}
	panel.DropCallback = func(di drag.Info, _ geom.Point, mods mod.Modifiers) bool {
		handled := false
		if reroutePanel != nil {
			target := reroutePanel
			reroutePanel = nil
			if target.DropCallback != nil {
				handled = target.DropCallback(di, geom.Point{Y: 100000000}, mods)
			}
		}
		return handled
	}
	panel.DrawOverCallback = func(gc *unison.Canvas, _ geom.Rect) {
		if reroutePanel != nil {
			r := panel.RectFromRoot(reroutePanel.RectToRoot(reroutePanel.ContentRect(true)))
			paint := unison.ThemeWarning.Paint(gc, r, paintstyle.Fill)
			paint.SetColorFilter(unison.Alpha30Filter())
			gc.DrawRect(r, paint)
		}
	}
}
