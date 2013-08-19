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
 * 2005-2011 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.character;

import com.trollworks.ttk.layout.ColumnLayout;
import com.trollworks.ttk.utility.LocalizedMessages;

import javax.swing.SwingConstants;

/** The character damage panel. */
public class LiftPanel extends DropPanel {
	private static String	MSG_LIFT_MOVE;
	private static String	MSG_BASIC_LIFT;
	private static String	MSG_BASIC_LIFT_TOOLTIP;
	private static String	MSG_ONE_HANDED_LIFT;
	private static String	MSG_ONE_HANDED_LIFT_TOOLTIP;
	private static String	MSG_TWO_HANDED_LIFT;
	private static String	MSG_TWO_HANDED_LIFT_TOOLTIP;
	private static String	MSG_SHOVE_KNOCK_OVER;
	private static String	MSG_SHOVE_KNOCK_OVER_TOOLTIP;
	private static String	MSG_RUNNING_SHOVE;
	private static String	MSG_RUNNING_SHOVE_TOOLTIP;
	private static String	MSG_CARRY_ON_BACK;
	private static String	MSG_CARRY_ON_BACK_TOOLTIP;
	private static String	MSG_SHIFT_SLIGHTLY;
	private static String	MSG_SHIFT_SLIGHTLY_TOOLTIP;

	static {
		LocalizedMessages.initialize(LiftPanel.class);
	}

	/**
	 * Creates a new damage panel.
	 * 
	 * @param character The character to display the data for.
	 */
	public LiftPanel(GURPSCharacter character) {
		super(new ColumnLayout(2, 2, 0), MSG_LIFT_MOVE);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_BASIC_LIFT, MSG_BASIC_LIFT, MSG_BASIC_LIFT_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_ONE_HANDED_LIFT, MSG_ONE_HANDED_LIFT, MSG_ONE_HANDED_LIFT_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_TWO_HANDED_LIFT, MSG_TWO_HANDED_LIFT, MSG_TWO_HANDED_LIFT_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_SHOVE_AND_KNOCK_OVER, MSG_SHOVE_KNOCK_OVER, MSG_SHOVE_KNOCK_OVER_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_RUNNING_SHOVE_AND_KNOCK_OVER, MSG_RUNNING_SHOVE, MSG_RUNNING_SHOVE_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_CARRY_ON_BACK, MSG_CARRY_ON_BACK, MSG_CARRY_ON_BACK_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_SHIFT_SLIGHTLY, MSG_SHIFT_SLIGHTLY, MSG_SHIFT_SLIGHTLY_TOOLTIP, SwingConstants.RIGHT);
	}
}
