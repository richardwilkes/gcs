/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.menu.item;

import com.trollworks.gcs.character.SheetDockable;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.notes.Note;
import com.trollworks.gcs.notes.NotesDockable;
import com.trollworks.gcs.template.TemplateDockable;
import com.trollworks.gcs.ui.widget.outline.ListOutline;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;
import java.awt.event.InputEvent;
import java.awt.event.KeyEvent;

/** Provides the "New Note" command. */
public class NewNoteCommand extends Command {
    /** The action command this command will issue. */
    public static final String         CMD_NEW_NOTE           = "NewNote";
    /** The action command this command will issue. */
    public static final String         CMD_NEW_NOTE_CONTAINER = "NewNoteContainer";
    /** The "New Note" command. */
    public static final NewNoteCommand INSTANCE               = new NewNoteCommand(false, I18n.Text("New Note"), CMD_NEW_NOTE, KeyEvent.VK_N, SHIFTED_COMMAND_MODIFIER);
    /** The "New Note Container" command. */
    public static final NewNoteCommand CONTAINER_INSTANCE     = new NewNoteCommand(true, I18n.Text("New Note Container"), CMD_NEW_NOTE_CONTAINER, KeyEvent.VK_N, COMMAND_MODIFIER | InputEvent.ALT_DOWN_MASK);
    private             boolean        mContainer;

    private NewNoteCommand(boolean container, String title, String cmd, int keyCode, int modifiers) {
        super(title, cmd, keyCode, modifiers);
        mContainer = container;
    }

    @Override
    public void adjust() {
        NotesDockable note = getTarget(NotesDockable.class);
        if (note != null) {
            setEnabled(!note.getOutline().getModel().isLocked());
        } else {
            SheetDockable sheet = getTarget(SheetDockable.class);
            if (sheet != null) {
                setEnabled(true);
            } else {
                setEnabled(getTarget(TemplateDockable.class) != null);
            }
        }
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        ListOutline   outline;
        DataFile      dataFile;
        NotesDockable eqpDockable = getTarget(NotesDockable.class);
        if (eqpDockable != null) {
            dataFile = eqpDockable.getDataFile();
            outline = eqpDockable.getOutline();
            if (outline.getModel().isLocked()) {
                return;
            }
        } else {
            SheetDockable sheet = getTarget(SheetDockable.class);
            if (sheet != null) {
                dataFile = sheet.getDataFile();
                outline = sheet.getSheet().getNoteOutline();
            } else {
                TemplateDockable template = getTarget(TemplateDockable.class);
                if (template != null) {
                    dataFile = template.getDataFile();
                    outline = template.getTemplate().getNoteOutline();
                } else {
                    return;
                }
            }
        }
        Note note = new Note(dataFile, mContainer);
        outline.addRow(note, getTitle(), false);
        outline.getModel().select(note, false);
        outline.scrollSelectionIntoView();
        outline.openDetailEditor(true);
    }
}
