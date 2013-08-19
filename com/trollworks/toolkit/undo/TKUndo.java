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

/** An object representing an edit that has been done, and that can be undone and redone. */
public interface TKUndo {
	/**
	 * Called to allow this edit to absorb the new edit if it can.
	 * 
	 * @param edit The edit to add.
	 * @return <code>true</code> if the edit was absorbed, <code>false</code> if it was not and
	 *         should be placed in the stack.
	 */
	public boolean addEdit(TKUndo edit);

	/**
	 * Called to ask if this edit should replace the specified edit that is already in the stack.
	 * 
	 * @param edit The edit that would be replaced.
	 * @return <code>true</code> to replace the specified edit.
	 */
	public boolean replaceEdit(TKUndo edit);

	/**
	 * @param forUndo Pass in <code>true</code> for the undoable version, or <code>false</code>
	 *            for the redoable version.
	 * @return <code>true</code> if it is still possible to do this operation.
	 */
	public boolean canApply(boolean forUndo);

	/**
	 * Undo/Redo the edit that was made.
	 * 
	 * @param forUndo Pass in <code>true</code> for the undoable version, or <code>false</code>
	 *            for the redoable version.
	 * @throws TKUndoException if the edit cannot be applied.
	 */
	public void apply(boolean forUndo) throws TKUndoException;

	/** Called when an edit should no longer be used. */
	public void obsolete();

	/** @return A localized, human-readable description of this edit. */
	public String getName();

	/**
	 * @param forUndo Pass in <code>true</code> for the undoable version, or <code>false</code>
	 *            for the redoable version.
	 * @return A localized, human-readable description of this edit. Typically, this is implemented
	 *         by calling {@link #getName()}and adding the appropriate undo/redo wording.
	 */
	public String getName(boolean forUndo);

	/**
	 * @return <code>false</code> if this edit is insignificant and should be coalesced with other
	 *         edits.
	 */
	public boolean isSignificant();
}
