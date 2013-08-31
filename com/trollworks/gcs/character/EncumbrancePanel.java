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

import static com.trollworks.gcs.character.EncumbrancePanel_LS.*;

import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.layout.ColumnLayout;
import com.trollworks.ttk.notification.NotifierTarget;
import com.trollworks.ttk.text.Numbers;
import com.trollworks.ttk.utility.UIUtilities;
import com.trollworks.ttk.widgets.Wrapper;

import java.awt.Color;
import java.awt.Container;
import java.awt.Dimension;
import java.text.MessageFormat;

import javax.swing.SwingConstants;

@Localized({
				@LS(key = "ENCUMBRANCE_MOVE_DODGE", msg = "Encumbrance, Move & Dodge"),
				@LS(key = "ENCUMBRANCE_LEVEL", msg = "Level"),
				@LS(key = "MAX_CARRY", msg = "Max Load"),
				@LS(key = "MOVE", msg = "Move"),
				@LS(key = "DODGE", msg = "Dodge"),
				@LS(key = "ENCUMBRANCE_TOOLTIP", msg = "The encumbrance level"),
				@LS(key = "MAX_CARRY_TOOLTIP", msg = "The maximum load a character can carry and still remain within a specific encumbrance level"),
				@LS(key = "MOVE_TOOLTIP", msg = "The character's ground movement rate for a specific encumbrance level"),
				@LS(key = "DODGE_TOOLTIP", msg = "The character's dodge for a specific encumbrance level"),
				@LS(key = "NONE", msg = "None"),
				@LS(key = "LIGHT", msg = "Light"),
				@LS(key = "MEDIUM", msg = "Medium"),
				@LS(key = "HEAVY", msg = "Heavy"),
				@LS(key = "EXTRA_HEAVY", msg = "X-Heavy"),
				@LS(key = "ENCUMBRANCE_FORMAT", msg = "{0} ({1})"),
				@LS(key = "CURRENT_ENCUMBRANCE_FORMAT", msg = "\u2022 {0} ({1})"),
})
/** The character encumbrance panel. */
public class EncumbrancePanel extends DropPanel implements NotifierTarget {
	private static final Color		CURRENT_ENCUMBRANCE_COLOR	= new Color(252, 242, 196);
	/** The various encumbrance titles. */
	public static final String[]	ENCUMBRANCE_TITLES			= new String[] { NONE, LIGHT, MEDIUM, HEAVY, EXTRA_HEAVY };
	private GURPSCharacter			mCharacter;
	private PageLabel[]				mMarkers;

	/**
	 * Creates a new encumbrance panel.
	 * 
	 * @param character The character to display the data for.
	 */
	public EncumbrancePanel(GURPSCharacter character) {
		super(new ColumnLayout(7, 2, 0), ENCUMBRANCE_MOVE_DODGE, true);
		mCharacter = character;
		mMarkers = new PageLabel[GURPSCharacter.ENCUMBRANCE_LEVELS];
		PageHeader header = createHeader(this, ENCUMBRANCE_LEVEL, ENCUMBRANCE_TOOLTIP);
		addHorizontalBackground(header, Color.black);
		addVerticalBackground(createDivider(), Color.black);
		createHeader(this, MAX_CARRY, MAX_CARRY_TOOLTIP);
		addVerticalBackground(createDivider(), Color.black);
		createHeader(this, MOVE, MOVE_TOOLTIP);
		addVerticalBackground(createDivider(), Color.black);
		createHeader(this, DODGE, DODGE_TOOLTIP);
		int current = character.getEncumbranceLevel();
		for (int i = 0; i < GURPSCharacter.ENCUMBRANCE_LEVELS; i++) {
			mMarkers[i] = new PageLabel(getMarkerText(i, current), header);
			add(mMarkers[i]);
			if (current == i) {
				addHorizontalBackground(mMarkers[i], CURRENT_ENCUMBRANCE_COLOR);
			}
			createDivider();
			createDisabledField(this, character, GURPSCharacter.MAXIMUM_CARRY_PREFIX + i, MAX_CARRY_TOOLTIP, SwingConstants.RIGHT);
			createDivider();
			createDisabledField(this, character, GURPSCharacter.MOVE_PREFIX + i, MOVE_TOOLTIP, SwingConstants.RIGHT);
			createDivider();
			createDisabledField(this, character, GURPSCharacter.DODGE_PREFIX + i, DODGE_TOOLTIP, SwingConstants.RIGHT);
		}
		character.addTarget(this, GURPSCharacter.ID_CARRIED_WEIGHT, GURPSCharacter.ID_BASIC_LIFT);
	}

	private static String getMarkerText(int which, int current) {
		return MessageFormat.format(which == current ? CURRENT_ENCUMBRANCE_FORMAT : ENCUMBRANCE_FORMAT, ENCUMBRANCE_TITLES[which], Numbers.format(which));
	}

	private Container createDivider() {
		Wrapper panel = new Wrapper();
		UIUtilities.setOnlySize(panel, new Dimension(1, 1));
		add(panel);
		return panel;
	}

	@Override
	public void handleNotification(Object producer, String type, Object data) {
		int current = mCharacter.getEncumbranceLevel();
		for (int i = 0; i < GURPSCharacter.ENCUMBRANCE_LEVELS; i++) {
			if (i == current) {
				addHorizontalBackground(mMarkers[i], CURRENT_ENCUMBRANCE_COLOR);
			} else {
				removeHorizontalBackground(mMarkers[i]);
			}
			mMarkers[i].setText(getMarkerText(i, current));
		}
		revalidate();
		repaint();
	}

	@Override
	public int getNotificationPriority() {
		return 0;
	}
}
