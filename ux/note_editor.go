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

package ux

import (
	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/unison"
)

// EditNote displays the editor for a note.
func EditNote(owner Rebuildable, note *model.Note) {
	displayEditor[*model.Note, *model.NoteEditData](owner, note, svg.GCSNotes, initNoteEditor)
}

func initNoteEditor(e *editor[*model.Note, *model.NoteEditData], content *unison.Panel) func() {
	addNotesLabelAndField(content, &e.editorData.Text)
	addPageRefLabelAndField(content, &e.editorData.PageRef)
	return nil
}
