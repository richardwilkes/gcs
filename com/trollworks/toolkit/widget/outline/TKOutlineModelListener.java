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

/** Objects interested in {@link TKOutlineModel} notifications must implement this interface. */
public interface TKOutlineModelListener {
	/**
	 * Called after rows are added.
	 * 
	 * @param model The model that was altered.
	 * @param rows The rows being added.
	 */
	public void rowsAdded(TKOutlineModel model, TKRow[] rows);

	/**
	 * Called prior to rows being removed.
	 * 
	 * @param model The affected model.
	 * @param rows The rows being removed.
	 */
	public void rowsWillBeRemoved(TKOutlineModel model, TKRow[] rows);

	/**
	 * Called after rows are removed.
	 * 
	 * @param model The affected model.
	 * @param rows The rows that were removed.
	 */
	public void rowsWereRemoved(TKOutlineModel model, TKRow[] rows);

	/**
	 * Called whenever the sort settings are cleared.
	 * 
	 * @param model The model whose sort settings were cleared.
	 */
	public void sortCleared(TKOutlineModel model);

	/**
	 * Called whenever the model is sorted.
	 * 
	 * @param model The model that was sorted.
	 * @param restoring <code>true</code> when the sort is being restored (usually due to row
	 *            disclosure).
	 */
	public void sorted(TKOutlineModel model, boolean restoring);

	/**
	 * Called whenever the "locked" state is about to change.
	 * 
	 * @param model The model whose state will change.
	 */
	public void lockedStateWillChange(TKOutlineModel model);

	/**
	 * Called whenever the "locked" state changes.
	 * 
	 * @param model The model whose state changed.
	 */
	public void lockedStateDidChange(TKOutlineModel model);

	/**
	 * Called whenever the selection is about to change.
	 * 
	 * @param model The model whose selection will change.
	 */
	public void selectionWillChange(TKOutlineModel model);

	/**
	 * Called whenever the selection changes.
	 * 
	 * @param model The model whose selection changed.
	 */
	public void selectionDidChange(TKOutlineModel model);

	/**
	 * Called whenever an undo/redo is about to be applied.
	 * 
	 * @param model The model which will be affected.
	 */
	public void undoWillHappen(TKOutlineModel model);

	/**
	 * Called whenever an undo/redo has been applied.
	 * 
	 * @param model The model which was affected.
	 */
	public void undoDidHappen(TKOutlineModel model);
}
