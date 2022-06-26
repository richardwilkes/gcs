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

package editors

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/unison"
)

// EditNote displays the editor for a note.
func EditNote(owner widget.Rebuildable, note *gurps.Note) {
	displayEditor[*gurps.Note, *gurps.NoteEditData](owner, note, res.GCSNotesSVG, initNoteEditor)
}

func initNoteEditor(e *editor[*gurps.Note, *gurps.NoteEditData], content *unison.Panel) func() {
	addNotesLabelAndField(content, &e.editorData.Text)
	addPageRefLabelAndField(content, &e.editorData.PageRef)
	return nil
}
