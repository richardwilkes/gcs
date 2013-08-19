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

package com.trollworks.gcs.ui.sheet;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.units.TKWeightUnits;

/** A weight input field. */
public class CSWeightField extends CSDoubleField {
	/**
	 * Creates a new weight input field.
	 * 
	 * @param character The character to listen to.
	 * @param consumedType The field to listen to.
	 * @param tooltip The tooltip to set.
	 */
	public CSWeightField(CMCharacter character, String consumedType, String tooltip) {
		this(character, consumedType, TKAlignment.LEFT, true, tooltip);
	}

	/**
	 * Creates a new weight input field.
	 * 
	 * @param character The character to listen to.
	 * @param consumedType The field to listen to.
	 * @param alignment The alignment of the field.
	 * @param editable Whether or not the user can edit this field.
	 * @param tooltip The tooltip to set.
	 */
	public CSWeightField(CMCharacter character, String consumedType, int alignment, boolean editable, String tooltip) {
		super(character, consumedType, alignment, false, 0.0, Double.MAX_VALUE / 10.0, editable, tooltip);
	}

	@Override public void handleNotification(Object producer, String type, Object data) {
		setText(TKWeightUnits.POUNDS.format(((Double) data).doubleValue()));
		invalidate();
	}
}
