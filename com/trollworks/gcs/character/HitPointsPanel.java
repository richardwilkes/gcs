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
 * Portions created by the Initial Developer are Copyright (C) 1998-2013 the
 * Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.character;

import static com.trollworks.gcs.character.HitPointsPanel_LS.*;

import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.layout.ColumnLayout;
import com.trollworks.ttk.layout.RowDistribution;
import com.trollworks.ttk.utility.UIUtilities;
import com.trollworks.ttk.widgets.Wrapper;

import java.awt.Color;
import java.awt.Container;
import java.awt.Dimension;

import javax.swing.SwingConstants;

@Localized({
				@LS(key = "FP_HP", msg = "Fatigue/Hit Points"),
				@LS(key = "HP", msg = "Basic HP:"),
				@LS(key = "HP_TOOLTIP", msg = "<html><body>Normal (i.e. unharmed) hit points.<br><b>{0} points</b> have been spent to modify <b>HP</b></body></html>"),
				@LS(key = "HP_CURRENT", msg = "Current HP:"),
				@LS(key = "HP_CURRENT_TOOLTIP", msg = "Current hit points"),
				@LS(key = "HP_REELING", msg = "Reeling:"),
				@LS(key = "HP_REELING_TOOLTIP", msg = "<html><body>Current hit points at or below this point indicate the character<br>is reeling from the pain, halving move, speed and dodge</body></html>"),
				@LS(key = "HP_UNCONSCIOUS_CHECKS", msg = "Collapse:"),
				@LS(key = "HP_UNCONSCIOUS_CHECKS_TOOLTIP", msg = "<html><body>Current hit points at or below this point indicate the character<br>is on the verge of collapse, causing the character to <b>roll vs. HT</b><br>(at -1 per full multiple of HP below zero) every second to avoid<br>falling unconscious</body></html>"),
				@LS(key = "HP_DEATH_CHECK_1", msg = "Check #1:"),
				@LS(key = "HP_DEATH_CHECK_1_TOOLTIP", msg = "<html><body>Current hit points at or below this point cause<br>the character to <b>roll vs. HT</b> to avoid death</body></html>"),
				@LS(key = "HP_DEATH_CHECK_2", msg = "Check #2:"),
				@LS(key = "HP_DEATH_CHECK_2_TOOLTIP", msg = "<html><body>Current hit points at or below this point cause<br>the character to <b>roll vs. HT</b> to avoid death</body></html>"),
				@LS(key = "HP_DEATH_CHECK_3", msg = "Check #3:"),
				@LS(key = "HP_DEATH_CHECK_3_TOOLTIP", msg = "<html><body>Current hit points at or below this point cause<br>the character to <b>roll vs. HT</b> to avoid death</body></html>"),
				@LS(key = "HP_DEATH_CHECK_4", msg = "Check #4:"),
				@LS(key = "HP_DEATH_CHECK_4_TOOLTIP", msg = "<html><body>Current hit points at or below this point cause<br>the character to <b>roll vs. HT</b> to avoid death</body></html>"),
				@LS(key = "HP_DEAD", msg = "Dead:"),
				@LS(key = "HP_DEAD_TOOLTIP", msg = "<html><body>Current hit points at or below this<br>point cause the character to die</body></html>"),
				@LS(key = "FP", msg = "Basic FP:"),
				@LS(key = "FP_TOOLTIP", msg = "<html><body>Normal (i.e. fully rested) fatigue points.<br><b>{0} points</b> have been spent to modify <b>FP</b></body></html>"),
				@LS(key = "FP_CURRENT", msg = "Current FP:"),
				@LS(key = "FP_CURRENT_TOOLTIP", msg = "Current fatigue points"),
				@LS(key = "FP_TIRED", msg = "Tired:"),
				@LS(key = "FP_TIRED_TOOLTIP", msg = "<html><body>Current fatigue points at or below this point indicate the<br>character is very tired, halving move, dodge and strength</body></html>"),
				@LS(key = "FP_UNCONSCIOUS_CHECKS", msg = "Collapse:"),
				@LS(key = "FP_UNCONSCIOUS_CHECKS_TOOLTIP", msg = "<html><body>Current fatigue points at or below this point indicate the<br>character is on the verge of collapse, causing the character<br>to roll vs. Will to do anything besides talk or rest</body></html>"),
				@LS(key = "FP_UNCONSCIOUS", msg = "Unconscious:"),
				@LS(key = "FP_UNCONSCIOUS_TOOLTIP", msg = "<html><body>Current fatigue points at or below this point<br>cause the character to fall unconscious</body></html>"),
})
/** The character hit points panel. */
public class HitPointsPanel extends DropPanel {
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
