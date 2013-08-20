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
 * 2005-2013 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.criteria;

import com.trollworks.ttk.units.WeightValue;
import com.trollworks.ttk.xml.XMLReader;

import java.io.IOException;

/** Manages weight comparison criteria. */
public class WeightCriteria extends NumericCriteria {
	private WeightValue	mQualifier;

	/**
	 * Creates a new double comparison.
	 * 
	 * @param type The {@link NumericCompareType} to use.
	 * @param qualifier The qualifier to match against.
	 */
	public WeightCriteria(NumericCompareType type, WeightValue qualifier) {
		super(type);
		setQualifier(qualifier);
	}

	/**
	 * Creates a new double comparison.
	 * 
	 * @param other A {@link WeightCriteria} to clone.
	 */
	public WeightCriteria(WeightCriteria other) {
		super(other.getType());
		mQualifier = new WeightValue(other.mQualifier);
	}

	@Override
	public boolean equals(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof WeightCriteria && super.equals(obj)) {
			return mQualifier.equals(((WeightCriteria) obj).mQualifier);
		}
		return false;
	}

	@Override
	public int hashCode() {
		return super.hashCode();
	}

	@Override
	public void load(XMLReader reader) throws IOException {
		super.load(reader);
		setQualifier(WeightValue.extract(reader.readText(), false));
	}

	/** @return The qualifier to match against. */
	public WeightValue getQualifier() {
		return mQualifier;
	}

	@Override
	public String getQualifierAsString(boolean allowAdornments) {
		return mQualifier.toString(allowAdornments);
	}

	/** @param qualifier The qualifier to match against. */
	public void setQualifier(WeightValue qualifier) {
		mQualifier = new WeightValue(qualifier);
	}

	/**
	 * @param data The data to match against.
	 * @return Whether the data matches this criteria.
	 */
	public boolean matches(WeightValue data) {
		switch (getType()) {
			case IS:
				return mQualifier.equals(data);
			case AT_LEAST:
			default:
				return data.getNormalizedValue() >= mQualifier.getNormalizedValue();
			case AT_MOST:
				return data.getNormalizedValue() <= mQualifier.getNormalizedValue();
		}
	}
}
