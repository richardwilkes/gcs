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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.character;

import com.trollworks.ttk.layout.Alignment;
import com.trollworks.ttk.layout.FlexComponent;
import com.trollworks.ttk.layout.FlexGrid;
import com.trollworks.ttk.layout.FlexRow;
import com.trollworks.ttk.layout.FlexSpacer;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.utility.UIUtilities;
import com.trollworks.ttk.widgets.Wrapper;

import java.awt.Color;
import java.awt.Dimension;

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
	private static String	MSG_FRIGHT_CHECK;
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
	private static String	MSG_BASIC_THRUST;
	private static String	MSG_BASIC_THRUST_TOOLTIP;
	private static String	MSG_BASIC_SWING;
	private static String	MSG_BASIC_SWING_TOOLTIP;

	static {
		LocalizedMessages.initialize(AttributesPanel.class);
	}

	/**
	 * Creates a new attributes panel.
	 * 
	 * @param character The character to display the data for.
	 */
	public AttributesPanel(GURPSCharacter character) {
		super(null, MSG_ATTRIBUTES, true);
		FlexGrid grid = new FlexGrid();
		grid.setVerticalGap(0);
		int row = 0;
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_STRENGTH, MSG_ST, MSG_ST_TOOLTIP, SwingConstants.RIGHT, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_DEXTERITY, MSG_DX, MSG_DX_TOOLTIP, SwingConstants.RIGHT, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_INTELLIGENCE, MSG_IQ, MSG_IQ_TOOLTIP, SwingConstants.RIGHT, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_HEALTH, MSG_HT, MSG_HT_TOOLTIP, SwingConstants.RIGHT, true);
		createDivider(grid, row++, false);
		createDivider(grid, row++, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_WILL, MSG_WILL, MSG_WILL_TOOLTIP, SwingConstants.RIGHT, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_FRIGHT_CHECK, MSG_FRIGHT_CHECK, null, SwingConstants.RIGHT, false);
		createDivider(grid, row++, false);
		createDivider(grid, row++, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_BASIC_SPEED, MSG_BASIC_SPEED, MSG_BASIC_SPEED_TOOLTIP, SwingConstants.RIGHT, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_BASIC_MOVE, MSG_BASIC_MOVE, MSG_BASIC_MOVE_TOOLTIP, SwingConstants.RIGHT, true);
		createDivider(grid, row++, false);
		createDivider(grid, row++, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_PERCEPTION, MSG_PERCEPTION, MSG_PERCEPTION_TOOLTIP, SwingConstants.RIGHT, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_VISION, MSG_VISION, null, SwingConstants.RIGHT, false);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_HEARING, MSG_HEARING, null, SwingConstants.RIGHT, false);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_TASTE_AND_SMELL, MSG_TASTE_SMELL, null, SwingConstants.RIGHT, false);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_TOUCH, MSG_TOUCH, null, SwingConstants.RIGHT, false);
		createDivider(grid, row++, false);
		createDivider(grid, row++, true);
		createDamageFields(grid, row++, character);
		grid.apply(this);
	}

	private void createLabelAndField(FlexGrid grid, int row, GURPSCharacter character, String key, String title, String tooltip, int alignment, boolean enabled) {
		PageField field = new PageField(character, key, alignment, enabled, tooltip);
		PageLabel label = new PageLabel(title, field);
		add(label);
		add(field);
		grid.add(new FlexComponent(label, Alignment.RIGHT_BOTTOM, Alignment.CENTER), row, 0);
		grid.add(field, row, 1);
	}

	private void createDamageFields(FlexGrid grid, int rowIndex, GURPSCharacter character) {
		FlexRow row = new FlexRow();
		row.setHorizontalAlignment(Alignment.CENTER);
		createDamageLabelAndField(row, character, GURPSCharacter.ID_BASIC_THRUST, MSG_BASIC_THRUST, MSG_BASIC_THRUST_TOOLTIP);
		row.add(new FlexSpacer(0, 0, false, false));
		createDamageLabelAndField(row, character, GURPSCharacter.ID_BASIC_SWING, MSG_BASIC_SWING, MSG_BASIC_SWING_TOOLTIP);
		grid.add(row, rowIndex, 0, 1, 2);
	}

	private void createDamageLabelAndField(FlexRow row, GURPSCharacter character, String key, String title, String tooltip) {
		PageField field = new PageField(character, key, SwingConstants.RIGHT, false, tooltip);
		PageLabel label = new PageLabel(title, field);
		add(label);
		add(field);
		row.add(new FlexComponent(label, true));
		row.add(new FlexComponent(field, true));
	}

	private void createDivider(FlexGrid grid, int row, boolean black) {
		Wrapper panel = new Wrapper();
		UIUtilities.setOnlySize(panel, new Dimension(1, 1));
		add(panel);
		if (black) {
			addHorizontalBackground(panel, Color.black);
		}
		grid.add(panel, row, 0, 1, 2);
	}
}
