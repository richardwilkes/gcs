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

package com.trollworks.gcs.model.weapon;

import com.trollworks.toolkit.widget.outline.TKColumn;
import com.trollworks.toolkit.widget.outline.TKRow;

/** A non-editable row for displaying weapon information. */
public class CMWeaponDisplayRow extends TKRow {
	private CMWeaponStats	mWeapon;

	/**
	 * Creates a new weapon display row.
	 * 
	 * @param weapon The weapon to display.
	 */
	public CMWeaponDisplayRow(CMWeaponStats weapon) {
		super();
		mWeapon = weapon;
	}

	/** @return The weapon. */
	public CMWeaponStats getWeapon() {
		return mWeapon;
	}

	@Override public Object getData(TKColumn column) {
		return CMWeaponColumnID.values()[column.getID()].getData(mWeapon);
	}

	@Override public String getDataAsText(TKColumn column) {
		return CMWeaponColumnID.values()[column.getID()].getDataAsText(mWeapon);
	}

	@Override public void setData(TKColumn column, Object data) {
		assert false : "setData() is not supported"; //$NON-NLS-1$
	}
}
