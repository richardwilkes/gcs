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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.widgets.outline;

import java.util.ArrayList;

/** The information an undo for the row needs to operate. */
public class RowUndoSnapshot {
	private Row				mParent;
	private boolean			mOpen;
	private ArrayList<Row>	mChildren;

	/**
	 * Creates a snapshot of the information needed to undo any changes to the row.
	 * 
	 * @param row The row to create a snapshot for.
	 */
	public RowUndoSnapshot(Row row) {
		mParent = row.getParent();
		mOpen = row.isOpen();
		mChildren = row.canHaveChildren() ? new ArrayList<Row>(row.getChildren()) : null;
	}

	/** @return The children. */
	public ArrayList<Row> getChildren() {
		return mChildren;
	}

	/** @return Whether the row should be open. */
	public boolean isOpen() {
		return mOpen;
	}

	/** @return The parent. */
	public Row getParent() {
		return mParent;
	}
}
