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
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.ui.sheet;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.ui.common.CSImage;
import com.trollworks.toolkit.undo.TKMultipleUndo;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.utility.units.TKLengthUnits;
import com.trollworks.toolkit.utility.units.TKWeightUnits;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKTextField;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.button.TKButton;
import com.trollworks.toolkit.widget.button.TKCheckbox;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.layout.TKCompassLayout;
import com.trollworks.toolkit.widget.layout.TKCompassPosition;
import com.trollworks.toolkit.window.TKDialog;
import com.trollworks.toolkit.window.TKOptionDialog;

import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;

/** A character description randomizer. */
public class CSDescriptionRandomizer extends TKPanel implements ActionListener {
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
	private CMCharacter			mCharacter;
	private TKCheckbox[]		mCheckBoxes;
	private TKTextField[]		mFields;
	private TKButton			mRandomize;

	/**
	 * Brings up a modal dialog that allows randomizing the description of the character.
	 * 
	 * @param character The character to work on.
	 */
	static public void randomize(CMCharacter character) {
		TKOptionDialog dialog = new TKOptionDialog(Msgs.RANDOMIZER, TKOptionDialog.TYPE_OK_CANCEL);
		CSDescriptionRandomizer panel = new CSDescriptionRandomizer(character);

		dialog.setOKButtonTitle(Msgs.APPLY);
		dialog.setResizable(true);
		dialog.setDefaultButton(panel.mRandomize);
		if (dialog.doModal(CSImage.getCharacterSheetIcon(true), panel) == TKDialog.OK) {
			panel.applyChanges();
		}
	}

	private CSDescriptionRandomizer(CMCharacter character) {
		super(new TKCompassLayout());
		mCharacter = character;
		mCheckBoxes = new TKCheckbox[COUNT];
		mFields = new TKTextField[COUNT];
		TKPanel wrapper = new TKPanel(new TKColumnLayout(2));
		wrapper.setBorder(new TKEmptyBorder(10));
		addField(wrapper, Msgs.GENDER, null, GENDER_INDEX, mCharacter.getGender());
		addField(wrapper, Msgs.AGE, null, AGE_INDEX, TKNumberUtils.format(mCharacter.getAge()));
		addField(wrapper, Msgs.BIRTHDAY, null, BIRTHDAY_INDEX, mCharacter.getBirthday());
		addField(wrapper, Msgs.HEIGHT, null, HEIGHT_INDEX, TKLengthUnits.FEET.format(mCharacter.getHeight()));
		addField(wrapper, Msgs.WEIGHT, null, WEIGHT_INDEX, TKWeightUnits.POUNDS.format(mCharacter.getWeight()));
		addField(wrapper, Msgs.HAIR, Msgs.HAIR_TOOLTIP, HAIR_INDEX, mCharacter.getHair());
		addField(wrapper, Msgs.EYE_COLOR, Msgs.EYE_COLOR_TOOLTIP, EYES_INDEX, mCharacter.getEyeColor());
		addField(wrapper, Msgs.SKIN_COLOR, Msgs.SKIN_COLOR_TOOLTIP, SKIN_INDEX, mCharacter.getSkinColor());
		addField(wrapper, Msgs.HANDEDNESS, Msgs.HANDEDNESS_TOOLTIP, HAND_INDEX, mCharacter.getHandedness());
		add(wrapper, TKCompassPosition.CENTER);
		mRandomize = new TKButton(Msgs.RANDOMIZE);
		mRandomize.addActionListener(this);
		add(mRandomize, TKCompassPosition.SOUTH);
	}

	private void addField(TKPanel wrapper, String title, String tooltip, int which, String value) {
		mCheckBoxes[which] = new TKCheckbox(title, true, TKAlignment.RIGHT);
		mCheckBoxes[which].setToolTipText(tooltip);
		wrapper.add(mCheckBoxes[which]);

		mFields[which] = new TKTextField(value, 200);
		mFields[which].setToolTipText(tooltip);
		mFields[which].setEnabled(false);
		wrapper.add(mFields[which]);
	}

	public void actionPerformed(ActionEvent event) {
		if (mCheckBoxes[GENDER_INDEX].isChecked()) {
			mFields[GENDER_INDEX].setText(CMCharacter.getRandomGender());
		}
		if (mCheckBoxes[AGE_INDEX].isChecked()) {
			mFields[AGE_INDEX].setText(TKNumberUtils.format(mCharacter.getRandomAge()));
		}
		if (mCheckBoxes[BIRTHDAY_INDEX].isChecked()) {
			mFields[BIRTHDAY_INDEX].setText(CMCharacter.getRandomMonthAndDay());
		}
		if (mCheckBoxes[HEIGHT_INDEX].isChecked()) {
			mFields[HEIGHT_INDEX].setText(TKNumberUtils.formatHeight(CMCharacter.getRandomHeight(mCharacter.getStrength())));
		}
		if (mCheckBoxes[WEIGHT_INDEX].isChecked()) {
			mFields[WEIGHT_INDEX].setText(TKWeightUnits.POUNDS.format(CMCharacter.getRandomWeight(mCharacter.getStrength(), mCharacter.getWeightMultiplier())));
		}
		if (mCheckBoxes[HAIR_INDEX].isChecked()) {
			mFields[HAIR_INDEX].setText(CMCharacter.getRandomHair());
		}
		if (mCheckBoxes[EYES_INDEX].isChecked()) {
			mFields[EYES_INDEX].setText(CMCharacter.getRandomEyeColor());
		}
		if (mCheckBoxes[SKIN_INDEX].isChecked()) {
			mFields[SKIN_INDEX].setText(CMCharacter.getRandomSkinColor());
		}
		if (mCheckBoxes[HAND_INDEX].isChecked()) {
			mFields[HAND_INDEX].setText(CMCharacter.getRandomHandedness());
		}
	}

	private void applyChanges() {
		TKMultipleUndo undo = new TKMultipleUndo(Msgs.UNDO_RANDOMIZE);

		mCharacter.addEdit(undo);
		mCharacter.startNotify();
		mCharacter.setGender(mFields[GENDER_INDEX].getText());
		mCharacter.setAge(TKNumberUtils.getInteger(mFields[AGE_INDEX].getText(), 18));
		mCharacter.setBirthday(mFields[BIRTHDAY_INDEX].getText());
		mCharacter.setHeight(TKNumberUtils.getHeight(mFields[HEIGHT_INDEX].getText()));
		mCharacter.setWeight(TKNumberUtils.getDouble(mFields[WEIGHT_INDEX].getText(), 0.0));
		mCharacter.setHair(mFields[HAIR_INDEX].getText());
		mCharacter.setEyeColor(mFields[EYES_INDEX].getText());
		mCharacter.setSkinColor(mFields[SKIN_INDEX].getText());
		mCharacter.setHandedness(mFields[HAND_INDEX].getText());
		mCharacter.endNotify();
		undo.end();
	}
}
