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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.utility.units;

import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.text.NumberUtils;

import java.text.MessageFormat;

/** Common length units. */
public enum LengthUnits implements Units {
	/** Points (1/72 of an inch). */
	POINTS(1.0 / 72.0) {
		@Override public String toString() {
			return MSG_POINTS_ABBREVIATION;
		}
	},
	/** Inches. */
	INCHES(1.0) {
		@Override public String toString() {
			return MSG_INCHES_ABBREVIATION;
		}
	},
	/** Feet. */
	FEET(12.0) {
		@Override public String toString() {
			return MSG_FEET_ABBREVIATION;
		}
	},
	/** Yards. */
	YARDS(36.0) {
		@Override public String toString() {
			return MSG_YARDS_ABBREVIATION;
		}
	},
	/** Miles. */
	MILES(5280.0 * 12.0) {
		@Override public String toString() {
			return MSG_MILES_ABBREVIATION;
		}
	},
	/** Millimeters. */
	MILLIMETERS(0.1 / 2.54) {
		@Override public String toString() {
			return MSG_MILLIMETERS_ABBREVIATION;
		}
	},
	/** Centimeters. */
	CENTIMETERS(1.0 / 2.54) {
		@Override public String toString() {
			return MSG_CENTIMETERS_ABBREVIATION;
		}
	},
	/** Meters. */
	METERS(100.0 / 2.54) {
		@Override public String toString() {
			return MSG_METERS_ABBREVIATION;
		}
	},
	/** Kilometers. */
	KILOMETERS(100000.0 / 2.54) {
		@Override public String toString() {
			return MSG_KILOMETERS_ABBREVIATION;
		}
	};

	static String	MSG_POINTS_ABBREVIATION;
	static String	MSG_INCHES_ABBREVIATION;
	static String	MSG_FEET_ABBREVIATION;
	static String	MSG_YARDS_ABBREVIATION;
	static String	MSG_MILES_ABBREVIATION;
	static String	MSG_MILLIMETERS_ABBREVIATION;
	static String	MSG_CENTIMETERS_ABBREVIATION;
	static String	MSG_METERS_ABBREVIATION;
	static String	MSG_KILOMETERS_ABBREVIATION;
	static String	MSG_FORMAT;
	private double	mFactor;

	static {
		LocalizedMessages.initialize(LengthUnits.class);
	}

	private LengthUnits(double factor) {
		mFactor = factor;
	}

	public double convert(Units units, double value) {
		return value * units.getFactor() / mFactor;
	}

	public double normalize(double value) {
		return value / mFactor;
	}

	public double getFactor() {
		return mFactor;
	}

	public String format(double value) {
		return MessageFormat.format(MSG_FORMAT, NumberUtils.format(value), toString());
	}

	public Units[] getCompatibleUnits() {
		return values();
	}
}
