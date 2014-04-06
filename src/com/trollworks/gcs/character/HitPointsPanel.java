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

import java.awt.Color;
import java.awt.Container;
import java.awt.Dimension;

import javax.swing.SwingConstants;

/** The character hit points panel. */
public class HitPointsPanel extends DropPanel {
	@Localize("Fatigue/Hit Points")
	private static String	FP_HP;
	@Localize("Basic HP:")
	private static String	HP;
	@Localize("<html><body>Normal (i.e. unharmed) hit points.<br><b>{0} points</b> have been spent to modify <b>HP</b></body></html>")
	private static String	HP_TOOLTIP;
	@Localize("Current HP:")
	private static String	HP_CURRENT;
	@Localize("Current hit points")
	private static String	HP_CURRENT_TOOLTIP;
	@Localize("Reeling:")
	private static String	HP_REELING;
	@Localize("<html><body>Current hit points at or below this point indicate the character<br>is reeling from the pain, halving move, speed and dodge</body></html>")
	private static String	HP_REELING_TOOLTIP;
	@Localize("Collapse:")
	private static String	HP_UNCONSCIOUS_CHECKS;
	@Localize("<html><body>Current hit points at or below this point indicate the character<br>is on the verge of collapse, causing the character to <b>roll vs. HT</b><br>(at -1 per full multiple of HP below zero) every second to avoid<br>falling unconscious</body></html>")
	private static String	HP_UNCONSCIOUS_CHECKS_TOOLTIP;
	@Localize("Check #1:")
	private static String	HP_DEATH_CHECK_1;
	@Localize("<html><body>Current hit points at or below this point cause<br>the character to <b>roll vs. HT</b> to avoid death</body></html>")
	private static String	HP_DEATH_CHECK_1_TOOLTIP;
	@Localize("Check #2:")
	private static String	HP_DEATH_CHECK_2;
	@Localize("<html><body>Current hit points at or below this point cause<br>the character to <b>roll vs. HT</b> to avoid death</body></html>")
	private static String	HP_DEATH_CHECK_2_TOOLTIP;
	@Localize("Check #3:")
	private static String	HP_DEATH_CHECK_3;
	@Localize("<html><body>Current hit points at or below this point cause<br>the character to <b>roll vs. HT</b> to avoid death</body></html>")
	private static String	HP_DEATH_CHECK_3_TOOLTIP;
	@Localize("Check #4:")
	private static String	HP_DEATH_CHECK_4;
	@Localize("<html><body>Current hit points at or below this point cause<br>the character to <b>roll vs. HT</b> to avoid death</body></html>")
	private static String	HP_DEATH_CHECK_4_TOOLTIP;
	@Localize("Dead:")
	private static String	HP_DEAD;
	@Localize("<html><body>Current hit points at or below this<br>point cause the character to die</body></html>")
	private static String	HP_DEAD_TOOLTIP;
	@Localize("Basic FP:")
	private static String	FP;
	@Localize("<html><body>Normal (i.e. fully rested) fatigue points.<br><b>{0} points</b> have been spent to modify <b>FP</b></body></html>")
	private static String	FP_TOOLTIP;
	@Localize("Current FP:")
	private static String	FP_CURRENT;
	@Localize("Current fatigue points")
	private static String	FP_CURRENT_TOOLTIP;
	@Localize("Tired:")
	private static String	FP_TIRED;
	@Localize("<html><body>Current fatigue points at or below this point indicate the<br>character is very tired, halving move, dodge and strength</body></html>")
	private static String	FP_TIRED_TOOLTIP;
	@Localize("Collapse:")
	private static String	FP_UNCONSCIOUS_CHECKS;
	@Localize("<html><body>Current fatigue points at or below this point indicate the<br>character is on the verge of collapse, causing the character<br>to roll vs. Will to do anything besides talk or rest</body></html>")
	private static String	FP_UNCONSCIOUS_CHECKS_TOOLTIP;
	@Localize("Unconscious:")
	private static String	FP_UNCONSCIOUS;
	@Localize("<html><body>Current fatigue points at or below this point<br>cause the character to fall unconscious</body></html>")
	private static String	FP_UNCONSCIOUS_TOOLTIP;

	static {
		Localization.initialize();
	}

	/**
	 * Creates a new hit points panel.
	 *
	 * @param character The character to display the data for.
	 */
	public HitPointsPanel(GURPSCharacter character) {
		super(new ColumnLayout(2, 2, 0, RowDistribution.DISTRIBUTE_HEIGHT), FP_HP);
		createLabelAndField(this, character, GURPSCharacter.ID_CURRENT_FATIGUE_POINTS, FP_CURRENT, FP_CURRENT_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndField(this, character, GURPSCharacter.ID_FATIGUE_POINTS, FP, FP_TOOLTIP, SwingConstants.RIGHT);
		createDivider();
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_TIRED_FATIGUE_POINTS, FP_TIRED, FP_TIRED_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_UNCONSCIOUS_CHECKS_FATIGUE_POINTS, FP_UNCONSCIOUS_CHECKS, FP_UNCONSCIOUS_CHECKS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_UNCONSCIOUS_FATIGUE_POINTS, FP_UNCONSCIOUS, FP_UNCONSCIOUS_TOOLTIP, SwingConstants.RIGHT);
		createDivider();
		createLabelAndField(this, character, GURPSCharacter.ID_CURRENT_HIT_POINTS, HP_CURRENT, HP_CURRENT_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndField(this, character, GURPSCharacter.ID_HIT_POINTS, HP, HP_TOOLTIP, SwingConstants.RIGHT);
		createDivider();
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_REELING_HIT_POINTS, HP_REELING, HP_REELING_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_UNCONSCIOUS_CHECKS_HIT_POINTS, HP_UNCONSCIOUS_CHECKS, HP_UNCONSCIOUS_CHECKS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_DEATH_CHECK_1_HIT_POINTS, HP_DEATH_CHECK_1, HP_DEATH_CHECK_1_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_DEATH_CHECK_2_HIT_POINTS, HP_DEATH_CHECK_2, HP_DEATH_CHECK_2_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_DEATH_CHECK_3_HIT_POINTS, HP_DEATH_CHECK_3, HP_DEATH_CHECK_3_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_DEATH_CHECK_4_HIT_POINTS, HP_DEATH_CHECK_4, HP_DEATH_CHECK_4_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_DEAD_HIT_POINTS, HP_DEAD, HP_DEAD_TOOLTIP, SwingConstants.RIGHT);
	}

	@Override
	public Dimension getMaximumSize() {
		Dimension size = super.getMaximumSize();

		size.width = getPreferredSize().width;
		return size;
	}

	private void createDivider() {
		createOneByOnePanel();
		createOneByOnePanel();
		addHorizontalBackground(createOneByOnePanel(), Color.black);
		createOneByOnePanel();
	}

	private Container createOneByOnePanel() {
		Wrapper panel = new Wrapper();
		UIUtilities.setOnlySize(panel, new Dimension(1, 1));
		add(panel);
		return panel;
	}
}
