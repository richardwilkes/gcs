/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.notification.NotifierTarget;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.text.Text;

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
        super(new PrecisionLayout().setColumns(2).setMargins(0).setSpacing(2, 0).setAlignment(PrecisionLayoutAlignment.FILL, PrecisionLayoutAlignment.FILL), getTitle(sheet.getCharacter()));
        mSheet = sheet;
        createLabelAndField(sheet, GURPSCharacter.ID_UNSPENT_POINTS, I18n.Text("Unspent"), I18n.Text("Points that have been earned but have not yet been spent"), true);
        createLabelAndField(sheet, GURPSCharacter.ID_RACE_POINTS, I18n.Text("Race"), I18n.Text("A summary of all points spent on a racial package for this character"), false);
        createLabelAndField(sheet, GURPSCharacter.ID_ATTRIBUTE_POINTS, I18n.Text("Attributes"), I18n.Text("A summary of all points spent on attributes for this character"), false);
        createLabelAndField(sheet, GURPSCharacter.ID_ADVANTAGE_POINTS, I18n.Text("Advantages"), I18n.Text("A summary of all points spent on advantages for this character"), false);
        createLabelAndField(sheet, GURPSCharacter.ID_DISADVANTAGE_POINTS, I18n.Text("Disadvantages"), I18n.Text("A summary of all points spent on disadvantages for this character"), false);
        createLabelAndField(sheet, GURPSCharacter.ID_QUIRK_POINTS, I18n.Text("Quirks"), I18n.Text("A summary of all points spent on quirks for this character"), false);
        createLabelAndField(sheet, GURPSCharacter.ID_SKILL_POINTS, I18n.Text("Skills"), I18n.Text("A summary of all points spent on skills for this character"), false);
        createLabelAndField(sheet, GURPSCharacter.ID_SPELL_POINTS, I18n.Text("Spells"), I18n.Text("A summary of all points spent on spells for this character"), false);
        sheet.getCharacter().addTarget(this, GURPSCharacter.ID_TOTAL_POINTS);
        Preferences.getInstance().getNotifier().add(this, Preferences.KEY_INCLUDE_UNSPENT_POINTS_IN_TOTAL);
    }

    private void createLabelAndField(CharacterSheet sheet, String key, String title, String tooltip, boolean enabled) {
        PageField field = new PageField(sheet, key, SwingConstants.RIGHT, enabled, Text.wrapPlainTextForToolTip(tooltip));
        add(field, new PrecisionLayoutData().setGrabHorizontalSpace(true).setHorizontalAlignment(PrecisionLayoutAlignment.FILL));
        add(new PageLabel(title, field));
    }

    @Override
    public void handleNotification(Object producer, String type, Object data) {
        getTitledBorder().setTitle(getTitle(mSheet.getCharacter()));
        invalidate();
        repaint();
    }

    private static String getTitle(GURPSCharacter character) {
        return MessageFormat.format(I18n.Text("{0} Points"), Numbers.format(Preferences.getInstance().includeUnspentPointsInTotal() ? character.getTotalPoints() : character.getSpentPoints()));
    }

    @Override
    public int getNotificationPriority() {
        return 0;
    }
}
