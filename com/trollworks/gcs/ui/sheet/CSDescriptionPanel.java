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
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;

import java.awt.Color;
import java.awt.Dimension;

/** The character description panel. */
public class CSDescriptionPanel extends CSDropPanel {
	/**
	 * Creates a new description panel.
	 * 
	 * @param character The character to display the data for.
	 */
	public CSDescriptionPanel(CMCharacter character) {
		super(new TKColumnLayout(5, 2, 0), Msgs.DESCRIPTION);

		TKPanel wrapper = new TKPanel(new TKColumnLayout(2, 2, 0));
		createLabelAndField(wrapper, character, CMCharacter.ID_RACE, Msgs.RACE, null);
		createLabelAndField(wrapper, character, CMCharacter.ID_GENDER, Msgs.GENDER, null);
		createLabelAndIntegerField(wrapper, character, CMCharacter.ID_AGE, Msgs.AGE, null, false, 0, Integer.MAX_VALUE);
		createLabelAndField(wrapper, character, CMCharacter.ID_BIRTHDAY, Msgs.BIRTHDAY, null);
		add(wrapper);

		createDivider();

		wrapper = new TKPanel(new TKColumnLayout(2, 2, 0));
		createLabelAndHeightField(wrapper, character);
		createLabelAndWeightField(wrapper, character);
		createLabelAndIntegerField(wrapper, character, CMCharacter.ID_SIZE_MODIFIER, Msgs.SIZE_MODIFIER, Msgs.SIZE_MODIFIER_TOOLTIP, true, -8, 20);
		createLabelAndField(wrapper, character, CMCharacter.ID_TECH_LEVEL, Msgs.TECH_LEVEL, Msgs.TECH_LEVEL_TOOLTIP);
		add(wrapper);

		createDivider();

		wrapper = new TKPanel(new TKColumnLayout(2, 2, 0));
		createLabelAndField(wrapper, character, CMCharacter.ID_HAIR, Msgs.HAIR, Msgs.HAIR_TOOLTIP);
		createLabelAndField(wrapper, character, CMCharacter.ID_EYE_COLOR, Msgs.EYE_COLOR, Msgs.EYE_COLOR_TOOLTIP);
		createLabelAndField(wrapper, character, CMCharacter.ID_SKIN_COLOR, Msgs.SKIN_COLOR, Msgs.SKIN_COLOR_TOOLTIP);
		createLabelAndField(wrapper, character, CMCharacter.ID_HANDEDNESS, Msgs.HANDEDNESS, Msgs.HANDEDNESS_TOOLTIP);
		add(wrapper);
	}

	private void createDivider() {
		TKPanel panel = new TKPanel();

		panel.setOnlySize(new Dimension(1, 1));
		add(panel);
		addVerticalBackground(panel, Color.black);
	}

	private void createLabelAndField(TKPanel panel, CMCharacter character, String key, String title, String tooltip) {
		CSField field = new CSField(character, key, tooltip);

		panel.add(new CSLabel(title, field));
		panel.add(field);
	}

	private void createLabelAndIntegerField(TKPanel panel, CMCharacter character, String key, String title, String tooltip, boolean forceSign, int minValue, int maxValue) {
		CSIntegerField field = new CSIntegerField(character, key, TKAlignment.LEFT, forceSign, minValue, maxValue, tooltip);

		panel.add(new CSLabel(title, field));
		panel.add(field);
	}

	private void createLabelAndWeightField(TKPanel panel, CMCharacter character) {
		CSWeightField field = new CSWeightField(character, CMCharacter.ID_WEIGHT, null);

		panel.add(new CSLabel(Msgs.WEIGHT, field));
		panel.add(field);
	}

	private void createLabelAndHeightField(TKPanel panel, CMCharacter character) {
		CSHeightField field = new CSHeightField(character, CMCharacter.ID_HEIGHT);

		panel.add(new CSLabel(Msgs.HEIGHT, field));
		panel.add(field);
	}
}
