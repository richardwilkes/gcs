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
import com.trollworks.gcs.utility.io.xml.XMLReader;
import com.trollworks.gcs.utility.io.xml.XMLWriter;

import java.io.IOException;
import java.text.MessageFormat;

/**
 * Holds a value and {@link Units} pair.
 * 
 * @param <T> The type of {@link Units} to use.
 */
public class UnitsValue<T extends Units> implements Comparable<UnitsValue<T>> {
	private static String		MSG_INVALID_UNITS;
	private static final String	ATTRIBUTE_UNITS	= "units";	//$NON-NLS-1$
	private double				mValue;
	private T					mUnits;

	static {
		LocalizedMessages.initialize(UnitsValue.class);
	}

	/**
	 * Creates a new {@link UnitsValue}.
	 * 
	 * @param value The value to use.
	 * @param units The {@link Units} to use.
	 */
	public UnitsValue(double value, T units) {
		mValue = value;
		mUnits = units;
	}

	/**
	 * Creates a new {@link UnitsValue}.
	 * 
	 * @param other A {@link UnitsValue} to clone.
	 */
	public UnitsValue(UnitsValue<T> other) {
		set(other);
	}

	/** @param other A {@link UnitsValue} to copy state from. */
	public void set(UnitsValue<T> other) {
		mValue = other.mValue;
		mUnits = other.mUnits;
	}

	/** @return The units. */
	public T getUnits() {
		return mUnits;
	}

	/** @param units The value to set for units. */
	public void setUnits(T units) {
		mUnits = units;
	}

	/** @return The value. */
	public double getValue() {
		return mValue;
	}

	/** @param value The value to set for value. */
	public void setValue(double value) {
		mValue = value;
	}

	/** @return The normalized value. */
	public double getNormalizedValue() {
		return mUnits.normalize(mValue);
	}

	/**
	 * @return The default units to use during a load if nothing matches. <code>null</code> may be
	 *         returned to indicate an error should occur instead.
	 */
	public T getDefaultUnits() {
		return null;
	}

	public int compareTo(UnitsValue<T> other) {
		if (this == other) {
			return 0;
		}

		double result = getNormalizedValue() - other.getNormalizedValue();

		if (result < 0.0) {
			return -1;
		}
		if (result > 0.0) {
			return 1;
		}
		return 0;
	}

	@Override public String toString() {
		return mUnits.format(mValue);
	}

	@Override public boolean equals(Object other) {
		if (this == other) {
			return true;
		}
		if (other instanceof UnitsValue) {
			UnitsValue<?> ouv = (UnitsValue<?>) other;

			return mUnits == ouv.mUnits && mValue == ouv.mValue;
		}
		return false;
	}

	/**
	 * Loads the contents of this {@link UnitsValue}.
	 * 
	 * @param reader The {@link XMLReader} to use.
	 * @throws IOException
	 */
	@SuppressWarnings("unchecked") public void load(XMLReader reader) throws IOException {
		String unitsAttr = reader.getAttribute(ATTRIBUTE_UNITS);

		mValue = reader.readDouble(0);

		for (Units one : mUnits.getCompatibleUnits()) {
			if (one.name().equals(unitsAttr)) {
				mUnits = (T) one;
				return;
			}
		}

		mUnits = getDefaultUnits();
		if (mUnits == null) {
			throw new IOException(MessageFormat.format(MSG_INVALID_UNITS, unitsAttr));
		}
	}

	/**
	 * Saves the contents of this {@link UnitsValue} out as an XML tag.
	 * 
	 * @param out The {@link XMLWriter} to use.
	 * @param tag The XML tag to use.
	 */
	public void save(XMLWriter out, String tag) {
		out.simpleTagWithAttribute(tag, getValue(), ATTRIBUTE_UNITS, getUnits().name());
	}
}
