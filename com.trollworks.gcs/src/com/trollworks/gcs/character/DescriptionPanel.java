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
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.IconButton;
import com.trollworks.gcs.ui.widget.Wrapper;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.units.LengthValue;
import com.trollworks.gcs.utility.units.WeightValue;

import java.awt.Container;
import javax.swing.SwingConstants;

/** The character description panel. */
public class DescriptionPanel extends DropPanel {
    private GURPSCharacter mCharacter;

    /**
     * Creates a new description panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public DescriptionPanel(CharacterSheet sheet) {
        super(new PrecisionLayout().setColumns(5).setMargins(0).setSpacing(2, 0), I18n.Text("Description"));
        mCharacter = sheet.getCharacter();

        Wrapper wrapper = new Wrapper(new PrecisionLayout().setColumns(3).setMargins(0).setSpacing(0, 0));
        createLabelAndField2(wrapper, sheet, Profile.ID_GENDER, I18n.Text("Gender:"), null);
        createLabelAndRandomizableField(wrapper, sheet, Profile.ID_AGE, I18n.Text("Age:"), I18n.Text("The character's age"), this::getRandomAge);
        createLabelAndRandomizableField(wrapper, sheet, Profile.ID_BIRTHDAY, I18n.Text("Birthday:"), I18n.Text("The character's birthday"), this::getRandomBirthday);
        createLabelAndField2(wrapper, sheet, Profile.ID_RELIGION, I18n.Text("Religion:"), null);
        add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        createDivider();

        wrapper = new Wrapper(new PrecisionLayout().setColumns(3).setMargins(0).setSpacing(0, 0));
        createLabelAndRandomizableField(wrapper, sheet, Profile.ID_HEIGHT, I18n.Text("Height:"), I18n.Text("The character's height"), this::getRandomHeight);
        createLabelAndRandomizableField(wrapper, sheet, Profile.ID_WEIGHT, I18n.Text("Weight:"), I18n.Text("The character's weight"), this::getRandomWeight);
        createLabelAndField2(wrapper, sheet, Profile.ID_SIZE_MODIFIER, I18n.Text("Size:"), I18n.Text("The character's size modifier"));
        createLabelAndField2(wrapper, sheet, Profile.ID_TECH_LEVEL, I18n.Text("TL:"), I18n.Text("<html><body>TL0: Stone Age<br>TL1: Bronze Age<br>TL2: Iron Age<br>TL3: Medieval<br>TL4: Age of Sail<br>TL5: Industrial Revolution<br>TL6: Mechanized Age<br>TL7: Nuclear Age<br>TL8: Digital Age<br>TL9: Microtech Age<br>TL10: Robotic Age<br>TL11: Age of Exotic Matter<br>TL12: Anything Goes</body></html>"));
        add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        createDivider();

        wrapper = new Wrapper(new PrecisionLayout().setColumns(3).setMargins(0).setSpacing(0, 0));
        createLabelAndRandomizableField(wrapper, sheet, Profile.ID_HAIR, I18n.Text("Hair:"), I18n.Text("The character's hair style and color"), this::getRandomHair);
        createLabelAndRandomizableField(wrapper, sheet, Profile.ID_EYE_COLOR, I18n.Text("Eyes:"), I18n.Text("The character's eye color"), this::getRandomEyeColor);
        createLabelAndRandomizableField(wrapper, sheet, Profile.ID_SKIN_COLOR, I18n.Text("Skin:"), I18n.Text("The character's skin color"), this::getRandomSkinColor);
        createLabelAndField2(wrapper, sheet, Profile.ID_HANDEDNESS, I18n.Text("Hand:"), I18n.Text("The character's preferred hand"));
        add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }

    private void createLabelAndField2(Container parent, CharacterSheet sheet, String key, String title, String tooltip) {
        PageField field = new PageField(sheet, key, SwingConstants.LEFT, true, tooltip);
        parent.add(new PageLabel(title, field), new PrecisionLayoutData().setEndHorizontalAlignment().setHorizontalSpan(2));
        parent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setLeftMargin(4));
    }

    private void createLabelAndRandomizableField(Container parent, CharacterSheet sheet, String key, String title, String tooltip, FieldRandomizer randomizer) {
        PageField field = new PageField(sheet, key, SwingConstants.LEFT, true, null);
        parent.add(new IconButton(Images.RANDOMIZE, String.format(I18n.Text("Randomize %s"), tooltip.toLowerCase()), () -> mCharacter.setValueForID(key, randomizer.Randomize())));
        parent.add(new PageLabel(title, field), new PrecisionLayoutData().setEndHorizontalAlignment());
        parent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setLeftMargin(4));
    }

    private void createDivider() {
        Wrapper panel = new Wrapper();
        panel.setOnlySize(1, 1);
        add(panel);
        addVerticalBackground(panel, ThemeColor.ON_PAGE);
    }

    private String getRandomAge() {
        Profile profile = mCharacter.getProfile();
        String  current = profile.getAge();
        String  result;
        do {
            result = Numbers.format(profile.getRandomAge());
        } while (result.equals(current));
        return result;
    }

    private String getRandomBirthday() {
        String current = mCharacter.getProfile().getBirthday();
        String result;
        do {
            result = Profile.getRandomMonthAndDay();
        } while (result.equals(current));
        return result;
    }

    private LengthValue getRandomHeight() {
        Profile     profile = mCharacter.getProfile();
        LengthValue length  = profile.getHeight();
        LengthValue result;
        do {
            result = profile.getRandomHeight(mCharacter.getStrength(), profile.getSizeModifier());
        } while (result.equals(length));
        return result;
    }

    private WeightValue getRandomWeight() {
        Profile     profile = mCharacter.getProfile();
        WeightValue weight  = profile.getWeight();
        WeightValue result;
        do {
            result = profile.getRandomWeight(mCharacter.getStrength(), profile.getSizeModifier(), profile.getWeightMultiplier());
        } while (result.equals(weight));
        return result;
    }

    private String getRandomHair() {
        return Profile.getRandomHair(mCharacter.getProfile().getHair());
    }

    private String getRandomEyeColor() {
        return Profile.getRandomEyeColor(mCharacter.getProfile().getEyeColor());
    }

    private String getRandomSkinColor() {
        return Profile.getRandomSkinColor(mCharacter.getProfile().getSkinColor());
    }
}
