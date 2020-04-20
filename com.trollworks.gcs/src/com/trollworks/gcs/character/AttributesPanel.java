/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.gcs.page.PageField;
import com.trollworks.gcs.page.PageLabel;
import com.trollworks.gcs.ui.layout.Alignment;
import com.trollworks.gcs.ui.layout.FlexComponent;
import com.trollworks.gcs.ui.layout.FlexGrid;
import com.trollworks.gcs.ui.layout.FlexRow;
import com.trollworks.gcs.ui.layout.FlexSpacer;
import com.trollworks.gcs.ui.widget.Wrapper;
import com.trollworks.gcs.utility.I18n;

import java.awt.Color;
import javax.swing.SwingConstants;

/** The character attributes panel. */
public class AttributesPanel extends DropPanel {
    private CharacterSheet mSheet;

    /**
     * Creates a new attributes panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public AttributesPanel(CharacterSheet sheet) {
        super(null, I18n.Text("Attributes"), true);
        mSheet = sheet;
        FlexGrid grid = new FlexGrid();
        grid.setVerticalGap(0);
        int row = 0;
        createLabelAndField(grid, row++, GURPSCharacter.ID_STRENGTH, I18n.Text("Strength (ST):"), I18n.Text("<html><body><b>{0} points</b> have been spent to modify <b>Strength</b></body></html>"), SwingConstants.RIGHT, true);
        createLabelAndField(grid, row++, GURPSCharacter.ID_DEXTERITY, I18n.Text("Dexterity (DX):"), I18n.Text("<html><body><b>{0} points</b> have been spent to modify <b>Dexterity</b></body></html>"), SwingConstants.RIGHT, true);
        createLabelAndField(grid, row++, GURPSCharacter.ID_INTELLIGENCE, I18n.Text("Intelligence (IQ):"), I18n.Text("<html><body><b>{0} points</b> have been spent to modify <b>Intelligence</b></body></html>"), SwingConstants.RIGHT, true);
        createLabelAndField(grid, row++, GURPSCharacter.ID_HEALTH, I18n.Text("Health (HT):"), I18n.Text("<html><body><b>{0} points</b> have been spent to modify <b>Health</b></body></html>"), SwingConstants.RIGHT, true);
        createDivider(grid, row++, false);
        createDivider(grid, row++, true);
        createLabelAndField(grid, row++, GURPSCharacter.ID_WILL, I18n.Text("Will:"), I18n.Text("<html><body><b>{0} points</b> have been spent to modify <b>Will</b></body></html>"), SwingConstants.RIGHT, true);
        createLabelAndField(grid, row++, GURPSCharacter.ID_FRIGHT_CHECK, I18n.Text("Fright Check:"), null, SwingConstants.RIGHT, false);
        createDivider(grid, row++, false);
        createDivider(grid, row++, true);
        createLabelAndField(grid, row++, GURPSCharacter.ID_BASIC_SPEED, I18n.Text("Basic Speed:"), I18n.Text("<html><body><b>{0} points</b> have been spent to modify <b>Basic Speed</b></body></html>"), SwingConstants.RIGHT, true);
        createLabelAndField(grid, row++, GURPSCharacter.ID_BASIC_MOVE, I18n.Text("Basic Move:"), I18n.Text("<html><body><b>{0} points</b> have been spent to modify <b>Basic Move</b></body></html>"), SwingConstants.RIGHT, true);
        createDivider(grid, row++, false);
        createDivider(grid, row++, true);
        createLabelAndField(grid, row++, GURPSCharacter.ID_PERCEPTION, I18n.Text("Perception:"), I18n.Text("<html><body><b>{0} points</b> have been spent to modify <b>Perception</b></body></html>"), SwingConstants.RIGHT, true);
        createLabelAndField(grid, row++, GURPSCharacter.ID_VISION, I18n.Text("Vision:"), null, SwingConstants.RIGHT, false);
        createLabelAndField(grid, row++, GURPSCharacter.ID_HEARING, I18n.Text("Hearing:"), null, SwingConstants.RIGHT, false);
        createLabelAndField(grid, row++, GURPSCharacter.ID_TASTE_AND_SMELL, I18n.Text("Taste & Smell:"), null, SwingConstants.RIGHT, false);
        createLabelAndField(grid, row++, GURPSCharacter.ID_TOUCH, I18n.Text("Touch:"), null, SwingConstants.RIGHT, false);
        createDivider(grid, row++, false);
        createDivider(grid, row++, true);
        createDamageFields(grid, row);
        grid.apply(this);
    }

    private void createLabelAndField(FlexGrid grid, int row, String key, String title, String tooltip, int alignment, boolean enabled) {
        PageField field = new PageField(mSheet, key, alignment, enabled, tooltip);
        PageLabel label = new PageLabel(title, field);
        add(label);
        add(field);
        grid.add(new FlexComponent(label, Alignment.RIGHT_BOTTOM, Alignment.CENTER), row, 0);
        grid.add(field, row, 1);
    }

    private void createDamageFields(FlexGrid grid, int rowIndex) {
        FlexRow row = new FlexRow();
        row.setHorizontalAlignment(Alignment.CENTER);
        createDamageLabelAndField(row, GURPSCharacter.ID_BASIC_THRUST, I18n.Text("thr:"), I18n.Text("The basic damage value for thrust attacks"));
        row.add(new FlexSpacer(0, 0, false, false));
        createDamageLabelAndField(row, GURPSCharacter.ID_BASIC_SWING, I18n.Text("sw:"), I18n.Text("The basic damage value for swing attacks"));
        grid.add(row, rowIndex, 0, 1, 2);
    }

    private void createDamageLabelAndField(FlexRow row, String key, String title, String tooltip) {
        PageField field = new PageField(mSheet, key, SwingConstants.RIGHT, false, tooltip);
        PageLabel label = new PageLabel(title, field);
        add(label);
        add(field);
        row.add(new FlexComponent(label, true));
        row.add(new FlexComponent(field, true));
    }

    private void createDivider(FlexGrid grid, int row, boolean black) {
        Wrapper panel = new Wrapper();
        panel.setOnlySize(1, 1);
        add(panel);
        if (black) {
            addHorizontalBackground(panel, Color.black);
        }
        grid.add(panel, row, 0, 1, 2);
    }
}
