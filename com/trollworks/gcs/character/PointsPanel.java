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

import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.notification.NotifierTarget;
import com.trollworks.gcs.utility.text.NumberUtils;
import com.trollworks.gcs.widgets.UIUtilities;
import com.trollworks.gcs.widgets.Wrapper;
import com.trollworks.gcs.widgets.layout.ColumnLayout;
import com.trollworks.gcs.widgets.layout.RowDistribution;

import java.awt.Color;
import java.awt.Dimension;
import java.text.MessageFormat;

import javax.swing.SwingConstants;

/** The character points panel. */
public class PointsPanel extends DropPanel implements NotifierTarget {
	private static String	MSG_POINTS;
	private static String	MSG_ATTRIBUTE_POINTS;
	private static String	MSG_ATTRIBUTE_POINTS_TOOLTIP;
	private static String	MSG_ADVANTAGE_POINTS;
	private static String	MSG_ADVANTAGE_POINTS_TOOLTIP;
	private static String	MSG_DISADVANTAGE_POINTS;
	private static String	MSG_DISADVANTAGE_POINTS_TOOLTIP;
	private static String	MSG_QUIRK_POINTS;
	private static String	MSG_QUIRK_POINTS_TOOLTIP;
	private static String	MSG_SKILL_POINTS;
	private static String	MSG_SKILL_POINTS_TOOLTIP;
	private static String	MSG_SPELL_POINTS;
	private static String	MSG_SPELL_POINTS_TOOLTIP;
	private static String	MSG_RACE_POINTS;
	private static String	MSG_RACE_POINTS_TOOLTIP;
	private static String	MSG_EARNED_POINTS;
	private static String	MSG_EARNED_POINTS_TOOLTIP;
	private GURPSCharacter	mCharacter;

	static {
		LocalizedMessages.initialize(PointsPanel.class);
	}

	/**
	 * Creates a new points panel.
	 * 
	 * @param character The character to display the data for.
	 */
	public PointsPanel(GURPSCharacter character) {
		super(new ColumnLayout(2, 2, 0, RowDistribution.DISTRIBUTE_HEIGHT), getTitle(character));
		mCharacter = character;
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_ATTRIBUTE_POINTS, MSG_ATTRIBUTE_POINTS, MSG_ATTRIBUTE_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_ADVANTAGE_POINTS, MSG_ADVANTAGE_POINTS, MSG_ADVANTAGE_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_DISADVANTAGE_POINTS, MSG_DISADVANTAGE_POINTS, MSG_DISADVANTAGE_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_QUIRK_POINTS, MSG_QUIRK_POINTS, MSG_QUIRK_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_SKILL_POINTS, MSG_SKILL_POINTS, MSG_SKILL_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_SPELL_POINTS, MSG_SPELL_POINTS, MSG_SPELL_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_RACE_POINTS, MSG_RACE_POINTS, MSG_RACE_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createDivider();
		createLabelAndField(this, character, GURPSCharacter.ID_EARNED_POINTS, MSG_EARNED_POINTS, MSG_EARNED_POINTS_TOOLTIP, SwingConstants.RIGHT);
		mCharacter.addTarget(this, GURPSCharacter.ID_TOTAL_POINTS);
	}

	@Override public Dimension getMaximumSize() {
		Dimension size = super.getMaximumSize();

		size.width = getPreferredSize().width;
		return size;
	}

	private void createDivider() {
		Wrapper panel = new Wrapper();
		UIUtilities.setOnlySize(panel, new Dimension(1, 1));
		add(panel);
		addHorizontalBackground(panel, Color.black);
		panel = new Wrapper();
		UIUtilities.setOnlySize(panel, new Dimension(1, 1));
		add(panel);
	}

	public void handleNotification(Object producer, String type, Object data) {
		getBoxedDropShadowBorder().setTitle(getTitle(mCharacter));
		invalidate();
		repaint();
	}

	private static String getTitle(GURPSCharacter character) {
		return MessageFormat.format(MSG_POINTS, NumberUtils.format(character.getTotalPoints()));
	}
}
