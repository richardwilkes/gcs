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

import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.text.NumberUtils;

import java.text.MessageFormat;

/** Common weight units. */
public enum WeightUnits implements Units {
	/** Ounces. */
	OUNCES(1.0 / 16.0) {
		@Override public String toString() {
			return MSG_OUNCES_ABBREVIATION;
		}
	},
	/** Pounds. */
	POUNDS(1.0) {
		@Override public String toString() {
			return MSG_POUNDS_ABBREVIATION;
		}

		@Override public String format(double value) {
			return MessageFormat.format(MSG_FORMAT, NumberUtils.format(value), value == 1.0 ? MSG_POUND_ABBREVIATION : MSG_POUNDS_ABBREVIATION);
		}
	},
	/** Grams. */
	GRAMS(0.002205) {
		@Override public String toString() {
			return MSG_GRAMS_ABBREVIATION;
		}
	},
	/** Kilograms. */
	KILOGRAMS(2.205) {
		@Override public String toString() {
			return MSG_KILOGRAMS_ABBREVIATION;
		}
	};

	static String	MSG_OUNCES_ABBREVIATION;
	static String	MSG_POUND_ABBREVIATION;
	static String	MSG_POUNDS_ABBREVIATION;
	static String	MSG_GRAMS_ABBREVIATION;
	static String	MSG_KILOGRAMS_ABBREVIATION;
	static String	MSG_FORMAT;
	private double	mFactor;

	static {
		LocalizedMessages.initialize(WeightUnits.class);
	}

	private WeightUnits(double factor) {
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
