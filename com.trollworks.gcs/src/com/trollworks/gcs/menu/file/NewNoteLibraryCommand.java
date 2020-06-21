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

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.library.LibraryExplorerDockable;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.notes.NoteList;
import com.trollworks.gcs.notes.NotesDockable;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;

/** Provides the "New Note Library" command. */
public class NewNoteLibraryCommand extends Command {
    /** The action command this command will issue. */
    public static final String                CMD_NEW_LIBRARY = "NewNoteLibrary";
    /** The singleton {@link NewNoteLibraryCommand}. */
    public static final NewNoteLibraryCommand INSTANCE        = new NewNoteLibraryCommand();

    private NewNoteLibraryCommand() {
        super(I18n.Text("New Note Library"), CMD_NEW_LIBRARY);
    }

    @Override
    public void adjust() {
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        newNoteLibrary();
    }

    /** @return The newly created a new {@link NotesDockable}. */
    public static NotesDockable newNoteLibrary() {
        LibraryExplorerDockable library = LibraryExplorerDockable.get();
        if (library != null) {
            NoteList list = new NoteList();
            list.getModel().setLocked(false);
            NotesDockable dockable = new NotesDockable(list);
            library.dockLibrary(dockable);
            return dockable;
        }
        return null;
    }
}
