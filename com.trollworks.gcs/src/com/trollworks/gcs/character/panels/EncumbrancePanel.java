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
import com.trollworks.gcs.character.Encumbrance;
import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.page.DropPanel;
import com.trollworks.gcs.page.PageField;
import com.trollworks.gcs.page.PageHeader;
import com.trollworks.gcs.page.PageLabel;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.ui.widget.Wrapper;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;

import java.awt.Color;
import java.awt.Container;
import java.awt.Dimension;
import java.text.MessageFormat;
import javax.swing.JComponent;
import javax.swing.SwingConstants;

/** The character encumbrance panel. */
public class EncumbrancePanel extends DropPanel {
    /**
     * Creates a new encumbrance panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public EncumbrancePanel(CharacterSheet sheet) {
        super(new ColumnLayout(8, 2, 0), I18n.Text("Encumbrance, Move & Dodge"), true);
        Encumbrance[] encumbranceValues = Encumbrance.values();
        Dimension     prefSize          = new PageLabel("•", null).getPreferredSize();
        PageLabel[]   markers           = new PageLabel[encumbranceValues.length];
        String        encLevelTooltip   = I18n.Text("The encumbrance level");
        PageHeader    bulletHeader      = createHeader("", encLevelTooltip);
        UIUtilities.setOnlySize(bulletHeader, prefSize);
        addHorizontalBackground(bulletHeader, ThemeColor.HEADER);
        PageHeader header = createHeader(I18n.Text("Level"), encLevelTooltip);
        addVerticalBackground(createDivider(), ThemeColor.DIVIDER);
        String maxLoadTooltip = I18n.Text("The maximum load a character can carry and still remain within a specific encumbrance level");
        createHeader(I18n.Text("Max Load"), maxLoadTooltip);
        addVerticalBackground(createDivider(), ThemeColor.DIVIDER);
        String moveTooltip = I18n.Text("The character's ground movement rate for a specific encumbrance level");
        createHeader(I18n.Text("Move"), moveTooltip);
        addVerticalBackground(createDivider(), ThemeColor.DIVIDER);
        String dodgeTooltip = I18n.Text("The character's dodge for a specific encumbrance level");
        createHeader(I18n.Text("Dodge"), dodgeTooltip);
        GURPSCharacter character = sheet.getCharacter();
        Encumbrance    current   = character.getEncumbranceLevel(false);
        boolean        band      = false;
        for (Encumbrance encumbrance : encumbranceValues) {
            boolean warn;
            Color   textColor;
            if (current == encumbrance) {
                warn = character.isCarryingGreaterThanMaxLoad(false);
                textColor = warn ? ThemeColor.ON_WARN : ThemeColor.ON_CURRENT;
            } else {
                warn = false;
                textColor = ThemeColor.ON_CONTENT;
            }
            int index = encumbrance.ordinal();
            markers[index] = new PageLabel(encumbrance == current ? "•" : "", textColor, header);
            UIUtilities.setOnlySize(markers[index], prefSize);
            add(markers[index]);
            if (current == encumbrance) {
                addHorizontalBackground(markers[index], warn ? ThemeColor.WARN : ThemeColor.CURRENT);
            } else if (band) {
                addHorizontalBackground(markers[index], ThemeColor.BANDING);
            }
            band = !band;
            JComponent field = new PageLabel(MessageFormat.format("{0} {1}", Numbers.format(-encumbrance.getEncumbrancePenalty()), encumbrance), textColor, header);
            add(field);
            createDivider();
            add(new PageField(FieldFactory.WEIGHT, character.getMaximumCarry(encumbrance), sheet, SwingConstants.RIGHT, maxLoadTooltip, textColor));
            createDivider();
            add(new PageField(FieldFactory.POSINT5, Integer.valueOf(character.getMove(encumbrance)), sheet, SwingConstants.RIGHT, moveTooltip, textColor));
            createDivider();
            add(new PageField(FieldFactory.POSINT5, Integer.valueOf(character.getDodge(encumbrance)), sheet, SwingConstants.RIGHT, dodgeTooltip, textColor));
        }
    }

    private PageHeader createHeader(String title, String tooltip) {
        PageHeader header = new PageHeader(title, tooltip);
        add(header);
        return header;
    }

    private Container createDivider() {
        Wrapper panel = new Wrapper();
        panel.setOnlySize(1, 1);
        add(panel);
        return panel;
    }
}
