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
import com.trollworks.toolkit.widget.layout.TKRowDistribution;

import java.awt.Color;
import java.awt.Dimension;

/** The character hit points panel. */
public class CSHitPointsPanel extends CSDropPanel {
	/**
	 * Creates a new hit points panel.
	 * 
	 * @param character The character to display the data for.
	 */
	public CSHitPointsPanel(CMCharacter character) {
		super(new TKColumnLayout(2, 2, 0, TKRowDistribution.DISTRIBUTE_HEIGHT), Msgs.FP_HP);
		createLabelAndEditableField(character, CMCharacter.ID_CURRENT_FATIGUE_POINTS, Msgs.FP_CURRENT, Msgs.FP_CURRENT_TOOLTIP);
		createLabelAndEditableNumberField(character, CMCharacter.ID_FATIGUE_POINTS, Msgs.FP, Msgs.FP_TOOLTIP);
		createDivider();
		createLabelAndField(character, CMCharacter.ID_TIRED_FATIGUE_POINTS, Msgs.FP_TIRED, Msgs.FP_TIRED_TOOLTIP);
		createLabelAndField(character, CMCharacter.ID_UNCONSCIOUS_CHECKS_FATIGUE_POINTS, Msgs.FP_UNCONSCIOUS_CHECKS, Msgs.FP_UNCONSCIOUS_CHECKS_TOOLTIP);
		createLabelAndField(character, CMCharacter.ID_UNCONSCIOUS_FATIGUE_POINTS, Msgs.FP_UNCONSCIOUS, Msgs.FP_UNCONSCIOUS_TOOLTIP);
		createDivider();
		createLabelAndEditableField(character, CMCharacter.ID_CURRENT_HIT_POINTS, Msgs.HP_CURRENT, Msgs.HP_CURRENT_TOOLTIP);
		createLabelAndEditableNumberField(character, CMCharacter.ID_HIT_POINTS, Msgs.HP, Msgs.HP_TOOLTIP);
		createDivider();
		createLabelAndField(character, CMCharacter.ID_REELING_HIT_POINTS, Msgs.HP_REELING, Msgs.HP_REELING_TOOLTIP);
		createLabelAndField(character, CMCharacter.ID_UNCONSCIOUS_CHECKS_HIT_POINTS, Msgs.HP_UNCONSCIOUS_CHECKS, Msgs.HP_UNCONSCIOUS_CHECKS_TOOLTIP);
		createLabelAndField(character, CMCharacter.ID_DEATH_CHECK_1_HIT_POINTS, Msgs.HP_DEATH_CHECK_1, Msgs.HP_DEATH_CHECK_1_TOOLTIP);
		createLabelAndField(character, CMCharacter.ID_DEATH_CHECK_2_HIT_POINTS, Msgs.HP_DEATH_CHECK_2, Msgs.HP_DEATH_CHECK_2_TOOLTIP);
		createLabelAndField(character, CMCharacter.ID_DEATH_CHECK_3_HIT_POINTS, Msgs.HP_DEATH_CHECK_3, Msgs.HP_DEATH_CHECK_3_TOOLTIP);
		createLabelAndField(character, CMCharacter.ID_DEATH_CHECK_4_HIT_POINTS, Msgs.HP_DEATH_CHECK_4, Msgs.HP_DEATH_CHECK_4_TOOLTIP);
		createLabelAndField(character, CMCharacter.ID_DEAD_HIT_POINTS, Msgs.HP_DEAD, Msgs.HP_DEAD_TOOLTIP);
	}

	@Override protected Dimension getMaximumSizeSelf() {
		Dimension size = super.getMaximumSizeSelf();

		size.width = getPreferredSize().width;
		return size;
	}

	private void createDivider() {
		createOneByOnePanel();
		createOneByOnePanel();
		addHorizontalBackground(createOneByOnePanel(), Color.black);
		createOneByOnePanel();
	}

	private TKPanel createOneByOnePanel() {
		TKPanel panel = new TKPanel();

		panel.setOnlySize(new Dimension(1, 1));
		add(panel);
		return panel;
	}

	private void createLabelAndField(CMCharacter character, String key, String title, String tooltip) {
		CSIntegerField field = new CSIntegerField(character, key, false, 0, 9999, false, tooltip);

		add(new CSLabel(title, field));
		add(field);
	}

	private void createLabelAndEditableField(CMCharacter character, String key, String title, String tooltip) {
		CSField field = new CSField(character, key, TKAlignment.RIGHT, tooltip);

		add(new CSLabel(title, field));
		add(field);
	}

	private void createLabelAndEditableNumberField(CMCharacter character, String key, String title, String tooltip) {
		CSIntegerField field = new CSIntegerField(character, key, false, 0, 999999, tooltip);

		add(new CSLabel(title, field));
		add(field);
	}
}
