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

import com.trollworks.toolkit.undo.TKSimpleUndo;
import com.trollworks.toolkit.undo.TKUndoException;

/** Provides undo support for {@link TKOutlineModel} changes. */
public class TKOutlineModelUndo extends TKSimpleUndo {
	private String						mName;
	private TKOutlineModel				mModel;
	private TKOutlineModelUndoSnapshot	mBefore;
	private TKOutlineModelUndoSnapshot	mAfter;

	/**
	 * Create a new model sort undo edit.
	 * 
	 * @param name The name for this undo edit.
	 * @param model The model to provide an undo edit for.
	 * @param before The original state.
	 * @param after The state after an action that changes the model.
	 */
	public TKOutlineModelUndo(String name, TKOutlineModel model, TKOutlineModelUndoSnapshot before, TKOutlineModelUndoSnapshot after) {
		super();
		mName = name;
		mModel = model;
		mBefore = before;
		mAfter = after;
	}

	@Override public String getName() {
		return mName;
	}

	@Override public void apply(boolean forUndo) throws TKUndoException {
		super.apply(forUndo);
		mModel.applyUndoSnapshot(forUndo ? mBefore : mAfter);
	}
}
