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

import com.trollworks.toolkit.io.TKMessages;

/** All localized strings in this package should be accessed via this class. */
class Msgs {
	/** Used by {@link TKWeightUnits}. */
	public static String	OUNCES_ABBREVIATION;
	/** Used by {@link TKWeightUnits}. */
	public static String	POUNDS_ABBREVIATION;
	/** Used by {@link TKWeightUnits}. */
	public static String	GRAMS_ABBREVIATION;
	/** Used by {@link TKWeightUnits}. */
	public static String	KILOGRAMS_ABBREVIATION;

	/** Used by {@link TKLengthUnits}. */
	public static String	POINTS_ABBREVIATION;
	/** Used by {@link TKLengthUnits}. */
	public static String	INCHES_ABBREVIATION;
	/** Used by {@link TKLengthUnits}. */
	public static String	FEET_ABBREVIATION;
	/** Used by {@link TKLengthUnits}. */
	public static String	YARDS_ABBREVIATION;
	/** Used by {@link TKLengthUnits}. */
	public static String	MILES_ABBREVIATION;
	/** Used by {@link TKLengthUnits}. */
	public static String	MILLIMETERS_ABBREVIATION;
	/** Used by {@link TKLengthUnits}. */
	public static String	CENTIMETERS_ABBREVIATION;
	/** Used by {@link TKLengthUnits}. */
	public static String	METERS_ABBREVIATION;
	/** Used by {@link TKLengthUnits}. */
	public static String	KILOMETERS_ABBREVIATION;
	/** Used by {@link TKLengthUnits}. */
	public static String	FEET_FORMAT;

	/** Used by {@link TKWeightUnits} and {@link TKLengthUnits}. */
	public static String	GENERIC_FORMAT;

	/** Used by {@link TKUnitsValue}. */
	public static String	INVALID_UNITS;

	static {
		TKMessages.initialize(Msgs.class);
	}
}
