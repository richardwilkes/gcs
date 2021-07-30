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
import com.trollworks.gcs.character.Profile;
import com.trollworks.gcs.page.DropPanel;
import com.trollworks.gcs.page.PageField;
import com.trollworks.gcs.page.PageLabel;
import com.trollworks.gcs.settings.GeneralSettingsWindow;
import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.FontAwesome;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.FontIconButton;
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
    private PageField mAgeField;
    private PageField mBirthdayField;
    private PageField mHeightField;
    private PageField mWeightField;
    private PageField mHairField;
    private PageField mEyeColorField;
    private PageField mSkinColorField;

    /**
     * Creates a new description panel.
     *
     * @param sheet The sheet to display the data for.
     */
    public DescriptionPanel(CharacterSheet sheet) {
        super(new PrecisionLayout().setColumns(5).setMargins(0).setSpacing(2, 0), I18n.text("Description"));
        GURPSCharacter gch     = sheet.getCharacter();
        Profile        profile = gch.getProfile();
        Wrapper        wrapper = new Wrapper(new PrecisionLayout().setColumns(3).setMargins(0).setSpacing(0, 0));
        createField(wrapper, sheet, FieldFactory.STRING, profile.getGender(), "gender",
                I18n.text("Gender"), null, (c, v) -> c.getProfile().setGender((String) v));
        mAgeField = createRandomizableField(wrapper, sheet, FieldFactory.STRING, profile.getAge(),
                "age", I18n.text("Age"), I18n.text("The character's age"),
                (c, v) -> c.getProfile().setAge((String) v), (b) -> {
                    mAgeField.attemptCommit();
                    mAgeField.requestFocus();
                    String current = profile.getAge();
                    String result;
                    do {
                        result = Numbers.format(profile.getRandomAge());
                    } while (result.equals(current));
                    profile.setAge(result);
                });
        mBirthdayField = createRandomizableField(wrapper, sheet, FieldFactory.STRING,
                profile.getBirthday(), "birthday", I18n.text("Birthday"),
                I18n.text("The character's birthday"),
                (c, v) -> c.getProfile().setBirthday((String) v), (b) -> {
                    mBirthdayField.attemptCommit();
                    mBirthdayField.requestFocus();
                    String current = profile.getBirthday();
                    String result;
                    do {
                        result = Profile.getRandomMonthAndDay();
                    } while (result.equals(current));
                    profile.setBirthday(result);
                });
        createField(wrapper, sheet, FieldFactory.STRING, profile.getReligion(), "religion",
                I18n.text("Religion"), null, (c, v) -> c.getProfile().setReligion((String) v));
        add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        createDivider();

        wrapper = new Wrapper(new PrecisionLayout().setColumns(3).setMargins(0).setSpacing(0, 0));
        mHeightField = createRandomizableField(wrapper, sheet, FieldFactory.HEIGHT,
                profile.getHeight(), "character height", I18n.text("Height"),
                I18n.text("The character's height"),
                (c, v) -> c.getProfile().setHeight((LengthValue) v), (b) -> {
                    mHeightField.attemptCommit();
                    mHeightField.requestFocus();
                    LengthValue length = profile.getHeight();
                    LengthValue result;
                    do {
                        result = profile.getRandomHeight(gch.getAttributeIntValue("st"),
                                profile.getSizeModifier());
                    } while (result.equals(length));
                    profile.setHeight(result);
                });
        mWeightField = createRandomizableField(wrapper, sheet, FieldFactory.WEIGHT,
                profile.getWeight(), "character weight", I18n.text("Weight"),
                I18n.text("The character's weight"),
                (c, v) -> c.getProfile().setWeight((WeightValue) v), (b) -> {
                    mWeightField.attemptCommit();
                    mWeightField.requestFocus();
                    WeightValue weight = profile.getWeight();
                    WeightValue result;
                    do {
                        result = profile.getRandomWeight(gch.getAttributeIntValue("st"), profile.getSizeModifier(), profile.getWeightMultiplier());
                    } while (result.equals(weight));
                    profile.setWeight(result);
                });
        createField(wrapper, sheet, FieldFactory.SM, Integer.valueOf(profile.getSizeModifier()),
                "SM", I18n.text("Size"), I18n.text("The character's size modifier"),
                (c, v) -> c.getProfile().setSizeModifier(((Integer) v).intValue()));
        createField(wrapper, sheet, FieldFactory.STRING, profile.getTechLevel(), "character TL",
                I18n.text("TL"), GeneralSettingsWindow.getTechLevelTooltip(),
                (c, v) -> c.getProfile().setTechLevel((String) v));
        add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        createDivider();

        wrapper = new Wrapper(new PrecisionLayout().setColumns(3).setMargins(0).setSpacing(0, 0));
        mHairField = createRandomizableField(wrapper, sheet, FieldFactory.STRING, profile.getHair(),
                "hair", I18n.text("Hair"), I18n.text("The character's hair style and color"),
                (c, v) -> c.getProfile().setHair((String) v), (b) -> {
                    mHairField.attemptCommit();
                    mHairField.requestFocus();
                    profile.setHair(Profile.getRandomHair(profile.getHair()));
                });
        mEyeColorField = createRandomizableField(wrapper, sheet, FieldFactory.STRING,
                profile.getEyeColor(), "eye color", I18n.text("Eyes"),
                I18n.text("The character's eye color"),
                (c, v) -> c.getProfile().setEyeColor((String) v), (b) -> {
                    mEyeColorField.attemptCommit();
                    mEyeColorField.requestFocus();
                    profile.setEyeColor(Profile.getRandomEyeColor(profile.getEyeColor()));
                });
        mSkinColorField = createRandomizableField(wrapper, sheet, FieldFactory.STRING,
                profile.getSkinColor(), "skin color", I18n.text("Skin"),
                I18n.text("The character's skin color"),
                (c, v) -> c.getProfile().setSkinColor((String) v), (b) -> {
                    mSkinColorField.attemptCommit();
                    mSkinColorField.requestFocus();
                    profile.setSkinColor(Profile.getRandomSkinColor(profile.getSkinColor()));
                });
        createField(wrapper, sheet, FieldFactory.STRING, profile.getHandedness(), "handedness", I18n.text("Hand"), I18n.text("The character's preferred hand"), (c, v) -> c.getProfile().setHandedness((String) v));
        add(wrapper, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }

    private static PageField createRandomizableField(Container parent, CharacterSheet sheet, AbstractFormatterFactory factory, Object value, String tag, String title, String tooltip, CharacterSetter setter, FontIconButton.ClickFunction randomizer) {
        FontIconButton button = new FontIconButton(FontAwesome.RANDOM, String.format(I18n.text("Randomize %s"), title), randomizer);
        button.setThemeFont(Fonts.FONT_ICON_PAGE_SMALL);
        button.setFocusable(false);
        parent.add(button);
        PageField field = new PageField(factory, value, setter, sheet, tag, SwingConstants.LEFT, true, tooltip);
        parent.add(new PageLabel(title), new PrecisionLayoutData().setEndHorizontalAlignment().setLeftMargin(1));
        parent.add(field, createFieldLayout());
        return field;
    }

    private static void createField(Container parent, CharacterSheet sheet, AbstractFormatterFactory factory, Object value, String tag, String title, String tooltip, CharacterSetter setter) {
        PageField field = new PageField(factory, value, setter, sheet, tag, SwingConstants.LEFT, true, tooltip);
        parent.add(new PageLabel(title), new PrecisionLayoutData().setEndHorizontalAlignment().setHorizontalSpan(2));
        parent.add(field, createFieldLayout());
    }

    private static PrecisionLayoutData createFieldLayout() {
        return new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setLeftMargin(4);
    }

    private void createDivider() {
        Wrapper panel = new Wrapper();
        panel.setOnlySize(1, 1);
        add(panel);
        addVerticalBackground(panel, Colors.DIVIDER);
    }
}
