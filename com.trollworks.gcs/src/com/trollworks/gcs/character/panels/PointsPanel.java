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

import com.trollworks.gcs.character.CharacterSetter;
import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.page.DropPanel;
import com.trollworks.gcs.page.PageField;
import com.trollworks.gcs.page.PageLabel;
import com.trollworks.gcs.settings.Settings;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;

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
        createLabelAndEditableField(gch.getUnspentPoints(), (c, v) -> c.setUnspentPoints(((Integer) v).intValue()), sheet, "unspent points", I18n.text("Unspent"), I18n.text("Points earned but not yet spent"));
        createLabelAndField(gch.getRacePoints(), sheet, I18n.text("Race"), I18n.text("Total points spent on a racial package"));
        createLabelAndField(gch.getAttributePoints(), sheet, I18n.text("Attributes"), I18n.text("Total points spent on attributes"));
        createLabelAndField(gch.getAdvantagePoints(), sheet, I18n.text("Advantages"), I18n.text("Total points spent on advantages"));
        createLabelAndField(gch.getDisadvantagePoints(), sheet, I18n.text("Disadvantages"), I18n.text("Total points spent on disadvantages"));
        createLabelAndField(gch.getQuirkPoints(), sheet, I18n.text("Quirks"), I18n.text("Total points spent on quirks"));
        createLabelAndField(gch.getSkillPoints(), sheet, I18n.text("Skills"), I18n.text("Total points spent on skills"));
        createLabelAndField(gch.getSpellPoints(), sheet, I18n.text("Spells"), I18n.text("Total points spent on spells"));
    }

    private void createLabelAndEditableField(int value, CharacterSetter setter, CharacterSheet sheet, String key, String title, String tooltip) {
        PageField field = new PageField(FieldFactory.INT6, Integer.valueOf(value), setter, sheet, key, SwingConstants.RIGHT, true, tooltip);
        add(field, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        add(new PageLabel(title));
    }

    private void createLabelAndField(int value, CharacterSheet sheet, String title, String tooltip) {
        PageLabel pts = new PageLabel(Numbers.format(value), tooltip);
        pts.setHorizontalAlignment(SwingConstants.RIGHT);
        pts.setBorder(new EmptyBorder(0, 2, 0, 2));
        add(pts, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
        add(new PageLabel(title, tooltip));
    }

    private static String getTitle(GURPSCharacter gch) {
        return MessageFormat.format(I18n.text("{0} Points"), Numbers.format(Settings.getInstance().getGeneralSettings().includeUnspentPointsInTotal() ? gch.getTotalPoints() : gch.getSpentPoints()));
    }
}
