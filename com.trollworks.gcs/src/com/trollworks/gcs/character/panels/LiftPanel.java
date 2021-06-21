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
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.page.DropPanel;
import com.trollworks.gcs.page.PageLabel;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.units.WeightValue;

/** The character lift panel. */
public class LiftPanel extends DropPanel {
    /**
     * Creates a new lift panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public LiftPanel(CharacterSheet sheet) {
        super(new PrecisionLayout().setColumns(2).setMargins(0).setSpacing(2, 0), I18n.text("Lifting & Moving Things"));
        GURPSCharacter gch = sheet.getCharacter();
        createRow(sheet, gch.getBasicLift(), I18n.text("Basic Lift"), I18n.text("The weight the character can lift overhead with one hand in one second"));
        createRow(sheet, gch.getOneHandedLift(), I18n.text("One-Handed Lift"), I18n.text("The weight the character can lift overhead with one hand in two seconds"));
        createRow(sheet, gch.getTwoHandedLift(), I18n.text("Two-Handed Lift"), I18n.text("The weight the character can lift overhead with both hands in four seconds"));
        createRow(sheet, gch.getShoveAndKnockOver(), I18n.text("Shove & Knock Over"), I18n.text("The weight of an object the character can shove and knock over"));
        createRow(sheet, gch.getRunningShoveAndKnockOver(), I18n.text("Running Shove & Knock Over"), I18n.text("The weight of an object the character can shove  and knock over with a running start"));
        createRow(sheet, gch.getCarryOnBack(), I18n.text("Carry On Back"), I18n.text("The weight the character can carry slung across the back"));
        createRow(sheet, gch.getShiftSlightly(), I18n.text("Shift Slightly"), I18n.text("The weight of an object the character can shift slightly on a floor"));
    }

    private void createRow(CharacterSheet sheet, WeightValue weight, String title, String tooltip) {
        add(new PageLabel(weight.toString(), tooltip), new PrecisionLayoutData().setHorizontalAlignment(PrecisionLayoutAlignment.END).setGrabHorizontalSpace(true));
        add(new PageLabel(title, tooltip), new PrecisionLayoutData().setGrabHorizontalSpace(true));
    }
}
