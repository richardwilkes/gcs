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

package com.trollworks.toolkit.undo;

import java.text.MessageFormat;

/** A basic undoable edit. */
public class TKSimpleUndo implements TKUndo {
	private boolean	mHasBeenDone;
	private boolean	mAlive;

	/** Creates a basic undoable edit. */
	public TKSimpleUndo() {
		mHasBeenDone = true;
		mAlive = true;
	}

	/** Always returns <code>false</code>. {@inheritDoc} */
	public boolean addEdit(TKUndo edit) {
		return false;
	}

	/** Always returns <code>false</code>. {@inheritDoc} */
	public boolean replaceEdit(TKUndo edit) {
		return false;
	}

	public boolean canApply(boolean forUndo) {
		return mAlive && mHasBeenDone == forUndo;
	}

	public void apply(boolean forUndo) throws TKUndoException {
		if (!canApply(forUndo)) {
			throw new TKUndoException(forUndo);
		}
		mHasBeenDone = !forUndo;
	}

	public void obsolete() {
		mAlive = false;
	}

	public String getName() {
		return ""; //$NON-NLS-1$
	}

	public String getName(boolean forUndo) {
		String undoRedo = forUndo ? Msgs.UNDO : Msgs.REDO;
		String name = getName();

		if (name == null || name.length() == 0) {
			return undoRedo;
		}

		return MessageFormat.format(Msgs.FORMAT, undoRedo, name);
	}

	/** Always returns <code>true</code>. {@inheritDoc} */
	public boolean isSignificant() {
		return true;
	}
}
