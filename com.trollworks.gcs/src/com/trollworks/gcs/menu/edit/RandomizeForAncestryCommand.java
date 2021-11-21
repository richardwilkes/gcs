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

package com.trollworks.gcs.menu.edit;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.character.Profile;
import com.trollworks.gcs.character.SheetDockable;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.settings.SheetSettings;
import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.units.LengthValue;
import com.trollworks.gcs.utility.units.WeightValue;

import java.awt.event.ActionEvent;

public final class RandomizeForAncestryCommand extends Command {
    /** The singleton {@link RandomizeForAncestryCommand}. */
    public static final RandomizeForAncestryCommand INSTANCE = new RandomizeForAncestryCommand();

    private RandomizeForAncestryCommand() {
        super(I18n.text("Randomize for Ancestry"), "randomize_for_ancestry");
    }

    @Override
    public void adjust() {
        setEnabled(getTarget(SheetDockable.class) != null);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        SheetDockable target = getTarget(SheetDockable.class);
        if (target != null) {
            GURPSCharacter gchar   = target.getDataFile();
            Profile        profile = gchar.getProfile();
            profile.setGender(profile.getRandomGender(""));
            profile.setAge(Numbers.format(profile.getRandomAge(0)));
            profile.setBirthday(profile.getRandomBirthday(""));
            SheetSettings sheetSettings = gchar.getSheetSettings();
            profile.setHeight(profile.getRandomHeight(new LengthValue(Fixed6.ZERO, sheetSettings.defaultLengthUnits())));
            profile.setWeight(profile.getRandomWeight(profile.getWeightMultiplier(), new WeightValue(Fixed6.ZERO, sheetSettings.defaultWeightUnits())));
            profile.setHair(profile.getRandomHair(""));
            profile.setEyeColor(profile.getRandomEyeColor(""));
            profile.setSkinColor(profile.getRandomSkin(""));
            profile.setHandedness(profile.getRandomHandedness(""));
            profile.setName(profile.getRandomName(""));
        }
    }
}
