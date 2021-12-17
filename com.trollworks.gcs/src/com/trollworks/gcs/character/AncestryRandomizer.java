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

package com.trollworks.gcs.character;

import com.trollworks.gcs.menu.edit.RandomizeForAncestryCommand;
import com.trollworks.gcs.settings.SheetSettings;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.Checkbox;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.ui.widget.Modal;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.units.LengthValue;
import com.trollworks.gcs.utility.units.WeightValue;

import java.awt.BorderLayout;
import java.awt.Container;

public class AncestryRandomizer implements Runnable {
    private CharacterSheet mSheet;
    private boolean        mUserInitiated;

    public AncestryRandomizer(CharacterSheet sheet, boolean userInitiated) {
        mSheet = sheet;
        mUserInitiated = userInitiated;
    }

    @Override
    public void run() {
        Modal          dialog  = new Modal(mSheet, mUserInitiated ? RandomizeForAncestryCommand.INSTANCE.getTitle() : I18n.text("Ancestry Changed"));
        Container      content = dialog.getContentPane();
        Panel          panel   = new Panel(new PrecisionLayout().setMargins(20).setColumns(3));
        GURPSCharacter gchar   = mSheet.getCharacter();
        Profile        profile = gchar.getProfile();
        if (!mUserInitiated) {
            panel.add(new Label(String.format(I18n.text("The ancestry of %s has changed to %s."), profile.getName(), gchar.getAncestryRef().name())), new PrecisionLayoutData().setHorizontalSpan(3));
        }
        panel.add(new Label(I18n.text("Check the fields you'd like to randomize for the ancestry:")), new PrecisionLayoutData().setTopMargin(20).setBottomMargin(20).setHorizontalSpan(3));
        Checkbox nameCheckbox = new Checkbox(I18n.text("Name"), true, null);
        panel.add(nameCheckbox, new PrecisionLayoutData().setLeftMargin(20));
        Checkbox heightCheckbox = new Checkbox(I18n.text("Height"), true, null);
        panel.add(heightCheckbox, new PrecisionLayoutData().setLeftMargin(20));
        Checkbox hairCheckbox = new Checkbox(I18n.text("Hair"), true, null);
        panel.add(hairCheckbox, new PrecisionLayoutData().setLeftMargin(20));
        Checkbox genderCheckbox = new Checkbox(I18n.text("Gender"), true, null);
        panel.add(genderCheckbox, new PrecisionLayoutData().setLeftMargin(20));
        Checkbox weightCheckbox = new Checkbox(I18n.text("Weight"), true, null);
        panel.add(weightCheckbox, new PrecisionLayoutData().setLeftMargin(20));
        Checkbox eyesCheckbox = new Checkbox(I18n.text("Eyes"), true, null);
        panel.add(eyesCheckbox, new PrecisionLayoutData().setLeftMargin(20));
        Checkbox ageCheckbox = new Checkbox(I18n.text("Age"), true, null);
        panel.add(ageCheckbox, new PrecisionLayoutData().setLeftMargin(20));
        Checkbox handCheckbox = new Checkbox(I18n.text("Hand"), true, null);
        panel.add(handCheckbox, new PrecisionLayoutData().setLeftMargin(20));
        Checkbox skinCheckbox = new Checkbox(I18n.text("Skin"), true, null);
        panel.add(skinCheckbox, new PrecisionLayoutData().setLeftMargin(20));
        Checkbox birthdayCheckbox = new Checkbox(I18n.text("Birthday"), true, null);
        panel.add(birthdayCheckbox, new PrecisionLayoutData().setLeftMargin(20));
        content.add(panel, BorderLayout.CENTER);
        dialog.addCancelButton();
        dialog.addApplyButton();
        dialog.presentToUser();
        if (dialog.getResult() == Modal.OK) {
            if (nameCheckbox.isChecked()) {
                profile.setGender(profile.getRandomGender(""));
            }
            if (ageCheckbox.isChecked()) {
                profile.setAge(Numbers.format(profile.getRandomAge(0)));
            }
            if (birthdayCheckbox.isChecked()) {
                profile.setBirthday(profile.getRandomBirthday(""));
            }
            SheetSettings sheetSettings = gchar.getSheetSettings();
            if (heightCheckbox.isChecked()) {
                profile.setHeight(profile.getRandomHeight(new LengthValue(Fixed6.ZERO, sheetSettings.defaultLengthUnits())));
            }
            if (weightCheckbox.isChecked()) {
                profile.setWeight(profile.getRandomWeight(profile.getWeightMultiplier(), new WeightValue(Fixed6.ZERO, sheetSettings.defaultWeightUnits())));
            }
            if (hairCheckbox.isChecked()) {
                profile.setHair(profile.getRandomHair(""));
            }
            if (eyesCheckbox.isChecked()) {
                profile.setEyeColor(profile.getRandomEyeColor(""));
            }
            if (skinCheckbox.isChecked()) {
                profile.setSkinColor(profile.getRandomSkin(""));
            }
            if (handCheckbox.isChecked()) {
                profile.setHandedness(profile.getRandomHandedness(""));
            }
            if (nameCheckbox.isChecked()) {
                profile.setName(profile.getRandomName(""));
            }
        }
        if (!mUserInitiated) {
            mSheet.setAncestryChangePending(false);
        }
    }
}
