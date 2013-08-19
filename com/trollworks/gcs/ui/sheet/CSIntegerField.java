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
import com.trollworks.gcs.ui.common.CSMenuKeys;
import com.trollworks.gcs.ui.common.CSWindow;
import com.trollworks.toolkit.text.TKNumberFilter;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.widget.menu.TKMenuItem;

/** An integer input field. */
public class CSIntegerField extends CSField {
	private int		mMinValue;
	private int		mMaxValue;
	private boolean	mForceSign;
	private boolean	mDisplayZero;

	/**
	 * Creates a new, right-aligned, integer input field.
	 * 
	 * @param character The character to listen to.
	 * @param consumedType The field to listen to.
	 * @param forceSign Whether or not a plus sign should be forced for positive numbers.
	 * @param minValue The minimum value allowed.
	 * @param maxValue The maximum value allowed.
	 * @param tooltip The tooltip to set.
	 */
	public CSIntegerField(CMCharacter character, String consumedType, boolean forceSign, int minValue, int maxValue, String tooltip) {
		this(character, consumedType, TKAlignment.RIGHT, forceSign, minValue, maxValue, true, tooltip);
	}

	/**
	 * Creates a new, right-aligned, integer input field.
	 * 
	 * @param character The character to listen to.
	 * @param consumedType The field to listen to.
	 * @param forceSign Whether or not a plus sign should be forced for positive numbers.
	 * @param minValue The minimum value allowed.
	 * @param maxValue The maximum value allowed.
	 * @param editable Whether or not the user can edit this field.
	 * @param tooltip The tooltip to set.
	 */
	public CSIntegerField(CMCharacter character, String consumedType, boolean forceSign, int minValue, int maxValue, boolean editable, String tooltip) {
		this(character, consumedType, TKAlignment.RIGHT, forceSign, minValue, maxValue, editable, tooltip);
	}

	/**
	 * Creates a new integer input field.
	 * 
	 * @param character The character to listen to.
	 * @param consumedType The field to listen to.
	 * @param alignment The alignment of the field.
	 * @param forceSign Whether or not a plus sign should be forced for positive numbers.
	 * @param minValue The minimum value allowed.
	 * @param maxValue The maximum value allowed.
	 * @param tooltip The tooltip to set.
	 */
	public CSIntegerField(CMCharacter character, String consumedType, int alignment, boolean forceSign, int minValue, int maxValue, String tooltip) {
		this(character, consumedType, alignment, forceSign, minValue, maxValue, true, tooltip);
	}

	/**
	 * Creates a new integer input field.
	 * 
	 * @param character The character to listen to.
	 * @param consumedType The field to listen to.
	 * @param alignment The alignment of the field.
	 * @param forceSign Whether or not a plus sign should be forced for positive numbers.
	 * @param minValue The minimum value allowed.
	 * @param maxValue The maximum value allowed.
	 * @param editable Whether or not the user can edit this field.
	 * @param tooltip The tooltip to set.
	 */
	public CSIntegerField(CMCharacter character, String consumedType, int alignment, boolean forceSign, int minValue, int maxValue, boolean editable, String tooltip) {
		super(character, consumedType, alignment, editable, tooltip);
		mDisplayZero = true;
		mForceSign = forceSign;
		mMinValue = minValue;
		mMaxValue = maxValue;
		int max = TKNumberUtils.format(maxValue, mForceSign).length();
		int min = TKNumberUtils.format(minValue, mForceSign).length();

		if (min > max) {
			max = min;
		}
		setKeyEventFilter(new TKNumberFilter(false, minValue < 0, max));
		initialize();
	}

	@Override public boolean adjustMenuItem(String command, TKMenuItem item) {
		if (CSWindow.CMD_INCREMENT.equals(command)) {
			item.setTitle(CSMenuKeys.getTitle(command));
			item.setEnabled(isEnabled() && canIncrement());
		} else if (CSWindow.CMD_DECREMENT.equals(command)) {
			item.setTitle(CSMenuKeys.getTitle(command));
			item.setEnabled(isEnabled() && canDecrement());
		} else {
			return super.adjustMenuItem(command, item);
		}
		return true;
	}

	@Override public boolean obeyCommand(String command, TKMenuItem item) {
		if (CSWindow.CMD_INCREMENT.equals(command)) {
			adjustValue(1);
		} else if (CSWindow.CMD_DECREMENT.equals(command)) {
			adjustValue(-1);
		} else {
			return super.obeyCommand(command, item);
		}
		return true;
	}

	private boolean canIncrement() {
		return mMaxValue > TKNumberUtils.getInteger(getText(), mMinValue <= 0 && mMaxValue >= 0 ? 0 : mMinValue);
	}

	private void adjustValue(int amt) {
		int value = TKNumberUtils.getInteger(getText(), mMinValue <= 0 && mMaxValue >= 0 ? 0 : mMinValue) + amt;

		if (value < mMinValue) {
			value = mMinValue;
		} else if (value > mMaxValue) {
			value = mMaxValue;
		}
		setText(TKNumberUtils.format(value, mForceSign));
	}

	private boolean canDecrement() {
		return mMinValue < TKNumberUtils.getInteger(getText(), mMinValue <= 0 && mMaxValue >= 0 ? 0 : mMinValue);
	}

	@Override protected Object getObjectToSet() {
		int number = TKNumberUtils.getInteger(getText(), mMinValue <= 0 && mMaxValue >= 0 ? 0 : mMinValue);

		if (number < mMinValue) {
			number = mMinValue;
		} else if (number > mMaxValue) {
			number = mMaxValue;
		}
		return new Integer(number);
	}

	@Override public void setText(String text) {
		if (!mDisplayZero && TKNumberUtils.getInteger(text, 0) == 0) {
			text = ""; //$NON-NLS-1$
		}
		super.setText(text);
	}

	@Override public void handleNotification(Object producer, String type, Object data) {
		setText(TKNumberUtils.format(((Integer) data).intValue(), mForceSign));
		invalidate();
	}

	/** @return Whether the value of 0 should be displayed or not. */
	public boolean displayZero() {
		return mDisplayZero;
	}

	/** @param displayZero Whether the value of 0 should be displayed or not. */
	public void setDisplayZero(boolean displayZero) {
		mDisplayZero = displayZero;
		setText(getText());
	}
}
