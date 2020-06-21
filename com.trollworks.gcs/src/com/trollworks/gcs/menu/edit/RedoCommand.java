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

package com.trollworks.gcs.menu.edit;

import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import javax.swing.undo.UndoManager;

/** Provides the "Redo" command. */
public class RedoCommand extends Command {
    /** The action command this command will issue. */
    public static final String CMD_REDO = "Redo";

    /** The singleton {@link RedoCommand}. */
    public static final RedoCommand INSTANCE = new RedoCommand();

    private RedoCommand() {
        super(I18n.Text("Can't Redo"), CMD_REDO, KeyEvent.VK_Y);
    }

    @Override
    public void adjust() {
        Undoable undoable = getTarget(Undoable.class);
        if (undoable != null) {
            UndoManager mgr = undoable.getUndoManager();
            setEnabled(mgr.canRedo());
            setTitle(mgr.getRedoPresentationName());
        } else {
            setEnabled(false);
            setTitle(I18n.Text("Can't Redo"));
        }
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Undoable undoable = getTarget(Undoable.class);
        if (undoable != null) {
            undoable.getUndoManager().redo();
        }
    }
}
