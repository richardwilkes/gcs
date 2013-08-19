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

package com.trollworks.gcs.character;

import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.notification.NotifierTarget;
import com.trollworks.gcs.utility.text.NumberUtils;
import com.trollworks.gcs.widgets.UIUtilities;
import com.trollworks.gcs.widgets.layout.ColumnLayout;

import java.awt.Color;
import java.awt.Container;
import java.awt.Dimension;
import java.text.MessageFormat;

import javax.swing.JPanel;
import javax.swing.SwingConstants;

/** The character encumbrance panel. */
public class EncumbrancePanel extends DropPanel implements NotifierTarget {
	private static String			MSG_ENCUMBRANCE_MOVE_DODGE;
	private static String			MSG_ENCUMBRANCE_LEVEL;
	private static String			MSG_MAX_CARRY;
	private static String			MSG_MOVE;
	private static String			MSG_DODGE;
	private static String			MSG_ENCUMBRANCE_TOOLTIP;
	private static String			MSG_MAX_CARRY_TOOLTIP;
	private static String			MSG_MOVE_TOOLTIP;
	private static String			MSG_DODGE_TOOLTIP;
	private static String			MSG_NONE;
	private static String			MSG_LIGHT;
	private static String			MSG_MEDIUM;
	private static String			MSG_HEAVY;
	private static String			MSG_EXTRA_HEAVY;
	static String					MSG_ENCUMBRANCE_FORMAT;
	static String					MSG_CURRENT_ENCUMBRANCE_FORMAT;

	static {
		LocalizedMessages.initialize(EncumbrancePanel.class);
	}

	private static final Color		CURRENT_ENCUMBRANCE_COLOR	= new Color(252, 242, 196);
	/** The various encumbrance titles. */
	public static final String[]	ENCUMBRANCE_TITLES			= new String[] { MSG_NONE, MSG_LIGHT, MSG_MEDIUM, MSG_HEAVY, MSG_EXTRA_HEAVY };
	private GURPSCharacter			mCharacter;
	private PageLabel[]				mMarkers;

	/**
	 * Creates a new encumbrance panel.
	 * 
	 * @param character The character to display the data for.
	 */
	public EncumbrancePanel(GURPSCharacter character) {
		super(new ColumnLayout(7, 2, 0), MSG_ENCUMBRANCE_MOVE_DODGE, true);
		mCharacter = character;
		mMarkers = new PageLabel[GURPSCharacter.ENCUMBRANCE_LEVELS];
		PageHeader header = createHeader(this, MSG_ENCUMBRANCE_LEVEL, MSG_ENCUMBRANCE_TOOLTIP);
		addHorizontalBackground(header, Color.black);
		addVerticalBackground(createDivider(), Color.black);
		createHeader(this, MSG_MAX_CARRY, MSG_MAX_CARRY_TOOLTIP);
		addVerticalBackground(createDivider(), Color.black);
		createHeader(this, MSG_MOVE, MSG_MOVE_TOOLTIP);
		addVerticalBackground(createDivider(), Color.black);
		createHeader(this, MSG_DODGE, MSG_DODGE_TOOLTIP);
		int current = character.getEncumbranceLevel();
		for (int i = 0; i < GURPSCharacter.ENCUMBRANCE_LEVELS; i++) {
			mMarkers[i] = new PageLabel(MessageFormat.format(i == current ? MSG_CURRENT_ENCUMBRANCE_FORMAT : MSG_ENCUMBRANCE_FORMAT, ENCUMBRANCE_TITLES[i], NumberUtils.format(i)), header);
			add(mMarkers[i]);
			if (current == i) {
				addHorizontalBackground(mMarkers[i], CURRENT_ENCUMBRANCE_COLOR);
			}
			createDivider();
			createDisabledField(this, character, GURPSCharacter.MAXIMUM_CARRY_PREFIX + i, MSG_MAX_CARRY_TOOLTIP, SwingConstants.RIGHT);
			createDivider();
			createDisabledField(this, character, GURPSCharacter.MOVE_PREFIX + i, MSG_MOVE_TOOLTIP, SwingConstants.RIGHT);
			createDivider();
			createDisabledField(this, character, GURPSCharacter.DODGE_PREFIX + i, MSG_DODGE_TOOLTIP, SwingConstants.RIGHT);
		}
		character.addTarget(this, GURPSCharacter.ID_CARRIED_WEIGHT, GURPSCharacter.ID_BASIC_LIFT);
	}

	private Container createDivider() {
		JPanel panel = new JPanel();
		panel.setOpaque(false);
		UIUtilities.setOnlySize(panel, new Dimension(1, 1));
		add(panel);
		return panel;
	}

	public void handleNotification(Object producer, String type, Object data) {
		int current = mCharacter.getEncumbranceLevel();
		for (int i = 0; i < GURPSCharacter.ENCUMBRANCE_LEVELS; i++) {
			if (i == current) {
				addHorizontalBackground(mMarkers[i], CURRENT_ENCUMBRANCE_COLOR);
			} else {
				removeHorizontalBackground(mMarkers[i]);
			}
		}
		repaint();
	}
}
