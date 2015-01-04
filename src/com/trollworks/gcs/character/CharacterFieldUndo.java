/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character;

import javax.swing.undo.AbstractUndoableEdit;
import javax.swing.undo.CannotRedoException;
import javax.swing.undo.CannotUndoException;

/** Provides undo support for character fields. */
public class CharacterFieldUndo extends AbstractUndoableEdit {
	private GURPSCharacter	mCharacter;
	private String			mName;
	private String			mID;
	private Object			mBefore;
	private Object			mAfter;

	/**
	 * Create a new character field undo edit.
	 * 
	 * @param character The character to provide an undo edit for.
	 * @param name The name of the undo edit.
	 * @param id The ID of the field being changed.
	 * @param before The original value.
	 * @param after The new value.
	 */
	public CharacterFieldUndo(GURPSCharacter character, String name, String id, Object before, Object after) {
		super();
		mCharacter = character;
		mName = name;
		mID = id;
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
		mCharacter.setValueForID(mID, mBefore);
	}

	@Override
	public void redo() throws CannotRedoException {
		super.redo();
		mCharacter.setValueForID(mID, mAfter);
	}
}
