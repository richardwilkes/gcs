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

import java.util.ArrayList;

/** Provides a convenient way to collect multiple undos into a single undo. */
public class TKMultipleUndo extends TKSimpleUndo {
	private String				mName;
	private boolean				mInProgress;
	private ArrayList<TKUndo>	mEdits;

	/**
	 * Create a multiple undo edit.
	 * 
	 * @param name The name of the undo edit.
	 */
	public TKMultipleUndo(String name) {
		super();
		mName = name;
		mInProgress = true;
		mEdits = new ArrayList<TKUndo>();
	}

	@Override public boolean addEdit(TKUndo edit) {
		if (mInProgress) {
			mEdits.add(edit);
			return true;
		}
		return false;
	}

	@Override public boolean canApply(boolean forUndo) {
		return !mInProgress && super.canApply(forUndo);
	}

	@Override public void apply(boolean forUndo) throws TKUndoException {
		super.apply(forUndo);

		int size = mEdits.size();
		int start = forUndo ? size - 1 : 0;
		int end = forUndo ? -1 : size;
		int inc = forUndo ? -1 : 1;

		for (int i = start; i != end; i += inc) {
			mEdits.get(i).apply(forUndo);
		}
	}

	@Override public void obsolete() {
		int i = mEdits.size();

		while (i-- > 0) {
			mEdits.get(i).obsolete();
		}
		super.obsolete();
	}

	/** Ends the ability of this multiple edit to absorb more edits. */
	public void end() {
		mInProgress = false;
	}

	/** @return Whether this undo can still absorb more undos. */
	public boolean isInProgress() {
		return mInProgress;
	}

	@Override public boolean isSignificant() {
		return !mEdits.isEmpty();
	}

	@Override public String getName() {
		return mName;
	}
}
