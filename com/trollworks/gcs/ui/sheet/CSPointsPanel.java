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
import com.trollworks.toolkit.notification.TKNotifierTarget;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.layout.TKRowDistribution;

import java.awt.Color;
import java.awt.Dimension;
import java.text.MessageFormat;

/** The character points panel. */
public class CSPointsPanel extends CSDropPanel implements TKNotifierTarget {
	private CMCharacter	mCharacter;

	/**
	 * Creates a new points panel.
	 * 
	 * @param character The character to display the data for.
	 */
	public CSPointsPanel(CMCharacter character) {
		super(new TKColumnLayout(2, 2, 0, TKRowDistribution.DISTRIBUTE_HEIGHT), getTitle(character));
		mCharacter = character;
		createLabelAndDisplayField(character, CMCharacter.ID_ATTRIBUTE_POINTS, Msgs.ATTRIBUTE_POINTS, Msgs.ATTRIBUTE_POINTS_TOOLTIP);
		createLabelAndDisplayField(character, CMCharacter.ID_ADVANTAGE_POINTS, Msgs.ADVANTAGE_POINTS, Msgs.ADVANTAGE_POINTS_TOOLTIP);
		createLabelAndDisplayField(character, CMCharacter.ID_DISADVANTAGE_POINTS, Msgs.DISADVANTAGE_POINTS, Msgs.DISADVANTAGE_POINTS_TOOLTIP);
		createLabelAndDisplayField(character, CMCharacter.ID_QUIRK_POINTS, Msgs.QUIRK_POINTS, Msgs.QUIRK_POINTS_TOOLTIP);
		createLabelAndDisplayField(character, CMCharacter.ID_SKILL_POINTS, Msgs.SKILL_POINTS, Msgs.SKILL_POINTS_TOOLTIP);
		createLabelAndDisplayField(character, CMCharacter.ID_SPELL_POINTS, Msgs.SPELL_POINTS, Msgs.SPELL_POINTS_TOOLTIP);
		createLabelAndDisplayField(character, CMCharacter.ID_RACE_POINTS, Msgs.RACE_POINTS, Msgs.RACE_POINTS_TOOLTIP);
		createDivider();
		createLabelAndField(character, CMCharacter.ID_EARNED_POINTS, Msgs.EARNED_POINTS, Msgs.EARNED_POINTS_TOOLTIP);
		mCharacter.addTarget(this, CMCharacter.ID_TOTAL_POINTS);
	}

	@Override protected Dimension getMaximumSizeSelf() {
		Dimension size = super.getMaximumSizeSelf();

		size.width = getPreferredSize().width;
		return size;
	}

	private void createDivider() {
		TKPanel panel = new TKPanel();

		panel.setOnlySize(new Dimension(1, 1));
		add(panel);
		addHorizontalBackground(panel, Color.black);
		panel = new TKPanel();
		panel.setOnlySize(new Dimension(1, 1));
		add(panel);
	}

	private void createLabelAndDisplayField(CMCharacter character, String key, String title, String tooltip) {
		CSIntegerField field = new CSIntegerField(character, key, false, 0, 9999, false, tooltip);

		add(new CSLabel(title, field));
		add(field);
	}

	private void createLabelAndField(CMCharacter character, String key, String title, String tooltip) {
		CSIntegerField field = new CSIntegerField(character, key, false, -9999, 9999, tooltip);

		add(new CSLabel(title, field));
		add(field);
	}

	public void handleNotification(Object producer, String type, Object data) {
		getBoxedDropShadowBorder().setTitle(getTitle(mCharacter));
		invalidate();
	}

	private static String getTitle(CMCharacter character) {
		return MessageFormat.format(Msgs.POINTS, TKNumberUtils.format(character.getTotalPoints()));
	}
}
