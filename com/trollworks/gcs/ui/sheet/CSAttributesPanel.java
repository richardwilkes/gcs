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

/** The character attributes panel. */
public class CSAttributesPanel extends CSDropPanel {
	/**
	 * Creates a new attributes panel.
	 * 
	 * @param character The character to display the data for.
	 */
	public CSAttributesPanel(CMCharacter character) {
		super(new TKColumnLayout(2, 2, 0), Msgs.ATTRIBUTES, true);
		createLabelAndEditableField(character, CMCharacter.ID_STRENGTH, Msgs.ST, Msgs.ST_TOOLTIP);
		createLabelAndEditableField(character, CMCharacter.ID_DEXTERITY, Msgs.DX, Msgs.DX_TOOLTIP);
		createLabelAndEditableField(character, CMCharacter.ID_INTELLIGENCE, Msgs.IQ, Msgs.IQ_TOOLTIP);
		createLabelAndEditableField(character, CMCharacter.ID_HEALTH, Msgs.HT, Msgs.HT_TOOLTIP);
		createDivider(false);
		createDivider(true);
		createLabelAndEditableField(character, CMCharacter.ID_WILL, Msgs.WILL, Msgs.WILL_TOOLTIP);
		createDivider(false);
		createDivider(true);
		createLabelAndDoubleField(character, CMCharacter.ID_BASIC_SPEED, Msgs.BASIC_SPEED, Msgs.BASIC_SPEED_TOOLTIP);
		createLabelAndEditableField(character, CMCharacter.ID_BASIC_MOVE, Msgs.BASIC_MOVE, Msgs.BASIC_MOVE_TOOLTIP);
		createDivider(false);
		createDivider(true);
		createLabelAndEditableField(character, CMCharacter.ID_PERCEPTION, Msgs.PERCEPTION, Msgs.PERCEPTION_TOOLTIP);
		createLabelAndField(character, CMCharacter.ID_VISION, Msgs.VISION, null);
		createLabelAndField(character, CMCharacter.ID_HEARING, Msgs.HEARING, null);
		createLabelAndField(character, CMCharacter.ID_TASTE_AND_SMELL, Msgs.TASTE_SMELL, null);
		createLabelAndField(character, CMCharacter.ID_TOUCH, Msgs.TOUCH, null);
		createDivider(false);
		createDivider(true);
		createLabelAndDamageField(character, CMCharacter.ID_BASIC_THRUST, Msgs.THRUST, null);
		createLabelAndDamageField(character, CMCharacter.ID_BASIC_SWING, Msgs.SWING, null);
	}

	private void createDivider(boolean black) {
		TKPanel panel = new TKPanel();

		panel.setOnlySize(new Dimension(1, 1));
		add(panel);
		if (black) {
			addHorizontalBackground(panel, Color.black);
		}
		panel = new TKPanel();
		panel.setOnlySize(new Dimension(1, 1));
		add(panel);
	}

	private void createLabelAndEditableField(CMCharacter character, String key, String title, String tooltip) {
		CSIntegerField field = new CSIntegerField(character, key, false, 0, 9999, tooltip);

		add(new CSLabel(title, field));
		add(field);
	}

	private void createLabelAndDoubleField(CMCharacter character, String key, String title, String tooltip) {
		CSDoubleField field = new CSDoubleField(character, key, false, 0.0, 9999.0, tooltip);

		add(new CSLabel(title, field));
		add(field);
	}

	private void createLabelAndField(CMCharacter character, String key, String title, String tooltip) {
		CSIntegerField field = new CSIntegerField(character, key, false, 0, 9999, false, tooltip);

		add(new CSLabel(title, field));
		add(field);
	}

	private void createLabelAndDamageField(CMCharacter character, String key, String title, String tooltip) {
		CSField field = new CSField(character, key, TKAlignment.RIGHT, false, tooltip);

		add(new CSLabel(title, field));
		add(field);
	}
}
