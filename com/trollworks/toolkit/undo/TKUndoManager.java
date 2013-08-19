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

/** Provides an undo manager. */
public class TKUndoManager extends TKCompoundUndo {
	private int								mIndexOfNextAdd;
	private int								mLimit;
	private boolean							mUndoBeingApplied;
	private ArrayList<TKUndoManagerMonitor>	mMonitors;

	/** Creates a new undo manager that can hold up to 20 edits. */
	public TKUndoManager() {
		this(20);
	}

	/**
	 * Creates a new undo manager.
	 * 
	 * @param limit The maximum number of edits to hold.
	 */
	public TKUndoManager(int limit) {
		super();
		mIndexOfNextAdd = 0;
		mLimit = limit;
		mEdits.ensureCapacity(mLimit);
	}

	/** @param monitor The monitor to add. */
	public void addMonitor(TKUndoManagerMonitor monitor) {
		if (mMonitors == null) {
			mMonitors = new ArrayList<TKUndoManagerMonitor>(1);
		}
		mMonitors.add(monitor);
	}

	/** @param monitor The monitor to remove. */
	public void removeMonitor(TKUndoManagerMonitor monitor) {
		mMonitors.remove(monitor);
		if (mMonitors.isEmpty()) {
			mMonitors = null;
		}
	}

	private void notifyMonitors() {
		if (mMonitors != null) {
			for (TKUndoManagerMonitor monitor : new ArrayList<TKUndoManagerMonitor>(mMonitors)) {
				monitor.undoManagerStackChanged(this);
			}
		}
	}

	/** @return The current list of edits. */
	public ArrayList<TKUndo> getCurrentEdits() {
		return mEdits;
	}

	/** @return The index of the next addition. */
	public int getIndexOfNextAdd() {
		return mIndexOfNextAdd;
	}

	/** @return <code>true</code> when an undo or redo is being applied. */
	public boolean isUndoBeingApplied() {
		return mUndoBeingApplied;
	}

	@Override public synchronized boolean addEdit(TKUndo edit) {
		boolean retVal;

		trimEdits(mIndexOfNextAdd, mEdits.size() - 1);

		retVal = super.addEdit(edit);
		if (isInProgress()) {
			retVal = true;
		}

		mIndexOfNextAdd = mEdits.size();
		trimForLimit();

		notifyMonitors();
		return retVal;
	}

	@Override public synchronized boolean canApply(boolean forUndo) {
		if (isInProgress()) {
			TKUndo edit = nextEdit(forUndo);

			return edit != null && edit.canApply(forUndo);
		}
		return super.canApply(forUndo);
	}

	@Override public synchronized void apply(boolean forUndo) throws TKUndoException {
		mUndoBeingApplied = true;
		try {
			if (isInProgress()) {
				TKUndo edit = nextEdit(forUndo);
				boolean done = false;

				if (edit == null) {
					throw new TKUndoException(forUndo);
				}

				while (!done) {
					TKUndo next = mEdits.get(forUndo ? --mIndexOfNextAdd : mIndexOfNextAdd++);

					next.apply(forUndo);
					done = next == edit;
				}
			} else {
				super.apply(forUndo);
			}
		} finally {
			mUndoBeingApplied = false;
		}
		notifyMonitors();
	}

	/**
	 * Releases an edit.
	 * 
	 * @return <code>true</code> if an edit was released.
	 */
	public synchronized boolean releaseAnEdit() {
		int count = mEdits.size();

		if (count == 1) {
			discardAllEdits();
			return true;
		}

		if (count > 0) {
			int limit = getLimit();

			setLimit(count - 1);
			setLimit(limit);
			return true;
		}

		return false;
	}

	/** Resets the undo manager, marking all contained edits obsolete. */
	public synchronized void discardAllEdits() {
		int i = mEdits.size();

		while (i-- > 0) {
			mEdits.get(i).obsolete();
		}

		mEdits.clear();
		mIndexOfNextAdd = 0;
		notifyMonitors();
	}

	private TKUndo nextEdit(boolean forUndo) {
		TKUndo edit;
		int i;

		if (forUndo) {
			for (i = mIndexOfNextAdd - 1; i >= 0; i--) {
				edit = mEdits.get(i);
				if (edit.isSignificant()) {
					return edit;
				}
			}
		} else {
			for (i = mIndexOfNextAdd; i < mEdits.size(); i++) {
				edit = mEdits.get(i);
				if (edit.isSignificant()) {
					return edit;
				}
			}
		}
		return null;
	}

	/**
	 * Stops this undo manager from accepting any more edits, effectively turning it into a standard
	 * {@link TKCompoundUndo}. {@inheritDoc}
	 */
	@Override public synchronized void end() {
		super.end();
		trimEdits(mIndexOfNextAdd, mEdits.size() - 1);
	}

	@Override public synchronized String getName(boolean forUndo) {
		if (isInProgress()) {
			if (canApply(forUndo)) {
				return nextEdit(forUndo).getName(forUndo);
			}
			return forUndo ? Msgs.UNDO : Msgs.REDO;
		}
		return super.getName(forUndo);
	}

	/** @return The maximum number of edits this undo manager will hold. */
	public synchronized int getLimit() {
		return mLimit;
	}

	/** @param limit The maximum number of edits this undo manager will hold. */
	public synchronized void setLimit(int limit) {
		if (limit != mLimit) {
			if (!isInProgress()) {
				throw new RuntimeException("Attempt to call setLimit() after end() has been called"); //$NON-NLS-1$
			}
			mLimit = limit;
			trimForLimit();
			notifyMonitors();
		}
	}

	private void trimEdits(int from, int to) {
		if (from <= to) {
			for (int i = to; from <= i; i--) {
				mEdits.get(i).obsolete();
				mEdits.remove(i);
			}

			if (mIndexOfNextAdd > to) {
				mIndexOfNextAdd -= to - from + 1;
			} else if (mIndexOfNextAdd >= from) {
				mIndexOfNextAdd = from;
			}
		}
	}

	private void trimForLimit() {
		if (mLimit > 0) {
			int size = mEdits.size();

			if (size > mLimit) {
				int halfLimit = mLimit / 2;
				int keepFrom = mIndexOfNextAdd - 1 - halfLimit;
				int keepTo = mIndexOfNextAdd - 1 + halfLimit;

				if (keepTo - keepFrom + 1 > mLimit) {
					keepFrom++;
				}

				if (keepFrom < 0) {
					keepTo -= keepFrom;
					keepFrom = 0;
				}

				if (keepTo >= size) {
					int delta = size - keepTo - 1;

					keepTo += delta;
					keepFrom += delta;
				}

				trimEdits(keepTo + 1, size - 1);
				trimEdits(0, keepFrom - 1);
			}
		}
	}
}
