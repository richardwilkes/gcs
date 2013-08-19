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

package com.trollworks.gcs.model;

import com.trollworks.toolkit.undo.TKSimpleUndo;
import com.trollworks.toolkit.undo.TKUndoException;

/** Provides undo support for template notes. */
public class CMTemplateNotesUndo extends TKSimpleUndo {
	private CMTemplate	mTemplate;
	private String		mBefore;
	private String		mAfter;

	/**
	 * Create a new template notes undo edit.
	 * 
	 * @param template The template to provide an undo edit for.
	 * @param before The original value.
	 * @param after The new value.
	 */
	public CMTemplateNotesUndo(CMTemplate template, String before, String after) {
		super();
		mTemplate = template;
		mBefore = before;
		mAfter = after;
	}

	@Override public String getName() {
		return Msgs.NOTES_UNDO;
	}

	@Override public void apply(boolean forUndo) throws TKUndoException {
		super.apply(forUndo);
		mTemplate.setNotes(forUndo ? mBefore : mAfter);
	}
}
