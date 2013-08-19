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

package com.trollworks.gcs.utility.units;

/** Specifies the methods a type of unit must implement. */
public interface Units {
	/**
	 * @param value The value to format.
	 * @return The formatted value.
	 */
	String format(double value);

	/**
	 * Converts from a specified units type into this units type.
	 * 
	 * @param units The units to convert from.
	 * @param value The value to convert.
	 * @return The new value, in units of this type.
	 */
	double convert(Units units, double value);

	/**
	 * Normalizes a value to a common scale.
	 * 
	 * @param value The value to normalize.
	 * @return The normalized value.
	 */
	public double normalize(double value);

	/** @return The factor used. */
	double getFactor();

	/** @return The reference name of the units (not localized). */
	String name();

	/** @return An array of compatible units. */
	Units[] getCompatibleUnits();
}
