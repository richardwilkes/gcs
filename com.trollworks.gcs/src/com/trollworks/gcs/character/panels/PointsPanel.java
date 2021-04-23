/*
 * Copyright ©1998-2021 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character.panels;

import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.character.CharacterSetter;
import com.trollworks.gcs.page.DropPanel;
import com.trollworks.gcs.page.PageField;
import com.trollworks.gcs.page.PageLabel;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.text.Text;

import java.text.MessageFormat;
import javax.swing.SwingConstants;

/** The character points panel. */
public class PointsPanel extends DropPanel {
    /**
     * Creates a new points panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public PointsPanel(CharacterSheet sheet) {
        super(new PrecisionLayout().setColumns(2).setMargins(0).setSpacing(2, 0).setAlignment(PrecisionLayoutAlignment.FILL, PrecisionLayoutAlignment.FILL), getTitle(sheet.getCharacter()));
        GURPSCharacter gch = sheet.getCharacter();
        createLabelAndEditableField(gch.getUnspentPoints(), (c, v) -> c.setUnspentPoints(((Integer) v).intValue()), sheet, "unspent points", I18n.Text("Unspent"), I18n.Text("Points earned but not yet spent"));
        createLabelAndField(gch.getRacePoints(), sheet, I18n.Text("Race"), I18n.Text("Total points spent on a racial package"));
        createLabelAndField(gch.getAttributePoints(), sheet, I18n.Text("Attributes"), I18n.Text("Total points spent on attributes"));
        createLabelAndField(gch.getAdvantagePoints(), sheet, I18n.Text("Advantages"), I18n.Text("Total points spent on advantages"));
        createLabelAndField(gch.getDisadvantagePoints(), sheet, I18n.Text("Disadvantages"), I18n.Text("Total points spent on disadvantages"));
        createLabelAndField(gch.getQuirkPoints(), sheet, I18n.Text("Quirks"), I18n.Text("Total points spent on quirks"));
        createLabelAndField(gch.getSkillPoints(), sheet, I18n.Text("Skills"), I18n.Text("Total points spent on skills"));
        createLabelAndField(gch.getSpellPoints(), sheet, I18n.Text("Spells"), I18n.Text("Total points spent on spells"));
    }

    private void createLabelAndEditableField(int value, CharacterSetter setter, CharacterSheet sheet, String key, String title, String tooltip) {
        PageField field = new PageField(FieldFactory.INT6, Integer.valueOf(value), setter, sheet, key, SwingConstants.RIGHT, true, Text.wrapPlainTextForToolTip(tooltip), ThemeColor.ON_PAGE);
        add(field, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        add(new PageLabel(title, field));
    }

    private void createLabelAndField(int value, CharacterSheet sheet, String title, String tooltip) {
        PageField field = new PageField(FieldFactory.INT6, Integer.valueOf(value), sheet, SwingConstants.RIGHT, Text.wrapPlainTextForToolTip(tooltip), ThemeColor.ON_PAGE);
        add(field, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        add(new PageLabel(title, field));
    }

    private static String getTitle(GURPSCharacter gch) {
        return MessageFormat.format(I18n.Text("{0} Points"), Numbers.format(Preferences.getInstance().includeUnspentPointsInTotal() ? gch.getTotalPoints() : gch.getSpentPoints()));
    }
}
