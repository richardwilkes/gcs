/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2011 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.character;

import com.trollworks.ttk.layout.ColumnLayout;
import com.trollworks.ttk.text.Numbers;
import com.trollworks.ttk.undo.MultipleUndo;
import com.trollworks.ttk.units.LengthValue;
import com.trollworks.ttk.units.WeightValue;
import com.trollworks.ttk.utility.LocalizedMessages;

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
	private static String		MSG_RANDOMIZE;
	private static String		MSG_UNDO_RANDOMIZE;
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

	static {
		LocalizedMessages.initialize(DescriptionRandomizer.class);
	}

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
		addField(wrapper, DescriptionPanel.MSG_GENDER, null, GENDER_INDEX, description.getGender());
		addField(wrapper, DescriptionPanel.MSG_AGE, null, AGE_INDEX, Numbers.format(description.getAge()));
		addField(wrapper, DescriptionPanel.MSG_BIRTHDAY, null, BIRTHDAY_INDEX, description.getBirthday());
		addField(wrapper, DescriptionPanel.MSG_HEIGHT, null, HEIGHT_INDEX, description.getHeight().toString());
		addField(wrapper, DescriptionPanel.MSG_WEIGHT, null, WEIGHT_INDEX, description.getWeight().toString());
		addField(wrapper, DescriptionPanel.MSG_HAIR, DescriptionPanel.MSG_HAIR_TOOLTIP, HAIR_INDEX, description.getHair());
		addField(wrapper, DescriptionPanel.MSG_EYE_COLOR, DescriptionPanel.MSG_EYE_COLOR_TOOLTIP, EYES_INDEX, description.getEyeColor());
		addField(wrapper, DescriptionPanel.MSG_SKIN_COLOR, DescriptionPanel.MSG_SKIN_COLOR_TOOLTIP, SKIN_INDEX, description.getSkinColor());
		addField(wrapper, DescriptionPanel.MSG_HANDEDNESS, DescriptionPanel.MSG_HANDEDNESS_TOOLTIP, HAND_INDEX, description.getHandedness());
		add(wrapper, BorderLayout.CENTER);
		mRandomize = new JButton(MSG_RANDOMIZE);
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
		MultipleUndo edit = new MultipleUndo(MSG_UNDO_RANDOMIZE);
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
