/*
 * Copyright (c) 1998-2014 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.layout.ColumnLayout;
import com.trollworks.toolkit.utility.Localization;

import javax.swing.SwingConstants;

/** The character damage panel. */
public class LiftPanel extends DropPanel {
	@Localize("Lifting & Moving Things")
	private static String	LIFT_MOVE;
	@Localize("Basic Lift:")
	private static String	BASIC_LIFT;
	@Localize("<html><body>The weight the character can lift overhead<br>with one hand in one second</body></html>")
	private static String	BASIC_LIFT_TOOLTIP;
	@Localize("One-Handed Lift:")
	private static String	ONE_HANDED_LIFT;
	@Localize("<html><body>The weight the character can lift overhead<br>with one hand in two seconds</body></html>")
	private static String	ONE_HANDED_LIFT_TOOLTIP;
	@Localize("Two-Handed Lift:")
	private static String	TWO_HANDED_LIFT;
	@Localize("<html><body>The weight the character can lift overhead<br>with both hands in four seconds</body></html>")
	private static String	TWO_HANDED_LIFT_TOOLTIP;
	@Localize("Shove & Knock Over:")
	private static String	SHOVE_KNOCK_OVER;
	@Localize("<html><body>The weight of an object the character<br>can shove and knock over</body></html>")
	private static String	SHOVE_KNOCK_OVER_TOOLTIP;
	@Localize("Running Shove & Knock Over:")
	private static String	RUNNING_SHOVE;
	@Localize("<html><body>The weight of an object the character can shove<br> and knock over with a running start</body></html>")
	private static String	RUNNING_SHOVE_TOOLTIP;
	@Localize("Carry On Back:")
	private static String	CARRY_ON_BACK;
	@Localize("The weight the character can carry slung across the back")
	private static String	CARRY_ON_BACK_TOOLTIP;
	@Localize("Shift Slightly:")
	private static String	SHIFT_SLIGHTLY;
	@Localize("<html><body>The weight of an object the character<br>can shift slightly on a floor</body></html>")
	private static String	SHIFT_SLIGHTLY_TOOLTIP;

	static {
		Localization.initialize();
	}

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
