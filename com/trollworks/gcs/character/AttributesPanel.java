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
import com.trollworks.gcs.widgets.UIUtilities;
import com.trollworks.gcs.widgets.layout.ColumnLayout;

import java.awt.Color;
import java.awt.Dimension;

import javax.swing.JPanel;
import javax.swing.SwingConstants;

/** The character attributes panel. */
public class AttributesPanel extends DropPanel {
	private static String	MSG_ATTRIBUTES;
	private static String	MSG_ST;
	private static String	MSG_ST_TOOLTIP;
	private static String	MSG_DX;
	private static String	MSG_DX_TOOLTIP;
	private static String	MSG_IQ;
	private static String	MSG_IQ_TOOLTIP;
	private static String	MSG_HT;
	private static String	MSG_HT_TOOLTIP;
	private static String	MSG_WILL;
	private static String	MSG_WILL_TOOLTIP;
	private static String	MSG_PERCEPTION;
	private static String	MSG_PERCEPTION_TOOLTIP;
	private static String	MSG_VISION;
	private static String	MSG_HEARING;
	private static String	MSG_TOUCH;
	private static String	MSG_TASTE_SMELL;
	private static String	MSG_BASIC_SPEED;
	private static String	MSG_BASIC_SPEED_TOOLTIP;
	private static String	MSG_BASIC_MOVE;
	private static String	MSG_BASIC_MOVE_TOOLTIP;
	private static String	MSG_THRUST;
	private static String	MSG_SWING;

	static {
		LocalizedMessages.initialize(AttributesPanel.class);
	}

	/**
	 * Creates a new attributes panel.
	 * 
	 * @param character The character to display the data for.
	 */
	public AttributesPanel(GURPSCharacter character) {
		super(new ColumnLayout(2, 2, 0), MSG_ATTRIBUTES, true);
		createLabelAndField(this, character, GURPSCharacter.ID_STRENGTH, MSG_ST, MSG_ST_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndField(this, character, GURPSCharacter.ID_DEXTERITY, MSG_DX, MSG_DX_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndField(this, character, GURPSCharacter.ID_INTELLIGENCE, MSG_IQ, MSG_IQ_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndField(this, character, GURPSCharacter.ID_HEALTH, MSG_HT, MSG_HT_TOOLTIP, SwingConstants.RIGHT);
		createDivider(false);
		createDivider(true);
		createLabelAndField(this, character, GURPSCharacter.ID_WILL, MSG_WILL, MSG_WILL_TOOLTIP, SwingConstants.RIGHT);
		createDivider(false);
		createDivider(true);
		createLabelAndField(this, character, GURPSCharacter.ID_BASIC_SPEED, MSG_BASIC_SPEED, MSG_BASIC_SPEED_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndField(this, character, GURPSCharacter.ID_BASIC_MOVE, MSG_BASIC_MOVE, MSG_BASIC_MOVE_TOOLTIP, SwingConstants.RIGHT);
		createDivider(false);
		createDivider(true);
		createLabelAndField(this, character, GURPSCharacter.ID_PERCEPTION, MSG_PERCEPTION, MSG_PERCEPTION_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_VISION, MSG_VISION, null, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_HEARING, MSG_HEARING, null, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_TASTE_AND_SMELL, MSG_TASTE_SMELL, null, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_TOUCH, MSG_TOUCH, null, SwingConstants.RIGHT);
		createDivider(false);
		createDivider(true);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_BASIC_THRUST, MSG_THRUST, null, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_BASIC_SWING, MSG_SWING, null, SwingConstants.RIGHT);
	}

	private void createDivider(boolean black) {
		JPanel panel = new JPanel();
		panel.setOpaque(false);

		UIUtilities.setOnlySize(panel, new Dimension(1, 1));
		add(panel);
		if (black) {
			addHorizontalBackground(panel, Color.black);
		}
		panel = new JPanel();
		panel.setOpaque(false);
		UIUtilities.setOnlySize(panel, new Dimension(1, 1));
		add(panel);
	}
}
