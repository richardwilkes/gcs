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

import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.text.Text;
import com.trollworks.gcs.utility.undo.MultipleUndo;
import com.trollworks.gcs.utility.units.LengthValue;
import com.trollworks.gcs.utility.units.WeightValue;

import java.awt.BorderLayout;
import java.awt.Container;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import javax.swing.JButton;
import javax.swing.JCheckBox;
import javax.swing.JPanel;
import javax.swing.JTextField;

/** A character description randomizer. */
public class DescriptionRandomizer extends JPanel implements ActionListener {
    private static final int            GENDER_INDEX   = 0;
    private static final int            AGE_INDEX      = 1;
    private static final int            BIRTHDAY_INDEX = 2;
    private static final int            HEIGHT_INDEX   = 3;
    private static final int            WEIGHT_INDEX   = 4;
    private static final int            HAIR_INDEX     = 5;
    private static final int            EYES_INDEX     = 6;
    private static final int            SKIN_INDEX     = 7;
    private static final int            HAND_INDEX     = 8;
    private static final int            COUNT          = 9;
    private              GURPSCharacter mCharacter;
    private              JCheckBox[]    mCheckBoxes;
    private              JTextField[]   mFields;

    /**
     * Creates a new {@link DescriptionRandomizer}.
     *
     * @param character The {@link GURPSCharacter} to randomize the description of.
     */
    public DescriptionRandomizer(GURPSCharacter character) {
        super(new BorderLayout());
        mCharacter = character;
        mCheckBoxes = new JCheckBox[COUNT];
        mFields = new JTextField[COUNT];
        JPanel wrapper = new JPanel(new ColumnLayout(2));
        wrapper.setBorder(new EmptyBorder(10));
        Profile description = mCharacter.getProfile();
        addField(wrapper, I18n.Text("Gender:"), null, GENDER_INDEX, description.getGender());
        addField(wrapper, I18n.Text("Age:"), null, AGE_INDEX, description.getAge());
        addField(wrapper, I18n.Text("Birthday:"), null, BIRTHDAY_INDEX, description.getBirthday());
        addField(wrapper, I18n.Text("Height:"), null, HEIGHT_INDEX, description.getHeight().toString());
        addField(wrapper, I18n.Text("Weight:"), null, WEIGHT_INDEX, description.getWeight().toString());
        addField(wrapper, I18n.Text("Hair:"), I18n.Text("The character's hair style and color"), HAIR_INDEX, description.getHair());
        addField(wrapper, I18n.Text("Eyes:"), I18n.Text("The character's eye color"), EYES_INDEX, description.getEyeColor());
        addField(wrapper, I18n.Text("Skin:"), I18n.Text("The character's skin color"), SKIN_INDEX, description.getSkinColor());
        addField(wrapper, I18n.Text("Hand:"), I18n.Text("The character's preferred hand"), HAND_INDEX, description.getHandedness());
        add(wrapper, BorderLayout.CENTER);
        JButton randomize = new JButton(I18n.Text("Randomize"));
        randomize.addActionListener(this);
        add(randomize, BorderLayout.SOUTH);
    }

    private void addField(Container wrapper, String title, String tooltip, int which, String value) {
        mCheckBoxes[which] = new JCheckBox(title, true);
        String wrappedTooltip = Text.wrapPlainTextForToolTip(tooltip);
        mCheckBoxes[which].setToolTipText(wrappedTooltip);
        wrapper.add(mCheckBoxes[which]);
        mFields[which] = new JTextField(value, 20);
        mFields[which].setToolTipText(wrappedTooltip);
        mFields[which].setEnabled(false);
        wrapper.add(mFields[which]);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Profile description = mCharacter.getProfile();
        if (mCheckBoxes[GENDER_INDEX].isSelected()) {
            mFields[GENDER_INDEX].setText(Profile.getRandomGender());
        }
        if (mCheckBoxes[AGE_INDEX].isSelected()) {
            mFields[AGE_INDEX].setText(Numbers.format(description.getRandomAge()));
        }
        if (mCheckBoxes[BIRTHDAY_INDEX].isSelected()) {
            mFields[BIRTHDAY_INDEX].setText(Profile.getRandomMonthAndDay());
        }
        if (mCheckBoxes[HEIGHT_INDEX].isSelected()) {
            mFields[HEIGHT_INDEX].setText(description.getRandomHeight(mCharacter.getStrength(), description.getSizeModifier()).toString());
        }
        if (mCheckBoxes[WEIGHT_INDEX].isSelected()) {
            mFields[WEIGHT_INDEX].setText(description.getRandomWeight(mCharacter.getStrength(), description.getSizeModifier(), description.getWeightMultiplier()).toString());
        }
        if (mCheckBoxes[HAIR_INDEX].isSelected()) {
            mFields[HAIR_INDEX].setText(Profile.getRandomHair());
        }
        if (mCheckBoxes[EYES_INDEX].isSelected()) {
            mFields[EYES_INDEX].setText(Profile.getRandomEyeColor());
        }
        if (mCheckBoxes[SKIN_INDEX].isSelected()) {
            mFields[SKIN_INDEX].setText(Profile.getRandomSkinColor());
        }
        if (mCheckBoxes[HAND_INDEX].isSelected()) {
            mFields[HAND_INDEX].setText(Profile.getRandomHandedness());
        }
    }

    /** Apply the changes. */
    public void applyChanges() {
        MultipleUndo edit        = new MultipleUndo(I18n.Text("Description Randomization"));
        Profile      description = mCharacter.getProfile();
        mCharacter.addEdit(edit);
        mCharacter.startNotify();
        description.setGender(mFields[GENDER_INDEX].getText());
        description.setAge(mFields[AGE_INDEX].getText());
        description.setBirthday(mFields[BIRTHDAY_INDEX].getText());
        description.setHeight(LengthValue.extract(mFields[HEIGHT_INDEX].getText(), true));
        description.setWeight(WeightValue.extract(mFields[WEIGHT_INDEX].getText(), true));
        description.setHair(mFields[HAIR_INDEX].getText());
        description.setEyeColor(mFields[EYES_INDEX].getText());
        description.setSkinColor(mFields[SKIN_INDEX].getText());
        description.setHandedness(mFields[HAND_INDEX].getText());
        mCharacter.endNotify();
        edit.end();
    }
}
