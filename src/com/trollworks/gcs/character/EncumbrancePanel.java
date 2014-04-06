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
import com.trollworks.toolkit.ui.widget.Wrapper;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.notification.NotifierTarget;
import com.trollworks.toolkit.utility.text.Numbers;

import java.awt.Color;
import java.awt.Container;
import java.awt.Dimension;
import java.text.MessageFormat;

import javax.swing.SwingConstants;

/** The character encumbrance panel. */
public class EncumbrancePanel extends DropPanel implements NotifierTarget {
	@Localize("Encumbrance, Move & Dodge")
	private static String			ENCUMBRANCE_MOVE_DODGE;
	@Localize("Level")
	private static String			ENCUMBRANCE_LEVEL;
	@Localize("Max Load")
	private static String			MAX_CARRY;
	@Localize("Move")
	private static String			MOVE;
	@Localize("Dodge")
	private static String			DODGE;
	@Localize("The encumbrance level")
	private static String			ENCUMBRANCE_TOOLTIP;
	@Localize("The maximum load a character can carry and still remain within a specific encumbrance level")
	private static String			MAX_CARRY_TOOLTIP;
	@Localize("The character's ground movement rate for a specific encumbrance level")
	private static String			MOVE_TOOLTIP;
	@Localize("The character's dodge for a specific encumbrance level")
	private static String			DODGE_TOOLTIP;
	@Localize("None")
	private static String			NONE;
	@Localize("Light")
	private static String			LIGHT;
	@Localize("Medium")
	private static String			MEDIUM;
	@Localize("Heavy")
	private static String			HEAVY;
	@Localize("X-Heavy")
	private static String			EXTRA_HEAVY;
	@Localize("{0} ({1})")
	static String					ENCUMBRANCE_FORMAT;
	@Localize("\u2022 {0} ({1})")
	static String					CURRENT_ENCUMBRANCE_FORMAT;

	static {
		Localization.initialize();
	}

	private static final Color		CURRENT_ENCUMBRANCE_COLOR	= new Color(252, 242, 196);
	/** The various encumbrance titles. */
	public static final String[]	ENCUMBRANCE_TITLES			= new String[] { NONE, LIGHT, MEDIUM, HEAVY, EXTRA_HEAVY };
	private GURPSCharacter			mCharacter;
	private PageLabel[]				mMarkers;

	/**
	 * Creates a new encumbrance panel.
	 *
	 * @param character The character to display the data for.
	 */
	public EncumbrancePanel(GURPSCharacter character) {
		super(new ColumnLayout(7, 2, 0), ENCUMBRANCE_MOVE_DODGE, true);
		mCharacter = character;
		mMarkers = new PageLabel[GURPSCharacter.ENCUMBRANCE_LEVELS];
		PageHeader header = createHeader(this, ENCUMBRANCE_LEVEL, ENCUMBRANCE_TOOLTIP);
		addHorizontalBackground(header, Color.black);
		addVerticalBackground(createDivider(), Color.black);
		createHeader(this, MAX_CARRY, MAX_CARRY_TOOLTIP);
		addVerticalBackground(createDivider(), Color.black);
		createHeader(this, MOVE, MOVE_TOOLTIP);
		addVerticalBackground(createDivider(), Color.black);
		createHeader(this, DODGE, DODGE_TOOLTIP);
		int current = character.getEncumbranceLevel();
		for (int i = 0; i < GURPSCharacter.ENCUMBRANCE_LEVELS; i++) {
			mMarkers[i] = new PageLabel(getMarkerText(i, current), header);
			add(mMarkers[i]);
			if (current == i) {
				addHorizontalBackground(mMarkers[i], CURRENT_ENCUMBRANCE_COLOR);
			}
			createDivider();
			createDisabledField(this, character, GURPSCharacter.MAXIMUM_CARRY_PREFIX + i, MAX_CARRY_TOOLTIP, SwingConstants.RIGHT);
			createDivider();
			createDisabledField(this, character, GURPSCharacter.MOVE_PREFIX + i, MOVE_TOOLTIP, SwingConstants.RIGHT);
			createDivider();
			createDisabledField(this, character, GURPSCharacter.DODGE_PREFIX + i, DODGE_TOOLTIP, SwingConstants.RIGHT);
		}
		character.addTarget(this, GURPSCharacter.ID_CARRIED_WEIGHT, GURPSCharacter.ID_BASIC_LIFT);
	}

	private static String getMarkerText(int which, int current) {
		return MessageFormat.format(which == current ? CURRENT_ENCUMBRANCE_FORMAT : ENCUMBRANCE_FORMAT, ENCUMBRANCE_TITLES[which], Numbers.format(which));
	}

	private Container createDivider() {
		Wrapper panel = new Wrapper();
		UIUtilities.setOnlySize(panel, new Dimension(1, 1));
		add(panel);
		return panel;
	}

	@Override
	public void handleNotification(Object producer, String type, Object data) {
		int current = mCharacter.getEncumbranceLevel();
		for (int i = 0; i < GURPSCharacter.ENCUMBRANCE_LEVELS; i++) {
			if (i == current) {
				addHorizontalBackground(mMarkers[i], CURRENT_ENCUMBRANCE_COLOR);
			} else {
				removeHorizontalBackground(mMarkers[i]);
			}
			mMarkers[i].setText(getMarkerText(i, current));
		}
		revalidate();
		repaint();
	}

	@Override
	public int getNotificationPriority() {
		return 0;
	}
}
