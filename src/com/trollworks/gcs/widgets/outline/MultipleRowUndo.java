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

package com.trollworks.gcs.widgets.outline;

import java.util.Collection;

import javax.swing.undo.AbstractUndoableEdit;
import javax.swing.undo.CannotRedoException;
import javax.swing.undo.CannotUndoException;

/** An undo that contains one or more {@link RowUndo}s. */
public class MultipleRowUndo extends AbstractUndoableEdit {
	private RowUndo[]	mUndos;

	/**
	 * Creates a new {@link MultipleRowUndo}.
	 * 
	 * @param undos The {@link RowUndo}s to manage.
	 */
	public MultipleRowUndo(Collection<RowUndo> undos) {
		super();
		mUndos = undos.toArray(new RowUndo[0]);
		if (mUndos.length > 0) {
			mUndos[0].getDataFile().addEdit(this);
		}
		notifyDataFile();
	}

	private void notifyDataFile() {
		if (mUndos.length > 0) {
			mUndos[0].getDataFile().notifySingle(mUndos[0].getRow().getListChangedID(), null);
		}
	}

	@Override
	public void undo() throws CannotUndoException {
		super.undo();
		if (mUndos.length > 0) {
			for (int i = mUndos.length - 1; i != -1; i--) {
				mUndos[i].undo();
			}
			notifyDataFile();
		}
	}

	@Override
	public void redo() throws CannotRedoException {
		super.redo();
		if (mUndos.length > 0) {
			for (int i = 0; i != mUndos.length; i++) {
				mUndos[i].redo();
			}
			notifyDataFile();
		}
	}

	@Override
	public String getPresentationName() {
		if (mUndos.length == 0) {
			return super.getPresentationName();
		}
		return mUndos[0].getPresentationName();
	}
}
