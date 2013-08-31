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
 * Portions created by the Initial Developer are Copyright (C) 1998-2013 the
 * Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.character;

import static com.trollworks.gcs.character.LiftPanel_LS.*;

import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.layout.ColumnLayout;

import javax.swing.SwingConstants;

@Localized({
				@LS(key = "LIFT_MOVE", msg = "Lifting & Moving Things"),
				@LS(key = "BASIC_LIFT", msg = "Basic Lift:"),
				@LS(key = "BASIC_LIFT_TOOLTIP", msg = "<html><body>The weight the character can lift overhead<br>with one hand in one second</body></html>"),
				@LS(key = "ONE_HANDED_LIFT", msg = "One-Handed Lift:"),
				@LS(key = "ONE_HANDED_LIFT_TOOLTIP", msg = "<html><body>The weight the character can lift overhead<br>with one hand in two seconds</body></html>"),
				@LS(key = "TWO_HANDED_LIFT", msg = "Two-Handed Lift:"),
				@LS(key = "TWO_HANDED_LIFT_TOOLTIP", msg = "<html><body>The weight the character can lift overhead<br>with both hands in four seconds</body></html>"),
				@LS(key = "SHOVE_KNOCK_OVER", msg = "Shove & Knock Over:"),
				@LS(key = "SHOVE_KNOCK_OVER_TOOLTIP", msg = "<html><body>The weight of an object the character<br>can shove and knock over</body></html>"),
				@LS(key = "RUNNING_SHOVE", msg = "Running Shove & Knock Over:"),
				@LS(key = "RUNNING_SHOVE_TOOLTIP", msg = "<html><body>The weight of an object the character can shove<br> and knock over with a running start</body></html>"),
				@LS(key = "CARRY_ON_BACK", msg = "Carry On Back:"),
				@LS(key = "CARRY_ON_BACK_TOOLTIP", msg = "The weight the character can carry slung across the back"),
				@LS(key = "SHIFT_SLIGHTLY", msg = "Shift Slightly:"),
				@LS(key = "SHIFT_SLIGHTLY_TOOLTIP", msg = "<html><body>The weight of an object the character<br>can shift slightly on a floor</body></html>"),
})
/** The character damage panel. */
public class LiftPanel extends DropPanel {
	/**
	 * Creates a new damage panel.
	 * 
	 * @param character The character to display the data for.
	 */
	public LiftPanel(GURPSCharacter character) {
		super(new ColumnLayout(2, 2, 0), LIFT_MOVE);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_BASIC_LIFT, BASIC_LIFT, BASIC_LIFT_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_ONE_HANDED_LIFT, ONE_HANDED_LIFT, ONE_HANDED_LIFT_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_TWO_HANDED_LIFT, TWO_HANDED_LIFT, TWO_HANDED_LIFT_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_SHOVE_AND_KNOCK_OVER, SHOVE_KNOCK_OVER, SHOVE_KNOCK_OVER_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_RUNNING_SHOVE_AND_KNOCK_OVER, RUNNING_SHOVE, RUNNING_SHOVE_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_CARRY_ON_BACK, CARRY_ON_BACK, CARRY_ON_BACK_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_SHIFT_SLIGHTLY, SHIFT_SLIGHTLY, SHIFT_SLIGHTLY_TOOLTIP, SwingConstants.RIGHT);
	}
}
