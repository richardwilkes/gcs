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
import com.trollworks.gcs.ui.common.CSDropPanel;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;

/** The character damage panel. */
public class CSLiftPanel extends CSDropPanel {
	/**
	 * Creates a new damage panel.
	 * 
	 * @param character The character to display the data for.
	 */
	public CSLiftPanel(CMCharacter character) {
		super(new TKColumnLayout(2, 2, 0), Msgs.LIFT_MOVE);
		createLabelAndField(character, CMCharacter.ID_BASIC_LIFT, Msgs.BASIC_LIFT, Msgs.BASIC_LIFT_TOOLTIP);
		createLabelAndField(character, CMCharacter.ID_ONE_HANDED_LIFT, Msgs.ONE_HANDED_LIFT, Msgs.ONE_HANDED_LIFT_TOOLTIP);
		createLabelAndField(character, CMCharacter.ID_TWO_HANDED_LIFT, Msgs.TWO_HANDED_LIFT, Msgs.TWO_HANDED_LIFT_TOOLTIP);
		createLabelAndField(character, CMCharacter.ID_SHOVE_AND_KNOCK_OVER, Msgs.SHOVE_KNOCK_OVER, Msgs.SHOVE_KNOCK_OVER_TOOLTIP);
		createLabelAndField(character, CMCharacter.ID_RUNNING_SHOVE_AND_KNOCK_OVER, Msgs.RUNNING_SHOVE, Msgs.RUNNING_SHOVE_TOOLTIP);
		createLabelAndField(character, CMCharacter.ID_CARRY_ON_BACK, Msgs.CARRY_ON_BACK, Msgs.CARRY_ON_BACK_TOOLTIP);
		createLabelAndField(character, CMCharacter.ID_SHIFT_SLIGHTLY, Msgs.SHIFT_SLIGHTLY, Msgs.SHIFT_SLIGHTLY_TOOLTIP);
	}

	private void createLabelAndField(CMCharacter character, String key, String title, String tooltip) {
		CSWeightField field = new CSWeightField(character, key, TKAlignment.RIGHT, false, tooltip);

		add(new CSLabel(title, field));
		add(field);
	}
}
