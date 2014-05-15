/*
 * Copyright (c) 1998-2014 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.layout.ColumnLayout;
import com.trollworks.toolkit.ui.layout.RowDistribution;
import com.trollworks.toolkit.ui.widget.Wrapper;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.notification.NotifierTarget;
import com.trollworks.toolkit.utility.text.Numbers;

import java.awt.Color;
import java.awt.Dimension;
import java.text.MessageFormat;

import javax.swing.SwingConstants;

/** The character points panel. */
public class PointsPanel extends DropPanel implements NotifierTarget {
	@Localize("{0} Points")
	private static String	POINTS;
	@Localize("Attributes:")
	private static String	ATTRIBUTE_POINTS;
	@Localize("A summary of all points spent on attributes for this character")
	private static String	ATTRIBUTE_POINTS_TOOLTIP;
	@Localize("Advantages:")
	private static String	ADVANTAGE_POINTS;
	@Localize("A summary of all points spent on advantages for this character")
	private static String	ADVANTAGE_POINTS_TOOLTIP;
	@Localize("Disadvantages:")
	private static String	DISADVANTAGE_POINTS;
	@Localize("A summary of all points spent on disadvantages for this character")
	private static String	DISADVANTAGE_POINTS_TOOLTIP;
	@Localize("Quirks:")
	private static String	QUIRK_POINTS;
	@Localize("A summary of all points spent on quirks for this character")
	private static String	QUIRK_POINTS_TOOLTIP;
	@Localize("Skills:")
	private static String	SKILL_POINTS;
	@Localize("A summary of all points spent on skills for this character")
	private static String	SKILL_POINTS_TOOLTIP;
	@Localize("Spells:")
	private static String	SPELL_POINTS;
	@Localize("A summary of all points spent on spells for this character")
	private static String	SPELL_POINTS_TOOLTIP;
	@Localize("Race:")
	private static String	RACE_POINTS;
	@Localize("A summary of all points spent on a racial package for this character")
	private static String	RACE_POINTS_TOOLTIP;
	@Localize("Earned:")
	private static String	EARNED_POINTS;
	@Localize("Points that have been earned but not yet been spent")
	private static String	EARNED_POINTS_TOOLTIP;

	static {
		Localization.initialize();
	}

	private GURPSCharacter	mCharacter;

	/**
	 * Creates a new points panel.
	 *
	 * @param character The character to display the data for.
	 */
	public PointsPanel(GURPSCharacter character) {
		super(new ColumnLayout(2, 2, 0, RowDistribution.DISTRIBUTE_HEIGHT), getTitle(character));
		mCharacter = character;
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_RACE_POINTS, RACE_POINTS, RACE_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_ATTRIBUTE_POINTS, ATTRIBUTE_POINTS, ATTRIBUTE_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_ADVANTAGE_POINTS, ADVANTAGE_POINTS, ADVANTAGE_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_DISADVANTAGE_POINTS, DISADVANTAGE_POINTS, DISADVANTAGE_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_QUIRK_POINTS, QUIRK_POINTS, QUIRK_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_SKILL_POINTS, SKILL_POINTS, SKILL_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_SPELL_POINTS, SPELL_POINTS, SPELL_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createDivider();
		createLabelAndField(this, character, GURPSCharacter.ID_EARNED_POINTS, EARNED_POINTS, EARNED_POINTS_TOOLTIP, SwingConstants.RIGHT);
		mCharacter.addTarget(this, GURPSCharacter.ID_TOTAL_POINTS);
	}

	@Override
	public Dimension getMaximumSize() {
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

	@Override
	public void handleNotification(Object producer, String type, Object data) {
		getBoxedDropShadowBorder().setTitle(getTitle(mCharacter));
		invalidate();
		repaint();
	}

	private static String getTitle(GURPSCharacter character) {
		return MessageFormat.format(POINTS, Numbers.format(character.getTotalPoints()));
	}

	@Override
	public int getNotificationPriority() {
		return 0;
	}
}
