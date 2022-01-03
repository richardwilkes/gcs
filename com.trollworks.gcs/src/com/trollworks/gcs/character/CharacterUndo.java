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

package com.trollworks.gcs.character;

import javax.swing.undo.AbstractUndoableEdit;
import javax.swing.undo.CannotRedoException;
import javax.swing.undo.CannotUndoException;

/** Provides undo support for character fields. */
public class CharacterUndo extends AbstractUndoableEdit {
    private GURPSCharacter  mCharacter;
    private String          mName;
    private CharacterSetter mSetter;
    private Object          mBefore;
    private Object          mAfter;

    /**
     * Create a new character field undo edit.
     *
     * @param character The character to provide an undo edit for.
     * @param name      The name of the undo edit.
     * @param setter    The field setter.
     * @param before    The original value.
     * @param after     The new value.
     */
    public CharacterUndo(GURPSCharacter character, String name, CharacterSetter setter, Object before, Object after) {
        mCharacter = character;
        mName = name;
        mSetter = setter;
        mBefore = before;
        mAfter = after;
    }

    @Override
    public String getPresentationName() {
        return mName;
    }

    @Override
    public void undo() throws CannotUndoException {
        super.undo();
        mSetter.setValue(mCharacter, mBefore);
    }

    @Override
    public void redo() throws CannotRedoException {
        super.redo();
        mSetter.setValue(mCharacter, mAfter);
    }
}
