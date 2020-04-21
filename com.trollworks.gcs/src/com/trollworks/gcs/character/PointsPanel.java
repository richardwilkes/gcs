/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License, version 2.0.
 * If a copy of the MPL was not distributed with this file, You can obtain one at
 * http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined by the
 * Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character;

import com.trollworks.gcs.page.DropPanel;
import com.trollworks.gcs.preferences.DisplayPreferences;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.ui.layout.RowDistribution;
import com.trollworks.gcs.ui.widget.Wrapper;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Preferences;
import com.trollworks.gcs.utility.notification.NotifierTarget;
import com.trollworks.gcs.utility.text.Numbers;

import java.awt.Color;
import java.awt.Dimension;
import java.text.MessageFormat;
import javax.swing.SwingConstants;

/** The character points panel. */
public class PointsPanel extends DropPanel implements NotifierTarget {
    private CharacterSheet mSheet;

    /**
     * Creates a new points panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public PointsPanel(CharacterSheet sheet) {
        super(new ColumnLayout(2, 2, 0, RowDistribution.DISTRIBUTE_HEIGHT), getTitle(sheet.getCharacter()));
        mSheet = sheet;
        createLabelAndDisabledField(this, sheet, GURPSCharacter.ID_RACE_POINTS, I18n.Text("Race:"), I18n.Text("A summary of all points spent on a racial package for this character"), SwingConstants.RIGHT);
        createLabelAndDisabledField(this, sheet, GURPSCharacter.ID_ATTRIBUTE_POINTS, I18n.Text("Attributes:"), I18n.Text("A summary of all points spent on attributes for this character"), SwingConstants.RIGHT);
        createLabelAndDisabledField(this, sheet, GURPSCharacter.ID_ADVANTAGE_POINTS, I18n.Text("Advantages:"), I18n.Text("A summary of all points spent on advantages for this character"), SwingConstants.RIGHT);
        createLabelAndDisabledField(this, sheet, GURPSCharacter.ID_DISADVANTAGE_POINTS, I18n.Text("Disadvantages:"), I18n.Text("A summary of all points spent on disadvantages for this character"), SwingConstants.RIGHT);
        createLabelAndDisabledField(this, sheet, GURPSCharacter.ID_QUIRK_POINTS, I18n.Text("Quirks:"), I18n.Text("A summary of all points spent on quirks for this character"), SwingConstants.RIGHT);
        createLabelAndDisabledField(this, sheet, GURPSCharacter.ID_SKILL_POINTS, I18n.Text("Skills:"), I18n.Text("A summary of all points spent on skills for this character"), SwingConstants.RIGHT);
        createLabelAndDisabledField(this, sheet, GURPSCharacter.ID_SPELL_POINTS, I18n.Text("Spells:"), I18n.Text("A summary of all points spent on spells for this character"), SwingConstants.RIGHT);
        createDivider();
        createLabelAndField(this, sheet, GURPSCharacter.ID_UNSPENT_POINTS, I18n.Text("Unspent:"), I18n.Text("Points that have been earned but have not yet been spent"), SwingConstants.RIGHT);
        sheet.getCharacter().addTarget(this, GURPSCharacter.ID_TOTAL_POINTS);
        Preferences.getInstance().getNotifier().add(this, DisplayPreferences.TOTAL_POINTS_DISPLAY_PREF_KEY);
    }

    @Override
    public Dimension getMaximumSize() {
        Dimension size = super.getMaximumSize();
        size.width = getPreferredSize().width;
        return size;
    }

    private void createDivider() {
        Wrapper panel = new Wrapper();
        panel.setOnlySize(1, 1);
        add(panel);
        addHorizontalBackground(panel, Color.black);
        panel = new Wrapper();
        panel.setOnlySize(1, 1);
        add(panel);
    }

    @Override
    public void handleNotification(Object producer, String type, Object data) {
        getTitledBorder().setTitle(getTitle(mSheet.getCharacter()));
        invalidate();
        repaint();
    }

    private static String getTitle(GURPSCharacter character) {
        return MessageFormat.format(I18n.Text("{0} Points"), Numbers.format(DisplayPreferences.shouldIncludeUnspentPointsInTotalPointDisplay() ? character.getTotalPoints() : character.getSpentPoints()));
    }

    @Override
    public int getNotificationPriority() {
        return 0;
    }
}
