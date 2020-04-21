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

package com.trollworks.gcs.utility.undo;

import com.trollworks.gcs.menu.edit.Undoable;
import com.trollworks.gcs.ui.UIUtilities;

import java.awt.Component;
import javax.swing.undo.CannotRedoException;
import javax.swing.undo.CannotUndoException;
import javax.swing.undo.UndoManager;
import javax.swing.undo.UndoableEdit;

/** The standard {@link UndoManager} for use with our app's windows. */
public class StdUndoManager extends UndoManager {
    private boolean mInTransaction;

    @Override
    public synchronized void undo() throws CannotUndoException {
        mInTransaction = true;
        super.undo();
        mInTransaction = false;
    }

    @Override
    public synchronized void redo() throws CannotRedoException {
        mInTransaction = true;
        super.redo();
        mInTransaction = false;
    }

    /** @return Whether this {@link UndoManager} is currently processing an undo or redo. */
    public boolean isInTransaction() {
        return mInTransaction;
    }

    /**
     * Post an undoable edit to the undo manager responsible for the component.
     *
     * @param comp The component.
     * @param edit The undoable edit.
     */
    public static final void post(Component comp, UndoableEdit edit) {
        if (comp != null) {
            Undoable undoable = UIUtilities.getAncestorOfType(comp, Undoable.class);
            if (undoable != null) {
                undoable.getUndoManager().addEdit(edit);
            }
        }
    }
}
