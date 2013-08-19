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

import com.trollworks.ttk.layout.ColumnLayout;
import com.trollworks.ttk.layout.RowDistribution;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.utility.UIUtilities;
import com.trollworks.ttk.widgets.Wrapper;

import java.awt.Color;
import java.awt.Container;
import java.awt.Dimension;

import javax.swing.SwingConstants;

/** The character hit points panel. */
public class HitPointsPanel extends DropPanel {
	private static String	MSG_FP_HP;
	private static String	MSG_HP;
	private static String	MSG_HP_TOOLTIP;
	private static String	MSG_HP_CURRENT;
	private static String	MSG_HP_CURRENT_TOOLTIP;
	private static String	MSG_HP_REELING;
	private static String	MSG_HP_REELING_TOOLTIP;
	private static String	MSG_HP_UNCONSCIOUS_CHECKS;
	private static String	MSG_HP_UNCONSCIOUS_CHECKS_TOOLTIP;
	private static String	MSG_HP_DEATH_CHECK_1;
	private static String	MSG_HP_DEATH_CHECK_1_TOOLTIP;
	private static String	MSG_HP_DEATH_CHECK_2;
	private static String	MSG_HP_DEATH_CHECK_2_TOOLTIP;
	private static String	MSG_HP_DEATH_CHECK_3;
	private static String	MSG_HP_DEATH_CHECK_3_TOOLTIP;
	private static String	MSG_HP_DEATH_CHECK_4;
	private static String	MSG_HP_DEATH_CHECK_4_TOOLTIP;
	private static String	MSG_HP_DEAD;
	private static String	MSG_HP_DEAD_TOOLTIP;
	private static String	MSG_FP;
	private static String	MSG_FP_TOOLTIP;
	private static String	MSG_FP_CURRENT;
	private static String	MSG_FP_CURRENT_TOOLTIP;
	private static String	MSG_FP_TIRED;
	private static String	MSG_FP_TIRED_TOOLTIP;
	private static String	MSG_FP_UNCONSCIOUS_CHECKS;
	private static String	MSG_FP_UNCONSCIOUS_CHECKS_TOOLTIP;
	private static String	MSG_FP_UNCONSCIOUS;
	private static String	MSG_FP_UNCONSCIOUS_TOOLTIP;

	static {
		LocalizedMessages.initialize(HitPointsPanel.class);
	}

	/**
	 * Creates a new hit points panel.
	 * 
	 * @param character The character to display the data for.
	 */
	public HitPointsPanel(GURPSCharacter character) {
		super(new ColumnLayout(2, 2, 0, RowDistribution.DISTRIBUTE_HEIGHT), MSG_FP_HP);
		createLabelAndField(this, character, GURPSCharacter.ID_CURRENT_FATIGUE_POINTS, MSG_FP_CURRENT, MSG_FP_CURRENT_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndField(this, character, GURPSCharacter.ID_FATIGUE_POINTS, MSG_FP, MSG_FP_TOOLTIP, SwingConstants.RIGHT);
		createDivider();
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_TIRED_FATIGUE_POINTS, MSG_FP_TIRED, MSG_FP_TIRED_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_UNCONSCIOUS_CHECKS_FATIGUE_POINTS, MSG_FP_UNCONSCIOUS_CHECKS, MSG_FP_UNCONSCIOUS_CHECKS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_UNCONSCIOUS_FATIGUE_POINTS, MSG_FP_UNCONSCIOUS, MSG_FP_UNCONSCIOUS_TOOLTIP, SwingConstants.RIGHT);
		createDivider();
		createLabelAndField(this, character, GURPSCharacter.ID_CURRENT_HIT_POINTS, MSG_HP_CURRENT, MSG_HP_CURRENT_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndField(this, character, GURPSCharacter.ID_HIT_POINTS, MSG_HP, MSG_HP_TOOLTIP, SwingConstants.RIGHT);
		createDivider();
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_REELING_HIT_POINTS, MSG_HP_REELING, MSG_HP_REELING_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_UNCONSCIOUS_CHECKS_HIT_POINTS, MSG_HP_UNCONSCIOUS_CHECKS, MSG_HP_UNCONSCIOUS_CHECKS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_DEATH_CHECK_1_HIT_POINTS, MSG_HP_DEATH_CHECK_1, MSG_HP_DEATH_CHECK_1_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_DEATH_CHECK_2_HIT_POINTS, MSG_HP_DEATH_CHECK_2, MSG_HP_DEATH_CHECK_2_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_DEATH_CHECK_3_HIT_POINTS, MSG_HP_DEATH_CHECK_3, MSG_HP_DEATH_CHECK_3_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_DEATH_CHECK_4_HIT_POINTS, MSG_HP_DEATH_CHECK_4, MSG_HP_DEATH_CHECK_4_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_DEAD_HIT_POINTS, MSG_HP_DEAD, MSG_HP_DEAD_TOOLTIP, SwingConstants.RIGHT);
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
