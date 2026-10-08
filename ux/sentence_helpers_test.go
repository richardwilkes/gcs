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
	"cmp"

	"github.com/richardwilkes/unison"
)

// menuAction returns the action of the entry with the label, or nil.
func menuAction(entries []menuEntry, label string) func() {
	for _, one := range entries {
		if one.Label == label {
			return one.Act
		}
	}
	return nil
}

// menuLabels returns the labels of the entries, with "-" for a separator.
func menuLabels(entries []menuEntry) []string {
	labels := make([]string, 0, len(entries))
	for _, one := range entries {
		labels = append(labels, cmp.Or(one.Label, "-"))
	}
	return labels
}

// sentenceUndoHost stands in for the editor that holds a prerequisites or features panel: it provides the undo manager,
// syncs when marked modified, and records the Escape that would discard the editor's changes.
type sentenceUndoHost struct {
	unison.Panel
	mgr      *unison.UndoManager
	escapes  int
	modified int
}

func (h *sentenceUndoHost) UndoManager() *unison.UndoManager {
	return h.mgr
}

// MarkModified implements ModifiableRoot, counting the calls and syncing as the editor does.
func (h *sentenceUndoHost) MarkModified(_ unison.Paneler) {
	h.modified++
	DeepSync(h)
}
