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
import com.trollworks.toolkit.ui.layout.Alignment;
import com.trollworks.toolkit.ui.layout.FlexComponent;
import com.trollworks.toolkit.ui.layout.FlexGrid;
import com.trollworks.toolkit.ui.layout.FlexRow;
import com.trollworks.toolkit.ui.layout.FlexSpacer;
import com.trollworks.toolkit.ui.widget.Wrapper;
import com.trollworks.toolkit.utility.Localization;

import java.awt.Color;
import java.awt.Dimension;

import javax.swing.SwingConstants;

/** The character attributes panel. */
public class AttributesPanel extends DropPanel {
	@Localize("Attributes")
	private static String	ATTRIBUTES;
	@Localize("Strength (ST):")
	private static String	ST;
	@Localize("<html><body><b>{0} points</b> have been spent to modify <b>Strength</b></body></html>")
	private static String	ST_TOOLTIP;
	@Localize("Dexterity (DX):")
	private static String	DX;
	@Localize("<html><body><b>{0} points</b> have been spent to modify <b>Dexterity</b></body></html>")
	private static String	DX_TOOLTIP;
	@Localize("Intelligence (IQ):")
	private static String	IQ;
	@Localize("<html><body><b>{0} points</b> have been spent to modify <b>Intelligence</b></body></html>")
	private static String	IQ_TOOLTIP;
	@Localize("Health (HT):")
	private static String	HT;
	@Localize("<html><body><b>{0} points</b> have been spent to modify <b>Health</b></body></html>")
	private static String	HT_TOOLTIP;
	@Localize("Will:")
	private static String	WILL;
	@Localize("<html><body><b>{0} points</b> have been spent to modify <b>Will</b></body></html>")
	private static String	WILL_TOOLTIP;
	@Localize("Fright Check:")
	private static String	FRIGHT_CHECK;
	@Localize("Perception:")
	private static String	PERCEPTION;
	@Localize("<html><body><b>{0} points</b> have been spent to modify <b>Perception</b></body></html>")
	private static String	PERCEPTION_TOOLTIP;
	@Localize("Vision:")
	private static String	VISION;
	@Localize("Hearing:")
	private static String	HEARING;
	@Localize("Touch:")
	private static String	TOUCH;
	@Localize("Taste & Smell:")
	private static String	TASTE_SMELL;
	@Localize("Basic Speed:")
	private static String	BASIC_SPEED;
	@Localize("<html><body><b>{0} points</b> have been spent to modify <b>Basic Speed</b></body></html>")
	private static String	BASIC_SPEED_TOOLTIP;
	@Localize("Basic Move:")
	private static String	BASIC_MOVE;
	@Localize("<html><body><b>{0} points</b> have been spent to modify <b>Basic Move</b></body></html>")
	private static String	BASIC_MOVE_TOOLTIP;
	@Localize("thr:")
	private static String	BASIC_THRUST;
	@Localize("The basic damage value for thrust attacks")
	private static String	BASIC_THRUST_TOOLTIP;
	@Localize("sw:")
	private static String	BASIC_SWING;
	@Localize("The basic damage value for swing attacks")
	private static String	BASIC_SWING_TOOLTIP;

	static {
		Localization.initialize();
	}

	/**
	 * Creates a new attributes panel.
	 *
	 * @param character The character to display the data for.
	 */
	public AttributesPanel(GURPSCharacter character) {
		super(null, ATTRIBUTES, true);
		FlexGrid grid = new FlexGrid();
		grid.setVerticalGap(0);
		int row = 0;
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_STRENGTH, ST, ST_TOOLTIP, SwingConstants.RIGHT, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_DEXTERITY, DX, DX_TOOLTIP, SwingConstants.RIGHT, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_INTELLIGENCE, IQ, IQ_TOOLTIP, SwingConstants.RIGHT, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_HEALTH, HT, HT_TOOLTIP, SwingConstants.RIGHT, true);
		createDivider(grid, row++, false);
		createDivider(grid, row++, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_WILL, WILL, WILL_TOOLTIP, SwingConstants.RIGHT, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_FRIGHT_CHECK, FRIGHT_CHECK, null, SwingConstants.RIGHT, false);
		createDivider(grid, row++, false);
		createDivider(grid, row++, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_BASIC_SPEED, BASIC_SPEED, BASIC_SPEED_TOOLTIP, SwingConstants.RIGHT, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_BASIC_MOVE, BASIC_MOVE, BASIC_MOVE_TOOLTIP, SwingConstants.RIGHT, true);
		createDivider(grid, row++, false);
		createDivider(grid, row++, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_PERCEPTION, PERCEPTION, PERCEPTION_TOOLTIP, SwingConstants.RIGHT, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_VISION, VISION, null, SwingConstants.RIGHT, false);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_HEARING, HEARING, null, SwingConstants.RIGHT, false);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_TASTE_AND_SMELL, TASTE_SMELL, null, SwingConstants.RIGHT, false);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_TOUCH, TOUCH, null, SwingConstants.RIGHT, false);
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
		createDamageLabelAndField(row, character, GURPSCharacter.ID_BASIC_THRUST, BASIC_THRUST, BASIC_THRUST_TOOLTIP);
		row.add(new FlexSpacer(0, 0, false, false));
		createDamageLabelAndField(row, character, GURPSCharacter.ID_BASIC_SWING, BASIC_SWING, BASIC_SWING_TOOLTIP);
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
