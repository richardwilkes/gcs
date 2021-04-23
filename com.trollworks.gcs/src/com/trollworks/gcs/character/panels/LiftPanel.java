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
import com.trollworks.gcs.page.DropPanel;
import com.trollworks.gcs.page.PageField;
import com.trollworks.gcs.page.PageLabel;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.units.WeightValue;

import javax.swing.SwingConstants;

/** The character lift panel. */
public class LiftPanel extends DropPanel {
    /**
     * Creates a new lift panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public LiftPanel(CharacterSheet sheet) {
        super(new ColumnLayout(2, 2, 0), I18n.Text("Lifting & Moving Things"));
        GURPSCharacter gch = sheet.getCharacter();
        createRow(sheet, gch.getBasicLift(), I18n.Text("Basic Lift"), I18n.Text("The weight the character can lift overhead with one hand in one second"));
        createRow(sheet, gch.getOneHandedLift(), I18n.Text("One-Handed Lift"), I18n.Text("The weight the character can lift overhead with one hand in two seconds"));
        createRow(sheet, gch.getTwoHandedLift(), I18n.Text("Two-Handed Lift"), I18n.Text("The weight the character can lift overhead with both hands in four seconds"));
        createRow(sheet, gch.getShoveAndKnockOver(), I18n.Text("Shove & Knock Over"), I18n.Text("The weight of an object the character can shove and knock over"));
        createRow(sheet, gch.getRunningShoveAndKnockOver(), I18n.Text("Running Shove & Knock Over"), I18n.Text("The weight of an object the character can shove  and knock over with a running start"));
        createRow(sheet, gch.getCarryOnBack(), I18n.Text("Carry On Back"), I18n.Text("The weight the character can carry slung across the back"));
        createRow(sheet, gch.getShiftSlightly(), I18n.Text("Shift Slightly"), I18n.Text("The weight of an object the character can shift slightly on a floor"));
    }

    private void createRow(CharacterSheet sheet, WeightValue weight, String title, String tooltip) {
        PageField field = new PageField(FieldFactory.WEIGHT, weight, sheet, SwingConstants.RIGHT, tooltip, ThemeColor.ON_PAGE);
        add(field);
        add(new PageLabel(title, field));
    }
}
