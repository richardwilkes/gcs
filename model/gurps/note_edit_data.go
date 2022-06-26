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

package gurps

var _ EditorData[*Note] = &NoteEditData{}

// NoteEditData holds the Note data that can be edited by the UI detail editor.
type NoteEditData struct {
	Text    string `json:"text,omitempty"`
	PageRef string `json:"reference,omitempty"`
}

// CopyFrom implements node.EditorData.
func (d *NoteEditData) CopyFrom(note *Note) {
	d.copyFrom(&note.NoteEditData)
}

// ApplyTo implements node.EditorData.
func (d *NoteEditData) ApplyTo(note *Note) {
	note.NoteEditData.copyFrom(d)
}

func (d *NoteEditData) copyFrom(other *NoteEditData) {
	*d = *other
}
