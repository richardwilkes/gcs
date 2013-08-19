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
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.model;

import com.trollworks.toolkit.undo.TKSimpleUndo;
import com.trollworks.toolkit.undo.TKUndoException;

import java.util.Collection;

/** An undo that contains one or more {@link CMRowUndo}s. */
public class CMMultipleRowUndo extends TKSimpleUndo {
	private CMRowUndo[]	mUndos;

	/**
	 * Creates a new {@link CMMultipleRowUndo}.
	 * 
	 * @param undos The {@link CMRowUndo}s to manage.
	 */
	public CMMultipleRowUndo(Collection<CMRowUndo> undos) {
		super();
		mUndos = undos.toArray(new CMRowUndo[0]);
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

	@Override public void apply(boolean forUndo) throws TKUndoException {
		super.apply(forUndo);

		if (mUndos.length > 0) {
			int increment;
			int start;
			int finish;

			if (forUndo) {
				increment = -1;
				start = mUndos.length - 1;
				finish = -1;
			} else {
				increment = 1;
				start = 0;
				finish = mUndos.length;
			}
			for (int i = start; i != finish; i += increment) {
				mUndos[i].apply(forUndo);
			}
			notifyDataFile();
		}
	}

	@Override public String getName() {
		if (mUndos.length == 0) {
			return super.getName();
		}
		return mUndos[0].getName();
	}
}
