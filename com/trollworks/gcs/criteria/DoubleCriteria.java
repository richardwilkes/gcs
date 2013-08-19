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
 * 2005-2009 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.criteria;

import com.trollworks.ttk.collections.Enums;
import com.trollworks.ttk.units.WeightUnits;
import com.trollworks.ttk.xml.XMLReader;
import com.trollworks.ttk.xml.XMLWriter;

import java.io.IOException;

/** Manages double comparison criteria. */
public class DoubleCriteria extends NumericCriteria {
	private static final String	ATTRIBUTE_UNITS	= "units";	//$NON-NLS-1$
	private double				mQualifier;
	private boolean				mIsWeight;

	/**
	 * Creates a new double comparison.
	 * 
	 * @param type The {@link NumericCompareType} to use.
	 * @param qualifier The qualifier to match against.
	 * @param isWeight Whether this number represents a weight.
	 */
	public DoubleCriteria(NumericCompareType type, double qualifier, boolean isWeight) {
		super(type);
		setQualifier(qualifier);
		mIsWeight = isWeight;
	}

	/**
	 * Creates a new double comparison.
	 * 
	 * @param other A {@link DoubleCriteria} to clone.
	 */
	public DoubleCriteria(DoubleCriteria other) {
		super(other.getType());
		mQualifier = other.mQualifier;
		mIsWeight = other.mIsWeight;
	}

	@Override
	public boolean equals(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof DoubleCriteria && super.equals(obj)) {
			DoubleCriteria criteria = (DoubleCriteria) obj;
			return mIsWeight == criteria.mIsWeight && mQualifier == criteria.mQualifier;
		}
		return false;
	}

	@Override
	public void load(XMLReader reader) throws IOException {
		super.load(reader);
		if (mIsWeight) {
			setQualifier(WeightUnits.POUNDS.convert(Enums.extract(reader.getAttribute(ATTRIBUTE_UNITS), WeightUnits.values(), WeightUnits.POUNDS), reader.readDouble(0)));
		} else {
			setQualifier(reader.readDouble(0.0));
		}
	}

	@Override
	public void save(XMLWriter out, String tag) {
		if (mIsWeight) {
			out.startTag(tag);
			out.writeAttribute(ATTRIBUTE_COMPARE, getType().name().toLowerCase());
			out.writeAttribute(ATTRIBUTE_UNITS, WeightUnits.POUNDS.toString());
			out.finishTag();
			out.writeEncodedData(getQualifierAsString(false));
			out.endTagEOL(tag, false);
		} else {
			super.save(out, tag);
		}
	}

	/** @return Whether this number represents a weight. */
	public boolean isWeight() {
		return mIsWeight;
	}

	/** @return The qualifier to match against. */
	public double getQualifier() {
		return mQualifier;
	}

	@Override
	public String getQualifierAsString(boolean allowAdornments) {
		if (mIsWeight && allowAdornments) {
			return WeightUnits.POUNDS.format(mQualifier, true);
		}
		return Double.toString(mQualifier);
	}

	/** @param qualifier The qualifier to match against. */
	public void setQualifier(double qualifier) {
		mQualifier = qualifier;
	}

	/**
	 * @param data The data to match against.
	 * @return Whether the data matches this criteria.
	 */
	public boolean matches(double data) {
		switch (getType()) {
			case IS:
				return data == mQualifier;
			case AT_LEAST:
			default:
				return data >= mQualifier;
			case AT_MOST:
				return data <= mQualifier;
		}
	}
}
