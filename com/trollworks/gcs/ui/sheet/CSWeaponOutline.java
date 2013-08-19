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

import com.trollworks.gcs.model.weapon.CMWeaponColumnID;
import com.trollworks.gcs.model.weapon.CMWeaponStats;
import com.trollworks.toolkit.widget.outline.TKOutline;
import com.trollworks.toolkit.widget.outline.TKRow;

import java.awt.Color;
import java.awt.dnd.DropTargetDragEvent;

/** An outline for weapons. */
public class CSWeaponOutline extends TKOutline {
	/**
	 * Creates a new weapon outline.
	 * 
	 * @param weaponClass The class of weapon to generate an outline for.
	 */
	public CSWeaponOutline(Class<? extends CMWeaponStats> weaponClass) {
		super(false);
		CMWeaponColumnID.addColumns(this, weaponClass, false);
		getModel().getColumnWithID(CMWeaponColumnID.DESCRIPTION.ordinal()).setSortCriteria(0, true);
		setEnabled(false);
	}

	@Override public Color getBackground(int rowIndex, boolean selected, boolean active) {
		return super.getBackground(rowIndex, active ? selected : false, active);
	}

	@Override protected boolean isRowDragAcceptable(DropTargetDragEvent dtde, TKRow[] rows) {
		return false;
	}
}
