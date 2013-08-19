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

package com.trollworks.toolkit.utility.units;

import com.trollworks.toolkit.utility.TKNumberUtils;

import java.text.MessageFormat;

/** Common weight units. */
public enum TKWeightUnits implements TKUnits {
	/** Ounces. */
	OUNCES(Msgs.OUNCES_ABBREVIATION, 1.0 / 16.0),
	/** Pounds. */
	POUNDS(Msgs.POUNDS_ABBREVIATION, 1.0),
	/** Grams. */
	GRAMS(Msgs.GRAMS_ABBREVIATION, 0.002205),
	/** Kilograms. */
	KILOGRAMS(Msgs.KILOGRAMS_ABBREVIATION, 2.205);

	private String	mAbbreviation;
	private double	mFactor;

	private TKWeightUnits(String abbreviation, double factor) {
		mAbbreviation = abbreviation;
		mFactor = factor;
	}

	@Override public String toString() {
		return mAbbreviation;
	}

	public double convert(TKUnits units, double value) {
		return value * units.getFactor() / mFactor;
	}

	public double normalize(double value) {
		return value / mFactor;
	}

	public double getFactor() {
		return mFactor;
	}

	public String format(double value) {
		return MessageFormat.format(Msgs.GENERIC_FORMAT, TKNumberUtils.format(value), mAbbreviation);
	}

	public TKUnits[] getCompatibleUnits() {
		return values();
	}
}
