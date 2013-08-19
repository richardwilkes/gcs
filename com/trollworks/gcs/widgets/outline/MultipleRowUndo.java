/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2011 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

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

	@Override public void undo() throws CannotUndoException {
		super.undo();
		if (mUndos.length > 0) {
			for (int i = mUndos.length - 1; i != -1; i--) {
				mUndos[i].undo();
			}
			notifyDataFile();
		}
	}

	@Override public void redo() throws CannotRedoException {
		super.redo();
		if (mUndos.length > 0) {
			for (int i = 0; i != mUndos.length; i++) {
				mUndos[i].redo();
			}
			notifyDataFile();
		}
	}

	@Override public String getPresentationName() {
		if (mUndos.length == 0) {
			return super.getPresentationName();
		}
		return mUndos[0].getPresentationName();
	}
}
