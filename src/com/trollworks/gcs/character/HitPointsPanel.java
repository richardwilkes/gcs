/*
 * Copyright (c) 1998-2019 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character;

import com.trollworks.gcs.page.DropPanel;
import com.trollworks.toolkit.ui.layout.ColumnLayout;
import com.trollworks.toolkit.ui.layout.RowDistribution;
import com.trollworks.toolkit.ui.widget.Wrapper;
import com.trollworks.toolkit.utility.I18n;

import java.awt.Color;
import java.awt.Container;
import java.awt.Dimension;
import javax.swing.SwingConstants;

/** The character hit points panel. */
public class HitPointsPanel extends DropPanel {
    /**
     * Creates a new hit points panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public HitPointsPanel(CharacterSheet sheet) {
        super(new ColumnLayout(2, 2, 0, RowDistribution.DISTRIBUTE_HEIGHT), I18n.Text("Fatigue/Hit Points"));
        createLabelAndField(this, sheet, GURPSCharacter.ID_CURRENT_FATIGUE_POINTS, I18n.Text("Current FP:"), I18n.Text("Current fatigue points"), SwingConstants.RIGHT);
        createLabelAndField(this, sheet, GURPSCharacter.ID_FATIGUE_POINTS, I18n.Text("Basic FP:"), I18n.Text("<html><body>Normal (i.e. fully rested) fatigue points.<br><b>{0} points</b> have been spent to modify <b>FP</b></body></html>"), SwingConstants.RIGHT);
        createDivider();
        createLabelAndDisabledField(this, sheet, GURPSCharacter.ID_TIRED_FATIGUE_POINTS, I18n.Text("Tired:"), I18n.Text("<html><body>Current fatigue points at or below this point indicate the<br>character is very tired, halving move, dodge and strength</body></html>"), SwingConstants.RIGHT);
        createLabelAndDisabledField(this, sheet, GURPSCharacter.ID_UNCONSCIOUS_CHECKS_FATIGUE_POINTS, I18n.Text("Collapse:"), I18n.Text("<html><body>Current fatigue points at or below this point indicate the<br>character is on the verge of collapse, causing the character<br>to roll vs. Will to do anything besides talk or rest</body></html>"), SwingConstants.RIGHT);
        createLabelAndDisabledField(this, sheet, GURPSCharacter.ID_UNCONSCIOUS_FATIGUE_POINTS, I18n.Text("Unconscious:"), I18n.Text("<html><body>Current fatigue points at or below this point<br>cause the character to fall unconscious</body></html>"), SwingConstants.RIGHT);
        createDivider();
        createLabelAndField(this, sheet, GURPSCharacter.ID_CURRENT_HIT_POINTS, I18n.Text("Current HP:"), I18n.Text("Current hit points"), SwingConstants.RIGHT);
        createLabelAndField(this, sheet, GURPSCharacter.ID_HIT_POINTS, I18n.Text("Basic HP:"), I18n.Text("<html><body>Normal (i.e. unharmed) hit points.<br><b>{0} points</b> have been spent to modify <b>HP</b></body></html>"), SwingConstants.RIGHT);
        createDivider();
        createLabelAndDisabledField(this, sheet, GURPSCharacter.ID_REELING_HIT_POINTS, I18n.Text("Reeling:"), I18n.Text("<html><body>Current hit points at or below this point indicate the character<br>is reeling from the pain, halving move, speed and dodge</body></html>"), SwingConstants.RIGHT);
        createLabelAndDisabledField(this, sheet, GURPSCharacter.ID_UNCONSCIOUS_CHECKS_HIT_POINTS, I18n.Text("Collapse:"), I18n.Text("<html><body>Current hit points at or below this point indicate the character<br>is on the verge of collapse, causing the character to <b>roll vs. HT</b><br>(at -1 per full multiple of HP below zero) every second to avoid<br>falling unconscious</body></html>"), SwingConstants.RIGHT);
        String deathCheckTooltip = I18n.Text("<html><body>Current hit points at or below this point cause<br>the character to <b>roll vs. HT</b> to avoid death</body></html>");
        createLabelAndDisabledField(this, sheet, GURPSCharacter.ID_DEATH_CHECK_1_HIT_POINTS, I18n.Text("Check #1:"), deathCheckTooltip, SwingConstants.RIGHT);
        createLabelAndDisabledField(this, sheet, GURPSCharacter.ID_DEATH_CHECK_2_HIT_POINTS, I18n.Text("Check #2:"), deathCheckTooltip, SwingConstants.RIGHT);
        createLabelAndDisabledField(this, sheet, GURPSCharacter.ID_DEATH_CHECK_3_HIT_POINTS, I18n.Text("Check #3:"), deathCheckTooltip, SwingConstants.RIGHT);
        createLabelAndDisabledField(this, sheet, GURPSCharacter.ID_DEATH_CHECK_4_HIT_POINTS, I18n.Text("Check #4:"), deathCheckTooltip, SwingConstants.RIGHT);
        createLabelAndDisabledField(this, sheet, GURPSCharacter.ID_DEAD_HIT_POINTS, I18n.Text("Dead:"), I18n.Text("<html><body>Current hit points at or below this<br>point cause the character to die</body></html>"), SwingConstants.RIGHT);
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
        panel.setOnlySize(1, 1);
        add(panel);
        return panel;
    }
}
