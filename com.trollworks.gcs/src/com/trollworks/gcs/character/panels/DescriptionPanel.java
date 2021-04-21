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
import com.trollworks.gcs.character.Profile;
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
import javax.swing.JFormattedTextField.AbstractFormatterFactory;
import javax.swing.SwingConstants;

/** The character description panel. */
public class DescriptionPanel extends DropPanel {
    /**
     * Creates a new description panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public DescriptionPanel(CharacterSheet sheet) {
        super(new PrecisionLayout().setColumns(5).setMargins(0).setSpacing(2, 0), I18n.Text("Description"));
        GURPSCharacter gch     = sheet.getCharacter();
        Profile        profile = gch.getProfile();
        Wrapper        wrapper = new Wrapper(new PrecisionLayout().setColumns(3).setMargins(0).setSpacing(0, 0));
        createField(wrapper, sheet, FieldFactory.STRING, profile.getGender(), "gender", I18n.Text("Gender"), null, (c, v) -> c.getProfile().setGender((String) v));
        createRandomizableField(wrapper, sheet, FieldFactory.STRING, profile.getAge(), "age", I18n.Text("Age"), I18n.Text("The character's age"), (c, v) -> c.getProfile().setAge((String) v), () -> {
            String current = profile.getAge();
            String result;
            do {
                result = Numbers.format(profile.getRandomAge());
            } while (result.equals(current));
            profile.setAge(result);
        });
        createRandomizableField(wrapper, sheet, FieldFactory.STRING, profile.getBirthday(), "birthday", I18n.Text("Birthday"), I18n.Text("The character's birthday"), (c, v) -> c.getProfile().setBirthday((String) v), () -> {
            String current = profile.getBirthday();
            String result;
            do {
                result = Profile.getRandomMonthAndDay();
            } while (result.equals(current));
            profile.setBirthday(result);
        });
        createField(wrapper, sheet, FieldFactory.STRING, profile.getReligion(), "religion", I18n.Text("Religion"), null, (c, v) -> c.getProfile().setReligion((String) v));
        add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        createDivider();

        wrapper = new Wrapper(new PrecisionLayout().setColumns(3).setMargins(0).setSpacing(0, 0));
        createRandomizableField(wrapper, sheet, FieldFactory.HEIGHT, profile.getHeight(), "character height", I18n.Text("Height"), I18n.Text("The character's height"), (c, v) -> c.getProfile().setHeight((LengthValue) v), () -> {
            LengthValue length = profile.getHeight();
            LengthValue result;
            do {
                result = profile.getRandomHeight(gch.getAttributeIntValue("st"), profile.getSizeModifier());
            } while (result.equals(length));
            profile.setHeight(result);
        });
        createRandomizableField(wrapper, sheet, FieldFactory.WEIGHT, profile.getWeight(), "character weight", I18n.Text("Weight"), I18n.Text("The character's weight"), (c, v) -> c.getProfile().setWeight((WeightValue) v), () -> {
            WeightValue weight = profile.getWeight();
            WeightValue result;
            do {
                result = profile.getRandomWeight(gch.getAttributeIntValue("st"), profile.getSizeModifier(), profile.getWeightMultiplier());
            } while (result.equals(weight));
            profile.setWeight(result);
        });
        createField(wrapper, sheet, FieldFactory.SM, Integer.valueOf(profile.getSizeModifier()), "SM", I18n.Text("Size"), I18n.Text("The character's size modifier"), (c, v) -> c.getProfile().setSizeModifier(((Integer) v).intValue()));
        createField(wrapper, sheet, FieldFactory.STRING, profile.getTechLevel(), "character TL", I18n.Text("TL"), I18n.Text("""
                <html><body>
                TL0: Stone Age<br>
                TL1: Bronze Age<br>
                TL2: Iron Age<br>
                TL3: Medieval<br>
                TL4: Age of Sail<br>
                TL5: Industrial Revolution<br>
                TL6: Mechanized Age<br>
                TL7: Nuclear Age<br>
                TL8: Digital Age<br>
                TL9: Microtech Age<br>
                TL10: Robotic Age<br>
                TL11: Age of Exotic Matter<br>
                TL12: Anything Goes
                </body></html>
                """), (c, v) -> c.getProfile().setTechLevel((String) v));
        add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        createDivider();

        wrapper = new Wrapper(new PrecisionLayout().setColumns(3).setMargins(0).setSpacing(0, 0));
        createRandomizableField(wrapper, sheet, FieldFactory.STRING, profile.getHair(), "hair", I18n.Text("Hair"), I18n.Text("The character's hair style and color"), (c, v) -> c.getProfile().setHair((String) v), () -> profile.setHair(Profile.getRandomHair(profile.getHair())));
        createRandomizableField(wrapper, sheet, FieldFactory.STRING, profile.getEyeColor(), "eye color", I18n.Text("Eyes"), I18n.Text("The character's eye color"), (c, v) -> c.getProfile().setEyeColor((String) v), () -> profile.setEyeColor(Profile.getRandomEyeColor(profile.getEyeColor())));
        createRandomizableField(wrapper, sheet, FieldFactory.STRING, profile.getSkinColor(), "skin color", I18n.Text("Skin"), I18n.Text("The character's skin color"), (c, v) -> c.getProfile().setSkinColor((String) v), () -> profile.setSkinColor(Profile.getRandomSkinColor(profile.getSkinColor())));
        createField(wrapper, sheet, FieldFactory.STRING, profile.getHandedness(), "handedness", I18n.Text("Hand"), I18n.Text("The character's preferred hand"), (c, v) -> c.getProfile().setHandedness((String) v));
        add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }

    private void createRandomizableField(Container parent, CharacterSheet sheet, AbstractFormatterFactory factory, Object value, String tag, String title, String tooltip, CharacterSetter setter, Runnable randomizer) {
        IconButton button = new IconButton(Images.RANDOMIZE, null, randomizer);
        button.setToolTipText(String.format(I18n.Text("Randomize %s"), title));
        parent.add(button);
        PageField field = new PageField(factory, value, setter, sheet, tag, SwingConstants.LEFT, true, tooltip, ThemeColor.ON_PAGE);
        parent.add(new PageLabel(title + ":", field), new PrecisionLayoutData().setEndHorizontalAlignment());
        parent.add(field, createFieldLayout());
    }

    private void createField(Container parent, CharacterSheet sheet, AbstractFormatterFactory factory, Object value, String tag, String title, String tooltip, CharacterSetter setter) {
        PageField field = new PageField(factory, value, setter, sheet, tag, SwingConstants.LEFT, true, tooltip, ThemeColor.ON_PAGE);
        parent.add(new PageLabel(title + ":", field), new PrecisionLayoutData().setEndHorizontalAlignment().setHorizontalSpan(2));
        parent.add(field, createFieldLayout());
    }

    private PrecisionLayoutData createFieldLayout() {
        return new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setLeftMargin(4);
    }

    private void createDivider() {
        Wrapper panel = new Wrapper();
        panel.setOnlySize(1, 1);
        add(panel);
        addVerticalBackground(panel, ThemeColor.ON_PAGE);
    }
}
