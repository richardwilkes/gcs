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

/** Exception thrown when an undo or redo cannot be completed. */
public class TKUndoException extends RuntimeException {
	private boolean	mWasUndo;

	/**
	 * Create a new undo/redo exception.
	 * 
	 * @param wasUndo Pass in <code>true</code> if this was an undo, <code>false</code> if it
	 *            was a redo.
	 */
	public TKUndoException(boolean wasUndo) {
		super();
		mWasUndo = wasUndo;
	}

	/**
	 * Create a new undo/redo exception.
	 * 
	 * @param wasUndo Pass in <code>true</code> if this was an undo, <code>false</code> if it
	 *            was a redo.
	 * @param msg A message describing the failure.
	 */
	public TKUndoException(boolean wasUndo, String msg) {
		super(msg);
		mWasUndo = wasUndo;
	}

	/** @return <code>true</code> if this was an undo that failed. */
	public boolean wasUndo() {
		return mWasUndo;
	}

	/** @return <code>true</code> if this was an redo that failed. */
	public boolean wasRedo() {
		return !mWasUndo;
	}
}
