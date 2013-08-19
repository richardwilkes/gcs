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
 * 2005-2009 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.character;

import com.trollworks.ttk.layout.ColumnLayout;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.utility.UIUtilities;
import com.trollworks.ttk.widgets.Wrapper;

import java.awt.Color;
import java.awt.Dimension;

import javax.swing.SwingConstants;

/** The character description panel. */
public class DescriptionPanel extends DropPanel {
	private static String	MSG_DESCRIPTION;
	private static String	MSG_RACE;
	private static String	MSG_SIZE_MODIFIER;
	private static String	MSG_SIZE_MODIFIER_TOOLTIP;
	private static String	MSG_TECH_LEVEL;
	private static String	MSG_TECH_LEVEL_TOOLTIP;
	static String			MSG_AGE;
	static String			MSG_GENDER;
	static String			MSG_BIRTHDAY;
	static String			MSG_HEIGHT;
	static String			MSG_WEIGHT;
	static String			MSG_HAIR;
	static String			MSG_HAIR_TOOLTIP;
	static String			MSG_EYE_COLOR;
	static String			MSG_EYE_COLOR_TOOLTIP;
	static String			MSG_SKIN_COLOR;
	static String			MSG_SKIN_COLOR_TOOLTIP;
	static String			MSG_HANDEDNESS;
	static String			MSG_HANDEDNESS_TOOLTIP;

	static {
		LocalizedMessages.initialize(DescriptionPanel.class);
	}

	/**
	 * Creates a new description panel.
	 * 
	 * @param character The character to display the data for.
	 */
	public DescriptionPanel(GURPSCharacter character) {
		super(new ColumnLayout(5, 2, 0), MSG_DESCRIPTION);

		Wrapper wrapper = new Wrapper(new ColumnLayout(2, 2, 0));
		createLabelAndField(wrapper, character, Profile.ID_RACE, MSG_RACE, null, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_GENDER, MSG_GENDER, null, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_AGE, MSG_AGE, null, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_BIRTHDAY, MSG_BIRTHDAY, null, SwingConstants.LEFT);
		add(wrapper);

		createDivider();

		wrapper = new Wrapper(new ColumnLayout(2, 2, 0));
		createLabelAndField(wrapper, character, Profile.ID_HEIGHT, MSG_HEIGHT, null, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_WEIGHT, MSG_WEIGHT, null, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_SIZE_MODIFIER, MSG_SIZE_MODIFIER, MSG_SIZE_MODIFIER_TOOLTIP, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_TECH_LEVEL, MSG_TECH_LEVEL, MSG_TECH_LEVEL_TOOLTIP, SwingConstants.LEFT);
		add(wrapper);

		createDivider();

		wrapper = new Wrapper(new ColumnLayout(2, 2, 0));
		createLabelAndField(wrapper, character, Profile.ID_HAIR, MSG_HAIR, MSG_HAIR_TOOLTIP, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_EYE_COLOR, MSG_EYE_COLOR, MSG_EYE_COLOR_TOOLTIP, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_SKIN_COLOR, MSG_SKIN_COLOR, MSG_SKIN_COLOR_TOOLTIP, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_HANDEDNESS, MSG_HANDEDNESS, MSG_HANDEDNESS_TOOLTIP, SwingConstants.LEFT);
		add(wrapper);
	}

	private void createDivider() {
		Wrapper panel = new Wrapper();
		UIUtilities.setOnlySize(panel, new Dimension(1, 1));
		add(panel);
		addVerticalBackground(panel, Color.black);
	}
}
