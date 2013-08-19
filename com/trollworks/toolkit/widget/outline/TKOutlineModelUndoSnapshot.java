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

package com.trollworks.toolkit.widget.outline;

import com.trollworks.toolkit.widget.TKSelection;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;

/** The information an undo for the {@link TKOutlineModel} needs to operate. */
public class TKOutlineModelUndoSnapshot {
	private ArrayList<TKRow>					mRows;
	private HashMap<TKRow, TKRowUndoSnapshot>	mMap;
	private TKSelection							mSelection;
	private String								mSortConfig;

	/**
	 * Creates a snapshot of the information needed to undo any changes to the model.
	 * 
	 * @param model The model to create a snapshot for.
	 */
	public TKOutlineModelUndoSnapshot(TKOutlineModel model) {
		mRows = new ArrayList<TKRow>(model.getRows());
		mMap = new HashMap<TKRow, TKRowUndoSnapshot>();

		for (TKRow row : TKRowSorter.collectContainerRows(mRows, new HashSet<TKRow>())) {
			mMap.put(row, new TKRowUndoSnapshot(row));
		}

		mSelection = new TKSelection(model.getSelection());
		mSortConfig = model.getSortConfig();
	}

	/** @return The map. */
	public HashMap<TKRow, TKRowUndoSnapshot> getMap() {
		return mMap;
	}

	/** @return The rows. */
	public ArrayList<TKRow> getRows() {
		return mRows;
	}

	/** @return The selection. */
	public TKSelection getSelection() {
		return mSelection;
	}

	/** @return The sort config. */
	public String getSortConfig() {
		return mSortConfig;
	}
}
