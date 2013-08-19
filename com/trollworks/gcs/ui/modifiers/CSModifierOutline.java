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

package com.trollworks.gcs.ui.modifiers;

import com.trollworks.gcs.model.CMDataFile;
import com.trollworks.gcs.model.modifier.CMModifier;
import com.trollworks.gcs.ui.common.CSOutline;
import com.trollworks.toolkit.widget.outline.TKOutlineModel;

/** Outline to host a {@link CMModifier} list. */
public class CSModifierOutline extends CSOutline {
	/**
	 * Creates a new {@link CSModifierOutline}.
	 * 
	 * @param dataFile The owning data file.
	 * @param model The outline model to use.
	 * @param rowSetChangedID The notification ID to use when the row set changes.
	 */
	public CSModifierOutline(CMDataFile dataFile, TKOutlineModel model, String rowSetChangedID) {
		super(dataFile, model, rowSetChangedID);
		CSModifierColumnID.addColumns(this);
		setAllowColumnDrag(true);
		setAllowColumnResize(true);
		setAllowRowDrag(true);
		setMenuTargetDelegate(this);
	}
}
