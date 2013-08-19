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

/** Provides coalescing of multiple undos into a single undo. */
public class TKCompoundUndo extends TKSimpleUndo {
	/** The edits to be undone/redone. */
	protected ArrayList<TKUndo>	mEdits;
	private boolean				mInProgress;

	/** Creates a new compound edit. */
	public TKCompoundUndo() {
		super();
		mInProgress = true;
		mEdits = new ArrayList<TKUndo>();
	}

	/**
	 * The last edit added is given a chance to absorb the new edit. If it refuses (returns
	 * <code>false</code> from a call to {@link #addEdit(TKUndo)}), the new edit is given a
	 * chance to replace the last edit. If it refuses (returns <code>false</code> from a call to
	 * {@link #replaceEdit(TKUndo)}), then the new edit is added to the list of edits maintained by
	 * this compound edit.
	 * 
	 * @param edit The edit to add.
	 * @return <code>true</code> if the edit is progress and the passed in edit could be absorbed.
	 */
	@Override public boolean addEdit(TKUndo edit) {
		if (mInProgress) {
			TKUndo last = lastEdit();

			if (last == null) {
				mEdits.add(edit);
			} else if (!last.addEdit(edit)) {
				if (edit.replaceEdit(last)) {
					mEdits.remove(mEdits.size() - 1);
				}
				mEdits.add(edit);
			}
			return true;
		}
		return false;
	}

	@Override public boolean canApply(boolean forUndo) {
		return !isInProgress() && super.canApply(forUndo);
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

	/** Ends the ability of this compound edit to absorb more edits. */
	public void end() {
		mInProgress = false;
	}

	@Override public String getName() {
		TKUndo last = lastEdit();

		return last != null ? last.getName() : super.getName();
	}

	@Override public String getName(boolean forUndo) {
		TKUndo last = lastEdit();

		return last != null ? last.getName(forUndo) : super.getName(forUndo);
	}

	/** @return <code>true</code> if this edit is in progress. */
	public boolean isInProgress() {
		return mInProgress;
	}

	@Override public boolean isSignificant() {
		for (TKUndo edit : mEdits) {
			if (edit.isSignificant()) {
				return true;
			}
		}
		return false;
	}

	/** @return The last edit or <code>null</code> if there is none. */
	public TKUndo lastEdit() {
		int length = mEdits.size();

		return length > 0 ? mEdits.get(length - 1) : null;
	}
}
