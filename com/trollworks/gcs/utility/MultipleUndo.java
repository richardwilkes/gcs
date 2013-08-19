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

package com.trollworks.gcs.utility;

import javax.swing.UIManager;
import javax.swing.undo.CompoundEdit;

/** Provides a convenient way to collect multiple undos into a single undo. */
public class MultipleUndo extends CompoundEdit {
	private static final String	SPACE		= " ";															//$NON-NLS-1$
	private static final String	REDO_PREFIX	= UIManager.getString("AbstractUndoableEdit.redoText") + SPACE; //$NON-NLS-1$
	private static final String	UNDO_PREFIX	= UIManager.getString("AbstractUndoableEdit.undoText") + SPACE; //$NON-NLS-1$
	private String				mName;

	/**
	 * Create a multiple undo edit.
	 * 
	 * @param name The name of the undo edit.
	 */
	public MultipleUndo(String name) {
		super();
		mName = name;
	}

	@Override public String getPresentationName() {
		return mName;
	}

	@Override public String getRedoPresentationName() {
		return REDO_PREFIX + mName;
	}

	@Override public String getUndoPresentationName() {
		return UNDO_PREFIX + mName;
	}
}
