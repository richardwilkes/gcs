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

package com.trollworks.gcs.utility.text;

import java.text.ParseException;

import javax.swing.JFormattedTextField;

/** Provides integer field conversion. */
public class IntegerFormatter extends JFormattedTextField.AbstractFormatter {
	private int		mMinValue;
	private int		mMaxValue;
	private boolean	mForceSign;

	/**
	 * Creates a new {@link IntegerFormatter}.
	 * 
	 * @param minValue The minimum value allowed.
	 * @param maxValue The maximum value allowed.
	 * @param forceSign Whether or not a plus sign should be forced for positive numbers.
	 */
	public IntegerFormatter(int minValue, int maxValue, boolean forceSign) {
		mMinValue = minValue;
		mMaxValue = maxValue;
		mForceSign = forceSign;
	}

	@Override public Object stringToValue(String text) throws ParseException {
		return new Integer(Math.min(Math.max(NumberUtils.getInteger(text, mMinValue <= 0 && mMaxValue >= 0 ? 0 : mMinValue), mMinValue), mMaxValue));
	}

	@Override public String valueToString(Object value) throws ParseException {
		return NumberUtils.format(((Integer) value).intValue(), mForceSign);
	}
}
