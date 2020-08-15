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
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.utility.I18n;

import javax.swing.SwingConstants;

/** The character damage panel. */
public class LiftPanel extends DropPanel {
    /**
     * Creates a new damage panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public LiftPanel(CharacterSheet sheet) {
        super(new ColumnLayout(2, 2, 0), I18n.Text("Lifting & Moving Things"));
        createRow(sheet, GURPSCharacter.ID_BASIC_LIFT, I18n.Text("Basic Lift"), I18n.Text("The weight the character can lift overhead with one hand in one second"));
        createRow(sheet, GURPSCharacter.ID_ONE_HANDED_LIFT, I18n.Text("One-Handed Lift"), I18n.Text("The weight the character can lift overhead with one hand in two seconds"));
        createRow(sheet, GURPSCharacter.ID_TWO_HANDED_LIFT, I18n.Text("Two-Handed Lift"), I18n.Text("The weight the character can lift overhead with both hands in four seconds"));
        createRow(sheet, GURPSCharacter.ID_SHOVE_AND_KNOCK_OVER, I18n.Text("Shove & Knock Over"), I18n.Text("The weight of an object the character can shove and knock over"));
        createRow(sheet, GURPSCharacter.ID_RUNNING_SHOVE_AND_KNOCK_OVER, I18n.Text("Running Shove & Knock Over"), I18n.Text("The weight of an object the character can shove  and knock over with a running start"));
        createRow(sheet, GURPSCharacter.ID_CARRY_ON_BACK, I18n.Text("Carry On Back"), I18n.Text("The weight the character can carry slung across the back"));
        createRow(sheet, GURPSCharacter.ID_SHIFT_SLIGHTLY, I18n.Text("Shift Slightly"), I18n.Text("The weight of an object the character can shift slightly on a floor"));
    }

    private void createRow(CharacterSheet sheet, String key, String title, String tooltip) {
        PageField field = new PageField(sheet, key, SwingConstants.RIGHT, false, tooltip);
        field.setForeground(ThemeColor.ON_PAGE);
        field.setDisabledTextColor(ThemeColor.ON_PAGE);
        add(field);
        add(new PageLabel(title, field));
    }
}
