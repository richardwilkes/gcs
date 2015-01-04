/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.layout.ColumnLayout;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.text.Numbers;
import com.trollworks.toolkit.utility.undo.MultipleUndo;
import com.trollworks.toolkit.utility.units.LengthValue;
import com.trollworks.toolkit.utility.units.WeightValue;

import java.awt.BorderLayout;
import java.awt.Container;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;

import javax.swing.JButton;
import javax.swing.JCheckBox;
import javax.swing.JPanel;
import javax.swing.JTextField;
import javax.swing.border.EmptyBorder;

/** A character description randomizer. */
public class DescriptionRandomizer extends JPanel implements ActionListener {
	@Localize("Randomize")
	@Localize(locale = "de", value = "Zufallswerte erzeugen")
	@Localize(locale = "ru", value = "Сгенерировать")
	private static String		RANDOMIZE;
	@Localize("Description Randomization")
	@Localize(locale = "de", value = "Zufällige Beschreibungen")
	@Localize(locale = "ru", value = "Создать описание")
	private static String		UNDO_RANDOMIZE;

	static {
		Localization.initialize();
	}

	private static final int	GENDER_INDEX	= 0;
	private static final int	AGE_INDEX		= 1;
	private static final int	BIRTHDAY_INDEX	= 2;
	private static final int	HEIGHT_INDEX	= 3;
	private static final int	WEIGHT_INDEX	= 4;
	private static final int	HAIR_INDEX		= 5;
	private static final int	EYES_INDEX		= 6;
	private static final int	SKIN_INDEX		= 7;
	private static final int	HAND_INDEX		= 8;
	private static final int	COUNT			= 9;
	private GURPSCharacter		mCharacter;
	private JCheckBox[]			mCheckBoxes;
	private JTextField[]		mFields;
	private JButton				mRandomize;

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
		wrapper.setBorder(new EmptyBorder(10, 10, 10, 10));
		Profile description = mCharacter.getDescription();
		addField(wrapper, DescriptionPanel.GENDER, null, GENDER_INDEX, description.getGender());
		addField(wrapper, DescriptionPanel.AGE, null, AGE_INDEX, Numbers.format(description.getAge()));
		addField(wrapper, DescriptionPanel.BIRTHDAY, null, BIRTHDAY_INDEX, description.getBirthday());
		addField(wrapper, DescriptionPanel.HEIGHT_FIELD, null, HEIGHT_INDEX, description.getHeight().toString());
		addField(wrapper, DescriptionPanel.WEIGHT, null, WEIGHT_INDEX, description.getWeight().toString());
		addField(wrapper, DescriptionPanel.HAIR, DescriptionPanel.HAIR_TOOLTIP, HAIR_INDEX, description.getHair());
		addField(wrapper, DescriptionPanel.EYE_COLOR, DescriptionPanel.EYE_COLOR_TOOLTIP, EYES_INDEX, description.getEyeColor());
		addField(wrapper, DescriptionPanel.SKIN_COLOR, DescriptionPanel.SKIN_COLOR_TOOLTIP, SKIN_INDEX, description.getSkinColor());
		addField(wrapper, DescriptionPanel.HANDEDNESS, DescriptionPanel.HANDEDNESS_TOOLTIP, HAND_INDEX, description.getHandedness());
		add(wrapper, BorderLayout.CENTER);
		mRandomize = new JButton(RANDOMIZE);
		mRandomize.addActionListener(this);
		add(mRandomize, BorderLayout.SOUTH);
	}

	private void addField(Container wrapper, String title, String tooltip, int which, String value) {
		mCheckBoxes[which] = new JCheckBox(title, true);
		mCheckBoxes[which].setToolTipText(tooltip);
		wrapper.add(mCheckBoxes[which]);

		mFields[which] = new JTextField(value, 20);
		mFields[which].setToolTipText(tooltip);
		mFields[which].setEnabled(false);
		wrapper.add(mFields[which]);
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		Profile description = mCharacter.getDescription();
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
			mFields[HEIGHT_INDEX].setText(Profile.getRandomHeight(mCharacter.getStrength(), description.getSizeModifier()).toString());
		}
		if (mCheckBoxes[WEIGHT_INDEX].isSelected()) {
			mFields[WEIGHT_INDEX].setText(Profile.getRandomWeight(mCharacter.getStrength(), description.getSizeModifier(), description.getWeightMultiplier()).toString());
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
		MultipleUndo edit = new MultipleUndo(UNDO_RANDOMIZE);
		Profile description = mCharacter.getDescription();
		mCharacter.addEdit(edit);
		mCharacter.startNotify();
		description.setGender(mFields[GENDER_INDEX].getText());
		description.setAge(Numbers.getLocalizedInteger(mFields[AGE_INDEX].getText(), 18));
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
