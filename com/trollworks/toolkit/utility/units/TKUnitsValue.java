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

import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;

import java.io.IOException;
import java.text.MessageFormat;

/**
 * Holds a value and {@link TKUnits} pair.
 * 
 * @param <T> The type of {@link TKUnits} to use.
 */
public class TKUnitsValue<T extends TKUnits> implements Comparable<TKUnitsValue<T>> {
	private static final String	ATTRIBUTE_UNITS	= "units";	//$NON-NLS-1$
	private double				mValue;
	private T					mUnits;

	/**
	 * Creates a new {@link TKUnitsValue}.
	 * 
	 * @param value The value to use.
	 * @param units The {@link TKUnits} to use.
	 */
	public TKUnitsValue(double value, T units) {
		mValue = value;
		mUnits = units;
	}

	/**
	 * Creates a new {@link TKUnitsValue}.
	 * 
	 * @param other A {@link TKUnitsValue} to clone.
	 */
	public TKUnitsValue(TKUnitsValue<T> other) {
		set(other);
	}

	/** @param other A {@link TKUnitsValue} to copy state from. */
	public void set(TKUnitsValue<T> other) {
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

	public int compareTo(TKUnitsValue<T> other) {
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
		if (other instanceof TKUnitsValue) {
			TKUnitsValue<?> ouv = (TKUnitsValue<?>) other;

			return mUnits == ouv.mUnits && mValue == ouv.mValue;
		}
		return false;
	}

	/**
	 * Loads the contents of this {@link TKUnitsValue}.
	 * 
	 * @param reader The {@link TKXMLReader} to use.
	 * @throws IOException
	 */
	@SuppressWarnings("unchecked") public void load(TKXMLReader reader) throws IOException {
		String unitsAttr = reader.getAttribute(ATTRIBUTE_UNITS);

		mValue = reader.readDouble(0);

		for (TKUnits one : mUnits.getCompatibleUnits()) {
			if (one.name().equals(unitsAttr)) {
				mUnits = (T) one;
				return;
			}
		}

		mUnits = getDefaultUnits();
		if (mUnits == null) {
			throw new IOException(MessageFormat.format(Msgs.INVALID_UNITS, unitsAttr));
		}
	}

	/**
	 * Saves the contents of this {@link TKUnitsValue} out as an XML tag.
	 * 
	 * @param out The {@link TKXMLWriter} to use.
	 * @param tag The XML tag to use.
	 */
	public void save(TKXMLWriter out, String tag) {
		out.simpleTagWithAttribute(tag, getValue(), ATTRIBUTE_UNITS, getUnits().name());
	}
}
