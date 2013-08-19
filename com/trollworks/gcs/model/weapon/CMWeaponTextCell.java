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

import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.widget.TKKeyEventFilter;
import com.trollworks.toolkit.widget.outline.TKOutline;
import com.trollworks.toolkit.widget.outline.TKColumn;
import com.trollworks.toolkit.widget.outline.TKRow;
import com.trollworks.toolkit.widget.outline.TKTextCell;

import java.awt.Color;

/** Represents cells in a {@link TKOutline}. */
public class CMWeaponTextCell extends TKTextCell {
	/**
	 * Create a new text cell.
	 * 
	 * @param alignment The horizontal text alignment to use.
	 * @param compareType The type of comparison to perform when asked.
	 * @param keyFilter The key event filter to use.
	 * @param wrapped Pass in <code>true</code> to enable wrapping.
	 */
	public CMWeaponTextCell(int alignment, int compareType, TKKeyEventFilter keyFilter, boolean wrapped) {
		super(null, null, null, alignment, TKAlignment.TOP, compareType, false, TKAlignment.CENTER, keyFilter, wrapped);
	}

	@Override public Color getColor(boolean selected, boolean active, TKRow row, TKColumn column) {
		return super.getColor(active ? selected : false, active, row, column);
	}
}
