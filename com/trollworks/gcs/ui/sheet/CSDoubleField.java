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
import com.trollworks.toolkit.text.TKNumberFilter;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKNumberUtils;

/** A double input field. */
public class CSDoubleField extends CSField {
	private double	mMinValue;
	private double	mMaxValue;
	private boolean	mForceSign;

	/**
	 * Creates a new, right-aligned, double input field.
	 * 
	 * @param character The character to listen to.
	 * @param consumedType The field to listen to.
	 * @param forceSign Whether or not a plus sign should be forced for positive numbers.
	 * @param minValue The minimum value allowed.
	 * @param maxValue The maximum value allowed.
	 * @param tooltip The tooltip to set.
	 */
	public CSDoubleField(CMCharacter character, String consumedType, boolean forceSign, double minValue, double maxValue, String tooltip) {
		this(character, consumedType, TKAlignment.RIGHT, forceSign, minValue, maxValue, true, tooltip);
	}

	/**
	 * Creates a new double input field.
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
	public CSDoubleField(CMCharacter character, String consumedType, int alignment, boolean forceSign, double minValue, double maxValue, boolean editable, String tooltip) {
		super(character, consumedType, alignment, editable, tooltip);
		mForceSign = forceSign;
		mMinValue = minValue;
		mMaxValue = maxValue;
		int max = TKNumberUtils.format(maxValue, mForceSign).length();
		int min = TKNumberUtils.format(minValue, mForceSign).length();

		if (min > max) {
			max = min;
		}
		setKeyEventFilter(new TKNumberFilter(true, minValue < 0, max));
		initialize();
	}

	@Override protected Object getObjectToSet() {
		double number = TKNumberUtils.getDouble(getText(), mMinValue <= 0 && mMaxValue >= 0 ? 0 : mMinValue);

		if (number < mMinValue) {
			number = mMinValue;
		} else if (number > mMaxValue) {
			number = mMaxValue;
		}
		return new Double(number);
	}

	@Override public void handleNotification(Object producer, String type, Object data) {
		setText(TKNumberUtils.format(((Double) data).doubleValue(), mForceSign));
		invalidate();
	}
}
